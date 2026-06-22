package models

import (
	"time"

	"gorm.io/gorm"
)

type ObstacleSensorType string

const (
	SensorTypeMillimeterWave ObstacleSensorType = "millimeter_wave_radar"
	SensorTypeStereoVision   ObstacleSensorType = "stereo_vision"
	SensorTypeLidar          ObstacleSensorType = "lidar"
	SensorTypeUltrasonic     ObstacleSensorType = "ultrasonic"
)

type AvoidanceSensitivity string

const (
	SensitivityFar    AvoidanceSensitivity = "far"
	SensitivityMedium AvoidanceSensitivity = "medium"
	SensitivityNear   AvoidanceSensitivity = "near"
)

type AvoidanceStrategy string

const (
	StrategyHover       AvoidanceStrategy = "hover"
	StrategyAscendBypass AvoidanceStrategy = "ascend_bypass"
	StrategyRetreatBypass AvoidanceStrategy = "retreat_bypass"
)

type ObstacleDirection string

const (
	DirectionFront  ObstacleDirection = "front"
	DirectionLeft   ObstacleDirection = "left"
	DirectionRight  ObstacleDirection = "right"
	DirectionTop    ObstacleDirection = "top"
	DirectionBottom ObstacleDirection = "bottom"
	DirectionRear   ObstacleDirection = "rear"
)

type AvoidanceActionStatus string

const (
	AvoidanceStatusDetecting AvoidanceActionStatus = "detecting"
	AvoidanceStatusTriggered AvoidanceActionStatus = "triggered"
	AvoidanceStatusAvoiding  AvoidanceActionStatus = "avoiding"
	AvoidanceStatusBypassing AvoidanceActionStatus = "bypassing"
	AvoidanceStatusCompleted AvoidanceActionStatus = "completed"
	AvoidanceStatusFailed    AvoidanceActionStatus = "failed"
)

type ObstacleAvoidanceConfig struct {
	ID             uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID          uint64              `gorm:"uniqueIndex;not null" json:"uav_id"`
	Enabled        bool                `gorm:"default:true" json:"enabled"`
	Sensitivity    AvoidanceSensitivity `gorm:"type:varchar(20);default:'medium'" json:"sensitivity"`
	Strategy       AvoidanceStrategy   `gorm:"type:varchar(20);default:'ascend_bypass'" json:"strategy"`
	SensorType     ObstacleSensorType  `gorm:"type:varchar(30);default:'millimeter_wave_radar'" json:"sensor_type"`
	DetectionRange float64             `gorm:"type:decimal(5,1);default:15" json:"detection_range"`
	MinObstacleSize float64            `gorm:"type:decimal(5,2);default:0.5" json:"min_obstacle_size"`
	AscendHeight   float64             `gorm:"type:decimal(5,1);default:5" json:"ascend_height"`
	RetreatDistance float64             `gorm:"type:decimal(5,1);default:10" json:"retreat_distance"`
	BypassAngle    float64             `gorm:"type:decimal(5,1);default:45" json:"bypass_angle"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

func (ObstacleAvoidanceConfig) TableName() string {
	return "obstacle_avoidance_configs"
}

type ObstacleDetectionLog struct {
	ID          uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID       uint64              `gorm:"index;not null" json:"uav_id"`
	SensorType  ObstacleSensorType  `gorm:"type:varchar(30)" json:"sensor_type"`
	Direction   ObstacleDirection   `gorm:"type:varchar(20)" json:"direction"`
	Distance    float64             `gorm:"type:decimal(8,2);not null" json:"distance"`
	RelativeAngle float64           `gorm:"type:decimal(8,2)" json:"relative_angle"`
	ObstacleSize float64            `gorm:"type:decimal(5,2)" json:"obstacle_size"`
	Confidence  float64             `gorm:"type:decimal(3,2)" json:"confidence"`
	Latitude    float64             `gorm:"type:decimal(10,7)" json:"lat"`
	Longitude   float64             `gorm:"type:decimal(10,7)" json:"lng"`
	Altitude    float64             `gorm:"type:decimal(8,2)" json:"alt"`
	CreatedAt   time.Time           `json:"created_at"`
}

func (ObstacleDetectionLog) TableName() string {
	return "obstacle_detection_logs"
}

type ObstacleAvoidanceEvent struct {
	ID             uint64               `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID          uint64               `gorm:"index;not null" json:"uav_id"`
	Strategy       AvoidanceStrategy    `gorm:"type:varchar(20);not null" json:"strategy"`
	Status         AvoidanceActionStatus `gorm:"type:varchar(20);default:'triggered'" json:"status"`
	DetectionID    uint64               `gorm:"index" json:"detection_id"`
	SensorType     ObstacleSensorType   `gorm:"type:varchar(30)" json:"sensor_type"`
	Direction      ObstacleDirection    `gorm:"type:varchar(20)" json:"direction"`
	ObstacleDistance float64            `gorm:"type:decimal(8,2)" json:"obstacle_distance"`
	StartLat       float64              `gorm:"type:decimal(10,7)" json:"start_lat"`
	StartLng       float64              `gorm:"type:decimal(10,7)" json:"start_lng"`
	StartAlt       float64              `gorm:"type:decimal(8,2)" json:"start_alt"`
	BypassPath     string               `gorm:"type:json" json:"bypass_path"`
	CompletedAt    *time.Time           `json:"completed_at"`
	FailReason     string               `gorm:"type:varchar(500)" json:"fail_reason"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	DeletedAt      gorm.DeletedAt       `gorm:"index" json:"-"`
}

func (ObstacleAvoidanceEvent) TableName() string {
	return "obstacle_avoidance_events"
}

type ObstacleHeatmapPoint struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64    `gorm:"index;not null" json:"uav_id"`
	Latitude        float64   `gorm:"type:decimal(10,7);not null" json:"lat"`
	Longitude       float64   `gorm:"type:decimal(10,7);not null" json:"lng"`
	Altitude        float64   `gorm:"type:decimal(8,2)" json:"alt"`
	TriggerCount    int       `gorm:"default:1" json:"trigger_count"`
	LastTriggerTime time.Time `json:"last_trigger_time"`
	AvgDistance     float64   `gorm:"type:decimal(8,2)" json:"avg_distance"`
	MinDistance     float64   `gorm:"type:decimal(8,2)" json:"min_distance"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (ObstacleHeatmapPoint) TableName() string {
	return "obstacle_heatmap_points"
}
