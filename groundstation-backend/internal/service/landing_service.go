package service

import (
	"errors"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	CMD_NAV_LAND                 = 22
	CMD_NAV_PRECISION_LAND       = 25
	CMD_DO_SET_MODE              = 176
	MAV_MODE_GUIDED              = 4
	MAV_MODE_GUIDED_ARMED        = 92

	LANDING_PHASE_APPROACH       = "approach"
	LANDING_PHASE_DESCEND        = "descend"
	LANDING_PHASE_PRECISION      = "precision"
	LANDING_PHASE_TOUCHDOWN      = "touchdown"

	RTK_FIX_TYPE_NONE            = 0
	RTK_FIX_TYPE_SINGLE          = 1
	RTK_FIX_TYPE_DGPS            = 2
	RTK_FIX_TYPE_PPP             = 3
	RTK_FIX_TYPE_RTK_FIXED       = 4
	RTK_FIX_TYPE_RTK_FLOAT       = 5

	MARKER_TYPE_QR_CODE          = "qr_code"
	MARKER_TYPE_H_MARKER         = "h_marker"
	MARKER_TYPE_APRILTAG         = "apriltag"
	MARKER_TYPE_ARUCO            = "aruco"

	PRECISION_LANDING_THRESHOLD  = 0.10
	VISION_DETECTION_THRESHOLD   = 0.85
)

type LandingService struct {
	landingRepo    *repository.LandingRepository
	uavRepo        *repository.UAVRepository
	missionRepo    *repository.MissionRepository
	alertService   *AlertService
	mu             sync.RWMutex
	activeSessions map[uint64]*models.LandingSession
	trajectorySeq  map[uint64]int
}

type LandingPlan struct {
	PrimaryLanding   *models.LandingPoint
	AlternateLandings []*models.LandingPoint
	MissionID        uint64
	RTKEnabled       bool
	VisionEnabled    bool
	MovingPlatform   bool
}

var landingService *LandingService
var landingOnce sync.Once

func NewLandingService() *LandingService {
	landingOnce.Do(func() {
		landingService = &LandingService{
			landingRepo:    repository.NewLandingRepository(),
			uavRepo:        repository.NewUAVRepository(),
			missionRepo:    repository.NewMissionRepository(),
			alertService:   NewAlertService(),
			activeSessions: make(map[uint64]*models.LandingSession),
			trajectorySeq:  make(map[uint64]int),
		}
	})
	return landingService
}

func (s *LandingService) CreateLandingPoint(point *models.LandingPoint) (*models.LandingPoint, error) {
	if err := s.validateLandingPoint(point); err != nil {
		return nil, err
	}
	if err := s.landingRepo.CreateLandingPoint(point); err != nil {
		return nil, err
	}
	return s.landingRepo.FindLandingPointByID(point.ID)
}

func (s *LandingService) validateLandingPoint(point *models.LandingPoint) error {
	if point.Latitude < -90 || point.Latitude > 90 {
		return errors.New("invalid latitude")
	}
	if point.Longitude < -180 || point.Longitude > 180 {
		return errors.New("invalid longitude")
	}
	if point.Radius <= 0 {
		return errors.New("radius must be positive")
	}
	return nil
}

func (s *LandingService) GetLandingPoint(id uint64) (*models.LandingPoint, error) {
	return s.landingRepo.FindLandingPointByID(id)
}

func (s *LandingService) ListLandingPoints(page, pageSize int, pointType string, status string, hasMarkers *bool) ([]models.LandingPoint, int64, error) {
	pagination := &utils.Pagination{Page: page, PageSize: pageSize}
	return s.landingRepo.ListLandingPoints(pagination, models.LandingPointType(pointType), models.LandingPointStatus(status), hasMarkers)
}

