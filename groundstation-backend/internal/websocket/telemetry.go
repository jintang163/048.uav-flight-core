package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"time"
)

var telemetryHub = NewHub()
var flightService = service.NewFlightService()

type TelemetryData struct {
	UAVID       uint64  `json:"uav_id"`
	Timestamp   int64   `json:"timestamp"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	Altitude    float64 `json:"altitude"`
	RelativeAlt float64 `json:"relative_alt"`
	Heading     float64 `json:"heading"`
	GroundSpeed float64 `json:"ground_speed"`
	AirSpeed    float64 `json:"air_speed"`
	Throttle    float64 `json:"throttle"`
	Battery     float64 `json:"battery"`
	BatteryCurrent float64 `json:"battery_current"`
	SignalStrength int  `json:"signal_strength"`
	Satellites  int     `json:"satellites"`
	ArmStatus   bool    `json:"armed"`
	FlightMode  string  `json:"flight_mode"`
	Roll        float64 `json:"roll"`
	Pitch       float64 `json:"pitch"`
	Yaw         float64 `json:"yaw"`
	RollSpeed   float64 `json:"roll_speed"`
	PitchSpeed  float64 `json:"pitch_speed"`
	YawSpeed    float64 `json:"yaw_speed"`
	Vx          float64 `json:"vx"`
	Vy          float64 `json:"vy"`
	Vz          float64 `json:"vz"`
}

func BroadcastTelemetry(uavID uint64, status *models.FlightStatus) {
	data := &TelemetryData{
		UAVID:       uavID,
		Timestamp:   time.Now().UnixNano() / 1e6,
		Latitude:    status.Latitude,
		Longitude:   status.Longitude,
		Altitude:    status.Altitude,
		RelativeAlt: status.RelativeAltitude,
		Heading:     status.Heading,
		GroundSpeed: status.GroundSpeed,
		AirSpeed:    status.AirSpeed,
		Throttle:    status.Throttle,
		Battery:     status.BatteryVoltage,
		BatteryCurrent: status.BatteryCurrent,
		SignalStrength: int(status.SignalStrength),
		Satellites:  status.SatellitesVisible,
		ArmStatus:   status.Armed,
		FlightMode:  status.FlightMode,
		Roll:        status.Roll,
		Pitch:       status.Pitch,
		Yaw:         status.Yaw,
		RollSpeed:   status.RollSpeed,
		PitchSpeed:  status.PitchSpeed,
		YawSpeed:    status.YawSpeed,
		Vx:          status.Vx,
		Vy:          status.Vy,
		Vz:          status.Vz,
	}

	telemetryHub.BroadcastUAVTelemetry(uavID, data)
}

func BroadcastUAVStatus(uavID uint64, status string, message string) {
	data := gin.H{
		"uav_id":    uavID,
		"status":    status,
		"message":   message,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type: "uav_status",
		Data: data,
		Time: time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.broadcast <- bytes
}

func BroadcastMissionProgress(uavID uint64, missionID uint64, currentWP int, totalWP int, progress float64) {
	data := gin.H{
		"uav_id":      uavID,
		"mission_id":  missionID,
		"current_wp":  currentWP,
		"total_wp":    totalWP,
		"progress":    progress,
		"timestamp":   time.Now().UnixNano() / 1e6,
	}

	telemetryHub.BroadcastMissionUpdate(uavID, data)
}

func GetRealtimeData(uavID uint64) (map[string]interface{}, error) {
	return flightService.GetRealtimeData(uavID)
}

func GetAllRealtimeData() (map[string]interface{}, error) {
	return flightService.GetAllRealtimeData()
}

import "github.com/gin-gonic/gin"
