package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type FlightRepository struct {
	*BaseRepository
}

func NewFlightRepository() *FlightRepository {
	return &FlightRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *FlightRepository) CreateFlightStatus(status *models.FlightStatus) error {
	return r.db.Create(status).Error
}

func (r *FlightRepository) GetLatestStatus(uavID uint64) (*models.FlightStatus, error) {
	var status models.FlightStatus
	err := r.db.Where("uav_id = ?", uavID).Order("timestamp DESC").First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *FlightRepository) GetStatusHistory(uavID uint64, pagination *utils.Pagination, startTime, endTime *time.Time) ([]models.FlightStatus, int64, error) {
	var statuses []models.FlightStatus
	query := r.db.Model(&models.FlightStatus{}).Where("uav_id = ?", uavID)
	if startTime != nil {
		query = query.Where("timestamp >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("timestamp <= ?", *endTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("timestamp DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&statuses).Error
	return statuses, total, err
}

func (r *FlightRepository) GetStatusByTimeRange(uavID uint64, startTime, endTime time.Time) ([]models.FlightStatus, error) {
	var statuses []models.FlightStatus
	err := r.db.Where("uav_id = ? AND timestamp >= ? AND timestamp <= ?",
		uavID, startTime, endTime).Order("timestamp ASC").Find(&statuses).Error
	return statuses, err
}

func (r *FlightRepository) BatchCreate(statuses []models.FlightStatus) error {
	return r.db.CreateInBatches(statuses, 100).Error
}

func (r *FlightRepository) CleanOldData(beforeDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -beforeDays)
	result := r.db.Where("timestamp < ?", cutoff).Delete(&models.FlightStatus{})
	return result.RowsAffected, result.Error
}

func (r *FlightRepository) GetFlightSummary(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	var result map[string]interface{}
	row := r.db.Model(&models.FlightStatus{}).
		Where("uav_id = ? AND timestamp >= ? AND timestamp <= ?", uavID, startTime, endTime).
		Select(`
			COUNT(*) as total_points,
			MAX(altitude_rel) as max_altitude,
			AVG(ground_speed) as avg_speed,
			MAX(ground_speed) as max_speed,
			MIN(battery_level) as min_battery,
			AVG(signal_strength) as avg_signal
		`).Row()
	err := row.Scan(
		&result["total_points"],
		&result["max_altitude"],
		&result["avg_speed"],
		&result["max_speed"],
		&result["min_battery"],
		&result["avg_signal"],
	)
	return result, err
}