func (s *LandingService) UpdateLandingPoint(id uint64, point *models.LandingPoint) (*models.LandingPoint, error) {
	existing, err := s.landingRepo.FindLandingPointByID(id)
	if err != nil {
		return nil, errors.New("landing point not found")
	}
	if point.Name != "" {
		existing.Name = point.Name
	}
	if point.Description != "" {
		existing.Description = point.Description
	}
	if point.Type != "" {
		existing.Type = point.Type
	}
	if point.Latitude != 0 {
		existing.Latitude = point.Latitude
	}
	if point.Longitude != 0 {
		existing.Longitude = point.Longitude
	}
	if point.Altitude >= 0 {
		existing.Altitude = point.Altitude
	}
	if point.Radius > 0 {
		existing.Radius = point.Radius
	}
	existing.HasMarkers = point.HasMarkers
	existing.MarkerType = point.MarkerType
	existing.IsMovingPlatform = point.IsMovingPlatform
	existing.MovingPlatformID = point.MovingPlatformID
	existing.Priority = point.Priority
	existing.Status = point.Status

	if err := s.validateLandingPoint(existing); err != nil {
		return nil, err
	}
	if err := s.landingRepo.UpdateLandingPoint(existing); err != nil {
		return nil, err
	}
	return s.landingRepo.FindLandingPointByID(id)
}

func (s *LandingService) DeleteLandingPoint(id uint64) error {
	_, err := s.landingRepo.FindLandingPointByID(id)
	if err != nil {
		return errors.New("landing point not found")
	}
	return s.landingRepo.DeleteLandingPoint(id)
}

func (s *LandingService) PlanLanding(uavID uint64, plan *LandingPlan) (*models.LandingSession, error) {
	uav, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}
	if uav.Status != models.UAVStatusFlying && uav.Status != models.UAVStatusHovering {
		return nil, errors.New("uav is not in flight")
	}

	_, err = s.landingRepo.GetActiveLandingSession(uavID)
	if err == nil {
		return nil, errors.New("uav already has an active landing session")
	}

	var primaryLanding *models.LandingPoint
	if plan.PrimaryLanding != nil {
		primaryLanding = plan.PrimaryLanding
	} else {
		primaryPoints, err := s.landingRepo.ListLandingPointsByType(models.LandingPointTypePrimary)
		if err != nil || len(primaryPoints) == 0 {
			return nil, errors.New("no primary landing point available")
		}
		primaryLanding = &primaryPoints[0]
	}

	var alternateLanding *models.LandingPoint
	if len(plan.AlternateLandings) > 0 {
		alternateLanding = plan.AlternateLandings[0]
	} else {
		alternatePoints, err := s.landingRepo.ListLandingPointsByType(models.LandingPointTypeAlternate)
		if err == nil && len(alternatePoints) > 0 {
			alternateLanding = &alternatePoints[0]
		}
	}

	now := time.Now()
	session := &models.LandingSession{
		UAVID:              uavID,
		MissionID:          plan.MissionID,
		PrimaryLandingID:   primaryLanding.ID,
		Status:             models.LandingSessionStatusPending,
		RTKEnabled:         plan.RTKEnabled,
		VisionEnabled:      plan.VisionEnabled,
		IsMovingPlatform:   plan.MovingPlatform,
		TargetLatitude:     primaryLanding.Latitude,
		TargetLongitude:    primaryLanding.Longitude,
		TargetAltitude:     primaryLanding.Altitude,
		StartTime:          &now,
	}

	if alternateLanding != nil {
		session.AlternateLandingID = alternateLanding.ID
	}

	if err := s.landingRepo.CreateLandingSession(session); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.activeSessions[uavID] = session
	s.trajectorySeq[uavID] = 0
	s.mu.Unlock()

	if err := s.landingRepo.UpdateLandingPointStatus(primaryLanding.ID, models.LandingPointStatusOccupied); err != nil {
		middleware.Logger.Warn("Failed to update landing point status", zap.Error(err))
	}

	websocket.BroadcastLandingStatus(uavID, string(session.Status), primaryLanding.ID, session.RTKEnabled, session.VisionEnabled)

	middleware.Logger.Info("Landing session created",
		zap.Uint64("uav_id", uavID),
		zap.Uint64("session_id", session.ID),
		zap.Uint64("primary_landing_id", primaryLanding.ID),
	)

	return s.landingRepo.FindLandingSessionByID(session.ID)
}

