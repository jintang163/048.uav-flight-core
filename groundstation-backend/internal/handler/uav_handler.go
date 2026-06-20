package handler

import (
	"net/http"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var uavService = service.NewUAVService()

type CreateUAVRequest struct {
	Name          string             `json:"name" binding:"required"`
	SerialNumber  string             `json:"serial_number" binding:"required"`
	Model         string             `json:"model" binding:"required"`
	Type          models.UAVType     `json:"type" binding:"required"`
	MaxAltitude   float64            `json:"max_altitude"`
	MaxSpeed      float64            `json:"max_speed"`
	MaxFlightTime int                `json:"max_flight_time"`
}

type UpdateUAVRequest struct {
	Name          string             `json:"name"`
	Model         string             `json:"model"`
	Type          models.UAVType     `json:"type"`
	MaxAltitude   float64            `json:"max_altitude"`
	MaxSpeed      float64            `json:"max_speed"`
	MaxFlightTime int                `json:"max_flight_time"`
	Status        models.UAVStatus   `json:"status"`
	StatusMessage string             `json:"status_message"`
}

func CreateUAV(c *gin.Context) {
	var req CreateUAVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	uav := &models.UAV{
		Name:          req.Name,
		SerialNumber:  req.SerialNumber,
		Model:         req.Model,
		Type:          req.Type,
		Status:        models.UAVStatusOffline,
		MaxAltitude:   req.MaxAltitude,
		MaxSpeed:      req.MaxSpeed,
		MaxFlightTime: req.MaxFlightTime,
	}

	result, err := uavService.Create(uav)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetUAV(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	uav, err := uavService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "无人机不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", uav)
}

func ListUAVs(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	status := c.Query("status")
	uavType := c.Query("type")
	keyword := c.Query("keyword")

	uavs, total, err := uavService.List(pagination, status, uavType, keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", uavs, total)
}

func UpdateUAV(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	var req UpdateUAVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	uav := &models.UAV{
		Name:          req.Name,
		Model:         req.Model,
		Type:          req.Type,
		MaxAltitude:   req.MaxAltitude,
		MaxSpeed:      req.MaxSpeed,
		MaxFlightTime: req.MaxFlightTime,
		Status:        req.Status,
		StatusMessage: req.StatusMessage,
	}

	result, err := uavService.Update(id, uav)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteUAV(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	if err := uavService.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func GetUAVStatus(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	status, err := uavService.GetFlightStatus(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "飞行状态不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", status)
}

func GetUAVStatistics(c *gin.Context) {
	stats, err := uavService.GetStatistics()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func GetUAVFlightHistory(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	history, total, err := uavService.GetFlightHistory(id, pagination, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", history, total)
}
