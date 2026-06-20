package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type AlertRepository struct {
	*BaseRepository
}

func NewAlertRepository() *AlertRepository {
	return &AlertRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *AlertRepository) Create(alert *models.AlertEvent) error {
	alert.UUID = utils.GenerateUUID()
	return r.db.Create(alert).Error
}

func (r *AlertRepository) FindByID(id uint64) (*models.AlertEvent, error) {
	var alert models.AlertEvent
	if err := r.db.Preload("UAV").First(&alert, id).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) FindByUUID(uuid string) (*models.AlertEvent, error) {
	var alert models.AlertEvent
	if err := r.db.Preload("UAV").Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *AlertRepository) List(pagination *utils.Pagination, uavID uint64, level models.AlertLevel, status models.AlertStatus, alertType models.AlertType) ([]models.AlertEvent, int64, error) {
	var alerts []models.AlertEvent
	query := r.db.Model(&models.AlertEvent{}).Preload("UAV")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if alertType != "" {
		query = query.Where("type = ?", alertType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}
	return alerts, total, nil
}

func (r *AlertRepository) UpdateStatus(id uint64, status models.AlertStatus, userID uint64, note string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	switch status {
	case models.AlertStatusAcknowledged:
		updates["acknowledged_by"] = userID
		updates["acknowledged_at"] = gorm.Expr("NOW()")
	case models.AlertStatusResolved:
		updates["resolved_by"] = userID
		updates["resolved_at"] = gorm.Expr("NOW()")
		updates["resolved_note"] = note
	}
	return r.db.Model(&models.AlertEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AlertRepository) MarkNotificationSent(id uint64) error {
	return r.db.Model(&models.AlertEvent{}).Where("id = ?", id).Update("notification_sent", true).Error
}

func (r *AlertRepository) MarkSMSSent(id uint64) error {
	return r.db.Model(&models.AlertEvent{}).Where("id = ?", id).Update("sms_sent", true).Error
}

func (r *AlertRepository) MarkEmailSent(id uint64) error {
	return r.db.Model(&models.AlertEvent{}).Where("id = ?", id).Update("email_sent", true).Error
}

func (r *AlertRepository) GetUnacknowledgedAlerts(uavID uint64) ([]models.AlertEvent, error) {
	var alerts []models.AlertEvent
	query := r.db.Where("status = ?", models.AlertStatusNew)
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	err := query.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *AlertRepository) GetAlertStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	var result struct {
		Total     int64
		Warning   int64
		Critical  int64
		Fatal     int64
		Resolved  int64
	}
	r.db.Model(&models.AlertEvent{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Select("COUNT(*) as total").Row().Scan(&result.Total)
	r.db.Model(&models.AlertEvent{}).
		Where("created_at >= ? AND created_at <= ? AND level = ?", startTime, endTime, models.AlertLevelWarning).
		Count(&result.Warning)
	r.db.Model(&models.AlertEvent{}).
		Where("created_at >= ? AND created_at <= ? AND level = ?", startTime, endTime, models.AlertLevelCritical).
		Count(&result.Critical)
	r.db.Model(&models.AlertEvent{}).
		Where("created_at >= ? AND created_at <= ? AND level = ?", startTime, endTime, models.AlertLevelFatal).
		Count(&result.Fatal)
	r.db.Model(&models.AlertEvent{}).
		Where("created_at >= ? AND created_at <= ? AND status = ?", startTime, endTime, models.AlertStatusResolved).
		Count(&result.Resolved)
	return map[string]interface{}{
		"total":    result.Total,
		"warning":  result.Warning,
		"critical": result.Critical,
		"fatal":    result.Fatal,
		"resolved": result.Resolved,
	}, nil
}

func (r *AlertRepository) CreateContact(contact *models.AlertContact) error {
	return r.db.Create(contact).Error
}

func (r *AlertRepository) GetActiveContacts(level models.AlertLevel) ([]models.AlertContact, error) {
	var contacts []models.AlertContact
	err := r.db.Where("is_active = ? AND alert_level IN ?", true,
		[]models.AlertLevel{level, models.AlertLevelInfo, models.AlertLevelWarning, models.AlertLevelCritical, models.AlertLevelFatal}).
		Find(&contacts).Error
	return contacts, err
}
