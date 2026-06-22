package service

import (
	"fmt"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	minSafeDistance     = 50.0
	warningDistance     = 100.0
	criticalDistance    = 30.0
	emergencyDistance   = 15.0
	defaultSpeedFactor  = 1.0
	slowSpeedFactor     = 0.5
	holdSpeedFactor     = 0.0
	positionStaleSecs   = 30
	timeDiffSafeSec     = 30.0
	holdDurationDefault = 10 * time.Second
)

type CollisionAvoidanceService struct {
	repo        *repository.CollisionRepository
	missionRepo *repository.MissionRepository
	uavRepo     *repository.UAVRepository

	positions      map[uint64]*models.UAVLivePosition
	positionsMu    sync.RWMutex
	alerts         map[string]*models.CollisionAlert
	alertsMu       sync.Mutex
	speedFactors   map[uint64]float64
	speedFactorsMu sync.RWMutex

	routeIntersections      map[string]*models.RouteIntersection
	routeIntersectionsMu    sync.Mutex
	intersectionAvoidActive map[string]bool
	intersectionAvoidMu     sync.Mutex

	enabled   bool
	enabledMu sync.RWMutex

	stopCh chan struct{}
}

var (
	collisionServiceInstance *CollisionAvoidanceService
	collisionServiceOnce     sync.Once
)

func NewCollisionAvoidanceService() *CollisionAvoidanceService {
	collisionServiceOnce.Do(func() {
		collisionServiceInstance = &CollisionAvoidanceService{
			repo:                    repository.NewCollisionRepository(),
			missionRepo:             repository.NewMissionRepository(),
			uavRepo:                 repository.NewUAVRepository(),
			positions:               make(map[uint64]*models.UAVLivePosition),
			alerts:                  make(map[string]*models.CollisionAlert),
			speedFactors:            make(map[uint64]float64),
			routeIntersections:      make(map[string]*models.RouteIntersection),
			intersectionAvoidActive: make(map[string]bool),
			enabled:                 true,
			stopCh:                  make(chan struct{}),
		}
		collisionServiceInstance.registerCommandCallbacks()
		go collisionServiceInstance.backgroundLoop()
	})
	return collisionServiceInstance
}

func (s *CollisionAvoidanceService) registerCommandCallbacks() {
	cmdMgr := mavlink.NewCommandManager()
	cmdMgr.RegisterAckCallback(50030, func(uavID uint64, cmdID uint16, result uint8, success bool) {
		s.handleAvoidCommandAck(uavID, success, result)
	})
}

func (s *CollisionAvoidanceService) ReportPosition(pos *models.UAVLivePosition) {
	if pos == nil || pos.UAVID == 0 {
		return
	}

	s.positionsMu.Lock()
	s.positions[pos.UAVID] = pos
	s.positionsMu.Unlock()

	if s.isEnabled() {
		s.checkAllCollisions()
		s.verifySpeedAdjustment(pos)
	}
}

func (s *CollisionAvoidanceService) isEnabled() bool {
	s.enabledMu.RLock()
	defer s.enabledMu.RUnlock()
	return s.enabled
}

func (s *CollisionAvoidanceService) SetEnabled(enabled bool) {
	s.enabledMu.Lock()
	s.enabled = enabled
	s.enabledMu.Unlock()

	middleware.Logger.Info("协同避让状态变更", zap.Bool("enabled", enabled))
}

func (s *CollisionAvoidanceService) IsEnabled() bool {
	return s.isEnabled()
}

func (s *CollisionAvoidanceService) checkAllCollisions() {
	s.positionsMu.RLock()
	positions := make([]*models.UAVLivePosition, 0, len(s.positions))
	now := time.Now()
	for _, pos := range s.positions {
		if now.Sub(pos.Timestamp) < positionStaleSecs*time.Second {
			positions = append(positions, pos)
		}
	}
	s.positionsMu.RUnlock()

	if len(positions) < 2 {
		return
	}

	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			s.checkPairCollision(positions[i], positions[j])
		}
	}
}

func (s *CollisionAvoidanceService) checkPairCollision(pos1, pos2 *models.UAVLivePosition) {
	distance := calculate3DDistance(
		pos1.Latitude, pos1.Longitude, pos1.Altitude,
		pos2.Latitude, pos2.Longitude, pos2.Altitude,
	)

	pairKey := getPairKey(pos1.UAVID, pos2.UAVID)

	var riskLevel models.CollisionRiskLevel
	switch {
	case distance <= emergencyDistance:
		riskLevel = models.CollisionRiskAvoiding
	case distance <= criticalDistance:
		riskLevel = models.CollisionRiskCritical
	case distance <= minSafeDistance:
		riskLevel = models.CollisionRiskWarning
	default:
		riskLevel = models.CollisionRiskSafe
	}

	ttc := calculateTTC(pos1, pos2, distance)

	if riskLevel == models.CollisionRiskSafe {
		s.resolveAlertIfExists(pairKey)
		return
	}

	decision := s.makeAvoidanceDecision(pos1, pos2, distance, riskLevel)

	s.createOrUpdateAlert(pairKey, pos1, pos2, distance, ttc, riskLevel, decision)

	s.executeAvoidance(decision)
}

