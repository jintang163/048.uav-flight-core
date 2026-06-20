package service

import (
	"encoding/json"
	"errors"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"math"
)

type GeofenceService struct {
	geofenceRepo      *repository.GeofenceRepository
	violationRepo     *repository.GeofenceViolationRepository
	unlockingRepo     *repository.TemporaryUnlockingRepository
	alertRepo         *repository.AlertRepository
	uavRepo           *repository.UAVRepository
	cache             *GeofenceCacheService
}

func NewGeofenceService() *GeofenceService {
	return &GeofenceService{
		geofenceRepo:  repository.NewGeofenceRepository(),
		violationRepo: repository.NewGeofenceViolationRepository(),
		unlockingRepo: repository.NewTemporaryUnlockingRepository(),
		alertRepo:     repository.NewAlertRepository(),
		uavRepo:       repository.NewUAVRepository(),
		cache:         NewGeofenceCacheService(),
	}
}

type GeofenceViolation struct {
	GeofenceID     uint64  `json:"geofence_id"`
	GeofenceName   string  `json:"geofence_name"`
	GeofenceCategory string `json:"geofence_category"`
	ViolationType  string  `json:"violation_type"`
	Severity       string  `json:"severity"`
	Distance       float64 `json:"distance"`
	Action         string  `json:"action"`
}

func (s *GeofenceService) Create(geofence *models.Geofence, coords [][]float64, uavIDs []uint64) (*models.Geofence, error) {
	if geofence.Shape == models.GeofenceShapePolygon && len(coords) < 3 {
		return nil, errors.New("polygon geofence requires at least 3 coordinates")
	}
	if geofence.Shape == models.GeofenceShapeCircle && geofence.Radius <= 0 {
		return nil, errors.New("circle geofence requires radius")
	}

	if geofence.MaxAltitude <= 0 {
		geofence.MaxAltitude = 120
	}
	if geofence.MaxDistance <= 0 {
		geofence.MaxDistance = 500
	}

	coordinateList := make([]models.Coordinate, len(coords))
	for i, c := range coords {
		if len(c) >= 2 {
			coordinateList[i] = models.Coordinate{Lat: c[0], Lng: c[1]}
		}
	}

	if len(coordinateList) > 0 {
		coordsJSON, _ := json.Marshal(coordinateList)
		geofence.Coordinates = string(coordsJSON)
	}

	if geofence.Source == "" {
		geofence.Source = models.GeofenceSourceUser
	}
	if geofence.Category == "" {
		geofence.Category = models.GeofenceCategoryCustom
	}
	if geofence.FailAction == "" {
		geofence.FailAction = models.FailActionWarn
	}

	if err := s.geofenceRepo.Create(geofence); err != nil {
		return nil, err
	}

	for _, uavID := range uavIDs {
		_ = s.geofenceRepo.AssignUAV(geofence.ID, uavID)
	}

	s.cache.Refresh()

	return s.geofenceRepo.FindByID(geofence.ID)
}

func (s *GeofenceService) GetByID(id uint64) (*models.Geofence, error) {
	return s.geofenceRepo.FindByID(id)
}

func (s *GeofenceService) GetByUUID(uuid string) (*models.Geofence, error) {
	return s.geofenceRepo.FindByUUID(uuid)
}

func (s *GeofenceService) List(pagination *utils.Pagination, gfType string, isActiveStr string, category string, source string) ([]models.Geofence, int64, error) {
	var isActive *bool
	if isActiveStr != "" {
		b := isActiveStr == "true" || isActiveStr == "1"
		isActive = &b
	}
	return s.geofenceRepo.ListFiltered(pagination, models.GeofenceType(gfType), isActive, models.GeofenceCategory(category), models.GeofenceSource(source))
}

