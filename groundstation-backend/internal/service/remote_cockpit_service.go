package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"groundstation-backend/internal/config"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/webrtc"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type VideoStreamConfig struct {
	Codec            models.VideoCodec `json:"codec"`
	Resolution       models.VideoResolution `json:"resolution"`
	BitrateKbps      int    `json:"bitrate_kbps"`
	FPS              int    `json:"fps"`
	KeyframeInterval int    `json:"keyframe_interval"`
	AdaptiveEnabled  bool   `json:"adaptive_enabled"`
	MinBitrateKbps   int    `json:"min_bitrate_kbps"`
	MaxBitrateKbps   int    `json:"max_bitrate_kbps"`
	MinResolution    models.VideoResolution `json:"min_resolution"`
	MaxResolution    models.VideoResolution `json:"max_resolution"`
}

type CockpitLinkStatus struct {
	PrimaryLink         models.LinkType  `json:"primary_link"`
	SecondaryLink       models.LinkType  `json:"secondary_link"`
	PrimaryState        models.LinkState `json:"primary_state"`
	SecondaryState      models.LinkState `json:"secondary_state"`
	PrimaryLatencyMs    int    `json:"primary_latency_ms"`
	SecondaryLatencyMs  int    `json:"secondary_latency_ms"`
	PrimaryPacketLoss   float64 `json:"primary_packet_loss"`
	SecondaryPacketLoss float64 `json:"secondary_packet_loss"`
	FailoverEnabled     bool   `json:"failover_enabled"`
	FailoverThresholdMs int    `json:"failover_threshold_ms"`
	FailoverCount       int    `json:"failover_count"`
	LastFailoverTime    *time.Time `json:"last_failover_time,omitempty"`
}

type NetworkMetrics struct {
	BandwidthKbps  float64 `json:"bandwidth_estimate_kbps"`
	RTTms          int     `json:"rtt_ms"`
	PacketLoss     float64 `json:"packet_loss"`
	JitterMs       int     `json:"jitter_ms"`
	ThroughputKbps float64 `json:"throughput_kbps"`
	Timestamp      int64   `json:"timestamp"`
}

type RemoteCockpitService struct {
	db                 *gorm.DB
	uavService         *UAVService
	missionService     *MissionService
	linkService        *LinkService
	uavRepo            interface{}
	activeSessions     map[uint64]*models.CockpitSession
	sessionsMu         sync.RWMutex
	videoStreams       map[uint64]*models.VideoStreamSession
	videoMu            sync.RWMutex
	linkSnapshots      map[uint64]*models.CockpitLinkSnapshot
	linkMu             sync.RWMutex
	signalingMgr       *webrtc.SignalingManager
	streamProxy        *webrtc.StreamProxy
}

var cockpitService *RemoteCockpitService
var cockpitOnce sync.Once

func NewRemoteCockpitService() *RemoteCockpitService {
	cockpitOnce.Do(func() {
		sm := webrtc.NewSignalingManager()
		cockpitService = &RemoteCockpitService{
			db:             config.DB,
			uavService:     NewUAVService(),
			missionService: NewMissionService(),
			linkService:    NewLinkService(),
			activeSessions: make(map[uint64]*models.CockpitSession),
			videoStreams:   make(map[uint64]*models.VideoStreamSession),
			linkSnapshots:  make(map[uint64]*models.CockpitLinkSnapshot),
			signalingMgr:   sm,
			streamProxy:    webrtc.NewStreamProxy(sm),
		}
	})
	return cockpitService
}

