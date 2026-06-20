package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type FormationRepository struct {
	*BaseRepository
}

func NewFormationRepository() *FormationRepository {
	return &FormationRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *FormationRepository) Create(formation *models.Formation) error {
	formation.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(formation)
}

func (r *FormationRepository) FindByID(id uint64) (*models.Formation, error) {
	var formation models.Formation
	if err := r.db.Preload("Members").Preload("Members.UAV").First(&formation, id).Error; err != nil {
		return nil, err
	}
	return &formation, nil
}

func (r *FormationRepository) FindByUUID(uuid string) (*models.Formation, error) {
	var formation models.Formation
	if err := r.db.Preload("Members").Preload("Members.UAV").Where("uuid = ?", uuid).First(&formation).Error; err != nil {
		return nil, err
	}
	return &formation, nil
}

func (r *FormationRepository) List(pagination *utils.Pagination, ownerID uint64) ([]models.Formation, int64, error) {
	var formations []models.Formation
	query := r.db.Model(&models.Formation{}).Preload("Members")
	if ownerID > 0 {
		query = query.Where("owner_id = ?", ownerID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).Order("created_at DESC").Find(&formations).Error; err != nil {
		return nil, 0, err
	}
	return formations, total, nil
}

func (r *FormationRepository) Update(id uint64, formation *models.Formation) error {
	return r.db.Model(&models.Formation{}).Where("id = ?", id).Updates(formation).Error
}

func (r *FormationRepository) UpdateStatus(id uint64, status models.FormationStatus) error {
	return r.db.Model(&models.Formation{}).Where("id = ?", id).Update("status", status).Error
}

func (r *FormationRepository) Delete(id uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("formation_id = ?", id).Delete(&models.FormationMember{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.Formation{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *FormationRepository) AddMember(member *models.FormationMember) error {
	return r.BaseRepository.Create(member)
}

func (r *FormationRepository) RemoveMember(formationID, uavID uint64) error {
	return r.db.Where("formation_id = ? AND uav_id = ?", formationID, uavID).Delete(&models.FormationMember{}).Error
}

func (r *FormationRepository) GetMembers(formationID uint64) ([]models.FormationMember, error) {
	var members []models.FormationMember
	if err := r.db.Preload("UAV").Where("formation_id = ?", formationID).Order("position_index ASC").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *FormationRepository) UpdateMember(memberID uint64, member *models.FormationMember) error {
	return r.db.Model(&models.FormationMember{}).Where("id = ?", memberID).Updates(member).Error
}

func (r *FormationRepository) SetLeader(formationID, uavID uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FormationMember{}).Where("formation_id = ?", formationID).Update("is_leader", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.FormationMember{}).Where("formation_id = ? AND uav_id = ?", formationID, uavID).Update("is_leader", true).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Formation{}).Where("id = ?", formationID).Update("leader_id", uavID).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *FormationRepository) GetActiveFormations() ([]models.Formation, error) {
	var formations []models.Formation
	if err := r.db.Where("status IN ?", []models.FormationStatus{
		models.FormationStatusReady,
		models.FormationStatusExecuting,
		models.FormationStatusPaused,
	}).Preload("Members").Preload("Members.UAV").Find(&formations).Error; err != nil {
		return nil, err
	}
	return formations, nil
}

func (r *FormationRepository) AddCollisionWarning(warning *models.FormationCollisionWarning) error {
	return r.BaseRepository.Create(warning)
}

func (r *FormationRepository) GetCollisionWarnings(formationID uint64, pagination *utils.Pagination) ([]models.FormationCollisionWarning, int64, error) {
	var warnings []models.FormationCollisionWarning
	query := r.db.Model(&models.FormationCollisionWarning{}).Where("formation_id = ?", formationID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).Order("timestamp DESC").Find(&warnings).Error; err != nil {
		return nil, 0, err
	}
	return warnings, total, nil
}

func (r *FormationRepository) GetUnresolvedWarnings(formationID uint64) ([]models.FormationCollisionWarning, error) {
	var warnings []models.FormationCollisionWarning
	if err := r.db.Where("formation_id = ? AND resolved = ?", formationID, false).Find(&warnings).Error; err != nil {
		return nil, err
	}
	return warnings, nil
}

func (r *FormationRepository) ResolveWarning(id uint64) error {
	return r.db.Model(&models.FormationCollisionWarning{}).Where("id = ?", id).Updates(map[string]interface{}{
		"resolved":    true,
		"resolved_at": gorm.Expr("NOW()"),
	}).Error
}
