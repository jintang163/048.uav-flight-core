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
	CAMERA_STATUS    = 179
	CAMERA_FEEDBACK  = 180
	CAMERA_SETTINGS  = 178
	VIDEO_STREAM_INFORMATION = 269
	NAMED_VALUE_FLOAT = 251
	NAMED_VALUE_INT   = 252
	PAYLOAD_STATUS    = 380
	PAYLOAD_TELEMETRY = 381
	ESC_STATUS        = 291
	ESC_INFO          = 290
	HIGHRES_IMU       = 105
	LINK_STATUS       = 390
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
	CMD_DO_DIGICAM_CONFIGURE              = 202
	CMD_DO_VIDEO_START                    = 2000
	CMD_DO_VIDEO_STOP                     = 2001
	CMD_DO_SPRAYER                        = 2002
	CMD_DO_PLAY_TUNE                      = 4003
	CMD_DO_LOITER_UNLIMITED               = 17
	CMD_GET_FLIGHT_MODE_CODE              = 2003
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

type CameraStatusData struct {
	TimeUsec        uint64
	TargetSystem    uint8
	TargetComponent uint8
	ImgStatus       uint8
	VideoTimeSec    uint32
	VideoOn         uint8
	PhotoInterval   uint32
	StorageFreeKB   uint32
	StorageTotalKB  uint32
	BatteryPct      uint8
	LensTemperature float32
	SensorTemp      float32
	ZoomLevel       float32
	FocusLevel      float32
	ISO             uint16
	ShutterSpeedMs  uint32
	Resolution      uint8
	FrameRate       uint8
	AccentX         int16
	AccentY         int16
}

func ParseCameraStatus(payload []byte) (*CameraStatusData, error) {
	if len(payload) < 48 {
		return nil, errors.New("payload too short for camera status")
	}
	return &CameraStatusData{
		TimeUsec:        binary.LittleEndian.Uint64(payload[0:8]),
		TargetSystem:    uint8(payload[8]),
		TargetComponent: uint8(payload[9]),
		ImgStatus:       uint8(payload[10]),
		VideoTimeSec:    binary.LittleEndian.Uint32(payload[12:16]),
		VideoOn:         uint8(payload[16]),
		PhotoInterval:   binary.LittleEndian.Uint32(payload[17:21]),
		StorageFreeKB:   binary.LittleEndian.Uint32(payload[21:25]),
		StorageTotalKB:  binary.LittleEndian.Uint32(payload[25:29]),
		BatteryPct:      uint8(payload[29]),
		LensTemperature: math.Float32frombits(binary.LittleEndian.Uint32(payload[30:34])),
		SensorTemp:      math.Float32frombits(binary.LittleEndian.Uint32(payload[34:38])),
		ZoomLevel:       math.Float32frombits(binary.LittleEndian.Uint32(payload[38:42])),
		FocusLevel:      math.Float32frombits(binary.LittleEndian.Uint32(payload[42:46])),
		ISO:             binary.LittleEndian.Uint16(payload[46:48]),
	}, nil
}

type CameraFeedbackData struct {
	ImgIdx         uint64
	TimeUsec       uint64
	TargetSystem   uint8
	TargetComponent uint8
	Flags          uint8
	Result         uint8
	Latitude       int32
	Longitude      int32
	Altitude       float32
	RelativeAlt    float32
	Elevation      float32
	Heading        float32
	Roll           float32
	Pitch          float32
	CameraID       uint8
	CaptureResult  uint8
	ImageSeq       uint16
	Distance       float32
	SizeKB         uint32
}

