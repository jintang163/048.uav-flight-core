package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"
)

type BatteryService struct {
	batteryRepo       *repository.BatteryRepository
	usageRecordRepo   *repository.BatteryUsageRecordRepository
	cellDataRepo      *repository.BatteryCellDataRepository
	maintenanceAlertRepo *repository.BatteryMaintenanceAlertRepository
	alertService      *AlertService
}

func NewBatteryService() *BatteryService {
	return &BatteryService{
		batteryRepo:        repository.NewBatteryRepository(),
		usageRecordRepo:    repository.NewBatteryUsageRecordRepository(),
		cellDataRepo:       repository.NewBatteryCellDataRepository(),
		maintenanceAlertRepo: repository.NewBatteryMaintenanceAlertRepository(),
		alertService:       NewAlertService(),
	}
}

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
	Voltage     float64 `json:"voltage" binding:"required"`
	Level       float64 `json:"level" binding:"required"`
	Temperature float64 `json:"temperature"`
	Current     float64 `json:"current"`
	CellVoltages []float64 `json:"cell_voltages"`
}

type MaintenanceAlertAckRequest struct {
	Note string `json:"note"`
}

func (s *BatteryService) Create(req *CreateBatteryRequest) (*models.Battery, error) {
	existing, _ := s.batteryRepo.FindByBatteryID(req.BatteryID)
	if existing != nil {
		return nil, errors.New("battery ID already exists")
	}

	battery := &models.Battery{
		BatteryID:    req.BatteryID,
		Model:        req.Model,
		Manufacturer: req.Manufacturer,
		Capacity:     req.Capacity,
		CapacityUnit: req.CapacityUnit,
		Voltage:      req.Voltage,
		CellCount:    req.CellCount,
		Status:       models.BatteryStatusIdle,
		HealthStatus: models.BatteryHealthExcellent,
		SOH:          100,
		Location:     req.Location,
		Notes:        req.Notes,
	}

	if battery.CapacityUnit == "" {
		battery.CapacityUnit = "mAh"
	}

	if err := s.batteryRepo.Create(battery); err != nil {
		return nil, err
	}

	return battery, nil
}

func (s *BatteryService) GetByID(id uint64) (*models.Battery, error) {
	return s.batteryRepo.FindByID(id)
}

func (s *BatteryService) GetByBatteryID(batteryID string) (*models.Battery, error) {
	return s.batteryRepo.FindByBatteryID(batteryID)
}

func (s *BatteryService) List(pagination *utils.Pagination, status string, healthStatus string, uavID uint64, keyword string) ([]models.Battery, int64, error) {
	return s.batteryRepo.List(pagination, status, healthStatus, uavID, keyword)
}

func (s *BatteryService) Update(id uint64, req *UpdateBatteryRequest) (*models.Battery, error) {
	battery, err := s.batteryRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("battery not found")
	}

	updates := &models.Battery{}
	if req.Model != "" {
		updates.Model = req.Model
	}
	if req.Manufacturer != "" {
		updates.Manufacturer = req.Manufacturer
	}
	if req.Capacity > 0 {
		updates.Capacity = req.Capacity
	}
	if req.CapacityUnit != "" {
		updates.CapacityUnit = req.CapacityUnit
	}
	if req.Voltage > 0 {
		updates.Voltage = req.Voltage
	}
	if req.CellCount > 0 {
		updates.CellCount = req.CellCount
	}
	if req.Status != "" {
		updates.Status = models.BatteryStatus(req.Status)
	}
	if req.Location != "" {
		updates.Location = req.Location
	}
	if req.Notes != "" {
		updates.Notes = req.Notes
	}

	if err := s.batteryRepo.Update(id, updates); err != nil {
		return nil, err
	}

	return s.batteryRepo.FindByID(id)
}

func (s *BatteryService) Delete(id uint64) error {
	_, err := s.batteryRepo.FindByID(id)
	if err != nil {
		return errors.New("battery not found")
	}
	return s.batteryRepo.SoftDelete(&models.Battery{}, id)
}

func (s *BatteryService) UpdateTelemetry(id uint64, req *BatteryTelemetryRequest) error {
	battery, err := s.batteryRepo.FindByID(id)
	if err != nil {
		return errors.New("battery not found")
	}

	err = s.batteryRepo.UpdateTelemetry(id, req.Voltage, req.Level, req.Temperature, req.Current)
	if err != nil {
		return err
	}

	if req.CellVoltages != nil && len(req.CellVoltages) > 0 {
		cellData := make([]models.BatteryCellData, 0, len(req.CellVoltages))
		now := time.Now()
		for i, voltage := range req.CellVoltages {
			cellData = append(cellData, models.BatteryCellData{
				BatteryID:  id,
				CellIndex:  i + 1,
				Voltage:    voltage,
				Status:     "normal",
				RecordedAt: now,
			})
		}
		_ = s.cellDataRepo.BatchCreate(cellData)
	}

	s.checkLowBatteryAlert(battery, req.Level)

	return nil
}

