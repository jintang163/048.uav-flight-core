package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type ChargingStationRepository struct {
	*BaseRepository
}

func NewChargingStationRepository() *ChargingStationRepository {
	return &ChargingStationRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *ChargingStationRepository) Create(station *models.ChargingStation) error {
	station.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(station)
}

func (r *ChargingStationRepository) FindByID(id uint64) (*models.ChargingStation, error) {
	var station models.ChargingStation
	if err := r.db.First(&station, id).Error; err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *ChargingStationRepository) FindByUUID(uuid string) (*models.ChargingStation, error) {
	var station models.ChargingStation
	if err := r.db.Where("uuid = ?", uuid).First(&station).Error; err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *ChargingStationRepository) FindByStationID(stationID string) (*models.ChargingStation, error) {
	var station models.ChargingStation
	if err := r.db.Where("station_id = ?", stationID).First(&station).Error; err != nil {
		return nil, err
	}
	return &station, nil
}

func (r *ChargingStationRepository) List(pagination *utils.Pagination, status string, keyword string) ([]models.ChargingStation, int64, error) {
	var stations []models.ChargingStation
	query := r.db.Model(&models.ChargingStation{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR station_id LIKE ? OR location LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order(pagination.OrderBy).Find(&stations).Error; err != nil {
		return nil, 0, err
	}
	return stations, total, nil
}

func (r *ChargingStationRepository) Update(id uint64, station *models.ChargingStation) error {
	return r.db.Model(&models.ChargingStation{}).Where("id = ?", id).Updates(station).Error
}

func (r *ChargingStationRepository) UpdateStatus(id uint64, status models.ChargingStationStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.ChargingStationStatusOnline {
		updates["last_online_at"] = time.Now()
	}
	return r.db.Model(&models.ChargingStation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ChargingStationRepository) UpdateSlotCounts(id uint64, occupied, charging int) error {
	return r.db.Model(&models.ChargingStation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"occupied_slots": occupied,
		"charging_slots": charging,
	}).Error
}

func (r *ChargingStationRepository) CountByStatus(status models.ChargingStationStatus) (int64, error) {
	return r.Count(&models.ChargingStation{}, "status = ?", status)
}

func (r *ChargingStationRepository) GetStatistics() (map[string]interface{}, error) {
	total, _ := r.Count(&models.ChargingStation{}, nil)
	online, _ := r.CountByStatus(models.ChargingStationStatusOnline)
	offline, _ := r.CountByStatus(models.ChargingStationStatusOffline)
	fault, _ := r.CountByStatus(models.ChargingStationStatusFault)
	maintenance, _ := r.CountByStatus(models.ChargingStationStatusMaintenance)

	return map[string]interface{}{
		"total":    total,
		"online":   online,
		"offline":  offline,
		"fault":    fault,
		"maintenance": maintenance,
	}, nil
}

type ChargingSlotRepository struct {
	*BaseRepository
}

func NewChargingSlotRepository() *ChargingSlotRepository {
	return &ChargingSlotRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *ChargingSlotRepository) Create(slot *models.ChargingSlot) error {
	return r.BaseRepository.Create(slot)
}

func (r *ChargingSlotRepository) FindByID(id uint64) (*models.ChargingSlot, error) {
	var slot models.ChargingSlot
	if err := r.db.Preload("Station").Preload("Battery").First(&slot, id).Error; err != nil {
		return nil, err
	}
	return &slot, nil
}

func (r *ChargingSlotRepository) FindByStation(stationID uint64) ([]models.ChargingSlot, error) {
	var slots []models.ChargingSlot
	err := r.db.Preload("Battery").Where("station_id = ?", stationID).
		Order("slot_index ASC").Find(&slots).Error
	return slots, err
}

func (r *ChargingSlotRepository) Update(id uint64, slot *models.ChargingSlot) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", id).Updates(slot).Error
}

func (r *ChargingSlotRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ChargingSlotRepository) AssignBattery(slotID uint64, batteryID uint64) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"battery_id": batteryID,
		"status":     "occupied",
	}).Error
}

func (r *ChargingSlotRepository) RemoveBattery(slotID uint64) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"battery_id": nil,
		"status":     "empty",
	}).Error
}