func ParseCameraFeedback(payload []byte) (*CameraFeedbackData, error) {
	if len(payload) < 60 {
		return nil, errors.New("payload too short for camera feedback")
	}
	return &CameraFeedbackData{
		ImgIdx:         binary.LittleEndian.Uint64(payload[0:8]),
		TimeUsec:       binary.LittleEndian.Uint64(payload[8:16]),
		TargetSystem:   uint8(payload[16]),
		TargetComponent: uint8(payload[17]),
		Flags:          uint8(payload[18]),
		Result:         uint8(payload[19]),
		Latitude:       int32(binary.LittleEndian.Uint32(payload[20:24])),
		Longitude:      int32(binary.LittleEndian.Uint32(payload[24:28])),
		Altitude:       math.Float32frombits(binary.LittleEndian.Uint32(payload[28:32])),
		RelativeAlt:    math.Float32frombits(binary.LittleEndian.Uint32(payload[32:36])),
		Elevation:      math.Float32frombits(binary.LittleEndian.Uint32(payload[36:40])),
		Heading:        math.Float32frombits(binary.LittleEndian.Uint32(payload[40:44])),
		Roll:           math.Float32frombits(binary.LittleEndian.Uint32(payload[44:48])),
		Pitch:          math.Float32frombits(binary.LittleEndian.Uint32(payload[48:52])),
		CameraID:       uint8(payload[52]),
		CaptureResult:  uint8(payload[53]),
		ImageSeq:       binary.LittleEndian.Uint16(payload[54:56]),
		Distance:       math.Float32frombits(binary.LittleEndian.Uint32(payload[56:60])),
	}, nil
}

type PayloadStatusData struct {
	TimeUsec       uint64
	PayloadType    uint8
	PayloadID      uint8
	Status         uint8
	SubStatus      uint8
	SubComponent   uint8
	BatteryPct     uint8
	Temperature    float32
	Pressure       float32
	FlowRate       float32
	RemainingQty   float32
	TotalCapacity  float32
	OperationalHrs uint32
	FirmwareVersion uint32
}

func ParsePayloadStatus(payload []byte) (*PayloadStatusData, error) {
	if len(payload) < 32 {
		return nil, errors.New("payload too short")
	}
	return &PayloadStatusData{
		TimeUsec:       binary.LittleEndian.Uint64(payload[0:8]),
		PayloadType:    uint8(payload[8]),
		PayloadID:      uint8(payload[9]),
		Status:         uint8(payload[10]),
		SubStatus:      uint8(payload[11]),
		SubComponent:   uint8(payload[12]),
		BatteryPct:     uint8(payload[13]),
		Temperature:    math.Float32frombits(binary.LittleEndian.Uint32(payload[14:18])),
		Pressure:       math.Float32frombits(binary.LittleEndian.Uint32(payload[18:22])),
		FlowRate:       math.Float32frombits(binary.LittleEndian.Uint32(payload[22:26])),
		RemainingQty:   math.Float32frombits(binary.LittleEndian.Uint32(payload[26:30])),
		TotalCapacity:  math.Float32frombits(binary.LittleEndian.Uint32(payload[30:34])),
	}, nil
}

type NamedValueFloatData struct {
	TimeBootMs uint32
	Value      float32
	Name       string
}

func ParseNamedValueFloat(payload []byte) (*NamedValueFloatData, error) {
	if len(payload) < 18 {
		return nil, errors.New("payload too short")
	}
	nameBytes := payload[8:18]
	name := string(nameBytes[:])
	for i, b := range nameBytes {
		if b == 0 {
			name = string(nameBytes[:i])
			break
		}
	}
	return &NamedValueFloatData{
		TimeBootMs: binary.LittleEndian.Uint32(payload[0:4]),
		Value:      math.Float32frombits(binary.LittleEndian.Uint32(payload[4:8])),
		Name:       name,
	}, nil
}

type NamedValueIntData struct {
	TimeBootMs uint32
	Value      int32
	Name       string
}

func ParseNamedValueInt(payload []byte) (*NamedValueIntData, error) {
	if len(payload) < 18 {
		return nil, errors.New("payload too short")
	}
	nameBytes := payload[8:18]
	name := string(nameBytes[:])
	for i, b := range nameBytes {
		if b == 0 {
			name = string(nameBytes[:i])
			break
		}
	}
	return &NamedValueIntData{
		TimeBootMs: binary.LittleEndian.Uint32(payload[0:4]),
		Value:      int32(binary.LittleEndian.Uint32(payload[4:8])),
		Name:       name,
	}, nil
}

