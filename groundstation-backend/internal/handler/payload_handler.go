package handler

import (
	"net/http"

	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var payloadService = service.NewPayloadService()

type CreatePayloadRequest struct {
	UAVID       uint64              `json:"uav_id" binding:"required"`
	Type        models.PayloadType  `json:"type" binding:"required"`
	Name        string              `json:"name" binding:"required"`
	Model       string              `json:"model"`
	Port        int                 `json:"port"`
	Slot        int                 `json:"slot"`
	Description string              `json:"description"`
	Config      string              `json:"config"`
}

type UpdatePayloadRequest struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	Description string `json:"description"`
	Config      string `json:"config"`
	Port        int    `json:"port"`
	Status      string `json:"status"`
}

type CameraSettingsRequest struct {
	Resolution   string  `json:"resolution"`
	FrameRate    int     `json:"frame_rate"`
	ZoomLevel    float64 `json:"zoom_level"`
	ISO          int     `json:"iso"`
	ShutterSpeed string  `json:"shutter_speed"`
}

type SprayerFlowRequest struct {
	FlowRate float64 `json:"flow_rate_lpm" binding:"required"`
}

func CreatePayload(c *gin.Context) {
	var req CreatePayloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	payload := &models.PayloadDevice{
		UAVID:       req.UAVID,
		Type:        req.Type,
		Name:        req.Name,
		Model:       req.Model,
		Port:        req.Port,
		Slot:        req.Slot,
		Description: req.Description,
		Config:      req.Config,
	}

	result, err := payloadService.CreatePayload(payload)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetPayload(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	payload, err := payloadService.GetPayload(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "载荷不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", payload)
}

func ListPayloads(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	payloadType := c.Query("type")
	status := c.Query("status")
	keyword := c.Query("keyword")

	payloads, total, err := payloadService.ListPayloads(pagination, uavID, payloadType, status, keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", payloads, total)
}

func ListUAVPayloads(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	payloads, err := payloadService.ListPayloadsByUAV(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", payloads)
}

func UpdatePayload(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req UpdatePayloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	payload := &models.PayloadDevice{
		Name:        req.Name,
		Model:       req.Model,
		Description: req.Description,
		Config:      req.Config,
		Port:        req.Port,
	}

	if req.Status != "" {
		payload.Status = models.PayloadStatus(req.Status)
	}

	result, err := payloadService.UpdatePayload(id, payload)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeletePayload(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	if err := payloadService.DeletePayload(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func GetCameraStatus(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	status, err := payloadService.GetCameraStatus(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "相机状态不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", status)
}

func TakePhoto(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	userID := middleware.GetCurrentUserID(c)
	_ = userID

	if err := payloadService.TakePhoto(uavID, id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "拍照命令已发送", nil)
}

func StartVideoRecording(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.StartVideoRecording(uavID, id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "录像已开始", nil)
}

func StopVideoRecording(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.StopVideoRecording(uavID, id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "录像已停止", nil)
}

func SetCameraMode(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req struct {
		Mode models.CameraMode `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.SetCameraMode(uavID, id, req.Mode); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "模式已设置", nil)
}

func SetCameraZoom(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req struct {
		ZoomLevel float64 `json:"zoom_level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.SetCameraZoom(uavID, id, req.ZoomLevel); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "变焦已设置", nil)
}

func SetCameraSettings(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req CameraSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	settings := map[string]interface{}{
		"resolution":    req.Resolution,
		"frame_rate":    req.FrameRate,
		"zoom_level":    req.ZoomLevel,
		"iso":           req.ISO,
		"shutter_speed": req.ShutterSpeed,
	}

	if err := payloadService.SetCameraSettings(uavID, id, settings); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	if req.ZoomLevel > 0 {
		_ = payloadService.SetCameraZoom(uavID, id, req.ZoomLevel)
	}

	utils.SuccessResponse(c, "相机设置已更新", nil)
}

func GetSprayerStatus(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	status, err := payloadService.GetSprayerStatus(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "喷药器状态不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", status)
}

func SetSprayerFlowRate(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req SprayerFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.SetSprayerFlowRate(uavID, id, req.FlowRate); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "流量已设置", nil)
}

func StartSpraying(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	var req struct {
		FlowRate float64 `json:"flow_rate_lpm"`
	}
	c.ShouldBindJSON(&req)

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.StartSpraying(uavID, id, req.FlowRate); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "喷药已开始", nil)
}

func StopSpraying(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.StopSpraying(uavID, id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "喷药已停止", nil)
}

type CreateSpeakerAudioRequest struct {
	PayloadID      uint64  `json:"payload_id" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Type           string  `json:"type"`
	Content        string  `json:"content"`
	DurationSec    int     `json:"duration_sec"`
	IsTextToSpeech bool    `json:"is_text_to_speech"`
	Voice          string  `json:"voice"`
	Speed          float64 `json:"speed"`
	Pitch          float64 `json:"pitch"`
	Volume         int     `json:"volume"`
}

func CreateSpeakerAudio(c *gin.Context) {
	var req CreateSpeakerAudioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	audio := &models.SpeakerAudio{
		PayloadID:      req.PayloadID,
		Name:           req.Name,
		Type:           req.Type,
		Content:        req.Content,
		DurationSec:    req.DurationSec,
		IsTextToSpeech: req.IsTextToSpeech,
		Voice:          req.Voice,
		Speed:          req.Speed,
		Pitch:          req.Pitch,
		Volume:         req.Volume,
	}

	result, err := payloadService.CreateSpeakerAudio(audio, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func ListSpeakerAudios(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	payloadID, _ := utils.ParseUint64(c.Query("payload_id"))
	isTTSStr := c.Query("is_tts")

	var isTTS *bool
	if isTTSStr != "" {
		val := isTTSStr == "true" || isTTSStr == "1"
		isTTS = &val
	}

	audios, total, err := payloadService.ListSpeakerAudios(pagination, payloadID, isTTS)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", audios, total)
}

func GetSpeakerAudio(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的音频ID", nil)
		return
	}

	audio, err := payloadService.GetSpeakerAudio(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "音频不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", audio)
}

func DeleteSpeakerAudio(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的音频ID", nil)
		return
	}

	if err := payloadService.DeleteSpeakerAudio(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func PlaySpeakerAudio(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	audioID, err := utils.ParseUint64(c.Param("audio_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的音频ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.PlaySpeakerAudio(uavID, id, audioID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "播放命令已发送", nil)
}

func StopSpeaker(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的载荷ID", nil)
		return
	}

	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	if uavID == 0 {
		payload, _ := payloadService.GetPayload(id)
		if payload != nil {
			uavID = payload.UAVID
		}
	}

	if err := payloadService.StopSpeaker(uavID, id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "播放已停止", nil)
}
