package handler

import (
	"net/http"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var flightService = service.NewFlightService()

func GetLatestFlightStatus(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	status, err := flightService.GetLatestStatus(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "飞行状态不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", status)
}

func GetFlightHistory(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	history, total, err := flightService.GetHistory(uavID, pagination, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", history, total)
}

func GetRealTimeTelemetry(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	data, err := flightService.GetRealtimeData(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "实时数据不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", data)
}

func GetAllUAVsRealtime(c *gin.Context) {
	data, err := flightService.GetAllRealtimeData()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", data)
}

func GetFlightStatistics(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	stats, err := flightService.GetStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func GetFlightTrack(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	track, err := flightService.GetTrack(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", track)
}
