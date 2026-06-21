package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
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
		"uavId":     uavID,
		"status":    status,
		"message":   message,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "uav_status",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.broadcast <- bytes
}

func BroadcastMissionProgress(uavID uint64, missionID uint64, currentWP int, totalWP int, progress float64) {
	data := gin.H{
		"uav_id":      uavID,
		"uavId":       uavID,
		"mission_id":  missionID,
		"missionId":   missionID,
		"current_wp":  currentWP,
		"currentWaypoint": currentWP,
		"total_wp":    totalWP,
		"totalWaypoints": totalWP,
		"progress":    progress,
		"timestamp":   time.Now().UnixNano() / 1e6,
	}

	telemetryHub.BroadcastMissionProgress(uavID, data)
}

func BroadcastMissionProgressSimple(uavID uint64, currentWP int, totalWP int) {
	var progress float64
	if totalWP > 0 {
		progress = float64(currentWP) / float64(totalWP)
	}
	BroadcastMissionProgress(uavID, 0, currentWP, totalWP, progress)
}

func BroadcastWaypointReached(uavID uint64, wpIndex int) {
	data := gin.H{
		"uav_id":    uavID,
		"uavId":     uavID,
		"wp_index":  wpIndex,
		"waypointIndex": wpIndex,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastWaypointReached(uavID, data)
}

func BroadcastMissionStatus(uavID uint64, status string, currentWP int, totalWP int) {
	data := gin.H{
		"uav_id":            uavID,
		"uavId":             uavID,
		"status":            status,
		"current_wp":        currentWP,
		"currentWaypoint":   currentWP,
		"total_wp":          totalWP,
		"totalWaypoints":    totalWP,
		"timestamp":         time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastMissionStatus(uavID, data)
}

func BroadcastUAVMode(uavID uint64, mode string) {
	data := gin.H{
		"uav_id":    uavID,
		"uavId":     uavID,
		"mode":      mode,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastUAVMode(uavID, data)
}

func BroadcastBattery(uavID uint64, voltage float64, remaining float64) {
	data := gin.H{
		"uav_id":    uavID,
		"uavId":     uavID,
		"voltage":   voltage,
		"remaining": remaining,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastBattery(uavID, data)
}

func BroadcastGeofenceViolation(uavID uint64, fenceID uint64, latitude float64, longitude float64) {
	data := gin.H{
		"uav_id":    uavID,
		"uavId":     uavID,
		"fence_id":  fenceID,
		"fenceId":   fenceID,
		"latitude":  latitude,
		"longitude": longitude,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastGeofenceViolation(uavID, data)
}

func GetRealtimeData(uavID uint64) (map[string]interface{}, error) {
	return flightService.GetRealtimeData(uavID)
}

func GetAllRealtimeData() (map[string]interface{}, error) {
	return flightService.GetAllRealtimeData()
}

func BroadcastLinkStatus(uavID uint64, status interface{}) {
	data := map[string]interface{}{
		"uav_id":    uavID,
		"uavId":     uavID,
		"status":    status,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "link_status",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.BroadcastUAVTelemetry(uavID, msg)
	telemetryHub.broadcast <- bytes
}

func BroadcastChargingStatus(stationID uint64, slotID uint64, level float64, voltage float64, current float64) {
	data := map[string]interface{}{
		"station_id": stationID,
		"stationId":  stationID,
		"slot_id":    slotID,
		"slotId":     slotID,
		"level":      level,
		"voltage":    voltage,
		"current":    current,
		"timestamp":  time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "charging_status",
		Data:    data,
		Payload: data,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.broadcast <- bytes
}

func BroadcastChargingStationStatus(stationID uint64, status string, occupiedSlots int, chargingSlots int) {
	data := map[string]interface{}{
		"station_id":     stationID,
		"stationId":      stationID,
		"status":         status,
		"occupied_slots": occupiedSlots,
		"occupiedSlots":  occupiedSlots,
		"charging_slots": chargingSlots,
		"chargingSlots":  chargingSlots,
		"timestamp":      time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "charging_station_status",
		Data:    data,
		Payload: data,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.broadcast <- bytes
}

func BroadcastBatteryMaintenanceAlert(alert interface{}) {
	msg := &Message{
		Type:    "battery_maintenance_alert",
		Data:    alert,
		Payload: alert,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, _ := json.Marshal(msg)
	telemetryHub.broadcast <- bytes
}