func (s *BatteryService) checkLowBatteryAlert(battery *models.Battery, level float64) {
	if battery.UAVID == nil {
		return
	}

	if level <= 15 {
		_, _ = s.alertService.CreateCustomAlert(*battery.UAVID,
			"电池电量严重不足",
			"电池电量已降至15%以下，执行强制降落！",
			models.AlertLevelCritical)
	} else if level <= 30 {
		_, _ = s.alertService.CreateCustomAlert(*battery.UAVID,
			"电池电量低",
			"电池电量已降至30%以下，建议自动返航",
			models.AlertLevelWarning)
	}
}

func (s *BatteryService) UpdateSOH(id uint64, soh float64) error {
	return s.batteryRepo.UpdateSOH(id, soh)
}

func (s *BatteryService) CalculateSOH(battery *models.Battery) float64 {
	baseSOH := 100.0
	cycleFactor := 0.02
	soh := baseSOH - float64(battery.CycleCount)*cycleFactor

	if soh < 0 {
		soh = 0
	}
	if soh > 100 {
		soh = 100
	}

	return soh
}

func (s *BatteryService) GetStatistics() (map[string]interface{}, error) {
	return s.batteryRepo.GetStatistics()
}

func (s *BatteryService) GetUsageRecords(batteryID uint64, pagination *utils.Pagination) ([]models.BatteryUsageRecord, int64, error) {
	return s.usageRecordRepo.ListByBattery(batteryID, pagination)
}

func (s *BatteryService) GetCellData(batteryID uint64) ([]models.BatteryCellData, error) {
	return s.cellDataRepo.FindByBattery(batteryID)
}

func (s *BatteryService) CheckMaintenanceReminders(maxStorageDays int) ([]models.BatteryMaintenanceAlert, error) {
	batteries, err := s.batteryRepo.GetBatteriesForMaintenanceCheck(maxStorageDays)
	if err != nil {
		return nil, err
	}

	var alerts []models.BatteryMaintenanceAlert
	now := time.Now()

	for _, battery := range batteries {
		existingAlerts, _, _ := s.maintenanceAlertRepo.List(
			&utils.Pagination{Page: 1, PageSize: 10, OrderBy: "id DESC"},
			battery.ID,
			string(models.AlertStatusNew),
			"storage_discharge",
			"",
		)

		if len(existingAlerts) == 0 {
			alert := &models.BatteryMaintenanceAlert{
				BatteryID:   battery.ID,
				AlertType:   "storage_discharge",
				Level:       models.AlertLevelWarning,
				Title:       "电池存放时间过长",
				Message:     "电池已存放超过7天未使用，建议进行放电保养以维护电池健康",
				Status:      models.AlertStatusNew,
				StorageDays: battery.StorageDays,
				SOH:         battery.SOH,
			}

			if err := s.maintenanceAlertRepo.Create(alert); err == nil {
				alerts = append(alerts, *alert)
			}

			_ = s.batteryRepo.SetMaintenance(battery.ID, true,
				"电池已存放超过7天，建议放电保养")
		}
	}

	return alerts, nil
}

func (s *BatteryService) GetMaintenanceAlerts(pagination *utils.Pagination, batteryID uint64, status string, alertType string, level string) ([]models.BatteryMaintenanceAlert, int64, error) {
	return s.maintenanceAlertRepo.List(pagination, batteryID, status, alertType, level)
}

func (s *BatteryService) AcknowledgeMaintenanceAlert(id uint64, userID uint64) error {
	_, err := s.maintenanceAlertRepo.FindByID(id)
	if err != nil {
		return errors.New("maintenance alert not found")
	}
	return s.maintenanceAlertRepo.Acknowledge(id, userID)
}

func (s *BatteryService) ResolveMaintenanceAlert(id uint64, userID uint64, note string) error {
	alert, err := s.maintenanceAlertRepo.FindByID(id)
	if err != nil {
		return errors.New("maintenance alert not found")
	}

	err = s.maintenanceAlertRepo.Resolve(id, userID, note)
	if err != nil {
		return err
	}

	if alert.BatteryID > 0 {
		_ = s.batteryRepo.SetMaintenance(alert.BatteryID, false, "")
		_ = s.batteryRepo.UpdateStorageDays(alert.BatteryID, 0)
	}

	return nil
}

func (s *BatteryService) GetUnacknowledgedMaintenanceCount() (int64, error) {
	return s.maintenanceAlertRepo.GetUnacknowledgedCount()
}

func (s *BatteryService) IdentifyBattery(batteryID string) (*models.Battery, error) {
	battery, err := s.batteryRepo.FindByBatteryID(batteryID)
	if err == nil {
		return battery, nil
	}

	return nil, errors.New("battery not found")
}

func (s *BatteryService) RegisterBatteryUse(batteryID uint64, uavID uint64) error {
	battery, err := s.batteryRepo.FindByID(batteryID)
	if err != nil {
		return errors.New("battery not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"uav_id":      uavID,
		"status":      models.BatteryStatusInUse,
		"last_used_at": &now,
		"storage_days": 0,
	}

	if battery.FirstUseDate == nil {
		updates["first_use_date"] = &now
	}

	return s.batteryRepo.Update(batteryID, &models.Battery{})
}

func (s *BatteryService) StartMaintenanceScheduler(interval time.Duration, maxStorageDays int) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			_, _ = s.CheckMaintenanceReminders(maxStorageDays)
		}
	}()
}
