package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type WeatherRepository struct {
	*BaseRepository
}

func NewWeatherRepository() *WeatherRepository {
	return &WeatherRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *WeatherRepository) CreateWeatherData(data *models.WeatherData) error {
	return r.db.Create(data).Error
}

func (r *WeatherRepository) GetLatestByUAVID(uavID uint64) (*models.WeatherData, error) {
	var data models.WeatherData
	err := r.db.Where("uav_id = ?", uavID).Order("created_at DESC").First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *WeatherRepository) GetWeatherHistory(uavID uint64, start, end time.Time) ([]models.WeatherData, error) {
	var data []models.WeatherData
	err := r.db.Where("uav_id = ? AND created_at BETWEEN ? AND ?", uavID, start, end).
		Order("created_at ASC").Find(&data).Error
	return data, err
}

func (r *WeatherRepository) CreateWeatherAlert(alert *models.WeatherAlertEvent) error {
	return r.db.Create(alert).Error
}

func (r *WeatherRepository) ResolveWeatherAlert(id uint64) error {
	now := time.Now()
	return r.db.Model(&models.WeatherAlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_resolved": true, "resolved_at": now}).Error
}

func (r *WeatherRepository) GetActiveAlerts(uavID uint64) ([]models.WeatherAlertEvent, error) {
	var alerts []models.WeatherAlertEvent
	err := r.db.Where("uav_id = ? AND is_resolved = false", uavID).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *WeatherRepository) ListAlerts(pagination *utils.Pagination, uavID uint64, alertType string, level string) ([]models.WeatherAlertEvent, int64, error) {
	query := r.db.Model(&models.WeatherAlertEvent{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if alertType != "" {
		query = query.Where("alert_type = ?", alertType)
	}
	if level != "" {
		query = query.Where("alert_level = ?", level)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var alerts []models.WeatherAlertEvent
	err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, total, err
}

func (r *WeatherRepository) CreateFlightWeatherLog(log *models.FlightWeatherLog) error {
	return r.db.Create(log).Error
}

func (r *WeatherRepository) UpdateFlightWeatherLog(log *models.FlightWeatherLog) error {
	return r.db.Save(log).Error
}

func (r *WeatherRepository) GetFlightWeatherLog(flightID uint64) (*models.FlightWeatherLog, error) {
	var log models.FlightWeatherLog
	err := r.db.Where("flight_id = ?", flightID).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *WeatherRepository) GetWeatherStats(uavID uint64, start, end time.Time) (map[string]interface{}, error) {
	baseQuery := r.db.Model(&models.WeatherData{}).Where("created_at BETWEEN ? AND ?", start, end)
	if uavID > 0 {
		baseQuery = baseQuery.Where("uav_id = ?", uavID)
	}

	var count int64
	baseQuery.Count(&count)

	var result struct {
		AvgWind float64
		MaxWind float64
		AvgTemp float64
	}
	r.db.Model(&models.WeatherData{}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Select("COALESCE(AVG(wind_speed), 0) as avg_wind, COALESCE(MAX(wind_speed), 0) as max_wind, COALESCE(AVG(temperature), 0) as avg_temp").
		Scan(&result)
	if uavID > 0 {
		r.db.Model(&models.WeatherData{}).
			Where("created_at BETWEEN ? AND ? AND uav_id = ?", start, end, uavID).
			Select("COALESCE(AVG(wind_speed), 0) as avg_wind, COALESCE(MAX(wind_speed), 0) as max_wind, COALESCE(AVG(temperature), 0) as avg_temp").
			Scan(&result)
	}

	stats := map[string]interface{}{
		"sample_count":    count,
		"avg_wind_speed":  result.AvgWind,
		"max_wind_speed":  result.MaxWind,
		"avg_temperature": result.AvgTemp,
	}
	return stats, nil
}

func (r *WeatherRepository) DeleteOldWeatherData(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&models.WeatherData{}).Error
}
