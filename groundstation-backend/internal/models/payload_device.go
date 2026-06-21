package models

import (
	"time"

	"gorm.io/gorm"
)

type PayloadType string

const (
	PayloadTypeCamera        PayloadType = "camera"
	PayloadTypeThermalCamera PayloadType = "thermal_camera"
	PayloadTypeSpeaker       PayloadType = "speaker"
	PayloadTypeSprayer       PayloadType = "sprayer"
)

type PayloadStatus string

const (
	PayloadStatusOffline  PayloadStatus = "offline"
	PayloadStatusOnline   PayloadStatus = "online"
	PayloadStatusActive   PayloadStatus = "active"
	PayloadStatusError    PayloadStatus = "error"
)

type CameraMode string

const (
	CameraModePhoto CameraMode = "photo"
	CameraModeVideo CameraMode = "video"
	CameraModeIdle  CameraMode = "idle"
)

type PayloadDevice struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID         string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID        uint64         `gorm:"index;not null" json:"uav_id"`
	Type         PayloadType    `gorm:"type:varchar(30);not null" json:"type"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Model        string         `gorm:"type:varchar(50)" json:"model"`
	Status       PayloadStatus  `gorm:"type:varchar(20);default:'offline'" json:"status"`
	Port         int            `json:"port"`
	Slot         int            `json:"slot"`
	FirmwareVer  string         `gorm:"type:varchar(50)" json:"firmware_version"`
	Description  string         `gorm:"type:text" json:"description"`
	Config       string         `gorm:"type:text" json:"config,omitempty"`
	LastActiveAt *time.Time     `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (PayloadDevice) TableName() string {
	return "payload_devices"
}

type CameraStatus struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PayloadID        uint64         `gorm:"index;not null" json:"payload_id"`
	Mode             CameraMode     `gorm:"type:varchar(20);default:'idle'" json:"mode"`
	IsRecording      bool           `gorm:"default:false" json:"is_recording"`
	StorageTotalMB   float64        `gorm:"type:decimal(12,2)" json:"storage_total_mb"`
	StorageUsedMB    float64        `gorm:"type:decimal(12,2)" json:"storage_used_mb"`
	StorageFreeMB    float64        `gorm:"type:decimal(12,2)" json:"storage_free_mb"`
	LensTemperature  float64        `gorm:"type:decimal(5,2)" json:"lens_temperature"`
	SensorTemperature float64       `gorm:"type:decimal(5,2)" json:"sensor_temperature"`
	PhotoCount       int            `gorm:"default:0" json:"photo_count"`
	VideoDurationSec int            `gorm:"default:0" json:"video_duration_sec"`
	Resolution       string         `gorm:"type:varchar(20)" json:"resolution"`
	FrameRate        int            `json:"frame_rate"`
	ZoomLevel        float64        `gorm:"type:decimal(5,2)" json:"zoom_level"`
	ISO              int            `json:"iso"`
	ShutterSpeed     string         `gorm:"type:varchar(20)" json:"shutter_speed"`
	LastCaptureAt    *time.Time     `json:"last_capture_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Payload *PayloadDevice `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
}

func (CameraStatus) TableName() string {
	return "camera_statuses"
}

type SpeakerAudio struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	PayloadID   uint64         `gorm:"index;not null" json:"payload_id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Type        string         `gorm:"type:varchar(20);default:'audio'" json:"type"`
	Content     string         `gorm:"type:text" json:"content"`
	DurationSec int            `gorm:"default:0" json:"duration_sec"`
	FileSizeKB  int            `gorm:"default:0" json:"file_size_kb"`
	IsTextToSpeech bool        `gorm:"default:false" json:"is_text_to_speech"`
	Voice       string         `gorm:"type:varchar(50)" json:"voice"`
	Speed       float64        `gorm:"type:decimal(3,2);default:1.0" json:"speed"`
	Pitch       float64        `gorm:"type:decimal(3,2);default:1.0" json:"pitch"`
	Volume      int            `gorm:"default:80" json:"volume"`
	CreatedBy   uint64         `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Payload *PayloadDevice `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
}

func (SpeakerAudio) TableName() string {
	return "speaker_audios"
}

type SprayerStatus struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PayloadID        uint64         `gorm:"index;not null" json:"payload_id"`
	IsSpraying       bool           `gorm:"default:false" json:"is_spraying"`
	FlowRate         float64        `gorm:"type:decimal(8,2)" json:"flow_rate_lpm"`
	TargetFlowRate   float64        `gorm:"type:decimal(8,2)" json:"target_flow_rate_lpm"`
	TankCapacityL    float64        `gorm:"type:decimal(8,2)" json:"tank_capacity_l"`
	TankRemainingL   float64        `gorm:"type:decimal(8,2)" json:"tank_remaining_l"`
	PressureKPa      float64        `gorm:"type:decimal(8,2)" json:"pressure_kpa"`
	NozzleCount      int            `gorm:"default:0" json:"nozzle_count"`
	ActiveNozzleMask int            `gorm:"default:0" json:"active_nozzle_mask"`
	TotalSprayedL    float64        `gorm:"type:decimal(10,2)" json:"total_sprayed_l"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Payload *PayloadDevice `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
}

func (SprayerStatus) TableName() string {
	return "sprayer_statuses"
}

type PayloadTelemetry struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID     uint64    `gorm:"index;not null" json:"uav_id"`
	PayloadID uint64    `gorm:"index;not null" json:"payload_id"`
	PayloadType string  `gorm:"type:varchar(30);not null" json:"payload_type"`
	Data      string    `gorm:"type:text" json:"data"`
	Timestamp time.Time `gorm:"index" json:"timestamp"`
}

func (PayloadTelemetry) TableName() string {
	return "payload_telemetry"
}
