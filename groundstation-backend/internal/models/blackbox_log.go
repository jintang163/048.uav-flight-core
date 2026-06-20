package models

import (
	"time"

	"gorm.io/gorm"
)

type BlackboxLogStatus string

const (
	BlackboxStatusUploading BlackboxLogStatus = "uploading"
	BlackboxStatusUploaded  BlackboxLogStatus = "uploaded"
	BlackboxStatusAnalyzed  BlackboxLogStatus = "analyzed"
	BlackboxStatusError     BlackboxLogStatus = "error"
)

type BlackboxLog struct {
	ID            uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string              `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	UAVID         uint64              `gorm:"index;not null" json:"uav_id"`
	MissionID     uint64              `gorm:"index" json:"mission_id"`
	FlightName    string              `gorm:"type:varchar(200)" json:"flight_name"`
	StartTime     *time.Time          `json:"start_time"`
	EndTime       *time.Time          `json:"end_time"`
	Duration      int                 `json:"duration"`
	FileSize      int64               `json:"file_size"`
	FileName      string              `gorm:"type:varchar(200)" json:"file_name"`
	FileURL       string              `gorm:"type:varchar(500)" json:"file_url"`
	LogType       string              `gorm:"type:varchar(50);default:'bin'" json:"log_type"`
	Status        BlackboxLogStatus   `gorm:"type:varchar(20);default:'uploading'" json:"status"`
	FileHash      string              `gorm:"type:varchar(64)" json:"file_hash"`
	UploaderID    uint64              `json:"uploader_id"`
	MaxAltitude   float64             `gorm:"type:decimal(8,2)" json:"max_altitude"`
	MaxSpeed      float64             `gorm:"type:decimal(6,2)" json:"max_speed"`
	Distance      float64             `gorm:"type:decimal(10,2)" json:"distance"`
	BatteryUsed   float64             `gorm:"type:decimal(5,2)" json:"battery_used"`
	CrashDetected bool                `gorm:"default:false" json:"crash_detected"`
	Tags          string              `gorm:"type:varchar(500)" json:"tags"`
	Notes         string              `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	DeletedAt     gorm.DeletedAt      `gorm:"index" json:"-"`

	UAV     *UAV          `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
	Mission *FlightMission `gorm:"foreignKey:MissionID" json:"mission,omitempty"`
}

func (BlackboxLog) TableName() string {
	return "blackbox_logs"
}

type LogAnalysisReport struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	LogID          uint64         `gorm:"index;not null" json:"log_id"`
	AnalyzerID     uint64         `json:"analyzer_id"`
	ReportType     string         `gorm:"type:varchar(50)" json:"report_type"`
	Summary        string         `gorm:"type:text" json:"summary"`
	FlightScore    int            `json:"flight_score"`
	Anomalies      string         `gorm:"type:json" json:"anomalies"`
	Recommendations string         `gorm:"type:text" json:"recommendations"`
	ReportData     string         `gorm:"type:json" json:"report_data"`
	ReportURL      string         `gorm:"type:varchar(500)" json:"report_url"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	Log *BlackboxLog `gorm:"foreignKey:LogID" json:"log,omitempty"`
}

func (LogAnalysisReport) TableName() string {
	return "log_analysis_reports"
}
