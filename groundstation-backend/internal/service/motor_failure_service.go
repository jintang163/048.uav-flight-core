package service

import (
	"encoding/binary"
	"errors"
	"fmt"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MotorFailureService struct {
	mu            sync.RWMutex
	motorStatus   map[uint64]map[int]*models.MotorStatus
	motorHistory  map[uint64]map[int][]int
	failureState  map[uint64]*MotorFailureState
	motorCount    map[uint64]int
	alertService  *AlertService
	uavService    *UAVService
	flightService *FlightService
}

type MotorFailureState struct {
	UAVID          uint64
	FailedMotors   []int
	PIDAdjusted    bool
	RTHTriggered   bool
	StartTime      time.Time
	LastUpdateTime time.Time
}

var motorFailureService *MotorFailureService
var motorFailureOnce sync.Once

func NewMotorFailureService() *MotorFailureService {
	motorFailureOnce.Do(func() {
		motorFailureService = &MotorFailureService{
			motorStatus:   make(map[uint64]map[int]*models.MotorStatus),
			motorHistory:  make(map[uint64]map[int][]int),
			failureState:  make(map[uint64]*MotorFailureState),
			motorCount:    make(map[uint64]int),
			alertService:  NewAlertService(),
			uavService:    NewUAVService(),
			flightService: NewFlightService(),
		}
	})
	return motorFailureService
}

func (s *MotorFailureService) UpdateMotorStatus(uavID uint64, status *models.MotorStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.motorStatus[uavID]; !ok {
		s.motorStatus[uavID] = make(map[int]*models.MotorStatus)
	}
	s.motorStatus[uavID][status.MotorIndex] = status

	if _, ok := s.motorHistory[uavID]; !ok {
		s.motorHistory[uavID] = make(map[int][]int)
	}
	history := s.motorHistory[uavID][status.MotorIndex]
	history = append(history, status.RPM)
	if len(history) > 30 {
		history = history[len(history)-30:]
	}
	s.motorHistory[uavID][status.MotorIndex] = history

	motorCount := len(s.motorStatus[uavID])
	s.motorCount[uavID] = motorCount

	return nil
}

func (s *MotorFailureService) UpdateMotorInfo(uavID uint64, motorIndex int, vendor string, model string, faultFlags int, errorCode int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if motors, ok := s.motorStatus[uavID]; ok {
		if motor, ok2 := motors[motorIndex]; ok2 {
			motor.Vendor = vendor
			motor.Model = model
			motor.FaultFlags = faultFlags
			motor.ErrorCode = errorCode
		}
	}
	return nil
}

func (s *MotorFailureService) DetectMotorFailure(uavID uint64, motorIndex int, status *models.MotorStatus) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	motorCount := s.motorCount[uavID]
	if motorCount < 6 {
		return false, nil
	}

	if status.Status != models.MotorStatusFault {
		return false, nil
	}

	history := s.motorHistory[uavID][motorIndex]
	consecutiveZeros := 0
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] == 0 {
			consecutiveZeros++
		} else {
			break
		}
	}

	isFailure := consecutiveZeros >= 3 || status.FaultFlags > 0
	if !isFailure {
		return false, nil
	}

	failureState, exists := s.failureState[uavID]
	if !exists {
		failureState = &MotorFailureState{
			UAVID:     uavID,
			StartTime: time.Now(),
		}
		s.failureState[uavID] = failureState
	}

	alreadyFailed := false
	for _, idx := range failureState.FailedMotors {
		if idx == motorIndex {
			alreadyFailed = true
			break
		}
	}
	if !alreadyFailed {
		failureState.FailedMotors = append(failureState.FailedMotors, motorIndex)
	}
	failureState.LastUpdateTime = time.Now()

	remainingMotors := motorCount - len(failureState.FailedMotors)
	if remainingMotors < 4 {
		middleware.Logger.Error("Critical: too many motor failures for controlled flight",
			zap.Uint64("uav_id", uavID),
			zap.Int("failed_count", len(failureState.FailedMotors)),
			zap.Int("total_count", motorCount),
		)
	}

	if !failureState.PIDAdjusted && motorCount >= 6 && remainingMotors >= 4 {
		s.adjustPIDParameters(uavID, failureState.FailedMotors, motorCount)
		failureState.PIDAdjusted = true
	}

	return true, nil
}

func (s *MotorFailureService) adjustPIDParameters(uavID uint64, failedMotors []int, totalMotors int) {
	cm := mavlink.NewCommandManager()

	type PIDConfig struct {
		P   float32
		I   float32
		D   float32
	}

	pidAdjustments := map[int]PIDConfig{
		6:  {P: 1.15, I: 1.10, D: 1.05},
		8:  {P: 1.10, I: 1.08, D: 1.03},
		10: {P: 1.08, I: 1.05, D: 1.02},
	}

	adjustment, ok := pidAdjustments[totalMotors]
	if !ok {
		adjustment = PIDConfig{P: 1.12, I: 1.08, D: 1.04}
	}

	paramIndices := []struct {
		index  uint8
		factor float32
	}{
		{0, adjustment.P},
		{1, adjustment.I},
		{2, adjustment.D},
	}

	for _, param := range paramIndices {
		encoder := mavlink.NewMAVLinkEncoder()
		cmdData := encoder.EncodeCommandLong(
			0, uint8(uavID), 0,
			mavlink.CMD_DO_SET_PARAMETER,
			param.index, 0, 0,
			0, 0, 0,
			float32(param.factor),
		)

		if err := cm.SendCommand(uavID, cmdData); err != nil {
			middleware.Logger.Error("Failed to send PID adjustment command",
				zap.Uint64("uav_id", uavID),
				zap.Error(err),
			)
		}
	}

	middleware.Logger.Info("PID parameters adjusted for motor failure",
		zap.Uint64("uav_id", uavID),
		zap.Ints("failed_motors", failedMotors),
		zap.Int("total_motors", totalMotors),
		zap.Float32("P_factor", adjustment.P),
		zap.Float32("I_factor", adjustment.I),
		zap.Float32("D_factor", adjustment.D),
	)
}

