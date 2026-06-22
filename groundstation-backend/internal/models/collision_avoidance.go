package models

import (
	"time"

	"gorm.io/gorm"
)

type CollisionRiskLevel string

const (
	CollisionRiskSafe     CollisionRiskLevel = "safe"
	CollisionRiskWarning  CollisionRiskLevel = "warning"
	CollisionRiskCritical CollisionRiskLevel = "critical"
	CollisionRiskAvoiding CollisionRiskLevel = "avoiding"
	CollisionRiskResolved CollisionRiskLevel = "resolved"
)

type AvoidanceActionType string

const (
	AvoidanceSpeedReduce  AvoidanceActionType = "speed_reduce"
	AvoidanceSpeedAdjust  AvoidanceActionType = "speed_adjust"
	AvoidanceHoldPosition AvoidanceActionType = "hold_position"
	AvoidanceWaypointHold AvoidanceActionType = "waypoint_hold"
	AvoidanceAltitude     AvoidanceActionType = "altitude_change"
	AvoidanceResume       AvoidanceActionType = "resume"
)

type CollisionAlert struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	AlertID      string           `gorm:"type:varchar(64);uniqueIndex;not null" json:"alert_id"`
	UAVID1       uint64           `gorm:"index;not null" json:"uav_id_1"`
	UAVID2       uint64           `gorm:"index;not null" json:"uav_id_2"`
	RiskLevel    CollisionRiskLevel `gorm:"type:varchar(20);not null" json:"risk_level"`
	MinDistance  float64          `gorm:"type:decimal(8,2)" json:"min_distance"`
	CurrentDistance float64       `gorm:"type:decimal(8,2)" json:"current_distance"`
	TimeToCollision float64       `gorm:"type:decimal(6,1)" json:"time_to_collision"`
	AlertType    string           `gorm:"type:varchar(30)" json:"alert_type"`
	Action       AvoidanceActionType `gorm:"type:varchar(30)" json:"action_taken"`
	ActionDetail string           `gorm:"type:varchar(200)" json:"action_detail"`
	IsResolved   bool             `gorm:"default:false" json:"is_resolved"`
	ResolvedAt   *time.Time       `json:"resolved_at"`
	CreatedAt    time.Time        `gorm:"index" json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"-"`

	UAV1 *UAV `gorm:"foreignKey:UAVID1" json:"uav1,omitempty"`
	UAV2 *UAV `gorm:"foreignKey:UAVID2" json:"uav2,omitempty"`
}

func (CollisionAlert) TableName() string {
	return "collision_alerts"
}

type RouteIntersection struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID1       uint64    `gorm:"index;not null" json:"uav_id_1"`
	UAVID2       uint64    `gorm:"index;not null" json:"uav_id_2"`
	MissionID1   uint64    `json:"mission_id_1"`
	MissionID2   uint64    `json:"mission_id_2"`
	WaypointSeq1 int       `json:"waypoint_seq_1"`
	WaypointSeq2 int       `json:"waypoint_seq_2"`
	Latitude     float64   `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude    float64   `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude     float64   `gorm:"type:decimal(8,2)" json:"altitude"`
	Distance     float64   `gorm:"type:decimal(8,2)" json:"distance_m"`
	ETA1         time.Time `json:"eta_1"`
	ETA2         time.Time `json:"eta_2"`
	TimeDiffSec  float64   `gorm:"type:decimal(6,1)" json:"time_diff_sec"`
	RiskLevel    CollisionRiskLevel `gorm:"type:varchar(20)" json:"risk_level"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RouteIntersection) TableName() string {
	return "route_intersections"
}

type UAVLivePosition struct {
	UAVID      uint64    `json:"uav_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Altitude   float64   `json:"altitude"`
	GroundSpeed float64  `json:"ground_speed"`
	Heading    float64   `json:"heading"`
	VelocityX  float64   `json:"velocity_x"`
	VelocityY  float64   `json:"velocity_y"`
	VelocityZ  float64   `json:"velocity_z"`
	Mode       string    `json:"mode"`
	Timestamp  time.Time `json:"timestamp"`
}

type AvoidanceDecision struct {
	PairKey      string              `json:"pair_key"`
	UAVID1       uint64              `json:"uav_id_1"`
	UAVID2       uint64              `json:"uav_id_2"`
	Distance     float64             `json:"distance"`
	RiskLevel    CollisionRiskLevel  `json:"risk_level"`
	PrimaryAction AvoidanceActionType `json:"primary_action"`
	SecondaryAction AvoidanceActionType `json:"secondary_action,omitempty"`
	SpeedFactor1 float64             `json:"speed_factor_1"`
	SpeedFactor2 float64             `json:"speed_factor_2"`
	HoldDuration time.Duration       `json:"hold_duration"`
	Reason       string              `json:"reason"`
}
