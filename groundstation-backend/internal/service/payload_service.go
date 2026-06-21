package service

import (
	"encoding/json"
	"errors"
	"time"

	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"

	"github.com/google/uuid"
)

type PayloadService struct {
	payloadRepo *repository.PayloadRepository
	uavRepo     *repository.UAVRepository
}

func NewPayloadService() *PayloadService {
	return &PayloadService{
		payloadRepo: repository.NewPayloadRepository(),
		uavRepo:     repository.NewUAVRepository(),
	}
}

func (s *PayloadService) CreatePayload(payload *models.PayloadDevice) (*models.PayloadDevice, error) {
	if _, err := s.uavRepo.FindByID(payload.UAVID); err != nil {
		return nil, errors.New("uav not found")
	}

	payload.UUID = uuid.New().String()
	payload.Status = models.PayloadStatusOffline

	if err := s.payloadRepo.CreatePayload(payload); err != nil {
		return nil, err
	}

	s.initializePayloadStatus(payload)

	return s.payloadRepo.FindPayloadByID(payload.ID)
}

func (s *PayloadService) initializePayloadStatus(payload *models.PayloadDevice) {
	switch payload.Type {
	case models.PayloadTypeCamera, models.PayloadTypeThermalCamera:
		status := &models.CameraStatus{
			PayloadID:       payload.ID,
			Mode:            models.CameraModeIdle,
			StorageFreeMB:   64000,
			StorageTotalMB:  64000,
			LensTemperature: 25.0,
		}
		_ = s.payloadRepo.UpsertCameraStatus(status)
	case models.PayloadTypeSprayer:
		status := &models.SprayerStatus{
			PayloadID:      payload.ID,
			TankCapacityL:  10.0,
			TankRemainingL: 10.0,
			NozzleCount:    4,
		}
		_ = s.payloadRepo.UpsertSprayerStatus(status)
	}
}

func (s *PayloadService) GetPayload(id uint64) (*models.PayloadDevice, error) {
	return s.payloadRepo.FindPayloadByID(id)
}

func (s *PayloadService) GetPayloadByUUID(uuidStr string) (*models.PayloadDevice, error) {
	return s.payloadRepo.FindPayloadByUUID(uuidStr)
}

func (s *PayloadService) ListPayloads(pagination *utils.Pagination, uavID uint64, payloadType string, status string, keyword string) ([]models.PayloadDevice, int64, error) {
	return s.payloadRepo.ListPayloads(pagination, uavID, payloadType, status, keyword)
}

func (s *PayloadService) ListPayloadsByUAV(uavID uint64) ([]models.PayloadDevice, error) {
	return s.payloadRepo.ListPayloadsByUAV(uavID)
}

func (s *PayloadService) UpdatePayload(payload *models.PayloadDevice) error {
	return s.payloadRepo.UpdatePayload(payload)
}

func (s *PayloadService) UpdatePayloadByID(id uint64, payload *models.PayloadDevice) (*models.PayloadDevice, error) {
	existing, err := s.payloadRepo.FindPayloadByID(id)
	if err != nil {
		return nil, errors.New("payload not found")
	}

	if payload.Name != "" {
		existing.Name = payload.Name
	}
	if payload.Model != "" {
		existing.Model = payload.Model
	}
	if payload.Description != "" {
		existing.Description = payload.Description
	}
	if payload.Config != "" {
		existing.Config = payload.Config
	}
	if payload.Port > 0 {
		existing.Port = payload.Port
	}

	if err := s.payloadRepo.UpdatePayload(existing); err != nil {
		return nil, err
	}

	return s.payloadRepo.FindPayloadByID(id)
}

func (s *PayloadService) UpdateCameraStatus(payloadID uint64, status *models.CameraStatus) error {
	existing, err := s.payloadRepo.GetCameraStatus(payloadID)
	if err != nil || existing == nil {
		status.PayloadID = payloadID
		return s.payloadRepo.UpsertCameraStatus(status)
	}
	status.ID = existing.ID
	status.PayloadID = payloadID
	return s.payloadRepo.UpsertCameraStatus(status)
}

func (s *PayloadService) UpdateSprayerStatus(payloadID uint64, status *models.SprayerStatus) error {
	existing, err := s.payloadRepo.GetSprayerStatus(payloadID)
	if err != nil || existing == nil {
		status.PayloadID = payloadID
		return s.payloadRepo.UpsertSprayerStatus(status)
	}
	status.ID = existing.ID
	status.PayloadID = payloadID
	return s.payloadRepo.UpsertSprayerStatus(status)
}

func (s *PayloadService) IncrementPhotoCount(payloadID uint64) error {
	status, err := s.payloadRepo.GetCameraStatus(payloadID)
	if err != nil || status == nil {
		return err
	}
	status.PhotoCount++
	now := time.Now()
	status.LastCaptureAt = &now
	return s.payloadRepo.UpsertCameraStatus(status)
}

