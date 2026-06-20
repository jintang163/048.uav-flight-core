package service

import (
	"encoding/json"
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"math"
)

type GeofenceService struct {
	geofenceRepo *repository.GeofenceRepository
	alertRepo     *repository.AlertRepository
}

func NewGeofenceService() *GeofenceService {
	return &GeofenceService{
		geofenceRepo: repository.NewGeofenceRepository(),
		alertRepo:     repository.NewAlertRepository(),
	}
}

type CreateGeofenceRequest struct {
	Name        string               `json:"name" binding:"required"`
	Description string               `json:"description"`
	Type        models.GeofenceType `json:"type" binding:"required"`
	Shape       models.GeofenceShape `json:"shape" binding:"required"`
	MaxAltitude float64              `json:"max_altitude"`
	MinAltitude float64              `json:"min_altitude"`
	CenterLat   float64              `json:"center_lat"`
	CenterLng   float64              `json:"center_lng"`
	Radius      float64              `json:"radius"`
	Coordinates []models.Coordinate  `json:"coordinates"`
	IsActive    bool                 `json:"is_active"`
}

type GeofenceViolation struct {
	GeofenceID uint64  `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	ViolationType string `json:"violation_type"`
	Distance    float64 `json:"distance"`
}

func (s *GeofenceService) Create(req *CreateGeofenceRequest, creatorID uint64) (*models.Geofence, error) {
	coordsJSON, _ := json.Marshal(req.Coordinates)

	geofence := &models.Geofence{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Shape:       req.Shape,
		CreatorID:   creatorID,
		IsActive:    req.IsActive,
		MaxAltitude: req.MaxAltitude,
		MinAltitude: req.MinAltitude,
		CenterLat:   req.CenterLat,
		CenterLng:   req.CenterLng,
		Radius:      req.Radius,
		Coordinates: string(coordsJSON),
	}

	if err := s.geofenceRepo.Create(geofence); err != nil {
		return nil, err
	}

	return geofence, nil
}

func (s *GeofenceService) GetByID(id uint64) (*models.Geofence, error) {
	return s.geofenceRepo.FindByID(id)
}

func (s *GeofenceService) GetByUUID(uuid string) (*models.Geofence, error) {
	return s.geofenceRepo.FindByUUID(uuid)
}

func (s *GeofenceService) List(pagination *utils.Pagination, gfType models.GeofenceType, isActive *bool) ([]models.Geofence, int64, error) {
	return s.geofenceRepo.List(pagination, gfType, isActive)
}

func (s *GeofenceService) Update(id uint64, req *CreateGeofenceRequest) (*models.Geofence, error) {
	geofence, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("geofence not found")
	}

	if req.Name != "" {
		geofence.Name = req.Name
	}
	if req.Description != "" {
		geofence.Description = req.Description
	}
	if req.Type != "" {
		geofence.Type = req.Type
	}
	if req.Shape != "" {
		geofence.Shape = req.Shape
	}
	if req.MaxAltitude > 0 {
		geofence.MaxAltitude = req.MaxAltitude
	}
	if req.MinAltitude > 0 {
		geofence.MinAltitude = req.MinAltitude
	}
	if req.CenterLat != 0 {
		geofence.CenterLat = req.CenterLat
	}
	if req.CenterLng != 0 {
		geofence.CenterLng = req.CenterLng
	}
	if req.Radius > 0 {
		geofence.Radius = req.Radius
	}
	if len(req.Coordinates) > 0 {
		coordsJSON, _ := json.Marshal(req.Coordinates)
		geofence.Coordinates = string(coordsJSON)
	}

	if err := s.geofenceRepo.Update(geofence); err != nil {
		return nil, err
	}

	return geofence, nil
}

func (s *GeofenceService) Delete(id uint64) error {
	_, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return errors.New("geofence not found")
	}
	return s.geofenceRepo.SoftDelete(&models.Geofence{}, id)
}

func (s *GeofenceService) UpdateStatus(id uint64, isActive bool) error {
	_, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return errors.New("geofence not found")
	}
	return s.geofenceRepo.UpdateStatus(id, isActive)
}

func (s *GeofenceService) AssignUAV(geofenceID, uavID uint64) error {
	_, err := s.geofenceRepo.FindByID(geofenceID)
	if err != nil {
		return errors.New("geofence not found")
	}
	return s.geofenceRepo.AssignUAV(geofenceID, uavID)
}

func (s *GeofenceService) UnassignUAV(geofenceID, uavID uint64) error {
	return s.geofenceRepo.UnassignUAV(geofenceID, uavID)
}

func (s *GeofenceService) GetUAVGeofences(uavID uint64) ([]models.Geofence, error) {
	return s.geofenceRepo.GetUAVGeofences(uavID)
}

func (s *GeofenceService) CheckViolation(uavID uint64, lat, lng, altitude float64) ([]GeofenceViolation, error) {
	geofences, err := s.geofenceRepo.GetUAVGeofences(uavID)
	if err != nil {
		return nil, err
	}

	var violations []GeofenceViolation

	for _, gf := range geofences {
		violation := s.checkSingleGeofence(&gf, lat, lng, altitude)
		if violation != nil {
			violations = append(violations, *violation)

			alert := &models.AlertEvent{
				UAVID:    uavID,
				Type:      models.AlertTypeGeofenceBreach,
				Level:     models.AlertLevelCritical,
				Title:     "电子围栏越界告警",
				Message:   "无人机越界: " + gf.Name + ", 违规类型: " + violation.ViolationType,
				Latitude:  lat,
				Longitude: lng,
				Altitude:  altitude,
			}
			_ = s.alertRepo.Create(alert)
		}
	}

	return violations, nil
}

func (s *GeofenceService) checkSingleGeofence(gf *models.Geofence, lat, lng, altitude float64) *GeofenceViolation {
	if altitude > gf.MaxAltitude && gf.MaxAltitude > 0 {
		return &GeofenceViolation{
			GeofenceID:   gf.ID,
			GeofenceName: gf.Name,
			ViolationType: "altitude_exceeded",
			Distance:    altitude - gf.MaxAltitude,
		}
	}

	if altitude < gf.MinAltitude && gf.MinAltitude > 0 && altitude > 0 {
		return &GeofenceViolation{
			GeofenceID:   gf.ID,
			GeofenceName: gf.Name,
			ViolationType: "altitude_too_low",
			Distance:    gf.MinAltitude - altitude,
		}
	}

	switch gf.Shape {
	case models.GeofenceShapeCircle:
		distance := utils.HaversineDistance(lat, lng, gf.CenterLat, gf.CenterLng)
		if gf.Type == models.GeofenceTypeExclusion {
			if distance <= gf.Radius {
				return &GeofenceViolation{
					GeofenceID:   gf.ID,
					GeofenceName: gf.Name,
					ViolationType: "inside_exclusion_zone",
					Distance:    gf.Radius - distance,
				}
			}
		} else if gf.Type == models.GeofenceTypeInclusion {
			if distance > gf.Radius {
				return &GeofenceViolation{
					GeofenceID:   gf.ID,
					GeofenceName: gf.Name,
					ViolationType: "outside_inclusion_zone",
					Distance:    distance - gf.Radius,
				}
			}
		}

	case models.GeofenceShapePolygon:
		var coords []models.Coordinate
		_ = json.Unmarshal([]byte(gf.Coordinates), &coords)
		polygon := make([][2]float64, len(coords))
		for i, c := range coords {
			polygon[i] = [2]float64{c.Lat, c.Lng}
		}
		inside := utils.IsPointInPolygon(lat, lng, polygon)

		if gf.Type == models.GeofenceTypeExclusion && inside {
			minDistance := math.Inf(1)
			for _, c := range coords {
				d := utils.HaversineDistance(lat, lng, c.Lat, c.Lng)
				if d < minDistance {
					minDistance = d
				}
			}
			return &GeofenceViolation{
				GeofenceID:   gf.ID,
				GeofenceName: gf.Name,
				ViolationType: "inside_exclusion_zone",
				Distance:    minDistance,
			}
		} else if gf.Type == models.GeofenceTypeInclusion && !inside {
			minDistance := math.Inf(1)
			for _, c := range coords {
				d := utils.HaversineDistance(lat, lng, c.Lat, c.Lng)
				if d < minDistance {
					minDistance = d
				}
			}
			return &GeofenceViolation{
				GeofenceID:   gf.ID,
				GeofenceName: gf.Name,
				ViolationType: "outside_inclusion_zone",
				Distance:    minDistance,
			}
		}
	}

	return nil
}

func (s *GeofenceService) GetActiveGeofences() ([]models.Geofence, error) {
	return s.geofenceRepo.GetActiveGeofences()
}