func (s *LandingService) StartLanding(uavID uint64, sessionID uint64) (*models.LandingSession, error) {
	session, err := s.landingRepo.FindLandingSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("landing session not found")
	}
	if session.UAVID != uavID {
		return nil, errors.New("session does not belong to this uav")
	}
	if session.Status != models.LandingSessionStatusPending {
		return nil, errors.New("landing session is not in pending state")
	}

	session.Status = models.LandingSessionStatusApproaching
	if err := s.landingRepo.UpdateLandingSession(session); err != nil {
		return nil, err
	}

	s.sendLandingCommand(uavID, session)

	s.mu.Lock()
	s.activeSessions[uavID] = session
	s.mu.Unlock()

	websocket.BroadcastLandingStatus(uavID, string(session.Status), session.PrimaryLandingID, session.RTKEnabled, session.VisionEnabled)

	middleware.Logger.Info("Landing started",
		zap.Uint64("uav_id", uavID),
		zap.Uint64("session_id", sessionID),
	)

	return session, nil
}

func (s *LandingService) sendLandingCommand(uavID uint64, session *models.LandingSession) {
	cm := mavlink.NewCommandManager()
	if cm == nil {
		return
	}

	modeData := mavlink.EncodeCommandLong(uavID, CMD_DO_SET_MODE, MAV_MODE_GUIDED_ARMED, 0, 0, 0, 0, 0, 0)
	_ = cm.SendCommand(uavID, modeData)

	time.Sleep(100 * time.Millisecond)

	var landData []byte
	if session.VisionEnabled || session.RTKEnabled {
		landData = mavlink.EncodeCommandLong(uavID, CMD_NAV_PRECISION_LAND, float32(session.TargetLatitude), float32(session.TargetLongitude), float32(session.TargetAltitude), 0, 0, 0, 0)
	} else {
		landData = mavlink.EncodeCommandLong(uavID, CMD_NAV_LAND, 0, 0, 0, 0, 0, 0, 0)
	}
	_ = cm.SendCommand(uavID, landData)
}

func (s *LandingService) RecordTrajectoryPoint(uavID uint64, telemetry *models.FlightStatus) error {
	s.mu.RLock()
	session, exists := s.activeSessions[uavID]
	s.mu.RUnlock()

	if !exists || session == nil {
		var err error
		session, err = s.landingRepo.GetActiveLandingSession(uavID)
		if err != nil {
			return nil
		}
		s.mu.Lock()
		s.activeSessions[uavID] = session
		s.mu.Unlock()
	}

	s.mu.Lock()
	seq := s.trajectorySeq[uavID]
	s.trajectorySeq[uavID]++
	s.mu.Unlock()

	phase := s.determineLandingPhase(session, telemetry)

	point := &models.LandingTrajectoryPoint{
		SessionID:      session.ID,
		Sequence:       seq,
		Timestamp:      time.Now(),
		Latitude:       telemetry.Latitude,
		Longitude:      telemetry.Longitude,
		AltitudeMSL:    telemetry.AltitudeMSL,
		AltitudeRel:    telemetry.AltitudeRel,
		VelocityX:      telemetry.VelocityX,
		VelocityY:      telemetry.VelocityY,
		VelocityZ:      telemetry.VelocityZ,
		Heading:        telemetry.Heading,
		Pitch:          telemetry.Pitch,
		Roll:           telemetry.Roll,
		Throttle:       telemetry.BatteryLevel,
		RTKFixType:     telemetry.GPSFixType,
		HDOP:           telemetry.HDOP,
		VDOP:           telemetry.VDOP,
		MarkerDetected: session.MarkerDetected,
		MarkerOffsetX:  session.MarkerOffsetX,
		MarkerOffsetY:  session.MarkerOffsetY,
		Phase:          phase,
	}

	if err := s.landingRepo.AddTrajectoryPoint(point); err != nil {
		return err
	}

	websocket.BroadcastLandingTrajectory(uavID, session.ID, point)

	s.updateSessionStatus(session, telemetry, phase)

	return nil
}

