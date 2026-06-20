package models

import (
	"time"

	"gorm.io/gorm"
)

type DetectionClass string

const (
	DetectionClassPerson    DetectionClass = "person"
	DetectionClassCar       DetectionClass = "car"
	DetectionClassTruck     DetectionClass = "truck"
	DetectionClassBus       DetectionClass = "bus"
	DetectionClassMotorcycle DetectionClass = "motorcycle"
	DetectionClassBicycle   DetectionClass = "bicycle"
	DetectionClassDog       DetectionClass = "dog"
	DetectionClassUnknown   DetectionClass = "unknown"
)

type DetectionTarget struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID        uint64         `gorm:"not null;index" json:"uav_id"`
	Class        DetectionClass `gorm:"type:varchar(32);not null" json:"class"`
	ClassName    string         `gorm:"type:varchar(64)" json:"class_name"`
	Confidence   float64        `gorm:"type:decimal(5,4)" json:"confidence"`
	BboxX        float64        `gorm:"type:decimal(8,4)" json:"bbox_x"`
	BboxY        float64        `gorm:"type:decimal(8,4)" json:"bbox_y"`
	BboxWidth    float64        `gorm:"type:decimal(8,4)" json:"bbox_width"`
	BboxHeight   float64        `gorm:"type:decimal(8,4)" json:"bbox_height"`
	FrameWidth   int            `json:"frame_width"`
	FrameHeight  int            `json:"frame_height"`
	Latitude     float64        `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude    float64        `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude     float64        `gorm:"type:decimal(8,2)" json:"altitude"`
	ImagePath    string         `gorm:"type:varchar(512)" json:"image_path"`
	TrackID      *string        `gorm:"type:varchar(64);index" json:"track_id"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DetectionTarget) TableName() string {
	return "detection_targets"
}

type TrackingStatus string

const (
	TrackingStatusIdle      TrackingStatus = "idle"
	TrackingStatusLocking   TrackingStatus = "locking"
	TrackingStatusTracking  TrackingStatus = "tracking"
	TrackingStatusSearching TrackingStatus = "searching"
	TrackingStatusLost      TrackingStatus = "lost"
	TrackingStatusCompleted TrackingStatus = "completed"
)

type TrackingTask struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID            uint64         `gorm:"not null;index" json:"uav_id"`
	Name             string         `gorm:"type:varchar(128)" json:"name"`
	TargetClass      DetectionClass `gorm:"type:varchar(32)" json:"target_class"`
	Status           TrackingStatus `gorm:"type:varchar(20);default:'idle'" json:"status"`
	InitialBboxX     float64        `gorm:"type:decimal(8,4)" json:"initial_bbox_x"`
	InitialBboxY     float64        `gorm:"type:decimal(8,4)" json:"initial_bbox_y"`
	InitialBboxWidth float64        `gorm:"type:decimal(8,4)" json:"initial_bbox_width"`
	InitialBboxHeight float64       `gorm:"type:decimal(8,4)" json:"initial_bbox_height"`
	CurrentBboxX     *float64       `gorm:"type:decimal(8,4)" json:"current_bbox_x"`
	CurrentBboxY     *float64       `gorm:"type:decimal(8,4)" json:"current_bbox_y"`
	CurrentBboxWidth *float64       `gorm:"type:decimal(8,4)" json:"current_bbox_width"`
	CurrentBboxHeight *float64      `gorm:"type:decimal(8,4)" json:"current_bbox_height"`
	CenterOffsetX    *float64       `gorm:"type:decimal(8,4)" json:"center_offset_x"`
	CenterOffsetY    *float64       `gorm:"type:decimal(8,4)" json:"center_offset_y"`
	SearchRadius     float64        `gorm:"type:decimal(8,2);default:10.0" json:"search_radius"`
	MaxSearchRadius  float64        `gorm:"type:decimal(8,2);default:50.0" json:"max_search_radius"`
	Confidence       *float64       `gorm:"type:decimal(5,4)" json:"confidence"`
	FramesVisible    int            `gorm:"default:0" json:"frames_visible"`
	FramesLost       int            `gorm:"default:0" json:"frames_lost"`
	TargetLatitude   *float64       `gorm:"type:decimal(10,7)" json:"target_latitude"`
	TargetLongitude  *float64       `gorm:"type:decimal(10,7)" json:"target_longitude"`
	StartTime        *time.Time     `json:"start_time"`
	EndTime          *time.Time     `json:"end_time"`
	CreatedBy        uint64         `json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (TrackingTask) TableName() string {
	return "tracking_tasks"
}
