package service

import (
	"errors"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"time"
)

type TemporaryUnlockingService struct {
	unlockingRepo *repository.TemporaryUnlockingRepository
	geofenceRepo  *repository.GeofenceRepository
	uavRepo       *repository.UAVRepository
}

func NewTemporaryUnlockingService() *TemporaryUnlockingService {
	return &TemporaryUnlockingService{
		unlockingRepo: repository.NewTemporaryUnlockingRepository(),
		geofenceRepo:  repository.NewGeofenceRepository(),
		uavRepo:       repository.NewUAVRepository(),
	}
}

func (s *TemporaryUnlockingService) Apply(unlocking *models.TemporaryUnlocking, applicantID uint64) (*models.TemporaryUnlocking, error) {
	_, err := s.uavRepo.FindByID(unlocking.UAVID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	if unlocking.StartTime == nil {
		now := time.Now()
		unlocking.StartTime = &now
	}
	if unlocking.EndTime == nil {
		end := unlocking.StartTime.Add(2 * time.Hour)
		unlocking.EndTime = &end
	}

	unlocking.ApplicantID = applicantID
	unlocking.Status = models.UnlockStatusPending

	if err := s.unlockingRepo.Create(unlocking); err != nil {
		return nil, err
	}

	return s.unlockingRepo.FindByID(unlocking.ID)
}

func (s *TemporaryUnlockingService) GetByID(id uint64) (*models.TemporaryUnlocking, error) {
	return s.unlockingRepo.FindByID(id)
}

func (s *TemporaryUnlockingService) GetByUUID(uuid string) (*models.TemporaryUnlocking, error) {
	return s.unlockingRepo.FindByUUID(uuid)
}

func (s *TemporaryUnlockingService) List(pagination *utils.Pagination, uavID uint64, applicantID uint64, status string, category string, startTime string, endTime string) ([]models.TemporaryUnlocking, int64, error) {
	return s.unlockingRepo.List(pagination, uavID, applicantID, status, category, startTime, endTime)
}

func (s *TemporaryUnlockingService) Approve(id uint64, approverID uint64, remark string) error {
	unlocking, err := s.unlockingRepo.FindByID(id)
	if err != nil {
		return errors.New("application not found")
	}
	if unlocking.Status != models.UnlockStatusPending {
		return errors.New("only pending applications can be approved")
	}
	return s.unlockingRepo.Approve(id, approverID, remark)
}

func (s *TemporaryUnlockingService) Reject(id uint64, approverID uint64, remark string) error {
	unlocking, err := s.unlockingRepo.FindByID(id)
	if err != nil {
		return errors.New("application not found")
	}
	if unlocking.Status != models.UnlockStatusPending {
		return errors.New("only pending applications can be rejected")
	}
	return s.unlockingRepo.Reject(id, approverID, remark)
}

func (s *TemporaryUnlockingService) Cancel(id uint64, applicantID uint64) error {
	unlocking, err := s.unlockingRepo.FindByID(id)
	if err != nil {
		return errors.New("application not found")
	}
	if unlocking.ApplicantID != applicantID {
		return errors.New("you can only cancel your own application")
	}
	if unlocking.Status != models.UnlockStatusPending && unlocking.Status != models.UnlockStatusApproved {
		return errors.New("application cannot be cancelled")
	}
	return s.unlockingRepo.Cancel(id)
}

func (s *TemporaryUnlockingService) GetActiveUnlockings(uavID uint64, category string) ([]models.TemporaryUnlocking, error) {
	return s.unlockingRepo.GetActiveUnlockings(uavID, category)
}

func (s *TemporaryUnlockingService) CheckActiveUnlock(uavID uint64, geofenceID uint64) (*models.TemporaryUnlocking, error) {
	return s.unlockingRepo.CheckActiveUnlock(uavID, geofenceID)
}

func (s *TemporaryUnlockingService) ExpireOld() (int64, error) {
	return s.unlockingRepo.ExpireOld()
}
