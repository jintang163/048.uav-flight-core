package mavlink

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"groundstation-backend/internal/config"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/nsq"
	"groundstation-backend/internal/service"
	"groundstation-backend/internal/websocket"
	"go.uber.org/zap"
)

type CommandManager struct {
	uavConns      map[uint64]net.Conn
	connMu        sync.RWMutex
	parser        *MAVLinkParser
	flightService *service.FlightService
	heartbeatMgr  *HeartbeatManager
	listenerTCP   net.Listener
	listenerUDP   *net.UDPConn
}

var commandManager *CommandManager
var commandOnce sync.Once

func NewCommandManager() *CommandManager {
	commandOnce.Do(func() {
		commandManager = &CommandManager{
			uavConns:      make(map[uint64]net.Conn),
			parser:        NewMAVLinkParser(),
			flightService: service.NewFlightService(),
			heartbeatMgr:  NewHeartbeatManager(),
		}
	})
	return commandManager
}

func (m *CommandManager) Start() error {
	if err := m.startTCPListener(); err != nil {
		return err
	}
	if err := m.startUDPListener(); err != nil {
		return err
	}
	go m.processMessages()
	return nil
}

func (m *CommandManager) startTCPListener() error {
	cfg := config.AppConfig.MAVLink
	addr := fmt.Sprintf(":%d", cfg.TCPPort)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	m.listenerTCP = listener
	go m.acceptTCPConnections()

	middleware.Logger.Info("MAVLink TCP server started", zap.String("addr", addr))
	return nil
}

func (m *CommandManager) startUDPListener() error {
	cfg := config.AppConfig.MAVLink
	addr := fmt.Sprintf(":%d", cfg.UDPPort)

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	m.listenerUDP = conn
	go m.handleUDPConnections()

	middleware.Logger.Info("MAVLink UDP server started", zap.String("addr", addr))
	return nil
}

func (m *CommandManager) acceptTCPConnections() {
	for {
		conn, err := m.listenerTCP.Accept()
		if err != nil {
			middleware.Logger.Error("TCP accept error", zap.Error(err))
			continue
		}

		uavID := m.generateUAVID(conn.RemoteAddr().String())
		m.connMu.Lock()
		m.uavConns[uavID] = conn
		m.connMu.Unlock()

		go m.handleConnection(uavID, conn)
	}
}

func (m *CommandManager) handleUDPConnections() {
	buf := make([]byte, 2048)
	for {
		n, remoteAddr, err := m.listenerUDP.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		uavID := m.generateUAVID(remoteAddr.String())
		m.parser.Parse(buf[:n])

		m.connMu.Lock()
		if _, exists := m.uavConns[uavID]; !exists {
			m.uavConns[uavID] = &UDPConn{addr: remoteAddr, conn: m.listenerUDP}
		}
		m.connMu.Unlock()
	}
}

type UDPConn struct {
	addr *net.UDPAddr
	conn *net.UDPConn
}

func (u *UDPConn) Write(b []byte) (int, error) {
	return u.conn.WriteToUDP(b, u.addr)
}

func (u *UDPConn) Read(b []byte) (int, error) {
	return 0, nil
}

func (u *UDPConn) Close() error {
	return nil
}

func (u *UDPConn) RemoteAddr() net.Addr {
	return u.addr
}

func (u *UDPConn) LocalAddr() net.Addr {
	return nil
}

func (u *UDPConn) SetDeadline(t time.Time) error {
	return nil
}

func (u *UDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (u *UDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (m *CommandManager) handleConnection(uavID uint64, conn net.Conn) {
	defer func() {
		conn.Close()
		m.connMu.Lock()
		delete(m.uavConns, uavID)
		m.connMu.Unlock()
	}()

	buf := make([]byte, 2048)
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		m.parser.Parse(buf[:n])
	}
}

func (m *CommandManager) processMessages() {
	for {
		msg, ok := m.parser.GetMessage()
		if !ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		m.processMAVLinkMessage(msg)
	}
}

func (m *CommandManager) processMAVLinkMessage(msg *MAVLinkMessage) {
	uavID := uint64(msg.SystemID)

	switch msg.MessageID {
	case HEARTBEAT:
		hb, err := ParseHeartbeat(msg.Payload)
		if err == nil {
			m.heartbeatMgr.ProcessHeartbeat(uavID, msg.SystemID, msg.ComponentID,
				hb.BaseMode, hb.CustomMode, hb.SystemStatus)
		}

	case SYS_STATUS:
		m.processSysStatus(uavID, msg.Payload)

	case GPS_RAW_INT:
		gps, err := ParseGPSRawInt(msg.Payload)
		if err == nil {
			flightStatus := &models.FlightStatus{
				UAVID:        uavID,
				Latitude:     gps.Latitude,
				Longitude:    gps.Longitude,
				Altitude:     gps.Altitude,
				SatellitesVisible: int(gps.Satellites),
				GPSFixType:   int(gps.FixType),
			}
			_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
				UAVID:     uavID,
				Latitude:  gps.Latitude,
				Longitude: gps.Longitude,
				Altitude:  gps.Altitude,
				Timestamp: time.Now(),
			})
			websocket.BroadcastTelemetry(uavID, flightStatus)
		}

	case GLOBAL_POSITION_INT:
		m.processGlobalPosition(uavID, msg.Payload)

	case ATTITUDE:
		att, err := ParseAttitude(msg.Payload)
		if err == nil {
			_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
				UAVID:     uavID,
				Roll:      att.Roll,
				Pitch:     att.Pitch,
				Yaw:       att.Yaw,
				RollSpeed: att.RollSpeed,
				PitchSpeed: att.PitchSpeed,
				YawSpeed:  att.YawSpeed,
				Timestamp: time.Now(),
			})
		}

	case BATTERY_STATUS:
		bat, err := ParseBatteryStatus(msg.Payload)
		if err == nil {
			_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
				UAVID:         uavID,
				BatteryVoltage: bat.Voltage,
				BatteryCurrent: bat.Current,
				BatteryRemaining: float64(bat.Remaining),
				Timestamp:     time.Now(),
			})
		}

	case RADIO_STATUS:
		m.processRadioStatus(uavID, msg.Payload)

	case MISSION_ITEM_REACHED:
		m.processMissionItemReached(uavID, msg.Payload)

	case COMMAND_ACK:
		m.processCommandAck(uavID, msg.Payload)
	}

	_ = nsq.Publish(nsq.TopicMAVLinkMessage, map[string]interface{}{
		"uav_id":     uavID,
		"message_id": msg.MessageID,
		"length":     msg.Length,
		"timestamp":  time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processSysStatus(uavID uint64, payload []byte) {
	if len(payload) >= 30 {
		batteryVoltage := float64(binary.LittleEndian.Uint16(payload[4:6])) / 1000
		batteryCurrent := float64(int16(binary.LittleEndian.Uint16(payload[6:8]))) / 100
		batteryRemaining := int8(payload[10])

		_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
			UAVID:             uavID,
			BatteryVoltage:    batteryVoltage,
			BatteryCurrent:    batteryCurrent,
			BatteryRemaining:  float64(batteryRemaining),
			Timestamp:         time.Now(),
		})
	}
}

