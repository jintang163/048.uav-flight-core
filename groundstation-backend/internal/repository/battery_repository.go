package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type BatteryRepository struct {
	*BaseRepository
}

func NewBatteryRepository() *BatteryRepository {
	return &BatteryRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *BatteryRepository) Create(battery *models.Battery) error {
	battery.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(battery)
}

func (r *BatteryRepository) FindByID(id uint64) (*models.Battery, error) {
	var battery models.Battery
	if err := r.db.Preload("UAV").First(&battery, id).Error; err != nil {
		return nil, err
	}
	return &battery, nil
}

func (r *BatteryRepository) FindByUUID(uuid string) (*models.Battery, error) {
	var battery models.Battery
	if err := r.db.Preload("UAV").Where("uuid = ?", uuid).First(&battery).Error; err != nil {
		return nil, err
	}
	return &battery, nil
}

func (r *BatteryRepository) FindByBatteryID(batteryID string) (*models.Battery, error) {
	var battery models.Battery
	if err := r.db.Preload("UAV").Where("battery_id = ?", batteryID).First(&battery).Error; err != nil {
		return nil, err
	}
	return &battery, nil
}

func (r *BatteryRepository) List(pagination *utils.Pagination, status string, healthStatus string, uavID uint64, keyword string) ([]models.Battery, int64, error) {
	var batteries []models.Battery
	query := r.db.Model(&models.Battery{}).Preload("UAV")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if healthStatus != "" {
		query = query.Where("health_status = ?", healthStatus)
	}
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if keyword != "" {
		query = query.Where("battery_id LIKE ? OR model LIKE ? OR manufacturer LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order(pagination.OrderBy).Find(&batteries).Error; err != nil {
		return nil, 0, err
	}
	return batteries, total, nil
}

func (r *BatteryRepository) Update(id uint64, battery *models.Battery) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).Updates(battery).Error
}

func (r *BatteryRepository) UpdateStatus(id uint64, status models.BatteryStatus) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BatteryRepository) UpdateTelemetry(id uint64, voltage, level, temperature, current float64) error {
	now := time.Now()
	return r.db.Model(&models.Battery{}).Where("id = ?", id).Updates(map[string]interface{}{
		"current_voltage":     voltage,
		"current_level":       level,
		"current_temperature": temperature,
		"current_current":     current,
		"last_used_at":        &now,
	}).Error
}

func (r *BatteryRepository) UpdateSOH(id uint64, soh float64) error {
	var healthStatus models.BatteryHealthStatus
	switch {
	case soh >= 90:
		healthStatus = models.BatteryHealthExcellent
	case soh >= 80:
		healthStatus = models.BatteryHealthGood
	case soh >= 70:
		healthStatus = models.BatteryHealthFair
	case soh >= 60:
		healthStatus = models.BatteryHealthPoor
	default:
		healthStatus = models.BatteryHealthCritical
	}

	return r.db.Model(&models.Battery{}).Where("id = ?", id).Updates(map[string]interface{}{
		"soh":           soh,
		"health_status": healthStatus,
	}).Error
}

func (r *BatteryRepository) IncrementCycleCount(id uint64) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).
		UpdateColumn("cycle_count", gorm.Expr("cycle_count + 1")).Error
}

func (r *BatteryRepository) IncrementChargeCount(id uint64) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).
		UpdateColumn("total_charge_count", gorm.Expr("total_charge_count + 1")).Error
}

func (r *BatteryRepository) SetMaintenance(id uint64, needsMaintenance bool, message string) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).Updates(map[string]interface{}{
		"needs_maintenance":   needsMaintenance,
		"maintenance_message": message,
	}).Error
}

func (r *BatteryRepository) UpdateStorageDays(id uint64, days int) error {
	return r.db.Model(&models.Battery{}).Where("id = ?", id).Update("storage_days", days).Error
}

func (r *BatteryRepository) GetBatteriesForMaintenanceCheck(maxStorageDays int) ([]models.Battery, error) {
	var batteries []models.Battery
	err := r.db.Where("status = ? AND storage_days >= ?", models.BatteryStatusStorage, maxStorageDays).
		Find(&batteries).Error
	return batteries, err
}

func (r *BatteryRepository) CountByStatus(status models.BatteryStatus) (int64, error) {
	return r.Count(&models.Battery{}, "status = ?", status)
}

func (r *BatteryRepository) CountByHealthStatus(healthStatus models.BatteryHealthStatus) (int64, error) {
	return r.Count(&models.Battery{}, "health_status = ?", healthStatus)
}

