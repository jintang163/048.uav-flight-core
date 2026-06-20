package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  10240,
	WriteBufferSize: 10240,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	userID  uint64
	role    string
	UAVIDs  []uint64
}

type ClientMessage struct {
	Action string          `json:"action"`
	UAVID  uint64          `json:"uav_id,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

func NewClient(hub *Hub, conn *websocket.Conn, userID uint64, role string) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		role:   role,
		UAVIDs: make([]uint64, 0),
	}
}

func ServeWS(c *gin.Context) {
	hub := NewHub()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "WebSocket升级失败: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	role := string(middleware.GetCurrentUserRole(c))

	client := NewClient(hub, conn, userID, role)
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				middleware.Logger.Error("WebSocket read error", zap.Error(err))
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(rawMessage []byte) {
	var msg ClientMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		c.hub.SendToClient(c, "error", "无效的消息格式")
		return
	}

	switch msg.Action {
	case "subscribe":
		if msg.UAVID > 0 {
			c.hub.Subscribe(c, msg.UAVID)
			c.UAVIDs = append(c.UAVIDs, msg.UAVID)
			c.hub.SendToClient(c, "subscribed", gin.H{"uav_id": msg.UAVID})
		}

	case "unsubscribe":
		if msg.UAVID > 0 {
			c.hub.Unsubscribe(c, msg.UAVID)
			for i, id := range c.UAVIDs {
				if id == msg.UAVID {
					c.UAVIDs = append(c.UAVIDs[:i], c.UAVIDs[i+1:]...)
					break
				}
			}
			c.hub.SendToClient(c, "unsubscribed", gin.H{"uav_id": msg.UAVID})
		}

	case "command":
		c.handleCommand(msg)

	case "ping":
		c.hub.SendToClient(c, "pong", time.Now().UnixNano()/1e6)

	default:
		c.hub.SendToClient(c, "error", "未知的操作类型")
	}
}

import "go.uber.org/zap"
import "github.com/gin-gonic/gin"