func (m *CommandManager) processGlobalPosition(uavID uint64, payload []byte) {
	if len(payload) >= 28 {
		lat := float64(int32(binary.LittleEndian.Uint32(payload[4:8]))) / 1e7
		lng := float64(int32(binary.LittleEndian.Uint32(payload[8:12]))) / 1e7
		alt := float64(int32(binary.LittleEndian.Uint32(payload[12:16]))) / 1000
		relativeAlt := float64(int32(binary.LittleEndian.Uint32(payload[16:20]))) / 1000
		vx := float64(int16(binary.LittleEndian.Uint16(payload[20:22]))) / 100
		vy := float64(int16(binary.LittleEndian.Uint16(payload[22:24]))) / 100
		vz := float64(int16(binary.LittleEndian.Uint16(payload[24:26]))) / 100
		heading := float64(binary.LittleEndian.Uint16(payload[26:28])) / 100

		_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
			UAVID:           uavID,
			Latitude:        lat,
			Longitude:       lng,
			Altitude:        alt,
			RelativeAltitude: relativeAlt,
			Vx:              vx,
			Vy:              vy,
			Vz:              vz,
			Heading:         heading,
			GroundSpeed:     math.Sqrt(vx*vx + vy*vy),
			Timestamp:       time.Now(),
		})
	}
}

func (m *CommandManager) processRadioStatus(uavID uint64, payload []byte) {
	if len(payload) >= 9 {
		rssi := int8(payload[0])
		remoteRSSI := int8(payload[1])
		txbuf := uint8(payload[5])
		noise := int8(payload[6])
		remoteNoise := int8(payload[7])

		signalStrength := 100 + int(rssi)
		if signalStrength > 100 {
			signalStrength = 100
		}
		if signalStrength < 0 {
			signalStrength = 0
		}

		_ = m.flightService.ProcessTelemetry(&service.TelemetryData{
			UAVID:          uavID,
			SignalStrength: float64(signalStrength),
			RSSI:           int(rssi),
			RemoteRSSI:     int(remoteRSSI),
			Timestamp:      time.Now(),
		})
	}
}

func (m *CommandManager) processMissionItemReached(uavID uint64, payload []byte) {
	if len(payload) >= 2 {
		seq := binary.LittleEndian.Uint16(payload[0:2])
		_ = service.NewMissionService().UpdateWaypointProgress(uavID, int(seq))
	}
}

func (m *CommandManager) processCommandAck(uavID uint64, payload []byte) {
	if len(payload) >= 3 {
		command := binary.LittleEndian.Uint16(payload[0:2])
		result := uint8(payload[2])

		success := result == 0
		message := getCommandResultMessage(result)

		websocket.BroadcastCommandResponse(uavID, fmt.Sprintf("CMD_%d", command), success, message)
	}
}

func (m *CommandManager) SendCommand(uavID uint64, data []byte) error {
	m.connMu.RLock()
	conn, exists := m.uavConns[uavID]
	m.connMu.RUnlock()

	if !exists {
		return errors.New("uav not connected")
	}

	_, err := conn.Write(data)
	return err
}

func (m *CommandManager) SendCustomCommand(uavID uint64, commandName string, params map[string]interface{}) error {
	encoder := NewMAVLinkEncoder()
	data := encoder.EncodeCommandLong(0, uavID, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	return m.SendCommand(uavID, data)
}

func (m *CommandManager) generateUAVID(addr string) uint64 {
	hash := uint64(0)
	for _, b := range addr {
		hash = hash*31 + uint64(b)
	}
	return hash % 1000000
}

func (m *CommandManager) Stop() {
	if m.listenerTCP != nil {
		m.listenerTCP.Close()
	}
	if m.listenerUDP != nil {
		m.listenerUDP.Close()
	}
	m.parser.Close()
}

func getCommandResultMessage(result uint8) string {
	switch result {
	case 0:
		return "成功"
	case 1:
		return "临时失败"
	case 2:
		return "永久失败"
	case 3:
		return "不支持"
	case 4:
		return "拒绝"
	case 5:
		return "执行中"
	default:
		return fmt.Sprintf("未知错误 (%d)", result)
	}
}
