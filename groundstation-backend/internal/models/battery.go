package models

import (
	"time"

	"gorm.io/gorm"
)

type BatteryStatus string

const (
	BatteryStatusIdle      BatteryStatus = "idle"
	BatteryStatusCharging  BatteryStatus = "charging"
	BatteryStatusInUse     BatteryStatus = "in_use"
	BatteryStatusDischarging BatteryStatus = "discharging"
	BatteryStatusStorage   BatteryStatus = "storage"
	BatteryStatusFault     BatteryStatus = "fault"
)

type BatteryHealthStatus string

const (
	BatteryHealthExcellent BatteryHealthStatus = "excellent"
	BatteryHealthGood      BatteryHealthStatus = "good"
	BatteryHealthFair      BatteryHealthStatus = "fair"
	BatteryHealthPoor      BatteryHealthStatus = "poor"
	BatteryHealthCritical  BatteryHealthStatus = "critical"
)

type Battery struct {
	ID                 uint64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID               string              `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	BatteryID          string              `gorm:"type:varchar(100);uniqueIndex;not null" json:"battery_id"`
	Model              string              `gorm:"type:varchar(100)" json:"model"`
	Manufacturer       string              `gorm:"type:varchar(100)" json:"manufacturer"`
	Capacity           float64             `gorm:"type:decimal(10,2)" json:"capacity"`
	CapacityUnit       string              `gorm:"type:varchar(10);default:'mAh'" json:"capacity_unit"`
	Voltage            float64             `gorm:"type:decimal(5,2)" json:"voltage"`
	CellCount          int                 `json:"cell_count"`
	CurrentVoltage     float64             `gorm:"type:decimal(5,2)" json:"current_voltage"`
	CurrentLevel       float64             `gorm:"type:decimal(5,2)" json:"current_level"`
	CurrentTemperature float64             `gorm:"type:decimal(5,2)" json:"current_temperature"`
	CurrentCurrent     float64             `gorm:"type:decimal(5,2)" json:"current_current"`
	SOH                float64             `gorm:"type:decimal(5,2);default:100" json:"soh"`
	HealthStatus       BatteryHealthStatus `gorm:"type:varchar(20);default:'excellent'" json:"health_status"`
	Status             BatteryStatus       `gorm:"type:varchar(20);default:'idle'" json:"status"`
	CycleCount         int                 `gorm:"default:0" json:"cycle_count"`
	TotalFlightTime    int                 `gorm:"default:0" json:"total_flight_time"`
	TotalChargeCount   int                 `gorm:"default:0" json:"total_charge_count"`
	ManufactureDate    *time.Time          `json:"manufacture_date"`
	FirstUseDate       *time.Time          `json:"first_use_date"`
	LastUsedAt         *time.Time          `json:"last_used_at"`
	LastChargedAt      *time.Time          `json:"last_charged_at"`
	StorageDays        int                 `gorm:"default:0" json:"storage_days"`
	NeedsMaintenance   bool                `gorm:"default:false" json:"needs_maintenance"`
	MaintenanceMessage string              `gorm:"type:varchar(255)" json:"maintenance_message"`
	UAVID              *uint64             `gorm:"index" json:"uav_id"`
	Location           string              `gorm:"type:varchar(100)" json:"location"`
	Notes              string              `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	DeletedAt          gorm.DeletedAt      `gorm:"index" json:"-"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (Battery) TableName() string {
	return "batteries"
}

type BatteryUsageRecord struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID          string     `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	BatteryID     uint64     `gorm:"index;not null" json:"battery_id"`
	UAVID         uint64     `gorm:"index" json:"uav_id"`
	FlightMissionID *uint64  `gorm:"index" json:"flight_mission_id"`
	StartLevel    float64    `gorm:"type:decimal(5,2)" json:"start_level"`
	EndLevel      float64    `gorm:"type:decimal(5,2)" json:"end_level"`
	StartVoltage  float64    `gorm:"type:decimal(5,2)" json:"start_voltage"`
	EndVoltage    float64    `gorm:"type:decimal(5,2)" json:"end_voltage"`
	MaxTemperature float64   `gorm:"type:decimal(5,2)" json:"max_temperature"`
	AvgCurrent    float64    `gorm:"type:decimal(5,2)" json:"avg_current"`
	MaxCurrent    float64    `gorm:"type:decimal(5,2)" json:"max_current"`
	Duration      int        `json:"duration"`
	Distance      float64    `gorm:"type:decimal(10,2)" json:"distance"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	CellVoltages  string     `gorm:"type:json" json:"cell_voltages,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Battery *Battery `gorm:"foreignKey:BatteryID" json:"battery,omitempty"`
	UAV     *UAV     `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (BatteryUsageRecord) TableName() string {
	return "battery_usage_records"
}

type BatteryCellData struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	BatteryID  uint64         `gorm:"index;not null" json:"battery_id"`
	CellIndex  int            `gorm:"not null" json:"cell_index"`
	Voltage    float64        `gorm:"type:decimal(6,3)" json:"voltage"`
	Resistance float64        `gorm:"type:decimal(6,2)" json:"resistance"`
	Status     string         `gorm:"type:varchar(20);default:'normal'" json:"status"`
	RecordedAt time.Time      `json:"recorded_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Battery *Battery `gorm:"foreignKey:BatteryID" json:"battery,omitempty"`
}