func (s *PayloadService) UpdateSprayerFlowRate(payloadID uint64, flowRate float64) error {
	status, err := s.payloadRepo.GetSprayerStatus(payloadID)
	if err != nil || status == nil {
		return nil
	}
	status.FlowRate = flowRate
	status.TargetFlowRate = flowRate
	status.IsSpraying = flowRate > 0
	return s.payloadRepo.UpsertSprayerStatus(status)
}

func (s *PayloadService) UpdateSprayerPressure(payloadID uint64, pressure float64) error {
	status, err := s.payloadRepo.GetSprayerStatus(payloadID)
	if err != nil || status == nil {
		return nil
	}
	status.Pressure = pressure
	return s.payloadRepo.UpsertSprayerStatus(status)
}

func (s *PayloadService) UpdateCameraLensTemp(payloadID uint64, temp float64) error {
	status, err := s.payloadRepo.GetCameraStatus(payloadID)
	if err != nil || status == nil {
		return nil
	}
	status.LensTemperature = temp
	status.LensTemperatureC = temp
	return s.payloadRepo.UpsertCameraStatus(status)
}

func (s *PayloadService) UpdateCameraZoom(payloadID uint64, zoom float64) error {
	status, err := s.payloadRepo.GetCameraStatus(payloadID)
	if err != nil || status == nil {
		return nil
	}
	status.ZoomLevel = zoom
	return s.payloadRepo.UpsertCameraStatus(status)
}

func (s *PayloadService) UpdatePayloadStatus(id uint64, status models.PayloadStatus) error {
	payload, err := s.payloadRepo.FindPayloadByID(id)
	if err != nil {
		return errors.New("payload not found")
	}

	now := time.Now()
	payload.Status = status
	if status == models.PayloadStatusActive || status == models.PayloadStatusOnline {
		payload.LastActiveAt = &now
	}

	return s.payloadRepo.UpdatePayload(payload)
}

func (s *PayloadService) DeletePayload(id uint64) error {
	_, err := s.payloadRepo.FindPayloadByID(id)
	if err != nil {
		return errors.New("payload not found")
	}
	return s.payloadRepo.DeletePayload(id)
}

func (s *PayloadService) GetCameraStatus(payloadID uint64) (*models.CameraStatus, error) {
	return s.payloadRepo.GetCameraStatus(payloadID)
}

func (s *PayloadService) TakePhoto(uavID uint64, payloadID uint64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}
	if payload.Type != models.PayloadTypeCamera && payload.Type != models.PayloadTypeThermalCamera {
		return errors.New("payload is not a camera")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_DIGICAM_CONTROL, 0, 1, 0, 0, 0, 0, 0)
	if err := cmdMgr.SendCommand(uavID, data); err != nil {
		return err
	}

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		status.Mode = models.CameraModePhoto
		status.PhotoCount++
		now := time.Now()
		status.LastCaptureAt = &now
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "photo_taken", nil)

	return nil
}

func (s *PayloadService) StartVideoRecording(uavID uint64, payloadID uint64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}
	if payload.Type != models.PayloadTypeCamera && payload.Type != models.PayloadTypeThermalCamera {
		return errors.New("payload is not a camera")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_VIDEO_START, 0, 0, 0, 0, 0, 0, 0)
	if err := cmdMgr.SendCommand(uavID, data); err != nil {
		return err
	}

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		status.Mode = models.CameraModeVideo
		status.IsRecording = true
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "recording_started", nil)

	return nil
}

func (s *PayloadService) StopVideoRecording(uavID uint64, payloadID uint64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_VIDEO_STOP, 0, 0, 0, 0, 0, 0, 0)
	if err := cmdMgr.SendCommand(uavID, data); err != nil {
		return err
	}

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		status.IsRecording = false
		status.Mode = models.CameraModeIdle
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "recording_stopped", nil)

	return nil
}

func (s *PayloadService) SetCameraMode(uavID uint64, payloadID uint64, mode models.CameraMode) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		status.Mode = mode
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "mode_changed", map[string]interface{}{
		"mode": mode,
	})

	return nil
}

func (s *PayloadService) SetCameraZoom(uavID uint64, payloadID uint64, zoomLevel float64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_DIGICAM_CONFIGURE, 4, float32(zoomLevel), 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(uavID, data)

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		status.ZoomLevel = zoomLevel
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "zoom_changed", map[string]interface{}{
		"zoom_level": zoomLevel,
	})

	return nil
}

func (s *PayloadService) SetCameraSettings(uavID uint64, payloadID uint64, settings map[string]interface{}) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}

	status, _ := s.payloadRepo.GetCameraStatus(payloadID)
	if status != nil {
		if resolution, ok := settings["resolution"].(string); ok {
			status.Resolution = resolution
		}
		if frameRate, ok := settings["frame_rate"].(float64); ok {
			status.FrameRate = int(frameRate)
		}
		if iso, ok := settings["iso"].(float64); ok {
			status.ISO = int(iso)
		}
		if shutterSpeed, ok := settings["shutter_speed"].(string); ok {
			status.ShutterSpeed = shutterSpeed
		}
		_ = s.payloadRepo.UpsertCameraStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "settings_updated", settings)

	return nil
}

