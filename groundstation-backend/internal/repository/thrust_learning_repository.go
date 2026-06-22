package repository

import (
	"errors"
	"groundstation-backend/internal/models"
	"time"
)

type ThrustLearningRepository struct {
	*BaseRepository
}

func NewThrustLearningRepository() *ThrustLearningRepository {
	return &ThrustLearningRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *ThrustLearningRepository) GetLatestStatus(uavID uint64) (*models.ThrustLearningStatus, error) {
	var status models.ThrustLearningStatus
	err := r.db.Where("uav_id = ?", uavID).Order("created_at DESC").First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *ThrustLearningRepository) UpsertStatus(status *models.ThrustLearningStatus) error {
	existing, err := r.GetLatestStatus(status.UAVID)
	if err != nil {
		return r.db.Create(status).Error
	}

	status.ID = existing.ID
	status.CreatedAt = existing.CreatedAt
	return r.db.Save(status).Error
}

func (r *ThrustLearningRepository) ListCurvePoints(uavID uint64) ([]models.ThrustCurvePoint, error) {
	var points []models.ThrustCurvePoint
	err := r.db.Where("uav_id = ?", uavID).Order("throttle ASC").Find(&points).Error
	return points, err
}

func (r *ThrustLearningRepository) UpsertCurvePoint(uavID uint64, throttle, thrust, rpm float64) error {
	var point models.ThrustCurvePoint
	err := r.db.Where("uav_id = ? AND throttle = ?", uavID, throttle).First(&point).Error

	if err != nil {
		newPoint := models.ThrustCurvePoint{
			UAVID:       uavID,
			Throttle:    throttle,
			ThrustN:     thrust,
			MotorRpmAvg: rpm,
			SampleCount: 1,
		}
		return r.db.Create(&newPoint).Error
	}

	point.ThrustN = thrust
	point.MotorRpmAvg = rpm
	point.SampleCount++
	return r.db.Save(&point).Error
}

func (r *ThrustLearningRepository) GetPIDProfile(uavID uint64) (*models.PIDGainProfile, error) {
	var profile models.PIDGainProfile
	err := r.db.Where("uav_id = ?", uavID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ThrustLearningRepository) UpsertPIDProfile(profile *models.PIDGainProfile) error {
	existing, err := r.GetPIDProfile(profile.UAVID)
	if err != nil {
		return r.db.Create(profile).Error
	}

	profile.ID = existing.ID
	profile.CreatedAt = existing.CreatedAt
	return r.db.Save(profile).Error
}

func (r *ThrustLearningRepository) AddSample(sample *models.ThrustLearningSample) error {
	return r.db.Create(sample).Error
}

func (r *ThrustLearningRepository) ListSamples(uavID uint64, limit int) ([]models.ThrustLearningSample, error) {
	var samples []models.ThrustLearningSample
	if limit <= 0 {
		limit = 1000
	}
	err := r.db.Where("uav_id = ?", uavID).Order("timestamp DESC").Limit(limit).Find(&samples).Error
	return samples, err
}

func (r *ThrustLearningRepository) OptimizeThrustCurve(uavID uint64) ([]models.ThrustCurvePoint, error) {
	samples, err := r.ListSamples(uavID, 5000)
	if err != nil {
		return nil, err
	}
	if len(samples) < 10 {
		return nil, errors.New("insufficient samples for optimization")
	}

	type aggregatedPoint struct {
		Throttle    float64
		ThrustSum   float64
		RpmSum      float64
		SampleCount int
	}

	bucketMap := make(map[float64]*aggregatedPoint)
	const bucketSize = 0.02

	for _, sample := range samples {
		if sample.AccelZ < 0.5 || sample.VZ > 0.5 || sample.VZ < -0.5 {
			continue
		}

		bucketThrottle := float64(int(sample.Throttle/bucketSize)) * bucketSize
		bucketThrottle = float64(int(bucketThrottle*1000)) / 1000

		thrust := sample.AccelZ * 9.81

		if bucket, exists := bucketMap[bucketThrottle]; exists {
			bucket.ThrustSum += thrust
			bucket.RpmSum += float64(sample.MotorPWM1+sample.MotorPWM2+sample.MotorPWM3+sample.MotorPWM4) / 4.0
			bucket.SampleCount++
		} else {
			bucketMap[bucketThrottle] = &aggregatedPoint{
				Throttle:    bucketThrottle,
				ThrustSum:   thrust,
				RpmSum:      float64(sample.MotorPWM1+sample.MotorPWM2+sample.MotorPWM3+sample.MotorPWM4) / 4.0,
				SampleCount: 1,
			}
		}
	}

	if len(bucketMap) < 3 {
		return nil, errors.New("insufficient valid samples for optimization")
	}

	optimizedPoints := make([]models.ThrustCurvePoint, 0, len(bucketMap))
	now := time.Now()

	for _, bucket := range bucketMap {
		if bucket.SampleCount < 3 {
			continue
		}
		optimizedPoints = append(optimizedPoints, models.ThrustCurvePoint{
			UAVID:       uavID,
			Throttle:    bucket.Throttle,
			ThrustN:     bucket.ThrustSum / float64(bucket.SampleCount),
			MotorRpmAvg: bucket.RpmSum / float64(bucket.SampleCount),
			SampleCount: bucket.SampleCount,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	if len(optimizedPoints) < 3 {
		return nil, errors.New("insufficient aggregated points for optimization")
	}

	err = r.db.Where("uav_id = ?", uavID).Delete(&models.ThrustCurvePoint{}).Error
	if err != nil {
		return nil, err
	}

	for i := range optimizedPoints {
		if err := r.db.Create(&optimizedPoints[i]).Error; err != nil {
			return nil, err
		}
	}

	return optimizedPoints, nil
}
