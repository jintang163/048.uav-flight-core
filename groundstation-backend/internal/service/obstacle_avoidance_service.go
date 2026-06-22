package service

import (
	"encoding/json"
	"fmt"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"time"
)

type ObstacleAvoidanceService struct {
	repo *repository.ObstacleAvoidanceRepository
}

func NewObstacleAvoidanceService() *ObstacleAvoidanceService {
	return &ObstacleAvoidanceService{
		repo: repository.NewObstacleAvoidanceRepository(),
	}
}

func (s *ObstacleAvoidanceService) GetConfig(uavID uint64) (*models.ObstacleAvoidanceConfig, error) {
	config, err := s.repo.GetConfig(uavID)
	if err != nil {
		config = &models.ObstacleAvoidanceConfig{
			UAVID:          uavID,
			Enabled:        true,
			Sensitivity:    models.SensitivityMedium,
			Strategy:       models.StrategyAscendBypass,
			SensorType:     models.SensorTypeMillimeterWave,
			DetectionRange: 15,
			AscendHeight:   5,
			RetreatDistance: 10,
			BypassAngle:    45,
		}
		if createErr := s.repo.CreateConfig(config); createErr != nil {
			return nil, createErr
		}
	}
	return config, nil
}

func (s *ObstacleAvoidanceService) UpdateConfig(uavID uint64, updates map[string]interface{}) (*models.ObstacleAvoidanceConfig, error) {
	if err := s.repo.UpdateConfig(uavID, updates); err != nil {
		return nil, err
	}
	return s.repo.GetConfig(uavID)
}

func (s *ObstacleAvoidanceService) RecordDetection(uavID uint64, detection *models.ObstacleDetectionLog) error {
	if err := s.repo.CreateDetectionLog(detection); err != nil {
		return err
	}
	return s.repo.UpsertHeatmapPoint(uavID, detection.Latitude, detection.Longitude, detection.Altitude, detection.Distance)
}

func (s *ObstacleAvoidanceService) CreateAvoidanceEvent(uavID uint64, strategy models.AvoidanceStrategy, detectionID uint64, sensorType models.ObstacleSensorType, direction models.ObstacleDirection, obstacleDistance, startLat, startLng, startAlt float64, bypassPath []BypassWaypoint) (*models.ObstacleAvoidanceEvent, error) {
	pathJSON, err := json.Marshal(bypassPath)
	if err != nil {
		pathJSON = []byte("[]")
	}
	event := &models.ObstacleAvoidanceEvent{
		UAVID:            uavID,
		Strategy:         strategy,
		Status:           models.AvoidanceStatusTriggered,
		DetectionID:      detectionID,
		SensorType:       sensorType,
		Direction:        direction,
		ObstacleDistance:  obstacleDistance,
		StartLat:         startLat,
		StartLng:         startLng,
		StartAlt:         startAlt,
		BypassPath:       string(pathJSON),
	}
	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *ObstacleAvoidanceService) UpdateAvoidanceEvent(eventID uint64, status models.AvoidanceActionStatus, failReason string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == models.AvoidanceStatusCompleted {
		updates["completed_at"] = time.Now()
	}
	if failReason != "" {
		updates["fail_reason"] = failReason
	}
	return s.repo.UpdateEvent(eventID, updates)
}

func (s *ObstacleAvoidanceService) GetEvent(eventID uint64) (*models.ObstacleAvoidanceEvent, error) {
	return s.repo.GetEvent(eventID)
}

func (s *ObstacleAvoidanceService) ListEvents(uavID uint64, status string, page, pageSize int) ([]models.ObstacleAvoidanceEvent, int64, error) {
	pagination := &repository.Pagination{Page: page, PageSize: pageSize}
	return s.repo.ListEvents(uavID, status, pagination)
}

func (s *ObstacleAvoidanceService) GetHeatmap(uavID uint64, startTime, endTime time.Time) ([]models.ObstacleHeatmapPoint, error) {
	return s.repo.GetHeatmapPoints(uavID, startTime, endTime)
}

func (s *ObstacleAvoidanceService) ClearHeatmap(uavID uint64) error {
	return s.repo.ClearHeatmap(uavID)
}

func (s *ObstacleAvoidanceService) GetStatistics(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	return s.repo.GetStatistics(uavID, startTime, endTime)
}

func (s *ObstacleAvoidanceService) TriggerAvoidanceTest(uavID uint64, strategy models.AvoidanceStrategy) error {
	config, err := s.GetConfig(uavID)
	if err != nil {
		return err
	}
	if !config.Enabled {
		return fmt.Errorf("obstacle avoidance is not enabled for UAV %d", uavID)
	}
	return nil
}

type BypassWaypoint struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Alt       float64 `json:"alt"`
	Timestamp int64   `json:"timestamp"`
	Type      string  `json:"type"`
}
