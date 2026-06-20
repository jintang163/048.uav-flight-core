package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"

	"gorm.io/gorm"
)

type OTARepository struct {
	*BaseRepository
}

func NewOTARepository() *OTARepository {
	return &OTARepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *OTARepository) CreateFirmware(firmware *models.Firmware) error {
	firmware.UUID = utils.GenerateUUID()
	return r.db.Create(firmware).Error
}

func (r *OTARepository) FindFirmwareByID(id uint64) (*models.Firmware, error) {
	var firmware models.Firmware
	if err := r.db.First(&firmware, id).Error; err != nil {
		return nil, err
	}
	return &firmware, nil
}

func (r *OTARepository) FindFirmwareByUUID(uuid string) (*models.Firmware, error) {
	var firmware models.Firmware
	if err := r.db.Where("uuid = ?", uuid).First(&firmware).Error; err != nil {
		return nil, err
	}
	return &firmware, nil
}

func (r *OTARepository) ListFirmwares(pagination *utils.Pagination, fwType models.FirmwareType, status models.FirmwareStatus, hardware string) ([]models.Firmware, int64, error) {
	var firmwares []models.Firmware
	query := r.db.Model(&models.Firmware{})
	if fwType != "" {
		query = query.Where("type = ?", fwType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if hardware != "" {
		query = query.Where("hardware LIKE ?", "%"+hardware+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&firmwares).Error; err != nil {
		return nil, 0, err
	}
	return firmwares, total, nil
}

func (r *OTARepository) GetLatestFirmware(fwType models.FirmwareType, hardware string) (*models.Firmware, error) {
	var firmware models.Firmware
	err := r.db.Where("type = ? AND hardware = ? AND status = ?",
		fwType, hardware, models.FirmwareStatusReleased).
		Order("version DESC").First(&firmware).Error
	if err != nil {
		return nil, err
	}
	return &firmware, nil
}

func (r *OTARepository) IncrementDownloadCount(id uint64) error {
	return r.db.Model(&models.Firmware{}).Where("id = ?", id).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *OTARepository) CreateUpdate(update *models.FirmwareUpdate) error {
	update.UUID = utils.GenerateUUID()
	return r.db.Create(update).Error
}

func (r *OTARepository) FindUpdateByID(id uint64) (*models.FirmwareUpdate, error) {
	var update models.FirmwareUpdate
	if err := r.db.Preload("UAV").Preload("Firmware").First(&update, id).Error; err != nil {
		return nil, err
	}
	return &update, nil
}

func (r *OTARepository) FindUpdateByUUID(uuid string) (*models.FirmwareUpdate, error) {
	var update models.FirmwareUpdate
	if err := r.db.Preload("UAV").Preload("Firmware").Where("uuid = ?", uuid).First(&update).Error; err != nil {
		return nil, err
	}
	return &update, nil
}

func (r *OTARepository) ListUpdates(pagination *utils.Pagination, uavID uint64, status string) ([]models.FirmwareUpdate, int64, error) {
	var updates []models.FirmwareUpdate
	query := r.db.Model(&models.FirmwareUpdate{}).Preload("UAV").Preload("Firmware")
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&updates).Error; err != nil {
		return nil, 0, err
	}
	return updates, total, nil
}

func (r *OTARepository) UpdateProgress(id uint64, progress int, status string) error {
	updates := map[string]interface{}{
		"progress": progress,
	}
	if status != "" {
		updates["status"] = status
	}
	return r.db.Model(&models.FirmwareUpdate{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OTARepository) CompleteUpdate(id uint64, success bool, errorMsg string) error {
	updates := map[string]interface{}{}
	if success {
		updates["status"] = "completed"
		updates["completed_at"] = gorm.Expr("NOW()")
	} else {
		updates["status"] = "failed"
		updates["error_message"] = errorMsg
		updates["completed_at"] = gorm.Expr("NOW()")
	}
	return r.db.Model(&models.FirmwareUpdate{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OTARepository) GetActiveUpdateByUAV(uavID uint64) (*models.FirmwareUpdate, error) {
	var update models.FirmwareUpdate
	err := r.db.Where("uav_id = ? AND status IN ?", uavID,
		[]string{"pending", "downloading", "installing"}).
		Preload("Firmware").Order("created_at DESC").First(&update).Error
	if err != nil {
		return nil, err
	}
	return &update, nil
}