func (s *RemoteCockpitService) StartSession(uavID, pilotID uint64) (*models.CockpitSession, error) {
	uav, err := s.uavService.GetByID(uavID)
	if err != nil {
		return nil, errors.New("无人机不存在")
	}
	if uav.Status == models.UAVStatusOffline {
		return nil, errors.New("无人机未连接")
	}

	s.sessionsMu.Lock()
	if existing, ok := s.activeSessions[uavID]; ok && existing.EndTime == nil {
		s.sessionsMu.Unlock()
		return existing, nil
	}

	sessionID := fmt.Sprintf("cockpit_%d_%d", uavID, time.Now().UnixNano())
	now := time.Now()
	session := &models.CockpitSession{
		SessionID: sessionID,
		UAVID:     uavID,
		PilotID:   pilotID,
		StartTime: now,
		Mode:      models.CockpitModeConnecting,
	}

	if err := s.db.Create(session).Error; err != nil {
		s.sessionsMu.Unlock()
		return nil, err
	}
	s.activeSessions[uavID] = session
	s.sessionsMu.Unlock()

	if err := s.initLinkSnapshot(uavID); err != nil {
		middleware.Logger.Warn("Failed to init link snapshot", zapError(err)...)
	}

	websocket.BroadcastRemoteCockpitSession(uavID, map[string]interface{}{
		"session_id": sessionID,
		"uav_id":     uavID,
		"pilot_id":   pilotID,
		"start_time": now.UnixMilli(),
		"mode":       string(models.CockpitModeConnecting),
	})

	return session, nil
}

func (s *RemoteCockpitService) EndSession(uavID uint64) (*models.CockpitSession, error) {
	s.sessionsMu.Lock()
	session, ok := s.activeSessions[uavID]
	if !ok {
		s.sessionsMu.Unlock()
		return nil, errors.New("会话不存在")
	}

	now := time.Now()
	session.EndTime = &now
	session.Mode = models.CockpitModeIdle
	session.TotalFlightMs = now.Sub(session.StartTime).Milliseconds()
	delete(s.activeSessions, uavID)
	s.sessionsMu.Unlock()

	s.db.Save(session)

	_ = s.StopVideoStream(uavID)

	websocket.BroadcastRemoteCockpitSessionEnd(uavID, session.SessionID)
	return session, nil
}

func (s *RemoteCockpitService) GetSession(uavID uint64) (*models.CockpitSession, error) {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	session, ok := s.activeSessions[uavID]
	if !ok {
		var dbSession models.CockpitSession
		if err := s.db.Where("uav_id = ? AND end_time IS NULL", uavID).First(&dbSession).Error; err != nil {
			return nil, err
		}
		return &dbSession, nil
	}
	return session, nil
}

func (s *RemoteCockpitService) StartVideoStream(uavID uint64, config *VideoStreamConfig) (*models.VideoStreamSession, error) {
	_, err := s.uavService.GetByID(uavID)
	if err != nil {
		return nil, errors.New("无人机不存在")
	}

	s.videoMu.Lock()
	defer s.videoMu.Unlock()

	var stream models.VideoStreamSession
	err = s.db.Where("uav_id = ? AND active = ?", uavID, true).First(&stream).Error
	if err == nil {
		return &stream, nil
	}

	if config == nil {
		config = &VideoStreamConfig{
			Codec:            models.VideoCodecH265,
			Resolution:       models.VideoRes720P,
			BitrateKbps:      4000,
			FPS:              30,
			KeyframeInterval: 60,
			AdaptiveEnabled:  true,
			MinBitrateKbps:   1000,
			MaxBitrateKbps:   8000,
			MinResolution:    models.VideoRes640P,
			MaxResolution:    models.VideoRes1080P,
		}
	}

	now := time.Now()
	localIP := utils.GetLocalIP()
	streamURL := fmt.Sprintf("webrtc://%s:8888/stream/uav_%d", localIP, uavID)

	stream = models.VideoStreamSession{
		UAVID:              uavID,
		Codec:              config.Codec,
		Resolution:         config.Resolution,
		TargetBitrateKbps:  config.BitrateKbps,
		CurrentBitrateKbps: config.BitrateKbps,
		FPS:                config.FPS,
		KeyframeInterval:   config.KeyframeInterval,
		AdaptiveEnabled:    config.AdaptiveEnabled,
		MinBitrateKbps:     config.MinBitrateKbps,
		MaxBitrateKbps:     config.MaxBitrateKbps,
		MinResolution:      config.MinResolution,
		MaxResolution:      config.MaxResolution,
		StreamURL:          streamURL,
		Protocol:           "webrtc",
		Active:             true,
		StartedAt:          &now,
	}

	if err := s.db.Create(&stream).Error; err != nil {
		return nil, err
	}

	s.videoStreams[uavID] = &stream

	s.streamProxy.RegisterStream(uavID, streamURL)

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(uavID, "video_start", map[string]interface{}{
		"param1": float32(getResolutionWidth(config.Resolution)),
		"param2": float32(getResolutionHeight(config.Resolution)),
		"param3": float32(config.BitrateKbps),
		"param4": float32(config.FPS),
		"param5": float32(1),
	})

	websocket.BroadcastVideoStreamStatus(uavID, map[string]interface{}{
		"active":            true,
		"codec":             config.Codec,
		"resolution":        config.Resolution,
		"current_bitrate_kbps": config.BitrateKbps,
		"target_bitrate_kbps":  config.BitrateKbps,
		"fps":               config.FPS,
		"stream_url":        streamURL,
		"protocol":          "webrtc",
		"latency_ms":        0,
		"jitter_ms":         0,
		"packet_loss":       0,
		"frames_decoded":    0,
		"frames_dropped":    0,
		"last_frame_time":   now.UnixMilli(),
	})

	return &stream, nil
}

