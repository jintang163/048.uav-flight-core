package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
)

type PayloadMissionRepository struct {
	*BaseRepository
}

func NewPayloadMissionRepository() *PayloadMissionRepository {
	return &PayloadMissionRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *PayloadMissionRepository) CreateOrbitMission(mission *models.OrbitMission) error {
	return r.db.Create(mission).Error
}

func (r *PayloadMissionRepository) FindOrbitMissionByID(id uint64) (*models.OrbitMission, error) {
	var mission models.OrbitMission
	if err := r.db.Preload("UAV").Preload("Payload").First(&mission, id).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *PayloadMissionRepository) ListOrbitMissions(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.OrbitMission, int64, error) {
	var missions []models.OrbitMission
	var total int64

	query := r.db.Model(&models.OrbitMission{})
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	if err := query.Preload("UAV").Preload("Payload").Order("created_at DESC").Find(&missions).Error; err != nil {
		return nil, 0, err
	}

	return missions, total, nil
}

func (r *PayloadMissionRepository) GetActiveOrbitMissionByUAV(uavID uint64) (*models.OrbitMission, error) {
	var mission models.OrbitMission
	if err := r.db.Where("uav_id = ? AND status IN (?)", uavID,
		[]models.OrbitMissionStatus{models.OrbitStatusActive, models.OrbitStatusPaused}).
		First(&mission).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *PayloadMissionRepository) UpdateOrbitMission(mission *models.OrbitMission) error {
	return r.db.Save(mission).Error
}

func (r *PayloadMissionRepository) UpdateOrbitStatus(id uint64, status models.OrbitMissionStatus) error {
	return r.db.Model(&models.OrbitMission{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *PayloadMissionRepository) DeleteOrbitMission(id uint64) error {
	return r.db.Delete(&models.OrbitMission{}, id).Error
}

func (r *PayloadMissionRepository) CreateOrthoMission(mission *models.OrthoMission) error {
	return r.db.Create(mission).Error
}

func (r *PayloadMissionRepository) FindOrthoMissionByID(id uint64) (*models.OrthoMission, error) {
	var mission models.OrthoMission
	if err := r.db.Preload("UAV").Preload("Payload").Preload("Mission").
		Preload("Waypoints").First(&mission, id).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *PayloadMissionRepository) ListOrthoMissions(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.OrthoMission, int64, error) {
	var missions []models.OrthoMission
	var total int64

	query := r.db.Model(&models.OrthoMission{})
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	if err := query.Preload("UAV").Preload("Payload").Order("created_at DESC").Find(&missions).Error; err != nil {
		return nil, 0, err
	}

	return missions, total, nil
}

func (r *PayloadMissionRepository) GetActiveOrthoMissionByUAV(uavID uint64) (*models.OrthoMission, error) {
	var mission models.OrthoMission
	if err := r.db.Where("uav_id = ? AND status IN (?)", uavID,
		[]models.OrthoMissionStatus{models.OrthoStatusActive, models.OrthoStatusPaused, models.OrthoStatusPlanning}).
		First(&mission).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *PayloadMissionRepository) UpdateOrthoMission(mission *models.OrthoMission) error {
	return r.db.Save(mission).Error
}

func (r *PayloadMissionRepository) UpdateOrthoStatus(id uint64, status models.OrthoMissionStatus) error {
	return r.db.Model(&models.OrthoMission{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *PayloadMissionRepository) DeleteOrthoMission(id uint64) error {
	return r.db.Delete(&models.OrthoMission{}, id).Error
}

func (r *PayloadMissionRepository) AddOrthoWaypoint(wp *models.OrthoWaypoint) error {
	return r.db.Create(wp).Error
}

func (r *PayloadMissionRepository) AddOrthoWaypoints(wps []models.OrthoWaypoint) error {
	if len(wps) == 0 {
		return nil
	}
	return r.db.Create(&wps).Error
}

func (r *PayloadMissionRepository) GetOrthoWaypoints(missionID uint64) ([]models.OrthoWaypoint, error) {
	var waypoints []models.OrthoWaypoint
	if err := r.db.Where("ortho_mission_id = ?", missionID).Order("seq ASC").Find(&waypoints).Error; err != nil {
		return nil, err
	}
	return waypoints, nil
}

func (r *PayloadMissionRepository) DeleteOrthoWaypoints(missionID uint64) error {
	return r.db.Where("ortho_mission_id = ?", missionID).Delete(&models.OrthoWaypoint{}).Error
}

func (r *PayloadMissionRepository) UpdateOrthoWaypointReached(wpID uint64, reached bool) error {
	return r.db.Model(&models.OrthoWaypoint{}).Where("id = ?", wpID).
		Updates(map[string]interface{}{
			"is_reached": reached,
		}).Error
}

func (r *PayloadMissionRepository) CreateTTSTask(task *models.TextToSpeechTask) error {
	return r.db.Create(task).Error
}

func (r *PayloadMissionRepository) FindTTSTaskByID(id uint64) (*models.TextToSpeechTask, error) {
	var task models.TextToSpeechTask
	if err := r.db.Preload("Payload").Preload("UAV").First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *PayloadMissionRepository) ListTTSTasks(pagination *utils.Pagination, payloadID uint64, uavID uint64, status string) ([]models.TextToSpeechTask, int64, error) {
	var tasks []models.TextToSpeechTask
	var total int64

	query := r.db.Model(&models.TextToSpeechTask{})
	if payloadID > 0 {
		query = query.Where("payload_id = ?", payloadID)
	}
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	if err := query.Preload("Payload").Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *PayloadMissionRepository) UpdateTTSTask(task *models.TextToSpeechTask) error {
	return r.db.Save(task).Error
}

func (r *PayloadMissionRepository) UpdateTTSStatus(id uint64, status string, errorMsg string, audioURL string, duration int) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	if audioURL != "" {
		updates["audio_url"] = audioURL
	}
	if duration > 0 {
		updates["duration"] = duration
	}
	return r.db.Model(&models.TextToSpeechTask{}).Where("id = ?", id).Updates(updates).Error
}
