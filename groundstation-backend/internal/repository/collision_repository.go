package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"
)

type CollisionRepository struct {
	*BaseRepository
}

func NewCollisionRepository() *CollisionRepository {
	return &CollisionRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *CollisionRepository) CreateAlert(alert *models.CollisionAlert) error {
	return r.db.Create(alert).Error
}

func (r *CollisionRepository) UpdateAlert(alert *models.CollisionAlert) error {
	return r.db.Save(alert).Error
}

func (r *CollisionRepository) ResolveAlert(id uint64) error {
	now := time.Now()
	return r.db.Model(&models.CollisionAlert{}).Where("id = ?", id).
		Updates(map[string]interface{}{"is_resolved": true, "resolved_at": now}).Error
}

func (r *CollisionRepository) GetActiveAlerts() ([]models.CollisionAlert, error) {
	var alerts []models.CollisionAlert
	err := r.db.Where("is_resolved = ?", false).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *CollisionRepository) GetAlertByID(id uint64) (*models.CollisionAlert, error) {
	var alert models.CollisionAlert
	err := r.db.Where("id = ?", id).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *CollisionRepository) GetAlertByAlertID(alertID string) (*models.CollisionAlert, error) {
	var alert models.CollisionAlert
	err := r.db.Where("alert_id = ?", alertID).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *CollisionRepository) ListAlerts(pagination *utils.Pagination, riskLevel string, uavID uint64, resolved bool) ([]models.CollisionAlert, int64, error) {
	query := r.db.Model(&models.CollisionAlert{})
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	if uavID > 0 {
		query = query.Where("uav_id_1 = ? OR uav_id_2 = ?", uavID, uavID)
	}
	if resolved {
		query = query.Where("is_resolved = ?", true)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var alerts []models.CollisionAlert
	err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, total, err
}

func (r *CollisionRepository) CreateIntersection(intersection *models.RouteIntersection) error {
	return r.db.Create(intersection).Error
}

func (r *CollisionRepository) GetActiveIntersections(uavID uint64) ([]models.RouteIntersection, error) {
	var intersections []models.RouteIntersection
	query := r.db.Where("is_active = ?", true)
	if uavID > 0 {
		query = query.Where("uav_id_1 = ? OR uav_id_2 = ?", uavID, uavID)
	}
	err := query.Order("created_at DESC").Find(&intersections).Error
	return intersections, err
}

func (r *CollisionRepository) ClearActiveIntersections() error {
	return r.db.Model(&models.RouteIntersection{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error
}

func (r *CollisionRepository) DeactivateIntersectionsForUAV(uavID uint64) error {
	return r.db.Model(&models.RouteIntersection{}).
		Where("uav_id_1 = ? OR uav_id_2 = ?", uavID, uavID).
		Update("is_active", false).Error
}

func (r *CollisionRepository) GetIntersectionHistory(uavID uint64, start, end time.Time) ([]models.RouteIntersection, error) {
	var intersections []models.RouteIntersection
	query := r.db.Where("created_at BETWEEN ? AND ?", start, end)
	if uavID > 0 {
		query = query.Where("uav_id_1 = ? OR uav_id_2 = ?", uavID, uavID)
	}
	err := query.Order("created_at DESC").Find(&intersections).Error
	return intersections, err
}

func (r *CollisionRepository) GetAlertStats(start, end time.Time) (map[string]interface{}, error) {
	var total int64
	var critical, warning, avoiding int64

	r.db.Model(&models.CollisionAlert{}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Count(&total)

	r.db.Model(&models.CollisionAlert{}).
		Where("created_at BETWEEN ? AND ? AND risk_level = ?", start, end, models.CollisionRiskCritical).
		Count(&critical)

	r.db.Model(&models.CollisionAlert{}).
		Where("created_at BETWEEN ? AND ? AND risk_level = ?", start, end, models.CollisionRiskWarning).
		Count(&warning)

	r.db.Model(&models.CollisionAlert{}).
		Where("created_at BETWEEN ? AND ? AND action_taken IS NOT NULL", start, end).
		Count(&avoiding)

	return map[string]interface{}{
		"total":     total,
		"critical":  critical,
		"warning":   warning,
		"avoided":   avoiding,
		"resolved":  total - avoiding,
	}, nil
}
