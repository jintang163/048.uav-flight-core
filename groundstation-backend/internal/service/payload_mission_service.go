package service

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"

	"github.com/google/uuid"
)

type PayloadMissionService struct {
	missionRepo  *repository.PayloadMissionRepository
	payloadRepo  *repository.PayloadRepository
	uavRepo      *repository.UAVRepository
	missionSvc   *MissionService
	payloadSvc   *PayloadService
}

func NewPayloadMissionService() *PayloadMissionService {
	return &PayloadMissionService{
		missionRepo: repository.NewPayloadMissionRepository(),
		payloadRepo: repository.NewPayloadRepository(),
		uavRepo:     repository.NewUAVRepository(),
		missionSvc:  NewMissionService(),
		payloadSvc:  NewPayloadService(),
	}
}

func (s *PayloadMissionService) CreateOrbitMission(mission *models.OrbitMission, creatorID uint64) (*models.OrbitMission, error) {
	if _, err := s.uavRepo.FindByID(mission.UAVID); err != nil {
		return nil, errors.New("uav not found")
	}

	if mission.Radius <= 0 {
		mission.Radius = 30
	}
	if mission.Altitude <= 0 {
		mission.Altitude = 50
	}
	if mission.Loops <= 0 {
		mission.Loops = 1
	}
	if mission.Velocity <= 0 {
		mission.Velocity = 5
	}
	if mission.Direction == 0 {
		mission.Direction = 1
	}
	if mission.CaptureInterval <= 0 {
		mission.CaptureInterval = 5
	}

	mission.UUID = uuid.New().String()
	mission.Status = models.OrbitStatusPending
	mission.CreatorID = creatorID

	if err := s.missionRepo.CreateOrbitMission(mission); err != nil {
		return nil, err
	}

	return s.missionRepo.FindOrbitMissionByID(mission.ID)
}

func (s *PayloadMissionService) GetOrbitMission(id uint64) (*models.OrbitMission, error) {
	return s.missionRepo.FindOrbitMissionByID(id)
}

func (s *PayloadMissionService) ListOrbitMissions(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.OrbitMission, int64, error) {
	return s.missionRepo.ListOrbitMissions(pagination, uavID, status, startTime, endTime)
}

func (s *PayloadMissionService) StartOrbitMission(id uint64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	if mission.Status != models.OrbitStatusPending && mission.Status != models.OrbitStatusPaused {
		return nil, errors.New("mission cannot be started from current status")
	}

	now := time.Now()
	mission.Status = models.OrbitStatusActive
	mission.StartAt = &now
	mission.CurrentLoop = 0
	mission.Progress = 0

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	s.sendOrbitCommand(mission)
	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) PauseOrbitMission(id uint64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	if mission.Status != models.OrbitStatusActive {
		return nil, errors.New("mission is not active")
	}

	mission.Status = models.OrbitStatusPaused

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusHovering)

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_DO_LOITER_UNLIMITED, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) ResumeOrbitMission(id uint64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	if mission.Status != models.OrbitStatusPaused {
		return nil, errors.New("mission is not paused")
	}

	mission.Status = models.OrbitStatusActive

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	s.sendOrbitCommand(mission)
	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) AbortOrbitMission(id uint64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	now := time.Now()
	mission.Status = models.OrbitStatusAborted
	mission.EndAt = &now

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) CompleteOrbitMission(id uint64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	now := time.Now()
	mission.Status = models.OrbitStatusCompleted
	mission.EndAt = &now
	mission.Progress = 100
	mission.CurrentLoop = mission.Loops

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)

	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) UpdateOrbitProgress(id uint64, currentLoop int, progress float64) (*models.OrbitMission, error) {
	mission, err := s.missionRepo.FindOrbitMissionByID(id)
	if err != nil {
		return nil, errors.New("orbit mission not found")
	}

	mission.CurrentLoop = currentLoop
	mission.Progress = progress

	if currentLoop >= mission.Loops && progress >= 100 {
		return s.CompleteOrbitMission(id)
	}

	if err := s.missionRepo.UpdateOrbitMission(mission); err != nil {
		return nil, err
	}

	s.broadcastOrbitStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) sendOrbitCommand(mission *models.OrbitMission) {
	cmdMgr := mavlink.NewCommandManager()

	radius := float32(mission.Radius)
	velocity := float32(mission.Velocity)
	direction := float32(mission.Direction)
	lat := float32(mission.CenterLatitude)
	lng := float32(mission.CenterLongitude)
	alt := float32(mission.Altitude)

	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_NAV_LOITER_TURNS,
		float32(mission.Loops), radius, velocity, direction, lat, lng, alt)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	if mission.AutoCapture {
		intervalData := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_DO_DIGICAM_CONTROL, 4,
			float32(mission.CaptureInterval), 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(mission.UAVID, intervalData)
	}
}