func (s *CollisionAvoidanceService) makeAvoidanceDecision(
	pos1, pos2 *models.UAVLivePosition,
	distance float64,
	riskLevel models.CollisionRiskLevel,
) *models.AvoidanceDecision {
	pairKey := getPairKey(pos1.UAVID, pos2.UAVID)
	decision := &models.AvoidanceDecision{
		PairKey:      pairKey,
		UAVID1:       pos1.UAVID,
		UAVID2:       pos2.UAVID,
		Distance:     distance,
		RiskLevel:    riskLevel,
		SpeedFactor1: defaultSpeedFactor,
		SpeedFactor2: defaultSpeedFactor,
	}

	switch riskLevel {
	case models.CollisionRiskWarning:
		decision.PrimaryAction = models.AvoidanceSpeedReduce
		decision.Reason = fmt.Sprintf("距离 %.1fm 接近安全阈值 %.1fm", distance, minSafeDistance)
		if pos1.GroundSpeed >= pos2.GroundSpeed {
			decision.SpeedFactor1 = slowSpeedFactor
			decision.SecondaryAction = models.AvoidanceSpeedAdjust
		} else {
			decision.SpeedFactor2 = slowSpeedFactor
			decision.SecondaryAction = models.AvoidanceSpeedAdjust
		}

	case models.CollisionRiskCritical:
		decision.PrimaryAction = models.AvoidanceHoldPosition
		decision.Reason = fmt.Sprintf("距离 %.1fm 低于安全阈值 %.1fm", distance, minSafeDistance)
		decision.HoldDuration = holdDurationDefault
		if pos1.GroundSpeed >= pos2.GroundSpeed {
			decision.SpeedFactor1 = holdSpeedFactor
			decision.SecondaryAction = models.AvoidanceSpeedReduce
			decision.SpeedFactor2 = slowSpeedFactor
		} else {
			decision.SpeedFactor2 = holdSpeedFactor
			decision.SecondaryAction = models.AvoidanceSpeedReduce
			decision.SpeedFactor1 = slowSpeedFactor
		}

	case models.CollisionRiskAvoiding:
		decision.PrimaryAction = models.AvoidanceHoldPosition
		decision.Reason = fmt.Sprintf("距离 %.1fm 紧急，双机悬停", distance)
		decision.SpeedFactor1 = holdSpeedFactor
		decision.SpeedFactor2 = holdSpeedFactor
		decision.HoldDuration = 15 * time.Second
	}

	return decision
}

func (s *CollisionAvoidanceService) executeAvoidance(decision *models.AvoidanceDecision) {
	cmdMgr := mavlink.NewCommandManager()

	s.speedFactorsMu.Lock()
	prevFactor1 := s.speedFactors[decision.UAVID1]
	prevFactor2 := s.speedFactors[decision.UAVID2]
	s.speedFactorsMu.Unlock()

	if decision.SpeedFactor1 != prevFactor1 {
		s.sendSpeedCommand(decision.UAVID1, decision.SpeedFactor1, cmdMgr)
		s.speedFactorsMu.Lock()
		s.speedFactors[decision.UAVID1] = decision.SpeedFactor1
		s.speedFactorsMu.Unlock()
	}

	if decision.SpeedFactor2 != prevFactor2 {
		s.sendSpeedCommand(decision.UAVID2, decision.SpeedFactor2, cmdMgr)
		s.speedFactorsMu.Lock()
		s.speedFactors[decision.UAVID2] = decision.SpeedFactor2
		s.speedFactorsMu.Unlock()
	}

	s.broadcastAvoidanceDecision(decision)
}

func (s *CollisionAvoidanceService) sendSpeedCommand(uavID uint64, factor float64, cmdMgr *mavlink.CommandManager) {
	action := "resume_normal_speed"
	if factor <= 0.01 {
		action = "hold_position"
	} else if factor < 0.6 {
		action = "reduce_speed"
	}

	baseSpeed := 0.0
	s.positionsMu.RLock()
	if pos, exists := s.positions[uavID]; exists {
		baseSpeed = pos.GroundSpeed
	}
	s.positionsMu.RUnlock()

	err := cmdMgr.SendCustomCommand(uavID, "collision_avoid_speed", map[string]interface{}{
		"param1": float32(factor),
		"param2": float32(0),
		"param3": float32(0),
	})

	if err == nil {
		s.trackAvoidCommand(uavID, factor, baseSpeed)
	}

	middleware.Logger.Info("碰撞避让: 调整速度",
		zap.Uint64("uav_id", uavID),
		zap.Float64("speed_factor", factor),
		zap.String("action", action),
		zap.Error(err),
	)
}

