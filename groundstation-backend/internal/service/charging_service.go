package service

import (
	"errors"
	"fmt"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"time"
)

type ChargingService struct {
	stationRepo *repository.ChargingStationRepository
	slotRepo    *repository.ChargingSlotRepository
	recordRepo  *repository.ChargingRecordRepository
	batteryRepo *repository.BatteryRepository
}

func NewChargingService() *ChargingService {
	return &ChargingService{
		stationRepo: repository.NewChargingStationRepository(),
		slotRepo:    repository.NewChargingSlotRepository(),
		recordRepo:  repository.NewChargingRecordRepository(),
		batteryRepo: repository.NewBatteryRepository(),
	}
}

type CreateChargingStationRequest struct {
	StationID  string `json:"station_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Model      string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	SlotCount  int    `json:"slot_count"`
	Location   string `json:"location"`
	IPAddress  string `json:"ip_address"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	MaxVoltage float64 `json:"max_voltage"`
	MaxCurrent float64 `json:"max_current"`
	Description string `json:"description"`
}

type UpdateChargingStationRequest struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	SlotCount  int    `json:"slot_count"`
	Location   string `json:"location"`
	IPAddress  string `json:"ip_address"`
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	MaxVoltage float64 `json:"max_voltage"`
	MaxCurrent float64 `json:"max_current"`
	Description string `json:"description"`
	Status     string `json:"status"`
}

type StartChargingRequest struct {
	BatteryID    uint64  `json:"battery_id" binding:"required"`
	ChargingMode string  `json:"charging_mode"`
	TargetVoltage float64 `json:"target_voltage"`
	TargetCurrent float64 `json:"target_current"`
}

type ChargingSlotTelemetryRequest struct {
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Level       float64 `json:"level"`
	Temperature float64 `json:"temperature"`
	ChargedCapacity float64 `json:"charged_capacity"`
	ChargingTime int   `json:"charging_time"`
	RemainingTime int  `json:"remaining_time"`
}

type StationHeartbeatRequest struct {
	Status         string `json:"status"`
	FirmwareVersion string `json:"firmware_version"`
	OccupiedSlots  int    `json:"occupied_slots"`
	ChargingSlots  int    `json:"charging_slots"`
}

func (s *ChargingService) CreateStation(req *CreateChargingStationRequest) (*models.ChargingStation, error) {
	existing, _ := s.stationRepo.FindByStationID(req.StationID)
	if existing != nil {
		return nil, errors.New("station ID already exists")
	}

	station := &models.ChargingStation{
		StationID:       req.StationID,
		Name:            req.Name,
		Model:           req.Model,
		Manufacturer:    req.Manufacturer,
		SlotCount:       req.SlotCount,
		Location:        req.Location,
		IPAddress:       req.IPAddress,
		Port:            req.Port,
		Protocol:        req.Protocol,
		Status:          models.ChargingStationStatusOffline,
		MaxVoltage:      req.MaxVoltage,
		MaxCurrent:      req.MaxCurrent,
		Description:     req.Description,
	}

	if station.Protocol == "" {
		station.Protocol = "tcp"
	}

	if err := s.stationRepo.Create(station); err != nil {
		return nil, err
	}

	if req.SlotCount > 0 {
		slots := make([]models.ChargingSlot, 0, req.SlotCount)
		for i := 0; i < req.SlotCount; i++ {
			slots = append(slots, models.ChargingSlot{
				StationID: station.ID,
				SlotIndex: i + 1,
				SlotName:  fmt.Sprintf("Slot %d", i+1),
				Status:    "empty",
			})
		}
		for _, slot := range slots {
			_ = s.slotRepo.Create(&slot)
		}
	}

	return station, nil
}

func (s *ChargingService) GetStationByID(id uint64) (*models.ChargingStation, error) {
	return s.stationRepo.FindByID(id)
}

func (s *ChargingService) GetStationByStationID(stationID string) (*models.ChargingStation, error) {
	return s.stationRepo.FindByStationID(stationID)
}

func (s *ChargingService) ListStations(pagination *utils.Pagination, status string, keyword string) ([]models.ChargingStation, int64, error) {
	return s.stationRepo.List(pagination, status, keyword)
}

