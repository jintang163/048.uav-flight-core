package handler

import (
	"net/http"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var chargingService = service.NewChargingService()

type CreateChargingStationRequest struct {
	StationID     string  `json:"station_id" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Model         string  `json:"model"`
	Manufacturer  string  `json:"manufacturer"`
	SlotCount     int     `json:"slot_count"`
	Location      string  `json:"location"`
	IPAddress     string  `json:"ip_address"`
	Port          int     `json:"port"`
	Protocol      string  `json:"protocol"`
	MaxVoltage    float64 `json:"max_voltage"`
	MaxCurrent    float64 `json:"max_current"`
	Description   string  `json:"description"`
}

type UpdateChargingStationRequest struct {
	Name          string  `json:"name"`
	Model         string  `json:"model"`
	Manufacturer  string  `json:"manufacturer"`
	SlotCount     int     `json:"slot_count"`
	Location      string  `json:"location"`
	IPAddress     string  `json:"ip_address"`
	Port          int     `json:"port"`
	Protocol      string  `json:"protocol"`
	MaxVoltage    float64 `json:"max_voltage"`
	MaxCurrent    float64 `json:"max_current"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
}

type StartChargingRequest struct {
	BatteryID     uint64  `json:"battery_id" binding:"required"`
	ChargingMode  string  `json:"charging_mode"`
	TargetVoltage float64 `json:"target_voltage"`
	TargetCurrent float64 `json:"target_current"`
}

type SlotTelemetryRequest struct {
	Voltage          float64 `json:"voltage"`
	Current          float64 `json:"current"`
	Level            float64 `json:"level"`
	Temperature      float64 `json:"temperature"`
	ChargedCapacity  float64 `json:"charged_capacity"`
	ChargingTime     int     `json:"charging_time"`
	RemainingTime    int     `json:"remaining_time"`
}

type StationHeartbeatRequest struct {
	Status          string `json:"status"`
	FirmwareVersion string `json:"firmware_version"`
	OccupiedSlots   int    `json:"occupied_slots"`
	ChargingSlots   int    `json:"charging_slots"`
}

type AssignBatteryRequest struct {
	BatteryID uint64 `json:"battery_id" binding:"required"`
}

func CreateChargingStation(c *gin.Context) {
	var req CreateChargingStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.CreateChargingStationRequest{
		StationID:     req.StationID,
		Name:          req.Name,
		Model:         req.Model,
		Manufacturer:  req.Manufacturer,
		SlotCount:     req.SlotCount,
		Location:      req.Location,
		IPAddress:     req.IPAddress,
		Port:          req.Port,
		Protocol:      req.Protocol,
		MaxVoltage:    req.MaxVoltage,
		MaxCurrent:    req.MaxCurrent,
		Description:   req.Description,
	}

	result, err := chargingService.CreateStation(serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetChargingStation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	station, err := chargingService.GetStationByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "充电柜不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", station)
}

func ListChargingStations(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	status := c.Query("status")
	keyword := c.Query("keyword")

	stations, total, err := chargingService.ListStations(pagination, status, keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", stations, total)
}

func UpdateChargingStation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	var req UpdateChargingStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.UpdateChargingStationRequest{
		Name:          req.Name,
		Model:         req.Model,
		Manufacturer:  req.Manufacturer,
		SlotCount:     req.SlotCount,
		Location:      req.Location,
		IPAddress:     req.IPAddress,
		Port:          req.Port,
		Protocol:      req.Protocol,
		MaxVoltage:    req.MaxVoltage,
		MaxCurrent:    req.MaxCurrent,
		Description:   req.Description,
		Status:        req.Status,
	}

	result, err := chargingService.UpdateStation(id, serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteChargingStation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	if err := chargingService.DeleteStation(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func GetChargingStationSlots(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	slots, err := chargingService.GetStationSlots(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", slots)
}

func GetChargingSlot(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	slot, err := chargingService.GetSlotByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "充电位不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", slot)
}

func StartCharging(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	var req StartChargingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.StartChargingRequest{
		BatteryID:     req.BatteryID,
		ChargingMode:  req.ChargingMode,
		TargetVoltage: req.TargetVoltage,
		TargetCurrent: req.TargetCurrent,
	}

	record, err := chargingService.StartCharging(slotID, serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "开始充电", record)
}

func StopCharging(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	var req struct {
		EndLevel float64 `json:"end_level"`
	}
	_ = c.ShouldBindJSON(&req)

	record, err := chargingService.StopCharging(slotID, req.EndLevel)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "停止充电", record)
}

func UpdateSlotTelemetry(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	var req SlotTelemetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.ChargingSlotTelemetryRequest{
		Voltage:         req.Voltage,
		Current:         req.Current,
		Level:           req.Level,
		Temperature:     req.Temperature,
		ChargedCapacity: req.ChargedCapacity,
		ChargingTime:    req.ChargingTime,
		RemainingTime:   req.RemainingTime,
	}

	if err := chargingService.UpdateSlotTelemetry(slotID, serviceReq); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", nil)
}

func StationHeartbeat(c *gin.Context) {
	stationID, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	var req StationHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.StationHeartbeatRequest{
		Status:          req.Status,
		FirmwareVersion: req.FirmwareVersion,
		OccupiedSlots:   req.OccupiedSlots,
		ChargingSlots:   req.ChargingSlots,
	}

	if err := chargingService.ProcessStationHeartbeat(stationID, serviceReq); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "心跳已处理", nil)
}

func GetChargingRecords(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	batteryID, _ := utils.ParseUint64(c.Query("battery_id"))
	stationID, _ := utils.ParseUint64(c.Query("station_id"))
	status := c.Query("status")

	records, total, err := chargingService.GetChargingRecords(pagination, batteryID, stationID, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", records, total)
}

func GetChargingRecord(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电记录ID", nil)
		return
	}

	record, err := chargingService.GetChargingRecordByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "充电记录不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", record)
}

func GetBatteryChargingRecords(c *gin.Context) {
	batteryID, err := utils.ParseUint64(c.Param("battery_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的电池ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)

	records, total, err := chargingService.GetBatteryChargingRecords(batteryID, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", records, total)
}

func GetStationChargingRecords(c *gin.Context) {
	stationID, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电柜ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)

	records, total, err := chargingService.GetStationChargingRecords(stationID, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", records, total)
}

func GetChargingStatistics(c *gin.Context) {
	stats, err := chargingService.GetStatistics()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func AssignBatteryToSlot(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	var req AssignBatteryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := chargingService.AssignBatteryToSlot(slotID, req.BatteryID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "分配成功", nil)
}

func RemoveBatteryFromSlot(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	if err := chargingService.RemoveBatteryFromSlot(slotID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "移除成功", nil)
}

func SetSlotFault(c *gin.Context) {
	slotID, err := utils.ParseUint64(c.Param("slot_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的充电位ID", nil)
		return
	}

	var req struct {
		FaultCode    int    `json:"fault_code"`
		FaultMessage string `json:"fault_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := chargingService.SetSlotFault(slotID, req.FaultCode, req.FaultMessage); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "设置成功", nil)
}
