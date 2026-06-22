package handler

import (
	"io"
	"net/http"
	"strconv"

	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type RemoteCockpitHandler struct {
	service *service.RemoteCockpitService
}

func NewRemoteCockpitHandler() *RemoteCockpitHandler {
	return &RemoteCockpitHandler{
		service: service.NewRemoteCockpitService(),
	}
}

var cockpitHandler = NewRemoteCockpitHandler()

func StartCockpitSession(c *gin.Context)       { cockpitHandler.StartSession(c) }
func EndCockpitSession(c *gin.Context)         { cockpitHandler.EndSession(c) }
func GetCockpitSession(c *gin.Context)         { cockpitHandler.GetSession(c) }
func StartCockpitVideoStream(c *gin.Context)   { cockpitHandler.StartVideoStream(c) }
func StopCockpitVideoStream(c *gin.Context)    { cockpitHandler.StopVideoStream(c) }
func GetCockpitVideoStream(c *gin.Context)     { cockpitHandler.GetVideoStream(c) }
func AdjustCockpitVideoQuality(c *gin.Context) { cockpitHandler.AdjustVideoQuality(c) }
func GetCockpitStreamURL(c *gin.Context)       { cockpitHandler.GetStreamURL(c) }
func GetCockpitLinkStatus(c *gin.Context)      { cockpitHandler.GetLinkStatus(c) }
func SetCockpitPrimaryLink(c *gin.Context)     { cockpitHandler.SetPrimaryLink(c) }
func SetCockpitFailoverEnabled(c *gin.Context) { cockpitHandler.SetFailoverEnabled(c) }
func SetCockpitAutoMissionFallback(c *gin.Context) {
	cockpitHandler.SetAutoMissionFallback(c)
}
func TriggerCockpitAutoMissionFallback(c *gin.Context) {
	cockpitHandler.TriggerAutoMissionFallback(c)
}
func GetAvailableCockpitUAVs(c *gin.Context)   { cockpitHandler.GetAvailableUAVs(c) }
func SwitchCockpitUAV(c *gin.Context)          { cockpitHandler.SwitchUAV(c) }
func SendCockpitFlightControl(c *gin.Context)  { cockpitHandler.SendFlightControl(c) }
func GetCockpitNetworkMetrics(c *gin.Context)  { cockpitHandler.GetNetworkMetrics(c) }
func GetCockpitFlightControlLogs(c *gin.Context) {
	cockpitHandler.GetFlightControlLogs(c)
}
func HandleCockpitSDPOffer(c *gin.Context)  { cockpitHandler.HandleSDPOffer(c) }
func GetCockpitWebRTCStats(c *gin.Context)   { cockpitHandler.GetWebRTCStats(c) }

type StartSessionRequest struct {
	UAVID   uint64 `json:"uav_id" binding:"required"`
	PilotID uint64 `json:"pilot_id" binding:"required"`
}

type StartVideoStreamRequest struct {
	Codec            string  `json:"codec"`
	Resolution       string  `json:"resolution"`
	BitrateKbps      int     `json:"bitrate_kbps"`
	FPS              int     `json:"fps"`
	KeyframeInterval int     `json:"keyframe_interval"`
	AdaptiveEnabled  *bool   `json:"adaptive_enabled"`
	MinBitrateKbps   int     `json:"min_bitrate_kbps"`
	MaxBitrateKbps   int     `json:"max_bitrate_kbps"`
	MinResolution    string  `json:"min_resolution"`
	MaxResolution    string  `json:"max_resolution"`
}

type AdjustQualityRequest struct {
	BitrateKbps *int   `json:"bitrate_kbps"`
	Resolution  string `json:"resolution"`
}

type SetPrimaryLinkRequest struct {
	LinkType models.LinkType `json:"link_type" binding:"required"`
}

type SetFailoverRequest struct {
	Enabled bool `json:"enabled" binding:"required"`
}

type SetAutoMissionFallbackRequest struct {
	Enabled bool `json:"enabled" binding:"required"`
}

type SendFlightControlRequest struct {
	Pitch    float64 `json:"pitch"`
	Roll     float64 `json:"roll"`
	Yaw      float64 `json:"yaw"`
	Throttle float64 `json:"throttle"`
	Source   string  `json:"source"`
	PilotID  uint64  `json:"pilot_id"`
}

type SwitchUAVRequest struct {
	FromUAVID uint64 `json:"from_uav_id"`
	ToUAVID   uint64 `json:"to_uav_id" binding:"required"`
	PilotID   uint64 `json:"pilot_id" binding:"required"`
}

type TriggerFallbackRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *RemoteCockpitHandler) StartSession(c *gin.Context) {
	var req StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	session, err := h.service.StartSession(req.UAVID, req.PilotID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, session)
}

func (h *RemoteCockpitHandler) EndSession(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	session, err := h.service.EndSession(uavID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, session)
}

func (h *RemoteCockpitHandler) GetSession(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	session, err := h.service.GetSession(uavID)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondSuccess(c, session)
}

