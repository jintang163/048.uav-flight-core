package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/service"
	"sync"
	"time"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex

	uavSubscriptions       map[uint64][]*Client
	formationSubscriptions map[uint64][]*Client
	metricsService         *service.MetricsService
	telemetryService       *service.FlightService
	formationService       *service.FormationService
}

type Message struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Payload interface{} `json:"payload"`
	UAVID   uint64      `json:"uav_id,omitempty"`
	UavID   uint64      `json:"uavId,omitempty"`
	Time    int64       `json:"time"`
}

var hub *Hub
var once sync.Once

func NewHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			broadcast:              make(chan []byte, 1024),
			register:               make(chan *Client),
			unregister:             make(chan *Client),
			clients:                make(map[*Client]bool),
			uavSubscriptions:       make(map[uint64][]*Client),
			formationSubscriptions: make(map[uint64][]*Client),
			metricsService:         service.NewMetricsService(),
			telemetryService:       service.NewFlightService(),
			formationService:       service.NewFormationService(),
		}
		go hub.run()
	})
	return hub
}

func (h *Hub) run() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.metricsService.SetWebSocketConnections(len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.removeClientFromSubscriptions(client)
			}
			h.mu.Unlock()
			h.metricsService.SetWebSocketConnections(len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.broadcastTelemetry()
		}
	}
}

func (h *Hub) removeClientFromSubscriptions(client *Client) {
	for uavID, clients := range h.uavSubscriptions {
		for i, c := range clients {
			if c == client {
				h.uavSubscriptions[uavID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}
	for formationID, clients := range h.formationSubscriptions {
		for i, c := range clients {
			if c == client {
				h.formationSubscriptions[formationID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
	}
}

func (h *Hub) broadcastTelemetry() {
	allData, _ := h.telemetryService.GetAllRealtimeData()
	if allData == nil {
		return
	}

	msg := &Message{
		Type:    "telemetry_all",
		Data:    allData,
		Payload: allData,
		Time:    time.Now().UnixNano() / 1e6,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) BroadcastUAVTelemetry(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "telemetry",
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

func (h *Hub) BroadcastAlert(data interface{}) {
	msg := &Message{
		Type:    "alert",
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

func (h *Hub) BroadcastMissionUpdate(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "mission_update",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastMissionProgress(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "mission_progress",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastWaypointReached(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "waypoint_reached",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastMissionStatus(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "mission_status",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastUAVStatus(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "uav_status",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastUAVMode(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "uav_mode",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastBattery(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "battery",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastGeofenceViolation(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "geofence_violation",
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

	h.broadcast <- bytes
}

func (h *Hub) Subscribe(client *Client, uavID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.uavSubscriptions[uavID] = append(h.uavSubscriptions[uavID], client)
}

func (h *Hub) Unsubscribe(client *Client, uavID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.uavSubscriptions[uavID]
	for i, c := range clients {
		if c == client {
			h.uavSubscriptions[uavID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) SendToClient(client *Client, msgType string, data interface{}) {
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

	select {
	case client.send <- bytes:
	default:
	}
}

func (h *Hub) BroadcastFormationUpdate(formationID uint64, data interface{}) {
	msg := &Message{
		Type:    "formation_update",
		Data:    data,
		Payload: data,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.formationSubscriptions[formationID]
	for _, client := range clients {
		select {
		case client.send <- bytes:
		default:
		}
	}

	h.broadcast <- bytes
}

func (h *Hub) BroadcastFormationStatus(formationID uint64, status string) {
	data := map[string]interface{}{
		"formation_id": formationID,
		"status":       status,
		"timestamp":    time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "formation_status",
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

func (h *Hub) BroadcastFormationCollisionWarning(formationID uint64, warning interface{}) {
	msg := &Message{
		Type:    "formation_collision_warning",
		Data:    warning,
		Payload: warning,
		Time:    time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.broadcast <- bytes
}

func (h *Hub) BroadcastFormationLightUpdate(formationID uint64, lightData interface{}) {
	data := map[string]interface{}{
		"formation_id": formationID,
		"light":        lightData,
		"timestamp":    time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "formation_light",
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

func (h *Hub) BroadcastObstacleAvoidanceEvent(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "obstacle_avoidance_start",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastObstacleAvoidanceStatus(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "obstacle_avoidance_config",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastObstacleHeatmapUpdate(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "obstacle_heatmap_update",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastThrustLearningStatus(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "thrust_learning_status",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastThrustCurveUpdate(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "thrust_curve_update",
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

	h.broadcast <- bytes
}

func (h *Hub) BroadcastPIDGainsUpdate(uavID uint64, data interface{}) {
	msg := &Message{
		Type:    "pid_gains_update",
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

	h.broadcast <- bytes
}

func (h *Hub) SubscribeFormation(client *Client, formationID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.formationSubscriptions[formationID] = append(h.formationSubscriptions[formationID], client)
}

func (h *Hub) UnsubscribeFormation(client *Client, formationID uint64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.formationSubscriptions[formationID]
	for i, c := range clients {
		if c == client {
			h.formationSubscriptions[formationID] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
}

func (h *Hub) broadcastFormations() {
	formations, err := h.formationService.GetActiveFormations()
	if err != nil {
		return
	}

	for _, formation := range formations {
		h.BroadcastFormationUpdate(formation.ID, formation)
	}
}
