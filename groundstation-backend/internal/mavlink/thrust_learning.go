package mavlink

import (
	"encoding/binary"
	"errors"
	"math"
)

type ThrustLearningStatusData struct {
	State         uint8
	WeightKG      float32
	HoverThrottle float32
	SampleCount   uint32
	Progress      float32
}

func ParseThrustLearningStatus(payload []byte) (state string, weight float64, hoverThrottle float64, sampleCount uint32, progress float64, err error) {
	if len(payload) < 15 {
		err = errors.New("payload too short for thrust learning status")
		return
	}

	stateCode := uint8(payload[0])
	switch stateCode {
	case 0:
		state = "idle"
	case 1:
		state = "weight_estimation"
	case 2:
		state = "data_collecting"
	case 3:
		state = "model_optimizing"
	case 4:
		state = "applied"
	default:
		state = "idle"
	}

	weight = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[1:5])))
	hoverThrottle = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[5:9])))
	sampleCount = binary.LittleEndian.Uint32(payload[9:13])
	progress = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[13:17])))

	return
}

type ThrustCurvePointData struct {
	Throttle float32
	ThrustN  float32
}

type ThrustCurveData struct {
	StartIdx uint8
	Count    uint8
	Points   []ThrustCurvePointData
}

func ParseThrustCurveData(payload []byte) (startIdx int, count int, points []struct {
	Throttle float64
	Thrust   float64
}, err error) {
	if len(payload) < 2 {
		err = errors.New("payload too short for thrust curve data")
		return
	}

	startIdx = int(uint8(payload[0]))
	count = int(uint8(payload[1]))

	expectedLen := 2 + count*8
	if len(payload) < expectedLen {
		err = errors.New("payload too short for thrust curve points")
		return
	}

	points = make([]struct {
		Throttle float64
		Thrust   float64
	}, count)

	offset := 2
	for i := 0; i < count; i++ {
		points[i].Throttle = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
		points[i].Thrust = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset+4 : offset+8])))
		offset += 8
	}

	return
}

type ThrustSampleData struct {
	Throttle    float32
	AccelZ      float32
	Altitude    float32
	VZ          float32
	MotorPWM1   uint16
	MotorPWM2   uint16
	MotorPWM3   uint16
	MotorPWM4   uint16
	Voltage     float32
	TimestampMs uint64
}

func ParseThrustSample(payload []byte) (*ThrustSampleData, error) {
	if len(payload) < 24 {
		return nil, errors.New("payload too short for thrust sample")
	}

	sample := &ThrustSampleData{
		Throttle:  math.Float32frombits(binary.LittleEndian.Uint32(payload[0:4])),
		AccelZ:    math.Float32frombits(binary.LittleEndian.Uint32(payload[4:8])),
		Altitude:  math.Float32frombits(binary.LittleEndian.Uint32(payload[8:12])),
		VZ:        math.Float32frombits(binary.LittleEndian.Uint32(payload[12:16])),
		MotorPWM1: binary.LittleEndian.Uint16(payload[16:18]),
		MotorPWM2: binary.LittleEndian.Uint16(payload[18:20]),
		MotorPWM3: binary.LittleEndian.Uint16(payload[20:22]),
		MotorPWM4: binary.LittleEndian.Uint16(payload[22:24]),
	}

	return sample, nil
}

func ParsePIDGainsReport(payload []byte) (map[string]float64, error) {
	if len(payload) < 60 {
		return nil, errors.New("payload too short for PID gains report")
	}

	gains := make(map[string]float64)
	offset := 0

	gains["roll_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["roll_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["roll_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["pitch_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["pitch_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["pitch_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["yaw_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["yaw_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["yaw_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["rate_roll_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_roll_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_roll_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["rate_pitch_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_pitch_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_pitch_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["rate_yaw_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_yaw_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["rate_yaw_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4

	gains["alt_kp"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["alt_ki"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))
	offset += 4
	gains["alt_kd"] = float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])))

	return gains, nil
}
