package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/models"
	"time"
)

func BroadcastPayloadStatus(uavID uint64, payloadID uint64, status *models.PayloadDevice) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastPayloadStatus(uavID, payloadID, status)
}

func BroadcastCameraStatus(uavID uint64, payloadID uint64, cameraStatus *models.CameraStatus) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastCameraStatus(uavID, payloadID, cameraStatus)
}

func BroadcastSprayerStatus(uavID uint64, payloadID uint64, sprayerStatus *models.SprayerStatus) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastSprayerStatus(uavID, payloadID, sprayerStatus)
}

func BroadcastCameraFeedback(uavID uint64, payloadID uint64, captureResult string, imageSeq int, sizeKB int) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastCameraFeedback(uavID, payloadID, captureResult, imageSeq, sizeKB)
}

func BroadcastOrbitMissionProgress(uavID uint64, missionID uint64, status string, currentLoop int, totalLoops int, progress float64) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastOrbitMissionProgress(uavID, missionID, status, currentLoop, totalLoops, progress)
}

func BroadcastOrthoMissionProgress(uavID uint64, missionID uint64, status string, currentWP int, totalWP int, progress float64, photosCaptured int) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastOrthoMissionProgress(uavID, missionID, status, currentWP, totalWP, progress, photosCaptured)
}

func BroadcastTTSTaskProgress(taskID uint64, status string, audioURL string, errorMsg string) {
	hub := GetHub()
	if hub == nil {
		return
	}
	hub.BroadcastTTSTaskProgress(taskID, status, audioURL, errorMsg)
}

func (h *Hub) BroadcastPayloadStatus(uavID uint64, payloadID uint64, status *models.PayloadDevice) {
	data := map[string]interface{}{
		"uavId":     uavID,
		"payloadId": payloadID,
		"status":    status,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "payload_status", data)
}

func (h *Hub) BroadcastCameraStatus(uavID uint64, payloadID uint64, cameraStatus *models.CameraStatus) {
	data := map[string]interface{}{
		"uavId":        uavID,
		"payloadId":    payloadID,
		"cameraStatus": cameraStatus,
		"timestamp":    time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "camera_status", data)
}

func (h *Hub) BroadcastSprayerStatus(uavID uint64, payloadID uint64, sprayerStatus *models.SprayerStatus) {
	data := map[string]interface{}{
		"uavId":         uavID,
		"payloadId":     payloadID,
		"sprayerStatus": sprayerStatus,
		"timestamp":     time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "sprayer_status", data)
}

func (h *Hub) BroadcastCameraFeedback(uavID uint64, payloadID uint64, captureResult string, imageSeq int, sizeKB int) {
	data := map[string]interface{}{
		"uavId":         uavID,
		"payloadId":     payloadID,
		"captureResult": captureResult,
		"imageSeq":      imageSeq,
		"sizeKB":        sizeKB,
		"timestamp":     time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "camera_feedback", data)
}

func (h *Hub) BroadcastOrbitMissionProgress(uavID uint64, missionID uint64, status string, currentLoop int, totalLoops int, progress float64) {
	data := map[string]interface{}{
		"uavId":       uavID,
		"missionId":   missionID,
		"status":      status,
		"currentLoop": currentLoop,
		"totalLoops":  totalLoops,
		"progress":    progress,
		"timestamp":   time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "orbit_progress", data)
}

func (h *Hub) BroadcastOrthoMissionProgress(uavID uint64, missionID uint64, status string, currentWP int, totalWP int, progress float64, photosCaptured int) {
	data := map[string]interface{}{
		"uavId":           uavID,
		"missionId":       missionID,
		"status":          status,
		"currentWaypoint": currentWP,
		"totalWaypoints":  totalWP,
		"progress":        progress,
		"photosCaptured":  photosCaptured,
		"timestamp":       time.Now().UnixNano() / 1e6,
	}
	h.broadcastUAVMsg(uavID, "ortho_progress", data)
}

func (h *Hub) BroadcastTTSTaskProgress(taskID uint64, status string, audioURL string, errorMsg string) {
	data := map[string]interface{}{
		"taskId":    taskID,
		"status":    status,
		"audioURL":  audioURL,
		"error":     errorMsg,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	h.broadcastGlobalMsg("tts_task_progress", data)
}

func (h *Hub) broadcastUAVMsg(uavID uint64, msgType string, data interface{}) {
	msg := &Message{
		Type:    msgType,
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

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.uavSubscriptions[uavID]
	for _, client := range clients {
		select {
		case client.send <- bytes:
		default:
		}
	}

	h.broadcast <- bytes
}

func (h *Hub) broadcastGlobalMsg(msgType string, data interface{}) {
	msg := &Message{
		Type:    msgType,
		Data:    data,
		Payload: data,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.broadcast <- bytes
}
