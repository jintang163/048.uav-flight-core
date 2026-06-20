package models

import (
	"time"

	"gorm.io/gorm"
)

type MissionType string

const (
	MissionTypeSurvey   MissionType = "survey"
	MissionTypePatrol   MissionType = "patrol"
	MissionTypeDelivery MissionType = "delivery"
	MissionTypeInspect  MissionType = "inspect"
	MissionTypeCustom   MissionType = "custom"
)

type MissionTemplate struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Type        MissionType    `gorm:"type:varchar(20);default:'custom'" json:"type"`
	CreatorID   uint64         `json:"creator_id"`
	MaxAltitude float64        `gorm:"type:decimal(8,2)" json:"max_altitude"`
	MinAltitude float64        `gorm:"type:decimal(8,2)" json:"min_altitude"`
	Speed       float64        `gorm:"type:decimal(6,2)" json:"speed"`
	MaxDuration int            `json:"max_duration"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Waypoints []MissionWaypoint `gorm:"foreignKey:TemplateID" json:"waypoints,omitempty"`
}

func (MissionTemplate) TableName() string {
	return "mission_templates"
}