func (s *PayloadMissionService) broadcastOrbitStatus(mission *models.OrbitMission) {
	hub := websocket.NewHub()
	hub.BroadcastMissionStatus(mission.UAVID, map[string]interface{}{
		"type":         "orbit_mission",
		"mission_id":   mission.ID,
		"status":       mission.Status,
		"current_loop": mission.CurrentLoop,
		"total_loops":  mission.Loops,
		"progress":     mission.Progress,
	})
}

func (s *PayloadMissionService) CreateOrthoMission(mission *models.OrthoMission, creatorID uint64) (*models.OrthoMission, error) {
	if _, err := s.uavRepo.FindByID(mission.UAVID); err != nil {
		return nil, errors.New("uav not found")
	}

	if mission.Altitude <= 0 {
		mission.Altitude = 100
	}
	if mission.Speed <= 0 {
		mission.Speed = 8
	}
	if mission.OverlapFront <= 0 {
		mission.OverlapFront = 80
	}
	if mission.OverlapSide <= 0 {
		mission.OverlapSide = 70
	}
	if mission.TriggerMode == "" {
		mission.TriggerMode = "distance"
	}

	mission.UUID = uuid.New().String()
	mission.Status = models.OrthoStatusPending
	mission.CreatorID = creatorID
	mission.Progress = 0
	mission.CapturedPhotos = 0

	if err := s.missionRepo.CreateOrthoMission(mission); err != nil {
		return nil, err
	}

	return s.missionRepo.FindOrthoMissionByID(mission.ID)
}

func (s *PayloadMissionService) PlanOrthoMission(id uint64, areaCoords [][]float64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	mission.Status = models.OrthoStatusPlanning

	waypoints, totalDistance, totalPhotos := s.calculateOrthoWaypoints(
		areaCoords, mission.Altitude, mission.Speed, mission.OverlapFront, mission.OverlapSide,
	)

	_ = s.missionRepo.DeleteOrthoWaypoints(id)

	for i := range waypoints {
		waypoints[i].OrthoMissionID = id
		waypoints[i].Seq = i
	}

	if err := s.missionRepo.AddOrthoWaypoints(waypoints); err != nil {
		return nil, err
	}

	areaGeoJSON, _ := json.Marshal(map[string]interface{}{
		"type":        "Polygon",
		"coordinates": [][][]float64{areaCoords},
	})

	mission.SurveyArea = string(areaGeoJSON)
	mission.TotalDistance = totalDistance
	mission.TotalPhotos = totalPhotos
	mission.TotalWaypoints = len(waypoints)
	mission.CurrentWaypoint = 0
	mission.Progress = 0

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	return s.missionRepo.FindOrthoMissionByID(id)
}

func (s *PayloadMissionService) calculateOrthoWaypoints(areaCoords [][]float64, altitude, speed, overlapFront, overlapSide float64) ([]models.OrthoWaypoint, float64, int) {
	if len(areaCoords) < 3 {
		return nil, 0, 0
	}

	var minLat, maxLat, minLng, maxLng float64
	minLat, maxLat = areaCoords[0][0], areaCoords[0][0]
	minLng, maxLng = areaCoords[0][1], areaCoords[0][1]

	for _, coord := range areaCoords {
		if coord[0] < minLat {
			minLat = coord[0]
		}
		if coord[0] > maxLat {
			maxLat = coord[0]
		}
		if coord[1] < minLng {
			minLng = coord[1]
		}
		if coord[1] > maxLng {
			maxLng = coord[1]
		}
	}

	footprintWidth := altitude * 2 * math.Tan(0.52) * 0.3048
	footprintHeight := altitude * 2 * math.Tan(0.39) * 0.3048

	lapSpacing := footprintHeight * (1 - overlapSide/100)
	photoSpacing := footprintWidth * (1 - overlapFront/100)

	latDiff := maxLat - minLat
	lngDiff := maxLng - minLng

	centerLat := (minLat + maxLat) / 2
	metersPerDegLat := 111320.0
	metersPerDegLng := 111320.0 * math.Cos(centerLat*math.Pi/180)

	areaWidth := lngDiff * metersPerDegLng
	areaHeight := latDiff * metersPerDegLat

	numLaps := int(math.Ceil(areaHeight / lapSpacing))
	photosPerLap := int(math.Ceil(areaWidth / photoSpacing))

	var waypoints []models.OrthoWaypoint
	var totalDistance float64

	for lap := 0; lap < numLaps; lap++ {
		latOffset := float64(lap) * lapSpacing / metersPerDegLat
		lat := minLat + latOffset

		direction := 1
		if lap%2 == 1 {
			direction = -1
		}

		for photo := 0; photo <= photosPerLap; photo++ {
			var lngOffset float64
			if direction == 1 {
				lngOffset = float64(photo) * photoSpacing / metersPerDegLng
			} else {
				lngOffset = float64(photosPerLap-photo) * photoSpacing / metersPerDegLng
			}
			lng := minLng + lngOffset

			isTurnPoint := photo == 0 || photo == photosPerLap
			triggerPhoto := !isTurnPoint

			wp := models.OrthoWaypoint{
				Latitude:     lat,
				Longitude:    lng,
				Altitude:     altitude,
				IsTurnPoint:  isTurnPoint,
				TriggerPhoto: triggerPhoto,
			}
			waypoints = append(waypoints, wp)

			if photo > 0 {
				totalDistance += photoSpacing
			}
		}

		if lap < numLaps-1 {
			totalDistance += lapSpacing
		}
	}

	totalPhotos := numLaps * (photosPerLap - 1)

	return waypoints, totalDistance, totalPhotos
}