func (s *RemoteCockpitService) StopVideoStream(uavID uint64) error {
	s.videoMu.Lock()
	defer s.videoMu.Unlock()

	var stream models.VideoStreamSession
	if err := s.db.Where("uav_id = ? AND active = ?", uavID, true).First(&stream).Error; err != nil {
		return nil
	}

	now := time.Now()
	stream.Active = false
	stream.StoppedAt = &now
	s.db.Save(&stream)
	delete(s.videoStreams, uavID)

	s.signalingMgr.ClosePeerConnection(uavID)
	s.streamProxy.RemoveStream(uavID)

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(uavID, "video_stop", map[string]interface{}{})

	websocket.BroadcastVideoStreamDisconnected(uavID)
	return nil
}

func (s *RemoteCockpitService) GetVideoStream(uavID uint64) (*models.VideoStreamSession, error) {
	s.videoMu.RLock()
	if stream, ok := s.videoStreams[uavID]; ok {
		s.videoMu.RUnlock()
		return stream, nil
	}
	s.videoMu.RUnlock()

	var stream models.VideoStreamSession
	if err := s.db.Where("uav_id = ? AND active = ?", uavID, true).First(&stream).Error; err != nil {
		return nil, err
	}
	return &stream, nil
}

func (s *RemoteCockpitService) AdjustVideoQuality(uavID uint64, bitrateKbps *int, resolution *models.VideoResolution) (*models.VideoStreamSession, error) {
	stream, err := s.GetVideoStream(uavID)
	if err != nil {
		return nil, errors.New("视频流未启动")
	}

	if bitrateKbps != nil {
		stream.TargetBitrateKbps = *bitrateKbps
		stream.CurrentBitrateKbps = *bitrateKbps
	}
	if resolution != nil {
		stream.Resolution = *resolution
	}
	stream.QualityAdjustments++
	stream.UpdatedAt = time.Now()

	s.db.Save(stream)

	s.videoMu.Lock()
	s.videoStreams[uavID] = stream
	s.videoMu.Unlock()

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(uavID, "video_quality_adjust", map[string]interface{}{
		"bitrate_kbps": stream.TargetBitrateKbps,
		"resolution":   stream.Resolution,
	})

	websocket.BroadcastVideoQualityAdjusted(uavID, map[string]interface{}{
		"new_bitrate_kbps": stream.TargetBitrateKbps,
		"new_resolution":   stream.Resolution,
	})

	return stream, nil
}

