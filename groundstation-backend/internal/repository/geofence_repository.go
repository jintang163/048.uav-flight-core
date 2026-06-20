package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
)

type GeofenceRepository struct {
	*BaseRepository
}

func NewGeofenceRepository() *GeofenceRepository {
	return &GeofenceRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *GeofenceRepository) Create(geofence *models.Geofence) error {
	geofence.UUID = utils.GenerateUUID()
	return r.db.Create(geofence).Error
}

func (r *GeofenceRepository) FindByID(id uint64) (*models.Geofence, error) {
	var geofence models.Geofence
	if err := r.db.Preload("UAVs").First(&geofence, id).Error; err != nil {
		return nil, err
	}
	return &geofence, nil
}

func (r *GeofenceRepository) FindByUUID(uuid string) (*models.Geofence, error) {
	var geofence models.Geofence
	if err := r.db.Preload("UAVs").Where("uuid = ?", uuid).First(&geofence).Error; err != nil {
		return nil, err
	}
	return &geofence, nil
}

func (r *GeofenceRepository) List(pagination *utils.Pagination, gfType models.GeofenceType, isActive *bool) ([]models.Geofence, int64, error) {
	var geofences []models.Geofence
	query := r.db.Model(&models.Geofence{})
	if gfType != "" {
		query = query.Where("type = ?", gfType)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&geofences).Error; err != nil {
		return nil, 0, err
	}
	return geofences, total, nil
}

func (r *GeofenceRepository) ListFiltered(pagination *utils.Pagination, gfType models.GeofenceType, isActive *bool, category models.GeofenceCategory, source models.GeofenceSource) ([]models.Geofence, int64, error) {
	var geofences []models.Geofence
	query := r.db.Model(&models.Geofence{})
	if gfType != "" {
		query = query.Where("type = ?", gfType)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&geofences).Error; err != nil {
		return nil, 0, err
	}
	return geofences, total, nil
}

func (r *GeofenceRepository) ClearUAVs(geofenceID uint64) error {
	return r.db.Where("geofence_id = ?", geofenceID).Delete(&models.UAVGeofence{}).Error
}

func (r *GeofenceRepository) GetActiveGeofences() ([]models.Geofence, error) {
	var geofences []models.Geofence
	err := r.db.Where("is_active = ?", true).Find(&geofences).Error
	return geofences, err
}

func (r *GeofenceRepository) GetUAVGeofences(uavID uint64) ([]models.Geofence, error) {
	var geofences []models.Geofence
	err := r.db.Joins("JOIN uav_geofences ON uav_geofences.geofence_id = geofences.id").
		Where("uav_geofences.uav_id = ? AND geofences.is_active = ?", uavID, true).
		Find(&geofences).Error
	return geofences, err
}

func (r *GeofenceRepository) AssignUAV(geofenceID, uavID uint64) error {
	ug := &models.UAVGeofence{
		UAVID:      uavID,
		GeofenceID: geofenceID,
	}
	return r.db.Create(ug).Error
}

func (r *GeofenceRepository) UnassignUAV(geofenceID, uavID uint64) error {
	return r.db.Where("uav_id = ? AND geofence_id = ?", uavID, geofenceID).Delete(&models.UAVGeofence{}).Error
}

func (r *GeofenceRepository) UpdateStatus(id uint64, isActive bool) error {
	return r.db.Model(&models.Geofence{}).Where("id = ?", id).Update("is_active", isActive).Error
}
