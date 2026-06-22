package mavlink

import (
	"encoding/binary"
	"errors"
	"math"
)

type ObstacleAvoidanceEventData struct {
	Timestamp       uint64
	EventID         uint32
	UAVIDFromFC     uint8
	Strategy        uint8
	Status          uint8
	SensorType      uint8
	Direction       uint8
	ObstacleDistance float32
	StartLat        float32
	StartLng        float32
	StartAlt        float32
	BypassPathCount uint8
}

func ParseObstacleAvoidanceEvent(payload []byte) (*ObstacleAvoidanceEventData, error) {
	if len(payload) < 38 {
		return nil, errors.New("payload too short for obstacle avoidance event")
	}
	return &ObstacleAvoidanceEventData{
		Timestamp:        binary.LittleEndian.Uint64(payload[0:8]),
		EventID:          binary.LittleEndian.Uint32(payload[8:12]),
		Strategy:         uint8(payload[12]),
		Status:           uint8(payload[13]),
		SensorType:       uint8(payload[14]),
		Direction:        uint8(payload[15]),
		ObstacleDistance: math.Float32frombits(binary.LittleEndian.Uint32(payload[16:20])),
		StartLat:         math.Float32frombits(binary.LittleEndian.Uint32(payload[20:24])),
		StartLng:         math.Float32frombits(binary.LittleEndian.Uint32(payload[24:28])),
		StartAlt:         math.Float32frombits(binary.LittleEndian.Uint32(payload[28:32])),
		BypassPathCount:  uint8(payload[32]),
	}, nil
}

type ObstacleAvoidanceStatusData struct {
	Timestamp       uint64
	Enabled         uint8
	Sensitivity     uint8
	Strategy        uint8
	DetectionRange  float32
	TotalDetections uint32
	TotalEvents     uint32
}

func ParseObstacleAvoidanceStatus(payload []byte) (*ObstacleAvoidanceStatusData, error) {
	if len(payload) < 28 {
		return nil, errors.New("payload too short for obstacle avoidance status")
	}
	return &ObstacleAvoidanceStatusData{
		Timestamp:        binary.LittleEndian.Uint64(payload[0:8]),
		Enabled:          uint8(payload[8]),
		Sensitivity:      uint8(payload[9]),
		Strategy:         uint8(payload[10]),
		DetectionRange:   math.Float32frombits(binary.LittleEndian.Uint32(payload[11:15])),
		TotalDetections:  binary.LittleEndian.Uint32(payload[15:19]),
		TotalEvents:      binary.LittleEndian.Uint32(payload[19:23]),
	}, nil
}

type ObstacleHeatmapPointData struct {
	Latitude     float32
	Longitude    float32
	Altitude     float32
	TriggerCount uint16
	MinDistance   float32
}

type ObstacleHeatmapUpdateData struct {
	PointCount uint8
	Points     [8]ObstacleHeatmapPointData
}

func ParseObstacleHeatmapUpdate(payload []byte) (*ObstacleHeatmapUpdateData, error) {
	if len(payload) < 19 {
		return nil, errors.New("payload too short for obstacle heatmap update")
	}

	result := &ObstacleHeatmapUpdateData{
		PointCount: uint8(payload[0]),
	}

	offset := 1
	for i := 0; i < int(result.PointCount) && i < 8; i++ {
		if offset+18 > len(payload) {
			return nil, errors.New("payload too short for heatmap point data")
		}
		result.Points[i] = ObstacleHeatmapPointData{
			Latitude:     math.Float32frombits(binary.LittleEndian.Uint32(payload[offset : offset+4])),
			Longitude:    math.Float32frombits(binary.LittleEndian.Uint32(payload[offset+4 : offset+8])),
			Altitude:     math.Float32frombits(binary.LittleEndian.Uint32(payload[offset+8 : offset+12])),
			TriggerCount: binary.LittleEndian.Uint16(payload[offset+12 : offset+14]),
			MinDistance:  math.Float32frombits(binary.LittleEndian.Uint32(payload[offset+14 : offset+18])),
		}
		offset += 18
	}

	return result, nil
}
