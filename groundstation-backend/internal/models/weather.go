package models

import (
	"time"

	"gorm.io/gorm"
)

type WeatherCondition string

const (
	WeatherConditionClear     WeatherCondition = "clear"
	WeatherConditionCloudy    WeatherCondition = "cloudy"
	WeatherConditionRain      WeatherCondition = "rain"
	WeatherConditionSnow      WeatherCondition = "snow"
	WeatherConditionThunderstorm WeatherCondition = "thunderstorm"
	WeatherConditionFog       WeatherCondition = "fog"
	WeatherConditionHail      WeatherCondition = "hail"
)

type WeatherAlertType string

const (
	WeatherAlertHighWind        WeatherAlertType = "high_wind"
	WeatherAlertGust            WeatherAlertType = "gust"
	WeatherAlertThunderstorm    WeatherAlertType = "thunderstorm"
	WeatherAlertLowTemperature  WeatherAlertType = "low_temperature"
	WeatherAlertHeavyRain       WeatherAlertType = "heavy_rain"
)

type WeatherAlertLevel string

const (
	WeatherAlertLevelInfo     WeatherAlertLevel = "info"
	WeatherAlertLevelWarning  WeatherAlertLevel = "warning"
	WeatherAlertLevelCritical WeatherAlertLevel = "critical"
)

type WeatherData struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID           uint64           `gorm:"index;not null" json:"uav_id"`
	Source          string           `gorm:"type:varchar(20);not null" json:"source"`
	WindSpeed       float64          `gorm:"type:decimal(5,2);not null" json:"wind_speed"`
	WindDirection   float64          `gorm:"type:decimal(5,1)" json:"wind_direction"`
	WindGustSpeed   float64          `gorm:"type:decimal(5,2)" json:"wind_gust_speed"`
	Temperature     float64          `gorm:"type:decimal(5,1)" json:"temperature"`
	Humidity        float64          `gorm:"type:decimal(5,1)" json:"humidity"`
	Pressure        float64          `gorm:"type:decimal(7,1)" json:"pressure"`
	Visibility      float64          `gorm:"type:decimal(6,1)" json:"visibility"`
	Condition       WeatherCondition `gorm:"type:varchar(20)" json:"condition"`
	IsThunderstorm  bool             `gorm:"default:false" json:"is_thunderstorm"`
	Precipitation   float64          `gorm:"type:decimal(5,2)" json:"precipitation"`
	Latitude        float64          `gorm:"type:decimal(10,7)" json:"latitude"`
	Longitude       float64          `gorm:"type:decimal(10,7)" json:"longitude"`
	Altitude        float64          `gorm:"type:decimal(8,2)" json:"altitude"`
	CreatedAt       time.Time        `json:"created_at"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (WeatherData) TableName() string {
	return "weather_data"
}

type FlightWeatherLog struct {
	ID            uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	FlightID      uint64           `gorm:"index;not null" json:"flight_id"`
	UAVID         uint64           `gorm:"index;not null" json:"uav_id"`
	TakeoffWeatherID uint64        `json:"takeoff_weather_id"`
	LandingWeatherID uint64        `json:"landing_weather_id"`
	AvgWindSpeed  float64          `gorm:"type:decimal(5,2)" json:"avg_wind_speed"`
	MaxWindSpeed  float64          `gorm:"type:decimal(5,2)" json:"max_wind_speed"`
	MaxGustSpeed  float64          `gorm:"type:decimal(5,2)" json:"max_gust_speed"`
	AvgTemp       float64          `gorm:"type:decimal(5,1)" json:"avg_temperature"`
	MinTemp       float64          `gorm:"type:decimal(5,1)" json:"min_temperature"`
	Condition     WeatherCondition `gorm:"type:varchar(20)" json:"condition"`
	HadThunderstorm bool           `gorm:"default:false" json:"had_thunderstorm"`
	SampleCount   int              `json:"sample_count"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`

	TakeoffWeather *WeatherData `gorm:"foreignKey:TakeoffWeatherID" json:"takeoff_weather,omitempty"`
	LandingWeather *WeatherData `gorm:"foreignKey:LandingWeatherID" json:"landing_weather,omitempty"`
}

func (FlightWeatherLog) TableName() string {
	return "flight_weather_logs"
}

type WeatherAlertEvent struct {
	ID          uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	UAVID       uint64           `gorm:"index;not null" json:"uav_id"`
	AlertType   WeatherAlertType `gorm:"type:varchar(30);not null" json:"alert_type"`
	AlertLevel  WeatherAlertLevel `gorm:"type:varchar(20);not null" json:"alert_level"`
	WindSpeed   float64          `gorm:"type:decimal(5,2)" json:"wind_speed"`
	GustSpeed   float64          `gorm:"type:decimal(5,2)" json:"gust_speed"`
	Temperature float64          `gorm:"type:decimal(5,1)" json:"temperature"`
	Message     string           `gorm:"type:text;not null" json:"message"`
	ActionTaken string           `gorm:"type:varchar(50)" json:"action_taken"`
	IsResolved  bool             `gorm:"default:false" json:"is_resolved"`
	ResolvedAt  *time.Time       `json:"resolved_at"`
	CreatedAt   time.Time        `json:"created_at"`

	UAV *UAV `gorm:"foreignKey:UAVID" json:"uav,omitempty"`
}

func (WeatherAlertEvent) TableName() string {
	return "weather_alert_events"
}
