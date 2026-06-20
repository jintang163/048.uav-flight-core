package service

import (
	"errors"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"math"
	"time"
)

type FormationService struct {
	formationRepo *repository.FormationRepository
	uavRepo       *repository.UAVRepository
	flightRepo    *repository.FlightRepository
	missionRepo   *repository.MissionRepository
}

func NewFormationService() *FormationService {
	return &FormationService{
		formationRepo: repository.NewFormationRepository(),
		uavRepo:       repository.NewUAVRepository(),
		flightRepo:    repository.NewFlightRepository(),
		missionRepo:   repository.NewMissionRepository(),
	}
}

type CreateFormationRequest struct {
	Name        string               `json:"name" binding:"required"`
	Type        models.FormationType `json:"type" binding:"required"`
	Spacing     float64              `json:"spacing"`
	Description string               `json:"description"`
	OwnerID     uint64               `json:"owner_id"`
	UAVIDs      []uint64             `json:"uav_ids"`
	LeaderID    uint64               `json:"leader_id"`
}

type UpdateFormationRequest struct {
	Name        string               `json:"name"`
	Type        models.FormationType `json:"type"`
	Spacing     float64              `json:"spacing"`
	Description string               `json:"description"`
	Status      string               `json:"status"`
}

type AddMemberRequest struct {
	UAVID uint64 `json:"uav_id" binding:"required"`
}

type LightConfigRequest struct {
	Red    uint8              `json:"red"`
	Green  uint8              `json:"green"`
	Blue   uint8              `json:"blue"`
	Effect models.LightEffect `json:"effect"`
}

func (s *FormationService) Create(req *CreateFormationRequest) (*models.Formation, error) {
	if req.Spacing <= 0 {
		req.Spacing = 5.0
	}

	formation := &models.Formation{
		Name:        req.Name,
		Type:        req.Type,
		Status:      models.FormationStatusIdle,
		Spacing:     req.Spacing,
		Description: req.Description,
		OwnerID:     req.OwnerID,
	}

	if err := s.formationRepo.Create(formation); err != nil {
		return nil, err
	}

	if len(req.UAVIDs) > 0 {
		for i, uavID := range req.UAVIDs {
			offset := s.calculateFormationPosition(i, len(req.UAVIDs), req.Type, req.Spacing)
			isLeader := false
			if req.LeaderID == 0 && i == 0 {
				isLeader = true
			} else if req.LeaderID == uavID {
				isLeader = true
			}

			member := &models.FormationMember{
				FormationID:   formation.ID,
				UAVID:         uavID,
				PositionIndex: i,
				OffsetX:       offset.X,
				OffsetY:       offset.Y,
				OffsetZ:       offset.Z,
				IsLeader:      isLeader,
				Status:        "idle",
			}
			if err := s.formationRepo.AddMember(member); err != nil {
				return nil, err
			}
		}

		if req.LeaderID > 0 {
			if err := s.formationRepo.SetLeader(formation.ID, req.LeaderID); err != nil {
				return nil, err
			}
		}
	}

	return s.formationRepo.FindByID(formation.ID)
}

type FormationOffset struct {
	X float64
	Y float64
	Z float64
}

func (s *FormationService) calculateFormationPosition(index, total int, fType models.FormationType, spacing float64) FormationOffset {
	switch fType {
	case models.FormationLine:
		return s.calculateLineFormation(index, total, spacing)
	case models.FormationTriangle:
		return s.calculateTriangleFormation(index, total, spacing)
	case models.FormationCircle:
		return s.calculateCircleFormation(index, total, spacing)
	default:
		return FormationOffset{X: 0, Y: 0, Z: 0}
	}
}

func (s *FormationService) calculateLineFormation(index, total int, spacing float64) FormationOffset {
	startOffset := -float64(total-1) * spacing / 2.0
	return FormationOffset{
		X: 0,
		Y: startOffset + float64(index)*spacing,
		Z: 0,
	}
}

func (s *FormationService) calculateTriangleFormation(index, total int, spacing float64) FormationOffset {
	if index == 0 {
		return FormationOffset{X: 0, Y: 0, Z: 0}
	}

	row := 0
	countInRow := 1
	placed := 1

	for placed < index+1 {
		row++
		countInRow++
		placed += countInRow
	}

	placed -= countInRow
	posInRow := index - placed

	rowWidth := float64(countInRow-1) * spacing
	startY := -rowWidth / 2.0
	rowX := -float64(row) * spacing * 0.866

	return FormationOffset{
		X: rowX,
		Y: startY + float64(posInRow)*spacing,
		Z: 0,
	}
}

