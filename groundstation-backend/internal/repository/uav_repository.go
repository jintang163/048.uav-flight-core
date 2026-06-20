package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type UAVRepository struct {
	*BaseRepository
}

func NewUAVRepository() *UAVRepository {
	return &UAVRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *UAVRepository) Create(uav *models.UAV) error {
	uav.UUID = utils.GenerateUUID()
	return r.BaseRepository.Create(uav)
}

func (r *UAVRepository) FindByID(id uint64) (*models.UAV, error) {
	var uav models.UAV
	if err := r.db.Preload("FlightStatus").First(&uav, id).Error; err != nil {
		return nil, err
	}
	return &uav, nil
}

func (r *UAVRepository) FindByUUID(uuid string) (*models.UAV, error) {
	var uav models.UAV
	if err := r.db.Preload("FlightStatus").Where("uuid = ?", uuid).First(&uav).Error; err != nil {
		return nil, err
	}
	return &uav, nil
}

func (r *UAVRepository) FindBySerialNumber(sn string) (*models.UAV, error) {
	var uav models.UAV
	if err := r.db.Where("serial_number = ?", sn).First(&uav).Error; err != nil {
		return nil, err
	}
	return &uav, nil
}

func (r *UAVRepository) List(pagination *utils.Pagination, status string, ownerID uint64) ([]models.UAV, int64, error) {
	var uavs []models.UAV
	query := r.db.Model(&models.UAV{}).Preload("FlightStatus")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if ownerID > 0 {
		query = query.Where("owner_id = ?", ownerID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&uavs).Error; err != nil {
		return nil, 0, err
	}
	return uavs, total, nil
}

func (r *UAVRepository) UpdateStatus(id uint64, status models.UAVStatus) error {
	return r.db.Model(&models.UAV{}).Where("id = ?", id).Update("status", status).Error
}

func (r *UAVRepository) UpdateLastSeen(id uint64) error {
	return r.db.Model(&models.UAV{}).Where("id = ?", id).Update("last_seen_at", gorm.Expr("NOW()")).Error
}

func (r *UAVRepository) FindByIP(ip string, port int) (*models.UAV, error) {
	var uav models.UAV
	if err := r.db.Where("ip_address = ? AND port = ?", ip, port).First(&uav).Error; err != nil {
		return nil, err
	}
	return &uav, nil
}

func (r *UAVRepository) CountByStatus(status models.UAVStatus) (int64, error) {
	return r.Count(&models.UAV{}, "status = ?", status)
}

func (r *UAVRepository) GetOnlineUAVs() ([]models.UAV, error) {
	var uavs []models.UAV
	err := r.db.Where("status IN ?", []models.UAVStatus{
		models.UAVStatusOnline,
		models.UAVStatusFlying,
		models.UAVStatusHovering,
	}).Preload("FlightStatus").Find(&uavs).Error
	return uavs, err
}