func (s *PayloadMissionService) GetOrthoMission(id uint64) (*models.OrthoMission, error) {
	return s.missionRepo.FindOrthoMissionByID(id)
}

func (s *PayloadMissionService) ListOrthoMissions(pagination *utils.Pagination, uavID uint64, status string, startTime string, endTime string) ([]models.OrthoMission, int64, error) {
	return s.missionRepo.ListOrthoMissions(pagination, uavID, status, startTime, endTime)
}

func (s *PayloadMissionService) StartOrthoMission(id uint64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	if mission.Status != models.OrthoStatusPending &&
		mission.Status != models.OrthoStatusPlanning &&
		mission.Status != models.OrthoStatusPaused {
		return nil, errors.New("mission cannot be started from current status")
	}

	if mission.TotalWaypoints == 0 {
		return nil, errors.New("mission has not been planned yet")
	}

	now := time.Now()
	mission.Status = models.OrthoStatusActive
	mission.StartAt = &now

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	s.uploadOrthoMissionToFlight(mission)
	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) uploadOrthoMissionToFlight(mission *models.OrthoMission) {
	waypoints, _ := s.missionRepo.GetOrthoWaypoints(mission.ID)
	if len(waypoints) == 0 {
		return
	}

	cmdMgr := mavlink.NewCommandManager()

	countData := mavlink.EncodeMissionCount(uint16(len(waypoints)))
	_ = cmdMgr.SendCommand(mission.UAVID, countData)

	for i, wp := range waypoints {
		itemData := mavlink.EncodeMissionItem(uint16(i), wp.Latitude, wp.Longitude, wp.Altitude)
		_ = cmdMgr.SendCommand(mission.UAVID, itemData)
	}

	startData := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_MISSION_START, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, startData)

	if mission.TriggerMode == "distance" && mission.TriggerDistance > 0 {
		triggerData := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_DO_DIGICAM_CONFIGURE, 3,
			float32(mission.TriggerDistance), 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(mission.UAVID, triggerData)
	} else if mission.TriggerMode == "time" && mission.TriggerInterval > 0 {
		triggerData := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_DO_DIGICAM_CONFIGURE, 4,
			float32(mission.TriggerInterval), 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(mission.UAVID, triggerData)
	}
}

