package service

import (
	"context"
	"encoding/json"
	"errors"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type FlightService struct {
	flightRepo  *repository.FlightRepository
	uavRepo     *repository.UAVRepository
	missionRepo *repository.MissionRepository
	alertRepo   *repository.AlertRepository
	redis       *redis.Client
}

func NewFlightService() *FlightService {
	return &FlightService{
		flightRepo:  repository.NewFlightRepository(),
		uavRepo:     repository.NewUAVRepository(),
		missionRepo: repository.NewMissionRepository(),
		alertRepo:   repository.NewAlertRepository(),
		redis:       config.Redis,
	}
}

type TelemetryData struct {
	UAVID          uint64    `json:"uav_id"`
	Timestamp      time.Time `json:"timestamp"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	AltitudeRel    float64   `json:"altitude_rel"`
	AltitudeMSL    float64   `json:"altitude_msl"`
	GroundSpeed    float64   `json:"ground_speed"`
	AirSpeed       float64   `json:"air_speed"`
	Heading        float64   `json:"heading"`
	Pitch          float64   `json:"pitch"`
	Roll           float64   `json:"roll"`
	Yaw            float64   `json:"yaw"`
	BatteryLevel   float64   `json:"battery_level"`
	BatteryVoltage float64   `json:"battery_voltage"`
	SignalStrength int       `json:"signal_strength"`
	Satellites     int       `json:"satellites"`
	GPSFixType     int       `json:"gps_fix_type"`
	Mode           string    `json:"mode"`
	ArmStatus      bool      `json:"arm_status"`
	FlightTime     int       `json:"flight_time"`
}

func (s *FlightService) ProcessTelemetry(data *TelemetryData) error {
	status := &models.FlightStatus{
		UAVID:          data.UAVID,
		Timestamp:      data.Timestamp,
		Latitude:       data.Latitude,
		Longitude:      data.Longitude,
		AltitudeMSL:    data.AltitudeMSL,
		AltitudeRel:    data.AltitudeRel,
		GroundSpeed:    data.GroundSpeed,
		AirSpeed:       data.AirSpeed,
		Heading:        data.Heading,
		Pitch:          data.Pitch,
		Roll:           data.Roll,
		Yaw:            data.Yaw,
		BatteryVoltage: data.BatteryVoltage,
		BatteryLevel:   data.BatteryLevel,
		SignalStrength: data.SignalStrength,
		Satellites:     data.Satellites,
		GPSFixType:     data.GPSFixType,
		Mode:           data.Mode,
		ArmStatus:      data.ArmStatus,
		FlightTime:     data.FlightTime,
	}

	if err := s.flightRepo.CreateFlightStatus(status); err != nil {
		return err
	}

	_ = s.uavRepo.UpdateLastSeen(data.UAVID)

	cacheKey := "uav:telemetry:" + utils.Uint64ToString(data.UAVID)
	cacheData, _ := json.Marshal(data)
	_ = s.redis.Set(ctx, cacheKey, cacheData, time.Minute).Err()

	go s.checkAlerts(data)

	return nil
}

func (s *FlightService) checkAlerts(data *TelemetryData) {
	threshold := config.AppConfig.Alert.LowBatteryThreshold
	if data.BatteryLevel > 0 && data.BatteryLevel <= threshold {
		alert := &models.AlertEvent{
			UAVID:        data.UAVID,
			Type:         models.AlertTypeLowBattery,
			Level:        models.AlertLevelWarning,
			Title:        "低电量告警",
			Message:      "无人机电量低于 " + utils.Float64ToString(threshold) + "%",
			Latitude:     data.Latitude,
			Longitude:    data.Longitude,
			Altitude:     data.AltitudeRel,
			BatteryLevel: data.BatteryLevel,
			SignalStrength: data.SignalStrength,
		}
		_ = s.alertRepo.Create(alert)
	}

	signalThreshold := config.AppConfig.Alert.SignalLossThreshold
	if data.SignalStrength > 0 && data.SignalStrength < signalThreshold {
		alert := &models.AlertEvent{
			UAVID:        data.UAVID,
			Type:         models.AlertTypeSignalLoss,
			Level:        models.AlertLevelCritical,
			Title:        "信号丢失告警",
			Message:      "无人机信号强度过低",
			Latitude:     data.Latitude,
			Longitude:    data.Longitude,
			Altitude:     data.AltitudeRel,
			BatteryLevel: data.BatteryLevel,
			SignalStrength: data.SignalStrength,
		}
		_ = s.alertRepo.Create(alert)
	}

	if data.GPSFixType < 3 {
		alert := &models.AlertEvent{
			UAVID:        data.UAVID,
			Type:         models.AlertTypeGPSLost,
			Level:        models.AlertLevelWarning,
			Title:        "GPS信号丢失",
			Message:      "无人机GPS定位精度不足",
			Latitude:     data.Latitude,
			Longitude:    data.Longitude,
			Altitude:     data.AltitudeRel,
			BatteryLevel: data.BatteryLevel,
			SignalStrength: data.SignalStrength,
		}
		_ = s.alertRepo.Create(alert)
	}
}

func (s *FlightService) GetLatestStatus(uavID uint64) (*models.FlightStatus, error) {
	cacheKey := "uav:telemetry:" + utils.Uint64ToString(uavID)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var data TelemetryData
		if json.Unmarshal([]byte(cached), &data) == nil {
			return &models.FlightStatus{
				UAVID:          data.UAVID,
				Timestamp:      data.Timestamp,
				Latitude:       data.Latitude,
				Longitude:      data.Longitude,
				AltitudeMSL:    data.AltitudeMSL,
				AltitudeRel:    data.AltitudeRel,
				GroundSpeed:    data.GroundSpeed,
				AirSpeed:       data.AirSpeed,
				Heading:        data.Heading,
				Pitch:          data.Pitch,
				Roll:           data.Roll,
				Yaw:            data.Yaw,
				BatteryVoltage: data.BatteryVoltage,
				BatteryLevel:   data.BatteryLevel,
				SignalStrength: data.SignalStrength,
				Satellites:     data.Satellites,
				GPSFixType:     data.GPSFixType,
				Mode:           data.Mode,
				ArmStatus:      data.ArmStatus,
				FlightTime:     data.FlightTime,
			}, nil
		}
	}
	return s.flightRepo.GetLatestStatus(uavID)
}

func (s *FlightService) GetStatusHistory(uavID uint64, pagination *utils.Pagination, startTime, endTime *time.Time) ([]models.FlightStatus, int64, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, 0, errors.New("uav not found")
	}
	return s.flightRepo.GetStatusHistory(uavID, pagination, startTime, endTime)
}

func (s *FlightService) GetRealtimeData(uavIDs []uint64) (map[uint64]*models.FlightStatus, error) {
	result := make(map[uint64]*models.FlightStatus)
	for _, uavID := range uavIDs {
		status, err := s.GetLatestStatus(uavID)
		if err == nil {
			result[uavID] = status
		}
	}
	return result, nil
}

func (s *FlightService) GetAllRealtimeData() (map[uint64]*models.FlightStatus, error) {
	uavs, err := s.uavRepo.GetOnlineUAVs()
	if err != nil {
		return nil, err
	}
	uavIDs := make([]uint64, len(uavs))
	for i, uav := range uavs {
		uavIDs[i] = uav.ID
	}
	return s.GetRealtimeData(uavIDs)
}

func (s *FlightService) GetFlightSummary(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}
	return s.flightRepo.GetFlightSummary(uavID, startTime, endTime)
}

func (s *FlightService) CleanOldData(beforeDays int) (int64, error) {
	return s.flightRepo.CleanOldData(beforeDays)
}