func (BatteryCellData) TableName() string {
	return "battery_cell_data"
}

type ChargingStationStatus string

const (
	ChargingStationStatusOnline  ChargingStationStatus = "online"
	ChargingStationStatusOffline ChargingStationStatus = "offline"
	ChargingStationStatusFault   ChargingStationStatus = "fault"
	ChargingStationStatusMaintenance ChargingStationStatus = "maintenance"
)

type ChargingStation struct {
	ID             uint64                `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID           string                `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	StationID      string                `gorm:"type:varchar(100);uniqueIndex;not null" json:"station_id"`
	Name           string                `gorm:"type:varchar(100);not null" json:"name"`
	Model          string                `gorm:"type:varchar(100)" json:"model"`
	Manufacturer   string                `gorm:"type:varchar(100)" json:"manufacturer"`
	Status         ChargingStationStatus `gorm:"type:varchar(20);default:'offline'" json:"status"`
	SlotCount      int                   `gorm:"default:0" json:"slot_count"`
	OccupiedSlots  int                   `gorm:"default:0" json:"occupied_slots"`
	ChargingSlots  int                   `gorm:"default:0" json:"charging_slots"`
	TotalCharged   int                   `gorm:"default:0" json:"total_charged"`
	Location       string                `gorm:"type:varchar(200)" json:"location"`
	IPAddress      string                `gorm:"type:varchar(45)" json:"ip_address"`
	Port           int                   `json:"port"`
	Protocol       string                `gorm:"type:varchar(20);default:'tcp'" json:"protocol"`
	LastOnlineAt   *time.Time            `json:"last_online_at"`
	FirmwareVersion string               `gorm:"type:varchar(50)" json:"firmware_version"`
	MaxVoltage     float64               `gorm:"type:decimal(5,2)" json:"max_voltage"`
	MaxCurrent     float64               `gorm:"type:decimal(5,2)" json:"max_current"`
	Description    string                `gorm:"type:text" json:"description"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	DeletedAt      gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (ChargingStation) TableName() string {
	return "charging_stations"
}

type ChargingSlot struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	StationID       uint64         `gorm:"index;not null" json:"station_id"`
	SlotIndex       int            `gorm:"not null" json:"slot_index"`
	SlotName        string         `gorm:"type:varchar(50)" json:"slot_name"`
	Status          string         `gorm:"type:varchar(20);default:'empty'" json:"status"`
	BatteryID       *uint64        `gorm:"index" json:"battery_id"`
	ChargingMode    string         `gorm:"type:varchar(20)" json:"charging_mode"`
	TargetVoltage   float64        `gorm:"type:decimal(5,2)" json:"target_voltage"`
	TargetCurrent   float64        `gorm:"type:decimal(5,2)" json:"target_current"`
	CurrentVoltage  float64        `gorm:"type:decimal(5,2)" json:"current_voltage"`
	CurrentCurrent  float64        `gorm:"type:decimal(5,2)" json:"current_current"`
	CurrentLevel    float64        `gorm:"type:decimal(5,2)" json:"current_level"`
	Temperature     float64        `gorm:"type:decimal(5,2)" json:"temperature"`
	ChargedCapacity float64        `gorm:"type:decimal(10,2)" json:"charged_capacity"`
	ChargingTime    int            `json:"charging_time"`
	RemainingTime   int            `json:"remaining_time"`
	StartTime       *time.Time     `json:"start_time"`
	EndTime         *time.Time     `json:"end_time"`
	FaultCode       int            `json:"fault_code"`
	FaultMessage    string         `gorm:"type:varchar(255)" json:"fault_message"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Station *ChargingStation `gorm:"foreignKey:StationID" json:"station,omitempty"`
	Battery *Battery         `gorm:"foreignKey:BatteryID" json:"battery,omitempty"`
}