func (s *CollisionAvoidanceService) createOrUpdateAlert(
	pairKey string,
	pos1, pos2 *models.UAVLivePosition,
	distance float64,
	ttc float64,
	riskLevel models.CollisionRiskLevel,
	decision *models.AvoidanceDecision,
) {
	s.alertsMu.Lock()
	existing, exists := s.alerts[pairKey]
	s.alertsMu.Unlock()

	alertID := fmt.Sprintf("col_%s", uuid.New().String()[:8])

	if exists && existing != nil {
		existing.RiskLevel = riskLevel
		existing.CurrentDistance = distance
		existing.TimeToCollision = ttc
		existing.Action = decision.PrimaryAction
		existing.ActionDetail = decision.Reason
		existing.UpdatedAt = time.Now()
		_ = s.repo.UpdateAlert(existing)
	} else {
		alert := &models.CollisionAlert{
			AlertID:        alertID,
			UAVID1:         pos1.UAVID,
			UAVID2:         pos2.UAVID,
			RiskLevel:      riskLevel,
			MinDistance:    distance,
			CurrentDistance: distance,
			TimeToCollision: ttc,
			AlertType:      "proximity",
			Action:         decision.PrimaryAction,
			ActionDetail:   decision.Reason,
		}
		_ = s.repo.CreateAlert(alert)

		s.alertsMu.Lock()
		s.alerts[pairKey] = alert
		s.alertsMu.Unlock()

		s.broadcastNewAlert(alert, pos1, pos2)
	}
}

func (s *CollisionAvoidanceService) resolveAlertIfExists(pairKey string) {
	s.alertsMu.Lock()
	alert, exists := s.alerts[pairKey]
	if exists {
		delete(s.alerts, pairKey)
	}
	s.alertsMu.Unlock()

	if exists && alert != nil && !alert.IsResolved {
		_ = s.repo.ResolveAlert(alert.ID)

		s.speedFactorsMu.RLock()
		_, hasFactor1 := s.speedFactors[alert.UAVID1]
		_, hasFactor2 := s.speedFactors[alert.UAVID2]
		s.speedFactorsMu.RUnlock()

		cmdMgr := mavlink.NewCommandManager()
		if hasFactor1 {
			baseSpeed1 := s.getUAVGroundSpeed(alert.UAVID1)
			err1 := cmdMgr.SendCustomCommand(alert.UAVID1, "collision_avoid_speed", map[string]interface{}{
				"param1": float32(defaultSpeedFactor),
			})
			if err1 == nil {
				s.trackAvoidCommand(alert.UAVID1, defaultSpeedFactor, baseSpeed1)
			}
			s.speedFactorsMu.Lock()
			delete(s.speedFactors, alert.UAVID1)
			s.speedFactorsMu.Unlock()
		}
		if hasFactor2 {
			baseSpeed2 := s.getUAVGroundSpeed(alert.UAVID2)
			err2 := cmdMgr.SendCustomCommand(alert.UAVID2, "collision_avoid_speed", map[string]interface{}{
				"param1": float32(defaultSpeedFactor),
			})
			if err2 == nil {
				s.trackAvoidCommand(alert.UAVID2, defaultSpeedFactor, baseSpeed2)
			}
			s.speedFactorsMu.Lock()
			delete(s.speedFactors, alert.UAVID2)
			s.speedFactorsMu.Unlock()
		}

		websocket.NewHub().Broadcast(map[string]interface{}{
			"type":      "collision_resolved",
			"alert_id":  alert.AlertID,
			"uav_id_1":  alert.UAVID1,
			"uav_id_2":  alert.UAVID2,
			"timestamp": time.Now().UnixNano() / 1e6,
		})
	}
}

