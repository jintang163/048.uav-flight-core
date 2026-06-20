package models

import (
	"time"

	"gorm.io/gorm"
)

type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
	AlertLevelFatal    AlertLevel = "fatal"
)

type AlertType string

const (
	AlertTypeLowBattery      AlertType = "low_battery"
	AlertTypeSignalLoss      AlertType = "signal_loss"
	AlertTypeGeofenceBreach  AlertType = "geofence_breach"
	AlertTypeConnectionLost  AlertType = "connection_lost"
	AlertTypeGPSLost         AlertType = "gps_lost"
	AlertTypeMotorError      AlertType = "motor_error"
	AlertTypeTemperatureHigh AlertType = "temperature_high"
	AlertTypeMissionAborted  AlertType = "mission_aborted"
	AlertTypeCustom          AlertType = "custom"
)

type AlertStatus string

const (
	AlertStatusNew       AlertStatus = "new"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved  AlertStatus = "resolved"
	AlertStatusIgnored   AlertStatus = "ignored"
)

type AlertEvent struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID         uint64         `gorm:"index;not null" json:"uav_id"`
	Type          AlertType      `gorm:"type:varchar(50);not null" json:"type"`
	Level         AlertLevel     `gorm:"type:varchar(20);not null" json:"level"`
	Status        AlertStatus    `gorm:"type:varchar(20);default:'new'" json:"status"`
	Title         string         `gorm:"type:varchar(200);not null" json:"title"`
	Message       string         `gorm:"type:text;not null" json:"message"`
	Latitude      float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude     float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude      float64        `gorm:"type:decimal(8,2)" json:"altitude"`
	BatteryLevel  float64        `gorm:"type:decimal(5,2)" json:"battery_level"`
	SignalStrength int           `json:"signal_strength"`
	AcknowledgedBy *uint64       `json:"acknowledged_by"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at"`
	ResolvedBy    *uint64        `json:"resolved_by"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	ResolvedNote  string         `gorm:"type:text" json:"resolved_note"`
	NotificationSent bool        `gorm:"default:false" json:"notification_sent"`
	SMSSent       bool           `gorm:"default:false" json:"sms_sent"`
	EmailSent     bool           `gorm:"default:false" json:"email_sent"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (AlertEvent) TableName() string {
	return "alert_events"
}

type AlertContact struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone"`
	Email     string         `gorm:"type:varchar(100)" json:"email"`
	AlertLevel AlertLevel    `gorm:"type:varchar(20);default:'warning'" json:"alert_level"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AlertContact) TableName() string {
	return "alert_contacts"
}
