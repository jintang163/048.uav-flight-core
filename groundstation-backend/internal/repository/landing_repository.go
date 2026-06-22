package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type LandingRepository struct {
	*BaseRepository
}

func NewLandingRepository() *LandingRepository {
	return &LandingRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *LandingRepository) CreateLandingPoint(point *models.LandingPoint) error {
	return r.db.Create(point).Error
}

func (r *LandingRepository) FindLandingPointByID(id uint64) (*models.LandingPoint, error) {
	var point models.LandingPoint
	if err := r.db.First(&point, id).Error; err != nil {
		return nil, err
	}
	return &point, nil
}

func (r *LandingRepository) ListLandingPoints(pagination *utils.Pagination, pointType models.LandingPointType, status models.LandingPointStatus, hasMarkers *bool) ([]models.LandingPoint, int64, error) {
	var points []models.LandingPoint
	query := r.db.Model(&models.LandingPoint{})
	if pointType != "" {
		query = query.Where("type = ?", pointType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if hasMarkers != nil {
		query = query.Where("has_markers = ?", *hasMarkers)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("priority DESC, created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&points).Error; err != nil {
		return nil, 0, err
	}
	return points, total, nil
}

func (r *LandingRepository) ListLandingPointsByType(pointType models.LandingPointType) ([]models.LandingPoint, error) {
	var points []models.LandingPoint
	err := r.db.Where("type = ?", pointType).Order("priority DESC, created_at DESC").Find(&points).Error
	return points, err
}

func (r *LandingRepository) FindAvailableLandingPoints(minRadius float64) ([]models.LandingPoint, error) {
	var points []models.LandingPoint
	err := r.db.Where("status = ? AND radius >= ?", models.LandingPointStatusAvailable, minRadius).
		Order("priority DESC, created_at DESC").Find(&points).Error
	return points, err
}

func (r *LandingRepository) UpdateLandingPoint(point *models.LandingPoint) error {
	return r.db.Save(point).Error
}

func (r *LandingRepository) UpdateLandingPointStatus(id uint64, status models.LandingPointStatus) error {
	return r.db.Model(&models.LandingPoint{}).Where("id = ?", id).Update("status", status).Error
}

func (r *LandingRepository) DeleteLandingPoint(id uint64) error {
	return r.db.Delete(&models.LandingPoint{}, id).Error
}

func (r *LandingRepository) CreateLandingSession(session *models.LandingSession) error {
	return r.db.Create(session).Error
}

func (r *LandingRepository) FindLandingSessionByID(id uint64) (*models.LandingSession, error) {
	var session models.LandingSession
	if err := r.db.Preload("PrimaryLanding").Preload("AlternateLanding").First(&session, id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *LandingRepository) GetActiveLandingSession(uavID uint64) (*models.LandingSession, error) {
	var session models.LandingSession
	err := r.db.Preload("PrimaryLanding").Preload("AlternateLanding").
		Where("uav_id = ? AND status IN ?", uavID, []models.LandingSessionStatus{
			models.LandingSessionStatusPending,
			models.LandingSessionStatusApproaching,
			models.LandingSessionStatusDescending,
			models.LandingSessionStatusPrecision,
		}).Order("created_at DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *LandingRepository) ListLandingSessions(pagination *utils.Pagination, uavID uint64, status models.LandingSessionStatus, startTime, endTime string) ([]models.LandingSession, int64, error) {
	var sessions []models.LandingSession
	query := r.db.Model(&models.LandingSession{}).Preload("PrimaryLanding").Preload("AlternateLanding")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
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
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (r *LandingRepository) UpdateLandingSession(session *models.LandingSession) error {
	return r.db.Save(session).Error
}

func (r *LandingRepository) UpdateLandingSessionStatus(id uint64, status models.LandingSessionStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.LandingSessionStatusLanded || status == models.LandingSessionStatusAborted || status == models.LandingSessionStatusFailed {
		now := time.Now()
		updates["end_time"] = &now
	}
	return r.db.Model(&models.LandingSession{}).Where("id = ?", id).Updates(updates).Error
}

func (r *LandingRepository) AddTrajectoryPoint(point *models.LandingTrajectoryPoint) error {
	return r.db.Create(point).Error
}

func (r *LandingRepository) GetTrajectoryBySession(sessionID uint64) ([]models.LandingTrajectoryPoint, error) {
	var points []models.LandingTrajectoryPoint
	err := r.db.Where("session_id = ?", sessionID).Order("sequence ASC, timestamp ASC").Find(&points).Error
	return points, err
}

func (r *LandingRepository) GetTrajectoryBySessionWithLimit(sessionID uint64, limit int) ([]models.LandingTrajectoryPoint, error) {
	var points []models.LandingTrajectoryPoint
	err := r.db.Where("session_id = ?", sessionID).Order("sequence ASC, timestamp ASC").Limit(limit).Find(&points).Error
	return points, err
}

func (r *LandingRepository) GetLastTrajectoryPoint(sessionID uint64) (*models.LandingTrajectoryPoint, error) {
	var point models.LandingTrajectoryPoint
	err := r.db.Where("session_id = ?", sessionID).Order("sequence DESC").First(&point).Error
	if err != nil {
		return nil, err
	}
	return &point, nil
}

func (r *LandingRepository) CreateForcedLandingEvent(event *models.ForcedLandingEvent) error {
	return r.db.Create(event).Error
}

func (r *LandingRepository) FindForcedLandingEventByID(id uint64) (*models.ForcedLandingEvent, error) {
	var event models.ForcedLandingEvent
	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *LandingRepository) GetActiveForcedLandingEvent(uavID uint64) (*models.ForcedLandingEvent, error) {
	var event models.ForcedLandingEvent
	err := r.db.Where("uav_id = ? AND resolved_at IS NULL", uavID).Order("triggered_at DESC").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *LandingRepository) ListForcedLandingEvents(pagination *utils.Pagination, uavID uint64, triggerType string, isResolved *bool) ([]models.ForcedLandingEvent, int64, error) {
	var events []models.ForcedLandingEvent
	query := r.db.Model(&models.ForcedLandingEvent{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if triggerType != "" {
		query = query.Where("trigger_type = ?", triggerType)
	}
	if isResolved != nil {
		if *isResolved {
			query = query.Where("resolved_at IS NOT NULL")
		} else {
			query = query.Where("resolved_at IS NULL")
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("triggered_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, total, nil
}

func (r *LandingRepository) ResolveForcedLandingEvent(id uint64, resolvedBy uint64, notes string) error {
	now := time.Now()
	return r.db.Model(&models.ForcedLandingEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"resolved_at":       &now,
		"resolved_by":       resolvedBy,
		"resolution_notes":  notes,
	}).Error
}

func (r *LandingRepository) AddVisionLandingData(data *models.VisionLandingData) error {
	return r.db.Create(data).Error
}

func (r *LandingRepository) GetVisionLandingDataBySession(sessionID uint64, limit int) ([]models.VisionLandingData, error) {
	var data []models.VisionLandingData
	query := r.db.Where("session_id = ?", sessionID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&data).Error
	return data, err
}

func (r *LandingRepository) GetLatestVisionLandingData(uavID uint64) (*models.VisionLandingData, error) {
	var data models.VisionLandingData
	err := r.db.Where("uav_id = ?", uavID).Order("timestamp DESC").First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *LandingRepository) AddRTKPositionData(data *models.RTKPositionData) error {
	return r.db.Create(data).Error
}

func (r *LandingRepository) GetRTKPositionDataBySession(sessionID uint64, limit int) ([]models.RTKPositionData, error) {
	var data []models.RTKPositionData
	query := r.db.Where("session_id = ?", sessionID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&data).Error
	return data, err
}

func (r *LandingRepository) GetLatestRTKPositionData(uavID uint64) (*models.RTKPositionData, error) {
	var data models.RTKPositionData
	err := r.db.Where("uav_id = ?", uavID).Order("timestamp DESC").First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *LandingRepository) GetLandingStatistics(uavID uint64, startTime, endTime string) (map[string]interface{}, error) {
	query := r.db.Model(&models.LandingSession{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
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

	var landedCount int64
	if err := query.Where("status = ?", models.LandingSessionStatusLanded).Count(&landedCount).Error; err != nil {
		return nil, err
	}

	var abortedCount int64
	if err := query.Where("status = ?", models.LandingSessionStatusAborted).Count(&abortedCount).Error; err != nil {
		return nil, err
	}

	var failedCount int64
	if err := query.Where("status = ?", models.LandingSessionStatusFailed).Count(&failedCount).Error; err != nil {
		return nil, err
	}

	var avgError float64
	err := query.Where("status = ?", models.LandingSessionStatusLanded).
		Select("COALESCE(AVG(landing_error), 0)").Scan(&avgError).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_landings":     total,
		"successful_landings": landedCount,
		"aborted_landings":   abortedCount,
		"failed_landings":    failedCount,
		"success_rate":       float64(landedCount) / float64(total) * 100,
		"avg_landing_error":  avgError,
	}, nil
}
