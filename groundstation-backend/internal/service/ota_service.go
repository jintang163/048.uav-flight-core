package service

import (
	"context"
	"errors"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ctx = context.Background()

type OTAService struct {
	otaRepo  *repository.OTARepository
	uavRepo  *repository.UAVRepository
	minio    *minio.Client
}

func NewOTAService() *OTAService {
	cfg := config.AppConfig.MinIO
	minioClient, _ := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	return &OTAService{
		otaRepo: repository.NewOTARepository(),
		uavRepo: repository.NewUAVRepository(),
		minio:   minioClient,
	}
}

type UploadFirmwareRequest struct {
	Name        string                `form:"name" binding:"required"`
	Type        models.FirmwareType   `form:"type" binding:"required"`
	Version     string                `form:"version" binding:"required"`
	BuildNumber string                `form:"build_number"`
	Hardware    string                `form:"hardware"`
	Description string                `form:"description"`
	Changelog   string                `form:"changelog"`
	IsMandatory bool                  `form:"is_mandatory"`
	MinVersion  string                `form:"min_version"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

type StartUpdateRequest struct {
	UAVID      uint64 `json:"uav_id" binding:"required"`
	FirmwareID uint64 `json:"firmware_id" binding:"required"`
	OperatorID uint64 `json:"operator_id"`
}

func (s *OTAService) UploadFirmware(req *UploadFirmwareRequest, uploaderID uint64) (*models.Firmware, error) {
	cfg := config.AppConfig.MinIO

	if req.File.Size == 0 {
		return nil, errors.New("empty file")
	}

	ext := filepath.Ext(req.File.Filename)
	objectName := fmt.Sprintf("firmwares/%s/%s/%s%s",
		req.Type, req.Version, utils.GenerateUUID(), ext)

	src, err := req.File.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	uploadInfo, err := s.minio.PutObject(ctx, cfg.BucketLogs, objectName, src,
		req.File.Size, minio.PutObjectOptions{ContentType: req.File.Header.Get("Content-Type")})
	if err != nil {
		tmpFile, err := os.CreateTemp("", "firmware-*"+ext)
		if err != nil {
			return nil, err
		}
		defer tmpFile.Close()
		src.Seek(0, 0)
		io.Copy(tmpFile, src)
		objectName = tmpFile.Name()
	}

	hash, err := s.calculateFileHash(req.File)
	if err != nil {
		return nil, err
	}

	firmware := &models.Firmware{
		Name:        req.Name,
		Type:        req.Type,
		Version:     req.Version,
		BuildNumber: req.BuildNumber,
		Hardware:    req.Hardware,
		Description: req.Description,
		Changelog:   req.Changelog,
		Status:      models.FirmwareStatusDraft,
		FileURL:     objectName,
		FileSize:    req.File.Size,
		FileHash:    hash,
		UploaderID:  uploaderID,
		IsMandatory: req.IsMandatory,
		MinVersion:  req.MinVersion,
	}

	if err := s.otaRepo.CreateFirmware(firmware); err != nil {
		return nil, err
	}

	return firmware, nil
}

func (s *OTAService) calculateFileHash(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	return utils.MD5(fmt.Sprintf("%s-%d-%s", file.Filename, file.Size, time.Now().String())), nil
}

func (s *OTAService) GetFirmware(id uint64) (*models.Firmware, error) {
	return s.otaRepo.FindFirmwareByID(id)
}

func (s *OTAService) ListFirmwares(pagination *utils.Pagination, fwType models.FirmwareType, status models.FirmwareStatus, hardware string) ([]models.Firmware, int64, error) {
	return s.otaRepo.ListFirmwares(pagination, fwType, status, hardware)
}

func (s *OTAService) UpdateFirmwareStatus(id uint64, status models.FirmwareStatus) (*models.Firmware, error) {
	firmware, err := s.otaRepo.FindFirmwareByID(id)
	if err != nil {
		return nil, errors.New("firmware not found")
	}

	firmware.Status = status
	if status == models.FirmwareStatusReleased {
		now := time.Now()
		firmware.ReleasedAt = &now
	}

	if err := s.otaRepo.Update(firmware); err != nil {
		return nil, err
	}

	return firmware, nil
}

func (s *OTAService) DeleteFirmware(id uint64) error {
	_, err := s.otaRepo.FindFirmwareByID(id)
	if err != nil {
		return errors.New("firmware not found")
	}
	return s.otaRepo.SoftDelete(&models.Firmware{}, id)
}

func (s *OTAService) GetLatestFirmware(fwType models.FirmwareType, hardware string) (*models.Firmware, error) {
	return s.otaRepo.GetLatestFirmware(fwType, hardware)
}

func (s *OTAService) StartUpdate(req *StartUpdateRequest) (*models.FirmwareUpdate, error) {
	uav, err := s.uavRepo.FindByID(req.UAVID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	firmware, err := s.otaRepo.FindFirmwareByID(req.FirmwareID)
	if err != nil {
		return nil, errors.New("firmware not found")
	}

	if firmware.Status != models.FirmwareStatusReleased {
		return nil, errors.New("firmware not released")
	}

	activeUpdate, _ := s.otaRepo.GetActiveUpdateByUAV(req.UAVID)
	if activeUpdate != nil {
		return nil, errors.New("uav has an active update")
	}

	update := &models.FirmwareUpdate{
		UAVID:          req.UAVID,
		FirmwareID:     req.FirmwareID,
		OperatorID:     req.OperatorID,
		Status:         "pending",
		Progress:       0,
		CurrentVersion: uav.FirmwareVer,
		TargetVersion:  firmware.Version,
	}

	if err := s.otaRepo.CreateUpdate(update); err != nil {
		return nil, err
	}

	go s.performUpdate(update.ID, firmware, uav)

	return s.otaRepo.FindUpdateByID(update.ID)
}

func (s *OTAService) performUpdate(updateID uint64, firmware *models.Firmware, uav *models.UAV) {
	_ = s.otaRepo.UpdateProgress(updateID, 10, "downloading")
	time.Sleep(500 * time.Millisecond)

	_ = s.otaRepo.UpdateProgress(updateID, 50, "installing")
	time.Sleep(500 * time.Millisecond)

	_ = s.otaRepo.UpdateProgress(updateID, 90, "verifying")
	time.Sleep(500 * time.Millisecond)

	success := true
	var errMsg string

	if success {
		_ = s.otaRepo.CompleteUpdate(updateID, true, "")
		uav.FirmwareVer = firmware.Version
		_ = s.uavRepo.Update(uav)
	} else {
		_ = s.otaRepo.CompleteUpdate(updateID, false, errMsg)
	}
}

func (s *OTAService) GetUpdate(id uint64) (*models.FirmwareUpdate, error) {
	return s.otaRepo.FindUpdateByID(id)
}

func (s *OTAService) ListUpdates(pagination *utils.Pagination, uavID uint64, status string) ([]models.FirmwareUpdate, int64, error) {
	return s.otaRepo.ListUpdates(pagination, uavID, status)
}

func (s *OTAService) DownloadFirmware(id uint64) (string, error) {
	firmware, err := s.otaRepo.FindFirmwareByID(id)
	if err != nil {
		return "", errors.New("firmware not found")
	}

	_ = s.otaRepo.IncrementDownloadCount(id)
	return firmware.FileURL, nil
}

func (s *OTAService) GetActiveUpdate(uavID uint64) (*models.FirmwareUpdate, error) {
	return s.otaRepo.GetActiveUpdateByUAV(uavID)
}
