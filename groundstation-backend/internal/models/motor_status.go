package models

import (
	"time"

	"gorm.io/gorm"
)

type MotorStatusType string

const (
	MotorStatusNormal  MotorStatusType = "normal"
	MotorStatusWarning MotorStatusType = "warning"
	MotorStatusFault   MotorStatusType = "fault"
	MotorStatusOffline MotorStatusType = "offline"
)

type MotorStatus struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID       uint64          `gorm:"index;not null" json:"uav_id"`
	MotorIndex  int             `gorm:"not null" json:"motor_index"`
	Status      MotorStatusType `gorm:"type:varchar(20);default:'normal'" json:"status"`
	RPM         int             `json:"rpm"`
	Voltage     float64         `gorm:"type:decimal(8,4)" json:"voltage"`
	Current     float64         `gorm:"type:decimal(8,4)" json:"current"`
	Temperature int             `json:"temperature"`
	Throttle    float64         `gorm:"type:decimal(5,2)" json:"throttle"`
	FaultFlags  int             `json:"fault_flags"`
	ErrorCount  int             `json:"error_count"`
	Vendor      string          `gorm:"type:varchar(50)" json:"vendor"`
	Model       string          `gorm:"type:varchar(50)" json:"model"`
	ErrorCode   int             `json:"error_code"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (MotorStatus) TableName() string {
	return "motor_status"
}

type MotorFailureEvent struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID         uint64         `gorm:"index;not null" json:"uav_id"`
	MotorIndex    int            `gorm:"not null" json:"motor_index"`
	FaultFlags    int            `json:"fault_flags"`
	ErrorCode     int            `json:"error_code"`
	RPMAtFailure  int            `json:"rpm_at_failure"`
	TempAtFailure int            `json:"temp_at_failure"`
	AlertID       uint64         `json:"alert_id"`
	ActionTaken   string         `gorm:"type:varchar(50)" json:"action_taken"`
	Resolved      bool           `gorm:"default:false" json:"resolved"`
	ResolvedAt    *time.Time     `json:"resolved_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MotorFailureEvent) TableName() string {
	return "motor_failure_events"
}