func (s *FormationService) calculateCircleFormation(index, total int, spacing float64) FormationOffset {
	if total <= 1 {
		return FormationOffset{X: 0, Y: 0, Z: 0}
	}

	radius := spacing / (2.0 * math.Sin(math.Pi/float64(total)))
	angle := 2.0*math.Pi*float64(index)/float64(total) - math.Pi/2.0

	return FormationOffset{
		X: radius * math.Cos(angle),
		Y: radius * math.Sin(angle),
		Z: 0,
	}
}

func (s *FormationService) GetByID(id uint64) (*models.Formation, error) {
	return s.formationRepo.FindByID(id)
}

func (s *FormationService) GetByUUID(uuid string) (*models.Formation, error) {
	return s.formationRepo.FindByUUID(uuid)
}

func (s *FormationService) List(pagination *utils.Pagination, ownerID uint64) ([]models.Formation, int64, error) {
	return s.formationRepo.List(pagination, ownerID)
}

func (s *FormationService) Update(id uint64, req *UpdateFormationRequest) (*models.Formation, error) {
	formation, err := s.formationRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("formation not found")
	}

	if req.Name != "" {
		formation.Name = req.Name
	}
	if req.Type != "" {
		formation.Type = req.Type
	}
	if req.Spacing > 0 {
		formation.Spacing = req.Spacing
	}
	if req.Description != "" {
		formation.Description = req.Description
	}
	if req.Status != "" {
		formation.Status = models.FormationStatus(req.Status)
	}

	if err := s.formationRepo.Update(id, formation); err != nil {
		return nil, err
	}

	if req.Type != "" || req.Spacing > 0 {
		s.recalculateFormationOffsets(id)
	}

	return s.formationRepo.FindByID(id)
}

func (s *FormationService) recalculateFormationOffsets(formationID uint64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return err
	}

	members, err := s.formationRepo.GetMembers(formationID)
	if err != nil {
		return err
	}

	for i, member := range members {
		offset := s.calculateFormationPosition(i, len(members), formation.Type, formation.Spacing)
		member.OffsetX = offset.X
		member.OffsetY = offset.Y
		member.OffsetZ = offset.Z
		if err := s.formationRepo.UpdateMember(member.ID, &member); err != nil {
			return err
		}
	}

	return nil
}

func (s *FormationService) Delete(id uint64) error {
	_, err := s.formationRepo.FindByID(id)
	if err != nil {
		return errors.New("formation not found")
	}
	return s.formationRepo.Delete(id)
}

func (s *FormationService) AddMember(formationID uint64, req *AddMemberRequest) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	if _, err := s.uavRepo.FindByID(req.UAVID); err != nil {
		return errors.New("uav not found")
	}

	members, err := s.formationRepo.GetMembers(formationID)
	if err != nil {
		return err
	}

	for _, m := range members {
		if m.UAVID == req.UAVID {
			return errors.New("uav already in formation")
		}
	}

	index := len(members)
	offset := s.calculateFormationPosition(index, index+1, formation.Type, formation.Spacing)

	member := &models.FormationMember{
		FormationID:   formationID,
		UAVID:         req.UAVID,
		PositionIndex: index,
		OffsetX:       offset.X,
		OffsetY:       offset.Y,
		OffsetZ:       offset.Z,
		IsLeader:      index == 0 && formation.LeaderID == 0,
		Status:        "idle",
	}

	if err := s.formationRepo.AddMember(member); err != nil {
		return err
	}

	return s.recalculateFormationOffsets(formationID)
}

func (s *FormationService) RemoveMember(formationID, uavID uint64) error {
	if _, err := s.formationRepo.FindByID(formationID); err != nil {
		return errors.New("formation not found")
	}

	if err := s.formationRepo.RemoveMember(formationID, uavID); err != nil {
		return err
	}

	return s.recalculateFormationOffsets(formationID)
}

func (s *FormationService) SetLeader(formationID, uavID uint64) error {
	return s.formationRepo.SetLeader(formationID, uavID)
}

func (s *FormationService) GetMembers(formationID uint64) ([]models.FormationMember, error) {
	return s.formationRepo.GetMembers(formationID)
}

