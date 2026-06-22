package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/models"
	"time"
)

func BroadcastLandingStatus(uavID uint64, status string, landingPointID uint64, rtkEnabled, visionEnabled bool) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":           uavID,
		"status":           status,
		"landing_point_id": landingPointID,
		"rtk_enabled":      rtkEnabled,
		"vision_enabled":   visionEnabled,
		"timestamp":        time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "landing_status",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}

func BroadcastLandingTrajectory(uavID, sessionID uint64, point *models.LandingTrajectoryPoint) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":     uavID,
		"session_id": sessionID,
		"trajectory": point,
		"timestamp":  time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "landing_trajectory",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}

func BroadcastLandingVision(uavID uint64, visionData *models.VisionLandingData) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":      uavID,
		"session_id":  visionData.SessionID,
		"vision_data": visionData,
		"timestamp":   time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "landing_vision",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}

func BroadcastLandingRTK(uavID uint64, rtkData *models.RTKPositionData) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":   uavID,
		"session_id": rtkData.SessionID,
		"rtk_data": rtkData,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "landing_rtk",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}

func BroadcastForcedLanding(uavID uint64, event *models.ForcedLandingEvent) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":    uavID,
		"event":     event,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "forced_landing",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}

func BroadcastForcedLandingResolved(uavID uint64, eventID uint64) {
	hub := GetHub()
	if hub == nil {
		return
	}

	data := map[string]interface{}{
		"uav_id":    uavID,
		"event_id":  eventID,
		"resolved":  true,
		"timestamp": time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "forced_landing_resolved",
		Data:    data,
		Payload: data,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	hub.broadcast <- bytes
}