func (s *RemoteCockpitService) GetStreamURL(uavID uint64, protocol string) (map[string]interface{}, error) {
	stream, err := s.GetVideoStream(uavID)
	if err != nil {
		return nil, errors.New("视频流未启动")
	}

	localIP := utils.GetLocalIP()
	var url string
	switch protocol {
	case "hls":
		url = fmt.Sprintf("http://%s:8080/hls/uav_%d/index.m3u8", localIP, uavID)
	case "ws":
		url = fmt.Sprintf("ws://%s:8889/ws/uav_%d", localIP, uavID)
	default:
		url = stream.StreamURL
	}

	return map[string]interface{}{
		"url":      url,
		"protocol": protocol,
	}, nil
}

func (s *RemoteCockpitService) initLinkSnapshot(uavID uint64) error {
	s.linkMu.Lock()
	defer s.linkMu.Unlock()

	if _, ok := s.linkSnapshots[uavID]; ok {
		return nil
	}

	now := time.Now()
	snapshot := &models.CockpitLinkSnapshot{
		UAVID:               uavID,
		PrimaryLink:         models.LinkTypeLTE,
		SecondaryLink:       models.LinkTypeRadio,
		PrimaryState:        models.LinkStateDisconnected,
		SecondaryState:      models.LinkStateDisconnected,
		FailoverEnabled:     true,
		FailoverThresholdMs: 200,
		AutoMissionFallback: true,
		Timestamp:           now,
	}

	if latest, err := s.linkService.GetLatestByUAVID(uavID); err == nil && latest != nil {
		if latest.ActiveLink == models.LinkTypeRadio {
			snapshot.PrimaryLink = models.LinkTypeRadio
			snapshot.SecondaryLink = models.LinkTypeLTE
		}
		if latest.RadioConnected {
			if snapshot.PrimaryLink == models.LinkTypeRadio {
				snapshot.PrimaryState = models.LinkStateConnected
			} else {
				snapshot.SecondaryState = models.LinkStateConnected
			}
			snapshot.PrimaryLatencyMs = int(latest.LatencyMs)
			snapshot.PrimaryPacketLoss = latest.PacketLoss
		}
		if latest.LteConnected {
			if snapshot.PrimaryLink == models.LinkTypeLTE {
				snapshot.PrimaryState = models.LinkStateConnected
			} else {
				snapshot.SecondaryState = models.LinkStateConnected
			}
			snapshot.SecondaryLatencyMs = int(latest.LatencyMs)
			snapshot.SecondaryPacketLoss = latest.PacketLoss
		}
		snapshot.AutoMissionFallback = latest.AutoSwitchEnabled
	}

	s.db.Create(snapshot)
	s.linkSnapshots[uavID] = snapshot

	return nil
}

func (s *RemoteCockpitService) GetLinkStatus(uavID uint64) (*CockpitLinkStatus, error) {
	s.linkMu.RLock()
	snapshot, ok := s.linkSnapshots[uavID]
	s.linkMu.RUnlock()

	if !ok {
		if err := s.initLinkSnapshot(uavID); err != nil {
			return nil, err
		}
		s.linkMu.RLock()
		snapshot = s.linkSnapshots[uavID]
		s.linkMu.RUnlock()
	}

	status := &CockpitLinkStatus{
		PrimaryLink:         snapshot.PrimaryLink,
		SecondaryLink:       snapshot.SecondaryLink,
		PrimaryState:        snapshot.PrimaryState,
		SecondaryState:      snapshot.SecondaryState,
		PrimaryLatencyMs:    snapshot.PrimaryLatencyMs,
		SecondaryLatencyMs:  snapshot.SecondaryLatencyMs,
		PrimaryPacketLoss:   snapshot.PrimaryPacketLoss,
		SecondaryPacketLoss: snapshot.SecondaryPacketLoss,
		FailoverEnabled:     snapshot.FailoverEnabled,
		FailoverThresholdMs: snapshot.FailoverThresholdMs,
		FailoverCount:       snapshot.FailoverCount,
		LastFailoverTime:    snapshot.LastFailoverTime,
	}

	return status, nil
}

