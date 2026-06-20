package service

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsService struct {
	uavRepo       *repository.UAVRepository
	flightRepo    *repository.FlightRepository
	missionRepo   *repository.MissionRepository
	alertRepo     *repository.AlertRepository
	blackboxRepo  *repository.BlackboxRepository

	mu sync.RWMutex

	uavTotal           prometheus.Gauge
	uavOnline          prometheus.Gauge
	uavFlying          prometheus.Gauge
	uavOffline         prometheus.Gauge
	uavError           prometheus.Gauge

	missionTotal       prometheus.Counter
	missionRunning     prometheus.Gauge
	missionCompleted   prometheus.Counter
	missionFailed      prometheus.Counter

	alertTotal         prometheus.Counter
	alertWarning       prometheus.Counter
	alertCritical      prometheus.Counter
	alertFatal         prometheus.Counter
	alertResolved      prometheus.Counter

	telemetryReceived  prometheus.Counter
	telemetryLatency   prometheus.Histogram

	apiRequestTotal    prometheus.Counter
	apiRequestDuration prometheus.Histogram
	apiErrorTotal      prometheus.Counter

	websocketConnections prometheus.Gauge
}

func NewMetricsService() *MetricsService {
	return &MetricsService{
		uavRepo:      repository.NewUAVRepository(),
		flightRepo:   repository.NewFlightRepository(),
		missionRepo:  repository.NewMissionRepository(),
		alertRepo:    repository.NewAlertRepository(),
		blackboxRepo: repository.NewBlackboxRepository(),

		uavTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "uav_total",
			Help: "Total number of UAVs",
		}),
		uavOnline: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "uav_online",
			Help: "Number of online UAVs",
		}),
		uavFlying: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "uav_flying",
			Help: "Number of flying UAVs",
		}),
		uavOffline: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "uav_offline",
			Help: "Number of offline UAVs",
		}),
		uavError: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "uav_error",
			Help: "Number of UAVs in error state",
		}),

		missionTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mission_total",
			Help: "Total number of missions",
		}),
		missionRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mission_running",
			Help: "Number of running missions",
		}),
		missionCompleted: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mission_completed_total",
			Help: "Total number of completed missions",
		}),
		missionFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mission_failed_total",
			Help: "Total number of failed missions",
		}),

		alertTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alert_total",
			Help: "Total number of alerts",
		}),
		alertWarning: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alert_warning_total",
			Help: "Total number of warning alerts",
		}),
		alertCritical: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alert_critical_total",
			Help: "Total number of critical alerts",
		}),
		alertFatal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alert_fatal_total",
			Help: "Total number of fatal alerts",
		}),
		alertResolved: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alert_resolved_total",
			Help: "Total number of resolved alerts",
		}),

		telemetryReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "telemetry_received_total",
			Help: "Total number of telemetry messages received",
		}),
		telemetryLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "telemetry_latency_seconds",
			Help:    "Telemetry message latency",
			Buckets: prometheus.DefBuckets,
		}),

		apiRequestTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "api_request_total",
			Help: "Total number of API requests",
		}),
		apiRequestDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "API request duration",
			Buckets: prometheus.DefBuckets,
		}),
		apiErrorTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "api_error_total",
			Help: "Total number of API errors",
		}),

		websocketConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Number of active WebSocket connections",
		}),
	}
}

func (s *MetricsService) CollectUAVMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	total, _ := s.uavRepo.Count(&models.UAV{}, nil)
	online, _ := s.uavRepo.CountByStatus(models.UAVStatusOnline)
	flying, _ := s.uavRepo.CountByStatus(models.UAVStatusFlying)
	offline, _ := s.uavRepo.CountByStatus(models.UAVStatusOffline)
	errorStatus, _ := s.uavRepo.CountByStatus(models.UAVStatusError)

	s.uavTotal.Set(float64(total))
	s.uavOnline.Set(float64(online))
	s.uavFlying.Set(float64(flying))
	s.uavOffline.Set(float64(offline))
	s.uavError.Set(float64(errorStatus))
}

func (s *MetricsService) CollectMissionMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	running, _ := s.missionRepo.Count(&models.FlightMission{}, "status IN ?",
		[]models.MissionStatus{models.MissionStatusExecuting, models.MissionStatusPaused})

	s.missionRunning.Set(float64(running))
}

func (s *MetricsService) IncrementMission(status models.MissionStatus) {
	switch status {
	case models.MissionStatusCompleted:
		s.missionCompleted.Inc()
	case models.MissionStatusAborted, models.MissionStatusFailed:
		s.missionFailed.Inc()
	}
	s.missionTotal.Inc()
}

func (s *MetricsService) IncrementAlert(level models.AlertLevel, status models.AlertStatus) {
	s.alertTotal.Inc()
	switch level {
	case models.AlertLevelWarning:
		s.alertWarning.Inc()
	case models.AlertLevelCritical:
		s.alertCritical.Inc()
	case models.AlertLevelFatal:
		s.alertFatal.Inc()
	}
	if status == models.AlertStatusResolved {
		s.alertResolved.Inc()
	}
}

func (s *MetricsService) RecordTelemetry(latency time.Duration) {
	s.telemetryReceived.Inc()
	s.telemetryLatency.Observe(latency.Seconds())
}

func (s *MetricsService) RecordAPIRequest(method, path string, statusCode int, duration time.Duration) {
	s.apiRequestTotal.Inc()
	s.apiRequestDuration.Observe(duration.Seconds())
	if statusCode >= 400 {
		s.apiErrorTotal.Inc()
	}
}

func (s *MetricsService) SetWebSocketConnections(count int) {
	s.websocketConnections.Set(float64(count))
}

func (s *MetricsService) GetSummary() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalUAVs, _ := s.uavRepo.Count(&models.UAV{}, nil)
	onlineUAVs, _ := s.uavRepo.CountByStatus(models.UAVStatusOnline)
	flyingUAVs, _ := s.uavRepo.CountByStatus(models.UAVStatusFlying)

	runningMissions, _ := s.missionRepo.Count(&models.FlightMission{}, "status = ?", models.MissionStatusExecuting)
	today := time.Now().Truncate(24 * time.Hour)
	todayMissions, _ := s.missionRepo.Count(&models.FlightMission{}, "created_at >= ?", today)

	newAlerts, _ := s.alertRepo.Count(&models.AlertEvent{}, "status = ?", models.AlertStatusNew)
	todayAlerts, _ := s.alertRepo.Count(&models.AlertEvent{}, "created_at >= ?", today)

	totalFlights, _ := s.blackboxRepo.Count(&models.BlackboxLog{}, nil)
	todayFlights, _ := s.blackboxRepo.Count(&models.BlackboxLog{}, "created_at >= ?", today)

	return map[string]interface{}{
		"uavs": map[string]interface{}{
			"total":   totalUAVs,
			"online":  onlineUAVs,
			"flying":  flyingUAVs,
			"offline": totalUAVs - onlineUAVs,
		},
		"missions": map[string]interface{}{
			"running": runningMissions,
			"today":   todayMissions,
		},
		"alerts": map[string]interface{}{
			"new":   newAlerts,
			"today": todayAlerts,
		},
		"flights": map[string]interface{}{
			"total": totalFlights,
			"today": todayFlights,
		},
	}, nil
}

func (s *MetricsService) StartCollector(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CollectUAVMetrics()
			s.CollectMissionMetrics()
		}
	}()
}
