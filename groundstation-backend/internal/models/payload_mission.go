package models

import (
	"time"

	"gorm.io/gorm"
)

type OrbitMissionStatus string

const (
	OrbitStatusPending   OrbitMissionStatus = "pending"
	OrbitStatusActive    OrbitMissionStatus = "active"
	OrbitStatusPaused    OrbitMissionStatus = "paused"
	OrbitStatusCompleted OrbitMissionStatus = "completed"
	OrbitStatusAborted   OrbitMissionStatus = "aborted"
)

type OrbitMission struct {
	ID              uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID            string              `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID           uint64              `gorm:"index;not null" json:"uav_id"`
	PayloadID       uint64              `gorm:"index" json:"payload_id"`
	Name            string              `gorm:"type:varchar(100);not null" json:"name"`
	Status          OrbitMissionStatus  `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CenterLatitude  float64             `gorm:"type:decimal(10,7);not null" json:"center_latitude"`
	CenterLongitude float64             `gorm:"type:decimal(10,7);not null" json:"center_longitude"`
	Altitude        float64             `gorm:"type:decimal(8,2);not null" json:"altitude"`
	Radius          float64             `gorm:"type:decimal(8,2);not null" json:"radius"`
	Velocity        float64             `gorm:"type:decimal(6,2)" json:"velocity"`
	YawRate         float64             `gorm:"type:decimal(6,2)" json:"yaw_rate"`
	Direction       int                 `gorm:"default:1" json:"direction"`
	Loops           int                 `gorm:"default:1" json:"loops"`
	CurrentLoop     int                 `gorm:"default:0" json:"current_loop"`
	Progress        float64             `gorm:"type:decimal(5,2);default:0" json:"progress"`
	AutoCapture     bool                `gorm:"default:true" json:"auto_capture"`
	CaptureInterval int                 `gorm:"default:5" json:"capture_interval_sec"`
	CameraGimbalPitch float64           `gorm:"type:decimal(5,2)" json:"camera_gimbal_pitch"`
	CameraGimbalYaw   float64           `gorm:"type:decimal(5,2)" json:"camera_gimbal_yaw"`
	StartAt         *time.Time          `json:"start_at"`
	EndAt           *time.Time          `json:"end_at"`
	PlannedStart    *time.Time          `json:"planned_start"`
	OperatorID      uint64              `json:"operator_id"`
	CreatorID       uint64              `json:"creator_id"`
	Notes           string              `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`

	UAV     *UAV           `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Payload *PayloadDevice `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
}

func (OrbitMission) TableName() string {
	return "orbit_missions"
}

type OrthoMissionStatus string

const (
	OrthoStatusPending   OrthoMissionStatus = "pending"
	OrthoStatusPlanning  OrthoMissionStatus = "planning"
	OrthoStatusActive    OrthoMissionStatus = "active"
	OrthoStatusPaused    OrthoMissionStatus = "paused"
	OrthoStatusCompleted OrthoMissionStatus = "completed"
	OrthoStatusAborted   OrthoMissionStatus = "aborted"
)