func (s *MotorFailureService) TriggerEmergencyRTH(uavID uint64, motorIndex int) error {
	s.mu.Lock()
	failureState, exists := s.failureState[uavID]
	if !exists {
		s.mu.Unlock()
		return errors.New("no failure state found")
	}

	if failureState.RTHTriggered {
		s.mu.Unlock()
		return nil
	}
	failureState.RTHTriggered = true
	s.mu.Unlock()

	cm := mavlink.NewCommandManager()
	encoder := mavlink.NewMAVLinkEncoder()
	rthData := encoder.EncodeCommandLong(
		0, uint8(uavID), 0,
		mavlink.CMD_NAV_RETURN_TO_LAUNCH,
		0, 0, 0, 0, 0, 0, 0,
	)

	if err := cm.SendCommand(uavID, rthData); err != nil {
		middleware.Logger.Error("Failed to send emergency RTH command",
			zap.Uint64("uav_id", uavID),
			zap.Error(err),
		)
		return err
	}

	middleware.Logger.Warn("Emergency RTH triggered due to motor failure",
		zap.Uint64("uav_id", uavID),
		zap.Int("failed_motor", motorIndex),
	)

	return nil
}

func (s *MotorFailureService) CreateMotorFailureAlert(uavID uint64, motorIndex int, status *models.MotorStatus) (*models.AlertEvent, error) {
	title := fmt.Sprintf("电机 #%d 失效告警", motorIndex+1)
	message := fmt.Sprintf("无人机 %d 的电机 #%d 检测到故障。故障标志: 0x%04X, RPM: %d, 温度: %d°C。飞控已自动调整PID参数并触发紧急返航。",
		uavID, motorIndex+1, status.FaultFlags, status.RPM, status.Temperature)

	alert, err := s.alertService.CreateCustomAlert(uavID, title, message, models.AlertLevelCritical)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if failureState, ok := s.failureState[uavID]; ok {
		repo := repository.NewBaseRepository()
		event := &models.MotorFailureEvent{
			UAVID:         uavID,
			MotorIndex:    motorIndex,
			FaultFlags:    status.FaultFlags,
			ErrorCode:     status.ErrorCode,
			RPMAtFailure:  status.RPM,
			TempAtFailure: status.Temperature,
			AlertID:       alert.ID,
			ActionTaken:   "pid_adjusted_rth",
		}
		_ = repo.Create(event)
	}
	s.mu.Unlock()

	return alert, nil
}

func (s *MotorFailureService) GetMotorStatuses(uavID uint64) []*models.MotorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	motors, ok := s.motorStatus[uavID]
	if !ok {
		return nil
	}

	result := make([]*models.MotorStatus, 0, len(motors))
	for _, status := range motors {
		result = append(result, status)
	}
	return result
}

func (s *MotorFailureService) GetFailureState(uavID uint64) *MotorFailureState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.failureState[uavID]
	if !ok {
		return nil
	}
	return state
}

func (s *MotorFailureService) GetMotorCount(uavID uint64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.motorCount[uavID]
}

func (s *MotorFailureService) ResolveFailure(uavID uint64, motorIndex int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	failureState, ok := s.failureState[uavID]
	if !ok {
		return nil
	}

	newFailed := make([]int, 0)
	for _, idx := range failureState.FailedMotors {
		if idx != motorIndex {
			newFailed = append(newFailed, idx)
		}
	}
	failureState.FailedMotors = newFailed

	if len(newFailed) == 0 {
		delete(s.failureState, uavID)
	}

	return nil
}

func (s *MotorFailureService) SendManualPIDAdjustment(uavID uint64, params map[string]float64) error {
	cm := mavlink.NewCommandManager()
	encoder := mavlink.NewMAVLinkEncoder()

	for paramIdx, value := range params {
		cmdData := encoder.EncodeCommandLong(
			0, uint8(uavID), 0,
			mavlink.CMD_DO_SET_PARAMETER,
			uint8(paramIdx), 0, 0, 0, 0, 0,
			float32(value),
		)
		if err := cm.SendCommand(uavID, cmdData); err != nil {
			return err
		}
	}
	return nil
}

func (s *MotorFailureService) TriggerManualRTH(uavID uint64) error {
	return s.TriggerEmergencyRTH(uavID, -1)
}

func (s *MotorFailureService) TriggerLand(uavID uint64) error {
	cm := mavlink.NewCommandManager()
	encoder := mavlink.NewMAVLinkEncoder()
	landData := encoder.EncodeCommandLong(
		0, uint8(uavID), 0,
		mavlink.CMD_NAV_LAND,
		0, 0, 0, 0, 0, 0, 0,
	)
	return cm.SendCommand(uavID, landData)
}

func init() {
	_ = binary.LittleEndian.Uint16
}