func (s *CollisionAvoidanceService) broadcastNewAlert(
	alert *models.CollisionAlert,
	pos1, pos2 *models.UAVLivePosition,
) {
	data := map[string]interface{}{
		"alert_id":         alert.AlertID,
		"uav_id_1":         alert.UAVID1,
		"uav_id_2":         alert.UAVID2,
		"risk_level":       string(alert.RiskLevel),
		"current_distance": alert.CurrentDistance,
		"min_distance":     alert.MinDistance,
		"time_to_collision": alert.TimeToCollision,
		"action_taken":     string(alert.Action),
		"action_detail":    alert.ActionDetail,
		"uav1_pos": map[string]float64{
			"lat": pos1.Latitude,
			"lon": pos1.Longitude,
			"alt": pos1.Altitude,
		},
		"uav2_pos": map[string]float64{
			"lat": pos2.Latitude,
			"lon": pos2.Longitude,
			"alt": pos2.Altitude,
		},
		"timestamp": time.Now().UnixNano() / 1e6,
	}
	websocket.NewHub().BroadcastCollisionAlert(alert.UAVID1, data)
}

func (s *CollisionAvoidanceService) broadcastAvoidanceDecision(decision *models.AvoidanceDecision) {
	websocket.NewHub().Broadcast(map[string]interface{}{
		"type":            "avoidance_decision",
		"pair_key":        decision.PairKey,
		"uav_id_1":        decision.UAVID1,
		"uav_id_2":        decision.UAVID2,
		"distance":        decision.Distance,
		"risk_level":      string(decision.RiskLevel),
		"primary_action":  string(decision.PrimaryAction),
		"secondary_action": string(decision.SecondaryAction),
		"speed_factor_1":  decision.SpeedFactor1,
		"speed_factor_2":  decision.SpeedFactor2,
		"reason":          decision.Reason,
		"timestamp":       time.Now().UnixNano() / 1e6,
	})
}

func (s *CollisionAvoidanceService) DetectRouteIntersections() ([]models.RouteIntersection, error) {
	positions := s.getActivePositions()
	if len(positions) < 2 {
		return nil, nil
	}

	var allWaypoints map[uint64][]models.MissionWaypoint
	allWaypoints = make(map[uint64][]models.MissionWaypoint)

	for _, pos := range positions {
		mission, err := s.missionRepo.GetActiveMissionByUAV(pos.UAVID)
		if err != nil || mission == nil {
			continue
		}
		waypoints, err := s.missionRepo.GetMissionWaypoints(mission.ID)
		if err != nil {
			continue
		}
		allWaypoints[pos.UAVID] = waypoints
	}

	if len(allWaypoints) < 2 {
		return nil, nil
	}

	_ = s.repo.ClearActiveIntersections()

	var intersections []models.RouteIntersection
	uavIDs := make([]uint64, 0, len(allWaypoints))
	for id := range allWaypoints {
		uavIDs = append(uavIDs, id)
	}

	for i := 0; i < len(uavIDs); i++ {
		for j := i + 1; j < len(uavIDs); j++ {
			wps1 := allWaypoints[uavIDs[i]]
			wps2 := allWaypoints[uavIDs[j]]

			for _, wp1 := range wps1 {
				for _, wp2 := range wps2 {
					dist := calculate2DDistance(wp1.Latitude, wp1.Longitude, wp2.Latitude, wp2.Longitude)
					if dist > warningDistance {
						continue
					}

					altDiff := math.Abs(wp1.Altitude - wp2.Altitude)
					if altDiff > 30 {
						continue
					}

					pos1 := positions[uavIDs[i]]
					pos2 := positions[uavIDs[j]]

					eta1 := s.estimateETA(pos1, wp1)
					eta2 := s.estimateETA(pos2, wp2)
					timeDiff := math.Abs(eta1.Sub(eta2).Seconds())

					riskLevel := models.CollisionRiskSafe
					switch {
					case timeDiff < 10 && dist < minSafeDistance:
						riskLevel = models.CollisionRiskCritical
					case timeDiff < timeDiffSafeSec && dist < minSafeDistance:
						riskLevel = models.CollisionRiskWarning
					case dist < warningDistance:
						riskLevel = models.CollisionRiskWarning
					}

					intersection := models.RouteIntersection{
						UAVID1:       uavIDs[i],
						UAVID2:       uavIDs[j],
						WaypointSeq1: wp1.Seq,
						WaypointSeq2: wp2.Seq,
						Latitude:     (wp1.Latitude + wp2.Latitude) / 2,
						Longitude:    (wp1.Longitude + wp2.Longitude) / 2,
						Altitude:     (wp1.Altitude + wp2.Altitude) / 2,
						Distance:     dist,
						ETA1:         eta1,
						ETA2:         eta2,
						TimeDiffSec:  timeDiff,
						RiskLevel:    riskLevel,
						IsActive:     true,
					}

					_ = s.repo.CreateIntersection(&intersection)
					intersections = append(intersections, intersection)
				}
			}
		}
	}

	s.broadcastIntersections(intersections)
	return intersections, nil
}

