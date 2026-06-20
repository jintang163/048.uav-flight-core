package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type TemporaryUnlockingRepository struct {
	*BaseRepository
}

func NewTemporaryUnlockingRepository() *TemporaryUnlockingRepository {
	return &TemporaryUnlockingRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *TemporaryUnlockingRepository) Create(unlocking *models.TemporaryUnlocking) error {
	return r.db.Create(unlocking).Error
}

func (r *TemporaryUnlockingRepository) FindByID(id uint64) (*models.TemporaryUnlocking, error) {
	var unlocking models.TemporaryUnlocking
	if err := r.db.Preload("UAV").Preload("Geofence").Preload("Applicant").Preload("Approver").First(&unlocking, id).Error; err != nil {
		return nil, err
	}
	return &unlocking, nil
}

func (r *TemporaryUnlockingRepository) FindByUUID(uuid string) (*models.TemporaryUnlocking, error) {
	var unlocking models.TemporaryUnlocking
	if err := r.db.Preload("UAV").Preload("Geofence").Preload("Applicant").Preload("Approver").Where("uuid = ?", uuid).First(&unlocking).Error; err != nil {
		return nil, err
	}
	return &unlocking, nil
}

func (r *TemporaryUnlockingRepository) List(pagination *utils.Pagination, uavID uint64, applicantID uint64, status string, category string, startTime string, endTime string) ([]models.TemporaryUnlocking, int64, error) {
	var unlockings []models.TemporaryUnlocking
	query := r.db.Model(&models.TemporaryUnlocking{}).Preload("UAV").Preload("Applicant").Preload("Approver")

	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if applicantID > 0 {
		query = query.Where("applicant_id = ?", applicantID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if startTime != "" {
		query = query.Where("start_time >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("end_time <= ?", endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&unlockings).Error; err != nil {
		return nil, 0, err
	}
	return unlockings, total, nil
}

func (r *TemporaryUnlockingRepository) Approve(id uint64, approverID uint64, remark string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":          models.UnlockStatusApproved,
		"approver_id":     approverID,
		"approval_remark": remark,
		"approved_at":     &now,
	}
	return r.db.Model(&models.TemporaryUnlocking{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TemporaryUnlockingRepository) Reject(id uint64, approverID uint64, remark string) error {
	updates := map[string]interface{}{
		"status":          models.UnlockStatusRejected,
		"approver_id":     approverID,
		"approval_remark": remark,
	}
	return r.db.Model(&models.TemporaryUnlocking{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TemporaryUnlockingRepository) Cancel(id uint64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.UnlockStatusCancelled,
		"cancelled_at": &now,
	}
	return r.db.Model(&models.TemporaryUnlocking{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TemporaryUnlockingRepository) GetActiveUnlockings(uavID uint64, category string) ([]models.TemporaryUnlocking, error) {
	var unlockings []models.TemporaryUnlocking
	query := r.db.Where("uav_id = ? AND status = ?", uavID, models.UnlockStatusApproved)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	now := time.Now()
	query = query.Where("start_time <= ? AND end_time >= ?", &now, &now)
	err := query.Order("created_at DESC").Find(&unlockings).Error
	return unlockings, err
}

func (r *TemporaryUnlockingRepository) CheckActiveUnlock(uavID uint64, geofenceID uint64) (*models.TemporaryUnlocking, error) {
	var unlocking models.TemporaryUnlocking
	now := time.Now()
	err := r.db.Where("uav_id = ? AND geofence_id = ? AND status = ? AND start_time <= ? AND end_time >= ?",
		uavID, geofenceID, models.UnlockStatusApproved, &now, &now).
		Order("created_at DESC").First(&unlocking).Error
	if err != nil {
		return nil, err
	}
	return &unlocking, nil
}

func (r *TemporaryUnlockingRepository) ExpireOld() (int64, error) {
	now := time.Now()
	result := r.db.Model(&models.TemporaryUnlocking{}).
		Where("status = ? AND end_time < ?", models.UnlockStatusApproved, &now).
		Update("status", models.UnlockStatusExpired)
	return result.RowsAffected, result.Error
}