func (s *PayloadMissionService) PauseOrthoMission(id uint64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	if mission.Status != models.OrthoStatusActive {
		return nil, errors.New("mission is not active")
	}

	mission.Status = models.OrthoStatusPaused

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusHovering)

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_DO_LOITER_UNLIMITED, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) ResumeOrthoMission(id uint64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	if mission.Status != models.OrthoStatusPaused {
		return nil, errors.New("mission is not paused")
	}

	mission.Status = models.OrthoStatusActive

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusFlying)

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_MISSION_START, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) AbortOrthoMission(id uint64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	now := time.Now()
	mission.Status = models.OrthoStatusAborted
	mission.EndAt = &now

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(mission.UAVID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(mission.UAVID, data)

	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) CompleteOrthoMission(id uint64) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	now := time.Now()
	mission.Status = models.OrthoStatusCompleted
	mission.EndAt = &now
	mission.Progress = 100
	mission.CapturedPhotos = mission.TotalPhotos
	mission.CurrentWaypoint = mission.TotalWaypoints

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	_ = s.uavRepo.UpdateStatus(mission.UAVID, models.UAVStatusOnline)

	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) UpdateOrthoWaypointProgress(id uint64, wpIndex int) (*models.OrthoMission, error) {
	mission, err := s.missionRepo.FindOrthoMissionByID(id)
	if err != nil {
		return nil, errors.New("ortho mission not found")
	}

	waypoints, _ := s.missionRepo.GetOrthoWaypoints(id)
	if wpIndex >= 0 && wpIndex < len(waypoints) {
		_ = s.missionRepo.UpdateOrthoWaypointReached(waypoints[wpIndex].ID, true)
		if waypoints[wpIndex].TriggerPhoto {
			mission.CapturedPhotos++
		}
	}

	mission.CurrentWaypoint = wpIndex
	if mission.TotalWaypoints > 0 {
		mission.Progress = float64(wpIndex) / float64(mission.TotalWaypoints) * 100
	}

	if wpIndex >= mission.TotalWaypoints-1 {
		return s.CompleteOrthoMission(id)
	}

	if err := s.missionRepo.UpdateOrthoMission(mission); err != nil {
		return nil, err
	}

	s.broadcastOrthoStatus(mission)

	return mission, nil
}

func (s *PayloadMissionService) broadcastOrthoStatus(mission *models.OrthoMission) {
	hub := websocket.NewHub()
	hub.BroadcastMissionStatus(mission.UAVID, map[string]interface{}{
		"type":              "ortho_mission",
		"mission_id":        mission.ID,
		"status":            mission.Status,
		"current_waypoint":  mission.CurrentWaypoint,
		"total_waypoints":   mission.TotalWaypoints,
		"captured_photos":   mission.CapturedPhotos,
		"total_photos":      mission.TotalPhotos,
		"progress":          mission.Progress,
	})
}

func (s *PayloadMissionService) CreateTTSTask(task *models.TextToSpeechTask, creatorID uint64) (*models.TextToSpeechTask, error) {
	if _, err := s.uavRepo.FindByID(task.UAVID); err != nil {
		return nil, errors.New("uav not found")
	}
	if task.Text == "" {
		return nil, errors.New("text content is required")
	}

	task.UUID = uuid.New().String()
	task.CreatedBy = creatorID
	task.Status = "pending"
	if task.Language == "" {
		task.Language = "zh-CN"
	}
	if task.Speed <= 0 {
		task.Speed = 1.0
	}
	if task.Pitch <= 0 {
		task.Pitch = 1.0
	}
	if task.Volume <= 0 {
		task.Volume = 80
	}

	if err := s.missionRepo.CreateTTSTask(task); err != nil {
		return nil, err
	}

	go s.processTTSTask(task.ID)

	return s.missionRepo.FindTTSTaskByID(task.ID)
}

func (s *PayloadMissionService) processTTSTask(taskID uint64) {
	task, err := s.missionRepo.FindTTSTaskByID(taskID)
	if err != nil {
		return
	}

	_ = s.missionRepo.UpdateTTSStatus(taskID, "processing", "", "", 0)

	audio := &models.SpeakerAudio{
		PayloadID:      task.PayloadID,
		Name:           "TTS_" + task.UUID,
		Type:           "tts",
		Content:        task.Text,
		IsTextToSpeech: true,
		Voice:          task.Voice,
		Speed:          task.Speed,
		Pitch:          task.Pitch,
		Volume:         task.Volume,
		CreatedBy:      task.CreatedBy,
		DurationSec:    len(task.Text) * 2,
	}

	createdAudio, err := s.payloadSvc.CreateSpeakerAudio(audio, task.CreatedBy)
	if err != nil {
		_ = s.missionRepo.UpdateTTSStatus(taskID, "failed", err.Error(), "", 0)
		return
	}

	audioURL := "/api/v1/payloads/speaker/audios/" + createdAudio.UUID
	duration := len(task.Text) * 2

	_ = s.missionRepo.UpdateTTSStatus(taskID, "completed", "", audioURL, duration)

	if task.PayloadID > 0 {
		_ = s.payloadSvc.PlaySpeakerAudio(task.UAVID, task.PayloadID, createdAudio.ID)
	}
}

func (s *PayloadMissionService) GetTTSTask(id uint64) (*models.TextToSpeechTask, error) {
	return s.missionRepo.FindTTSTaskByID(id)
}

func (s *PayloadMissionService) ListTTSTasks(pagination *utils.Pagination, payloadID uint64, uavID uint64, status string) ([]models.TextToSpeechTask, int64, error) {
	return s.missionRepo.ListTTSTasks(pagination, payloadID, uavID, status)
}
