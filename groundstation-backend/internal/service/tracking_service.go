package service

import (
	"errors"
	"fmt"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"math"
	"sync"
	"time"
)

const (
	TargetCenterTolerance  = 0.05
	FramesToLock           = 10
	FramesToSearch         = 15
	FramesToLost           = 60
	SearchRadiusIncrement  = 5.0
	FrameCenterX           = 0.5
	FrameCenterY           = 0.5
	DefaultSearchRadius    = 10.0
	DefaultMaxSearchRadius = 50.0
)

type TrackingService struct {
	trackingRepo    *repository.TrackingRepository
	flightRepo      *repository.FlightRepository
	uavRepo         *repository.UAVRepository
	yolo            *YOLOv8Service
	mavlinkMgr      *mavlink.CommandManager
	mu              sync.RWMutex
	activeTasks     map[uint64]*models.TrackingTask
}

var trackingService *TrackingService
var trackingServiceOnce sync.Once

func NewTrackingService() *TrackingService {
	trackingServiceOnce.Do(func() {
		trackingService = &TrackingService{
			trackingRepo: repository.NewTrackingRepository(),
			flightRepo:   repository.NewFlightRepository(),
			uavRepo:      repository.NewUAVRepository(),
			yolo:         NewYOLOv8Service(),
			mavlinkMgr:   mavlink.GetCommandManager(),
			activeTasks:  make(map[uint64]*models.TrackingTask),
		}
	})
	return trackingService
}

func GetTrackingService() *TrackingService {
	return NewTrackingService()
}

type LockTargetRequest struct {
	UAVID        uint64
	BboxX        float64
	BboxY        float64
	BboxWidth    float64
	BboxHeight   float64
	FrameWidth   int
	FrameHeight  int
	TargetClass  models.DetectionClass
	Name         string
	CreatedBy    uint64
	SearchRadius float64
	MaxRadius    float64
}

