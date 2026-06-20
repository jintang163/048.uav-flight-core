package models

import (
	"time"

	"gorm.io/gorm"
)

type FirmwareStatus string

const (
	FirmwareStatusDraft     FirmwareStatus = "draft"
	FirmwareStatusTesting   FirmwareStatus = "testing"
	FirmwareStatusReleased  FirmwareStatus = "released"
	FirmwareStatusDeprecated FirmwareStatus = "deprecated"
)

type FirmwareType string

const (
	FirmwareTypeFlightController FirmwareType = "flight_controller"
	FirmwareTypeESC              FirmwareType = "esc"
	FirmwareTypeRadio            FirmwareType = "radio"
	FirmwareTypeGPS              FirmwareType = "gps"
	FirmwareTypeCamera           FirmwareType = "camera"
)

type Firmware struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Type        FirmwareType   `gorm:"type:varchar(30);not null" json:"type"`
	Version     string         `gorm:"type:varchar(50);not null" json:"version"`
	BuildNumber string         `gorm:"type:varchar(50)" json:"build_number"`
	Hardware    string         `gorm:"type:varchar(100)" json:"hardware"`
	Description string         `gorm:"type:text" json:"description"`
	Changelog   string         `gorm:"type:text" json:"changelog"`
	Status      FirmwareStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	FileURL     string         `gorm:"type:varchar(500)" json:"file_url"`
	FileSize    int64          `json:"file_size"`
	FileHash    string         `gorm:"type:varchar(64)" json:"file_hash"`
	UploaderID  uint64         `json:"uploader_id"`
	IsMandatory bool           `gorm:"default:false" json:"is_mandatory"`
	MinVersion  string         `gorm:"type:varchar(50)" json:"min_version"`
	ReleasedAt  *time.Time     `json:"released_at"`
	DownloadCount int          `gorm:"default:0" json:"download_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Firmware) TableName() string {
	return "firmwares"
}

type FirmwareUpdate struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID           string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID          uint64         `gorm:"index;not null" json:"uav_id"`
	FirmwareID     uint64         `gorm:"index;not null" json:"firmware_id"`
	OperatorID     uint64         `json:"operator_id"`
	Status         string         `gorm:"type:varchar(30);default:'pending'" json:"status"`
	Progress       int            `gorm:"default:0" json:"progress"`
	CurrentVersion string         `gorm:"type:varchar(50)" json:"current_version"`
	TargetVersion  string         `gorm:"type:varchar(50)" json:"target_version"`
	StartedAt      *time.Time     `json:"started_at"`
	CompletedAt    *time.Time     `json:"completed_at"`
	ErrorMessage   string         `gorm:"type:text" json:"error_message"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	UAV      *UAV      `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Firmware *Firmware `gorm:"foreignKey:FirmwareID" json:"firmware,omitempty"`
}

func (FirmwareUpdate) TableName() string {
	return "firmware_updates"
}
