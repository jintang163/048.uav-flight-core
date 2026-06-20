package models

import (
	"time"

	"gorm.io/gorm"
)

type FormationType string

const (
	FormationLine     FormationType = "line"
	FormationTriangle FormationType = "triangle"
	FormationCircle   FormationType = "circle"
)

type FormationStatus string

const (
	FormationStatusIdle      FormationStatus = "idle"
	FormationStatusReady     FormationStatus = "ready"
	FormationStatusExecuting FormationStatus = "executing"
	FormationStatusPaused    FormationStatus = "paused"
	FormationStatusCompleted FormationStatus = "completed"
)

type LightEffect string

const (
	LightEffectStatic    LightEffect = "static"
	LightEffectBlink     LightEffect = "blink"
	LightEffectRainbow   LightEffect = "rainbow"
	LightEffectBreathing LightEffect = "breathing"
)

type Formation struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string          `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name        string          `gorm:"type:varchar(100);not null" json:"name"`
	Type        FormationType   `gorm:"type:varchar(20);not null" json:"type"`
	Status      FormationStatus `gorm:"type:varchar(20);default:'idle'" json:"status"`
	LeaderID    uint64          `json:"leader_id"`
	Spacing     float64         `gorm:"type:decimal(8,2);default:5.0" json:"spacing"`
	Description string          `gorm:"type:text" json:"description"`
	OwnerID     uint64          `json:"owner_id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	Members []FormationMember `gorm:"foreignKey:FormationID" json:"members,omitempty"`
}

func (Formation) TableName() string {
	return "formations"
}

type FormationMember struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	FormationID   uint64         `gorm:"index;not null" json:"formation_id"`
	UAVID         uint64         `gorm:"index;not null" json:"uav_id"`
	PositionIndex int            `json:"position_index"`
	OffsetX       float64        `gorm:"type:decimal(10,3)" json:"offset_x"`
	OffsetY       float64        `gorm:"type:decimal(10,3)" json:"offset_y"`
	OffsetZ       float64        `gorm:"type:decimal(10,3)" json:"offset_z"`
	IsLeader      bool           `json:"is_leader"`
	Status        string         `gorm:"type:varchar(20);default:'idle'" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (FormationMember) TableName() string {
	return "formation_members"
}

type FormationLightConfig struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Red       uint8          `json:"red"`
	Green     uint8          `json:"green"`
	Blue      uint8          `json:"blue"`
	Effect    LightEffect    `gorm:"type:varchar(20);default:'static'" json:"effect"`
	OwnerID   uint64         `json:"owner_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FormationLightConfig) TableName() string {
	return "formation_light_configs"
}

type FormationCollisionWarning struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormationID    uint64    `gorm:"index;not null" json:"formation_id"`
	UAVID1         uint64    `gorm:"index;not null" json:"uav_id_1"`
	UAVID2         uint64    `gorm:"index;not null" json:"uav_id_2"`
	Distance       float64   `gorm:"type:decimal(10,3)" json:"distance"`
	WarningLevel   string    `gorm:"type:varchar(20)" json:"warning_level"`
	Timestamp      time.Time `gorm:"index;not null" json:"timestamp"`
	Resolved       bool      `gorm:"default:false" json:"resolved"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

func (FormationCollisionWarning) TableName() string {
	return "formation_collision_warnings"
}
