package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type MissionRepository struct {
	*BaseRepository
}

func NewMissionRepository() *MissionRepository {
	return &MissionRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *MissionRepository) CreateTemplate(template *models.MissionTemplate) error {
	template.UUID = utils.GenerateUUID()
	return r.db.Create(template).Error
}

func (r *MissionRepository) FindTemplateByID(id uint64) (*models.MissionTemplate, error) {
	var template models.MissionTemplate
	if err := r.db.Preload("Waypoints").First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *MissionRepository) FindTemplateByUUID(uuid string) (*models.MissionTemplate, error) {
	var template models.MissionTemplate
	if err := r.db.Preload("Waypoints").Where("uuid = ?", uuid).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *MissionRepository) ListTemplates(pagination *utils.Pagination, templateType string, creatorID uint64, isPublic *bool) ([]models.MissionTemplate, int64, error) {
	var templates []models.MissionTemplate
	query := r.db.Model(&models.MissionTemplate{}).Preload("Waypoints")
	if templateType != "" {
		query = query.Where("type = ?", templateType)
	}
	if creatorID > 0 {
		query = query.Where("creator_id = ?", creatorID)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func (r *MissionRepository) CreateMission(mission *models.FlightMission) error {
	mission.UUID = utils.GenerateUUID()
	return r.db.Create(mission).Error
}

func (r *MissionRepository) FindMissionByID(id uint64) (*models.FlightMission, error) {
	var mission models.FlightMission
	if err := r.db.Preload("UAV").Preload("Template").Preload("Waypoints").First(&mission, id).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *MissionRepository) FindMissionByUUID(uuid string) (*models.FlightMission, error) {
	var mission models.FlightMission
	if err := r.db.Preload("UAV").Preload("Template").Preload("Waypoints").Where("uuid = ?", uuid).First(&mission).Error; err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *MissionRepository) ListMissions(pagination *utils.Pagination, uavID uint64, status models.MissionStatus, operatorID uint64) ([]models.FlightMission, int64, error) {
	var missions []models.FlightMission
	query := r.db.Model(&models.FlightMission{}).Preload("UAV").Preload("Template")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if operatorID > 0 {
		query = query.Where("operator_id = ?", operatorID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&missions).Error; err != nil {
		return nil, 0, err
	}
	return missions, total, nil
}

func (r *MissionRepository) UpdateMissionStatus(id uint64, status models.MissionStatus) error {
	return r.db.Model(&models.FlightMission{}).Where("id = ?", id).Update("status", status).Error
}

func (r *MissionRepository) UpdateCurrentWaypoint(id uint64, wpIndex int) error {
	return r.db.Model(&models.FlightMission{}).Where("id = ?", id).Update("current_waypoint", wpIndex).Error
}

func (r *MissionRepository) AddWaypoint(wp *models.MissionWaypoint) error {
	return r.db.Create(wp).Error
}

func (r *MissionRepository) UpdateWaypointReached(wpID uint64, reached bool) error {
	updates := map[string]interface{}{
		"is_reached": reached,
	}
	if reached {
		updates["reached_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&models.MissionWaypoint{}).Where("id = ?", wpID).Updates(updates).Error
}

func (r *MissionRepository) GetMissionWaypoints(missionID uint64) ([]models.MissionWaypoint, error) {
	var waypoints []models.MissionWaypoint
	err := r.db.Where("mission_id = ?", missionID).Order("seq ASC").Find(&waypoints).Error
	return waypoints, err
}

func (r *MissionRepository) GetActiveMissionByUAV(uavID uint64) (*models.FlightMission, error) {
	var mission models.FlightMission
	err := r.db.Where("uav_id = ? AND status IN ?", uavID, []models.MissionStatus{
		models.MissionStatusExecuting,
		models.MissionStatusPaused,
		models.MissionStatusReady,
	}).Preload("Waypoints").Order("created_at DESC").First(&mission).Error
	if err != nil {
		return nil, err
	}
	return &mission, nil
}

func (r *MissionRepository) ListTemplatesByCategory(pagination *utils.Pagination, category string, keyword string) ([]models.MissionTemplate, int64, error) {
	var templates []models.MissionTemplate
	query := r.db.Model(&models.MissionTemplate{}).Preload("Waypoints")
	if category != "" {
		query = query.Where("category = ? OR type = ?", category, category)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

func (r *MissionRepository) UpdateTemplate(template *models.MissionTemplate) error {
	return r.db.Save(template).Error
}

func (r *MissionRepository) DeleteTemplateWaypoints(templateID uint64) error {
	return r.db.Where("template_id = ?", templateID).Delete(&models.MissionWaypoint{}).Error
}

func (r *MissionRepository) ListMissionsFiltered(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.FlightMission, int64, error) {
	var missions []models.FlightMission
	query := r.db.Model(&models.FlightMission{}).Preload("UAV").Preload("Template")
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
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&missions).Error; err != nil {
		return nil, 0, err
	}
	return missions, total, nil
}

func (r *MissionRepository) Update(mission *models.FlightMission) error {
	return r.db.Save(mission).Error
}