func (s *GeofenceService) Update(id uint64, geofence *models.Geofence, coords [][]float64, uavIDs []uint64) (*models.Geofence, error) {
	existing, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("geofence not found")
	}

	if geofence.Name != "" {
		existing.Name = geofence.Name
	}
	if geofence.Description != "" {
		existing.Description = geofence.Description
	}
	if geofence.Type != "" {
		existing.Type = geofence.Type
	}
	if geofence.Shape != "" {
		existing.Shape = geofence.Shape
	}
	if geofence.Category != "" {
		existing.Category = geofence.Category
	}
	if geofence.MaxAltitude > 0 {
		existing.MaxAltitude = geofence.MaxAltitude
	}
	if geofence.MinAltitude > 0 {
		existing.MinAltitude = geofence.MinAltitude
	}
	if geofence.MaxDistance > 0 {
		existing.MaxDistance = geofence.MaxDistance
	}
	if geofence.CenterLat != 0 {
		existing.CenterLat = geofence.CenterLat
	}
	if geofence.CenterLng != 0 {
		existing.CenterLng = geofence.CenterLng
	}
	if geofence.Radius > 0 {
		existing.Radius = geofence.Radius
	}
	if geofence.FailAction != "" {
		existing.FailAction = geofence.FailAction
	}
	if geofence.CountryCode != "" {
		existing.CountryCode = geofence.CountryCode
	}
	if geofence.CityName != "" {
		existing.CityName = geofence.CityName
	}

	if len(coords) > 0 {
		coordinateList := make([]models.Coordinate, len(coords))
		for i, c := range coords {
			if len(c) >= 2 {
				coordinateList[i] = models.Coordinate{Lat: c[0], Lng: c[1]}
			}
		}
		coordsJSON, _ := json.Marshal(coordinateList)
		existing.Coordinates = string(coordsJSON)
	}

	if err := s.geofenceRepo.Update(existing); err != nil {
		return nil, err
	}

	if len(uavIDs) > 0 {
		_ = s.geofenceRepo.ClearUAVs(id)
		for _, uavID := range uavIDs {
			_ = s.geofenceRepo.AssignUAV(id, uavID)
		}
	}

	s.cache.Refresh()

	return s.geofenceRepo.FindByID(id)
}

func (s *GeofenceService) Delete(id uint64) error {
	_, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return errors.New("geofence not found")
	}
	err = s.geofenceRepo.SoftDelete(&models.Geofence{}, id)
	if err == nil {
		s.cache.Refresh()
	}
	return err
}