func (s *TrackingService) LockTarget(req *LockTargetRequest) (*models.TrackingTask, error) {
	if req.UAVID == 0 {
		return nil, errors.New("uav id required")
	}

	existing, _ := s.trackingRepo.GetActiveTrackingByUAV(req.UAVID)
	if existing != nil {
		return nil, fmt.Errorf("uav %d already has active tracking task #%d", req.UAVID, existing.ID)
	}

	if req.SearchRadius <= 0 {
		req.SearchRadius = DefaultSearchRadius
	}
	if req.MaxRadius <= 0 {
		req.MaxRadius = DefaultMaxSearchRadius
	}

	now := time.Now()
	task := &models.TrackingTask{
		UAVID:             req.UAVID,
		Name:              req.Name,
		TargetClass:       req.TargetClass,
		Status:            models.TrackingStatusLocking,
		InitialBboxX:      req.BboxX,
		InitialBboxY:      req.BboxY,
		InitialBboxWidth:  req.BboxWidth,
		InitialBboxHeight: req.BboxHeight,
		CurrentBboxX:      &req.BboxX,
		CurrentBboxY:      &req.BboxY,
		CurrentBboxWidth:  &req.BboxWidth,
		CurrentBboxHeight: &req.BboxHeight,
		SearchRadius:      req.SearchRadius,
		MaxSearchRadius:   req.MaxRadius,
		FramesVisible:     0,
		FramesLost:        0,
		StartTime:         &now,
		CreatedBy:         req.CreatedBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.trackingRepo.CreateTracking(task); err != nil {
		return nil, fmt.Errorf("create tracking task: %w", err)
	}

	s.mu.Lock()
	s.activeTasks[task.UAVID] = task
	s.mu.Unlock()

	websocket.BroadcastTrackingUpdate(task.UAVID, task)

	return task, nil
}

func (s *TrackingService) ProcessDetections(uavID uint64, detections []*models.DetectionTarget, frameWidth, frameHeight int) error {
	s.mu.RLock()
	task, ok := s.activeTasks[uavID]
	s.mu.RUnlock()

	if !ok {
		task, _ = s.trackingRepo.GetActiveTrackingByUAV(uavID)
		if task == nil {
			return nil
		}
		s.mu.Lock()
		s.activeTasks[uavID] = task
		s.mu.Unlock()
	}

	matched := s.findMatchingDetection(task, detections, frameWidth, frameHeight)

	if matched != nil {
		s.handleTargetVisible(task, matched, frameWidth, frameHeight)
	} else {
		s.handleTargetLost(task)
	}

	s.trackingRepo.UpdateTracking(task)
	websocket.BroadcastTrackingUpdate(uavID, task)

	if task.Status == models.TrackingStatusTracking || task.Status == models.TrackingStatusSearching {
		s.sendTrackingCommand(task, frameWidth, frameHeight)
	}

	return nil
}

func (s *TrackingService) findMatchingDetection(task *models.TrackingTask, detections []*models.DetectionTarget, frameWidth, frameHeight int) *models.DetectionTarget {
	if len(detections) == 0 {
		return nil
	}

	var searchBoxX, searchBoxY, searchBoxW, searchBoxH float64
	if task.CurrentBboxX != nil {
		searchBoxX = *task.CurrentBboxX
		searchBoxY = *task.CurrentBboxY
		searchBoxW = *task.CurrentBboxWidth
		searchBoxH = *task.CurrentBboxHeight
	} else {
		searchBoxX = task.InitialBboxX
		searchBoxY = task.InitialBboxY
		searchBoxW = task.InitialBboxWidth
		searchBoxH = task.InitialBboxHeight
	}

	searchCenterX := searchBoxX + searchBoxW/2
	searchCenterY := searchBoxY + searchBoxH/2

	radiusFactor := 1.0
	if task.Status == models.TrackingStatusSearching {
		radiusFactor = 2.5
	}
	searchRadiusPx := math.Max(searchBoxW, searchBoxH) * radiusFactor

	var best *models.DetectionTarget
	bestScore := -1.0

	for _, d := range detections {
		if task.TargetClass != "" && d.Class != task.TargetClass && d.Class != models.DetectionClassUnknown {
			continue
		}

		dCenterX := d.BboxX + d.BboxWidth/2
		dCenterY := d.BboxY + d.BboxHeight/2
		dx := dCenterX - searchCenterX
		dy := dCenterY - searchCenterY
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > searchRadiusPx {
			continue
		}

		distanceNorm := distance / searchRadiusPx
		score := d.Confidence * (1.0 - distanceNorm*0.5)

		if score > bestScore {
			bestScore = score
			best = d
		}
	}

	return best
}

func (s *TrackingService) handleTargetVisible(task *models.TrackingTask, detection *models.DetectionTarget, frameWidth, frameHeight int) {
	task.CurrentBboxX = &detection.BboxX
	task.CurrentBboxY = &detection.BboxY
	task.CurrentBboxWidth = &detection.BboxWidth
	task.CurrentBboxHeight = &detection.BboxHeight
	task.Confidence = &detection.Confidence

	targetCenterX := detection.BboxX + detection.BboxWidth/2
	targetCenterY := detection.BboxY + detection.BboxHeight/2
	offsetX := (targetCenterX/float64(frameWidth) - FrameCenterX) * 2
	offsetY := (targetCenterY/float64(frameHeight) - FrameCenterY) * 2
	task.CenterOffsetX = &offsetX
	task.CenterOffsetY = &offsetY

	if detection.Latitude != 0 {
		task.TargetLatitude = &detection.Latitude
		task.TargetLongitude = &detection.Longitude
	}

	task.FramesVisible++
	task.FramesLost = 0

	switch task.Status {
	case models.TrackingStatusLocking:
		if task.FramesVisible >= FramesToLock {
			task.Status = models.TrackingStatusTracking
			task.SearchRadius = DefaultSearchRadius
		}
	case models.TrackingStatusSearching:
		task.Status = models.TrackingStatusTracking
		task.SearchRadius = DefaultSearchRadius
	}
}

func (s *TrackingService) handleTargetLost(task *models.TrackingTask) {
	task.FramesLost++
	task.FramesVisible = 0

	switch task.Status {
	case models.TrackingStatusTracking, models.TrackingStatusLocking:
		if task.FramesLost >= FramesToSearch {
			task.Status = models.TrackingStatusSearching
			if task.SearchRadius < task.MaxSearchRadius {
				task.SearchRadius = math.Min(task.SearchRadius+SearchRadiusIncrement, task.MaxSearchRadius)
			}
		}
	case models.TrackingStatusSearching:
		if task.FramesLost >= FramesToSearch && task.FramesLost%FramesToSearch == 0 {
			if task.SearchRadius < task.MaxSearchRadius {
				task.SearchRadius = math.Min(task.SearchRadius+SearchRadiusIncrement, task.MaxSearchRadius)
			}
		}
		if task.FramesLost >= FramesToLost {
			task.Status = models.TrackingStatusLost
			now := time.Now()
			task.EndTime = &now
			s.mu.Lock()
			delete(s.activeTasks, task.UAVID)
			s.mu.Unlock()
		}
	}
}

func (s *TrackingService) sendTrackingCommand(task *models.TrackingTask, frameWidth, frameHeight int) {
	var offsetX, offsetY float64
	if task.CenterOffsetX != nil {
		offsetX = *task.CenterOffsetX
	}
	if task.CenterOffsetY != nil {
		offsetY = *task.CenterOffsetY
	}

	maxVel := 3.0
	velocityE := -offsetX * maxVel
	velocityN := -offsetY * maxVel

	if math.Abs(offsetX) < TargetCenterTolerance {
		velocityE = 0
	}
	if math.Abs(offsetY) < TargetCenterTolerance {
		velocityN = 0
	}

	cmd := map[string]interface{}{
		"type":       "tracking",
		"velocity_n": velocityN,
		"velocity_e": velocityE,
		"velocity_d": 0.0,
		"yaw_rate":   -offsetX * 0.5,
		"search":     task.Status == models.TrackingStatusSearching,
		"radius":     task.SearchRadius,
	}

	if task.TargetLatitude != nil && task.TargetLongitude != nil {
		cmd["target_lat"] = *task.TargetLatitude
		cmd["target_lon"] = *task.TargetLongitude
	}

	s.mavlinkMgr.SendCustomCommand(task.UAVID, "SET_POSITION_TARGET_LOCAL_NED", cmd)
}

func (s *TrackingService) StopTracking(id uint64) error {
	task, err := s.trackingRepo.GetTrackingByID(id)
	if err != nil {
		return err
	}

	task.Status = models.TrackingStatusCompleted
	now := time.Now()
	task.EndTime = &now

	if err := s.trackingRepo.UpdateTracking(task); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.activeTasks, task.UAVID)
	s.mu.Unlock()

	websocket.BroadcastTrackingUpdate(task.UAVID, task)

	return nil
}