func (s *RemoteCockpitService) SetPrimaryLink(uavID uint64, linkType models.LinkType) (*CockpitLinkStatus, error) {
	s.linkMu.Lock()
	defer s.linkMu.Unlock()

	snapshot, ok := s.linkSnapshots[uavID]
	if !ok {
		if err := s.initLinkSnapshot(uavID); err != nil {
			return nil, err
		}
		snapshot = s.linkSnapshots[uavID]
	}

	oldPrimary := snapshot.PrimaryLink
	if oldPrimary == linkType {
		return s.snapshotToStatus(snapshot), nil
	}

	snapshot.PrimaryLink = linkType
	if linkType == models.LinkTypeLTE {
		snapshot.SecondaryLink = models.LinkTypeRadio
	} else {
		snapshot.SecondaryLink = models.LinkTypeLTE
	}
	snapshot.Timestamp = time.Now()
	s.db.Save(snapshot)

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(uavID, "set_primary_link", map[string]interface{}{
		"link_type": linkType,
	})

	websocket.BroadcastRemoteCockpitLinkStatus(uavID, s.snapshotToStatus(snapshot))

	return s.snapshotToStatus(snapshot), nil
}

func (s *RemoteCockpitService) SetFailoverEnabled(uavID uint64, enabled bool) (*CockpitLinkStatus, error) {
	s.linkMu.Lock()
	defer s.linkMu.Unlock()

	snapshot, ok := s.linkSnapshots[uavID]
	if !ok {
		if err := s.initLinkSnapshot(uavID); err != nil {
			return nil, err
		}
		snapshot = s.linkSnapshots[uavID]
	}

	snapshot.FailoverEnabled = enabled
	snapshot.Timestamp = time.Now()
	s.db.Save(snapshot)

	return s.snapshotToStatus(snapshot), nil
}

func (s *RemoteCockpitService) SetAutoMissionFallback(uavID uint64, enabled bool) error {
	s.linkMu.Lock()
	defer s.linkMu.Unlock()

	snapshot, ok := s.linkSnapshots[uavID]
	if !ok {
		if err := s.initLinkSnapshot(uavID); err != nil {
			return err
		}
		snapshot = s.linkSnapshots[uavID]
	}

	snapshot.AutoMissionFallback = enabled
	snapshot.Timestamp = time.Now()
	s.db.Save(snapshot)

	return nil
}

func (s *RemoteCockpitService) TriggerAutoMissionFallback(uavID uint64, reason string) error {
	s.linkMu.RLock()
	snapshot, ok := s.linkSnapshots[uavID]
	s.linkMu.RUnlock()

	if !ok || !snapshot.AutoMissionFallback {
		return errors.New("自动回退未启用")
	}

	s.sessionsMu.RLock()
	session, hasSession := s.activeSessions[uavID]
	s.sessionsMu.RUnlock()

	if hasSession {
		session.Mode = models.CockpitModeMission
		session.AutoFallbackUsed = true
		s.db.Save(session)
	}

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(uavID, "set_mode", map[string]interface{}{
		"mode":   "AUTO",
		"reason": reason,
	})

	activeMission, err := s.missionService.GetActiveMission(uavID)
	if err == nil && activeMission != nil {
		_, _ = s.missionService.ResumeMission(activeMission.ID)
	} else {
		_ = s.activateHoldMission(uavID)
	}

	websocket.BroadcastAutoMissionFallback(uavID, reason)

	uav, _ := s.uavService.GetByID(uavID)
	if uav != nil {
		_ = s.uavService.UpdateStatus(uavID, models.UAVStatusFlying)
	}

	return nil
}

func (s *RemoteCockpitService) CheckAndTriggerFallback(uavID uint64, bothLinksDown bool, videoDisconnected bool) {
	if !bothLinksDown && !videoDisconnected {
		return
	}

	s.linkMu.RLock()
	snapshot, ok := s.linkSnapshots[uavID]
	s.linkMu.RUnlock()

	if !ok || !snapshot.AutoMissionFallback {
		return
	}

	s.sessionsMu.RLock()
	session, hasSession := s.activeSessions[uavID]
	s.sessionsMu.RUnlock()

	if !hasSession {
		return
	}

	if session.Mode == models.CockpitModeMission {
		return
	}

	reason := "all_links_disconnected"
	if videoDisconnected && !bothLinksDown {
		reason = "video_link_disconnected"
	}

	_ = s.TriggerAutoMissionFallback(uavID, reason)
}