func (s *LandingService) determineLandingPhase(session *models.LandingSession, telemetry *models.FlightStatus) string {
	distance := s.calculateHorizontalDistance(telemetry.Latitude, telemetry.Longitude, session.TargetLatitude, session.TargetLongitude)
	altitude := telemetry.AltitudeRel

	switch {
	case altitude > 10.0 && distance > 5.0:
		return LANDING_PHASE_APPROACH
	case altitude > 1.0 && altitude <= 10.0:
		return LANDING_PHASE_DESCEND
	case altitude <= 1.0 && altitude > 0.1:
		return LANDING_PHASE_PRECISION
	default:
		return LANDING_PHASE_TOUCHDOWN
	}
}

func (s *LandingService) calculateHorizontalDistance(lat1, lon1, lat2, lon2 float64) float64 {
	radLat1 := lat1 * math.Pi / 180.0
	radLon1 := lon1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0
	radLon2 := lon2 * math.Pi / 180.0

	dLat := radLat2 - radLat1
	dLon := radLon2 - radLon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6371000.0 * c
}

func (s *LandingService) updateSessionStatus(session *models.LandingSession, telemetry *models.FlightStatus, phase string) {
	session.HorizontalAccuracy = telemetry.HDOP
	session.VerticalAccuracy = telemetry.VDOP
	session.RTKFixType = telemetry.GPSFixType

	if telemetry.AltitudeRel <= 0.1 && telemetry.GroundSpeed < 0.5 {
		session.Status = models.LandingSessionStatusLanded
		session.LandingError = s.calculateLandingError(session, telemetry)

		if session.PrimaryLandingID > 0 {
			_ = s.landingRepo.UpdateLandingPointStatus(session.PrimaryLandingID, models.LandingPointStatusAvailable)
		}

		s.createLandingCompletionAlert(session)
	} else if phase == LANDING_PHASE_PRECISION && session.Status != models.LandingSessionStatusPrecision {
		session.Status = models.LandingSessionStatusPrecision
	} else if phase == LANDING_PHASE_DESCEND && session.Status != models.LandingSessionStatusDescending {
		session.Status = models.LandingSessionStatusDescending
	}

	_ = s.landingRepo.UpdateLandingSession(session)
	websocket.BroadcastLandingStatus(session.UAVID, string(session.Status), session.PrimaryLandingID, session.RTKEnabled, session.VisionEnabled)
}

func (s *LandingService) calculateLandingError(session *models.LandingSession, telemetry *models.FlightStatus) float64 {
	return s.calculateHorizontalDistance(telemetry.Latitude, telemetry.Longitude, session.TargetLatitude, session.TargetLongitude)
}

func (s *LandingService) createLandingCompletionAlert(session *models.LandingSession) {
	title := "着陆完成"
	var message string
	if session.LandingError <= PRECISION_LANDING_THRESHOLD {
		message = fmt.Sprintf("无人机 %d 已精确着陆，着陆误差: %.2f cm，达到厘米级精度要求。", session.UAVID, session.LandingError*100)
	} else {
		message = fmt.Sprintf("无人机 %d 已着陆，着陆误差: %.2f cm，未达到厘米级精度要求。", session.UAVID, session.LandingError*100)
	}

	level := models.AlertLevelInfo
	if session.LandingError > PRECISION_LANDING_THRESHOLD {
		level = models.AlertLevelWarning
	}

	_, _ = s.alertService.CreateCustomAlert(session.UAVID, title, message, level)

	s.mu.Lock()
	delete(s.activeSessions, session.UAVID)
	delete(s.trajectorySeq, session.UAVID)
	s.mu.Unlock()
}

func (s *LandingService) UpdateVisionData(uavID uint64, visionData *models.VisionLandingData) error {
	visionData.UAVID = uavID
	visionData.Timestamp = time.Now()

	s.mu.RLock()
	session, exists := s.activeSessions[uavID]
	s.mu.RUnlock()

	if exists && session != nil {
		visionData.SessionID = session.ID
	}

	if err := s.landingRepo.AddVisionLandingData(visionData); err != nil {
		return err
	}

	if exists && session != nil && visionData.Confidence >= VISION_DETECTION_THRESHOLD {
		session.MarkerDetected = visionData.MarkerDetected
		session.MarkerType = visionData.MarkerType
		session.MarkerOffsetX = visionData.OffsetX
		session.MarkerOffsetY = visionData.OffsetY
		_ = s.landingRepo.UpdateLandingSession(session)

		if visionData.MarkerDetected {
			s.sendPrecisionCorrection(uavID, visionData)
		}
	}

	websocket.BroadcastLandingVision(uavID, visionData)

	return nil
}

