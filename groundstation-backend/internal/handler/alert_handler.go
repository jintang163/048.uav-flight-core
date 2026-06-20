package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var alertService = service.NewAlertService()

type AcknowledgeRequest struct {
	Note string `json:"note"`
}

type ResolveRequest struct {
	ResolutionNote string `json:"resolution_note" binding:"required"`
}

type BatchAcknowledgeRequest struct {
	AlertIDs []uint64 `json:"alert_ids" binding:"required,min=1"`
}

func GetAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的告警ID", nil)
		return
	}

	alert, err := alertService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "告警不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", alert)
}

func ListAlerts(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	level := c.Query("level")
	alertType := c.Query("type")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	alerts, total, err := alertService.List(pagination, uavID, level, alertType, status, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", alerts, total)
}

func AcknowledgeAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的告警ID", nil)
		return
	}

	var req AcknowledgeRequest
	c.ShouldBindJSON(&req)

	userID := middleware.GetCurrentUserID(c)
	alert, err := alertService.Acknowledge(id, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已确认", alert)
}

func ResolveAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的告警ID", nil)
		return
	}

	var req ResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	alert, err := alertService.Resolve(id, userID, req.ResolutionNote)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已解决", alert)
}

func BatchAcknowledge(c *gin.Context) {
	var req BatchAcknowledgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	count, err := alertService.BatchAcknowledge(req.AlertIDs, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "批量确认成功", gin.H{"acknowledged": count})
}

func GetAlertStatistics(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	stats, err := alertService.GetStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func SendTestNotification(c *gin.Context) {
	var data struct {
		Phone  string `json:"phone"`
		Email  string `json:"email"`
		Title  string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	err := alertService.SendTestNotification(data.Phone, data.Email, data.Title, data.Message)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "测试通知已发送", nil)
}
