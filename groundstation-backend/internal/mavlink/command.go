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

const (
	MAV_COMP_ID_TELEMETRY_RADIO = 68
	MAV_COMP_ID_UDP_BRIDGE      = 240
)

type linkStatusSnapshot struct {
	ActiveLink     uint8
	RadioRSSI      int8
	RadioConnected bool
	LteRSSI        int8
	LteConnected   bool
	LteNetworkType string
	PacketLoss     float64
	LatencyMs      uint32
	LastUpdate     time.Time
}

type CommandManager struct {
	uavConns        map[uint64]net.Conn
	connMu          sync.RWMutex
	parser          *MAVLinkParser
	flightService   *service.FlightService
	heartbeatMgr    *HeartbeatManager
	listenerTCP     net.Listener
	listenerUDP     *net.UDPConn
	linkStatusCache map[uint64]*linkStatusSnapshot
	linkStatusMu    sync.Mutex
}

var commandManager *CommandManager
var commandOnce sync.Once

func NewCommandManager() *CommandManager {
	commandOnce.Do(func() {
		commandManager = &CommandManager{
			uavConns:        make(map[uint64]net.Conn),
			parser:          NewMAVLinkParser(),
			flightService:   service.NewFlightService(),
			heartbeatMgr:    NewHeartbeatManager(),
			linkStatusCache: make(map[uint64]*linkStatusSnapshot),
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
		m.processRadioStatus(uavID, msg.ComponentID, msg.Payload)

	case MISSION_ITEM_REACHED:
		m.processMissionItemReached(uavID, msg.Payload)

	case COMMAND_ACK:
		m.processCommandAck(uavID, msg.Payload)

	case CAMERA_STATUS:
		m.processCameraStatus(uavID, msg.Payload)

	case CAMERA_FEEDBACK:
		m.processCameraFeedback(uavID, msg.Payload)

	case CAMERA_SETTINGS:
		m.processCameraSettings(uavID, msg.Payload)

	case PAYLOAD_STATUS:
		m.processPayloadStatus(uavID, msg.Payload)

	case NAMED_VALUE_FLOAT:
		m.processNamedValueFloat(uavID, msg.Payload)

	case NAMED_VALUE_INT:
		m.processNamedValueInt(uavID, msg.Payload)

	case VIDEO_STREAM_INFORMATION:
		m.processVideoStreamInfo(uavID, msg.Payload)

	case ESC_STATUS:
		m.processESCStatus(uavID, msg.Payload)

	case ESC_INFO:
		m.processESCInfo(uavID, msg.Payload)

	case LINK_STATUS:
		m.processLinkStatus(uavID, msg.Payload)

	case OBSTACLE_AVOIDANCE_EVENT:
		m.processObstacleAvoidanceEvent(uavID, msg.Payload)

	case OBSTACLE_AVOIDANCE_STATUS:
		m.processObstacleAvoidanceStatus(uavID, msg.Payload)

	case OBSTACLE_HEATMAP_UPDATE:
		m.processObstacleHeatmapUpdate(uavID, msg.Payload)

	case THRUST_LEARNING_STATUS:
		m.processThrustLearningStatus(uavID, msg.Payload)

	case THRUST_CURVE_DATA:
		m.processThrustCurveData(uavID, msg.Payload)

	case PID_GAINS_REPORT:
		m.processPIDGainsReport(uavID, msg.Payload)

	case THRUST_SAMPLE:
		m.processThrustSample(uavID, msg.Payload)
	}

	_ = nsq.Publish(nsq.TopicMAVLinkMessage, map[string]interface{}{
		"uav_id":     uavID,
		"message_id": msg.MessageID,
		"length":     msg.Length,
		"timestamp":  time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processCameraStatus(uavID uint64, payload []byte) {
	camStatus, err := ParseCameraStatus(payload)
	if err != nil {
		return
	}

	payloadService := service.NewPayloadService()

	payloadID := uint64(camStatus.TargetComponent)
	if payloadID == 0 {
		payloadID = uavID * 10 + 1
	}

	storageFreeMB := float64(camStatus.StorageFreeKB) / 1024
	storageTotalMB := float64(camStatus.StorageTotalKB) / 1024
	lensTemp := float64(camStatus.LensTemperature)

	mode := models.CameraModeIdle
	imgStatus := camStatus.ImgStatus
	if imgStatus&0x02 != 0 {
		mode = models.CameraModeVideo
	} else if imgStatus&0x01 != 0 {
		mode = models.CameraModePhoto
	}

	recordingTime := 0
	if camStatus.VideoOn > 0 {
		recordingTime = int(camStatus.VideoTimeSec)
	}

	cameraStatus := &models.CameraStatus{
		PayloadID:           payloadID,
		Mode:                mode,
		Recording:           camStatus.VideoOn > 0,
		RecordingTimeSec:    recordingTime,
		PhotoCount:          int(camStatus.ImgStatus >> 4),
		StorageFreeMB:       storageFreeMB,
		StorageTotalMB:      storageTotalMB,
		StorageUsedPercent:  0,
		LensTemperatureC:    lensTemp,
		SensorTemperatureC:  float64(camStatus.SensorTemp),
		ZoomLevel:           float64(camStatus.ZoomLevel),
		FocusLevel:          float64(camStatus.FocusLevel),
		ISO:                 int(camStatus.ISO),
		ShutterSpeedMs:      float64(camStatus.ShutterSpeedMs),
		BatteryPercent:      int(camStatus.BatteryPct),
	}
	if cameraStatus.StorageTotalMB > 0 {
		cameraStatus.StorageUsedPercent = int((cameraStatus.StorageTotalMB - cameraStatus.StorageFreeMB) / cameraStatus.StorageTotalMB * 100)
	}

	_ = payloadService.UpdateCameraStatus(payloadID, cameraStatus)
	websocket.BroadcastCameraStatus(uavID, payloadID, cameraStatus)

	device, _ := payloadService.GetPayload(payloadID)
	if device != nil {
		websocket.BroadcastPayloadStatus(uavID, payloadID, device)
	}
}

func (m *CommandManager) processCameraFeedback(uavID uint64, payload []byte) {
	fb, err := ParseCameraFeedback(payload)
	if err != nil {
		return
	}

	payloadID := uint64(fb.CameraID)
	if payloadID == 0 {
		payloadID = uavID * 10 + 1
	}

	payloadService := service.NewPayloadService()
	_ = payloadService.IncrementPhotoCount(payloadID)

	result := "success"
	if fb.CaptureResult != 0 {
		result = "failed"
	}
	websocket.BroadcastCameraFeedback(uavID, payloadID, result, int(fb.ImageSeq), int(fb.SizeKB))
}

func (m *CommandManager) processCameraSettings(uavID uint64, payload []byte) {
	if len(payload) < 40 {
		return
	}

	payloadID := uint64(payload[8])
	if payloadID == 0 {
		payloadID = uavID * 10 + 1
	}

	mode := models.CameraModeIdle
	modeByte := uint8(payload[10])
	if modeByte == 1 {
		mode = models.CameraModePhoto
	} else if modeByte == 2 {
		mode = models.CameraModeVideo
	}

	payloadService := service.NewPayloadService()
	status, _ := payloadService.GetCameraStatus(payloadID)
	if status == nil {
		status = &models.CameraStatus{PayloadID: payloadID}
	}
	status.Mode = mode
	_ = payloadService.UpdateCameraStatus(payloadID, status)
	websocket.BroadcastCameraStatus(uavID, payloadID, status)
}

func (m *CommandManager) processPayloadStatus(uavID uint64, payload []byte) {
	ps, err := ParsePayloadStatus(payload)
	if err != nil {
		return
	}

	payloadService := service.NewPayloadService()
	payloadID := uint64(ps.PayloadID)
	if payloadID == 0 {
		payloadID = uavID*10 + uint64(ps.PayloadType+1)
	}

	statusCode := "offline"
	switch ps.Status {
	case 1:
		statusCode = "online"
	case 2:
		statusCode = "active"
	case 3:
		statusCode = "error"
	}

	payloadType := ""
	switch ps.PayloadType {
	case 1:
		payloadType = string(models.PayloadTypeCamera)
	case 2:
		payloadType = string(models.PayloadTypeThermalCamera)
	case 3:
		payloadType = string(models.PayloadTypeSpeaker)
	case 4:
		payloadType = string(models.PayloadTypeSprayer)
	}

	device, _ := payloadService.GetPayload(payloadID)
	if device == nil {
		device = &models.PayloadDevice{
			ID:          payloadID,
			UAVID:       uavID,
			Name:        fmt.Sprintf("Payload-%d", payloadID),
			Type:        models.PayloadType(payloadType),
			Status:      models.PayloadStatus(statusCode),
			Temperature: float64(ps.Temperature),
			Battery:     int(ps.BatteryPct),
		}
		_ = payloadService.CreatePayload(device)
	} else {
		device.Status = models.PayloadStatus(statusCode)
		device.Temperature = float64(ps.Temperature)
		device.Battery = int(ps.BatteryPct)
		_ = payloadService.UpdatePayload(device)
	}

	if payloadType == string(models.PayloadTypeSprayer) {
		sprayerStatus := &models.SprayerStatus{
			PayloadID:       payloadID,
			FlowRate:        float64(ps.FlowRate),
			RemainingVolume: float64(ps.RemainingQty),
			TotalCapacity:   float64(ps.TotalCapacity),
			Pressure:        float64(ps.Pressure),
			Spraying:        ps.SubStatus == 1,
		}
		if sprayerStatus.TotalCapacity > 0 {
			sprayerStatus.RemainingPercent = int(sprayerStatus.RemainingVolume / sprayerStatus.TotalCapacity * 100)
		}
		_ = payloadService.UpdateSprayerStatus(payloadID, sprayerStatus)
		websocket.BroadcastSprayerStatus(uavID, payloadID, sprayerStatus)
	}

	websocket.BroadcastPayloadStatus(uavID, payloadID, device)
}

func (m *CommandManager) processNamedValueFloat(uavID uint64, payload []byte) {
	nv, err := ParseNamedValueFloat(payload)
	if err != nil {
		return
	}

	payloadService := service.NewPayloadService()

	switch nv.Name {
	case "spray_flow":
		_ = payloadService.UpdateSprayerFlowRate(uavID*10+4, float64(nv.Value))
	case "spray_pres":
		_ = payloadService.UpdateSprayerPressure(uavID*10+4, float64(nv.Value))
	case "lens_temp":
		_ = payloadService.UpdateCameraLensTemp(uavID*10+1, float64(nv.Value))
	case "cam_zoom":
		_ = payloadService.UpdateCameraZoom(uavID*10+1, float64(nv.Value))
	}

	_ = nsq.Publish(nsq.TopicTelemetryData, map[string]interface{}{
		"uav_id":      uavID,
		"name":        nv.Name,
		"value":       nv.Value,
		"time_boot_ms": nv.TimeBootMs,
		"timestamp":   time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processNamedValueInt(uavID uint64, payload []byte) {
	nv, err := ParseNamedValueInt(payload)
	if err != nil {
		return
	}

	m.linkStatusMu.Lock()
	cache, exists := m.linkStatusCache[uavID]
	if !exists {
		cache = &linkStatusSnapshot{}
		m.linkStatusCache[uavID] = cache
	}

	updated := false

	switch nv.Name {
	case "link_active":
		cache.ActiveLink = uint8(nv.Value)
		updated = true
	case "lte_nettype":
		switch nv.Value {
		case 0:
			cache.LteNetworkType = "none"
		case 1:
			cache.LteNetworkType = "2G"
		case 2:
			cache.LteNetworkType = "3G"
		case 3:
			cache.LteNetworkType = "4G"
		case 4:
			cache.LteNetworkType = "5G"
		}
		updated = true
	}

	if updated {
		cache.LastUpdate = time.Now()
	}
	m.linkStatusMu.Unlock()

	if updated {
		m.updateLinkStatusIfChanged(uavID)
	}

	_ = nsq.Publish(nsq.TopicTelemetryData, map[string]interface{}{
		"uav_id":       uavID,
		"name":         nv.Name,
		"value":        nv.Value,
		"time_boot_ms": nv.TimeBootMs,
		"timestamp":    time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processVideoStreamInfo(uavID uint64, payload []byte) {
	if len(payload) < 20 {
		return
	}

	frameRate := int(payload[17])
	resolutionH := int(payload[18])
	resolutionV := int(payload[19])

	payloadID := uavID*10 + 1
	payloadService := service.NewPayloadService()
	status, _ := payloadService.GetCameraStatus(payloadID)
	if status != nil {
		status.FrameRate = frameRate
		status.Resolution = fmt.Sprintf("%dx%d", resolutionH, resolutionV)
		_ = payloadService.UpdateCameraStatus(payloadID, status)
		websocket.BroadcastCameraStatus(uavID, payloadID, status)
	}
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

func (m *CommandManager) processRadioStatus(uavID uint64, componentID uint8, payload []byte) {
	if len(payload) < 9 {
		return
	}

	rssi := int8(payload[0])
	fixed := binary.LittleEndian.Uint16(payload[7:9])
	packetLoss := float64(fixed) / 10.0

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
		Timestamp:      time.Now(),
	})

	m.linkStatusMu.Lock()
	cache, exists := m.linkStatusCache[uavID]
	if !exists {
		cache = &linkStatusSnapshot{}
		m.linkStatusCache[uavID] = cache
	}

	connected := rssi > -100

	switch componentID {
	case MAV_COMP_ID_TELEMETRY_RADIO:
		cache.RadioRSSI = rssi
		cache.RadioConnected = connected
	case MAV_COMP_ID_UDP_BRIDGE:
		cache.LteRSSI = rssi
		cache.LteConnected = connected
	default:
		cache.RadioRSSI = rssi
		cache.RadioConnected = connected
	}

	if packetLoss > 0 {
		cache.PacketLoss = packetLoss
	}

	cache.LastUpdate = time.Now()
	m.linkStatusMu.Unlock()

	m.updateLinkStatusIfChanged(uavID)
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
	var cmd uint16
	var p [7]float32

	switch commandName {
	case "video_start":
		cmd = 50001
		p[0] = floatParam(params, "param1")
		p[1] = floatParam(params, "param2")
		p[2] = floatParam(params, "param3")
		p[3] = floatParam(params, "param4")
		p[4] = 1
	case "video_stop":
		cmd = 50002
	case "video_quality_adjust":
		cmd = 50003
		p[0] = floatParam(params, "bitrate_kbps")
	case "set_primary_link":
		cmd = 50010
		if lt, ok := params["link_type"]; ok {
			switch v := lt.(type) {
			case models.LinkType:
				p[0] = float32(v)
			case float64:
				p[0] = float32(v)
			case int:
				p[0] = float32(v)
			}
		}
	case "set_mode":
		cmd = CMD_DO_SET_MODE
		p[0] = 1
		modeStr := "AUTO"
		if v, ok := params["mode"].(string); ok {
			modeStr = v
		}
		p[1] = float32(GetFlightModeCode(modeStr))
	case "velocity":
		cmd = CMD_DO_SET_POSITION_TARGET_LOCAL_NED
		p[0] = floatParam(params, "vx")
		p[1] = floatParam(params, "vy")
		p[2] = floatParam(params, "vz")
		p[3] = floatParam(params, "yaw_rate")
	case "wind_adapt_speed":
		cmd = 50020
		p[0] = floatParam(params, "param1")
		p[1] = floatParam(params, "param2")
		p[2] = floatParam(params, "param3")
	case "gust_protection":
		cmd = 50021
		p[0] = floatParam(params, "param1")
		p[1] = floatParam(params, "param2")
		p[2] = floatParam(params, "param3")
	case "collision_avoid_speed":
		cmd = 50030
		p[0] = floatParam(params, "param1")
		p[1] = floatParam(params, "param2")
		p[2] = floatParam(params, "param3")
	default:
		cmd = 50000
		p[0] = floatParam(params, "param1")
		p[1] = floatParam(params, "param2")
		p[2] = floatParam(params, "param3")
		p[3] = floatParam(params, "param4")
		p[4] = floatParam(params, "param5")
		p[5] = floatParam(params, "param6")
		p[6] = floatParam(params, "param7")
	}

	data := EncodeCommandLong(uavID, cmd, p[0], p[1], p[2], p[3], p[4], p[5], p[6])
	return m.SendCommand(uavID, data)
}

func floatParam(params map[string]interface{}, key string) float32 {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case float64:
			return float32(val)
		case float32:
			return val
		case int:
			return float32(val)
		case int64:
			return float32(val)
		}
	}
	return 0
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

func (m *CommandManager) processESCStatus(uavID uint64, payload []byte) {
	esc, err := ParseESCStatus(payload)
	if err != nil {
		return
	}

	motorService := service.NewMotorFailureService()
	motorStatus := &models.MotorStatus{
		UAVID:       uavID,
		MotorIndex:  int(esc.Index),
		RPM:         int(esc.RPM),
		Voltage:     float64(esc.Voltage),
		Current:     float64(esc.Current),
		Temperature: int(esc.Temperature),
		Throttle:    float64(esc.Throttle),
		FaultFlags:  int(esc.FaultFlags),
		ErrorCount:  int(esc.ErrorCount),
	}

	if err := motorService.UpdateMotorStatus(uavID, motorStatus); err != nil {
		return
	}

	failureDetected, statusOverride := motorService.DetectMotorFailure(uavID, int(esc.Index), motorStatus)
	if statusOverride != "" {
		motorStatus.Status = statusOverride
	}
	websocket.BroadcastMotorStatus(uavID, motorStatus)

	if failureDetected {
		alert, _ := motorService.CreateMotorFailureAlert(uavID, int(esc.Index), motorStatus)
		if alert != nil {
			websocket.BroadcastAlert(alert)
		}
		websocket.BroadcastMotorFailure(uavID, int(esc.Index), motorStatus)

		motorService.RecalculateMotorMixing(uavID)
		motorService.AdjustPIDParameters(uavID)

		go motorService.TriggerEmergencyRTH(uavID, int(esc.Index))
	}
}

func (m *CommandManager) processESCInfo(uavID uint64, payload []byte) {
	escInfo, err := ParseESCInfo(payload)
	if err != nil {
		return
	}

	motorService := service.NewMotorFailureService()
	_ = motorService.UpdateMotorInfo(uavID, int(escInfo.Index), escInfo.Vendor, escInfo.Model, int(escInfo.FaultFlags), int(escInfo.ErrorCode))
}

func (m *CommandManager) updateLinkStatusIfChanged(uavID uint64) {
	m.linkStatusMu.Lock()
	cache, exists := m.linkStatusCache[uavID]
	if !exists {
		m.linkStatusMu.Unlock()
		return
	}

	hasRadioRSSI := cache.RadioRSSI != 0 || cache.RadioConnected
	hasLteRSSI := cache.LteRSSI != 0 || cache.LteConnected

	if !hasRadioRSSI && !hasLteRSSI {
		m.linkStatusMu.Unlock()
		return
	}

	report := &service.LinkStatusReport{
		ActiveLink:     cache.ActiveLink,
		RadioRSSI:      cache.RadioRSSI,
		RadioConnected: cache.RadioConnected,
		LteRSSI:        cache.LteRSSI,
		LteConnected:   cache.LteConnected,
		LteNetworkType: cache.LteNetworkType,
		PacketLoss:     cache.PacketLoss,
		LatencyMs:      cache.LatencyMs,
	}
	m.linkStatusMu.Unlock()

	_, _ = service.NewLinkService().ReportStatus(uavID, report)
}

func (m *CommandManager) processLinkStatus(uavID uint64, payload []byte) {
	linkStatus, err := ParseLinkStatus(payload)
	if err != nil {
		return
	}

	linkService := service.NewLinkService()
	report := &service.LinkStatusReport{
		ActiveLink:     linkStatus.ActiveLink,
		RadioRSSI:      linkStatus.RadioRSSI,
		RadioConnected: linkStatus.RadioConnected != 0,
		LteRSSI:        linkStatus.LteRSSI,
		LteConnected:   linkStatus.LteConnected != 0,
		LteNetworkType: linkStatus.LteNetworkType,
		PacketLoss:     float64(linkStatus.PacketLoss),
		LatencyMs:      linkStatus.LatencyMs,
	}
	_, _ = linkService.ReportStatus(uavID, report)
}

func (m *CommandManager) processObstacleAvoidanceEvent(uavID uint64, payload []byte) {
	event, err := ParseObstacleAvoidanceEvent(payload)
	if err != nil {
		return
	}

	oaService := service.NewObstacleAvoidanceService()

	strategy := models.AvoidanceStrategy("hover")
	switch event.Strategy {
	case 1:
		strategy = models.StrategyAscendBypass
	case 2:
		strategy = models.StrategyRetreatBypass
	}

	status := models.AvoidanceStatusTriggered
	switch event.Status {
	case 3:
		status = models.AvoidanceStatusAvoiding
	case 4:
		status = models.AvoidanceStatusBypassing
	case 5:
		status = models.AvoidanceStatusCompleted
	case 6:
		status = models.AvoidanceStatusFailed
	}

	sensorType := models.SensorTypeMillimeterWave
	switch event.SensorType {
	case 1:
		sensorType = models.SensorTypeStereoVision
	case 2:
		sensorType = models.SensorTypeLidar
	case 3:
		sensorType = models.SensorTypeUltrasonic
	}

	direction := models.DirectionFront
	switch event.Direction {
	case 1:
		direction = models.DirectionLeft
	case 2:
		direction = models.DirectionRight
	case 3:
		direction = models.DirectionTop
	case 4:
		direction = models.DirectionBottom
	case 5:
		direction = models.DirectionRear
	}

	_, _ = oaService.CreateAvoidanceEvent(uavID, strategy, 0, sensorType, direction,
		float64(event.ObstacleDistance), float64(event.StartLat), float64(event.StartLng), float64(event.StartAlt), nil)

	websocket.BroadcastObstacleAvoidanceEvent(uavID, map[string]interface{}{
		"uav_id":      uavID,
		"uavId":       uavID,
		"event_id":    event.EventID,
		"strategy":    strategy,
		"status":      status,
		"sensor_type": sensorType,
		"direction":   direction,
		"distance":    float64(event.ObstacleDistance),
		"start_lat":   float64(event.StartLat),
		"start_lng":   float64(event.StartLng),
		"start_alt":   float64(event.StartAlt),
		"bypass_path_count": event.BypassPathCount,
		"timestamp":   time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processObstacleAvoidanceStatus(uavID uint64, payload []byte) {
	status, err := ParseObstacleAvoidanceStatus(payload)
	if err != nil {
		return
	}

	websocket.BroadcastObstacleAvoidanceStatus(uavID, map[string]interface{}{
		"uav_id":           uavID,
		"uavId":            uavID,
		"enabled":          status.Enabled == 1,
		"sensitivity":      status.Sensitivity,
		"strategy":         status.Strategy,
		"detection_range":  float64(status.DetectionRange),
		"total_detections": status.TotalDetections,
		"total_events":     status.TotalEvents,
		"timestamp":        time.Now().UnixNano() / 1e6,
	})
}

func (m *CommandManager) processObstacleHeatmapUpdate(uavID uint64, payload []byte) {
	heatmap, err := ParseObstacleHeatmapUpdate(payload)
	if err != nil {
		return
	}

	oaService := service.NewObstacleAvoidanceService()

	points := make([]map[string]interface{}, 0, heatmap.PointCount)
	for i := 0; i < int(heatmap.PointCount) && i < 8; i++ {
		p := heatmap.Points[i]
		_ = oaService.RecordDetection(uavID, &models.ObstacleDetectionLog{
			UAVID:       uavID,
			SensorType:  models.SensorTypeMillimeterWave,
			Direction:   models.DirectionFront,
			Distance:    float64(p.MinDistance),
			Latitude:    float64(p.Latitude),
			Longitude:   float64(p.Longitude),
			Altitude:    float64(p.Altitude),
		})

		points = append(points, map[string]interface{}{
			"lat":            float64(p.Latitude),
			"lng":            float64(p.Longitude),
			"alt":            float64(p.Altitude),
			"trigger_count":  p.TriggerCount,
			"min_distance":   float64(p.MinDistance),
			"intensity":      float64(p.TriggerCount) / 20.0,
			"avg_distance":   float64(p.MinDistance),
		})
	}

	websocket.BroadcastObstacleHeatmapUpdate(uavID, map[string]interface{}{
		"uav_id":    uavID,
		"uavId":     uavID,
		"points":    points,
		"timestamp": time.Now().UnixNano() / 1e6,
	})
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

func (m *CommandManager) processThrustLearningStatus(uavID uint64, payload []byte) {
	state, weight, hoverThrottle, sampleCount, progress, err := ParseThrustLearningStatus(payload)
	if err != nil {
		return
	}

	tlService := service.NewThrustLearningService()
	_ = tlService.UpdateStatusFromMAVLink(uavID, state, weight, hoverThrottle, sampleCount, progress)

	status, _ := tlService.GetStatus(uavID)
	websocket.BroadcastThrustLearningStatus(uavID, status)
}

func (m *CommandManager) processThrustCurveData(uavID uint64, payload []byte) {
	_, count, points, err := ParseThrustCurveData(payload)
	if err != nil {
		return
	}

	tlService := service.NewThrustLearningService()

	curvePoints := make([]struct {
		Throttle float64
		Thrust   float64
		Rpm      float64
	}, count)
	for i, p := range points {
		curvePoints[i].Throttle = p.Throttle
		curvePoints[i].Thrust = p.Thrust
		curvePoints[i].Rpm = 0
	}

	_ = tlService.StoreCurvePoints(uavID, curvePoints)

	updatedPoints, _ := tlService.GetThrustCurve(uavID)
	websocket.BroadcastThrustCurveUpdate(uavID, updatedPoints)
}

func (m *CommandManager) processPIDGainsReport(uavID uint64, payload []byte) {
	gains, err := ParsePIDGainsReport(payload)
	if err != nil {
		return
	}

	tlService := service.NewThrustLearningService()
	_ = tlService.StorePIDGains(uavID, gains)

	profile, _ := tlService.GetPIDGains(uavID)
	websocket.BroadcastPIDGainsUpdate(uavID, profile)
}

func (m *CommandManager) processThrustSample(uavID uint64, payload []byte) {
	sampleData, err := ParseThrustSample(payload)
	if err != nil {
		return
	}

	tlService := service.NewThrustLearningService()
	_ = tlService.StoreSample(uavID, sampleData)

	wsData := map[string]interface{}{
		"uav_id":    uavID,
		"uavId":     uavID,
		"id":        time.Now().UnixNano() / 1e6,
		"throttle":  float64(sampleData.Throttle),
		"accel_z":   float64(sampleData.AccelZ),
		"altitude":  float64(sampleData.Altitude),
		"vz":        float64(sampleData.VZ),
		"motor_pwm": []uint16{sampleData.MotorPWM1, sampleData.MotorPWM2, sampleData.MotorPWM3, sampleData.MotorPWM4},
		"voltage":   float64(sampleData.Voltage),
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	websocket.BroadcastThrustLearningSample(uavID, wsData)
}