func (s *PayloadService) GetSprayerStatus(payloadID uint64) (*models.SprayerStatus, error) {
	return s.payloadRepo.GetSprayerStatus(payloadID)
}

func (s *PayloadService) SetSprayerFlowRate(uavID uint64, payloadID uint64, flowRate float64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}
	if payload.Type != models.PayloadTypeSprayer {
		return errors.New("payload is not a sprayer")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_SPRAYER, 1, float32(flowRate), 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(uavID, data)

	status, _ := s.payloadRepo.GetSprayerStatus(payloadID)
	if status != nil {
		status.TargetFlowRate = flowRate
		status.FlowRate = flowRate
		status.IsSpraying = flowRate > 0
		_ = s.payloadRepo.UpsertSprayerStatus(status)
	}

	s.broadcastPayloadStatus(uavID, payloadID, "flow_rate_changed", map[string]interface{}{
		"flow_rate_lpm": flowRate,
		"is_spraying":   flowRate > 0,
	})

	return nil
}

func (s *PayloadService) StartSpraying(uavID uint64, payloadID uint64, flowRate float64) error {
	if flowRate <= 0 {
		flowRate = 2.0
	}
	return s.SetSprayerFlowRate(uavID, payloadID, flowRate)
}

func (s *PayloadService) StopSpraying(uavID uint64, payloadID uint64) error {
	return s.SetSprayerFlowRate(uavID, payloadID, 0)
}

func (s *PayloadService) CreateSpeakerAudio(audio *models.SpeakerAudio, creatorID uint64) (*models.SpeakerAudio, error) {
	audio.UUID = uuid.New().String()
	audio.CreatedBy = creatorID
	if err := s.payloadRepo.CreateSpeakerAudio(audio); err != nil {
		return nil, err
	}
	return s.payloadRepo.FindSpeakerAudioByID(audio.ID)
}

func (s *PayloadService) ListSpeakerAudios(pagination *utils.Pagination, payloadID uint64, isTTS *bool) ([]models.SpeakerAudio, int64, error) {
	return s.payloadRepo.ListSpeakerAudios(pagination, payloadID, isTTS)
}

func (s *PayloadService) GetSpeakerAudio(id uint64) (*models.SpeakerAudio, error) {
	return s.payloadRepo.FindSpeakerAudioByID(id)
}

func (s *PayloadService) DeleteSpeakerAudio(id uint64) error {
	_, err := s.payloadRepo.FindSpeakerAudioByID(id)
	if err != nil {
		return errors.New("audio not found")
	}
	return s.payloadRepo.DeleteSpeakerAudio(id)
}

func (s *PayloadService) PlaySpeakerAudio(uavID uint64, payloadID uint64, audioID uint64) error {
	payload, err := s.payloadRepo.FindPayloadByID(payloadID)
	if err != nil {
		return errors.New("payload not found")
	}
	if payload.Type != models.PayloadTypeSpeaker {
		return errors.New("payload is not a speaker")
	}

	audio, err := s.payloadRepo.FindSpeakerAudioByID(audioID)
	if err != nil {
		return errors.New("audio not found")
	}

	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_PLAY_TUNE, float32(audioID), 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(uavID, data)

	s.broadcastPayloadStatus(uavID, payloadID, "audio_playing", map[string]interface{}{
		"audio_id":   audioID,
		"audio_name": audio.Name,
		"duration":   audio.DurationSec,
	})

	return nil
}

func (s *PayloadService) StopSpeaker(uavID uint64, payloadID uint64) error {
	cmdMgr := mavlink.NewCommandManager()
	data := mavlink.EncodeCommandLong(uavID, mavlink.CMD_DO_PLAY_TUNE, 0, 0, 0, 0, 0, 0, 0)
	_ = cmdMgr.SendCommand(uavID, data)

	s.broadcastPayloadStatus(uavID, payloadID, "audio_stopped", nil)

	return nil
}

func (s *PayloadService) UpdatePayloadTelemetry(uavID uint64, payloadID uint64, payloadType string, data interface{}) error {
	dataBytes, _ := json.Marshal(data)
	telemetry := &models.PayloadTelemetry{
		UAVID:       uavID,
		PayloadID:   payloadID,
		PayloadType: payloadType,
		Data:        string(dataBytes),
		Timestamp:   time.Now(),
	}
	return s.payloadRepo.InsertPayloadTelemetry(telemetry)
}

func (s *PayloadService) broadcastPayloadStatus(uavID uint64, payloadID uint64, event string, data interface{}) {
	payload := map[string]interface{}{
		"uav_id":     uavID,
		"payload_id": payloadID,
		"event":      event,
		"data":       data,
		"timestamp":  time.Now().UnixNano() / 1e6,
	}

	hub := websocket.NewHub()
	hub.BroadcastUAVTelemetry(uavID, map[string]interface{}{
		"type":    "payload_status",
		"payload": payload,
	})
}
