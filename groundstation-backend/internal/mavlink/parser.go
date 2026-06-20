package mavlink

import (
	"encoding/binary"
	"errors"
	"groundstation-backend/internal/models"
	"math"
	"time"
)

const (
	MAVLINK_STX = 0xFD

	HEARTBEAT        = 0
	SYS_STATUS       = 1
	SYSTEM_TIME      = 2
	GPS_RAW_INT      = 24
	GLOBAL_POSITION_INT = 33
	ATTITUDE         = 30
	RAW_IMU          = 27
	BATTERY_STATUS   = 147
	RADIO_STATUS     = 109
	COMMAND_LONG     = 76
	COMMAND_ACK      = 77
	MISSION_COUNT    = 44
	MISSION_ITEM     = 39
	MISSION_CURRENT  = 42
	MISSION_ITEM_REACHED = 46
)

const (
	CMD_NAV_WAYPOINT                      = 16
	CMD_NAV_LOITER_UNLIM                  = 17
	CMD_NAV_LOITER_TURNS                  = 18
	CMD_NAV_LOITER_TIME                   = 19
	CMD_NAV_RETURN_TO_LAUNCH              = 20
	CMD_NAV_LAND                          = 21
	CMD_NAV_TAKEOFF                       = 22
	CMD_DO_JUMP                           = 177
	CMD_DO_CHANGE_SPEED                   = 178
	CMD_DO_SET_HOME                       = 179
	CMD_DO_SET_PARAMETER                  = 23
	CMD_DO_SET_MODE                       = 176
	CMD_COMPONENT_ARM_DISARM              = 400
	CMD_MISSION_START                     = 300
	CMD_MISSION_CLEAR_ALL                 = 45
	CMD_MISSION_SET_CURRENT               = 41
	CMD_DO_SET_POSITION_TARGET_LOCAL_NED  = 84
	CMD_CONDITION_YAW                     = 115
	CMD_PREFLIGHT_REBOOT_SHUTDOWN         = 246
	CMD_PREFLIGHT_CALIBRATION             = 241
	CMD_DO_DIGICAM_CONTROL                = 203
)

type MAVLinkMessage struct {
	STX         byte
	Length      byte
	IncompatFlags byte
	CompatFlags byte
	Seq         byte
	SystemID    byte
	ComponentID byte
	MessageID   uint32
	Payload     []byte
	Checksum    uint16
}

type MAVLinkParser struct {
	buffer       []byte
	messageChan  chan *MAVLinkMessage
	parsing      bool
}

func NewMAVLinkParser() *MAVLinkParser {
	return &MAVLinkParser{
		buffer:      make([]byte, 0),
		messageChan: make(chan *MAVLinkMessage, 1024),
		parsing:     true,
	}
}

func (p *MAVLinkParser) Parse(data []byte) error {
	p.buffer = append(p.buffer, data...)

	for len(p.buffer) >= 12 {
		if p.buffer[0] != MAVLINK_STX {
			p.buffer = p.buffer[1:]
			continue
		}

		length := int(p.buffer[1])
		totalLen := 12 + length + 2

		if len(p.buffer) < totalLen {
			break
		}

		msg := &MAVLinkMessage{
			STX:         p.buffer[0],
			Length:      p.buffer[1],
			IncompatFlags: p.buffer[2],
			CompatFlags: p.buffer[3],
			Seq:         p.buffer[4],
			SystemID:    p.buffer[5],
			ComponentID: p.buffer[6],
			MessageID:   uint32(p.buffer[7]) | uint32(p.buffer[8])<<8 | uint32(p.buffer[9])<<16,
			Payload:     p.buffer[12 : 12+length],
			Checksum:    binary.LittleEndian.Uint16(p.buffer[12+length : 12+length+2]),
		}

		if verifyChecksum(msg) {
			p.messageChan <- msg
		}

		p.buffer = p.buffer[totalLen:]
	}

	return nil
}

func verifyChecksum(msg *MAVLinkMessage) bool {
	return true
}

func (p *MAVLinkParser) GetMessage() (*MAVLinkMessage, bool) {
	select {
	case msg := <-p.messageChan:
		return msg, true
	default:
		return nil, false
	}
}

func (p *MAVLinkParser) Close() {
	p.parsing = false
	close(p.messageChan)
}

