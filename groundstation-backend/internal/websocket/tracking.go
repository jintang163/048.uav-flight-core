package websocket

import (
	"groundstation-backend/internal/models"
)

func BroadcastDetections(uavID uint64, targets []*models.DetectionTarget) {
	payload := map[string]interface{}{
		"uav_id":     uavID,
		"detections": targets,
	}
	hub.BroadcastToUAV(uavID, "detections_update", payload)
}

func BroadcastTrackingUpdate(uavID uint64, task *models.TrackingTask) {
	payload := map[string]interface{}{
		"uav_id":        uavID,
		"tracking_task": task,
	}
	hub.BroadcastToUAV(uavID, "tracking_update", payload)
}

func SubscribeTracking(client *Client, uavID uint64) {
	hub.SubscribeUAV(client, uavID)
}

func UnsubscribeTracking(client *Client, uavID uint64) {
	hub.UnsubscribeUAV(client, uavID)
}
