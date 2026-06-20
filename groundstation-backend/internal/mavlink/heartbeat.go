package mavlink

import (
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/nsq"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/service"
	"sync"
	"time"
)

type HeartbeatManager struct {
	uavRepo     *repository.UAVRepository
	uavStatuses map[uint64]*UAVHeartbeatStatus
	mu          sync.RWMutex
	interval    time.Duration
	timeout     time.Duration
}

type UAVHeartbeatStatus struct {
	SystemID     uint8
	ComponentID  uint8
	LastSeen     time.Time
	Status       models.UAVStatus
	BaseMode     uint8
	CustomMode   uint32
	SystemStatus uint8
}

var heartbeatManager *HeartbeatManager
var heartbeatOnce sync.Once

func NewHeartbeatManager() *HeartbeatManager {
	heartbeatOnce.Do(func() {
		cfg := config.AppConfig.Heartbeat
		heartbeatManager = &HeartbeatManager{
			uavRepo:     repository.NewUAVRepository(),
			uavStatuses: make(map[uint64]*UAVHeartbeatStatus),
			interval:    time.Duration(cfg.Interval) * time.Millisecond,
			timeout:     time.Duration(cfg.Timeout) * time.Millisecond,
		}
		go heartbeatManager.startMonitor()
	})
	return heartbeatManager
}

func (m *HeartbeatManager) startMonitor() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for range ticker.C {
		m.checkTimeouts()
	}
}

func (m *HeartbeatManager) checkTimeouts() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for uavID, status := range m.uavStatuses {
		if now.Sub(status.LastSeen) > m.timeout {
			if status.Status != models.UAVStatusOffline {
				status.Status = models.UAVStatusOffline
				_ = m.updateUAVStatus(uavID, status)
				_ = service.NewAlertService().CreateSignalLossAlert(uavID)
			}
		}
	}
}

func (m *HeartbeatManager) ProcessHeartbeat(uavID uint64, systemID, componentID uint8,
	baseMode uint8, customMode uint32, systemStatus uint8) {

	m.mu.Lock()
	defer m.mu.Unlock()

	status, exists := m.uavStatuses[uavID]
	if !exists {
		status = &UAVHeartbeatStatus{}
		m.uavStatuses[uavID] = status
	}

	status.SystemID = systemID
	status.ComponentID = componentID
	status.LastSeen = time.Now()
	status.BaseMode = baseMode
	status.CustomMode = customMode
	status.SystemStatus = systemStatus

	newStatus := models.UAVStatusOnline
	if (baseMode & 0x80) != 0 {
		newStatus = models.UAVStatusFlying
	}

	if status.Status != newStatus || !exists {
		status.Status = newStatus
		_ = m.updateUAVStatus(uavID, status)
	}

	_ = nsq.Publish(nsq.TopicUAVHeartbeat, map[string]interface{}{
		"uav_id":       uavID,
		"system_id":    systemID,
		"component_id": componentID,
		"base_mode":    baseMode,
		"custom_mode":  customMode,
		"system_status": systemStatus,
		"flight_mode":  GetFlightModeName(customMode),
		"timestamp":    time.Now().UnixNano() / 1e6,
	})
}

func (m *HeartbeatManager) updateUAVStatus(uavID uint64, status *UAVHeartbeatStatus) error {
	uav, err := m.uavRepo.FindByID(uavID)
	if err != nil {
		return err
	}

	uav.Status = status.Status
	uav.LastHeartbeat = status.LastSeen
	uav.FlightMode = GetFlightModeName(status.CustomMode)
	uav.Armed = (status.BaseMode & 0x80) != 0

	return m.uavRepo.Update(uav)
}

func (m *HeartbeatManager) GetStatus(uavID uint64) (*UAVHeartbeatStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.uavStatuses[uavID]
	return status, exists
}

func (m *HeartbeatManager) GetAllStatuses() map[uint64]*UAVHeartbeatStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uint64]*UAVHeartbeatStatus)
	for k, v := range m.uavStatuses {
		result[k] = v
	}
	return result
}

func (m *HeartbeatManager) GetOnlineUAVs() []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]uint64, 0)
	now := time.Now()
	for uavID, status := range m.uavStatuses {
		if now.Sub(status.LastSeen) <= m.timeout {
			result = append(result, uavID)
		}
	}
	return result
}
