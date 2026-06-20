package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type GeofenceViolationRepository struct {
	*BaseRepository
}

func NewGeofenceViolationRepository() *GeofenceViolationRepository {
	return &GeofenceViolationRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *GeofenceViolationRepository) Create(log *models.GeofenceViolationLog) error {
	return r.db.Create(log).Error
}

func (r *GeofenceViolationRepository) FindByID(id uint64) (*models.GeofenceViolationLog, error) {
	var log models.GeofenceViolationLog
	if err := r.db.Preload("UAV").Preload("Geofence").First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *GeofenceViolationRepository) List(pagination *utils.Pagination, uavID uint64, geofenceID uint64, severity string, violationType string, isResolved *bool, startTime string, endTime string) ([]models.GeofenceViolationLog, int64, error) {
	var logs []models.GeofenceViolationLog
	query := r.db.Model(&models.GeofenceViolationLog{}).Preload("UAV").Preload("Geofence")

	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if geofenceID > 0 {
		query = query.Where("geofence_id = ?", geofenceID)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if violationType != "" {
		query = query.Where("violation_type = ?", violationType)
	}
	if isResolved != nil {
		query = query.Where("is_resolved = ?", *isResolved)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *GeofenceViolationRepository) MarkResolved(id uint64, notes string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"is_resolved": true,
		"resolved_at": &now,
	}
	if notes != "" {
		updates["notes"] = notes
	}
	return r.db.Model(&models.GeofenceViolationLog{}).Where("id = ?", id).Updates(updates).Error
}

func (r *GeofenceViolationRepository) GetStatistics(uavID uint64, geofenceID uint64, startTime string, endTime string) (map[string]interface{}, error) {
	query := r.db.Model(&models.GeofenceViolationLog{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if geofenceID > 0 {
		query = query.Where("geofence_id = ?", geofenceID)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var unresolved int64
	if err := query.Where("is_resolved = ?", false).Count(&unresolved).Error; err != nil {
		return nil, err
	}

	var critical int64
	if err := query.Where("severity = ?", models.ViolationSeverityCritical).Count(&critical).Error; err != nil {
		return nil, err
	}

	var byType []struct {
		ViolationType string `gorm:"column:violation_type"`
		Count         int64
	}
	if err := query.Select("violation_type, count(*) as count").Group("violation_type").Scan(&byType).Error; err != nil {
		return nil, err
	}

	typeMap := make(map[string]int64)
	for _, item := range byType {
		typeMap[item.ViolationType] = item.Count
	}

	return map[string]interface{}{
		"total":      total,
		"unresolved": unresolved,
		"critical":   critical,
		"by_type":    typeMap,
	}, nil
}

func (r *GeofenceViolationRepository) CleanOldData(days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	result := r.db.Where("created_at < ?", cutoff).Delete(&models.GeofenceViolationLog{})
	return result.RowsAffected, result.Error
}