func (s *GeofenceService) UpdateStatus(id uint64, isActive bool) error {
	_, err := s.geofenceRepo.FindByID(id)
	if err != nil {
		return errors.New("geofence not found")
	}
	err = s.geofenceRepo.UpdateStatus(id, isActive)
	if err == nil {
		s.cache.Refresh()
	}
	return err
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

func (s *GeofenceService) GetByUAV(uavID uint64) ([]models.Geofence, error) {
	return s.geofenceRepo.GetUAVGeofences(uavID)
}

func (s *GeofenceService) CheckViolation(uavID uint64, lat, lng, altitude float64) ([]GeofenceViolation, error) {
	uavGeofences, err := s.cache.GetUAVGeofences(uavID)
	if err != nil {
		uavGeofences, err = s.geofenceRepo.GetUAVGeofences(uavID)
		if err != nil {
			return nil, err
		}
	}

	nationalGeofences, err := s.cache.GetNational()
	if err != nil {
		nationalGeofences, _ = s.GetAllNational()
	}

	allGeofences := make([]models.Geofence, 0, len(uavGeofences)+len(nationalGeofences))
	allGeofences = append(allGeofences, uavGeofences...)
	for _, ng := range nationalGeofences {
		found := false
		for _, ug := range uavGeofences {
			if ug.ID == ng.ID {
				found = true
				break
			}
		}
		if !found {
			allGeofences = append(allGeofences, ng)
		}
	}

	var violations []GeofenceViolation

	for i := range allGeofences {
		gf := &allGeofences[i]
		if !gf.IsActive {
			continue
		}

		if s.hasActiveUnlock(uavID, gf.ID, gf.Category) {
			continue
		}

		violation := s.checkSingleGeofence(gf, lat, lng, altitude, uavID)
		if violation != nil {
			severity := models.ViolationSeverityWarning
			if gf.Category == models.GeofenceCategoryAirport ||
				gf.Category == models.GeofenceCategoryMilitary ||
				gf.Category == models.GeofenceCategoryNuclear ||
				gf.Category == models.GeofenceCategoryNational {
				severity = models.ViolationSeverityCritical
			}
			if gf.FailAction == models.FailActionRTL || gf.FailAction == models.FailActionLand {
				severity = models.ViolationSeverityFatal
			}

			action := gf.FailAction
			s.executeFailAction(uavID, *gf, action)

			_, _ = NewGeofenceViolationService().LogViolation(
				uavID, gf.ID, models.ViolationType(violation.ViolationType),
				severity, lat, lng, altitude, violation.Distance, action,
			)

			violations = append(violations, GeofenceViolation{
				GeofenceID:       gf.ID,
				GeofenceName:     gf.Name,
				GeofenceCategory: string(gf.Category),
				ViolationType:    violation.ViolationType,
				Severity:         string(severity),
				Distance:         violation.Distance,
				Action:           string(action),
			})

			alert := &models.AlertEvent{
				UAVID:     uavID,
				Type:      models.AlertTypeGeofenceBreach,
				Level:     models.AlertLevel(severity),
				Title:     "电子围栏越界告警",
				Message:   "无人机越界: " + gf.Name + ", 违规类型: " + violation.ViolationType + ", 处置: " + string(action),
				Latitude:  lat,
				Longitude: lng,
				Altitude:  altitude,
			}
			_ = s.alertRepo.Create(alert)
		}
	}

	return violations, nil
}

func (s *GeofenceService) hasActiveUnlock(uavID uint64, geofenceID uint64, category models.GeofenceCategory) bool {
	unlock, err := s.unlockingRepo.CheckActiveUnlock(uavID, geofenceID)
	if err == nil && unlock != nil {
		return true
	}
	activeUnlocks, err := s.unlockingRepo.GetActiveUnlockings(uavID, string(category))
	if err == nil && len(activeUnlocks) > 0 {
		return true
	}
	return false
}

func (s *GeofenceService) checkSingleGeofence(gf *models.Geofence, lat, lng, altitude float64, uavID uint64) *GeofenceViolation {
	if altitude > gf.MaxAltitude && gf.MaxAltitude > 0 {
		return &GeofenceViolation{
			GeofenceID:     gf.ID,
			GeofenceName:   gf.Name,
			ViolationType:  string(models.ViolationTypeAltitudeExceeded),
			Distance:       altitude - gf.MaxAltitude,
		}
	}

	if altitude < gf.MinAltitude && gf.MinAltitude > 0 && altitude > 0 {
		return &GeofenceViolation{
			GeofenceID:     gf.ID,
			GeofenceName:   gf.Name,
			ViolationType:  string(models.ViolationTypeAltitudeTooLow),
			Distance:       gf.MinAltitude - altitude,
		}
	}

	if gf.MaxDistance > 0 && uavID > 0 {
		uav, err := s.uavRepo.FindByID(uavID)
		if err == nil && uav != nil && uav.HomeLatitude != 0 && uav.HomeLongitude != 0 {
			dist := utils.HaversineDistance(lat, lng, uav.HomeLatitude, uav.HomeLongitude)
			if dist > gf.MaxDistance {
				return &GeofenceViolation{
					GeofenceID:     gf.ID,
					GeofenceName:   gf.Name,
					ViolationType:  string(models.ViolationTypeDistanceExceeded),
					Distance:       dist - gf.MaxDistance,
				}
			}
		}
	}

	switch gf.Shape {
	case models.GeofenceShapeCircle:
		distance := utils.HaversineDistance(lat, lng, gf.CenterLat, gf.CenterLng)
		if gf.Type == models.GeofenceTypeExclusion {
			if distance <= gf.Radius {
				return &GeofenceViolation{
					GeofenceID:     gf.ID,
					GeofenceName:   gf.Name,
					ViolationType:  string(models.ViolationTypeInsideExclusionZone),
					Distance:       gf.Radius - distance,
				}
			}
		} else if gf.Type == models.GeofenceTypeInclusion {
			if distance > gf.Radius {
				return &GeofenceViolation{
					GeofenceID:     gf.ID,
					GeofenceName:   gf.Name,
					ViolationType:  string(models.ViolationTypeOutsideInclusionZone),
					Distance:       distance - gf.Radius,
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
				GeofenceID:     gf.ID,
				GeofenceName:   gf.Name,
				ViolationType:  string(models.ViolationTypeInsideExclusionZone),
				Distance:       minDistance,
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
				GeofenceID:     gf.ID,
				GeofenceName:   gf.Name,
				ViolationType:  string(models.ViolationTypeOutsideInclusionZone),
				Distance:       minDistance,
			}
		}
	}

	return nil
}

func (s *GeofenceService) executeFailAction(uavID uint64, gf models.Geofence, action models.FailAction) {
	cmdMgr := mavlink.NewCommandManager()
	if cmdMgr == nil {
		return
	}

	switch action {
	case models.FailActionRTL:
		data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(uavID, data)
	case models.FailActionLand:
		data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_NAV_LAND, 0, 0, 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(uavID, data)
	case models.FailActionHover:
		data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_GUIDED_LIMITS, 0, 0, 0, 0, 0, 0, 0)
		_ = cmdMgr.SendCommand(uavID, data)
	}
}

func (s *GeofenceService) GetActiveGeofences() ([]models.Geofence, error) {
	return s.cache.GetAllActive()
}

func (s *GeofenceService) GetAllNational() ([]models.Geofence, error) {
	pagination := &utils.Pagination{Page: 1, PageSize: 1000}
	gfType := models.GeofenceType("")
	isActive := true
	category := models.GeofenceCategory("")
	source := models.GeofenceSourceNational
	result, _, err := s.geofenceRepo.ListFiltered(pagination, gfType, &isActive, category, source)
	return result, err
}

func (s *GeofenceService) ImportNationalGeofences(geofences []models.Geofence) (int, error) {
	count := 0
	for _, gf := range geofences {
		gf.Source = models.GeofenceSourceNational
		gf.IsActive = true
		if gf.FailAction == "" {
			gf.FailAction = models.FailActionHover
		}
		if gf.Category == "" {
			gf.Category = models.GeofenceCategoryNational
		}
		if err := s.geofenceRepo.Create(&gf); err == nil {
			count++
		}
	}
	if count > 0 {
		s.cache.Refresh()
	}
	return count, nil
}

func (s *GeofenceService) RefreshCache() {
	s.cache.Refresh()
}

type TakeoffCheckResult struct {
	Allowed     bool     `json:"allowed"`
	Reason      string   `json:"reason"`
	FenceIDs    []uint64 `json:"fence_ids"`
	FenceNames  []string `json:"fence_names"`
	Severity    string   `json:"severity"`
}

func (s *GeofenceService) CheckTakeoffPermission(uavID uint64, lat, lng, alt float64) (*TakeoffCheckResult, error) {
	result := &TakeoffCheckResult{
		Allowed:    true,
		Severity:   string(models.ViolationSeverityWarning),
	}

	nationalGeofences, err := s.cache.GetNational()
	if err != nil {
		nationalGeofences, _ = s.GetAllNational()
	}

	uavGeofences, err := s.cache.GetUAVGeofences(uavID)
	if err != nil {
		uavGeofences, err = s.geofenceRepo.GetUAVGeofences(uavID)
		if err != nil {
			uavGeofences = nil
		}
	}

	allGeofences := make([]models.Geofence, 0, len(uavGeofences)+len(nationalGeofences))
	allGeofences = append(allGeofences, nationalGeofences...)
	allGeofences = append(allGeofences, uavGeofences...)

	for i := range allGeofences {
		gf := &allGeofences[i]
		if !gf.IsActive || gf.Type != models.GeofenceTypeExclusion {
			continue
		}

		if s.hasActiveUnlock(uavID, gf.ID, gf.Category) {
			continue
		}

		inside := false
		switch gf.Shape {
		case models.GeofenceShapeCircle:
			dist := utils.HaversineDistance(lat, lng, gf.CenterLat, gf.CenterLng)
			inside = dist <= gf.Radius
		case models.GeofenceShapePolygon:
			var coords []models.Coordinate
			_ = json.Unmarshal([]byte(gf.Coordinates), &coords)
			polygon := make([][2]float64, len(coords))
			for j, c := range coords {
				polygon[j] = [2]float64{c.Lat, c.Lng}
			}
			inside = utils.IsPointInPolygon(lat, lng, polygon)
		}

		if inside {
			result.Allowed = false
			result.FenceIDs = append(result.FenceIDs, gf.ID)
			result.FenceNames = append(result.FenceNames, gf.Name)
			if gf.Category == models.GeofenceCategoryAirport ||
				gf.Category == models.GeofenceCategoryMilitary ||
				gf.Category == models.GeofenceCategoryNuclear ||
				gf.Category == models.GeofenceCategoryNational {
				result.Severity = string(models.ViolationSeverityCritical)
			}
		}
	}

	if !result.Allowed {
		result.Reason = "位于禁飞区内，无法起飞："
		for i, name := range result.FenceNames {
			if i > 0 {
				result.Reason += "、"
			}
			result.Reason += name
		}
	}

	return result, nil
}
