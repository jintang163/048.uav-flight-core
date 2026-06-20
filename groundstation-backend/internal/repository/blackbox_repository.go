package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type BlackboxRepository struct {
	*BaseRepository
}

func NewBlackboxRepository() *BlackboxRepository {
	return &BlackboxRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *BlackboxRepository) Create(log *models.BlackboxLog) error {
	log.UUID = utils.GenerateUUID()
	return r.db.Create(log).Error
}

func (r *BlackboxRepository) FindByID(id uint64) (*models.BlackboxLog, error) {
	var log models.BlackboxLog
	if err := r.db.Preload("UAV").Preload("Mission").First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *BlackboxRepository) FindByUUID(uuid string) (*models.BlackboxLog, error) {
	var log models.BlackboxLog
	if err := r.db.Preload("UAV").Preload("Mission").Where("uuid = ?", uuid).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *BlackboxRepository) List(pagination *utils.Pagination, uavID uint64, missionID uint64, status models.BlackboxLogStatus, crashDetected *bool) ([]models.BlackboxLog, int64, error) {
	var logs []models.BlackboxLog
	query := r.db.Model(&models.BlackboxLog{}).Preload("UAV").Preload("Mission")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if missionID > 0 {
		query = query.Where("mission_id = ?", missionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if crashDetected != nil {
		query = query.Where("crash_detected = ?", *crashDetected)
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

func (r *BlackboxRepository) UpdateStatus(id uint64, status models.BlackboxLogStatus) error {
	return r.db.Model(&models.BlackboxLog{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BlackboxRepository) UpdateAnalysisResult(id uint64, maxAlt, maxSpeed, distance, batteryUsed float64, crashDetected bool) error {
	return r.db.Model(&models.BlackboxLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"max_altitude":   maxAlt,
		"max_speed":      maxSpeed,
		"distance":       distance,
		"battery_used":   batteryUsed,
		"crash_detected": crashDetected,
	}).Error
}

func (r *BlackboxRepository) GetLogsByTimeRange(uavID uint64, startTime, endTime time.Time) ([]models.BlackboxLog, error) {
	var logs []models.BlackboxLog
	query := r.db.Where("uav_id = ?", uavID)
	if !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}
	err := query.Order("start_time DESC").Find(&logs).Error
	return logs, err
}

func (r *BlackboxRepository) CreateAnalysisReport(report *models.LogAnalysisReport) error {
	return r.db.Create(report).Error
}

func (r *BlackboxRepository) GetReportsByLogID(logID uint64) ([]models.LogAnalysisReport, error) {
	var reports []models.LogAnalysisReport
	err := r.db.Where("log_id = ?", logID).Preload("Log").Order("created_at DESC").Find(&reports).Error
	return reports, err
}

func (r *BlackboxRepository) GetFlightStats(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	var totalFlights int64
	var totalDuration int64
	var totalDistance float64
	var crashCount int64

	r.db.Model(&models.BlackboxLog{}).
		Where("uav_id = ? AND start_time >= ? AND end_time <= ? AND status = ?",
			uavID, startTime, endTime, models.BlackboxStatusAnalyzed).
		Count(&totalFlights)
	r.db.Model(&models.BlackboxLog{}).
		Where("uav_id = ? AND start_time >= ? AND end_time <= ? AND status = ?",
			uavID, startTime, endTime, models.BlackboxStatusAnalyzed).
		Select("COALESCE(SUM(duration), 0)").Row().Scan(&totalDuration)
	r.db.Model(&models.BlackboxLog{}).
		Where("uav_id = ? AND start_time >= ? AND end_time <= ? AND status = ?",
			uavID, startTime, endTime, models.BlackboxStatusAnalyzed).
		Select("COALESCE(SUM(distance), 0)").Row().Scan(&totalDistance)
	r.db.Model(&models.BlackboxLog{}).
		Where("uav_id = ? AND start_time >= ? AND end_time <= ? AND crash_detected = ?",
			uavID, startTime, endTime, true).
		Count(&crashCount)

	return map[string]interface{}{
		"total_flights":    totalFlights,
		"total_duration":   totalDuration,
		"total_distance":   totalDistance,
		"crash_count":      crashCount,
		"avg_duration":     func() int64 { if totalFlights > 0 { return totalDuration / totalFlights }; return 0 }(),
		"avg_distance":     func() float64 { if totalFlights > 0 { return totalDistance / float64(totalFlights) }; return 0 }(),
	}, nil
}

func (r *BlackboxRepository) MarkUploadComplete(id uint64, fileSize int64, fileHash string) error {
	return r.db.Model(&models.BlackboxLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    models.BlackboxStatusUploaded,
		"file_size": fileSize,
		"file_hash": fileHash,
	}).Error
}