func (r *BatteryRepository) GetStatistics() (map[string]interface{}, error) {
	total, _ := r.Count(&models.Battery{}, nil)
	inUse, _ := r.CountByStatus(models.BatteryStatusInUse)
	charging, _ := r.CountByStatus(models.BatteryStatusCharging)
	idle, _ := r.CountByStatus(models.BatteryStatusIdle)
	storage, _ := r.CountByStatus(models.BatteryStatusStorage)
	fault, _ := r.CountByStatus(models.BatteryStatusFault)

	excellent, _ := r.CountByHealthStatus(models.BatteryHealthExcellent)
	good, _ := r.CountByHealthStatus(models.BatteryHealthGood)
	fair, _ := r.CountByHealthStatus(models.BatteryHealthFair)
	poor, _ := r.CountByHealthStatus(models.BatteryHealthPoor)
	critical, _ := r.CountByHealthStatus(models.BatteryHealthCritical)

	needsMaintenance, _ := r.Count(&models.Battery{}, "needs_maintenance = ?", true)

	return map[string]interface{}{
		"total": total,
		"status": map[string]interface{}{
			"in_use":    inUse,
			"charging":  charging,
			"idle":      idle,
			"storage":   storage,
			"fault":     fault,
		},
		"health": map[string]interface{}{
			"excellent": excellent,
			"good":      good,
			"fair":      fair,
			"poor":      poor,
			"critical":  critical,
		},
		"needs_maintenance": needsMaintenance,
	}, nil
}

type BatteryUsageRecordRepository struct {
	*BaseRepository
}

func NewBatteryUsageRecordRepository() *BatteryUsageRecordRepository {
	return &BatteryUsageRecordRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *BatteryUsageRecordRepository) Create(record *models.BatteryUsageRecord) error {
	record.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(record)
}

func (r *BatteryUsageRecordRepository) FindByID(id uint64) (*models.BatteryUsageRecord, error) {
	var record models.BatteryUsageRecord
	if err := r.db.Preload("Battery").Preload("UAV").First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *BatteryUsageRecordRepository) ListByBattery(batteryID uint64, pagination *utils.Pagination) ([]models.BatteryUsageRecord, int64, error) {
	var records []models.BatteryUsageRecord
	query := r.db.Model(&models.BatteryUsageRecord{}).
		Preload("Battery").Preload("UAV").
		Where("battery_id = ?", batteryID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func (r *BatteryUsageRecordRepository) ListByUAV(uavID uint64, pagination *utils.Pagination) ([]models.BatteryUsageRecord, int64, error) {
	var records []models.BatteryUsageRecord
	query := r.db.Model(&models.BatteryUsageRecord{}).
		Preload("Battery").Preload("UAV").
		Where("uav_id = ?", uavID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

type BatteryCellDataRepository struct {
	*BaseRepository
}

func NewBatteryCellDataRepository() *BatteryCellDataRepository {
	return &BatteryCellDataRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *BatteryCellDataRepository) Create(data *models.BatteryCellData) error {
	return r.BaseRepository.Create(data)
}

func (r *BatteryCellDataRepository) FindByBattery(batteryID uint64) ([]models.BatteryCellData, error) {
	var data []models.BatteryCellData
	err := r.db.Where("battery_id = ?", batteryID).
		Order("cell_index ASC").Find(&data).Error
	return data, err
}

func (r *BatteryCellDataRepository) BatchCreate(data []models.BatteryCellData) error {
	if len(data) == 0 {
		return nil
	}
	return r.db.Create(&data).Error
}

type BatteryMaintenanceAlertRepository struct {
	*BaseRepository
}

func NewBatteryMaintenanceAlertRepository() *BatteryMaintenanceAlertRepository {
	return &BatteryMaintenanceAlertRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *BatteryMaintenanceAlertRepository) Create(alert *models.BatteryMaintenanceAlert) error {
	alert.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(alert)
}

func (r *BatteryMaintenanceAlertRepository) FindByID(id uint64) (*models.BatteryMaintenanceAlert, error) {
	var alert models.BatteryMaintenanceAlert
	if err := r.db.Preload("Battery").First(&alert, id).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *BatteryMaintenanceAlertRepository) List(pagination *utils.Pagination, batteryID uint64, status string, alertType string, level string) ([]models.BatteryMaintenanceAlert, int64, error) {
	var alerts []models.BatteryMaintenanceAlert
	query := r.db.Model(&models.BatteryMaintenanceAlert{}).Preload("Battery")

	if batteryID > 0 {
		query = query.Where("battery_id = ?", batteryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order("created_at DESC").Find(&alerts).Error; err != nil {
		return nil, 0, err
	}
	return alerts, total, nil
}

func (r *BatteryMaintenanceAlertRepository) Acknowledge(id uint64, userID uint64) error {
	now := time.Now()
	return r.db.Model(&models.BatteryMaintenanceAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          models.AlertStatusAcknowledged,
		"acknowledged_by": userID,
		"acknowledged_at": &now,
	}).Error
}

func (r *BatteryMaintenanceAlertRepository) Resolve(id uint64, userID uint64, note string) error {
	now := time.Now()
	return r.db.Model(&models.BatteryMaintenanceAlert{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        models.AlertStatusResolved,
		"resolved_by":   userID,
		"resolved_at":   &now,
		"resolved_note": note,
	}).Error
}

func (r *BatteryMaintenanceAlertRepository) GetUnacknowledgedCount() (int64, error) {
	return r.Count(&models.BatteryMaintenanceAlert{}, "status = ?", models.AlertStatusNew)
}
