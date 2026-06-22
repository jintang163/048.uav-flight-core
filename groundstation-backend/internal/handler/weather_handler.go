package handler

import (
	"net/http"
	"strconv"
	"time"

	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var weatherService = service.NewWeatherService()

func GetLatestWeather(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	data, err := weatherService.GetLatestWeather(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "气象数据不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", data)
}

func GetWeatherHistory(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

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

	data, err := weatherService.GetWeatherHistory(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", data)
}

func FetchWeatherFromAPI(c *gin.Context) {
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lon, _ := strconv.ParseFloat(c.Query("lon"), 64)

	if lat == 0 || lon == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "需要提供有效的经纬度", nil)
		return
	}

	data, err := weatherService.FetchWeatherFromAPI(lat, lon)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", data)
}

func GetWeatherAlerts(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	alertType := c.Query("alert_type")
	level := c.Query("alert_level")

	alerts, total, err := weatherService.ListAlerts(pagination, uavID, alertType, level)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", alerts, total)
}

func GetActiveWeatherAlerts(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	alerts, err := weatherService.GetActiveAlerts(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", alerts)
}

func ResolveWeatherAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的预警ID", nil)
		return
	}

	if err := weatherService.ResolveAlert(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "预警已解除", nil)
}

func CheckTakeoffWeather(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	result, err := weatherService.CheckTakeoffWeather(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "检查完成", result)
}

func GetWeatherThresholds(c *gin.Context) {
	utils.SuccessResponse(c, "获取成功", weatherService.GetThresholds())
}

func UpdateWeatherThresholds(c *gin.Context) {
	var req service.WeatherThresholds
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	weatherService.UpdateThresholds(&req)
	utils.SuccessResponse(c, "更新成功", weatherService.GetThresholds())
}

func GetFlightWeatherLog(c *gin.Context) {
	flightID, err := utils.ParseUint64(c.Param("flight_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的飞行ID", nil)
		return
	}

	log, err := weatherService.GetFlightWeatherLog(flightID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "飞行气象日志不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", log)
}

func ReportWeatherSensorData(c *gin.Context) {
	var data service.WeatherSensorData
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	if err := weatherService.ProcessSensorData(&data); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "数据已处理", nil)
}
