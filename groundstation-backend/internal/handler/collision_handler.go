package handler

import (
	"net/http"
	"strconv"
	"time"

	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var collisionAvoidService = service.NewCollisionAvoidanceService()

func GetActiveCollisionAlerts(c *gin.Context) {
	alerts, err := collisionAvoidService.GetActiveAlerts()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "获取成功", alerts)
}

func GetAllUAVPositions(c *gin.Context) {
	positions := collisionAvoidService.GetAllPositions()
	utils.SuccessResponse(c, "获取成功", positions)
}

func GetRouteIntersections(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	intersections, err := collisionAvoidService.GetActiveIntersections(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "获取成功", intersections)
}

func DetectRouteIntersections(c *gin.Context) {
	intersections, err := collisionAvoidService.DetectRouteIntersections()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "检测完成", map[string]interface{}{
		"intersections": intersections,
		"count":         len(intersections),
	})
}

func ResolveCollisionAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的预警ID", nil)
		return
	}
	if err := collisionAvoidService.ResolveAlert(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "预警已解除", nil)
}

func GetCollisionAvoidanceStatus(c *gin.Context) {
	positions := collisionAvoidService.GetAllPositions()
	alerts, _ := collisionAvoidService.GetActiveAlerts()
	intersections, _ := collisionAvoidService.GetActiveIntersections(0)

	enabled := false
	if len(positions) > 0 {
		enabled = true
	}

	status := map[string]interface{}{
		"enabled":            enabled,
		"active_uavs":        len(positions),
		"active_alerts":      len(alerts),
		"intersections":      len(intersections),
		"safe_distance_m":    50.0,
		"warning_distance_m": 100.0,
	}
	utils.SuccessResponse(c, "获取成功", status)
}

func ToggleCollisionAvoidance(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}
	collisionAvoidService.SetEnabled(req.Enabled)
	utils.SuccessResponse(c, "操作成功", map[string]interface{}{
		"enabled": req.Enabled,
	})
}

func ManualCollisionAvoidance(c *gin.Context) {
	var req struct {
		UAVID  uint64  `json:"uav_id" binding:"required"`
		Action string  `json:"action" binding:"required"`
		Param  float64 `json:"param"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}
	if err := collisionAvoidService.ManualAvoidance(req.UAVID, req.Action, req.Param); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "指令已下发", nil)
}

func ListCollisionAlerts(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	riskLevel := c.Query("risk_level")
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	resolved := c.Query("resolved") == "true"

	repo := repository.NewCollisionRepository()
	alerts, total, err := repo.ListAlerts(pagination, riskLevel, uavID, resolved)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponseWithTotal(c, "获取成功", alerts, total)
}

func GetUAVSpeedFactor(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}
	factor := collisionAvoidService.GetSpeedFactor(uavID)
	utils.SuccessResponse(c, "获取成功", map[string]interface{}{
		"uav_id":       uavID,
		"speed_factor": factor,
	})
}

func GetCollisionStats(c *gin.Context) {
	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	if s := c.Query("start_time"); s != "" {
		if t, err := strconv.ParseInt(s, 10, 64); err == nil {
			startTime = time.Unix(t/1000, 0)
		}
	}
	if e := c.Query("end_time"); e != "" {
		if t, err := strconv.ParseInt(e, 10, 64); err == nil {
			endTime = time.Unix(t/1000, 0)
		}
	}

	repo := repository.NewCollisionRepository()
	stats, err := repo.GetAlertStats(startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, "获取成功", stats)
}