func ParseHeartbeat(payload []byte) (*models.HeartbeatData, error) {
	if len(payload) < 9 {
		return nil, errors.New("payload too short")
	}

	return &models.HeartbeatData{
		Type:           uint8(payload[0]),
		Autopilot:      uint8(payload[1]),
		BaseMode:       uint8(payload[2]),
		CustomMode:     binary.LittleEndian.Uint32(payload[3:7]),
		SystemStatus:   uint8(payload[7]),
		MavlinkVersion: uint8(payload[8]),
	}, nil
}

func ParseGPSRawInt(payload []byte) (*models.GPSData, error) {
	if len(payload) < 36 {
		return nil, errors.New("payload too short")
	}

	return &models.GPSData{
		TimeUsec:     binary.LittleEndian.Uint64(payload[0:8]),
		FixType:      uint8(payload[8]),
		Satellites:   uint8(payload[9]),
		Latitude:     float64(int32(binary.LittleEndian.Uint32(payload[12:16]))) / 1e7,
		Longitude:    float64(int32(binary.LittleEndian.Uint32(payload[16:20]))) / 1e7,
		Altitude:     float64(int32(binary.LittleEndian.Uint32(payload[20:24]))) / 1000,
		EPH:          float64(binary.LittleEndian.Uint16(payload[24:26])) / 100,
		EPV:          float64(binary.LittleEndian.Uint16(payload[26:28])) / 100,
		Vel:          float64(binary.LittleEndian.Uint16(payload[28:30])) / 100,
		COG:          float64(binary.LittleEndian.Uint16(payload[30:32])) / 100,
	}, nil
}

func ParseAttitude(payload []byte) (*models.AttitudeData, error) {
	if len(payload) < 28 {
		return nil, errors.New("payload too short")
	}

	return &models.AttitudeData{
		TimeUsec:  binary.LittleEndian.Uint64(payload[0:8]),
		Roll:      float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[8:12]))),
		Pitch:     float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[12:16]))),
		Yaw:       float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[16:20]))),
		RollSpeed: float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[20:24]))),
		PitchSpeed: float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[24:28]))),
		YawSpeed:  float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[28:32]))),
	}, nil
}

func ParseBatteryStatus(payload []byte) (*models.BatteryData, error) {
	if len(payload) < 34 {
		return nil, errors.New("payload too short")
	}

	return &models.BatteryData{
		Voltage:    float64(binary.LittleEndian.Uint16(payload[4:6])) / 1000,
		Current:    float64(int16(binary.LittleEndian.Uint16(payload[6:8]))) / 100,
		Remaining:  int8(payload[10]),
	}, nil
}

func ParseSystemTime(payload []byte) (*models.SystemTimeData, error) {
	if len(payload) < 12 {
		return nil, errors.New("payload too short")
	}

	return &models.SystemTimeData{
		UnixUsec:    binary.LittleEndian.Uint64(payload[0:8]),
		TimeBootMs:  binary.LittleEndian.Uint32(payload[8:12]),
	}, nil
}

func GetFlightModeCode(mode string) uint32 {
	modeMap := map[string]uint32{
		"STABILIZE": 0,
		"ACRO":      1,
		"ALT_HOLD":  2,
		"AUTO":      3,
		"GUIDED":    4,
		"LOITER":    5,
		"RTL":       6,
		"CIRCLE":    7,
		"LAND":      9,
		"DRIFT":     11,
		"SPORT":     13,
		"FLIP":      14,
		"AUTOTUNE":  15,
		"POSHOLD":   16,
		"BRAKE":     17,
		"THROW":     18,
		"AVOID_ADSB": 19,
		"GUIDED_NOGPS": 20,
	}
	return modeMap[mode]
}

func GetFlightModeName(code uint32) string {
	modeMap := map[uint32]string{
		0: "STABILIZE",
		1: "ACRO",
		2: "ALT_HOLD",
		3: "AUTO",
		4: "GUIDED",
		5: "LOITER",
		6: "RTL",
		7: "CIRCLE",
		9: "LAND",
		11: "DRIFT",
		13: "SPORT",
		14: "FLIP",
		15: "AUTOTUNE",
		16: "POSHOLD",
		17: "BRAKE",
		18: "THROW",
		19: "AVOID_ADSB",
		20: "GUIDED_NOGPS",
	}
	if name, ok := modeMap[code]; ok {
		return name
	}
	return "UNKNOWN"
}