func (s *FormationService) Start(formationID uint64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	if formation.Status == models.FormationStatusExecuting {
		return errors.New("formation already executing")
	}

	if err := s.formationRepo.UpdateStatus(formationID, models.FormationStatusExecuting); err != nil {
		return err
	}

	s.broadcastFormationStatus(formationID, "executing")

	return nil
}

func (s *FormationService) Pause(formationID uint64) error {
	if _, err := s.formationRepo.FindByID(formationID); err != nil {
		return errors.New("formation not found")
	}

	if err := s.formationRepo.UpdateStatus(formationID, models.FormationStatusPaused); err != nil {
		return err
	}

	s.broadcastFormationStatus(formationID, "paused")
	return nil
}

func (s *FormationService) Resume(formationID uint64) error {
	if _, err := s.formationRepo.FindByID(formationID); err != nil {
		return errors.New("formation not found")
	}

	if err := s.formationRepo.UpdateStatus(formationID, models.FormationStatusExecuting); err != nil {
		return err
	}

	s.broadcastFormationStatus(formationID, "executing")
	return nil
}

func (s *FormationService) Stop(formationID uint64) error {
	if err := s.formationRepo.UpdateStatus(formationID, models.FormationStatusIdle); err != nil {
		return err
	}

	s.broadcastFormationStatus(formationID, "idle")
	return nil
}

func (s *FormationService) GetActiveFormations() ([]models.Formation, error) {
	return s.formationRepo.GetActiveFormations()
}

func (s *FormationService) CheckCollisionsRealtime(formationID uint64) ([]models.FormationCollisionWarning, error) {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return nil, err
	}

	var warnings []models.FormationCollisionWarning
	members := formation.Members

	type uavPos struct {
		lat float64
		lng float64
		alt float64
		ok  bool
	}

	positions := make(map[uint64]uavPos)
	for _, m := range members {
		status, err := s.flightRepo.GetLatestStatus(m.UAVID)
		if err != nil {
			positions[m.UAVID] = uavPos{ok: false}
			continue
		}
		positions[m.UAVID] = uavPos{
			lat: status.Latitude,
			lng: status.Longitude,
			alt: status.AltitudeMSL,
			ok:  true,
		}
	}

	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			p1, ok1 := positions[members[i].UAVID]
			p2, ok2 := positions[members[j].UAVID]

			if !ok1 || !ok2 {
				continue
			}

			dx, dy := latLngToMeters(p1.lat, p1.lng, p2.lat, p2.lng)
			dz := p1.alt - p2.alt
			distance := math.Sqrt(dx*dx+dy*dy+dz*dz)

			if distance < 5.0 {
				warning := models.FormationCollisionWarning{
					FormationID:  formationID,
					UAVID1:       members[i].UAVID,
					UAVID2:       members[j].UAVID,
					Distance:     distance,
					WarningLevel: "warning",
				}
				if distance < 3.0 {
					warning.WarningLevel = "critical"
				}
				warnings = append(warnings, warning)

				s.formationRepo.AddCollisionWarning(&warning)

				websocket.BroadcastFormationCollisionWarning(formationID, members[i].UAVID, members[j].UAVID, distance, string(warning.WarningLevel))
			}
		}
	}

	return warnings, nil
}

func latLngToMeters(lat1, lng1, lat2, lng2 float64) (dx, dy float64) {
	avgLat := (lat1 + lat2) / 2.0 * math.Pi / 180.0
	dx = (lng2 - lng1) * math.Pi / 180.0 * 6378137.0 * math.Cos(avgLat)
	dy = (lat2 - lat1) * math.Pi / 180.0 * 6378137.0
	return dx, dy
}

func (s *FormationService) GetCollisionWarnings(formationID uint64, pagination *utils.Pagination) ([]models.FormationCollisionWarning, int64, error) {
	return s.formationRepo.GetCollisionWarnings(formationID, pagination)
}

func (s *FormationService) SetLightConfig(formationID uint64, req *LightConfigRequest) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	for _, member := range formation.Members {
		cmdMgr := mavlink.NewCommandManager()
		lightData := map[string]interface{}{
			"formation_id": formationID,
			"uav_id":       member.UAVID,
			"red":          req.Red,
			"green":        req.Green,
			"blue":         req.Blue,
			"effect":       string(req.Effect),
		}

		_ = cmdMgr.SendCommand(member.UAVID, encodeFormationLightCommand(req.Red, req.Green, req.Blue, req.Effect))

		_ = lightData
	}

	websocket.BroadcastFormationLight(formationID, req.Red, req.Green, req.Blue, string(req.Effect))

	return nil
}