func (s *CollisionAvoidanceService) getActivePositions() map[uint64]*models.UAVLivePosition {
	s.positionsMu.RLock()
	defer s.positionsMu.RUnlock()

	result := make(map[uint64]*models.UAVLivePosition)
	now := time.Now()
	for id, pos := range s.positions {
		if now.Sub(pos.Timestamp) < positionStaleSecs*time.Second {
			result[id] = pos
		}
	}
	return result
}

func (s *CollisionAvoidanceService) estimateETA(pos *models.UAVLivePosition, wp models.MissionWaypoint) time.Time {
	if pos == nil {
		return time.Now().Add(1 * time.Hour)
	}

	distance := calculate2DDistance(pos.Latitude, pos.Longitude, wp.Latitude, wp.Longitude)
	speed := pos.GroundSpeed
	if speed < 1 {
		speed = 5
	}

	seconds := distance / speed
	return time.Now().Add(time.Duration(seconds) * time.Second)
}

func (s *CollisionAvoidanceService) GetActiveAlerts() ([]models.CollisionAlert, error) {
	return s.repo.GetActiveAlerts()
}

func (s *CollisionAvoidanceService) GetActiveIntersections(uavID uint64) ([]models.RouteIntersection, error) {
	return s.repo.GetActiveIntersections(uavID)
}

func (s *CollisionAvoidanceService) GetAllPositions() map[uint64]*models.UAVLivePosition {
	return s.getActivePositions()
}

func (s *CollisionAvoidanceService) GetSpeedFactor(uavID uint64) float64 {
	s.speedFactorsMu.RLock()
	defer s.speedFactorsMu.RUnlock()
	if factor, ok := s.speedFactors[uavID]; ok {
		return factor
	}
	return defaultSpeedFactor
}

func (s *CollisionAvoidanceService) ResolveAlert(id uint64) error {
	alert, err := s.repo.GetAlertByID(id)
	if err != nil {
		return err
	}

	pairKey := getPairKey(alert.UAVID1, alert.UAVID2)
	s.resolveAlertIfExists(pairKey)
	return nil
}

func (s *CollisionAvoidanceService) ManualAvoidance(uavID uint64, action string, param float64) error {
	cmdMgr := mavlink.NewCommandManager()
	return cmdMgr.SendCustomCommand(uavID, action, map[string]interface{}{
		"param1": float32(param),
	})
}

func (s *CollisionAvoidanceService) broadcastIntersections(intersections []models.RouteIntersection) {
	if len(intersections) == 0 {
		return
	}

	data := map[string]interface{}{
		"type":          "route_intersections",
		"intersections": intersections,
		"count":         len(intersections),
		"timestamp":     time.Now().UnixNano() / 1e6,
	}
	websocket.NewHub().Broadcast(data)
}

func calculate3DDistance(lat1, lon1, alt1, lat2, lon2, alt2 float64) float64 {
	groundDist := calculate2DDistance(lat1, lon1, lat2, lon2)
	altDiff := alt2 - alt1
	return math.Sqrt(groundDist*groundDist + altDiff*altDiff)
}

func calculate2DDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func calculateTTC(pos1, pos2 *models.UAVLivePosition, distance float64) float64 {
	if pos1.VelocityX == 0 && pos1.VelocityY == 0 && pos2.VelocityX == 0 && pos2.VelocityY == 0 {
		return -1
	}

	relVx := pos2.VelocityX - pos1.VelocityX
	relVy := pos2.VelocityY - pos1.VelocityY
	relSpeed := math.Sqrt(relVx*relVx + relVy*relVy)

	if relSpeed < 0.1 {
		return -1
	}

	if distance <= 0 {
		return 0
	}

	ttc := distance / relSpeed
	if ttc < 0 {
		return -1
	}

	return ttc
}

func getPairKey(id1, id2 uint64) string {
	if id1 < id2 {
		return fmt.Sprintf("%d_%d", id1, id2)
	}
	return fmt.Sprintf("%d_%d", id2, id1)
}

func (s *CollisionAvoidanceService) backgroundLoop() {
	routeTicker := time.NewTicker(5 * time.Second)
	cleanupTicker := time.NewTicker(30 * time.Second)
	defer routeTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-routeTicker.C:
			if s.isEnabled() {
				s.checkRouteIntersectionsAndAvoid()
			}
		case <-cleanupTicker.C:
			s.cleanupStaleData()
		}
	}
}

func (s *CollisionAvoidanceService) Stop() {
	close(s.stopCh)
}

