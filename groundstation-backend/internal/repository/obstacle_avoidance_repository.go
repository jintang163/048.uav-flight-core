package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type ObstacleAvoidanceRepository struct {
	*BaseRepository
}

func NewObstacleAvoidanceRepository() *ObstacleAvoidanceRepository {
	return &ObstacleAvoidanceRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *ObstacleAvoidanceRepository) GetConfig(uavID uint64) (*models.ObstacleAvoidanceConfig, error) {
	var config models.ObstacleAvoidanceConfig
	if err := r.db.Where("uav_id = ?", uavID).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *ObstacleAvoidanceRepository) CreateConfig(config *models.ObstacleAvoidanceConfig) error {
	return r.db.Create(config).Error
}

func (r *ObstacleAvoidanceRepository) UpdateConfig(uavID uint64, updates map[string]interface{}) error {
	return r.db.Model(&models.ObstacleAvoidanceConfig{}).Where("uav_id = ?", uavID).Updates(updates).Error
}

func (r *ObstacleAvoidanceRepository) CreateDetectionLog(log *models.ObstacleDetectionLog) error {
	return r.db.Create(log).Error
}

func (r *ObstacleAvoidanceRepository) ListDetectionLogs(uavID uint64, startTime, endTime time.Time, pagination *utils.Pagination) ([]models.ObstacleDetectionLog, int64, error) {
	var logs []models.ObstacleDetectionLog
	var total int64
	query := r.db.Model(&models.ObstacleDetectionLog{}).Where("uav_id = ?", uavID)
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *ObstacleAvoidanceRepository) CreateEvent(event *models.ObstacleAvoidanceEvent) error {
	return r.db.Create(event).Error
}

func (r *ObstacleAvoidanceRepository) UpdateEvent(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&models.ObstacleAvoidanceEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ObstacleAvoidanceRepository) GetEvent(id uint64) (*models.ObstacleAvoidanceEvent, error) {
	var event models.ObstacleAvoidanceEvent
	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *ObstacleAvoidanceRepository) ListEvents(uavID uint64, status string, pagination *utils.Pagination) ([]models.ObstacleAvoidanceEvent, int64, error) {
	var events []models.ObstacleAvoidanceEvent
	var total int64
	query := r.db.Model(&models.ObstacleAvoidanceEvent{}).Where("uav_id = ?", uavID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *ObstacleAvoidanceRepository) GetHeatmapPoints(uavID uint64, startTime, endTime time.Time) ([]models.ObstacleHeatmapPoint, error) {
	var points []models.ObstacleHeatmapPoint
	query := r.db.Model(&models.ObstacleHeatmapPoint{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if !startTime.IsZero() {
		query = query.Where("last_trigger_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("last_trigger_time <= ?", endTime)
	}
	if err := query.Order("trigger_count DESC").Limit(500).Find(&points).Error; err != nil {
		return nil, err
	}
	return points, nil
}

func (r *ObstacleAvoidanceRepository) UpsertHeatmapPoint(uavID uint64, lat, lng, alt, distance float64) error {
	var point models.ObstacleHeatmapPoint
	latRound := float64(int(lat*100000)) / 100000
	lngRound := float64(int(lng*100000)) / 100000

	err := r.db.Where("uav_id = ? AND latitude >= ? AND latitude <= ? AND longitude >= ? AND longitude <= ?",
		uavID, latRound-0.00001, latRound+0.00001, lngRound-0.00001, lngRound+0.00001).First(&point).Error

	if err != nil {
		newPoint := models.ObstacleHeatmapPoint{
			UAVID:           uavID,
			Latitude:        latRound,
			Longitude:       lngRound,
			Altitude:        alt,
			TriggerCount:    1,
			LastTriggerTime: time.Now(),
			AvgDistance:     distance,
			MinDistance:     distance,
		}
		return r.db.Create(&newPoint).Error
	}

	point.TriggerCount++
	point.LastTriggerTime = time.Now()
	point.AvgDistance = (point.AvgDistance*float64(point.TriggerCount-1) + distance) / float64(point.TriggerCount)
	if distance < point.MinDistance {
		point.MinDistance = distance
	}
	return r.db.Save(&point).Error
}

func (r *ObstacleAvoidanceRepository) ClearHeatmap(uavID uint64) error {
	query := r.db.Where("1 = 1")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	return query.Delete(&models.ObstacleHeatmapPoint{}).Error
}

func (r *ObstacleAvoidanceRepository) GetStatistics(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	detectionQuery := r.db.Model(&models.ObstacleDetectionLog{})
	eventQuery := r.db.Model(&models.ObstacleAvoidanceEvent{})
	if uavID > 0 {
		detectionQuery = detectionQuery.Where("uav_id = ?", uavID)
		eventQuery = eventQuery.Where("uav_id = ?", uavID)
	}
	if !startTime.IsZero() {
		detectionQuery = detectionQuery.Where("created_at >= ?", startTime)
		eventQuery = eventQuery.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		detectionQuery = detectionQuery.Where("created_at <= ?", endTime)
		eventQuery = eventQuery.Where("created_at <= ?", endTime)
	}

	var totalDetections int64
	detectionQuery.Count(&totalDetections)
	stats["total_detections"] = totalDetections

	var totalEvents int64
	eventQuery.Count(&totalEvents)
	stats["total_avoidance_events"] = totalEvents

	var successfulCount int64
	r.db.Model(&models.ObstacleAvoidanceEvent{}).Where("status = ?", models.AvoidanceStatusCompleted).Count(&successfulCount)
	stats["successful_avoidances"] = successfulCount

	var failedCount int64
	r.db.Model(&models.ObstacleAvoidanceEvent{}).Where("status = ?", models.AvoidanceStatusFailed).Count(&failedCount)
	stats["failed_avoidances"] = failedCount

	var nearestDistance float64
	r.db.Model(&models.ObstacleDetectionLog{}).Select("COALESCE(MIN(distance), 0)").Scan(&nearestDistance)
	stats["nearest_obstacle_distance"] = nearestDistance

	return stats, nil
}
