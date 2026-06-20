package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

type AlertService struct {
	alertRepo *repository.AlertRepository
}

func NewAlertService() *AlertService {
	return &AlertService{
		alertRepo: repository.NewAlertRepository(),
	}
}

type AcknowledgeAlertRequest struct {
	Note string `json:"note"`
}

type ResolveAlertRequest struct {
	Note string `json:"note" binding:"required"`
}

type CreateContactRequest struct {
	Name       string           `json:"name" binding:"required"`
	Phone      string           `json:"phone"`
	Email      string           `json:"email"`
	AlertLevel models.AlertLevel `json:"alert_level"`
}

func (s *AlertService) GetAlert(id uint64) (*models.AlertEvent, error) {
	return s.alertRepo.FindByID(id)
}

func (s *AlertService) List(pagination *utils.Pagination, uavID uint64, level models.AlertLevel, status models.AlertStatus, alertType models.AlertType) ([]models.AlertEvent, int64, error) {
	return s.alertRepo.List(pagination, uavID, level, status, alertType)
}

func (s *AlertService) Acknowledge(id uint64, userID uint64) (*models.AlertEvent, error) {
	alert, err := s.alertRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("alert not found")
	}

	if alert.Status == models.AlertStatusAcknowledged || alert.Status == models.AlertStatusResolved {
		return nil, errors.New("alert already acknowledged or resolved")
	}

	err = s.alertRepo.UpdateStatus(id, models.AlertStatusAcknowledged, userID, "")
	if err != nil {
		return nil, err
	}

	return s.alertRepo.FindByID(id)
}

func (s *AlertService) Resolve(id uint64, userID uint64, note string) (*models.AlertEvent, error) {
	alert, err := s.alertRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("alert not found")
	}

	if alert.Status == models.AlertStatusResolved {
		return nil, errors.New("alert already resolved")
	}

	err = s.alertRepo.UpdateStatus(id, models.AlertStatusResolved, userID, note)
	if err != nil {
		return nil, err
	}

	return s.alertRepo.FindByID(id)
}

func (s *AlertService) Ignore(id uint64, userID uint64) (*models.AlertEvent, error) {
	_, err := s.alertRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("alert not found")
	}

	err = s.alertRepo.UpdateStatus(id, models.AlertStatusIgnored, userID, "")
	if err != nil {
		return nil, err
	}

	return s.alertRepo.FindByID(id)
}

func (s *AlertService) GetUnacknowledged(uavID uint64) ([]models.AlertEvent, error) {
	return s.alertRepo.GetUnacknowledgedAlerts(uavID)
}

func (s *AlertService) GetStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	return s.alertRepo.GetAlertStats(startTime, endTime)
}

func (s *AlertService) CreateContact(req *CreateContactRequest) (*models.AlertContact, error) {
	contact := &models.AlertContact{
		Name:       req.Name,
		Phone:      req.Phone,
		Email:      req.Email,
		AlertLevel: req.AlertLevel,
		IsActive:   true,
	}

	if err := s.alertRepo.CreateContact(contact); err != nil {
		return nil, err
	}

	return contact, nil
}

func (s *AlertService) SendNotifications(alert *models.AlertEvent) error {
	contacts, err := s.alertRepo.GetActiveContacts(alert.Level)
	if err != nil {
		return err
	}

	for _, contact := range contacts {
		if contact.Phone != "" {
			go s.sendSMS(contact.Phone, alert)
		}
		if contact.Email != "" {
			go s.sendEmail(contact.Email, alert)
		}
	}

	_ = s.alertRepo.MarkNotificationSent(alert.ID)
	return nil
}

func (s *AlertService) sendSMS(phone string, alert *models.AlertEvent) {
	cfg := config.AppConfig.Alert
	if cfg.SMSAPI == "" || cfg.SMSAPIKey == "" {
		return
	}

	message := fmt.Sprintf("[UAV告警] %s: %s. 位置: %.6f, %.6f",
		alert.Level, alert.Title, alert.Latitude, alert.Longitude)

	payload := map[string]interface{}{
		"phone":   phone,
		"message": message,
	}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", cfg.SMSAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.SMSAPIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		_ = s.alertRepo.MarkSMSSent(alert.ID)
	}
}

func (s *AlertService) sendEmail(email string, alert *models.AlertEvent) error {
	cfg := config.AppConfig.Alert
	if cfg.EmailSMTPHost == "" || cfg.EmailUser == "" || cfg.EmailPassword == "" {
		return errors.New("email not configured")
	}

	subject := fmt.Sprintf("[UAV告警][%s] %s", strings.ToUpper(string(alert.Level)), alert.Title)
	body := fmt.Sprintf(`
告警级别: %s
告警类型: %s
告警标题: %s
告警内容: %s

无人机ID: %d
位置: %.6f, %.6f
高度: %.2f m
电量: %.1f %%
信号强度: %d %%

时间: %s
`,
		alert.Level, alert.Type, alert.Title, alert.Message,
		alert.UAVID, alert.Latitude, alert.Longitude, alert.Altitude,
		alert.BatteryLevel, alert.SignalStrength,
		alert.CreatedAt.Format("2006-01-02 15:04:05"),
	)

	smtpAuth := smtp.PlainAuth("", cfg.EmailUser, cfg.EmailPassword, cfg.EmailSMTPHost)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.EmailUser, email, subject, body)

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", cfg.EmailSMTPHost, cfg.EmailSMTPPort),
		smtpAuth,
		cfg.EmailUser,
		[]string{email},
		[]byte(msg),
	)

	if err == nil {
		_ = s.alertRepo.MarkEmailSent(alert.ID)
	}

	return err
}

func (s *AlertService) CreateCustomAlert(uavID uint64, title, message string, level models.AlertLevel) (*models.AlertEvent, error) {
	alert := &models.AlertEvent{
		UAVID:   uavID,
		Type:    models.AlertTypeCustom,
		Level:   level,
		Title:   title,
		Message: message,
	}

	if err := s.alertRepo.Create(alert); err != nil {
		return nil, err
	}

	go s.SendNotifications(alert)

	return alert, nil
}
