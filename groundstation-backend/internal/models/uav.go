package models

import (
	"time"

	"gorm.io/gorm"
)

type UAVStatus string

const (
	UAVStatusOffline   UAVStatus = "offline"
	UAVStatusOnline    UAVStatus = "online"
	UAVStatusFlying    UAVStatus = "flying"
	UAVStatusHovering  UAVStatus = "hovering"
	UAVStatusLanded    UAVStatus = "landed"
	UAVStatusError     UAVStatus = "error"
)

type UAVType string

const (
	UAVTypeMultirotor UAVType = "multirotor"
	UAVTypeFixedWing UAVType = "fixed_wing"
	UAVTypeHybrid    UAVType = "hybrid"
)

type UAV struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID         string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Model        string         `gorm:"type:varchar(50)" json:"model"`
	Type         UAVType        `gorm:"type:varchar(20)" json:"type"`
	SerialNumber string         `gorm:"type:varchar(100);uniqueIndex" json:"serial_number"`
	FirmwareVer  string         `gorm:"type:varchar(50)" json:"firmware_version"`
	Status       UAVStatus      `gorm:"type:varchar(20);default:'offline'" json:"status"`
	StatusMessage string        `gorm:"type:varchar(255)" json:"status_message"`
	IPAddress    string         `gorm:"type:varchar(45)" json:"ip_address"`
	Port         int            `json:"port"`
	Protocol     string         `gorm:"type:varchar(10);default:'tcp'" json:"protocol"`
	OwnerID      uint64         `json:"owner_id"`
	LastSeenAt   *time.Time     `json:"last_seen_at"`
	HomeLatitude float64        `gorm:"type:decimal(10,7)" json:"home_latitude"`
	HomeLongitude float64       `gorm:"type:decimal(10,7)" json:"home_longitude"`
	HomeAltitude float64        `gorm:"type:decimal(8,2)" json:"home_altitude"`
	MaxAltitude   float64        `gorm:"type:decimal(8,2)" json:"max_altitude"`
	MaxSpeed     float64        `gorm:"type:decimal(6,2)" json:"max_speed"`
	MaxFlightTime int           `json:"max_flight_time"`
	Description  string         `gorm:"type:text" json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	FlightStatus  *FlightStatus  `gorm:"foreignKey:UAVID" json:"flight_status,omitempty"`
	CurrentMission *FlightMission `gorm:"foreignKey:UAVID" json:"current_mission,omitempty"`
	Payloads     []PayloadDevice `gorm:"foreignKey:UAVID" json:"payloads,omitempty"`
}

func (UAV) TableName() string {
	return "uavs"
}