func (h *RemoteCockpitHandler) StartVideoStream(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req StartVideoStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	config := &service.VideoStreamConfig{}
	if req.Codec != "" {
		config.Codec = models.VideoCodec(req.Codec)
	} else {
		config.Codec = models.VideoCodecH265
	}
	if req.Resolution != "" {
		config.Resolution = models.VideoResolution(req.Resolution)
	} else {
		config.Resolution = models.VideoRes720P
	}
	if req.BitrateKbps > 0 {
		config.BitrateKbps = req.BitrateKbps
	} else {
		config.BitrateKbps = 4000
	}
	if req.FPS > 0 {
		config.FPS = req.FPS
	} else {
		config.FPS = 30
	}
	if req.KeyframeInterval > 0 {
		config.KeyframeInterval = req.KeyframeInterval
	} else {
		config.KeyframeInterval = 60
	}
	if req.AdaptiveEnabled != nil {
		config.AdaptiveEnabled = *req.AdaptiveEnabled
	} else {
		config.AdaptiveEnabled = true
	}
	if req.MinBitrateKbps > 0 {
		config.MinBitrateKbps = req.MinBitrateKbps
	} else {
		config.MinBitrateKbps = 1000
	}
	if req.MaxBitrateKbps > 0 {
		config.MaxBitrateKbps = req.MaxBitrateKbps
	} else {
		config.MaxBitrateKbps = 8000
	}
	if req.MinResolution != "" {
		config.MinResolution = models.VideoResolution(req.MinResolution)
	} else {
		config.MinResolution = models.VideoRes640P
	}
	if req.MaxResolution != "" {
		config.MaxResolution = models.VideoResolution(req.MaxResolution)
	} else {
		config.MaxResolution = models.VideoRes1080P
	}

	stream, err := h.service.StartVideoStream(uavID, config)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, stream)
}

func (h *RemoteCockpitHandler) StopVideoStream(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	if err := h.service.StopVideoStream(uavID); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"message": "视频流已停止"})
}

func (h *RemoteCockpitHandler) GetVideoStream(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	stream, err := h.service.GetVideoStream(uavID)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondSuccess(c, stream)
}

func (h *RemoteCockpitHandler) AdjustVideoQuality(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req AdjustQualityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	var resolution *models.VideoResolution
	if req.Resolution != "" {
		r := models.VideoResolution(req.Resolution)
		resolution = &r
	}

	stream, err := h.service.AdjustVideoQuality(uavID, req.BitrateKbps, resolution)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, stream)
}

func (h *RemoteCockpitHandler) GetStreamURL(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	protocol := c.DefaultQuery("protocol", "webrtc")

	url, err := h.service.GetStreamURL(uavID, protocol)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondSuccess(c, url)
}

func (h *RemoteCockpitHandler) GetLinkStatus(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	status, err := h.service.GetLinkStatus(uavID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, status)
}

func (h *RemoteCockpitHandler) SetPrimaryLink(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req SetPrimaryLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	status, err := h.service.SetPrimaryLink(uavID, req.LinkType)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, status)
}

func (h *RemoteCockpitHandler) SetFailoverEnabled(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req SetFailoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	status, err := h.service.SetFailoverEnabled(uavID, req.Enabled)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, status)
}

func (h *RemoteCockpitHandler) SetAutoMissionFallback(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req SetAutoMissionFallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.SetAutoMissionFallback(uavID, req.Enabled); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"enabled": req.Enabled})
}

func (h *RemoteCockpitHandler) TriggerAutoMissionFallback(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req TriggerFallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.TriggerAutoMissionFallback(uavID, req.Reason); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"message": "已切换到航线飞行模式", "reason": req.Reason})
}

func (h *RemoteCockpitHandler) GetAvailableUAVs(c *gin.Context) {
	pilotIDStr := c.DefaultQuery("pilot_id", "1")
	pilotID, _ := strconv.ParseUint(pilotIDStr, 10, 64)

	uavs, err := h.service.GetAvailableUAVs(pilotID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"uav_ids": uavs, "count": len(uavs)})
}

func (h *RemoteCockpitHandler) SwitchUAV(c *gin.Context) {
	var req SwitchUAVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.SwitchUAV(req.FromUAVID, req.ToUAVID, req.PilotID); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"message": "已切换无人机", "to_uav_id": req.ToUAVID})
}

func (h *RemoteCockpitHandler) SendFlightControl(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	var req SendFlightControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	source := req.Source
	if source == "" {
		source = "gamepad"
	}

	if err := h.service.SendFlightControl(uavID, req.PilotID, req.Pitch, req.Roll, req.Yaw, req.Throttle, source); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, gin.H{"message": "控制指令已发送"})
}

func (h *RemoteCockpitHandler) GetNetworkMetrics(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	offset := (page - 1) * pageSize

	var logs []models.NetworkMetricsLog
	var total int64

	db := config.DB
	db.Model(&models.NetworkMetricsLog{}).Where("uav_id = ?", uavID).Count(&total)
	db.Where("uav_id = ?", uavID).Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&logs)

	utils.RespondSuccess(c, gin.H{
		"metrics":  logs,
		"total":    total,
		"page":     page,
		"page_size": pageSize,
	})
}

func (h *RemoteCockpitHandler) GetFlightControlLogs(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))

	offset := (page - 1) * pageSize

	var logs []models.FlightControlLog
	var total int64

	db := config.DB
	db.Model(&models.FlightControlLog{}).Where("uav_id = ?", uavID).Count(&total)
	db.Where("uav_id = ?", uavID).Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&logs)

	utils.RespondSuccess(c, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *RemoteCockpitHandler) HandleSDPOffer(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	sdpOffer, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无法读取SDP Offer")
		return
	}

	sdpAnswer, err := h.service.HandleSDPOffer(uavID, string(sdpOffer))
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.Data(http.StatusCreated, "application/sdp", []byte(sdpAnswer))
}

func (h *RemoteCockpitHandler) GetWebRTCStats(c *gin.Context) {
	uavIDStr := c.Param("uavId")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "无效的无人机ID")
		return
	}

	stats, err := h.service.GetWebRTCStats(uavID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, stats)
}
