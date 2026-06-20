package models

import (
	"time"

	"gorm.io/gorm"
)

type ViolationType string

const (
	ViolationTypeAltitudeExceeded    ViolationType = "altitude_exceeded"
	ViolationTypeAltitudeTooLow      ViolationType = "altitude_too_low"
	ViolationTypeInsideExclusionZone ViolationType = "inside_exclusion_zone"
	ViolationTypeOutsideInclusionZone ViolationType = "outside_inclusion_zone"
	ViolationTypeDistanceExceeded    ViolationType = "distance_exceeded"
)

type ViolationSeverity string

const (
	ViolationSeverityWarning  ViolationSeverity = "warning"
	ViolationSeverityCritical ViolationSeverity = "critical"
	ViolationSeverityFatal    ViolationSeverity = "fatal"
)

type GeofenceViolationLog struct {
	ID               uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID            uint64            `gorm:"index;not null" json:"uav_id"`
	GeofenceID       uint64            `gorm:"index;not null" json:"geofence_id"`
	GeofenceName     string            `gorm:"type:varchar(100)" json:"geofence_name"`
	GeofenceCategory GeofenceCategory  `gorm:"type:varchar(30)" json:"geofence_category"`
	ViolationType    ViolationType     `gorm:"type:varchar(30);not null" json:"violation_type"`
	Severity         ViolationSeverity `gorm:"type:varchar(20);default:'warning'" json:"severity"`
	Latitude         float64           `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude        float64           `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude         float64           `gorm:"type:decimal(8,2)" json:"altitude"`
	Distance         float64           `gorm:"type:decimal(10,2)" json:"distance"`
	Duration         int               `gorm:"default:0" json:"duration"`
	ActionTaken      FailAction        `gorm:"type:varchar(20);default:'warn'" json:"action_taken"`
	ActionResult     string            `gorm:"type:varchar(100)" json:"action_result"`
	IsResolved       bool              `gorm:"default:false" json:"is_resolved"`
	ResolvedAt       *time.Time        `json:"resolved_at"`
	Notes            string            `gorm:"type:text" json:"notes"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	DeletedAt        gorm.DeletedAt    `gorm:"index" json:"-"`

	UAV      *UAV      `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Geofence *Geofence `gorm:"foreignKey:GeofenceID" json:"geofence,omitempty"`
}

func (GeofenceViolationLog) TableName() string {
	return "geofence_violation_logs"
}
