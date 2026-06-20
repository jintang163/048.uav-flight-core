package models

import (
	"time"

	"gorm.io/gorm"
)

type WaypointCommand int

const (
	WPCmdNavWaypoint   WaypointCommand = 16
	WPCmdTakeoff       WaypointCommand = 22
	WPCmdLand          WaypointCommand = 21
	WPCmdLoiterTime    WaypointCommand = 19
	WPCmdLoiterTurns   WaypointCommand = 18
	WPCmdReturnToLaunch WaypointCommand = 20
	WPCmdDelay         WaypointCommand = 93
	WPCmdChangeSpeed   WaypointCommand = 178
	WPCmdChangeAlt     WaypointCommand = 113
)

type MissionWaypoint struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TemplateID uint64         `gorm:"index;not null" json:"template_id"`
	MissionID  uint64         `gorm:"index" json:"mission_id,omitempty"`
	Seq        int            `gorm:"not null" json:"seq"`
	Command    WaypointCommand `gorm:"not null" json:"command"`
	Latitude   float64        `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude  float64        `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Altitude   float64        `gorm:"type:decimal(8,2);not null" json:"altitude"`
	Param1     float64        `gorm:"type:decimal(10,4)" json:"param1"`
	Param2     float64        `gorm:"type:decimal(10,4)" json:"param2"`
	Param3     float64        `gorm:"type:decimal(10,4)" json:"param3"`
	Param4     float64        `gorm:"type:decimal(10,4)" json:"param4"`
	HoldTime   int            `json:"hold_time"`
	Radius     float64        `gorm:"type:decimal(8,2)" json:"radius"`
	IsReached  bool           `gorm:"default:false" json:"is_reached"`
	ReachedAt  *time.Time     `json:"reached_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MissionWaypoint) TableName() string {
	return "mission_waypoints"
}
