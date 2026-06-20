package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"
)

type UAVService struct {
	uavRepo    *repository.UAVRepository
	flightRepo *repository.FlightRepository
}

func NewUAVService() *UAVService {
	return &UAVService{
		uavRepo:    repository.NewUAVRepository(),
		flightRepo: repository.NewFlightRepository(),
	}
}

type CreateUAVRequest struct {
	Name         string           `json:"name" binding:"required"`
	Model        string           `json:"model"`
	SerialNumber string           `json:"serial_number" binding:"required"`
	FirmwareVer  string           `json:"firmware_version"`
	IPAddress    string           `json:"ip_address"`
	Port         int              `json:"port"`
	Protocol     string           `json:"protocol"`
	Description  string           `json:"description"`
	OwnerID      uint64           `json:"owner_id"`
}

type UpdateUAVRequest struct {
	Name         string `json:"name"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
	FirmwareVer  string `json:"firmware_version"`
	IPAddress    string `json:"ip_address"`
	Port         int    `json:"port"`
	Protocol     string `json:"protocol"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

func (s *UAVService) Create(req *CreateUAVRequest) (*models.UAV, error) {
	existing, _ := s.uavRepo.FindBySerialNumber(req.SerialNumber)
	if existing != nil {
		return nil, errors.New("serial number already exists")
	}

	uav := &models.UAV{
		Name:         req.Name,
		Model:        req.Model,
		SerialNumber: req.SerialNumber,
		FirmwareVer:  req.FirmwareVer,
		IPAddress:    req.IPAddress,
		Port:         req.Port,
		Protocol:     req.Protocol,
		Description:  req.Description,
		OwnerID:      req.OwnerID,
		Status:       models.UAVStatusOffline,
	}

	if err := s.uavRepo.Create(uav); err != nil {
		return nil, err
	}

	return uav, nil
}

func (s *UAVService) GetByID(id uint64) (*models.UAV, error) {
	return s.uavRepo.FindByID(id)
}

func (s *UAVService) GetByUUID(uuid string) (*models.UAV, error) {
	return s.uavRepo.FindByUUID(uuid)
}

func (s *UAVService) List(pagination *utils.Pagination, status string, ownerID uint64) ([]models.UAV, int64, error) {
	return s.uavRepo.List(pagination, status, ownerID)
}

func (s *UAVService) Update(id uint64, req *UpdateUAVRequest) (*models.UAV, error) {
	uav, err := s.uavRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	if req.Name != "" {
		uav.Name = req.Name
	}
	if req.Model != "" {
		uav.Model = req.Model
	}
	if req.SerialNumber != "" && req.SerialNumber != uav.SerialNumber {
		existing, _ := s.uavRepo.FindBySerialNumber(req.SerialNumber)
		if existing != nil {
			return nil, errors.New("serial number already exists")
		}
		uav.SerialNumber = req.SerialNumber
	}
	if req.FirmwareVer != "" {
		uav.FirmwareVer = req.FirmwareVer
	}
	if req.IPAddress != "" {
		uav.IPAddress = req.IPAddress
	}
	if req.Port > 0 {
		uav.Port = req.Port
	}
	if req.Protocol != "" {
		uav.Protocol = req.Protocol
	}
	if req.Description != "" {
		uav.Description = req.Description
	}
	if req.Status != "" {
		uav.Status = models.UAVStatus(req.Status)
	}

	if err := s.uavRepo.Update(uav); err != nil {
		return nil, err
	}

	return uav, nil
}

func (s *UAVService) Delete(id uint64) error {
	_, err := s.uavRepo.FindByID(id)
	if err != nil {
		return errors.New("uav not found")
	}
	return s.uavRepo.SoftDelete(&models.UAV{}, id)
}

func (s *UAVService) UpdateStatus(id uint64, status models.UAVStatus) error {
	_, err := s.uavRepo.FindByID(id)
	if err != nil {
		return errors.New("uav not found")
	}
	return s.uavRepo.UpdateStatus(id, status)
}

func (s *UAVService) GetOnlineUAVs() ([]models.UAV, error) {
	return s.uavRepo.GetOnlineUAVs()
}

func (s *UAVService) GetStatus() (map[string]interface{}, error) {
	total, _ := s.uavRepo.Count(&models.UAV{}, nil)
	online, _ := s.uavRepo.CountByStatus(models.UAVStatusOnline)
	flying, _ := s.uavRepo.CountByStatus(models.UAVStatusFlying)
	hovering, _ := s.uavRepo.CountByStatus(models.UAVStatusHovering)
	offline, _ := s.uavRepo.CountByStatus(models.UAVStatusOffline)
	errorStatus, _ := s.uavRepo.CountByStatus(models.UAVStatusError)

	return map[string]interface{}{
		"total":    total,
		"online":   online,
		"flying":   flying,
		"hovering": hovering,
		"offline":  offline,
		"error":    errorStatus,
	}, nil
}

func (s *UAVService) GetLatestFlightStatus(uavID uint64) (*models.FlightStatus, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}
	return s.flightRepo.GetLatestStatus(uavID)
}

func (s *UAVService) GetFlightHistory(uavID uint64, pagination *utils.Pagination, startTime, endTime *int64) ([]models.FlightStatus, int64, error) {
	_, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, 0, errors.New("uav not found")
	}

	var start, end *utils.Time
	if startTime != nil {
		t := utils.ParseTime(string(*startTime))
		start = &t
	}
	if endTime != nil {
		t := utils.ParseTime(string(*endTime))
		end = &t
	}

	return s.flightRepo.GetStatusHistory(uavID, pagination, (*time.Time)(start), (*time.Time)(end))
}
