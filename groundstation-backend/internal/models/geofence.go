package models

import (
	"time"

	"gorm.io/gorm"
)

type GeofenceType string

const (
	GeofenceTypeInclusion GeofenceType = "inclusion"
	GeofenceTypeExclusion GeofenceType = "exclusion"
)

type GeofenceShape string

const (
	GeofenceShapePolygon   GeofenceShape = "polygon"
	GeofenceShapeCircle    GeofenceShape = "circle"
	GeofenceShapeRectangle GeofenceShape = "rectangle"
)

type Geofence struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Type        GeofenceType   `gorm:"type:varchar(20);not null" json:"type"`
	Shape       GeofenceShape  `gorm:"type:varchar(20);not null" json:"shape"`
	CreatorID   uint64         `json:"creator_id"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	MaxAltitude float64        `gorm:"type:decimal(8,2)" json:"max_altitude"`
	MinAltitude float64        `gorm:"type:decimal(8,2)" json:"min_altitude"`
	CenterLat   float64        `gorm:"type:decimal(10,7)" json:"center_lat"`
	CenterLng   float64        `gorm:"type:decimal(10,7)" json:"center_lng"`
	Radius      float64        `gorm:"type:decimal(10,2)" json:"radius"`
	Coordinates string         `gorm:"type:json" json:"coordinates"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	UAVs []UAV `gorm:"many2many:uav_geofences;" json:"uavs,omitempty"`
}

func (Geofence) TableName() string {
	return "geofences"
}

type UAVGeofence struct {
	UAVID      uint64    `gorm:"primaryKey" json:"uav_id"`
	GeofenceID uint64    `gorm:"primaryKey" json:"geofence_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UAVGeofence) TableName() string {
	return "uav_geofences"
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
