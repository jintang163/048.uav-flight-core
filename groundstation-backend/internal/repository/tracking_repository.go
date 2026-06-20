package repository

import (
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
)

type TrackingRepository struct{}

func NewTrackingRepository() *TrackingRepository {
	return &TrackingRepository{}
}

func (r *TrackingRepository) CreateDetection(target *models.DetectionTarget) error {
	return config.DB.Create(target).Error
}

func (r *TrackingRepository) BatchCreateDetections(targets []*models.DetectionTarget) error {
	if len(targets) == 0 {
		return nil
	}
	return config.DB.Create(&targets).Error
}

func (r *TrackingRepository) ListDetectionsByUAV(uavID uint64, page, pageSize int) ([]models.DetectionTarget, int64, error) {
	var targets []models.DetectionTarget
	var total int64

	query := config.DB.Model(&models.DetectionTarget{}).Where("uav_id = ?", uavID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&targets).Error
	return targets, total, err
}

func (r *TrackingRepository) GetLatestDetections(uavID uint64, limit int) ([]models.DetectionTarget, error) {
	var targets []models.DetectionTarget
	err := config.DB.Where("uav_id = ?", uavID).Order("created_at desc").Limit(limit).Find(&targets).Error
	return targets, err
}

func (r *TrackingRepository) CreateTracking(task *models.TrackingTask) error {
	return config.DB.Create(task).Error
}

func (r *TrackingRepository) GetTrackingByID(id uint64) (*models.TrackingTask, error) {
	var task models.TrackingTask
	err := config.DB.Preload("UAV").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TrackingRepository) GetActiveTrackingByUAV(uavID uint64) (*models.TrackingTask, error) {
	var task models.TrackingTask
	err := config.DB.Where("uav_id = ? AND status IN ?", uavID, []models.TrackingStatus{
		models.TrackingStatusLocking,
		models.TrackingStatusTracking,
		models.TrackingStatusSearching,
	}).Order("created_at desc").First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *TrackingRepository) ListTrackings(uavID *uint64, status *models.TrackingStatus, page, pageSize int) ([]models.TrackingTask, int64, error) {
	var tasks []models.TrackingTask
	var total int64

	query := config.DB.Model(&models.TrackingTask{}).Preload("UAV")
	if uavID != nil {
		query = query.Where("uav_id = ?", *uavID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (r *TrackingRepository) UpdateTracking(task *models.TrackingTask) error {
	return config.DB.Save(task).Error
}

func (r *TrackingRepository) UpdateTrackingStatus(id uint64, status models.TrackingStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.TrackingStatusCompleted || status == models.TrackingStatusLost {
		now := config.DB.NowFunc()
		updates["end_time"] = now
	}
	return config.DB.Model(&models.TrackingTask{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TrackingRepository) DeleteTracking(id uint64) error {
	return config.DB.Delete(&models.TrackingTask{}, id).Error
}
