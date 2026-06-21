package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type LinkStatusRepository struct {
	*BaseRepository
}

func NewLinkStatusRepository() *LinkStatusRepository {
	return &LinkStatusRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *LinkStatusRepository) Create(status *models.LinkStatus) error {
	return r.db.Create(status).Error
}

func (r *LinkStatusRepository) GetLatestByUAVID(uavID uint64) (*models.LinkStatus, error) {
	var status models.LinkStatus
	err := r.db.Where("uav_id = ?", uavID).Order("timestamp DESC").First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *LinkStatusRepository) ListByUAVID(uavID uint64, page, pageSize int, startTime, endTime *time.Time) ([]*models.LinkStatus, int64, error) {
	var statuses []*models.LinkStatus
	pagination := &utils.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	query := r.db.Model(&models.LinkStatus{}).Where("uav_id = ?", uavID)
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

func (r *LinkStatusRepository) GetCurrentActiveLinkCount() (map[models.LinkType]int64, error) {
	type result struct {
		ActiveLink models.LinkType
		Count      int64
	}

	var results []result
	err := r.db.Model(&models.LinkStatus{}).
		Select("active_link, COUNT(*) as count").
		Where("id IN (SELECT MAX(id) FROM link_status GROUP BY uav_id)").
		Group("active_link").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[models.LinkType]int64)
	for _, r := range results {
		counts[r.ActiveLink] = r.Count
	}

	return counts, nil
}

func (r *LinkStatusRepository) DeleteOldRecords(beforeTime time.Time) (int64, error) {
	result := r.db.Where("timestamp < ?", beforeTime).Delete(&models.LinkStatus{})
	return result.RowsAffected, result.Error
}
