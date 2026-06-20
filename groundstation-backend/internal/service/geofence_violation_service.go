package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
)

type GeofenceViolationService struct {
	violationRepo *repository.GeofenceViolationRepository
	geofenceRepo  *repository.GeofenceRepository
	uavRepo       *repository.UAVRepository
}

func NewGeofenceViolationService() *GeofenceViolationService {
	return &GeofenceViolationService{
		violationRepo: repository.NewGeofenceViolationRepository(),
		geofenceRepo:  repository.NewGeofenceRepository(),
		uavRepo:       repository.NewUAVRepository(),
	}
}

func (s *GeofenceViolationService) LogViolation(uavID uint64, geofenceID uint64, violationType models.ViolationType, severity models.ViolationSeverity, lat, lng, alt, distance float64, action models.FailAction) (*models.GeofenceViolationLog, error) {
	gf, err := s.geofenceRepo.FindByID(geofenceID)
	if err != nil {
		return nil, err
	}

	log := &models.GeofenceViolationLog{
		UAVID:            uavID,
		GeofenceID:       geofenceID,
		GeofenceName:     gf.Name,
		GeofenceCategory: gf.Category,
		ViolationType:    violationType,
		Severity:         severity,
		Latitude:         lat,
		Longitude:        lng,
		Altitude:         alt,
		Distance:         distance,
		ActionTaken:      action,
		IsResolved:       false,
	}

	if err := s.violationRepo.Create(log); err != nil {
		return nil, err
	}

	websocket.BroadcastGeofenceViolation(uavID, geofenceID, string(violationType), string(severity), lat, lng, alt, distance)

	return log, nil
}

func (s *GeofenceViolationService) GetByID(id uint64) (*models.GeofenceViolationLog, error) {
	return s.violationRepo.FindByID(id)
}

func (s *GeofenceViolationService) List(pagination *utils.Pagination, uavID uint64, geofenceID uint64, severity string, violationType string, isResolved *bool, startTime string, endTime string) ([]models.GeofenceViolationLog, int64, error) {
	return s.violationRepo.List(pagination, uavID, geofenceID, severity, violationType, isResolved, startTime, endTime)
}

func (s *GeofenceViolationService) Resolve(id uint64, notes string) error {
	_, err := s.violationRepo.FindByID(id)
	if err != nil {
		return errors.New("violation log not found")
	}
	return s.violationRepo.MarkResolved(id, notes)
}

func (s *GeofenceViolationService) GetStatistics(uavID uint64, geofenceID uint64, startTime string, endTime string) (map[string]interface{}, error) {
	return s.violationRepo.GetStatistics(uavID, geofenceID, startTime, endTime)
}

func (s *GeofenceViolationService) CleanOld(days int) (int64, error) {
	if days < 1 {
		days = 30
	}
	return s.violationRepo.CleanOldData(days)
}
