package websocket

import (
	"encoding/json"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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
	Type    string          `json:"type"`
	Action  string          `json:"action"`
	UAVID   interface{}     `json:"uav_id,omitempty"`
	UavID   interface{}     `json:"uavId,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Time    int64           `json:"time,omitempty"`
}

type OutboundMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	UAVID   uint64      `json:"uav_id,omitempty"`
	Time    int64       `json:"time"`
}

type CommandPayload struct {
	UAVID   interface{}         `json:"uavId,omitempty"`
	UavID   interface{}         `json:"uav_id,omitempty"`
	Command string              `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

func parseUint64(val interface{}) uint64 {
	switch v := val.(type) {
	case float64:
		return uint64(v)
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0
		}
		return n
	case uint64:
		return v
	case int64:
		return uint64(v)
	case int:
		return uint64(v)
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0
		}
		return uint64(n)
	default:
		return 0
	}
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
		c.hub.SendToClient(c, "error", "无效的消息格式: "+err.Error())
		return
	}

	msgType := msg.Type
	if msgType == "" {
		msgType = msg.Action
	}

	rawData := msg.Data
	if len(rawData) == 0 {
		rawData = msg.Payload
	}

	uavID := parseUint64(msg.UAVID)
	if uavID == 0 {
		uavID = parseUint64(msg.UavID)
	}

	switch msgType {
	case "subscribe", "subscribe_uav":
		if uavID > 0 {
			c.hub.Subscribe(c, uavID)
			c.UAVIDs = append(c.UAVIDs, uavID)
			c.hub.SendToClient(c, "subscribed", gin.H{"uavId": uavID, "uav_id": uavID})
		}

	case "unsubscribe", "unsubscribe_uav":
		if uavID > 0 {
			c.hub.Unsubscribe(c, uavID)
			for i, id := range c.UAVIDs {
				if id == uavID {
					c.UAVIDs = append(c.UAVIDs[:i], c.UAVIDs[i+1:]...)
					break
				}
			}
			c.hub.SendToClient(c, "unsubscribed", gin.H{"uavId": uavID, "uav_id": uavID})
		}

	case "command":
		var cmdPayload CommandPayload
		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &cmdPayload); err != nil {
				c.hub.SendToClient(c, "error", "无效的指令格式: "+err.Error())
				return
			}
		} else {
			cmdPayload.UAVID = msg.UAVID
			cmdPayload.UavID = msg.UavID
		}

		cmdUAVID := parseUint64(cmdPayload.UAVID)
		if cmdUAVID == 0 {
			cmdUAVID = parseUint64(cmdPayload.UavID)
		}
		if cmdUAVID == 0 {
			cmdUAVID = uavID
		}

		handler.HandleUAVCommand(c.userID, cmdUAVID, cmdPayload.Command, cmdPayload.Params)
		c.hub.SendToClient(c, "command_ack", gin.H{
			"command": cmdPayload.Command,
			"uavId":   cmdUAVID,
			"result":  true,
		})

	case "ping", "heartbeat":
		c.hub.SendToClient(c, "pong", time.Now().UnixNano()/1e6)

	case "subscribe_alerts":
		c.hub.SendToClient(c, "subscribed_alerts", gin.H{"success": true})

	case "unsubscribe_alerts":
		c.hub.SendToClient(c, "unsubscribed_alerts", gin.H{"success": true})

	case "request_telemetry":
		if uavID > 0 {
			handler.RequestUAVTelemetry(uavID)
		}

	case "subscribe_formation":
		var formationPayload struct {
			FormationID uint64 `json:"formation_id"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &formationPayload)
		}
		if formationPayload.FormationID > 0 {
			c.hub.SubscribeFormation(c, formationPayload.FormationID)
			c.hub.SendToClient(c, "subscribed_formation", gin.H{"formation_id": formationPayload.FormationID})
		}

	case "unsubscribe_formation":
		var formationPayload struct {
			FormationID uint64 `json:"formation_id"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &formationPayload)
		}
		if formationPayload.FormationID > 0 {
			c.hub.UnsubscribeFormation(c, formationPayload.FormationID)
			c.hub.SendToClient(c, "unsubscribed_formation", gin.H{"formation_id": formationPayload.FormationID})
		}

	case "cockpit_start_session":
		cockpitService := service.NewRemoteCockpitService()
		session, err := cockpitService.StartSession(uavID, c.userID)
		if err != nil {
			c.hub.SendToClient(c, "error", "启动驾驶舱会话失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_session_started", session)

	case "cockpit_end_session":
		cockpitService := service.NewRemoteCockpitService()
		session, err := cockpitService.EndSession(uavID)
		if err != nil {
			c.hub.SendToClient(c, "error", "结束驾驶舱会话失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_session_ended", session)

	case "cockpit_start_video":
		cockpitService := service.NewRemoteCockpitService()
		var videoReq struct {
			Codec            string `json:"codec"`
			Resolution       string `json:"resolution"`
			BitrateKbps      int    `json:"bitrate_kbps"`
			FPS              int    `json:"fps"`
			AdaptiveEnabled  *bool  `json:"adaptive_enabled"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &videoReq)
		}
		config := &service.VideoStreamConfig{
			Codec:           models.VideoCodecH265,
			Resolution:      models.VideoRes720P,
			BitrateKbps:     4000,
			FPS:             30,
			AdaptiveEnabled: true,
		}
		if videoReq.Codec != "" {
			config.Codec = models.VideoCodec(videoReq.Codec)
		}
		if videoReq.Resolution != "" {
			config.Resolution = models.VideoResolution(videoReq.Resolution)
		}
		if videoReq.BitrateKbps > 0 {
			config.BitrateKbps = videoReq.BitrateKbps
		}
		if videoReq.FPS > 0 {
			config.FPS = videoReq.FPS
		}
		if videoReq.AdaptiveEnabled != nil {
			config.AdaptiveEnabled = *videoReq.AdaptiveEnabled
		}
		stream, err := cockpitService.StartVideoStream(uavID, config)
		if err != nil {
			c.hub.SendToClient(c, "error", "启动视频流失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "video_stream_status", stream)

	case "cockpit_stop_video":
		cockpitService := service.NewRemoteCockpitService()
		if err := cockpitService.StopVideoStream(uavID); err != nil {
			c.hub.SendToClient(c, "error", "停止视频流失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "video_stream_disconnected", gin.H{"uav_id": uavID, "active": false})

	case "cockpit_adjust_quality":
		cockpitService := service.NewRemoteCockpitService()
		var qualityReq struct {
			BitrateKbps *int   `json:"bitrate_kbps"`
			Resolution  string `json:"resolution"`
		}
		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &qualityReq); err != nil {
				c.hub.SendToClient(c, "error", "无效的画质调整参数: "+err.Error())
				return
			}
		}
		var resolution *models.VideoResolution
		if qualityReq.Resolution != "" {
			r := models.VideoResolution(qualityReq.Resolution)
			resolution = &r
		}
		stream, err := cockpitService.AdjustVideoQuality(uavID, qualityReq.BitrateKbps, resolution)
		if err != nil {
			c.hub.SendToClient(c, "error", "调整画质失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "video_quality_adjusted", stream)

	case "cockpit_set_primary_link":
		cockpitService := service.NewRemoteCockpitService()
		var linkReq struct {
			LinkType models.LinkType `json:"link_type"`
		}
		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &linkReq); err != nil {
				c.hub.SendToClient(c, "error", "无效的链路参数: "+err.Error())
				return
			}
		}
		status, err := cockpitService.SetPrimaryLink(uavID, linkReq.LinkType)
		if err != nil {
			c.hub.SendToClient(c, "error", "切换主链路失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_link_status", status)

	case "cockpit_set_failover":
		cockpitService := service.NewRemoteCockpitService()
		var failoverReq struct {
			Enabled bool `json:"enabled"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &failoverReq)
		}
		status, err := cockpitService.SetFailoverEnabled(uavID, failoverReq.Enabled)
		if err != nil {
			c.hub.SendToClient(c, "error", "设置链路故障切换失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_link_status", status)

	case "cockpit_set_auto_mission_fallback":
		cockpitService := service.NewRemoteCockpitService()
		var fallbackReq struct {
			Enabled bool `json:"enabled"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &fallbackReq)
		}
		if err := cockpitService.SetAutoMissionFallback(uavID, fallbackReq.Enabled); err != nil {
			c.hub.SendToClient(c, "error", "设置自动航线回退失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_auto_mission_fallback_updated", gin.H{"enabled": fallbackReq.Enabled})

	case "cockpit_switch_uav":
		cockpitService := service.NewRemoteCockpitService()
		var switchReq struct {
			FromUAVID uint64 `json:"from_uav_id"`
			ToUAVID   uint64 `json:"to_uav_id"`
		}
		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &switchReq); err != nil {
				c.hub.SendToClient(c, "error", "无效的切换参数: "+err.Error())
				return
			}
		}
		if switchReq.ToUAVID == 0 {
			switchReq.ToUAVID = uavID
		}
		if err := cockpitService.SwitchUAV(switchReq.FromUAVID, switchReq.ToUAVID, c.userID); err != nil {
			c.hub.SendToClient(c, "error", "切换无人机失败: "+err.Error())
			return
		}
		BroadcastCockpitUAVSwitched(switchReq.FromUAVID, switchReq.ToUAVID, c.userID)
		c.hub.SendToClient(c, "cockpit_uav_switched", gin.H{
			"from_uav_id": switchReq.FromUAVID,
			"to_uav_id":   switchReq.ToUAVID,
		})

	case "cockpit_send_control":
		cockpitService := service.NewRemoteCockpitService()
		var ctrlReq struct {
			Pitch    float64 `json:"pitch"`
			Roll     float64 `json:"roll"`
			Yaw      float64 `json:"yaw"`
			Throttle float64 `json:"throttle"`
			Source   string  `json:"source"`
		}
		if len(rawData) > 0 {
			if err := json.Unmarshal(rawData, &ctrlReq); err != nil {
				c.hub.SendToClient(c, "error", "无效的控制参数: "+err.Error())
				return
			}
		}
		source := ctrlReq.Source
		if source == "" {
			source = "gamepad"
		}
		if err := cockpitService.SendFlightControl(uavID, c.userID, ctrlReq.Pitch, ctrlReq.Roll, ctrlReq.Yaw, ctrlReq.Throttle, source); err != nil {
			c.hub.SendToClient(c, "error", "发送控制指令失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "cockpit_control_ack", gin.H{"success": true})

	case "cockpit_trigger_fallback":
		cockpitService := service.NewRemoteCockpitService()
		var fallbackReq struct {
			Reason string `json:"reason"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &fallbackReq)
		}
		reason := fallbackReq.Reason
		if reason == "" {
			reason = "手动触发"
		}
		if err := cockpitService.TriggerAutoMissionFallback(uavID, reason); err != nil {
			c.hub.SendToClient(c, "error", "触发航线回退失败: "+err.Error())
			return
		}
		c.hub.SendToClient(c, "auto_mission_fallback_triggered", gin.H{"uav_id": uavID, "reason": reason})

	case "cockpit_network_metrics":
		cockpitService := service.NewRemoteCockpitService()
		var metrics struct {
			BandwidthKbps  float64 `json:"bandwidth_estimate_kbps"`
			RTTms          int     `json:"rtt_ms"`
			PacketLoss     float64 `json:"packet_loss"`
			JitterMs       int     `json:"jitter_ms"`
			ThroughputKbps float64 `json:"throughput_kbps"`
			Timestamp      int64   `json:"timestamp"`
		}
		if len(rawData) > 0 {
			json.Unmarshal(rawData, &metrics)
		}
		if metrics.Timestamp == 0 {
			metrics.Timestamp = time.Now().UnixMilli()
		}
		_ = cockpitService.LogNetworkMetrics(uavID, &service.NetworkMetrics{
			BandwidthKbps:  metrics.BandwidthKbps,
			RTTms:          metrics.RTTms,
			PacketLoss:     metrics.PacketLoss,
			JitterMs:       metrics.JitterMs,
			ThroughputKbps: metrics.ThroughputKbps,
			Timestamp:      metrics.Timestamp,
		})

	default:
		c.hub.SendToClient(c, "error", "未知的消息类型: "+msgType)
	}
}