func (s *RemoteCockpitService) activateHoldMission(uavID uint64) error {
	uav, err := s.uavService.GetByID(uavID)
	if err != nil {
		return err
	}

	flightStatus, err := s.uavService.GetFlightStatus(uavID)
	if err != nil {
		return err
	}

	waypoints := []models.MissionWaypoint{
		{
			Latitude:    flightStatus.Latitude,
			Longitude:   flightStatus.Longitude,
			Altitude:    flightStatus.Altitude + 10,
			HoldTime:    180,
			AcceptRadius: 5,
		},
		{
			Latitude:    uav.HomeLatitude,
			Longitude:   uav.HomeLongitude,
			Altitude:    uav.HomeAltitude + 30,
			AcceptRadius: 5,
		},
	}

	mission := &models.FlightMission{
		UAVID:        uavID,
		Name:         "Emergency Hold - " + time.Now().Format("2006-01-02 15:04:05"),
		Description:  "由图传断连自动触发的应急航线",
		Waypoints:    waypoints,
		MaxAltitude:  flightStatus.Altitude + 50,
		Speed:        5,
	}

	created, err := s.missionService.CreateMission(mission)
	if err != nil {
		return err
	}

	_, _ = s.missionService.StartMission(created.ID)
	return nil
}

func (s *RemoteCockpitService) GetAvailableUAVs(pilotID uint64) ([]uint64, error) {
	uavs, _, err := s.uavService.List(&utils.Pagination{Page: 1, PageSize: 100}, "", "", "")
	if err != nil {
		return nil, err
	}

	var available []uint64
	for _, uav := range uavs {
		if uav.Status == models.UAVStatusOnline || uav.Status == models.UAVStatusFlying || uav.Status == models.UAVStatusHovering {
			available = append(available, uav.ID)
		}
	}
	return available, nil
}

func (s *RemoteCockpitService) SwitchUAV(fromUavID, toUavID, pilotID uint64) error {
	if _, err := s.uavService.GetByID(toUavID); err != nil {
		return errors.New("目标无人机不存在")
	}

	hasActiveFrom := false
	s.sessionsMu.RLock()
	if _, ok := s.activeSessions[fromUavID]; ok {
		hasActiveFrom = true
	}
	s.sessionsMu.RUnlock()

	if hasActiveFrom {
		_, _ = s.EndSession(fromUavID)
	}

	_, err := s.StartSession(toUavID, pilotID)
	if err != nil {
		return err
	}

	return nil
}

func (s *RemoteCockpitService) SendFlightControl(uavID, pilotID uint64, pitch, roll, yaw, throttle float64, source string) error {
	s.sessionsMu.RLock()
	session, ok := s.activeSessions[uavID]
	s.sessionsMu.RUnlock()
	if !ok {
		return errors.New("会话未启动")
	}

	clampedPitch := math.Max(-1, math.Min(1, pitch))
	clampedRoll := math.Max(-1, math.Min(1, roll))
	clampedYaw := math.Max(-1, math.Min(1, yaw))
	clampedThrottle := math.Max(0, math.Min(1, throttle))

	cmdMgr := mavlink.NewCommandManager()
	vx := clampedRoll * 5
	vy := clampedPitch * 5
	vz := (clampedThrottle - 0.5) * 5
	yawRate := clampedYaw * 90

	_ = cmdMgr.SendCustomCommand(uavID, "velocity", map[string]interface{}{
		"vx":       vx,
		"vy":       vy,
		"vz":       vz,
		"yaw_rate": yawRate,
		"source":   source,
	})

	log := &models.FlightControlLog{
		SessionID: session.SessionID,
		UAVID:     uavID,
		PilotID:   pilotID,
		Pitch:     clampedPitch,
		Roll:      clampedRoll,
		Yaw:       clampedYaw,
		Throttle:  clampedThrottle,
		Source:    source,
		Timestamp: time.Now(),
	}
	s.db.Create(log)

	s.sessionsMu.Lock()
	session.CommandsSent++
	s.db.Save(session)
	s.sessionsMu.Unlock()

	return nil
}

