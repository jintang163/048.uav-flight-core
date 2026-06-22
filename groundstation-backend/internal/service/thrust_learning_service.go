package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
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

func (s *ThrustLearningService) StoreSample(uavID uint64, sample *models.ThrustLearningSample) error {
	sample.UAVID = uavID
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

	points, err := s.repo.OptimizeThrustCurve(uavID)
	if err != nil {
		status.State = "data_collecting"
		status.UpdatedAt = now
		_ = s.repo.UpsertStatus(status)
		return nil, err
	}

	if len(points) > 0 {
		hoverThrottle := 0.5
		estimatedWeight := 0.0
		for _, p := range points {
			if p.Throttle >= 0.45 && p.Throttle <= 0.55 {
				hoverThrottle = p.Throttle
				estimatedWeight = p.ThrustN / 9.81
				break
			}
		}

		status.HoverThrottle = hoverThrottle
		status.EstimatedWeight = estimatedWeight
	}

	status.State = "applied"
	status.CompletedAt = &now
	status.UpdatedAt = now
	if err := s.repo.UpsertStatus(status); err != nil {
		return nil, err
	}

	pidProfile, _ := s.repo.GetPIDProfile(uavID)
	if pidProfile != nil {
		pidProfile.IsAutoTuned = true
		pidProfile.UpdatedAt = now
		_ = s.repo.UpsertPIDProfile(pidProfile)
	}

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
