package models

import (
	"time"
)

type ThrustLearningStatus struct {
	ID              uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64     `gorm:"index;not null" json:"uav_id"`
	State           string     `gorm:"type:varchar(30);default:'idle'" json:"state"`
	EstimatedWeight float64    `gorm:"type:decimal(6,3)" json:"estimated_weight_kg"`
	HoverThrottle   float64    `gorm:"type:decimal(5,3)" json:"hover_throttle"`
	SampleCount     uint32     `gorm:"default:0" json:"sample_count"`
	Progress        float64    `gorm:"type:decimal(5,2);default:0" json:"progress_pct"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ThrustLearningStatus) TableName() string {
	return "thrust_learning_statuses"
}

type ThrustCurvePoint struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID       uint64    `gorm:"index;not null" json:"uav_id"`
	Throttle    float64   `gorm:"type:decimal(5,3)" json:"throttle"`
	ThrustN     float64   `gorm:"type:decimal(7,3)" json:"thrust_n"`
	MotorRpmAvg float64   `gorm:"type:decimal(8,1)" json:"motor_rpm_avg"`
	SampleCount int       `gorm:"default:1" json:"sample_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ThrustCurvePoint) TableName() string {
	return "thrust_curve_points"
}

type PIDGainProfile struct {
	ID          uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID       uint64  `gorm:"uniqueIndex;not null" json:"uav_id"`
	ProfileName string  `gorm:"type:varchar(50);default:'default'" json:"profile_name"`
	IsAutoTuned bool    `gorm:"default:false" json:"is_auto_tuned"`

	RollKP  float64 `gorm:"type:decimal(8,4);default:0" json:"roll_kp"`
	RollKI  float64 `gorm:"type:decimal(8,4);default:0" json:"roll_ki"`
	RollKD  float64 `gorm:"type:decimal(8,4);default:0" json:"roll_kd"`
	PitchKP float64 `gorm:"type:decimal(8,4);default:0" json:"pitch_kp"`
	PitchKI float64 `gorm:"type:decimal(8,4);default:0" json:"pitch_ki"`
	PitchKD float64 `gorm:"type:decimal(8,4);default:0" json:"pitch_kd"`
	YawKP   float64 `gorm:"type:decimal(8,4);default:0" json:"yaw_kp"`
	YawKI   float64 `gorm:"type:decimal(8,4);default:0" json:"yaw_ki"`
	YawKD   float64 `gorm:"type:decimal(8,4);default:0" json:"yaw_kd"`

	RateRollKP  float64 `gorm:"type:decimal(8,4);default:0" json:"rate_roll_kp"`
	RateRollKI  float64 `gorm:"type:decimal(8,4);default:0" json:"rate_roll_ki"`
	RateRollKD  float64 `gorm:"type:decimal(8,4);default:0" json:"rate_roll_kd"`
	RatePitchKP float64 `gorm:"type:decimal(8,4);default:0" json:"rate_pitch_kp"`
	RatePitchKI float64 `gorm:"type:decimal(8,4);default:0" json:"rate_pitch_ki"`
	RatePitchKD float64 `gorm:"type:decimal(8,4);default:0" json:"rate_pitch_kd"`
	RateYawKP   float64 `gorm:"type:decimal(8,4);default:0" json:"rate_yaw_kp"`
	RateYawKI   float64 `gorm:"type:decimal(8,4);default:0" json:"rate_yaw_ki"`
	RateYawKD   float64 `gorm:"type:decimal(8,4);default:0" json:"rate_yaw_kd"`

	AltKP float64 `gorm:"type:decimal(8,4);default:0" json:"alt_kp"`
	AltKI float64 `gorm:"type:decimal(8,4);default:0" json:"alt_ki"`
	AltKD float64 `gorm:"type:decimal(8,4);default:0" json:"alt_kd"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PIDGainProfile) TableName() string {
	return "pid_gain_profiles"
}

type ThrustLearningSample struct {
	ID        uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID     uint64  `gorm:"index;not null" json:"uav_id"`
	Throttle  float64 `json:"throttle"`
	AccelZ    float64 `json:"accel_z"`
	Altitude  float64 `json:"altitude"`
	VZ        float64 `json:"vz"`
	MotorPWM1 uint16  `json:"motor_pwm_1"`
	MotorPWM2 uint16  `json:"motor_pwm_2"`
	MotorPWM3 uint16  `json:"motor_pwm_3"`
	MotorPWM4 uint16  `json:"motor_pwm_4"`
	Voltage   float64 `json:"voltage"`
	Timestamp int64   `json:"timestamp"`
}

func (ThrustLearningSample) TableName() string {
	return "thrust_learning_samples"
}
