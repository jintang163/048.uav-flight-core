package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"math"
)

type FormationService struct {
	formationRepo *repository.FormationRepository
	uavRepo       *repository.UAVRepository
}

func NewFormationService() *FormationService {
	return &FormationService{
		formationRepo: repository.NewFormationRepository(),
		uavRepo:       repository.NewUAVRepository(),
	}
}

type CreateFormationRequest struct {
	Name        string              `json:"name" binding:"required"`
	Type        models.FormationType `json:"type" binding:"required"`
	Spacing     float64             `json:"spacing"`
	Description string              `json:"description"`
	OwnerID     uint64              `json:"owner_id"`
	UAVIDs      []uint64            `json:"uav_ids"`
	LeaderID    uint64              `json:"leader_id"`
}

type UpdateFormationRequest struct {
	Name        string              `json:"name"`
	Type        models.FormationType `json:"type"`
	Spacing     float64             `json:"spacing"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
}

type AddMemberRequest struct {
	UAVID uint64 `json:"uav_id" binding:"required"`
}

type LightConfigRequest struct {
	Red    uint8            `json:"red"`
	Green  uint8            `json:"green"`
	Blue   uint8            `json:"blue"`
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

	uav, err := s.uavRepo.FindByID(req.UAVID)
	if err != nil {
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
	_, err := s.formationRepo.FindByID(formationID)
	if err != nil {
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

	return s.formationRepo.UpdateStatus(formationID, models.FormationStatusExecuting)
}

func (s *FormationService) Pause(formationID uint64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	if formation.Status != models.FormationStatusExecuting {
		return errors.New("formation not executing")
	}

	return s.formationRepo.UpdateStatus(formationID, models.FormationStatusPaused)
}

func (s *FormationService) Resume(formationID uint64) error {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}

	if formation.Status != models.FormationStatusPaused {
		return errors.New("formation not paused")
	}

	return s.formationRepo.UpdateStatus(formationID, models.FormationStatusExecuting)
}

func (s *FormationService) Stop(formationID uint64) error {
	return s.formationRepo.UpdateStatus(formationID, models.FormationStatusIdle)
}

func (s *FormationService) GetActiveFormations() ([]models.Formation, error) {
	return s.formationRepo.GetActiveFormations()
}

func (s *FormationService) CheckCollisions(formationID uint64) ([]models.FormationCollisionWarning, error) {
	formation, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return nil, err
	}

	var warnings []models.FormationCollisionWarning
	members := formation.Members

	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			dx := members[i].OffsetX - members[j].OffsetX
			dy := members[i].OffsetY - members[j].OffsetY
			dz := members[i].OffsetZ - members[j].OffsetZ
			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

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
			}
		}
	}

	return warnings, nil
}

func (s *FormationService) GetCollisionWarnings(formationID uint64, pagination *utils.Pagination) ([]models.FormationCollisionWarning, int64, error) {
	return s.formationRepo.GetCollisionWarnings(formationID, pagination)
}

func (s *FormationService) SetLightConfig(formationID uint64, req *LightConfigRequest) error {
	_, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}
	return nil
}

func (s *FormationService) SyncWaypoints(formationID uint64, missionID uint64) error {
	_, err := s.formationRepo.FindByID(formationID)
	if err != nil {
		return errors.New("formation not found")
	}
	return nil
}