func encodeFormationLightCommand(red, green, blue uint8, effect models.LightEffect) []byte {
	buf := make([]byte, 8)
	buf[0] = 0xFE
	buf[1] = 5
	buf[2] = 0x01
	buf[3] = 0x00
	buf[4] = red
	buf[5] = green
	buf[6] = blue
	switch effect {
	case models.LightEffectStatic:
		buf[7] = 0x00
	case models.LightEffectBlink:
		buf[7] = 0x01
	case models.LightEffectRainbow:
		buf[7] = 0x02
	case models.LightEffectBreathing:
		buf[7] = 0x03
	default:
		buf[7] = 0x00
	}
	return buf
}

func (s *FormationService) SyncWaypoints(formationID uint64, missionID uint64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	mission, err := s.missionRepo.FindByID(missionID)
	if err != nil {
		return errors.New("mission not found")
	}

	waypoints, err := s.missionRepo.GetWaypoints(missionID)
	if err != nil {
		return errors.New("failed to get waypoints")
	}

	cmdMgr := mavlink.NewCommandManager()

	for _, member := range formation.Members {
		encodedWaypoints := encodeMissionWaypoints(waypoints)
		_ = cmdMgr.SendCommand(member.UAVID, encodedWaypoints)

		websocket.BroadcastMissionUpdate(member.UAVID, map[string]interface{}{
			"formation_id": formationID,
			"mission_id":   missionID,
			"action":       "sync_waypoints",
			"waypoint_count": len(waypoints),
		})
	}

	_ = mission

	websocket.BroadcastFormationUpdate(formationID, map[string]interface{}{
		"formation_id":   formationID,
		"mission_id":     missionID,
		"action":         "waypoints_synced",
		"waypoint_count": len(waypoints),
		"timestamp":      time.Now().UnixNano() / 1e6,
	})

	return nil
}

func encodeMissionWaypoints(waypoints []models.MissionWaypoint) []byte {
	totalLen := 6 + len(waypoints)*20
	buf := make([]byte, totalLen)
	buf[0] = 0xFE
	buf[1] = byte(totalLen - 6)
	buf[2] = 0x02
	buf[3] = byte(len(waypoints))

	for i, wp := range waypoints {
		offset := 4 + i*20
		latBytes := float64ToBytes(wp.Latitude)
		lngBytes := float64ToBytes(wp.Longitude)
		altBytes := float32ToBytes(float32(wp.Altitude))
		copy(buf[offset:offset+8], latBytes)
		copy(buf[offset+8:offset+16], lngBytes)
		copy(buf[offset+16:offset+20], altBytes)
	}

	buf[totalLen-2] = 0x00
	buf[totalLen-1] = 0x00
	return buf
}

func float64ToBytes(v float64) []byte {
	bits := math.Float64bits(v)
	buf := make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(bits >> uint(i*8))
	}
	return buf
}

func float32ToBytes(v float32) []byte {
	bits := math.Float32bits(v)
	buf := make([]byte, 4)
	for i := 0; i < 4; i++ {
		buf[i] = byte(bits >> uint(i*8))
	}
	return buf
}

func (s *FormationService) MultiTakeoff(formationID uint64, altitude float64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	cmdMgr := mavlink.NewCommandManager()

	for _, member := range formation.Members {
		takeoffCmd := encodeTakeoffCommand(altitude)
		_ = cmdMgr.SendCommand(member.UAVID, takeoffCmd)

		websocket.BroadcastUAVStatus(member.UAVID, "takeoff", "编队一键起飞")
	}

	websocket.BroadcastFormationStatus(formationID, "executing")

	return nil
}

func encodeTakeoffCommand(altitude float64) []byte {
	buf := make([]byte, 8)
	buf[0] = 0xFE
	buf[1] = 2
	buf[2] = 0x03
	buf[3] = 0x00
	altBytes := float32ToBytes(float32(altitude))
	copy(buf[4:8], altBytes)
	return buf
}

func (s *FormationService) broadcastFormationStatus(formationID uint64, status string) {
	websocket.BroadcastFormationStatus(formationID, status)
}
