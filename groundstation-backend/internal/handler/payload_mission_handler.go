package handler

import (
	"net/http"

	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var payloadMissionService = service.NewPayloadMissionService()

type CreateOrbitMissionRequest struct {
	UAVID            uint64  `json:"uav_id" binding:"required"`
	PayloadID        uint64  `json:"payload_id"`
	Name             string  `json:"name" binding:"required"`
	CenterLatitude   float64 `json:"center_latitude" binding:"required"`
	CenterLongitude  float64 `json:"center_longitude" binding:"required"`
	Altitude         float64 `json:"altitude"`
	Radius           float64 `json:"radius"`
	Velocity         float64 `json:"velocity"`
	YawRate          float64 `json:"yaw_rate"`
	Direction        int     `json:"direction"`
	Loops            int     `json:"loops"`
	AutoCapture      bool    `json:"auto_capture"`
	CaptureInterval  int     `json:"capture_interval_sec"`
	CameraGimbalPitch float64 `json:"camera_gimbal_pitch"`
	CameraGimbalYaw   float64 `json:"camera_gimbal_yaw"`
	Notes            string  `json:"notes"`
}

type UpdateOrbitProgressRequest struct {
	CurrentLoop int     `json:"current_loop"`
	Progress    float64 `json:"progress"`
}

func CreateOrbitMission(c *gin.Context) {
	var req CreateOrbitMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	mission := &models.OrbitMission{
		UAVID:            req.UAVID,
		PayloadID:        req.PayloadID,
		Name:             req.Name,
		CenterLatitude:   req.CenterLatitude,
		CenterLongitude:  req.CenterLongitude,
		Altitude:         req.Altitude,
		Radius:           req.Radius,
		Velocity:         req.Velocity,
		YawRate:          req.YawRate,
		Direction:        req.Direction,
		Loops:            req.Loops,
		AutoCapture:      req.AutoCapture,
		CaptureInterval:  req.CaptureInterval,
		CameraGimbalPitch: req.CameraGimbalPitch,
		CameraGimbalYaw:   req.CameraGimbalYaw,
		Notes:            req.Notes,
	}

	result, err := payloadMissionService.CreateOrbitMission(mission, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetOrbitMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	mission, err := payloadMissionService.GetOrbitMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "环绕任务不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", mission)
}

func ListOrbitMissions(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	missions, total, err := payloadMissionService.ListOrbitMissions(pagination, uavID, status, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", missions, total)
}

func StartOrbitMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	mission, err := payloadMissionService.StartOrbitMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已启动", mission)
}

func PauseOrbitMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	mission, err := payloadMissionService.PauseOrbitMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已暂停", mission)
}

func ResumeOrbitMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	mission, err := payloadMissionService.ResumeOrbitMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已恢复", mission)
}

func AbortOrbitMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	mission, err := payloadMissionService.AbortOrbitMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已中止", mission)
}

func UpdateOrbitProgress(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的环绕任务ID", nil)
		return
	}

	var req UpdateOrbitProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	mission, err := payloadMissionService.UpdateOrbitProgress(id, req.CurrentLoop, req.Progress)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "进度已更新", mission)
}

type CreateOrthoMissionRequest struct {
	UAVID          uint64  `json:"uav_id" binding:"required"`
	PayloadID      uint64  `json:"payload_id"`
	Name           string  `json:"name" binding:"required"`
	Altitude       float64 `json:"altitude"`
	Speed          float64 `json:"speed"`
	OverlapFront   float64 `json:"overlap_front"`
	OverlapSide    float64 `json:"overlap_side"`
	GSD            float64 `json:"gsd_cm"`
	CameraAngle    float64 `json:"camera_angle"`
	DirectionAngle float64 `json:"direction_angle"`
	TriggerMode    string  `json:"trigger_mode"`
	TriggerDistance float64 `json:"trigger_distance_m"`
	TriggerInterval int     `json:"trigger_interval_sec"`
	ReturnToHome   bool    `json:"return_to_home"`
	Notes          string  `json:"notes"`
}

type PlanOrthoMissionRequest struct {
	SurveyArea [][]float64 `json:"survey_area" binding:"required"`
}

func CreateOrthoMission(c *gin.Context) {
	var req CreateOrthoMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	mission := &models.OrthoMission{
		UAVID:          req.UAVID,
		PayloadID:      req.PayloadID,
		Name:           req.Name,
		Altitude:       req.Altitude,
		Speed:          req.Speed,
		OverlapFront:   req.OverlapFront,
		OverlapSide:    req.OverlapSide,
		GSD:            req.GSD,
		CameraAngle:    req.CameraAngle,
		DirectionAngle: req.DirectionAngle,
		TriggerMode:    req.TriggerMode,
		TriggerDistance: req.TriggerDistance,
		TriggerInterval: req.TriggerInterval,
		ReturnToHome:   req.ReturnToHome,
		Notes:          req.Notes,
	}

	result, err := payloadMissionService.CreateOrthoMission(mission, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func PlanOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	var req PlanOrthoMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	mission, err := payloadMissionService.PlanOrthoMission(id, req.SurveyArea)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "航线规划完成", mission)
}

func GetOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	mission, err := payloadMissionService.GetOrthoMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "正射任务不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", mission)
}

func ListOrthoMissions(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	missions, total, err := payloadMissionService.ListOrthoMissions(pagination, uavID, status, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", missions, total)
}

func StartOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	mission, err := payloadMissionService.StartOrthoMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已启动", mission)
}

func PauseOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	mission, err := payloadMissionService.PauseOrthoMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已暂停", mission)
}

func ResumeOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	mission, err := payloadMissionService.ResumeOrthoMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已恢复", mission)
}

func AbortOrthoMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的正射任务ID", nil)
		return
	}

	mission, err := payloadMissionService.AbortOrthoMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已中止", mission)
}

type CreateTTSRequest struct {
	UAVID     uint64  `json:"uav_id" binding:"required"`
	PayloadID uint64  `json:"payload_id" binding:"required"`
	Text      string  `json:"text" binding:"required"`
	Language  string  `json:"language"`
	Voice     string  `json:"voice"`
	Speed     float64 `json:"speed"`
	Pitch     float64 `json:"pitch"`
	Volume    int     `json:"volume"`
}

func CreateTTS(c *gin.Context) {
	var req CreateTTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	task := &models.TextToSpeechTask{
		UAVID:     req.UAVID,
		PayloadID: req.PayloadID,
		Text:      req.Text,
		Language:  req.Language,
		Voice:     req.Voice,
		Speed:     req.Speed,
		Pitch:     req.Pitch,
		Volume:    req.Volume,
	}

	result, err := payloadMissionService.CreateTTSTask(task, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "TTS任务已创建", result)
}

func GetTTSTask(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的TTS任务ID", nil)
		return
	}

	task, err := payloadMissionService.GetTTSTask(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "TTS任务不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", task)
}

func ListTTSTasks(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	payloadID, _ := utils.ParseUint64(c.Query("payload_id"))
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	status := c.Query("status")

	tasks, total, err := payloadMissionService.ListTTSTasks(pagination, payloadID, uavID, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", tasks, total)
}
