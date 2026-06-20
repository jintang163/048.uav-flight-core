package service

import (
	"context"
	"errors"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"strconv"
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

func (s *MissionService) CreateTemplate(template *models.MissionTemplate, creatorID uint64) (*models.MissionTemplate, error) {
	template.CreatorID = creatorID
	waypoints := template.Waypoints
	template.Waypoints = nil

	if err := s.missionRepo.CreateTemplate(template); err != nil {
		return nil, err
	}

	for i := range waypoints {
		waypoints[i].TemplateID = template.ID
		waypoints[i].MissionID = 0
		if waypoints[i].Seq == 0 {
			waypoints[i].Seq = i
		}
		if err := s.missionRepo.AddWaypoint(&waypoints[i]); err != nil {
			return nil, err
		}
	}

	return s.missionRepo.FindTemplateByID(template.ID)
}

func (s *MissionService) GetTemplate(id uint64) (*models.MissionTemplate, error) {
	return s.missionRepo.FindTemplateByID(id)
}

func (s *MissionService) ListTemplates(pagination *utils.Pagination, category string, keyword string) ([]models.MissionTemplate, int64, error) {
	return s.missionRepo.ListTemplatesByCategory(pagination, category, keyword)
}

func (s *MissionService) UpdateTemplate(id uint64, template *models.MissionTemplate, waypoints []models.MissionWaypoint) (*models.MissionTemplate, error) {
	existing, err := s.missionRepo.FindTemplateByID(id)
	if err != nil {
		return nil, errors.New("template not found")
	}

	if template.Name != "" {
		existing.Name = template.Name
	}
	if template.Description != "" {
		existing.Description = template.Description
	}
	if template.Category != "" {
		existing.Category = template.Category
	}

	if err := s.missionRepo.UpdateTemplate(existing); err != nil {
		return nil, err
	}

	if len(waypoints) > 0 {
		_ = s.missionRepo.DeleteTemplateWaypoints(id)
		for i := range waypoints {
			waypoints[i].TemplateID = id
			waypoints[i].MissionID = 0
			if waypoints[i].Seq == 0 {
				waypoints[i].Seq = i
			}
			if err := s.missionRepo.AddWaypoint(&waypoints[i]); err != nil {
				return nil, err
			}
		}
	}

	return s.missionRepo.FindTemplateByID(id)
}

func (s *MissionService) DeleteTemplate(id uint64) error {
	_, err := s.missionRepo.FindTemplateByID(id)
	if err != nil {
		return errors.New("template not found")
	}
	return s.missionRepo.SoftDelete(&models.MissionTemplate{}, id)
}

func (s *MissionService) CreateMission(mission *models.FlightMission) (*models.FlightMission, error) {
	uav, err := s.uavRepo.FindByID(mission.UAVID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	if uav.Status == models.UAVStatusFlying || uav.Status == models.UAVStatusHovering {
		activeMission, _ := s.missionRepo.GetActiveMissionByUAV(mission.UAVID)
		if activeMission != nil {
			return nil, errors.New("uav has an active mission")
		}
	}

	waypoints := mission.Waypoints
	mission.Waypoints = nil
	mission.Status = models.MissionStatusPending
	mission.TotalWP = len(waypoints)

	if err := s.missionRepo.CreateMission(mission); err != nil {
		return nil, err
	}

	for i := range waypoints {
		waypoints[i].MissionID = mission.ID
		waypoints[i].TemplateID = 0
		if waypoints[i].Seq == 0 {
			waypoints[i].Seq = i
		}
		if err := s.missionRepo.AddWaypoint(&waypoints[i]); err != nil {
			return nil, err
		}
	}

	return s.missionRepo.FindMissionByID(mission.ID)
}

func (s *MissionService) GetMission(id uint64) (*models.FlightMission, error) {
	return s.missionRepo.FindMissionByID(id)
}

func (s *MissionService) ListMissions(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.FlightMission, int64, error) {
	return s.missionRepo.ListMissionsFiltered(pagination, uavID, status, startTime, endTime)
}

func (s *MissionService) UpdateMission(id uint64, mission *models.FlightMission) (*models.FlightMission, error) {
	existing, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}
	if mission.Name != "" {
		existing.Name = mission.Name
	}
	if mission.Description != "" {
		existing.Description = mission.Description
	}
	if mission.MaxAltitude > 0 {
		existing.MaxAltitude = mission.MaxAltitude
	}
	if mission.MaxSpeed > 0 {
		existing.MaxSpeed = mission.MaxSpeed
	}
	if mission.Speed > 0 {
		existing.Speed = mission.Speed
	}
	if err := s.missionRepo.Update(existing); err != nil {
		return nil, err
	}
	return s.missionRepo.FindMissionByID(id)
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
	mission.CurrentWP = 0
	mission.ResumeFromWP = 0

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)

	s.sendMissionStartToMavlink(mission)
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
	mission.ResumeFromWP = mission.CurrentWP

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusHovering)
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)

	s.cacheBreakpoint(mission)
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
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)
	return mission, nil
}