func (s *LandingService) sendPrecisionCorrection(uavID uint64, visionData *models.VisionLandingData) {
	cm := mavlink.NewCommandManager()
	if cm == nil {
		return
	}

	param1 := float32(visionData.OffsetX)
	param2 := float32(visionData.OffsetY)
	param3 := float32(visionData.OffsetZ)
	param4 := float32(visionData.YawError)

	data := mavlink.EncodeCommandLong(uavID, CMD_NAV_PRECISION_LAND, param1, param2, param3, param4, 0, 0, 0)
	_ = cm.SendCommand(uavID, data)
}

func (s *LandingService) UpdateRTKData(uavID uint64, rtkData *models.RTKPositionData) error {
	rtkData.UAVID = uavID
	rtkData.Timestamp = time.Now()

	s.mu.RLock()
	session, exists := s.activeSessions[uavID]
	s.mu.RUnlock()

	if exists && session != nil {
		rtkData.SessionID = session.ID
	}

	if err := s.landingRepo.AddRTKPositionData(rtkData); err != nil {
		return err
	}

	if exists && session != nil {
		session.RTKFixType = rtkData.FixType
		session.HorizontalAccuracy = rtkData.HorizontalAcc
		session.VerticalAccuracy = rtkData.VerticalAcc
		_ = s.landingRepo.UpdateLandingSession(session)
	}

	websocket.BroadcastLandingRTK(uavID, rtkData)

	return nil
}

func (s *LandingService) UpdateMovingPlatformPosition(uavID uint64, lat, lon, velX, velY float64) error {
	s.mu.RLock()
	session, exists := s.activeSessions[uavID]
	s.mu.RUnlock()

	if !exists || session == nil {
		return errors.New("no active landing session")
	}
	if !session.IsMovingPlatform {
		return errors.New("not a moving platform landing")
	}

	session.TargetLatitude = lat
	session.TargetLongitude = lon
	session.MovingPlatformVelocityX = velX
	session.MovingPlatformVelocityY = velY

	if err := s.landingRepo.UpdateLandingSession(session); err != nil {
		return err
	}

	s.sendMovingPlatformUpdate(uavID, lat, lon, velX, velY)

	websocket.BroadcastLandingStatus(uavID, string(session.Status), session.PrimaryLandingID, session.RTKEnabled, session.VisionEnabled)

	return nil
}

func (s *LandingService) sendMovingPlatformUpdate(uavID uint64, lat, lon, velX, velY float64) {
	cm := mavlink.NewCommandManager()
	if cm == nil {
		return
	}

	param1 := float32(lat)
	param2 := float32(lon)
	param3 := float32(velX)
	param4 := float32(velY)

	data := mavlink.EncodeCommandLong(uavID, CMD_NAV_PRECISION_LAND, param1, param2, param3, param4, 1, 0, 0)
	_ = cm.SendCommand(uavID, data)
}

func (s *LandingService) TriggerForcedLanding(uavID uint64, triggerType string, reason string, lockArms bool) (*models.ForcedLandingEvent, error) {
	uav, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	_, err = s.landingRepo.GetActiveForcedLandingEvent(uavID)
	if err == nil {
		return nil, errors.New("active forced landing event already exists")
	}

	var lat, lon, alt float64
	if uav.Latitude != nil {
		lat = *uav.Latitude
	}
	if uav.Longitude != nil {
		lon = *uav.Longitude
	}
	if uav.Altitude != nil {
		alt = *uav.Altitude
	}

	event := &models.ForcedLandingEvent{
		UAVID:         uavID,
		TriggerType:   triggerType,
		Reason:        reason,
		ArmLocked:     lockArms,
		EmergencyMode: "LAND",
		Latitude:      lat,
		Longitude:     lon,
		Altitude:      alt,
		TriggeredAt:   time.Now(),
	}

	s.mu.RLock()
	session, hasSession := s.activeSessions[uavID]
	s.mu.RUnlock()

	if hasSession && session != nil {
		event.SessionID = session.ID
	}

	if err := s.landingRepo.CreateForcedLandingEvent(event); err != nil {
		return nil, err
	}

	s.sendForcedLandingCommand(uavID, lockArms)

	title := "强制降落触发"
	message := fmt.Sprintf("无人机 %d 触发强制降落，触发类型: %s，原因: %s。%s",
		uavID, triggerType, reason,
		map[bool]string{true: "机臂已锁定，将立即切断动力。", false: "机臂未锁定，将执行受控降落。"}[lockArms])
	_, _ = s.alertService.CreateCustomAlert(uavID, title, message, models.AlertLevelCritical)

	websocket.BroadcastForcedLanding(uavID, event)

	middleware.Logger.Warn("Forced landing triggered",
		zap.Uint64("uav_id", uavID),
		zap.String("trigger_type", triggerType),
		zap.Bool("arms_locked", lockArms),
	)

	return event, nil
}