func (s *RemoteCockpitService) LogNetworkMetrics(uavID uint64, metrics *NetworkMetrics) error {
	log := &models.NetworkMetricsLog{
		UAVID:          uavID,
		BandwidthKbps:  metrics.BandwidthKbps,
		RTTms:          metrics.RTTms,
		PacketLoss:     metrics.PacketLoss,
		JitterMs:       metrics.JitterMs,
		ThroughputKbps: metrics.ThroughputKbps,
		Timestamp:      time.UnixMilli(metrics.Timestamp),
	}
	return s.db.Create(log).Error
}

func (s *RemoteCockpitService) UpdateSessionMode(uavID uint64, mode models.CockpitMode) error {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	session, ok := s.activeSessions[uavID]
	if !ok {
		return errors.New("会话未启动")
	}

	session.Mode = mode
	s.db.Save(session)

	websocket.BroadcastRemoteCockpitMode(uavID, string(mode))
	return nil
}

func (s *RemoteCockpitService) HandleSDPOffer(uavID uint64, sdpOffer string) (string, error) {
	answer, err := s.signalingMgr.HandleSDPOffer(uavID, sdpOffer)
	if err != nil {
		return "", err
	}

	if _, ok := s.streamProxy.GetUpstreamURL(uavID); ok {
		go func() {
			_ = s.streamProxy.StartForwarding(context.Background(), uavID)
		}()
	}

	return answer, nil
}

func (s *RemoteCockpitService) GetWebRTCStats(uavID uint64) (*webrtc.StreamStats, error) {
	stats := s.signalingMgr.GetStreamStats(uavID)
	if stats == nil {
		return nil, errors.New("WebRTC连接不存在")
	}
	return stats, nil
}

func (s *RemoteCockpitService) snapshotToStatus(snapshot *models.CockpitLinkSnapshot) *CockpitLinkStatus {
	return &CockpitLinkStatus{
		PrimaryLink:         snapshot.PrimaryLink,
		SecondaryLink:       snapshot.SecondaryLink,
		PrimaryState:        snapshot.PrimaryState,
		SecondaryState:      snapshot.SecondaryState,
		PrimaryLatencyMs:    snapshot.PrimaryLatencyMs,
		SecondaryLatencyMs:  snapshot.SecondaryLatencyMs,
		PrimaryPacketLoss:   snapshot.PrimaryPacketLoss,
		SecondaryPacketLoss: snapshot.SecondaryPacketLoss,
		FailoverEnabled:     snapshot.FailoverEnabled,
		FailoverThresholdMs: snapshot.FailoverThresholdMs,
		FailoverCount:       snapshot.FailoverCount,
		LastFailoverTime:    snapshot.LastFailoverTime,
	}
}

func getResolutionWidth(res models.VideoResolution) int {
	switch res {
	case models.VideoRes480P:
		return 480
	case models.VideoRes640P:
		return 640
	case models.VideoRes960P:
		return 960
	case models.VideoRes720P:
		return 1280
	case models.VideoRes1080P:
		return 1920
	}
	return 1280
}

func getResolutionHeight(res models.VideoResolution) int {
	switch res {
	case models.VideoRes480P:
		return 360
	case models.VideoRes640P:
		return 480
	case models.VideoRes960P:
		return 540
	case models.VideoRes720P:
		return 720
	case models.VideoRes1080P:
		return 1080
	}
	return 720
}

func zapError(err error) []interface{} {
	if err == nil {
		return nil
	}
	return []interface{}{"error", err}
}