func (s *MissionService) ResumeMissionFromBreakpoint(id uint64) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	if mission.Status != models.MissionStatusPaused && mission.Status != models.MissionStatusAborted {
		return nil, errors.New("mission must be paused or aborted to resume from breakpoint")
	}

	breakpoint := mission.ResumeFromWP
	if breakpoint <= 0 {
		breakpoint = s.loadBreakpointFromCache(mission.UAVID)
	}
	if breakpoint < 0 {
		breakpoint = mission.CurrentWP
	}
	if breakpoint >= mission.TotalWP {
		breakpoint = 0
	}

	mission.Status = models.MissionStatusExecuting
	mission.CurrentWP = breakpoint

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)

	s.sendSetCurrentWaypoint(mission.UAVID, uint16(breakpoint))
	return mission, nil
}

func (s *MissionService) SetCurrentWaypoint(id uint64, wpIndex int) (*models.FlightMission, error) {
	mission, err := s.missionRepo.FindMissionByID(id)
	if err != nil {
		return nil, errors.New("mission not found")
	}

	if wpIndex < 0 || (mission.TotalWP > 0 && wpIndex >= mission.TotalWP) {
		return nil, errors.New("invalid waypoint index")
	}

	mission.CurrentWP = wpIndex
	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)
	s.sendSetCurrentWaypoint(mission.UAVID, uint16(wpIndex))
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
	mission.ResumeFromWP = mission.CurrentWP
	if reason != "" {
		if mission.Notes == "" {
			mission.Notes = "Abort reason: " + reason
		} else {
			mission.Notes += "\nAbort reason: " + reason
		}
	}

	if err := s.missionRepo.Update(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)
	s.cacheBreakpoint(mission)
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
	websocket.BroadcastMissionStatus(mission.UAVID, string(mission.Status), mission.CurrentWP, mission.TotalWP)
	return mission, nil
}

func (s *MissionService) UpdateWaypointProgress(missionID uint64, wpIndex int) error {
	mission, err := s.missionRepo.FindMissionByID(missionID)
	if err != nil {
		return errors.New("mission not found")
	}

	if wpIndex >= 0 && wpIndex < len(mission.Waypoints) {
		_ = s.missionRepo.UpdateWaypointReached(mission.Waypoints[wpIndex].ID, true)
		websocket.BroadcastWaypointReached(mission.UAVID, wpIndex)
	}

	if err := s.missionRepo.UpdateCurrentWaypoint(missionID, wpIndex); err != nil {
		return err
	}

	websocket.BroadcastMissionProgress(mission.UAVID, missionID, wpIndex, mission.TotalWP, float64(wpIndex)/float64(mission.TotalWP))
	return nil
}

func (s *MissionService) GetActiveMission(uavID uint64) (*models.FlightMission, error) {
	return s.missionRepo.GetActiveMissionByUAV(uavID)
}

func (s *MissionService) GetMissionWaypoints(missionID uint64) ([]models.MissionWaypoint, error) {
	return s.missionRepo.GetMissionWaypoints(missionID)
}

func (s *MissionService) cacheBreakpoint(mission *models.FlightMission) {
	if config.Redis == nil {
		return
	}
	key := "mission:breakpoint:" + strconv.FormatUint(mission.UAVID, 10)
	_ = config.Redis.Set(context.Background(), key, mission.CurrentWP, 7*24*time.Hour).Err()
}

func (s *MissionService) loadBreakpointFromCache(uavID uint64) int {
	if config.Redis == nil {
		return -1
	}
	key := "mission:breakpoint:" + strconv.FormatUint(uavID, 10)
	val, err := config.Redis.Get(context.Background(), key).Int()
	if err != nil {
		return -1
	}
	return val
}

func (s *MissionService) sendMissionStartToMavlink(mission *models.FlightMission) {
	cmdMgr := mavlink.NewCommandManager()
	if cmdMgr == nil {
		return
	}
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_MISSION_START, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)
}

func (s *MissionService) sendSetCurrentWaypoint(uavID uint64, wpIndex uint16) {
	cmdMgr := mavlink.NewCommandManager()
	if cmdMgr == nil {
		return
	}
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_MISSION_SET_CURRENT, float32(wpIndex), 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(uavID, data)
}