func (r *ChargingSlotRepository) StartCharging(slotID uint64, mode string, targetVoltage, targetCurrent float64) error {
	now := time.Now()
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"status":         "charging",
		"charging_mode":  mode,
		"target_voltage": targetVoltage,
		"target_current": targetCurrent,
		"start_time":     &now,
	}).Error
}

func (r *ChargingSlotRepository) StopCharging(slotID uint64, endLevel float64) error {
	now := time.Now()
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"status":       "occupied",
		"current_level": endLevel,
		"end_time":     &now,
	}).Error
}

func (r *ChargingSlotRepository) UpdateTelemetry(slotID uint64, voltage, current, level, temperature float64, chargedCapacity float64, chargingTime int, remainingTime int) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"current_voltage":  voltage,
		"current_current":  current,
		"current_level":    level,
		"temperature":      temperature,
		"charged_capacity": chargedCapacity,
		"charging_time":    chargingTime,
		"remaining_time":   remainingTime,
	}).Error
}

func (r *ChargingSlotRepository) SetFault(slotID uint64, faultCode int, faultMessage string) error {
	return r.db.Model(&models.ChargingSlot{}).Where("id = ?", slotID).Updates(map[string]interface{}{
		"status":        "fault",
		"fault_code":    faultCode,
		"fault_message": faultMessage,
	}).Error
}

func (r *ChargingSlotRepository) CountByStationAndStatus(stationID uint64, status string) (int64, error) {
	return r.Count(&models.ChargingSlot{}, "station_id = ? AND status = ?", stationID, status)
}

type ChargingRecordRepository struct {
	*BaseRepository
}

func NewChargingRecordRepository() *ChargingRecordRepository {
	return &ChargingRecordRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *ChargingRecordRepository) Create(record *models.ChargingRecord) error {
	record.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(record)
}

func (r *ChargingRecordRepository) FindByID(id uint64) (*models.ChargingRecord, error) {
	var record models.ChargingRecord
	if err := r.db.Preload("Battery").Preload("Station").First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ChargingRecordRepository) List(pagination *utils.Pagination, batteryID uint64, stationID uint64, status string) ([]models.ChargingRecord, int64, error) {
	var records []models.ChargingRecord
	query := r.db.Model(&models.ChargingRecord{}).Preload("Battery").Preload("Station")

	if batteryID > 0 {
		query = query.Where("battery_id = ?", batteryID)
	}
	if stationID > 0 {
		query = query.Where("station_id = ?", stationID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

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

func (r *ChargingRecordRepository) ListByBattery(batteryID uint64, pagination *utils.Pagination) ([]models.ChargingRecord, int64, error) {
	return r.List(pagination, batteryID, 0, "")
}

func (r *ChargingRecordRepository) ListByStation(stationID uint64, pagination *utils.Pagination) ([]models.ChargingRecord, int64, error) {
	return r.List(pagination, 0, stationID, "")
}

func (r *ChargingRecordRepository) GetLatestByBattery(batteryID uint64) (*models.ChargingRecord, error) {
	var record models.ChargingRecord
	err := r.db.Where("battery_id = ?", batteryID).
		Order("created_at DESC").Limit(1).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ChargingRecordRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&models.ChargingRecord{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ChargingRecordRepository) Complete(id uint64, endLevel, endVoltage, chargedCapacity, energy float64, duration int) error {
	now := time.Now()
	return r.db.Model(&models.ChargingRecord{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":           "completed",
		"end_level":        endLevel,
		"end_voltage":      endVoltage,
		"charged_capacity": chargedCapacity,
		"energy_consumed":  energy,
		"charging_time":    duration,
		"end_time":         &now,
	}).Error
}

func (r *ChargingRecordRepository) CountByStatus(status string) (int64, error) {
	return r.Count(&models.ChargingRecord{}, "status = ?", status)
}

func (r *ChargingRecordRepository) GetStatistics() (map[string]interface{}, error) {
	total, _ := r.Count(&models.ChargingRecord{}, nil)
	charging, _ := r.CountByStatus("charging")
	completed, _ := r.CountByStatus("completed")
	failed, _ := r.CountByStatus("failed")

	return map[string]interface{}{
		"total":     total,
		"charging":  charging,
		"completed": completed,
		"failed":    failed,
	}, nil
}