func (s *TrackingService) GetActiveTask(uavID uint64) (*models.TrackingTask, error) {
	s.mu.RLock()
	task, ok := s.activeTasks[uavID]
	s.mu.RUnlock()
	if ok {
		return task, nil
	}
	return s.trackingRepo.GetActiveTrackingByUAV(uavID)
}

func (s *TrackingService) ListTasks(uavID *uint64, status *models.TrackingStatus, page, pageSize int) ([]models.TrackingTask, int64, error) {
	return s.trackingRepo.ListTrackings(uavID, status, page, pageSize)
}

func (s *TrackingService) GetTask(id uint64) (*models.TrackingTask, error) {
	return s.trackingRepo.GetTrackingByID(id)
}

func (s *TrackingService) DetectAndTrack(uavID uint64, imageData []byte, frameWidth, frameHeight int) ([]*models.DetectionTarget, error) {
	results, err := s.yolo.DetectFromImage(imageData, frameWidth, frameHeight)
	if err != nil {
		return nil, err
	}

	status, _ := s.flightRepo.GetLatestStatus(uavID)
	var lat, lng, alt float64
	if status != nil {
		lat = status.Latitude
		lng = status.Longitude
		alt = status.RelativeAltitude
	}

	targets := s.yolo.ConvertToTargets(uavID, results, frameWidth, frameHeight, lat, lng, alt)
	if len(targets) > 0 {
		s.trackingRepo.BatchCreateDetections(targets)
	}

	s.ProcessDetections(uavID, targets, frameWidth, frameHeight)

	websocket.BroadcastDetections(uavID, targets)

	return targets, nil
}

func (s *TrackingService) ListDetections(uavID uint64, page, pageSize int) ([]models.DetectionTarget, int64, error) {
	return s.trackingRepo.ListDetectionsByUAV(uavID, page, pageSize)
}
