package handler

import (
	"net/http"

	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var violationService = service.NewGeofenceViolationService()

func ListViolations(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	geofenceID, _ := utils.ParseUint64(c.Query("geofence_id"))
	severity := c.Query("severity")
	violationType := c.Query("violation_type")
	isResolvedStr := c.Query("is_resolved")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var isResolved *bool
	if isResolvedStr != "" {
		b := isResolvedStr == "true" || isResolvedStr == "1"
		isResolved = &b
	}

	logs, total, err := violationService.List(pagination, uavID, geofenceID, severity, violationType, isResolved, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", logs, total)
}

func GetViolation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	log, err := violationService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", log)
}

func ResolveViolation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	var data struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&data)

	if err := violationService.Resolve(id, data.Notes); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已标记为已处理", nil)
}

func GetViolationStatistics(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	geofenceID, _ := utils.ParseUint64(c.Query("geofence_id"))
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	stats, err := violationService.GetStatistics(uavID, geofenceID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func BatchResolveViolations(c *gin.Context) {
	var data struct {
		IDs   []uint64 `json:"ids" binding:"required"`
		Notes string   `json:"notes"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	count := 0
	for _, id := range data.IDs {
		if err := violationService.Resolve(id, data.Notes); err == nil {
			count++
		}
	}

	utils.SuccessResponse(c, "批量处理完成", gin.H{
		"success_count": count,
		"total":         len(data.IDs),
	})
}

func CheckGeofenceViolation(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	lat, _ := utils.ParseFloat64(c.Query("lat"))
	lng, _ := utils.ParseFloat64(c.Query("lng"))
	altitude, _ := utils.ParseFloat64(c.Query("altitude"))

	violations, err := geofenceService.CheckViolation(uavID, lat, lng, altitude)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "检测完成", gin.H{
		"has_violation": len(violations) > 0,
		"violations":    violations,
	})
}