func (s *CollisionAvoidanceService) checkRouteIntersectionsAndAvoid() {
	intersections, err := s.DetectRouteIntersections()
	if err != nil || len(intersections) == 0 {
		return
	}

	for i := range intersections {
		isec := &intersections[i]
		if isec.RiskLevel == models.CollisionRiskSafe {
			continue
		}

		pairKey := getPairKey(isec.UAVID1, isec.UAVID2)

		s.intersectionAvoidMu.Lock()
		alreadyActive := s.intersectionAvoidActive[pairKey]
		s.intersectionAvoidMu.Unlock()

		if alreadyActive {
			continue
		}

		if isec.RiskLevel == models.CollisionRiskWarning || isec.RiskLevel == models.CollisionRiskCritical {
			s.executeIntersectionAvoidance(isec)
		}
	}
}

func (s *CollisionAvoidanceService) executeIntersectionAvoidance(isec *models.RouteIntersection) {
	pairKey := getPairKey(isec.UAVID1, isec.UAVID2)

	s.intersectionAvoidMu.Lock()
	if s.intersectionAvoidActive[pairKey] {
		s.intersectionAvoidMu.Unlock()
		return
	}
	s.intersectionAvoidActive[pairKey] = true
	s.intersectionAvoidMu.Unlock()

	cmdMgr := mavlink.NewCommandManager()

	if isec.RiskLevel == models.CollisionRiskCritical {
		s.intersectionHoldAndPass(isec, cmdMgr)
	} else {
		s.intersectionSpeedAdjust(isec, cmdMgr)
	}

	s.broadcastIntersectionAvoidance(isec)

	go s.monitorIntersectionResolution(isec)
}

func (s *CollisionAvoidanceService) intersectionHoldAndPass(isec *models.RouteIntersection, cmdMgr *mavlink.CommandManager) {
	var holdUAV, passUAV uint64
	if isec.TimeDiffSec >= 0 {
		holdUAV = isec.UAVID1
		passUAV = isec.UAVID2
	} else {
		holdUAV = isec.UAVID2
		passUAV = isec.UAVID1
	}

	holdBaseSpeed := s.getUAVGroundSpeed(holdUAV)
	passBaseSpeed := s.getUAVGroundSpeed(passUAV)

	errHold := cmdMgr.SendCustomCommand(holdUAV, "collision_avoid_speed", map[string]interface{}{
		"param1": float32(holdSpeedFactor),
		"param2": float32(isec.WaypointSeq1),
	})
	if errHold == nil {
		s.trackAvoidCommand(holdUAV, holdSpeedFactor, holdBaseSpeed)
	}
	s.speedFactorsMu.Lock()
	s.speedFactors[holdUAV] = holdSpeedFactor
	s.speedFactorsMu.Unlock()

	errPass := cmdMgr.SendCustomCommand(passUAV, "collision_avoid_speed", map[string]interface{}{
		"param1": float32(0.7),
		"param2": float32(0),
	})
	if errPass == nil {
		s.trackAvoidCommand(passUAV, 0.7, passBaseSpeed)
	}
	s.speedFactorsMu.Lock()
	s.speedFactors[passUAV] = 0.7
	s.speedFactorsMu.Unlock()

	middleware.Logger.Info("航路交叉避让: 一停一慢",
		zap.Uint64("hold_uav", holdUAV),
		zap.Uint64("pass_uav", passUAV),
		zap.Float64("distance", isec.Distance),
		zap.Float64("time_diff_sec", isec.TimeDiffSec),
	)
}

func (s *CollisionAvoidanceService) getUAVGroundSpeed(uavID uint64) float64 {
	s.positionsMu.RLock()
	defer s.positionsMu.RUnlock()
	if pos, exists := s.positions[uavID]; exists {
		return pos.GroundSpeed
	}
	return 0
}

func (s *CollisionAvoidanceService) intersectionSpeedAdjust(isec *models.RouteIntersection, cmdMgr *mavlink.CommandManager) {
	var fastUAV, slowUAV uint64
	if isec.TimeDiffSec > 0 {
		fastUAV = isec.UAVID1
		slowUAV = isec.UAVID2
	} else {
		fastUAV = isec.UAVID2
		slowUAV = isec.UAVID1
	}

	baseSpeed := s.getUAVGroundSpeed(fastUAV)
	err := cmdMgr.SendCustomCommand(fastUAV, "collision_avoid_speed", map[string]interface{}{
		"param1": float32(slowSpeedFactor),
	})
	if err == nil {
		s.trackAvoidCommand(fastUAV, slowSpeedFactor, baseSpeed)
	}
	s.speedFactorsMu.Lock()
	s.speedFactors[fastUAV] = slowSpeedFactor
	s.speedFactorsMu.Unlock()

	middleware.Logger.Info("航路交叉避让: 减速错峰",
		zap.Uint64("slow_uav", fastUAV),
		zap.Uint64("other_uav", slowUAV),
		zap.Float64("distance", isec.Distance),
		zap.Float64("time_diff_sec", isec.TimeDiffSec),
	)
}