func (s *ChargingService) UpdateStation(id uint64, req *UpdateChargingStationRequest) (*models.ChargingStation, error) {
	station, err := s.stationRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("charging station not found")
	}

	updates := &models.ChargingStation{}
	if req.Name != "" {
		updates.Name = req.Name
	}
	if req.Model != "" {
		updates.Model = req.Model
	}
	if req.Manufacturer != "" {
		updates.Manufacturer = req.Manufacturer
	}
	if req.Location != "" {
		updates.Location = req.Location
	}
	if req.IPAddress != "" {
		updates.IPAddress = req.IPAddress
	}
	if req.Port > 0 {
		updates.Port = req.Port
	}
	if req.Protocol != "" {
		updates.Protocol = req.Protocol
	}
	if req.MaxVoltage > 0 {
		updates.MaxVoltage = req.MaxVoltage
	}
	if req.MaxCurrent > 0 {
		updates.MaxCurrent = req.MaxCurrent
	}
	if req.Description != "" {
		updates.Description = req.Description
	}
	if req.Status != "" {
		updates.Status = models.ChargingStationStatus(req.Status)
	}

	if err := s.stationRepo.Update(id, updates); err != nil {
		return nil, err
	}

	return s.stationRepo.FindByID(id)
}

func (s *ChargingService) DeleteStation(id uint64) error {
	_, err := s.stationRepo.FindByID(id)
	if err != nil {
		return errors.New("charging station not found")
	}
	return s.stationRepo.SoftDelete(&models.ChargingStation{}, id)
}

func (s *ChargingService) GetStationSlots(stationID uint64) ([]models.ChargingSlot, error) {
	return s.slotRepo.FindByStation(stationID)
}

func (s *ChargingService) GetSlotByID(id uint64) (*models.ChargingSlot, error) {
	return s.slotRepo.FindByID(id)
}

func (s *ChargingService) StartCharging(slotID uint64, req *StartChargingRequest) (*models.ChargingRecord, error) {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return nil, errors.New("charging slot not found")
	}

	if slot.Status == "charging" {
		return nil, errors.New("slot is already charging")
	}

	battery, err := s.batteryRepo.FindByID(req.BatteryID)
	if err != nil {
		return nil, errors.New("battery not found")
	}

	mode := req.ChargingMode
	if mode == "" {
		mode = "standard"
	}

	if err := s.slotRepo.AssignBattery(slotID, req.BatteryID); err != nil {
		return nil, err
	}

	if err := s.slotRepo.StartCharging(slotID, mode, req.TargetVoltage, req.TargetCurrent); err != nil {
		return nil, err
	}

	now := time.Now()
	record := &models.ChargingRecord{
		BatteryID:       req.BatteryID,
		StationID:       slot.StationID,
		SlotID:          slotID,
		StartLevel:      battery.CurrentLevel,
		StartVoltage:    battery.CurrentVoltage,
		ChargingMode:    mode,
		ChargingCurrent: req.TargetCurrent,
		Status:          "charging",
		StartTime:       &now,
	}

	if err := s.recordRepo.Create(record); err != nil {
		return nil, err
	}

	if err := s.batteryRepo.UpdateStatus(req.BatteryID, models.BatteryStatusCharging); err != nil {
		return nil, err
	}

	s.updateStationSlotCounts(slot.StationID)

	return record, nil
}

func (s *ChargingService) StopCharging(slotID uint64, endLevel float64) (*models.ChargingRecord, error) {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return nil, errors.New("charging slot not found")
	}

	if slot.Status != "charging" {
		return nil, errors.New("slot is not charging")
	}

	if err := s.slotRepo.StopCharging(slotID, endLevel); err != nil {
		return nil, err
	}

	record, err := s.recordRepo.GetLatestByBattery(slot.BatteryID)
	if err != nil {
		return nil, err
	}

	duration := int(time.Since(*slot.StartTime).Seconds())
	chargedCapacity := float64(duration) * slot.CurrentCurrent / 3600.0
	energy := chargedCapacity * slot.CurrentVoltage / 1000.0

	if err := s.recordRepo.Complete(record.ID, endLevel, slot.CurrentVoltage,
		chargedCapacity, energy, duration); err != nil {
		return nil, err
	}

	if slot.BatteryID > 0 {
		_ = s.batteryRepo.IncrementChargeCount(slot.BatteryID)
		_ = s.batteryRepo.UpdateStatus(slot.BatteryID, models.BatteryStatusIdle)
	}

	s.updateStationSlotCounts(slot.StationID)

	return s.recordRepo.FindByID(record.ID)
}

func (s *ChargingService) UpdateSlotTelemetry(slotID uint64, req *ChargingSlotTelemetryRequest) error {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return errors.New("charging slot not found")
	}

	if err := s.slotRepo.UpdateTelemetry(slotID, req.Voltage, req.Current, req.Level,
		req.Temperature, req.ChargedCapacity, req.ChargingTime, req.RemainingTime); err != nil {
		return err
	}

	if slot.BatteryID > 0 {
		_ = s.batteryRepo.UpdateTelemetry(slot.BatteryID, req.Voltage, req.Level, req.Temperature, req.Current)
	}

	websocket.BroadcastChargingStatus(slot.StationID, slotID, req.Level, req.Voltage, req.Current)

	return nil
}

