package models

import (
	"time"

	"gorm.io/gorm"
)

type LandingPointType string

const (
	LandingPointTypePrimary  LandingPointType = "primary"
	LandingPointTypeAlternate LandingPointType = "alternate"
	LandingPointTypeEmergency LandingPointType = "emergency"
)

type LandingPointStatus string

const (
	LandingPointStatusAvailable LandingPointStatus = "available"
	LandingPointStatusOccupied  LandingPointStatus = "occupied"
	LandingPointStatusBlocked   LandingPointStatus = "blocked"
)

type LandingPoint struct {
	ID          uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string           `gorm:"type:varchar(100);not null" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Type        LandingPointType `gorm:"type:varchar(20);not null" json:"type"`
	Latitude    float64          `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude   float64          `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Altitude    float64          `gorm:"type:decimal(8,2)" json:"altitude"`
	Heading     float64          `gorm:"type:decimal(5,2)" json:"heading"`
	Radius      float64          `gorm:"type:decimal(6,2);default:5.0" json:"radius"`
	Priority    int              `gorm:"default:0" json:"priority"`
	Status      LandingPointStatus `gorm:"type:varchar(20);default:'available'" json:"status"`
	HasMarkers  bool             `gorm:"default:false" json:"has_markers"`
	MarkerType  string           `gorm:"type:varchar(50)" json:"marker_type"`
	IsMovingPlatform bool        `gorm:"default:false" json:"is_moving_platform"`
	MovingPlatformID string      `gorm:"type:varchar(64)" json:"moving_platform_id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (LandingPoint) TableName() string {
	return "landing_points"
}

type LandingSessionStatus string

const (
	LandingSessionStatusPending    LandingSessionStatus = "pending"
	LandingSessionStatusApproaching LandingSessionStatus = "approaching"
	LandingSessionStatusDescending LandingSessionStatus = "descending"
	LandingSessionStatusPrecision  LandingSessionStatus = "precision"
	LandingSessionStatusLanded     LandingSessionStatus = "landed"
	LandingSessionStatusAborted    LandingSessionStatus = "aborted"
	LandingSessionStatusFailed     LandingSessionStatus = "failed"
)

type LandingSession struct {
	ID                uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID             uint64              `gorm:"index;not null" json:"uav_id"`
	MissionID         uint64              `gorm:"index" json:"mission_id"`
	PrimaryLandingID  uint64              `gorm:"index" json:"primary_landing_id"`
	AlternateLandingID uint64             `gorm:"index" json:"alternate_landing_id"`
	Status            LandingSessionStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	RTKEnabled        bool                `gorm:"default:true" json:"rtk_enabled"`
	VisionEnabled     bool                `gorm:"default:true" json:"vision_enabled"`
	RTKFixType        int                 `json:"rtk_fix_type"`
	HorizontalAccuracy float64            `gorm:"type:decimal(6,3)" json:"horizontal_accuracy"`
	VerticalAccuracy  float64             `gorm:"type:decimal(6,3)" json:"vertical_accuracy"`
	MarkerDetected    bool                `gorm:"default:false" json:"marker_detected"`
	MarkerType        string              `gorm:"type:varchar(50)" json:"marker_type"`
	MarkerOffsetX     float64             `gorm:"type:decimal(6,3)" json:"marker_offset_x"`
	MarkerOffsetY     float64             `gorm:"type:decimal(6,3)" json:"marker_offset_y"`
	TargetLatitude    float64             `gorm:"type:decimal(10,7)" json:"target_latitude"`
	TargetLongitude   float64             `gorm:"type:decimal(10,7)" json:"target_longitude"`
	TargetAltitude    float64             `gorm:"type:decimal(8,2)" json:"target_altitude"`
	LandingError      float64             `gorm:"type:decimal(6,3)" json:"landing_error"`
	IsMovingPlatform  bool                `gorm:"default:false" json:"is_moving_platform"`
	MovingPlatformVelocityX float64        `gorm:"type:decimal(6,2)" json:"moving_platform_velocity_x"`
	MovingPlatformVelocityY float64        `gorm:"type:decimal(6,2)" json:"moving_platform_velocity_y"`
	StartTime         *time.Time          `json:"start_time"`
	EndTime           *time.Time          `json:"end_time"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `gorm:"index" json:"-"`

	PrimaryLanding   *LandingPoint `gorm:"foreignKey:PrimaryLandingID" json:"primary_landing,omitempty"`
	AlternateLanding *LandingPoint `gorm:"foreignKey:AlternateLandingID" json:"alternate_landing,omitempty"`
}

func (LandingSession) TableName() string {
	return "landing_sessions"
}

type LandingTrajectoryPoint struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID       uint64    `gorm:"index;not null" json:"session_id"`
	Sequence        int       `gorm:"not null" json:"sequence"`
	Timestamp       time.Time `gorm:"not null" json:"timestamp"`
	Latitude        float64   `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude       float64   `gorm:"type:decimal(10,7);not null" json:"longitude"`
	AltitudeMSL     float64   `gorm:"type:decimal(8,2)" json:"altitude_msl"`
	AltitudeRel     float64   `gorm:"type:decimal(8,2)" json:"altitude_rel"`
	VelocityX       float64   `gorm:"type:decimal(6,2)" json:"velocity_x"`
	VelocityY       float64   `gorm:"type:decimal(6,2)" json:"velocity_y"`
	VelocityZ       float64   `gorm:"type:decimal(6,2)" json:"velocity_z"`
	Heading         float64   `gorm:"type:decimal(5,2)" json:"heading"`
	Pitch           float64   `gorm:"type:decimal(5,2)" json:"pitch"`
	Roll            float64   `gorm:"type:decimal(5,2)" json:"roll"`
	Throttle        float64   `gorm:"type:decimal(5,2)" json:"throttle"`
	RTKFixType      int       `json:"rtk_fix_type"`
	HDOP            float64   `gorm:"type:decimal(4,2)" json:"hdop"`
	VDOP            float64   `gorm:"type:decimal(4,2)" json:"vdop"`
	MarkerDetected  bool      `gorm:"default:false" json:"marker_detected"`
	MarkerOffsetX   float64   `gorm:"type:decimal(6,3)" json:"marker_offset_x"`
	MarkerOffsetY   float64   `gorm:"type:decimal(6,3)" json:"marker_offset_y"`
	Phase           string    `gorm:"type:varchar(20)" json:"phase"`
	CreatedAt       time.Time `json:"created_at"`
}