func (s *CollisionAvoidanceService) monitorIntersectionResolution(isec *models.RouteIntersection) {
	pairKey := getPairKey(isec.UAVID1, isec.UAVID2)
	checkTicker := time.NewTicker(2 * time.Second)
	defer checkTicker.Stop()

	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-timeout:
			s.resolveIntersectionAvoidance(pairKey, isec)
			return
		case <-checkTicker.C:
			pos := s.getActivePositions()
			pos1, ok1 := pos[isec.UAVID1]
			pos2, ok2 := pos[isec.UAVID2]
			if !ok1 || !ok2 {
				continue
			}

			dist := calculate3DDistance(
				pos1.Latitude, pos1.Longitude, pos1.Altitude,
				pos2.Latitude, pos2.Longitude, pos2.Altitude,
			)

			if dist > minSafeDistance*1.5 {
				s.resolveIntersectionAvoidance(pairKey, isec)
				return
			}
		}
	}
}

func (s *CollisionAvoidanceService) resolveIntersectionAvoidance(pairKey string, isec *models.RouteIntersection) {
	s.intersectionAvoidMu.Lock()
	delete(s.intersectionAvoidActive, pairKey)
	s.intersectionAvoidMu.Unlock()

	cmdMgr := mavlink.NewCommandManager()

	s.speedFactorsMu.Lock()
	_, hasFactor1 := s.speedFactors[isec.UAVID1]
	_, hasFactor2 := s.speedFactors[isec.UAVID2]
	s.speedFactorsMu.Unlock()

	if hasFactor1 {
		baseSpeed1 := s.getUAVGroundSpeed(isec.UAVID1)
		err1 := cmdMgr.SendCustomCommand(isec.UAVID1, "collision_avoid_speed", map[string]interface{}{
			"param1": float32(defaultSpeedFactor),
		})
		if err1 == nil {
			s.trackAvoidCommand(isec.UAVID1, defaultSpeedFactor, baseSpeed1)
		}
		s.speedFactorsMu.Lock()
		delete(s.speedFactors, isec.UAVID1)
		s.speedFactorsMu.Unlock()
	}
	if hasFactor2 {
		baseSpeed2 := s.getUAVGroundSpeed(isec.UAVID2)
		err2 := cmdMgr.SendCustomCommand(isec.UAVID2, "collision_avoid_speed", map[string]interface{}{
			"param1": float32(defaultSpeedFactor),
		})
		if err2 == nil {
			s.trackAvoidCommand(isec.UAVID2, defaultSpeedFactor, baseSpeed2)
		}
		s.speedFactorsMu.Lock()
		delete(s.speedFactors, isec.UAVID2)
		s.speedFactorsMu.Unlock()
	}

	middleware.Logger.Info("航路交叉避让解除",
		zap.Uint64("uav_id_1", isec.UAVID1),
		zap.Uint64("uav_id_2", isec.UAVID2),
	)
}

func (s *CollisionAvoidanceService) broadcastIntersectionAvoidance(isec *models.RouteIntersection) {
	data := map[string]interface{}{
		"type":            "route_intersection_avoidance",
		"uav_id_1":        isec.UAVID1,
		"uav_id_2":        isec.UAVID2,
		"waypoint_seq_1":  isec.WaypointSeq1,
		"waypoint_seq_2":  isec.WaypointSeq2,
		"latitude":        isec.Latitude,
		"longitude":       isec.Longitude,
		"altitude":        isec.Altitude,
		"distance":        isec.Distance,
		"time_diff_sec":   isec.TimeDiffSec,
		"risk_level":      string(isec.RiskLevel),
		"action":          "speed_adjust",
		"timestamp":       time.Now().UnixNano() / 1e6,
	}
	websocket.NewHub().Broadcast(data)
}

func (s *CollisionAvoidanceService) cleanupStaleData() {
	s.positionsMu.Lock()
	now := time.Now()
	for id, pos := range s.positions {
		if now.Sub(pos.Timestamp) > positionStaleSecs*time.Second {
			delete(s.positions, id)
		}
	}
	s.positionsMu.Unlock()

	s.alertsMu.Lock()
	for key, alert := range s.alerts {
		if alert.IsResolved && now.Sub(alert.UpdatedAt) > 5*time.Minute {
			delete(s.alerts, key)
		}
	}
	s.alertsMu.Unlock()
}

