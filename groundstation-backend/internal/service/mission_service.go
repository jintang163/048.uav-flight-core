package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"
)

type MissionService struct {
	missionRepo *repository.MissionRepository
	uavRepo     *repository.UAVRepository
}

func NewMissionService() *MissionService {
	return &MissionService{
		missionRepo: repository.NewMissionRepository(),
		uavRepo:     repository.NewUAVRepository(),
	}
}

type CreateTemplateRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description"`
	Type        models.MissionType     `json:"type" binding:"required"`
	MaxAltitude float64                 `json:"max_altitude"`
	MinAltitude float64                 `json:"min_altitude"`
	Speed       float64                 `json:"speed"`
	MaxDuration int                     `json:"max_duration"`
	IsPublic    bool                    `json:"is_public"`
	Waypoints   []models.MissionWaypoint `json:"waypoints"`
}

type CreateMissionRequest struct {
	Name           string              `json:"name" binding:"required"`
	UAVID          uint64              `json:"uav_id" binding:"required"`
	TemplateID     uint64              `json:"template_id"`
	OperatorID     uint64              `json:"operator_id"`
	PlannedStart   *time.Time          `json:"planned_start"`
	PlannedEnd     *time.Time          `json:"planned_end"`
	MaxAltitude    float64             `json:"max_altitude"`
	Speed          float64             `json:"speed"`
	EnableGeofence bool                `json:"enable_geofence"`
	FailSafe       string              `json:"fail_safe"`
	Notes          string              `json:"notes"`
	Waypoints      []models.MissionWaypoint `json:"waypoints"`
}

func (s *MissionService) CreateTemplate(req *CreateTemplateRequest, creatorID uint64) (*models.MissionTemplate, error) {
	template := &models.MissionTemplate{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		CreatorID:   creatorID,
		MaxAltitude: req.MaxAltitude,
		MinAltitude: req.MinAltitude,
		Speed:       req.Speed,
		MaxDuration: req.MaxDuration,
		IsPublic:    req.IsPublic,
	}

	if err := s.missionRepo.CreateTemplate(template); err != nil {
		return nil, err
	}

	for i := range req.Waypoints {
		req.Waypoints[i].TemplateID = template.ID
		req.Waypoints[i].Seq = i
		if err := s.missionRepo.AddWaypoint(&req.Waypoints[i]); err != nil {
			return nil, err
		}
	}

	return s.missionRepo.FindTemplateByID(template.ID)
}

func (s *MissionService) GetTemplate(id uint64) (*models.MissionTemplate, error) {
	return s.missionRepo.FindTemplateByID(id)
}

func (s *MissionService) ListTemplates(pagination *utils.Pagination, templateType string, creatorID uint64, isPublic *bool) ([]models.MissionTemplate, int64, error) {
	return s.missionRepo.ListTemplates(pagination, templateType, creatorID, isPublic)
}

func (s *MissionService) DeleteTemplate(id uint64) error {
	_, err := s.missionRepo.FindTemplateByID(id)
	if err != nil {
		return errors.New("template not found")
	}
	return s.missionRepo.SoftDelete(&models.MissionTemplate{}, id)
}

func (s *MissionService) CreateMission(req *CreateMissionRequest) (*models.FlightMission, error) {
	uav, err := s.uavRepo.FindByID(req.UAVID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	if uav.Status == models.UAVStatusFlying || uav.Status == models.UAVStatusHovering {
		activeMission, _ := s.missionRepo.GetActiveMissionByUAV(req.UAVID)
		if activeMission != nil {
			return nil, errors.New("uav has an active mission")
		}
	}

	mission := &models.FlightMission{
		Name:           req.Name,
		UAVID:          req.UAVID,
		TemplateID:     req.TemplateID,
		OperatorID:     req.OperatorID,
		Status:         models.MissionStatusPending,
		PlannedStart:   req.PlannedStart,
		PlannedEnd:     req.PlannedEnd,
		MaxAltitude:    req.MaxAltitude,
		Speed:          req.Speed,
		EnableGeofence: req.EnableGeofence,
		FailSafe:       req.FailSafe,
		Notes:          req.Notes,
		TotalWP:        len(req.Waypoints),
	}

	if err := s.missionRepo.CreateMission(mission); err != nil {
		return nil, err
	}

	for i := range req.Waypoints {
		req.Waypoints[i].MissionID = mission.ID
		req.Waypoints[i].TemplateID = 0
		if req.Waypoints[i].Seq == 0 {
			req.Waypoints[i].Seq = i
		}
		if err := s.missionRepo.AddWaypoint(&req.Waypoints[i]); err != nil {
			return nil, err
		}
	}

	return s.missionRepo.FindMissionByID(mission.ID)
}

func (s *MissionService) GetMission(id uint64) (*models.FlightMission, error) {
	return s.missionRepo.FindMissionByID(id)
}

func (s *MissionService) ListMissions(pagination *utils.Pagination, uavID uint64, status models.MissionStatus, operatorID uint64) ([]models.FlightMission, int64, error) {
	return s.missionRepo.ListMissions(pagination, uavID, status, operatorID)
}

func (s *MissionService) StartMission(id uint64) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	if mission.Status != models.MissionStatusPending && mission.Status != models.MissionStatusReady {
		return nil, errors.New("mission cannot be started from current status")
	}

	now := time.Now()
	mission.ActualStart = &now
	mission.Status = models.MissionStatusExecuting

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	return mission, nil
}

func (s *MissionService) PauseMission(id uint64) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	if mission.Status != models.MissionStatusExecuting {
		return nil, errors.New("mission is not executing")
	}

	mission.Status = models.MissionStatusPaused
	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusHovering)

	return mission, nil
}

func (s *MissionService) ResumeMission(id uint64) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	if mission.Status != models.MissionStatusPaused {
		return nil, errors.New("mission is not paused")
	}

	mission.Status = models.MissionStatusExecuting
	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	return mission, nil
}

func (s *MissionService) AbortMission(id uint64, reason string) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	now := time.Now()
	mission.Status = models.MissionStatusAborted
	mission.ActualEnd = &now
	mission.Notes += "\nAbort reason: " + reason

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)

	return mission, nil
}

func (s *MissionService) CompleteMission(id uint64) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	now := time.Now()
	mission.Status = models.MissionStatusCompleted
	mission.ActualEnd = &now
	mission.CurrentWP = mission.TotalWP

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusLanded)

	return mission, nil
}

func (s *MissionService) UpdateWaypointProgress(missionID uint64, wpIndex int) error {
	mission, err := s.missionRepo.FindMissionByID(missionID)
	if err != nil {
		return errors.New("mission not found")
	}

	if wpIndex >= 0 && wpIndex < len(mission.Waypoints) {
		_ = s.missionRepo.UpdateWaypointReached(mission.Waypoints[wpIndex].ID, true)
	}

	return s.missionRepo.UpdateCurrentWaypoint(missionID, wpIndex)
}

func (s *MissionService) GetActiveMission(uavID uint64) (*models.FlightMission, error) {
	return s.missionRepo.GetActiveMissionByUAV(uavID)
}

func (s *MissionService) GetMissionWaypoints(missionID uint64) ([]models.MissionWaypoint, error) {
	return s.missionRepo.GetMissionWaypoints(missionID)
}