type ESCStatusData struct {
	Index       uint8
	TimeUsec    uint64
	RPM         uint32
	Voltage     float32
	Current     float32
	Temperature int16
	FaultFlags  uint16
	ErrorCount  uint16
	Throttle    float32
}

func ParseESCStatus(payload []byte) (*ESCStatusData, error) {
	if len(payload) < 31 {
		return nil, errors.New("payload too short for ESC status")
	}
	return &ESCStatusData{
		Index:       uint8(payload[0]),
		TimeUsec:    binary.LittleEndian.Uint64(payload[1:9]),
		RPM:         binary.LittleEndian.Uint32(payload[9:13]),
		Voltage:     math.Float32frombits(binary.LittleEndian.Uint32(payload[13:17])),
		Current:     math.Float32frombits(binary.LittleEndian.Uint32(payload[17:21])),
		Temperature: int16(binary.LittleEndian.Uint16(payload[21:23])),
		FaultFlags:  binary.LittleEndian.Uint16(payload[23:25]),
		ErrorCount:  binary.LittleEndian.Uint16(payload[25:27]),
		Throttle:    math.Float32frombits(binary.LittleEndian.Uint32(payload[27:31])),
	}, nil
}

type ESCInfoData struct {
	Index          uint8
	Count          uint8
	ConnectionType uint8
	FaultFlags     uint16
	ErrorCode      uint32
	Vendor         string
	Model          string
	Version        uint32
	SN             string
}

func ParseESCInfo(payload []byte) (*ESCInfoData, error) {
	if len(payload) < 12 {
		return nil, errors.New("payload too short for ESC info")
	}
	vendor := string(payload[8:24])
	for i, b := range []byte(vendor) {
		if b == 0 {
			vendor = string([]byte(vendor)[:i])
			break
		}
	}
	model := string(payload[24:40])
	for i, b := range []byte(model) {
		if b == 0 {
			model = string([]byte(model)[:i])
			break
		}
	}
	return &ESCInfoData{
		Index:          uint8(payload[0]),
		Count:          uint8(payload[1]),
		ConnectionType: uint8(payload[2]),
		FaultFlags:     binary.LittleEndian.Uint16(payload[3:5]),
		ErrorCode:      binary.LittleEndian.Uint32(payload[5:9]),
		Vendor:         vendor,
		Model:          model,
		Version:        binary.LittleEndian.Uint32(payload[40:44]),
	}, nil
}

type LinkStatusData struct {
	ActiveLink     uint8
	RadioRSSI      int8
	RadioConnected uint8
	LteRSSI        int8
	LteConnected   uint8
	LteNetworkType string
	PacketLoss     float32
	LatencyMs      uint32
	BytesSent      uint64
	BytesReceived  uint64
}

func ParseLinkStatus(payload []byte) (*LinkStatusData, error) {
	if len(payload) < 45 {
		return nil, errors.New("payload too short for link status")
	}

	lteNetworkType := string(payload[5:21])
	for i, b := range payload[5:21] {
		if b == 0 {
			lteNetworkType = string(payload[5 : 5+i])
			break
		}
	}

	return &LinkStatusData{
		ActiveLink:     uint8(payload[0]),
		RadioRSSI:      int8(payload[1]),
		RadioConnected: uint8(payload[2]),
		LteRSSI:        int8(payload[3]),
		LteConnected:   uint8(payload[4]),
		LteNetworkType: lteNetworkType,
		PacketLoss:     math.Float32frombits(binary.LittleEndian.Uint32(payload[21:25])),
		LatencyMs:      binary.LittleEndian.Uint32(payload[25:29]),
		BytesSent:      binary.LittleEndian.Uint64(payload[29:37]),
		BytesReceived:  binary.LittleEndian.Uint64(payload[37:45]),
	}, nil
}
