package models

import (
	"time"

	"gorm.io/gorm"
)

type MissionStatus string

const (
	MissionStatusPending    MissionStatus = "pending"
	MissionStatusReady      MissionStatus = "ready"
	MissionStatusExecuting  MissionStatus = "executing"
	MissionStatusPaused     MissionStatus = "paused"
	MissionStatusCompleted  MissionStatus = "completed"
	MissionStatusAborted    MissionStatus = "aborted"
	MissionStatusFailed     MissionStatus = "failed"
)

type FlightMission struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	UAVID         uint64         `gorm:"index;not null" json:"uav_id"`
	TemplateID    uint64         `gorm:"index" json:"template_id"`
	OperatorID    uint64         `json:"operator_id"`
	Status        MissionStatus  `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CurrentWP     int            `gorm:"default:0" json:"current_waypoint"`
	TotalWP       int            `gorm:"default:0" json:"total_waypoints"`
	PlannedStart  *time.Time     `json:"planned_start"`
	PlannedEnd    *time.Time     `json:"planned_end"`
	ActualStart   *time.Time     `json:"actual_start"`
	ActualEnd     *time.Time     `json:"actual_end"`
	MaxAltitude   float64        `gorm:"type:decimal(8,2)" json:"max_altitude"`
	Speed         float64        `gorm:"type:decimal(6,2)" json:"speed"`
	Distance      float64        `gorm:"type:decimal(10,2)" json:"distance"`
	Duration      int            `json:"duration"`
	ResumeFromWP  int            `gorm:"default:0" json:"resume_from_waypoint"`
	EnableGeofence bool          `gorm:"default:true" json:"enable_geofence"`
	FailSafe      string         `gorm:"type:varchar(50);default:'rtl'" json:"fail_safe"`
	Notes         string         `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	UAV        *UAV             `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Template   *MissionTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	Waypoints  []MissionWaypoint `gorm:"foreignKey:MissionID" json:"waypoints,omitempty"`
}

func (FlightMission) TableName() string {
	return "flight_missions"
}
