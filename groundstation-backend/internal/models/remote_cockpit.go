package models

import (
	"time"

	"gorm.io/gorm"
)

type VideoCodec string

const (
	VideoCodecH264 VideoCodec = "h264"
	VideoCodecH265 VideoCodec = "h265"
)

type VideoResolution string

const (
	VideoRes480P   VideoResolution = "480x360"
	VideoRes640P   VideoResolution = "640x480"
	VideoRes960P   VideoResolution = "960x540"
	VideoRes720P   VideoResolution = "1280x720"
	VideoRes1080P  VideoResolution = "1920x1080"
)

type CockpitMode string

const (
	CockpitModeIdle         CockpitMode = "idle"
	CockpitModeConnecting   CockpitMode = "connecting"
	CockpitModeFlying       CockpitMode = "flying"
	CockpitModeMission      CockpitMode = "mission"
	CockpitModeEmergency    CockpitMode = "emergency"
	CockpitModeDisconnected CockpitMode = "disconnected"
)

type CockpitSession struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID        string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"session_id"`
	UAVID            uint64         `gorm:"index;not null" json:"uav_id"`
	PilotID          uint64         `gorm:"index;not null" json:"pilot_id"`
	StartTime        time.Time      `gorm:"index;not null" json:"start_time"`
	EndTime          *time.Time     `json:"end_time,omitempty"`
	Mode             CockpitMode    `gorm:"type:varchar(20);default:'idle'" json:"mode"`
	TotalFlightMs    int64          `json:"total_flight_time_ms"`
	CommandsSent     int64          `json:"commands_sent"`
	FailoverEvents   int            `json:"failover_events"`
	AutoFallbackUsed bool           `json:"auto_fallback_used"`
	CreatedAt        time.Time      `json:"created_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	UAV   *UAV   `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Pilot *User  `gorm:"foreignKey:PilotID" json:"pilot,omitempty"`
}

func (CockpitSession) TableName() string {
	return "cockpit_sessions"
}

type VideoStreamSession struct {
	ID               uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID            uint64         `gorm:"index;not null" json:"uav_id"`
	Codec            VideoCodec     `gorm:"type:varchar(10);default:'h265'" json:"codec"`
	Resolution       VideoResolution `gorm:"type:varchar(20);default:'1280x720'" json:"resolution"`
	TargetBitrateKbps int           `json:"target_bitrate_kbps"`
	CurrentBitrateKbps int          `json:"current_bitrate_kbps"`
	FPS              int            `json:"fps"`
	KeyframeInterval int            `json:"keyframe_interval"`
	AdaptiveEnabled  bool           `gorm:"default:true" json:"adaptive_enabled"`
	MinBitrateKbps   int            `json:"min_bitrate_kbps"`
	MaxBitrateKbps   int            `json:"max_bitrate_kbps"`
	MinResolution    VideoResolution `gorm:"type:varchar(20);default:'640x480'" json:"min_resolution"`
	MaxResolution    VideoResolution `gorm:"type:varchar(20);default:'1920x1080'" json:"max_resolution"`
	StreamURL        string         `gorm:"type:varchar(512)" json:"stream_url"`
	Protocol         string         `gorm:"type:varchar(20);default:'webrtc'" json:"protocol"`
	Active           bool           `gorm:"default:false" json:"active"`
	FramesDecoded    int64          `json:"frames_decoded"`
	FramesDropped    int64          `json:"frames_dropped"`
	LatencyMs        int            `json:"latency_ms"`
	JitterMs         int            `json:"jitter_ms"`
	PacketLoss       float64        `json:"packet_loss"`
	LastFrameTime    *time.Time     `json:"last_frame_time,omitempty"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	StoppedAt        *time.Time     `json:"stopped_at,omitempty"`
	QualityAdjustments int          `json:"quality_adjustments"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (VideoStreamSession) TableName() string {
	return "video_stream_sessions"
}

type CockpitLinkSnapshot struct {
	ID                   uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID                uint64         `gorm:"index;not null" json:"uav_id"`
	PrimaryLink          LinkType       `gorm:"not null;default:2" json:"primary_link"`
	SecondaryLink        LinkType       `gorm:"not null;default:1" json:"secondary_link"`
	PrimaryState         LinkState      `gorm:"not null;default:0" json:"primary_state"`
	SecondaryState       LinkState      `gorm:"not null;default:0" json:"secondary_state"`
	PrimaryLatencyMs     int            `json:"primary_latency_ms"`
	SecondaryLatencyMs   int            `json:"secondary_latency_ms"`
	PrimaryPacketLoss    float64        `json:"primary_packet_loss"`
	SecondaryPacketLoss  float64        `json:"secondary_packet_loss"`
	FailoverEnabled      bool           `gorm:"default:true" json:"failover_enabled"`
	FailoverThresholdMs  int            `gorm:"default:200" json:"failover_threshold_ms"`
	FailoverCount        int            `json:"failover_count"`
	LastFailoverTime     *time.Time     `json:"last_failover_time,omitempty"`
	AutoMissionFallback  bool           `gorm:"default:true" json:"auto_mission_fallback"`
	Timestamp            time.Time      `gorm:"index;not null" json:"timestamp"`
	CreatedAt            time.Time      `json:"created_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CockpitLinkSnapshot) TableName() string {
	return "cockpit_link_snapshots"
}

type NetworkMetricsLog struct {
	ID                  uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID               uint64         `gorm:"index;not null" json:"uav_id"`
	BandwidthKbps       float64        `json:"bandwidth_estimate_kbps"`
	RTTms               int            `json:"rtt_ms"`
	PacketLoss          float64        `json:"packet_loss"`
	JitterMs            int            `json:"jitter_ms"`
	ThroughputKbps      float64        `json:"throughput_kbps"`
	Timestamp           time.Time      `gorm:"index;not null" json:"timestamp"`
	CreatedAt           time.Time      `json:"created_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (NetworkMetricsLog) TableName() string {
	return "network_metrics_logs"
}

type FlightControlLog struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID  string         `gorm:"type:varchar(64);index" json:"session_id"`
	UAVID      uint64         `gorm:"index;not null" json:"uav_id"`
	PilotID    uint64         `gorm:"index;not null" json:"pilot_id"`
	Pitch      float64        `json:"pitch"`
	Roll       float64        `json:"roll"`
	Yaw        float64        `json:"yaw"`
	Throttle   float64        `json:"throttle"`
	Source     string         `gorm:"type:varchar(20);default:'gamepad'" json:"source"`
	Timestamp  time.Time      `gorm:"index;not null" json:"timestamp"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FlightControlLog) TableName() string {
	return "flight_control_logs"
}
