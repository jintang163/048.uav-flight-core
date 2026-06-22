package handler

import (
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateLandingPoint(c *gin.Context) {
	var point models.LandingPoint
	if err := c.ShouldBindJSON(&point); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	created, err := landingService.CreateLandingPoint(&point)
	if err != nil {
		middleware.Logger.Error("Failed to create landing point", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落点创建成功", created)
}

func GetLandingPoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	landingService := service.NewLandingService()
	point, err := landingService.GetLandingPoint(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "landing point not found", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", point)
}

func ListLandingPoints(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	pointType := c.Query("type")
	status := c.Query("status")

	var hasMarkers *bool
	if hasMarkersStr := c.Query("has_markers"); hasMarkersStr != "" {
		val, _ := strconv.ParseBool(hasMarkersStr)
		hasMarkers = &val
	}

	landingService := service.NewLandingService()
	points, total, err := landingService.ListLandingPoints(page, pageSize, pointType, status, hasMarkers)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{
		"list":  points,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func UpdateLandingPoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	var point models.LandingPoint
	if err := c.ShouldBindJSON(&point); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	updated, err := landingService.UpdateLandingPoint(id, &point)
	if err != nil {
		middleware.Logger.Error("Failed to update landing point", zap.Error(err), zap.Uint64("id", id))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落点更新成功", updated)
}

func DeleteLandingPoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	landingService := service.NewLandingService()
	if err := landingService.DeleteLandingPoint(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落点删除成功", nil)
}

type PlanLandingRequest struct {
	PrimaryLandingID   uint64   `json:"primary_landing_id"`
	AlternateLandingIDs []uint64 `json:"alternate_landing_ids"`
	MissionID          uint64   `json:"mission_id"`
	RTKEnabled         bool     `json:"rtk_enabled"`
	VisionEnabled      bool     `json:"vision_enabled"`
	MovingPlatform     bool     `json:"moving_platform"`
}

func PlanLanding(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req PlanLandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()

	var primaryLanding *models.LandingPoint
	if req.PrimaryLandingID > 0 {
		primaryLanding, err = landingService.GetLandingPoint(req.PrimaryLandingID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, 404001, "primary landing point not found", nil)
			return
		}
	}

	var alternateLandings []*models.LandingPoint
	for _, altID := range req.AlternateLandingIDs {
		alt, err := landingService.GetLandingPoint(altID)
		if err == nil {
			alternateLandings = append(alternateLandings, alt)
		}
	}

	plan := &service.LandingPlan{
		PrimaryLanding:   primaryLanding,
		AlternateLandings: alternateLandings,
		MissionID:        req.MissionID,
		RTKEnabled:       req.RTKEnabled,
		VisionEnabled:    req.VisionEnabled,
		MovingPlatform:   req.MovingPlatform,
	}

	session, err := landingService.PlanLanding(uavID, plan)
	if err != nil {
		middleware.Logger.Error("Failed to plan landing", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落计划创建成功", session)
}

func StartLanding(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid session_id", nil)
		return
	}

	landingService := service.NewLandingService()
	session, err := landingService.StartLanding(uavID, sessionID)
	if err != nil {
		middleware.Logger.Error("Failed to start landing", zap.Error(err), zap.Uint64("uav_id", uavID), zap.Uint64("session_id", sessionID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落已启动", session)
}

func GetActiveLandingSession(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	landingService := service.NewLandingService()
	session, err := landingService.GetActiveLandingSession(uavID)
	if err != nil {
		utils.SuccessResponse(c, "获取成功", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", session)
}

func GetLandingSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	landingService := service.NewLandingService()
	session, err := landingService.GetLandingSession(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "landing session not found", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", session)
}

func ListLandingSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	uavID, _ := strconv.ParseUint(c.Query("uav_id"), 10, 64)
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	landingService := service.NewLandingService()
	sessions, total, err := landingService.ListLandingSessions(page, pageSize, uavID, status, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{
		"list":  sessions,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetLandingTrajectory(c *gin.Context) {
	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid session_id", nil)
		return
	}

	landingService := service.NewLandingService()
	trajectory, err := landingService.GetLandingTrajectory(sessionID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{
		"session_id": sessionID,
		"trajectory": trajectory,
		"count":      len(trajectory),
	})
}

type AbortLandingRequest struct {
	Reason string `json:"reason"`
}

func AbortLanding(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid session_id", nil)
		return
	}

	var req AbortLandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "manual abort"
	}

	landingService := service.NewLandingService()
	session, err := landingService.AbortLanding(uavID, sessionID, req.Reason)
	if err != nil {
		middleware.Logger.Error("Failed to abort landing", zap.Error(err), zap.Uint64("uav_id", uavID), zap.Uint64("session_id", sessionID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "降落已中止", session)
}

type SwitchAlternateLandingRequest struct {
	AlternateLandingID uint64 `json:"alternate_landing_id" binding:"required"`
}

func SwitchToAlternateLanding(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid session_id", nil)
		return
	}

	var req SwitchAlternateLandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	session, err := landingService.SwitchToAlternateLanding(uavID, sessionID, req.AlternateLandingID)
	if err != nil {
		middleware.Logger.Error("Failed to switch to alternate landing", zap.Error(err), zap.Uint64("uav_id", uavID), zap.Uint64("session_id", sessionID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已切换至备降点", session)
}

func GetLandingStatistics(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uav_id"), 10, 64)
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	landingService := service.NewLandingService()
	stats, err := landingService.GetLandingStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

type ForcedLandingRequest struct {
	TriggerType string `json:"trigger_type" binding:"required"`
	Reason      string `json:"reason"`
	LockArms    bool   `json:"lock_arms"`
}

func TriggerForcedLanding(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req ForcedLandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	event, err := landingService.TriggerForcedLanding(uavID, req.TriggerType, req.Reason, req.LockArms)
	if err != nil {
		middleware.Logger.Error("Failed to trigger forced landing", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "强制降落已触发", event)
}

func GetActiveForcedLandingEvent(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	landingService := service.NewLandingService()
	event, err := landingService.GetActiveForcedLandingEvent(uavID)
	if err != nil {
		utils.SuccessResponse(c, "获取成功", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", event)
}

func ListForcedLandingEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	uavID, _ := strconv.ParseUint(c.Query("uav_id"), 10, 64)
	triggerType := c.Query("trigger_type")

	var isResolved *bool
	if resolvedStr := c.Query("is_resolved"); resolvedStr != "" {
		val, _ := strconv.ParseBool(resolvedStr)
		isResolved = &val
	}

	landingService := service.NewLandingService()
	events, total, err := landingService.ListForcedLandingEvents(page, pageSize, uavID, triggerType, isResolved)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{
		"list":  events,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

func GetForcedLandingEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	landingService := service.NewLandingService()
	event, err := landingService.GetForcedLandingEvent(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "forced landing event not found", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", event)
}

type ResolveForcedLandingRequest struct {
	Notes string `json:"notes"`
}

func ResolveForcedLanding(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid id", nil)
		return
	}

	var req ResolveForcedLandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	userID, _ := c.Get("user_id")
	resolvedBy, _ := userID.(uint64)

	landingService := service.NewLandingService()
	event, err := landingService.ResolveForcedLanding(id, resolvedBy, req.Notes)
	if err != nil {
		middleware.Logger.Error("Failed to resolve forced landing", zap.Error(err), zap.Uint64("id", id))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "强制降落事件已解决", event)
}

type UpdateVisionDataRequest struct {
	MarkerDetected bool    `json:"marker_detected"`
	MarkerType     string  `json:"marker_type"`
	MarkerID       string  `json:"marker_id"`
	Confidence     float64 `json:"confidence"`
	OffsetX        float64 `json:"offset_x"`
	OffsetY        float64 `json:"offset_y"`
	OffsetZ        float64 `json:"offset_z"`
	YawError       float64 `json:"yaw_error"`
	CameraHeight   float64 `json:"camera_height"`
	ImageWidth     int     `json:"image_width"`
	ImageHeight    int     `json:"image_height"`
}

func UpdateVisionLandingData(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdateVisionDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	visionData := &models.VisionLandingData{
		MarkerDetected: req.MarkerDetected,
		MarkerType:     req.MarkerType,
		MarkerID:       req.MarkerID,
		Confidence:     req.Confidence,
		OffsetX:        req.OffsetX,
		OffsetY:        req.OffsetY,
		OffsetZ:        req.OffsetZ,
		YawError:       req.YawError,
		CameraHeight:   req.CameraHeight,
		ImageWidth:     req.ImageWidth,
		ImageHeight:    req.ImageHeight,
	}

	landingService := service.NewLandingService()
	if err := landingService.UpdateVisionData(uavID, visionData); err != nil {
		middleware.Logger.Error("Failed to update vision landing data", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "视觉着陆数据已更新", nil)
}

type UpdateRTKDataRequest struct {
	Latitude        float64 `json:"latitude" binding:"required"`
	Longitude       float64 `json:"longitude" binding:"required"`
	Altitude        float64 `json:"altitude"`
	FixType         int     `json:"fix_type" binding:"required"`
	Satellites      int     `json:"satellites"`
	HorizontalAcc   float64 `json:"horizontal_accuracy"`
	VerticalAcc     float64 `json:"vertical_accuracy"`
	HDOP            float64 `json:"hdop"`
	VDOP            float64 `json:"vdop"`
	BaseStationID   string  `json:"base_station_id"`
	DifferentialAge float64 `json:"differential_age"`
}

func UpdateRTKPositionData(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdateRTKDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	rtkData := &models.RTKPositionData{
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
		Altitude:        req.Altitude,
		FixType:         req.FixType,
		Satellites:      req.Satellites,
		HorizontalAcc:   req.HorizontalAcc,
		VerticalAcc:     req.VerticalAcc,
		HDOP:            req.HDOP,
		VDOP:            req.VDOP,
		BaseStationID:   req.BaseStationID,
		DifferentialAge: req.DifferentialAge,
	}

	landingService := service.NewLandingService()
	if err := landingService.UpdateRTKData(uavID, rtkData); err != nil {
		middleware.Logger.Error("Failed to update RTK position data", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "RTK定位数据已更新", nil)
}

type UpdateMovingPlatformRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	VelocityX float64 `json:"velocity_x"`
	VelocityY float64 `json:"velocity_y"`
}

func UpdateMovingPlatformPosition(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdateMovingPlatformRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	if err := landingService.UpdateMovingPlatformPosition(uavID, req.Latitude, req.Longitude, req.VelocityX, req.VelocityY); err != nil {
		middleware.Logger.Error("Failed to update moving platform position", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "移动平台位置已更新", nil)
}

func RecordTrajectoryPoint(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var telemetry models.FlightStatus
	if err := c.ShouldBindJSON(&telemetry); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body: "+err.Error(), nil)
		return
	}

	landingService := service.NewLandingService()
	if err := landingService.RecordTrajectoryPoint(uavID, &telemetry); err != nil {
		middleware.Logger.Error("Failed to record trajectory point", zap.Error(err), zap.Uint64("uav_id", uavID))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "航迹点已记录", nil)
}