func (s *LandingService) sendForcedLandingCommand(uavID uint64, lockArms bool) {
	cm := mavlink.NewCommandManager()
	if cm == nil {
		return
	}

	if lockArms {
		disarmData := mavlink.EncodeCommandLong(uavID, 400, 0, 21196, 0, 0, 0, 0, 0)
		_ = cm.SendCommand(uavID, disarmData)
	} else {
		landData := mavlink.EncodeCommandLong(uavID, CMD_NAV_LAND, 0, 0, 0, 0, 0, 0, 0)
		_ = cm.SendCommand(uavID, landData)
	}
}

func (s *LandingService) ResolveForcedLanding(id uint64, resolvedBy uint64, notes string) (*models.ForcedLandingEvent, error) {
	event, err := s.landingRepo.FindForcedLandingEventByID(id)
	if err != nil {
		return nil, errors.New("forced landing event not found")
	}
	if event.ResolvedAt != nil {
		return nil, errors.New("event already resolved")
	}

	if err := s.landingRepo.ResolveForcedLandingEvent(id, resolvedBy, notes); err != nil {
		return nil, err
	}

	websocket.BroadcastForcedLandingResolved(event.UAVID, id)

	return s.landingRepo.FindForcedLandingEventByID(id)
}

func (s *LandingService) AbortLanding(uavID uint64, sessionID uint64, reason string) (*models.LandingSession, error) {
	session, err := s.landingRepo.FindLandingSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("landing session not found")
	}
	if session.UAVID != uavID {
		return nil, errors.New("session does not belong to this uav")
	}

	if session.Status == models.LandingSessionStatusLanded ||
		session.Status == models.LandingSessionStatusAborted ||
		session.Status == models.LandingSessionStatusFailed {
		return nil, errors.New("landing session already completed")
	}

	if err := s.landingRepo.UpdateLandingSessionStatus(sessionID, models.LandingSessionStatusAborted); err != nil {
		return nil, err
	}

	if session.PrimaryLandingID > 0 {
		_ = s.landingRepo.UpdateLandingPointStatus(session.PrimaryLandingID, models.LandingPointStatusAvailable)
	}

	cm := mavlink.NewCommandManager()
	if cm != nil {
		rtlData := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
		_ = cm.SendCommand(uavID, rtlData)
	}

	s.mu.Lock()
	delete(s.activeSessions, uavID)
	delete(s.trajectorySeq, uavID)
	s.mu.Unlock()

	title := "降落中止"
	message := fmt.Sprintf("无人机 %d 降落已中止，原因: %s。已触发返航。", uavID, reason)
	_, _ = s.alertService.CreateCustomAlert(uavID, title, message, models.AlertLevelWarning)

	websocket.BroadcastLandingStatus(uavID, string(models.LandingSessionStatusAborted), 0, false, false)

	return s.landingRepo.FindLandingSessionByID(sessionID)
}

func (s *LandingService) GetActiveLandingSession(uavID uint64) (*models.LandingSession, error) {
	return s.landingRepo.GetActiveLandingSession(uavID)
}

func (s *LandingService) GetLandingSession(id uint64) (*models.LandingSession, error) {
	return s.landingRepo.FindLandingSessionByID(id)
}

