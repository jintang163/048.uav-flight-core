package models

import (
	"time"

	"gorm.io/gorm"
)

type FlightStatus struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID          uint64         `gorm:"index;not null" json:"uav_id"`
	Timestamp      time.Time      `gorm:"index;not null" json:"timestamp"`
	Latitude       float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude      float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	AltitudeMSL    float64        `gorm:"type:decimal(8,2)" json:"altitude_msl"`
	AltitudeRel    float64        `gorm:"type:decimal(8,2)" json:"altitude_rel"`
	VelocityX      float64        `gorm:"type:decimal(6,2)" json:"velocity_x"`
	VelocityY      float64        `gorm:"type:decimal(6,2)" json:"velocity_y"`
	VelocityZ      float64        `gorm:"type:decimal(6,2)" json:"velocity_z"`
	GroundSpeed    float64        `gorm:"type:decimal(6,2)" json:"ground_speed"`
	AirSpeed       float64        `gorm:"type:decimal(6,2)" json:"air_speed"`
	Heading        float64        `gorm:"type:decimal(5,2)" json:"heading"`
	Pitch          float64        `gorm:"type:decimal(5,2)" json:"pitch"`
	Roll           float64        `gorm:"type:decimal(5,2)" json:"roll"`
	Yaw            float64        `gorm:"type:decimal(5,2)" json:"yaw"`
	BatteryVoltage float64        `gorm:"type:decimal(5,2)" json:"battery_voltage"`
	BatteryCurrent float64        `gorm:"type:decimal(5,2)" json:"battery_current"`
	BatteryLevel   float64        `gorm:"type:decimal(5,2)" json:"battery_level"`
	Satellites     int            `json:"satellites"`
	GPSFixType     int            `json:"gps_fix_type"`
	HDOP           float64        `gorm:"type:decimal(4,2)" json:"hdop"`
	VDOP           float64        `gorm:"type:decimal(4,2)" json:"vdop"`
	SignalStrength int            `json:"signal_strength"`
	LinkQuality    int            `json:"link_quality"`
	Mode           string         `gorm:"type:varchar(30)" json:"mode"`
	SystemStatus   string         `gorm:"type:varchar(30)" json:"system_status"`
	ArmStatus      bool           `json:"arm_status"`
	FlightTime     int            `json:"flight_time"`
	DistanceTraveled float64      `gorm:"type:decimal(10,2)" json:"distance_traveled"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlightStatus) TableName() string {
	return "flight_status"
}
