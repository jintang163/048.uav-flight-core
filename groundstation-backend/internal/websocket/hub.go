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

	uavSubscriptions map[uint64][]*Client
	metricsService   *service.MetricsService
	telemetryService *service.FlightService
}

type Message struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	UAVID   uint64      `json:"uav_id,omitempty"`
	Time    int64       `json:"time"`
}

var hub *Hub
var once sync.Once

func NewHub() *Hub {
	once.Do(func() {
		hub = &Hub{
			broadcast:        make(chan []byte, 1024),
			register:         make(chan *Client),
			unregister:       make(chan *Client),
			clients:          make(map[*Client]bool),
			uavSubscriptions: make(map[uint64][]*Client),
			metricsService:   service.NewMetricsService(),
			telemetryService: service.NewFlightService(),
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
}

func (h *Hub) broadcastTelemetry() {
	allData, _ := h.telemetryService.GetAllRealtimeData()
	if allData == nil {
		return
	}

	msg := &Message{
		Type: "telemetry_all",
		Data: allData,
		Time: time.Now().UnixNano() / 1e6,
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
		Type:  "telemetry",
		Data:  data,
		UAVID: uavID,
		Time:  time.Now().UnixNano() / 1e6,
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
		Type: "alert",
		Data: data,
		Time: time.Now().UnixNano() / 1e6,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.broadcast <- bytes
}

func (h *Hub) BroadcastMissionUpdate(uavID uint64, data interface{}) {
	msg := &Message{
		Type:  "mission_update",
		Data:  data,
		UAVID: uavID,
		Time:  time.Now().UnixNano() / 1e6,
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
		Type: msgType,
		Data: data,
		Time: time.Now().UnixNano() / 1e6,
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