type OrthoMission struct {
	ID                uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID              string              `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID             uint64              `gorm:"index;not null" json:"uav_id"`
	PayloadID         uint64              `gorm:"index" json:"payload_id"`
	Name              string              `gorm:"type:varchar(100);not null" json:"name"`
	Status            OrthoMissionStatus  `gorm:"type:varchar(20);default:'pending'" json:"status"`
	SurveyArea        string              `gorm:"type:text" json:"survey_area_geojson"`
	Altitude          float64             `gorm:"type:decimal(8,2);not null" json:"altitude"`
	Speed             float64             `gorm:"type:decimal(6,2)" json:"speed"`
	OverlapFront      float64             `gorm:"type:decimal(5,2);default:80" json:"overlap_front"`
	OverlapSide       float64             `gorm:"type:decimal(5,2);default:70" json:"overlap_side"`
	GSD               float64             `gorm:"type:decimal(6,3)" json:"gsd_cm"`
	CameraAngle       float64             `gorm:"type:decimal(5,2);default:-90" json:"camera_angle"`
	DirectionAngle    float64             `gorm:"type:decimal(5,2)" json:"direction_angle"`
	TotalDistance     float64             `gorm:"type:decimal(10,2)" json:"total_distance_m"`
	TotalPhotos       int                 `gorm:"default:0" json:"total_photos"`
	CapturedPhotos    int                 `gorm:"default:0" json:"captured_photos"`
	CurrentWaypoint   int                 `gorm:"default:0" json:"current_waypoint"`
	TotalWaypoints    int                 `gorm:"default:0" json:"total_waypoints"`
	Progress          float64             `gorm:"type:decimal(5,2);default:0" json:"progress"`
	TriggerMode       string              `gorm:"type:varchar(20);default:'distance'" json:"trigger_mode"`
	TriggerDistance   float64             `gorm:"type:decimal(8,2)" json:"trigger_distance_m"`
	TriggerInterval   int                 `gorm:"default:2" json:"trigger_interval_sec"`
	ReturnToHome      bool                `gorm:"default:true" json:"return_to_home"`
	MissionID         uint64              `gorm:"index" json:"mission_id"`
	StartAt           *time.Time          `json:"start_at"`
	EndAt             *time.Time          `json:"end_at"`
	PlannedStart      *time.Time          `json:"planned_start"`
	OperatorID        uint64              `json:"operator_id"`
	CreatorID         uint64              `json:"creator_id"`
	Notes             string              `gorm:"type:text" json:"notes"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `gorm:"index" json:"-"`

	UAV      *UAV              `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Payload  *PayloadDevice    `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
	Mission  *FlightMission    `gorm:"foreignKey:MissionID" json:"mission,omitempty"`
	Waypoints []OrthoWaypoint  `gorm:"foreignKey:OrthoMissionID" json:"waypoints,omitempty"`
}

func (OrthoMission) TableName() string {
	return "ortho_missions"
}

type OrthoWaypoint struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	OrthoMissionID uint64         `gorm:"index;not null" json:"ortho_mission_id"`
	Seq            int            `gorm:"not null" json:"seq"`
	Latitude       float64        `gorm:"type:decimal(10,7);not null" json:"latitude"`
	Longitude      float64        `gorm:"type:decimal(10,7);not null" json:"longitude"`
	Altitude       float64        `gorm:"type:decimal(8,2);not null" json:"altitude"`
	IsTurnPoint    bool           `gorm:"default:false" json:"is_turn_point"`
	TriggerPhoto   bool           `gorm:"default:false" json:"trigger_photo"`
	IsReached      bool           `gorm:"default:false" json:"is_reached"`
	ReachedAt      *time.Time     `json:"reached_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OrthoWaypoint) TableName() string {
	return "ortho_waypoints"
}

type TextToSpeechTask struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID       string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	PayloadID  uint64         `gorm:"index;not null" json:"payload_id"`
	UAVID      uint64         `gorm:"index;not null" json:"uav_id"`
	Text       string         `gorm:"type:text;not null" json:"text"`
	Language   string         `gorm:"type:varchar(10);default:'zh-CN'" json:"language"`
	Voice      string         `gorm:"type:varchar(50)" json:"voice"`
	Speed      float64        `gorm:"type:decimal(3,2);default:1.0" json:"speed"`
	Pitch      float64        `gorm:"type:decimal(3,2);default:1.0" json:"pitch"`
	Volume     int            `gorm:"default:80" json:"volume"`
	Status     string         `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AudioURL   string         `gorm:"type:varchar(500)" json:"audio_url"`
	Duration   int            `gorm:"default:0" json:"duration_sec"`
	ErrorMsg   string         `gorm:"type:text" json:"error_msg"`
	CreatedBy  uint64         `json:"created_by"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Payload *PayloadDevice `gorm:"foreignKey:PayloadID" json:"payload,omitempty"`
	UAV     *UAV           `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (TextToSpeechTask) TableName() string {
	return "tts_tasks"
}