func (LandingTrajectoryPoint) TableName() string {
	return "landing_trajectory_points"
}

type ForcedLandingEvent struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64         `gorm:"index;not null" json:"uav_id"`
	SessionID       uint64         `gorm:"index" json:"session_id"`
	TriggerType     string         `gorm:"type:varchar(50);not null" json:"trigger_type"`
	Reason          string         `gorm:"type:text" json:"reason"`
	ArmLocked       bool           `gorm:"default:false" json:"arm_locked"`
	EmergencyMode   string         `gorm:"type:varchar(20)" json:"emergency_mode"`
	Latitude        float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude       float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude        float64        `gorm:"type:decimal(8,2)" json:"altitude"`
	TriggeredAt     time.Time      `json:"triggered_at"`
	ResolvedAt      *time.Time     `json:"resolved_at"`
	ResolvedBy      uint64         `json:"resolved_by"`
	ResolutionNotes string         `gorm:"type:text" json:"resolution_notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ForcedLandingEvent) TableName() string {
	return "forced_landing_events"
}

type VisionLandingData struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64    `gorm:"index;not null" json:"uav_id"`
	SessionID       uint64    `gorm:"index" json:"session_id"`
	Timestamp       time.Time `gorm:"not null" json:"timestamp"`
	MarkerDetected  bool      `gorm:"default:false" json:"marker_detected"`
	MarkerType      string    `gorm:"type:varchar(50)" json:"marker_type"`
	MarkerID        string    `gorm:"type:varchar(100)" json:"marker_id"`
	Confidence      float64   `gorm:"type:decimal(5,2)" json:"confidence"`
	OffsetX         float64   `gorm:"type:decimal(6,3)" json:"offset_x"`
	OffsetY         float64   `gorm:"type:decimal(6,3)" json:"offset_y"`
	OffsetZ         float64   `gorm:"type:decimal(6,3)" json:"offset_z"`
	YawError        float64   `gorm:"type:decimal(5,2)" json:"yaw_error"`
	CameraHeight    float64   `gorm:"type:decimal(6,2)" json:"camera_height"`
	ImageWidth      int       `json:"image_width"`
	ImageHeight     int       `json:"image_height"`
	CreatedAt       time.Time `json:"created_at"`
}

func (VisionLandingData) TableName() string {
	return "vision_landing_data"
}

type RTKPositionData struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64    `gorm:"index;not null" json:"uav_id"`
	SessionID       uint64    `gorm:"index" json:"session_id"`
	Timestamp       time.Time `gorm:"not null" json:"timestamp"`
	Latitude        float64   `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude       float64   `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Altitude        float64   `gorm:"type:decimal(8,2)" json:"altitude"`
	FixType         int       `gorm:"not null" json:"fix_type"`
	Satellites      int       `json:"satellites"`
	HorizontalAcc   float64   `gorm:"type:decimal(6,3)" json:"horizontal_accuracy"`
	VerticalAcc     float64   `gorm:"type:decimal(6,3)" json:"vertical_accuracy"`
	HDOP            float64   `gorm:"type:decimal(4,2)" json:"hdop"`
	VDOP            float64   `gorm:"type:decimal(4,2)" json:"vdop"`
	BaseStationID   string    `gorm:"type:varchar(64)" json:"base_station_id"`
	DifferentialAge float64   `gorm:"type:decimal(6,2)" json:"differential_age"`
	CreatedAt       time.Time `json:"created_at"`
}

func (RTKPositionData) TableName() string {
	return "rtk_position_data"
}
