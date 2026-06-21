package service

import (
	"fmt"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MotorFailureService struct {
	mu                 sync.RWMutex
	motorStatus        map[uint64]map[int]*models.MotorStatus
	motorHistory       map[uint64]map[int][]int
	motorCurrentHist   map[uint64]map[int][]float64
	failureState       map[uint64]*MotorFailureState
	motorCount         map[uint64]int
	lastFailureTime    map[uint64]map[int]time.Time
	alertService       *AlertService
}

type MotorFailureState struct {
	UAVID            uint64
	FailedMotors     []int
	PIDAdjusted      bool
	RTHTriggered     bool
	MixingRecalc     bool
	StartTime        time.Time
	LastUpdateTime   time.Time
	PIDAdjustments   PIDGain
	MixingMatrix     [][]float64
	OriginalMixing   [][]float64
}

type PIDGain struct {
	RollP, RollI, RollD   float32
	PitchP, PitchI, PitchD float32
	YawP, YawI, YawD      float32
}

const (
	motorTypeM6  = 6
	motorTypeX8  = 8
	motorTypeO8  = 8

	rpmMinThreshold          = 800
	tempCriticalThresholdC   = 110
	tempWarningThresholdC    = 90
	currentMaxThresholdA     = 45
	consecutiveFaultRequired = 3
	cooldownBetweenFailures  = 5 * time.Second
)

var motorFailureService *MotorFailureService
var motorFailureOnce sync.Once

func NewMotorFailureService() *MotorFailureService {
	motorFailureOnce.Do(func() {
		motorFailureService = &MotorFailureService{
			motorStatus:      make(map[uint64]map[int]*models.MotorStatus),
			motorHistory:     make(map[uint64]map[int][]int),
			motorCurrentHist: make(map[uint64]map[int][]float64),
			failureState:     make(map[uint64]*MotorFailureState),
			motorCount:       make(map[uint64]int),
			lastFailureTime:  make(map[uint64]map[int]time.Time),
			alertService:     NewAlertService(),
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
	if _, ok := s.motorCurrentHist[uavID]; !ok {
		s.motorCurrentHist[uavID] = make(map[int][]float64)
	}
	if _, ok := s.lastFailureTime[uavID]; !ok {
		s.lastFailureTime[uavID] = make(map[int]time.Time)
	}

	rpmHistory := s.motorHistory[uavID][status.MotorIndex]
	rpmHistory = append(rpmHistory, status.RPM)
	if len(rpmHistory) > 20 {
		rpmHistory = rpmHistory[len(rpmHistory)-20:]
	}
	s.motorHistory[uavID][status.MotorIndex] = rpmHistory

	curHistory := s.motorCurrentHist[uavID][status.MotorIndex]
	curHistory = append(curHistory, status.Current)
	if len(curHistory) > 20 {
		curHistory = curHistory[len(curHistory)-20:]
	}
	s.motorCurrentHist[uavID][status.MotorIndex] = curHistory

	s.motorCount[uavID] = len(s.motorStatus[uavID])

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

func (s *MotorFailureService) DetectMotorFailure(uavID uint64, motorIndex int, status *models.MotorStatus) (bool, models.MotorStatusType) {
	s.mu.Lock()
	defer s.mu.Unlock()

	motorCount := s.motorCount[uavID]
	if motorCount < 6 {
		return false, ""
	}

	rpmHistory := s.motorHistory[uavID][motorIndex]
	curHistory := s.motorCurrentHist[uavID][motorIndex]

	isFault := false
	isWarning := false

	if status.FaultFlags > 0 {
		isFault = true
	}

	if len(rpmHistory) >= consecutiveFaultRequired && status.Throttle > 5.0 {
		consecutiveLowRPM := 0
		for i := len(rpmHistory) - 1; i >= 0 && consecutiveLowRPM < consecutiveFaultRequired; i-- {
			if rpmHistory[i] < rpmMinThreshold {
				consecutiveLowRPM++
			} else {
				break
			}
		}
		if consecutiveLowRPM >= consecutiveFaultRequired {
			isFault = true
		}
	}

	if status.Temperature > tempCriticalThresholdC {
		isFault = true
	} else if status.Temperature > tempWarningThresholdC {
		isWarning = true
	}

	if len(curHistory) >= 5 {
		avgCurrent := 0.0
		for i := len(curHistory) - 5; i < len(curHistory); i++ {
			avgCurrent += curHistory[i]
		}
		avgCurrent /= 5.0
		if avgCurrent > currentMaxThresholdA {
			isWarning = true
		}
		if status.Voltage > 0 && status.Throttle > 10.0 && status.Current < 0.5 && status.RPM < rpmMinThreshold {
			isFault = true
		}
	}

	var statusOverride models.MotorStatusType
	if isFault {
		statusOverride = models.MotorStatusFault
	} else if isWarning {
		statusOverride = models.MotorStatusWarning
	} else {
		statusOverride = models.MotorStatusNormal
	}

	if !isFault {
		return false, statusOverride
	}

	lastTime, ok := s.lastFailureTime[uavID][motorIndex]
	if ok && time.Since(lastTime) < cooldownBetweenFailures {
		return false, statusOverride
	}
	s.lastFailureTime[uavID][motorIndex] = time.Now()

	failureState, exists := s.failureState[uavID]
	if !exists {
		failureState = &MotorFailureState{
			UAVID:     uavID,
			StartTime: time.Now(),
			OriginalMixing: buildStandardMixingMatrix(motorCount),
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
		middleware.Logger.Error("CRITICAL: Too many motor failures for controlled flight",
			zap.Uint64("uav_id", uavID),
			zap.Int("failed_count", len(failureState.FailedMotors)),
			zap.Int("total_count", motorCount),
		)
	}

	status.Status = statusOverride
	s.motorStatus[uavID][motorIndex] = status

	middleware.Logger.Warn("Motor failure detected",
		zap.Uint64("uav_id", uavID),
		zap.Int("motor_index", motorIndex),
		zap.Int("rpm", status.RPM),
		zap.Int("temperature", status.Temperature),
		zap.Int("fault_flags", status.FaultFlags),
		zap.Float64("current", status.Current),
	)

	return true, statusOverride
}

func buildStandardMixingMatrix(motorCount int) [][]float64 {
	matrix := make([][]float64, motorCount)
	for i := range matrix {
		matrix[i] = make([]float64, 4)
	}

	switch motorCount {
	case 6:
		for i := 0; i < 6; i++ {
			angle := float64(i) * math.Pi / 3.0
			matrix[i][0] = 1.0
			matrix[i][1] = math.Sin(angle)
			matrix[i][2] = -math.Cos(angle)
			if i%2 == 0 {
				matrix[i][3] = -1.0
			} else {
				matrix[i][3] = 1.0
			}
		}
	case 8:
		for i := 0; i < 8; i++ {
			angle := float64(i) * math.Pi / 4.0
			matrix[i][0] = 1.0
			matrix[i][1] = math.Sin(angle)
			matrix[i][2] = -math.Cos(angle)
			if i%2 == 0 {
				matrix[i][3] = -1.0
			} else {
				matrix[i][3] = 1.0
			}
		}
	default:
		for i := 0; i < motorCount; i++ {
			angle := float64(i) * 2.0 * math.Pi / float64(motorCount)
			matrix[i][0] = 1.0
			matrix[i][1] = math.Sin(angle)
			matrix[i][2] = -math.Cos(angle)
			if i%2 == 0 {
				matrix[i][3] = -1.0
			} else {
				matrix[i][3] = 1.0
			}
		}
	}
	return matrix
}

func (s *MotorFailureService) RecalculateMotorMixing(uavID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	failureState, ok := s.failureState[uavID]
	if !ok || failureState.MixingRecalc {
		return
	}
	if failureState.OriginalMixing == nil {
		failureState.OriginalMixing = buildStandardMixingMatrix(s.motorCount[uavID])
	}

	motorCount := s.motorCount[uavID]
	original := failureState.OriginalMixing
	failedSet := make(map[int]bool)
	for _, idx := range failureState.FailedMotors {
		failedSet[idx] = true
	}

	remainingCount := motorCount - len(failureState.FailedMotors)
	remainingIndices := make([]int, 0, remainingCount)
	for i := 0; i < motorCount; i++ {
		if !failedSet[i] {
			remainingIndices = append(remainingIndices, i)
		}
	}

	reducedMatrix := make([][]float64, remainingCount)
	reducedIdx := 0
	for i := 0; i < motorCount; i++ {
		if !failedSet[i] {
			reducedMatrix[reducedIdx] = make([]float64, 4)
			copy(reducedMatrix[reducedIdx], original[i])
			reducedIdx++
		}
	}

	thrustScaleFactor := float64(motorCount) / float64(remainingCount)
	newMatrix := make([][]float64, motorCount)
	for i := 0; i < motorCount; i++ {
		newMatrix[i] = make([]float64, 4)
	}

	for rIdx, origIdx := range remainingIndices {
		thrustBias := 1.0
		if len(failureState.FailedMotors) == 1 {
			failedIdx := failureState.FailedMotors[0]
			failedAngle := 2.0 * math.Pi * float64(failedIdx) / float64(motorCount)
			origAngle := 2.0 * math.Pi * float64(origIdx) / float64(motorCount)
			angleDiff := math.Cos(origAngle - failedAngle)
			thrustBias = 1.0 + 0.15*(1.0-angleDiff)
		}

		for axis := 0; axis < 4; axis++ {
			if axis == 0 {
				newMatrix[origIdx][axis] = reducedMatrix[rIdx][axis] * thrustScaleFactor * thrustBias
			} else {
				newMatrix[origIdx][axis] = reducedMatrix[rIdx][axis] * 1.15
			}
		}
	}

	failureState.MixingMatrix = newMatrix
	failureState.MixingRecalc = true

	s.sendMixingParameters(uavID, newMatrix, motorCount)

	middleware.Logger.Info("Motor mixing matrix recalculated",
		zap.Uint64("uav_id", uavID),
		zap.Int("failed_motors", len(failureState.FailedMotors)),
		zap.Float64("thrust_scale", thrustScaleFactor),
	)
}

func (s *MotorFailureService) sendMixingParameters(uavID uint64, matrix [][]float64, motorCount int) {
	cm := mavlink.NewCommandManager()

	paramName := "MOT_"
	for i := 0; i < motorCount; i++ {
		for axis := 0; axis < 4; axis++ {
			axisName := []string{"THR", "ROL", "PIT", "YAW"}[axis]
			fullName := fmt.Sprintf("%s%d_%s", paramName, i+1, axisName)
			if len(fullName) > 16 {
				fullName = fullName[:16]
			}
			value := float32(matrix[i][axis])
			cmdData := mavlink.EncodeParamSet(fullName, value, 9)
			_ = cm.SendCommand(uavID, cmdData)
			time.Sleep(20 * time.Millisecond)
		}
	}

	mixType := uint16(0)
	switch motorCount {
	case 6:
		mixType = 12
	case 8:
		mixType = 14
	}
	if len(s.failureState[uavID].FailedMotors) > 0 {
		mixType = 100
	}
	cmdData := mavlink.EncodeParamSet("FRAME_CLASS", float32(mixType), 6)
	_ = cm.SendCommand(uavID, cmdData)
}

func (s *MotorFailureService) AdjustPIDParameters(uavID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	failureState, ok := s.failureState[uavID]
	if !ok || failureState.PIDAdjusted {
		return
	}

	motorCount := s.motorCount[uavID]
	failedCount := len(failureState.FailedMotors)
	remainingCount := motorCount - failedCount

	if remainingCount < 4 {
		return
	}

	pFactor := 1.0
	iFactor := 1.0
	dFactor := 1.0

	switch {
	case motorCount >= 10:
		pFactor = 1.08
		iFactor = 1.05
		dFactor = 1.03
	case motorCount >= 8:
		pFactor = 1.12
		iFactor = 1.08
		dFactor = 1.05
	default:
		pFactor = 1.18
		iFactor = 1.12
		dFactor = 1.07
	}

	degradationRatio := float64(failedCount) / float64(motorCount)
	pFactor += degradationRatio * 0.15
	iFactor += degradationRatio * 0.10
	dFactor += degradationRatio * 0.08

	if failedCount >= 2 {
		dFactor += 0.05
	}

	isSymmetric := isFailureSymmetric(failureState.FailedMotors, motorCount)
	if !isSymmetric {
		pFactor += 0.05
		dFactor += 0.10
	}

	failureState.PIDAdjustments = PIDGain{
		RollP:  float32(pFactor),
		RollI:  float32(iFactor),
		RollD:  float32(dFactor),
		PitchP: float32(pFactor),
		PitchI: float32(iFactor),
		PitchD: float32(dFactor),
		YawP:   float32(pFactor * 1.05),
		YawI:   float32(iFactor),
		YawD:   float32(dFactor * 0.95),
	}
	failureState.PIDAdjusted = true

	s.sendPIDParameters(uavID, failureState.PIDAdjustments)

	middleware.Logger.Info("PID parameters adaptively adjusted",
		zap.Uint64("uav_id", uavID),
		zap.Int("motor_count", motorCount),
		zap.Int("failed_count", failedCount),
		zap.Bool("symmetric", isSymmetric),
		zap.Float32("p_factor", failureState.PIDAdjustments.RollP),
		zap.Float32("i_factor", failureState.PIDAdjustments.RollI),
		zap.Float32("d_factor", failureState.PIDAdjustments.RollD),
	)
}

func isFailureSymmetric(failedMotors []int, motorCount int) bool {
	if len(failedMotors) != 1 && len(failedMotors) != 2 {
		return false
	}
	half := motorCount / 2
	for _, f := range failedMotors {
		opposite := (f + half) % motorCount
		found := false
		for _, f2 := range failedMotors {
			if f2 == opposite {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (s *MotorFailureService) sendPIDParameters(uavID uint64, gains PIDGain) {
	cm := mavlink.NewCommandManager()

	params := []struct {
		name  string
		value float32
	}{
		{"ATC_RAT_RLL_P", gains.RollP},
		{"ATC_RAT_RLL_I", gains.RollI},
		{"ATC_RAT_RLL_D", gains.RollD},
		{"ATC_RAT_PIT_P", gains.PitchP},
		{"ATC_RAT_PIT_I", gains.PitchI},
		{"ATC_RAT_PIT_D", gains.PitchD},
		{"ATC_RAT_YAW_P", gains.YawP},
		{"ATC_RAT_YAW_I", gains.YawI},
		{"ATC_RAT_YAW_D", gains.YawD},
		{"ATC_ANG_RLL_P", gains.RollP * 0.9},
		{"ATC_ANG_PIT_P", gains.PitchP * 0.9},
	}

	for _, p := range params {
		cmdData := mavlink.EncodeParamSet(p.name, p.value, 9)
		if err := cm.SendCommand(uavID, cmdData); err != nil {
			middleware.Logger.Warn("Failed to send PID param",
				zap.Uint64("uav_id", uavID),
				zap.String("param", p.name),
				zap.Error(err),
			)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func (s *MotorFailureService) TriggerEmergencyRTH(uavID uint64, motorIndex int) error {
	s.mu.Lock()
	failureState, exists := s.failureState[uavID]
	if exists && failureState.RTHTriggered {
		s.mu.Unlock()
		return nil
	}
	if exists {
		failureState.RTHTriggered = true
	}
	s.mu.Unlock()

	cm := mavlink.NewCommandManager()

	rthData := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
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
	failedMotorLabel := fmt.Sprintf("电机 #%d", motorIndex+1)
	title := fmt.Sprintf("%s 失效告警", failedMotorLabel)
	message := fmt.Sprintf("无人机 %d 的 %s 检测到故障。故障标志: 0x%04X, RPM: %d, 温度: %d°C, 电流: %.1fA。飞控已重新分配电机混控并自动调整PID参数，正在执行紧急返航。",
		uavID, failedMotorLabel, status.FaultFlags, status.RPM, status.Temperature, status.Current)

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
			ActionTaken:   "mixing_recalc_pid_adjust_rth",
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
		delete(s.lastFailureTime, uavID)
	} else {
		failureState.MixingRecalc = false
		failureState.PIDAdjusted = false
	}

	return nil
}

func (s *MotorFailureService) SendManualPIDAdjustment(uavID uint64, params map[string]float64) error {
	cm := mavlink.NewCommandManager()

	for paramName, value := range params {
		cmdData := mavlink.EncodeParamSet(paramName, float32(value), 9)
		if err := cm.SendCommand(uavID, cmdData); err != nil {
			return err
		}
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

func (s *MotorFailureService) TriggerManualRTH(uavID uint64) error {
	cm := mavlink.NewCommandManager()
	rthData := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
	return cm.SendCommand(uavID, rthData)
}

func (s *MotorFailureService) TriggerLand(uavID uint64) error {
	cm := mavlink.NewCommandManager()
	landData := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_LAND, 0, 0, 0, 0, 0, 0, 0)
	return cm.SendCommand(uavID, landData)
}