func (s *CollisionAvoidanceService) GetIntersectionAvoidanceActive(pairKey string) bool {
	s.intersectionAvoidMu.RLock()
	defer s.intersectionAvoidMu.RUnlock()
	return s.intersectionAvoidActive[pairKey]
}

type avoidanceCmdTracker struct {
	uavID         uint64
	targetFactor  float64
	sentAt        time.Time
	acked         bool
	ackResult     uint8
	verified      bool
	retryCount    int
	baseSpeed     float64
}

var (
	cmdTrackers   = make(map[uint64]*avoidanceCmdTracker)
	cmdTrackersMu sync.Mutex
)

func (s *CollisionAvoidanceService) handleAvoidCommandAck(uavID uint64, success bool, result uint8) {
	cmdTrackersMu.Lock()
	tracker, exists := cmdTrackers[uavID]
	cmdTrackersMu.Unlock()

	if !exists {
		return
	}

	tracker.acked = true
	tracker.ackResult = result

	if success {
		middleware.Logger.Info("避让指令已确认",
			zap.Uint64("uav_id", uavID),
			zap.Float64("target_factor", tracker.targetFactor),
		)
	} else {
		middleware.Logger.Warn("避让指令执行失败",
			zap.Uint64("uav_id", uavID),
			zap.Uint8("result", result),
			zap.Int("retry_count", tracker.retryCount),
		)

		if tracker.retryCount < 2 {
			tracker.retryCount++
			go s.retryAvoidCommand(tracker)
		}
	}

	s.broadcastCommandStatus(uavID, tracker)
}

func (s *CollisionAvoidanceService) retryAvoidCommand(tracker *avoidanceCmdTracker) {
	time.Sleep(500 * time.Millisecond)

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(tracker.uavID, "collision_avoid_speed", map[string]interface{}{
		"param1": float32(tracker.targetFactor),
	})

	tracker.sentAt = time.Now()

	middleware.Logger.Info("避让指令重试",
		zap.Uint64("uav_id", tracker.uavID),
		zap.Int("retry_count", tracker.retryCount),
	)
}

func (s *CollisionAvoidanceService) trackAvoidCommand(uavID uint64, targetFactor float64, baseSpeed float64) {
	tracker := &avoidanceCmdTracker{
		uavID:        uavID,
		targetFactor: targetFactor,
		sentAt:       time.Now(),
		baseSpeed:    baseSpeed,
	}

	cmdTrackersMu.Lock()
	cmdTrackers[uavID] = tracker
	cmdTrackersMu.Unlock()

	s.broadcastCommandStatus(uavID, tracker)
}

func (s *CollisionAvoidanceService) verifySpeedAdjustment(pos *models.UAVLivePosition) {
	cmdTrackersMu.Lock()
	tracker, exists := cmdTrackers[pos.UAVID]
	cmdTrackersMu.Unlock()

	if !exists || tracker.verified {
		return
	}

	if !tracker.acked {
		if time.Since(tracker.sentAt) > 3*time.Second {
			if tracker.retryCount < 2 {
				tracker.retryCount++
				go s.retryAvoidCommand(tracker)
			} else {
				tracker.ackResult = 99
				s.broadcastCommandStatus(pos.UAVID, tracker)
			}
		}
		return
	}

	if tracker.baseSpeed <= 0 {
		tracker.verified = true
		return
	}

	expectedSpeed := tracker.baseSpeed * tracker.targetFactor
	speedDiff := math.Abs(pos.GroundSpeed - expectedSpeed)
	tolerance := tracker.baseSpeed * 0.15

	if speedDiff < tolerance || tracker.targetFactor == defaultSpeedFactor {
		tracker.verified = true
		middleware.Logger.Info("避让速度调整验证通过",
			zap.Uint64("uav_id", pos.UAVID),
			zap.Float64("target_factor", tracker.targetFactor),
			zap.Float64("current_speed", pos.GroundSpeed),
			zap.Float64("expected_speed", expectedSpeed),
		)
		s.broadcastCommandStatus(pos.UAVID, tracker)
	}
}

func (s *CollisionAvoidanceService) broadcastCommandStatus(uavID uint64, tracker *avoidanceCmdTracker) {
	status := "pending"
	if tracker.verified {
		status = "verified"
	} else if tracker.acked {
		if tracker.ackResult == 0 {
			status = "acked"
		} else {
			status = "failed"
		}
	}

	data := map[string]interface{}{
		"type":           "avoid_command_status",
		"uav_id":         uavID,
		"status":         status,
		"target_factor":  tracker.targetFactor,
		"retry_count":    tracker.retryCount,
		"ack_result":     tracker.ackResult,
		"timestamp":      time.Now().UnixNano() / 1e6,
	}
	websocket.NewHub().Broadcast(data)
}