func (ChargingSlot) TableName() string {
	return "charging_slots"
}

type ChargingRecord struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID              string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	BatteryID         uint64         `gorm:"index;not null" json:"battery_id"`
	StationID         uint64         `gorm:"index;not null" json:"station_id"`
	SlotID            uint64         `gorm:"index" json:"slot_id"`
	StartLevel        float64        `gorm:"type:decimal(5,2)" json:"start_level"`
	EndLevel          float64        `gorm:"type:decimal(5,2)" json:"end_level"`
	StartVoltage      float64        `gorm:"type:decimal(5,2)" json:"start_voltage"`
	EndVoltage        float64        `gorm:"type:decimal(5,2)" json:"end_voltage"`
	ChargingMode      string         `gorm:"type:varchar(20)" json:"charging_mode"`
	ChargingCurrent   float64        `gorm:"type:decimal(5,2)" json:"charging_current"`
	MaxTemperature    float64        `gorm:"type:decimal(5,2)" json:"max_temperature"`
	AvgTemperature    float64        `gorm:"type:decimal(5,2)" json:"avg_temperature"`
	ChargedCapacity   float64        `gorm:"type:decimal(10,2)" json:"charged_capacity"`
	ChargingTime      int            `json:"charging_time"`
	EnergyConsumed    float64        `gorm:"type:decimal(10,2)" json:"energy_consumed"`
	Status            string         `gorm:"type:varchar(20);default:'completed'" json:"status"`
	StartTime         *time.Time     `json:"start_time"`
	EndTime           *time.Time     `json:"end_time"`
	CellVoltagesStart string         `gorm:"type:json" json:"cell_voltages_start,omitempty"`
	CellVoltagesEnd   string         `gorm:"type:json" json:"cell_voltages_end,omitempty"`
	FaultCode         int            `json:"fault_code"`
	FaultMessage      string         `gorm:"type:varchar(255)" json:"fault_message"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	Battery *Battery         `gorm:"foreignKey:BatteryID" json:"battery,omitempty"`
	Station *ChargingStation `gorm:"foreignKey:StationID" json:"station,omitempty"`
}

func (ChargingRecord) TableName() string {
	return "charging_records"
}

type BatteryMaintenanceAlert struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UUID        string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"uuid"`
	BatteryID   uint64         `gorm:"index;not null" json:"battery_id"`
	AlertType   string         `gorm:"type:varchar(50);not null" json:"alert_type"`
	Level       AlertLevel     `gorm:"type:varchar(20);not null" json:"level"`
	Title       string         `gorm:"type:varchar(200);not null" json:"title"`
	Message     string         `gorm:"type:text;not null" json:"message"`
	Status      AlertStatus    `gorm:"type:varchar(20);default:'new'" json:"status"`
	StorageDays int            `json:"storage_days"`
	SOH         float64        `gorm:"type:decimal(5,2)" json:"soh"`
	AcknowledgedBy *uint64     `json:"acknowledged_by"`
	AcknowledgedAt *time.Time  `json:"acknowledged_at"`
	ResolvedBy  *uint64        `json:"resolved_by"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	ResolvedNote string        `gorm:"type:text" json:"resolved_note"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Battery *Battery `gorm:"foreignKey:BatteryID" json:"battery,omitempty"`
}

func (BatteryMaintenanceAlert) TableName() string {
	return "battery_maintenance_alerts"
}
