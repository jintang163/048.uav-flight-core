package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var geofenceService = service.NewGeofenceService()

type CreateGeofenceRequest struct {
	Name        string               `json:"name" binding:"required"`
	Type        models.GeofenceType  `json:"type" binding:"required,oneof=inclusion exclusion"`
	Shape       models.GeofenceShape `json:"shape" binding:"required,oneof=polygon circle rectangle"`
	CenterLat   float64              `json:"center_lat"`
	CenterLng   float64              `json:"center_lng"`
	Radius      float64              `json:"radius"`
	MinAltitude float64              `json:"min_altitude"`
	MaxAltitude float64              `json:"max_altitude"`
	Coordinates [][]float64          `json:"coordinates"`
	UAVIDs      []uint64             `json:"uav_ids"`
}

type UpdateGeofenceRequest struct {
	Name        string               `json:"name"`
	Type        models.GeofenceType  `json:"type" binding:"omitempty,oneof=inclusion exclusion"`
	Shape       models.GeofenceShape `json:"shape" binding:"omitempty,oneof=polygon circle rectangle"`
	CenterLat   float64              `json:"center_lat"`
	CenterLng   float64              `json:"center_lng"`
	Radius      float64              `json:"radius"`
	MinAltitude float64              `json:"min_altitude"`
	MaxAltitude float64              `json:"max_altitude"`
	Coordinates [][]float64          `json:"coordinates"`
	IsActive    *bool                `json:"is_active"`
	UAVIDs      []uint64             `json:"uav_ids"`
}

func CreateGeofence(c *gin.Context) {
	var req CreateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	geofence := &models.Geofence{
		Name:        req.Name,
		Type:        req.Type,
		Shape:       req.Shape,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		Radius:      req.Radius,
		MinAltitude: req.MinAltitude,
		MaxAltitude: req.MaxAltitude,
		IsActive:    true,
		CreatorID:   userID,
	}

	result, err := geofenceService.Create(geofence, req.Coordinates, req.UAVIDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetGeofence(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的围栏ID", nil)
		return
	}

	geofence, err := geofenceService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "围栏不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", geofence)
}

func ListGeofences(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	gfType := c.Query("type")
	isActive := c.Query("is_active")

	geofences, total, err := geofenceService.List(pagination, gfType, isActive)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", geofences, total)
}

func UpdateGeofence(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的围栏ID", nil)
		return
	}

	var req UpdateGeofenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	geofence := &models.Geofence{
		Name:        req.Name,
		Type:        req.Type,
		Shape:       req.Shape,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		Radius:      req.Radius,
		MinAltitude: req.MinAltitude,
		MaxAltitude: req.MaxAltitude,
	}
	if req.IsActive != nil {
		geofence.IsActive = *req.IsActive
	}

	result, err := geofenceService.Update(id, geofence, req.Coordinates, req.UAVIDs)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteGeofence(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的围栏ID", nil)
		return
	}

	if err := geofenceService.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func CheckViolation(c *gin.Context) {
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

func GetUAVGeofences(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	geofences, err := geofenceService.GetByUAV(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", geofences)
}