func (s *LandingService) ListLandingSessions(page, pageSize int, uavID uint64, status string, startTime, endTime string) ([]models.LandingSession, int64, error) {
	pagination := &utils.Pagination{Page: page, PageSize: pageSize}
	return s.landingRepo.ListLandingSessions(pagination, uavID, models.LandingSessionStatus(status), startTime, endTime)
}

func (s *LandingService) GetLandingTrajectory(sessionID uint64) ([]models.LandingTrajectoryPoint, error) {
	return s.landingRepo.GetTrajectoryBySession(sessionID)
}

func (s *LandingService) GetLandingStatistics(uavID uint64, startTime, endTime string) (map[string]interface{}, error) {
	return s.landingRepo.GetLandingStatistics(uavID, startTime, endTime)
}

func (s *LandingService) ListForcedLandingEvents(page, pageSize int, uavID uint64, triggerType string, isResolved *bool) ([]models.ForcedLandingEvent, int64, error) {
	pagination := &utils.Pagination{Page: page, PageSize: pageSize}
	return s.landingRepo.ListForcedLandingEvents(pagination, uavID, triggerType, isResolved)
}

func (s *LandingService) GetForcedLandingEvent(id uint64) (*models.ForcedLandingEvent, error) {
	return s.landingRepo.FindForcedLandingEventByID(id)
}

func (s *LandingService) GetActiveForcedLandingEvent(uavID uint64) (*models.ForcedLandingEvent, error) {
	return s.landingRepo.GetActiveForcedLandingEvent(uavID)
}

func (s *LandingService) SwitchToAlternateLanding(uavID uint64, sessionID uint64, alternateID uint64) (*models.LandingSession, error) {
	session, err := s.landingRepo.FindLandingSessionByID(sessionID)
	if err != nil {
		return nil, errors.New("landing session not found")
	}
	if session.UAVID != uavID {
		return nil, errors.New("session does not belong to this uav")
	}

	alternate, err := s.landingRepo.FindLandingPointByID(alternateID)
	if err != nil {
		return nil, errors.New("alternate landing point not found")
	}

	if alternate.Type != models.LandingPointTypeAlternate && alternate.Type != models.LandingPointTypeEmergency {
		return nil, errors.New("landing point is not an alternate or emergency type")
	}

	if session.PrimaryLandingID > 0 {
		_ = s.landingRepo.UpdateLandingPointStatus(session.PrimaryLandingID, models.LandingPointStatusAvailable)
	}

	session.PrimaryLandingID = alternateID
	session.TargetLatitude = alternate.Latitude
	session.TargetLongitude = alternate.Longitude
	session.TargetAltitude = alternate.Altitude
	session.AlternateLandingID = 0

	if err := s.landingRepo.UpdateLandingSession(session); err != nil {
		return nil, err
	}

	_ = s.landingRepo.UpdateLandingPointStatus(alternateID, models.LandingPointStatusOccupied)

	s.sendLandingUpdate(uavID, alternate)

	websocket.BroadcastLandingStatus(uavID, string(session.Status), alternateID, session.RTKEnabled, session.VisionEnabled)

	title := "切换备降点"
	message := fmt.Sprintf("无人机 %d 已切换至备降点: %s (ID: %d)", uavID, alternate.Name, alternateID)
	_, _ = s.alertService.CreateCustomAlert(uavID, title, message, models.AlertLevelWarning)

	middleware.Logger.Info("Switched to alternate landing point",
		zap.Uint64("uav_id", uavID),
		zap.Uint64("session_id", sessionID),
		zap.Uint64("alternate_id", alternateID),
	)

	return s.landingRepo.FindLandingSessionByID(sessionID)
}

func (s *LandingService) sendLandingUpdate(uavID uint64, point *models.LandingPoint) {
	cm := mavlink.NewCommandManager()
	if cm == nil {
		return
	}

	param1 := float32(point.Latitude)
	param2 := float32(point.Longitude)
	param3 := float32(point.Altitude)

	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_MISSION_ITEM, param1, param2, param3, 0, 0, 0, 0)
	_ = cm.SendCommand(uavID, data)
}
