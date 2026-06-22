package service

import (
	"errors"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"sort"
	"time"
)

type ThrustLearningService struct {
	repo *repository.ThrustLearningRepository
}

func NewThrustLearningService() *ThrustLearningService {
	return &ThrustLearningService{
		repo: repository.NewThrustLearningRepository(),
	}
}

func (s *ThrustLearningService) GetStatus(uavID uint64) (*models.ThrustLearningStatus, error) {
	status, err := s.repo.GetLatestStatus(uavID)
	if err != nil {
		now := time.Now()
		status = &models.ThrustLearningStatus{
			UAVID:           uavID,
			State:           "idle",
			EstimatedWeight: 0,
			HoverThrottle:   0.5,
			SampleCount:     0,
			Progress:        0,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if createErr := s.repo.UpsertStatus(status); createErr != nil {
			return nil, createErr
		}
	}
	return status, nil
}

func (s *ThrustLearningService) UpdateStatusFromMAVLink(uavID uint64, state string, weight, hoverThrottle float64, samples uint32, progress float64) error {
	status, err := s.repo.GetLatestStatus(uavID)
	now := time.Now()

	if err != nil {
		status = &models.ThrustLearningStatus{
			UAVID:           uavID,
			State:           state,
			EstimatedWeight: weight,
			HoverThrottle:   hoverThrottle,
			SampleCount:     samples,
			Progress:        progress,
			StartedAt:       &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	} else {
		status.State = state
		status.EstimatedWeight = weight
		status.HoverThrottle = hoverThrottle
		status.SampleCount = samples
		status.Progress = progress
		status.UpdatedAt = now

		if state != "idle" && status.StartedAt == nil {
			status.StartedAt = &now
		}
		if state == "applied" && status.CompletedAt == nil {
			status.CompletedAt = &now
		}
	}

	return s.repo.UpsertStatus(status)
}

func (s *ThrustLearningService) StoreCurvePoints(uavID uint64, points []struct {
	Throttle float64
	Thrust   float64
	Rpm      float64
}) error {
	for _, p := range points {
		if err := s.repo.UpsertCurvePoint(uavID, p.Throttle, p.Thrust, p.Rpm); err != nil {
			return err
		}
	}
	return nil
}

func (s *ThrustLearningService) StorePIDGains(uavID uint64, gains map[string]float64) error {
	profile, err := s.repo.GetPIDProfile(uavID)
	now := time.Now()

	if err != nil {
		profile = &models.PIDGainProfile{
			UAVID:     uavID,
			CreatedAt: now,
			UpdatedAt: now,
		}
	} else {
		profile.UpdatedAt = now
	}

	if v, ok := gains["roll_kp"]; ok {
		profile.RollKP = v
	}
	if v, ok := gains["roll_ki"]; ok {
		profile.RollKI = v
	}
	if v, ok := gains["roll_kd"]; ok {
		profile.RollKD = v
	}
	if v, ok := gains["pitch_kp"]; ok {
		profile.PitchKP = v
	}
	if v, ok := gains["pitch_ki"]; ok {
		profile.PitchKI = v
	}
	if v, ok := gains["pitch_kd"]; ok {
		profile.PitchKD = v
	}
	if v, ok := gains["yaw_kp"]; ok {
		profile.YawKP = v
	}
	if v, ok := gains["yaw_ki"]; ok {
		profile.YawKI = v
	}
	if v, ok := gains["yaw_kd"]; ok {
		profile.YawKD = v
	}

	if v, ok := gains["rate_roll_kp"]; ok {
		profile.RateRollKP = v
	}
	if v, ok := gains["rate_roll_ki"]; ok {
		profile.RateRollKI = v
	}
	if v, ok := gains["rate_roll_kd"]; ok {
		profile.RateRollKD = v
	}
	if v, ok := gains["rate_pitch_kp"]; ok {
		profile.RatePitchKP = v
	}
	if v, ok := gains["rate_pitch_ki"]; ok {
		profile.RatePitchKI = v
	}
	if v, ok := gains["rate_pitch_kd"]; ok {
		profile.RatePitchKD = v
	}
	if v, ok := gains["rate_yaw_kp"]; ok {
		profile.RateYawKP = v
	}
	if v, ok := gains["rate_yaw_ki"]; ok {
		profile.RateYawKI = v
	}
	if v, ok := gains["rate_yaw_kd"]; ok {
		profile.RateYawKD = v
	}

	if v, ok := gains["alt_kp"]; ok {
		profile.AltKP = v
	}
	if v, ok := gains["alt_ki"]; ok {
		profile.AltKI = v
	}
	if v, ok := gains["alt_kd"]; ok {
		profile.AltKD = v
	}

	return s.repo.UpsertPIDProfile(profile)
}

func (s *ThrustLearningService) StoreSample(uavID uint64, sampleData interface{}) error {
	var sample *models.ThrustLearningSample

	switch v := sampleData.(type) {
	case *mavlink.ThrustSampleData:
		sample = &models.ThrustLearningSample{
			UAVID:     uavID,
			Throttle:  float64(v.Throttle),
			AccelZ:    float64(v.AccelZ),
			Altitude:  float64(v.Altitude),
			VZ:        float64(v.VZ),
			MotorPWM1: v.MotorPWM1,
			MotorPWM2: v.MotorPWM2,
			MotorPWM3: v.MotorPWM3,
			MotorPWM4: v.MotorPWM4,
			Voltage:   float64(v.Voltage),
			Timestamp: int64(v.TimestampMs),
		}
	case *models.ThrustLearningSample:
		sample = v
		sample.UAVID = uavID
	case map[string]interface{}:
		sample = &models.ThrustLearningSample{
			UAVID: uavID,
		}
		if val, ok := v["throttle"].(float64); ok {
			sample.Throttle = val
		}
		if val, ok := v["accel_z"].(float64); ok {
			sample.AccelZ = val
		}
		if val, ok := v["altitude"].(float64); ok {
			sample.Altitude = val
		}
		if val, ok := v["vz"].(float64); ok {
			sample.VZ = val
		}
		if val, ok := v["motor_pwm_1"].(float64); ok {
			sample.MotorPWM1 = uint16(val)
		}
		if val, ok := v["motor_pwm_2"].(float64); ok {
			sample.MotorPWM2 = uint16(val)
		}
		if val, ok := v["motor_pwm_3"].(float64); ok {
			sample.MotorPWM3 = uint16(val)
		}
		if val, ok := v["motor_pwm_4"].(float64); ok {
			sample.MotorPWM4 = uint16(val)
		}
		if val, ok := v["voltage"].(float64); ok {
			sample.Voltage = val
		}
		if val, ok := v["timestamp"].(float64); ok {
			sample.Timestamp = int64(val)
		}
	default:
		return errors.New("unsupported sample data type")
	}

	if sample.Timestamp == 0 {
		sample.Timestamp = time.Now().UnixNano() / 1e6
	}

	return s.repo.AddSample(sample)
}

func (s *ThrustLearningService) TriggerLearning(uavID uint64) error {
	now := time.Now()
	status := &models.ThrustLearningStatus{
		UAVID:           uavID,
		State:           "weight_estimation",
		EstimatedWeight: 0,
		HoverThrottle:   0,
		SampleCount:     0,
		Progress:        0,
		StartedAt:       &now,
		CompletedAt:     nil,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return s.repo.UpsertStatus(status)
}

func (s *ThrustLearningService) OptimizeModel(uavID uint64) ([]models.ThrustCurvePoint, error) {
	status, err := s.repo.GetLatestStatus(uavID)
	if err != nil {
		return nil, errors.New("no learning status found")
	}

	now := time.Now()
	status.State = "model_optimizing"
	status.UpdatedAt = now
	if err := s.repo.UpsertStatus(status); err != nil {
		return nil, err
	}

	weightKG := status.EstimatedWeight
	if weightKG <= 0 {
		weightKG = 2.0
	}

	samples, err := s.repo.GetRecentSamples(uavID, 5000)
	if err != nil {
		status.State = "data_collecting"
		status.UpdatedAt = now
		_ = s.repo.UpsertStatus(status)
		return nil, err
	}
	if len(samples) < 10 {
		status.State = "data_collecting"
		status.UpdatedAt = now
		_ = s.repo.UpsertStatus(status)
		return nil, errors.New("insufficient samples for optimization")
	}

	const numBuckets = 16
	type bucket struct {
		throttleMin float64
		throttleMax float64
		thrustSum   float64
		rpmSum      float64
		count       int
	}

	buckets := make([]bucket, numBuckets)
	step := 1.0 / float64(numBuckets)
	for i := 0; i < numBuckets; i++ {
		buckets[i].throttleMin = float64(i) * step
		buckets[i].throttleMax = float64(i+1) * step
	}

	for _, sample := range samples {
		if sample.VZ > 0.5 || sample.VZ < -0.5 {
			continue
		}

		thrustN := weightKG * 9.81 * sample.AccelZ

		bucketIdx := int(sample.Throttle / step)
		if bucketIdx >= numBuckets {
			bucketIdx = numBuckets - 1
		}
		if bucketIdx < 0 {
			bucketIdx = 0
		}

		buckets[bucketIdx].thrustSum += thrustN
		buckets[bucketIdx].rpmSum += float64(sample.MotorPWM1+sample.MotorPWM2+sample.MotorPWM3+sample.MotorPWM4) / 4.0
		buckets[bucketIdx].count++
	}

	var validBuckets []bucket
	for _, b := range buckets {
		if b.count >= 3 {
			validBuckets = append(validBuckets, b)
		}
	}
	if len(validBuckets) < 3 {
		status.State = "data_collecting"
		status.UpdatedAt = now
		_ = s.repo.UpsertStatus(status)
		return nil, errors.New("insufficient valid samples for optimization")
	}

	points := make([]models.ThrustCurvePoint, 0, len(validBuckets))
	err = s.repo.ClearCurvePoints(uavID)
	if err != nil {
		status.State = "data_collecting"
		status.UpdatedAt = now
		_ = s.repo.UpsertStatus(status)
		return nil, err
	}

	for _, b := range validBuckets {
		throttleMid := (b.throttleMin + b.throttleMax) / 2.0
		avgThrust := b.thrustSum / float64(b.count)
		avgRpm := b.rpmSum / float64(b.count)

		point := models.ThrustCurvePoint{
			UAVID:       uavID,
			Throttle:    throttleMid,
			ThrustN:     avgThrust,
			MotorRpmAvg: avgRpm,
			SampleCount: b.count,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err = s.repo.UpsertCurvePoint(uavID, throttleMid, avgThrust, avgRpm)
		if err != nil {
			status.State = "data_collecting"
			status.UpdatedAt = now
			_ = s.repo.UpsertStatus(status)
			return nil, err
		}
		points = append(points, point)
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Throttle < points[j].Throttle
	})

	if len(points) > 0 {
		hoverThrottle := 0.5
		for _, p := range points {
			if p.Throttle >= 0.45 && p.Throttle <= 0.55 {
				hoverThrottle = p.Throttle
				weightKG = p.ThrustN / 9.81
				break
			}
		}
		status.HoverThrottle = hoverThrottle
		status.EstimatedWeight = weightKG
	}

	status.State = "applied"
	status.CompletedAt = &now
	status.UpdatedAt = now
	if err := s.repo.UpsertStatus(status); err != nil {
		return nil, err
	}

	pidProfile, _ := s.repo.GetPIDProfile(uavID)
	if pidProfile == nil {
		pidProfile = &models.PIDGainProfile{
			UAVID:       uavID,
			ProfileName: "default",
			CreatedAt:   now,
		}
	}

	defaultWeight := 2.0
	scaleFactor := weightKG / defaultWeight
	if scaleFactor < 0.5 {
		scaleFactor = 0.5
	}
	if scaleFactor > 2.0 {
		scaleFactor = 2.0
	}

	if pidProfile.RollKP == 0 {
		pidProfile.RollKP = 4.5
		pidProfile.RollKI = 0.0
		pidProfile.RollKD = 0.0
		pidProfile.PitchKP = 4.5
		pidProfile.PitchKI = 0.0
		pidProfile.PitchKD = 0.0
		pidProfile.YawKP = 3.5
		pidProfile.YawKI = 0.0
		pidProfile.YawKD = 0.0
		pidProfile.RateRollKP = 0.15
		pidProfile.RateRollKI = 0.10
		pidProfile.RateRollKD = 0.003
		pidProfile.RatePitchKP = 0.15
		pidProfile.RatePitchKI = 0.10
		pidProfile.RatePitchKD = 0.003
		pidProfile.RateYawKP = 0.20
		pidProfile.RateYawKI = 0.10
		pidProfile.RateYawKD = 0.003
		pidProfile.AltKP = 1.0
		pidProfile.AltKI = 0.0
		pidProfile.AltKD = 0.0
	}

	pidProfile.RollKP *= scaleFactor
	pidProfile.PitchKP *= scaleFactor
	pidProfile.YawKP *= scaleFactor
	pidProfile.RateRollKP *= scaleFactor
	pidProfile.RatePitchKP *= scaleFactor
	pidProfile.RateYawKP *= scaleFactor
	pidProfile.AltKP *= scaleFactor

	pidProfile.IsAutoTuned = true
	pidProfile.UpdatedAt = now
	_ = s.repo.UpsertPIDProfile(pidProfile)

	websocket.BroadcastThrustCurveUpdate(uavID, points)
	websocket.BroadcastPIDGainsUpdate(uavID, pidProfile)

	return points, nil
}

func (s *ThrustLearningService) ApplyAutoTunedPID(uavID uint64) error {
	profile, err := s.repo.GetPIDProfile(uavID)
	if err != nil {
		return errors.New("no PID profile found")
	}

	profile.IsAutoTuned = true
	profile.UpdatedAt = time.Now()
	return s.repo.UpsertPIDProfile(profile)
}

func (s *ThrustLearningService) GetThrustCurve(uavID uint64) ([]models.ThrustCurvePoint, error) {
	return s.repo.ListCurvePoints(uavID)
}

func (s *ThrustLearningService) GetPIDGains(uavID uint64) (*models.PIDGainProfile, error) {
	profile, err := s.repo.GetPIDProfile(uavID)
	if err != nil {
		now := time.Now()
		profile = &models.PIDGainProfile{
			UAVID:       uavID,
			ProfileName: "default",
			IsAutoTuned: false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if createErr := s.repo.UpsertPIDProfile(profile); createErr != nil {
			return nil, createErr
		}
	}
	return profile, nil
}

func (s *ThrustLearningService) UpdatePIDGains(uavID uint64, gains map[string]float64) (*models.PIDGainProfile, error) {
	if err := s.StorePIDGains(uavID, gains); err != nil {
		return nil, err
	}
	return s.repo.GetPIDProfile(uavID)
}

func (s *ThrustLearningService) GetSamples(uavID uint64, limit int) ([]models.ThrustLearningSample, error) {
	return s.repo.ListSamples(uavID, limit)
}
