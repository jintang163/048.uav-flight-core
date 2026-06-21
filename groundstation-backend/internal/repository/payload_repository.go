package repository

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/pkg/utils"
)

type PayloadRepository struct {
	*BaseRepository
}

func NewPayloadRepository() *PayloadRepository {
	return &PayloadRepository{
		BaseRepository: NewBaseRepository(),
	}
}

func (r *PayloadRepository) CreatePayload(payload *models.PayloadDevice) error {
	return r.db.Create(payload).Error
}

func (r *PayloadRepository) FindPayloadByID(id uint64) (*models.PayloadDevice, error) {
	var payload models.PayloadDevice
	if err := r.db.Preload("UAV").First(&payload, id).Error; err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *PayloadRepository) FindPayloadByUUID(uuid string) (*models.PayloadDevice, error) {
	var payload models.PayloadDevice
	if err := r.db.Where("uuid = ?", uuid).Preload("UAV").First(&payload).Error; err != nil {
		return nil, err
	}
	return &payload, nil
}

func (r *PayloadRepository) ListPayloadsByUAV(uavID uint64) ([]models.PayloadDevice, error) {
	var payloads []models.PayloadDevice
	if err := r.db.Where("uav_id = ?", uavID).Find(&payloads).Error; err != nil {
		return nil, err
	}
	return payloads, nil
}

func (r *PayloadRepository) ListPayloadsByUAVAndType(uavID uint64, payloadType models.PayloadType) ([]models.PayloadDevice, error) {
	var payloads []models.PayloadDevice
	if err := r.db.Where("uav_id = ? AND type = ?", uavID, payloadType).Find(&payloads).Error; err != nil {
		return nil, err
	}
	return payloads, nil
}

func (r *PayloadRepository) ListPayloads(pagination *utils.Pagination, uavID uint64, payloadType string, status string, keyword string) ([]models.PayloadDevice, int64, error) {
	var payloads []models.PayloadDevice
	var total int64

	query := r.db.Model(&models.PayloadDevice{})
	if uavID > 0 {
		query = query.Where("uav_id = ?", uavID)
	}
	if payloadType != "" {
		query = query.Where("type = ?", payloadType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR model LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	if err := query.Preload("UAV").Order("created_at DESC").Find(&payloads).Error; err != nil {
		return nil, 0, err
	}

	return payloads, total, nil
}

func (r *PayloadRepository) UpdatePayload(payload *models.PayloadDevice) error {
	return r.db.Save(payload).Error
}

func (r *PayloadRepository) UpdatePayloadStatus(id uint64, status models.PayloadStatus) error {
	return r.db.Model(&models.PayloadDevice{}).Where("id = ?", id).
		Update("status", status).Error
}

func (r *PayloadRepository) DeletePayload(id uint64) error {
	return r.db.Delete(&models.PayloadDevice{}, id).Error
}

func (r *PayloadRepository) GetCameraStatus(payloadID uint64) (*models.CameraStatus, error) {
	var status models.CameraStatus
	if err := r.db.Where("payload_id = ?", payloadID).First(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *PayloadRepository) UpsertCameraStatus(status *models.CameraStatus) error {
	var existing models.CameraStatus
	err := r.db.Where("payload_id = ?", status.PayloadID).First(&existing).Error
	if err != nil {
		return r.db.Create(status).Error
	}
	status.ID = existing.ID
	return r.db.Save(status).Error
}

func (r *PayloadRepository) GetSprayerStatus(payloadID uint64) (*models.SprayerStatus, error) {
	var status models.SprayerStatus
	if err := r.db.Where("payload_id = ?", payloadID).First(&status).Error; err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *PayloadRepository) UpsertSprayerStatus(status *models.SprayerStatus) error {
	var existing models.SprayerStatus
	err := r.db.Where("payload_id = ?", status.PayloadID).First(&existing).Error
	if err != nil {
		return r.db.Create(status).Error
	}
	status.ID = existing.ID
	return r.db.Save(status).Error
}

func (r *PayloadRepository) CreateSpeakerAudio(audio *models.SpeakerAudio) error {
	return r.db.Create(audio).Error
}

func (r *PayloadRepository) ListSpeakerAudios(pagination *utils.Pagination, payloadID uint64, isTTS *bool) ([]models.SpeakerAudio, int64, error) {
	var audios []models.SpeakerAudio
	var total int64

	query := r.db.Model(&models.SpeakerAudio{})
	if payloadID > 0 {
		query = query.Where("payload_id = ?", payloadID)
	}
	if isTTS != nil {
		query = query.Where("is_text_to_speech = ?", *isTTS)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if pagination != nil {
		query = query.Offset(pagination.Offset()).Limit(pagination.Limit())
	}

	if err := query.Order("created_at DESC").Find(&audios).Error; err != nil {
		return nil, 0, err
	}

	return audios, total, nil
}

func (r *PayloadRepository) FindSpeakerAudioByID(id uint64) (*models.SpeakerAudio, error) {
	var audio models.SpeakerAudio
	if err := r.db.First(&audio, id).Error; err != nil {
		return nil, err
	}
	return &audio, nil
}

func (r *PayloadRepository) DeleteSpeakerAudio(id uint64) error {
	return r.db.Delete(&models.SpeakerAudio{}, id).Error
}

func (r *PayloadRepository) InsertPayloadTelemetry(telemetry *models.PayloadTelemetry) error {
	return r.db.Create(telemetry).Error
}
