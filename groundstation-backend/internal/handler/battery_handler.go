package handler

import (
	"net/http"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var batteryService = service.NewBatteryService()

type CreateBatteryRequest struct {
	BatteryID    string  `json:"battery_id" binding:"required"`
	Model        string  `json:"model"`
	Manufacturer string  `json:"manufacturer"`
	Capacity     float64 `json:"capacity"`
	CapacityUnit string  `json:"capacity_unit"`
	Voltage      float64 `json:"voltage"`
	CellCount    int     `json:"cell_count"`
	Location     string  `json:"location"`
	Notes        string  `json:"notes"`
}

type UpdateBatteryRequest struct {
	Model        string  `json:"model"`
	Manufacturer string  `json:"manufacturer"`
	Capacity     float64 `json:"capacity"`
	CapacityUnit string  `json:"capacity_unit"`
	Voltage      float64 `json:"voltage"`
	CellCount    int     `json:"cell_count"`
	Status       string  `json:"status"`
	Location     string  `json:"location"`
	Notes        string  `json:"notes"`
}

type BatteryTelemetryRequest struct {
	Voltage     float64   `json:"voltage" binding:"required"`
	Level       float64   `json:"level" binding:"required"`
	Temperature float64   `json:"temperature"`
	Current     float64   `json:"current"`
	CellVoltages []float64 `json:"cell_voltages"`
}

type MaintenanceAlertResolveRequest struct {
	Note string `json:"note" binding:"required"`
}

func CreateBattery(c *gin.Context) {
	var req CreateBatteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.CreateBatteryRequest{
		BatteryID:    req.BatteryID,
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		Capacity:     req.Capacity,
		CapacityUnit: req.CapacityUnit,
		Voltage:      req.Voltage,
		CellCount:    req.CellCount,
		Location:     req.Location,
		Notes:        req.Notes,
	}

	result, err := batteryService.Create(serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetBattery(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	battery, err := batteryService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "电池不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", battery)
}

func ListBatteries(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	status := c.Query("status")
	healthStatus := c.Query("health_status")
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	keyword := c.Query("keyword")

	batteries, total, err := batteryService.List(pagination, status, healthStatus, uavID, keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", batteries, total)
}

func UpdateBattery(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	var req UpdateBatteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.UpdateBatteryRequest{
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		Capacity:     req.Capacity,
		CapacityUnit: req.CapacityUnit,
		Voltage:      req.Voltage,
		CellCount:    req.CellCount,
		Status:       req.Status,
		Location:     req.Location,
		Notes:        req.Notes,
	}

	result, err := batteryService.Update(id, serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteBattery(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	if err := batteryService.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func GetBatteryStatistics(c *gin.Context) {
	stats, err := batteryService.GetStatistics()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func GetBatteryUsageRecords(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)

	records, total, err := batteryService.GetUsageRecords(id, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", records, total)
}

func GetBatteryCellData(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	cellData, err := batteryService.GetCellData(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", cellData)
}

func UpdateBatteryTelemetry(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	var req BatteryTelemetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.BatteryTelemetryRequest{
		Voltage:     req.Voltage,
		Level:       req.Level,
		Temperature: req.Temperature,
		Current:     req.Current,
		CellVoltages: req.CellVoltages,
	}

	if err := batteryService.UpdateTelemetry(id, serviceReq); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", nil)
}

func IdentifyBattery(c *gin.Context) {
	batteryID := c.Query("battery_id")
	if batteryID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "电池ID不能为空", nil)
		return
	}

	battery, err := batteryService.IdentifyBattery(batteryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "电池不存在", nil)
		return
	}

	utils.SuccessResponse(c, "识别成功", battery)
}

func GetMaintenanceAlerts(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	batteryID, _ := utils.ParseUint64(c.Query("battery_id"))
	status := c.Query("status")
	alertType := c.Query("alert_type")
	level := c.Query("level")

	alerts, total, err := batteryService.GetMaintenanceAlerts(pagination, batteryID, status, alertType, level)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", alerts, total)
}

func AcknowledgeMaintenanceAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的告警ID", nil)
		return
	}

	userID := c.GetUint64("user_id")

	if err := batteryService.AcknowledgeMaintenanceAlert(id, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "确认成功", nil)
}

func ResolveMaintenanceAlert(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的告警ID", nil)
		return
	}

	var req MaintenanceAlertResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	userID := c.GetUint64("user_id")

	if err := batteryService.ResolveMaintenanceAlert(id, userID, req.Note); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "处理成功", nil)
}

func GetUnacknowledgedMaintenanceCount(c *gin.Context) {
	count, err := batteryService.GetUnacknowledgedMaintenanceCount()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{"count": count})
}

func CheckMaintenanceReminders(c *gin.Context) {
	maxDays, _ := utils.ParseInt(c.DefaultQuery("max_days", "7"))

	alerts, err := batteryService.CheckMaintenanceReminders(maxDays)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "检查完成", gin.H{"alerts": alerts, "count": len(alerts)})
}

func RegisterBatteryUse(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	uavID, err := utils.ParseUint64(c.Query("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "无效的无人机ID", nil)
		return
	}

	if err := batteryService.RegisterBatteryUse(id, uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "登记成功", nil)
}

func UpdateBatterySOH(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	var req struct {
		SOH float64 `json:"soh" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := batteryService.UpdateSOH(id, req.SOH); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", nil)
}