func (s *ChargingService) ProcessStationHeartbeat(stationID uint64, req *StationHeartbeatRequest) error {
	station, err := s.stationRepo.FindByID(stationID)
	if err != nil {
		return errors.New("charging station not found")
	}

	status := models.ChargingStationStatus(req.Status)
	if status == "" {
		status = models.ChargingStationStatusOnline
	}

	if err := s.stationRepo.UpdateStatus(stationID, status); err != nil {
		return err
	}

	if req.FirmwareVersion != "" {
		_ = s.stationRepo.Update(stationID, &models.ChargingStation{
			FirmwareVersion: req.FirmwareVersion,
		})
	}

	if req.OccupiedSlots > 0 || req.ChargingSlots > 0 {
		_ = s.stationRepo.UpdateSlotCounts(stationID, req.OccupiedSlots, req.ChargingSlots)
	}

	_ = station

	return nil
}

func (s *ChargingService) GetChargingRecords(pagination *utils.Pagination, batteryID uint64, stationID uint64, status string) ([]models.ChargingRecord, int64, error) {
	return s.recordRepo.List(pagination, batteryID, stationID, status)
}

func (s *ChargingService) GetChargingRecordByID(id uint64) (*models.ChargingRecord, error) {
	return s.recordRepo.FindByID(id)
}

func (s *ChargingService) GetBatteryChargingRecords(batteryID uint64, pagination *utils.Pagination) ([]models.ChargingRecord, int64, error) {
	return s.recordRepo.ListByBattery(batteryID, pagination)
}

func (s *ChargingService) GetStationChargingRecords(stationID uint64, pagination *utils.Pagination) ([]models.ChargingRecord, int64, error) {
	return s.recordRepo.ListByStation(stationID, pagination)
}

func (s *ChargingService) GetStatistics() (map[string]interface{}, error) {
	stationStats, _ := s.stationRepo.GetStatistics()
	recordStats, _ := s.recordRepo.GetStatistics()

	return map[string]interface{}{
		"stations": stationStats,
		"records":  recordStats,
	}, nil
}

func (s *ChargingService) updateStationSlotCounts(stationID uint64) {
	occupied, _ := s.slotRepo.CountByStationAndStatus(stationID, "occupied")
	charging, _ := s.slotRepo.CountByStationAndStatus(stationID, "charging")
	_ = s.stationRepo.UpdateSlotCounts(stationID, int(occupied), int(charging))
}

func (s *ChargingService) SetSlotFault(slotID uint64, faultCode int, faultMessage string) error {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return errors.New("charging slot not found")
	}

	if err := s.slotRepo.SetFault(slotID, faultCode, faultMessage); err != nil {
		return err
	}

	record, err := s.recordRepo.GetLatestByBattery(slot.BatteryID)
	if err == nil && record.Status == "charging" {
		_ = s.recordRepo.UpdateStatus(record.ID, "failed")
	}

	s.updateStationSlotCounts(slot.StationID)

	return nil
}

func (s *ChargingService) AssignBatteryToSlot(slotID uint64, batteryID uint64) error {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return errors.New("charging slot not found")
	}

	if slot.Status != "empty" {
		return errors.New("slot is not empty")
	}

	_, err = s.batteryRepo.FindByID(batteryID)
	if err != nil {
		return errors.New("battery not found")
	}

	if err := s.slotRepo.AssignBattery(slotID, batteryID); err != nil {
		return err
	}

	if err := s.batteryRepo.UpdateStatus(batteryID, models.BatteryStatusIdle); err != nil {
		return err
	}

	s.updateStationSlotCounts(slot.StationID)

	return nil
}

func (s *ChargingService) RemoveBatteryFromSlot(slotID uint64) error {
	slot, err := s.slotRepo.FindByID(slotID)
	if err != nil {
		return errors.New("charging slot not found")
	}

	if slot.Status == "charging" {
		return errors.New("cannot remove battery while charging")
	}

	batteryID := slot.BatteryID

	if err := s.slotRepo.RemoveBattery(slotID); err != nil {
		return err
	}

	if batteryID != nil && *batteryID > 0 {
		_ = s.batteryRepo.UpdateStatus(*batteryID, models.BatteryStatusStorage)
	}

	s.updateStationSlotCounts(slot.StationID)

	return nil
}
