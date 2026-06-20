package websocket

import (
	"time"

	"github.com/gin-gonic/gin"
)

func BroadcastFormationUpdate(formationID uint64, data interface{}) {
	telemetryHub.BroadcastFormationUpdate(formationID, data)
}

func BroadcastFormationStatus(formationID uint64, status string) {
	telemetryHub.BroadcastFormationStatus(formationID, status)
}

func BroadcastFormationCollisionWarning(formationID uint64, uavID1, uavID2 uint64, distance float64, level string) {
	data := gin.H{
		"formation_id": formationID,
		"uav_id_1":     uavID1,
		"uav_id_2":     uavID2,
		"distance":     distance,
		"warning_level": level,
		"timestamp":    time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastFormationCollisionWarning(formationID, data)
}

func BroadcastFormationLight(formationID uint64, red, green, blue uint8, effect string) {
	data := gin.H{
		"red":       red,
		"green":     green,
		"blue":      blue,
		"effect":    effect,
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	telemetryHub.BroadcastFormationLightUpdate(formationID, data)
}

func SubscribeFormation(client *Client, formationID uint64) {
	telemetryHub.SubscribeFormation(client, formationID)
}

func UnsubscribeFormation(client *Client, formationID uint64) {
	telemetryHub.UnsubscribeFormation(client, formationID)
}
