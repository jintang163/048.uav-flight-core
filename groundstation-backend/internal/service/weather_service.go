package service

import (
	"encoding/json"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type WeatherService struct {
	repo         *repository.WeatherRepository
	alertRepo    *repository.AlertRepository
	latestData   map[uint64]*models.WeatherData
	latestMu     sync.RWMutex
	gustWindow   map[uint64][]float64
	gustMu       sync.Mutex
	flightLogs   map[uint64]*models.FlightWeatherLog
	flightLogMu  sync.Mutex
}

var (
	windSpeedReturnThreshold   float64 = 5.0
	gustSpeedProtectThreshold  float64 = 12.0
	thunderstormRejectTakeoff  bool    = true
	windAdaptSpeedThreshold    float64 = 8.0
	lowTempThreshold           float64 = -10.0
)

type WeatherThresholds struct {
	WindSpeedReturnMs    float64 `json:"wind_speed_return_ms"`
	GustProtectMs        float64 `json:"gust_protect_ms"`
	WindAdaptMs          float64 `json:"wind_adapt_ms"`
	LowTempC             float64 `json:"low_temp_c"`
	ThunderstormReject   bool    `json:"thunderstorm_reject"`
}

var DefaultWeatherThresholds = &WeatherThresholds{
	WindSpeedReturnMs:   5.0,
	GustProtectMs:       12.0,
	WindAdaptMs:         8.0,
	LowTempC:            -10.0,
	ThunderstormReject:  true,
}

func NewWeatherService() *WeatherService {
	return &WeatherService{
		repo:       repository.NewWeatherRepository(),
		alertRepo:  repository.NewAlertRepository(),
		latestData: make(map[uint64]*models.WeatherData),
		gustWindow: make(map[uint64][]float64),
		flightLogs: make(map[uint64]*models.FlightWeatherLog),
	}
}

type WeatherSensorData struct {
	UAVID         uint64  `json:"uav_id"`
	WindSpeed     float64 `json:"wind_speed"`
	WindDirection float64 `json:"wind_direction"`
	WindGustSpeed float64 `json:"wind_gust_speed"`
	Temperature   float64 `json:"temperature"`
	Humidity      float64 `json:"humidity"`
	Pressure      float64 `json:"pressure"`
	Condition     string  `json:"condition"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Altitude      float64 `json:"altitude"`
}

type OpenWeatherResponse struct {
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
		Pressure float64 `json:"pressure"`
	} `json:"main"`
	Weather []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Main        string `json:"main"`
	} `json:"weather"`
	Visibility float64 `json:"visibility"`
	Rain       struct {
		OneH float64 `json:"1h"`
	} `json:"rain"`
}

func (s *WeatherService) ProcessSensorData(data *WeatherSensorData) error {
	condition := models.WeatherConditionClear
	if data.Condition != "" {
		condition = models.WeatherCondition(data.Condition)
	}

	isThunderstorm := condition == models.WeatherConditionThunderstorm

	weatherData := &models.WeatherData{
		UAVID:         data.UAVID,
		Source:        "sensor",
		WindSpeed:     data.WindSpeed,
		WindDirection: data.WindDirection,
		WindGustSpeed: data.WindGustSpeed,
		Temperature:   data.Temperature,
		Humidity:      data.Humidity,
		Pressure:      data.Pressure,
		Condition:     condition,
		IsThunderstorm: isThunderstorm,
		Latitude:      data.Latitude,
		Longitude:     data.Longitude,
		Altitude:      data.Altitude,
	}

	if err := s.repo.CreateWeatherData(weatherData); err != nil {
		return fmt.Errorf("保存气象数据失败: %w", err)
	}

	s.latestMu.Lock()
	s.latestData[data.UAVID] = weatherData
	s.latestMu.Unlock()

	s.updateFlightWeatherLog(data)

	s.checkAndTriggerAlerts(weatherData)

	s.broadcastWeatherUpdate(weatherData)

	return nil
}

func (s *WeatherService) FetchWeatherFromAPI(lat, lon float64) (*models.WeatherData, error) {
	apiKey := ""
	if config.AppConfig.ExternalServices != nil {
		if key, ok := config.AppConfig.ExternalServices["openweather_api_key"]; ok {
			apiKey = key
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("气象API未配置")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?lat=%.6f&lon=%.6f&appid=%s&units=metric",
		lat, lon, apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求气象API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取气象API响应失败: %w", err)
	}

	var apiResp OpenWeatherResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析气象API响应失败: %w", err)
	}

	condition := mapWeatherCondition(apiResp)

	return &models.WeatherData{
		Source:         "api",
		WindSpeed:      apiResp.Wind.Speed,
		WindDirection:  apiResp.Wind.Deg,
		WindGustSpeed:  apiResp.Wind.Gust,
		Temperature:    apiResp.Main.Temp,
		Humidity:       apiResp.Main.Humidity,
		Pressure:       apiResp.Main.Pressure,
		Visibility:     apiResp.Visibility,
		Condition:      condition,
		IsThunderstorm: condition == models.WeatherConditionThunderstorm,
		Precipitation:  apiResp.Rain.OneH,
		Latitude:       lat,
		Longitude:      lon,
	}, nil
}

func (s *WeatherService) checkAndTriggerAlerts(data *models.WeatherData) {
	thresholds := DefaultWeatherThresholds

	if data.WindSpeed > thresholds.WindAdaptMs {
		s.triggerWindAdaptation(data)
	}

	if data.WindSpeed > thresholds.WindSpeedReturnMs {
		s.createAndBroadcastAlert(data, models.WeatherAlertHighWind,
			models.WeatherAlertLevelWarning,
			fmt.Sprintf("风速 %.1fm/s 超过安全阈值 %.1fm/s，建议返航", data.WindSpeed, thresholds.WindSpeedReturnMs),
			"suggest_rth")
	}

	if data.WindGustSpeed > thresholds.GustProtectMs {
		s.recordGust(data)
		s.triggerGustProtection(data)
		s.createAndBroadcastAlert(data, models.WeatherAlertGust,
			models.WeatherAlertLevelCritical,
			fmt.Sprintf("阵风 %.1fm/s 超过保护阈值 %.1fm/s，已触发姿态保护", data.WindGustSpeed, thresholds.GustProtectMs),
			"attitude_protection")
	}

	if data.IsThunderstorm {
		s.createAndBroadcastAlert(data, models.WeatherAlertThunderstorm,
			models.WeatherAlertLevelCritical,
			"检测到雷暴天气，起飞已被自动拒绝",
			"reject_takeoff")
	}

	if data.Temperature < thresholds.LowTempC {
		s.createAndBroadcastAlert(data, models.WeatherAlertLowTemperature,
			models.WeatherAlertLevelWarning,
			fmt.Sprintf("温度 %.1f°C 低于安全阈值 %.1f°C，注意电池性能下降", data.Temperature, thresholds.LowTempC),
			"warn_low_temp")
	}
}

func (s *WeatherService) triggerWindAdaptation(data *models.WeatherData) {
	speedFactor := 1.0
	if data.WindSpeed >= 12.0 {
		speedFactor = 0.4
	} else if data.WindSpeed >= 10.0 {
		speedFactor = 0.5
	} else if data.WindSpeed >= 8.0 {
		speedFactor = 0.7
	}

	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(data.UAVID, "wind_adapt_speed", map[string]interface{}{
		"param1": float32(speedFactor),
		"param2": float32(data.WindSpeed),
		"param3": float32(data.WindDirection),
	})

	middleware.Logger.Info("风速自适应: 降低飞行速度",
		zap.Uint64("uav_id", data.UAVID),
		zap.Float64("wind_speed", data.WindSpeed),
		zap.Float64("speed_factor", speedFactor),
	)
}

func (s *WeatherService) recordGust(data *models.WeatherData) {
	s.gustMu.Lock()
	defer s.gustMu.Unlock()

	window := s.gustWindow[data.UAVID]
	window = append(window, data.WindGustSpeed)
	if len(window) > 10 {
		window = window[len(window)-10:]
	}
	s.gustWindow[data.UAVID] = window
}

func (s *WeatherService) triggerGustProtection(data *models.WeatherData) {
	cmdMgr := mavlink.NewCommandManager()
	_ = cmdMgr.SendCustomCommand(data.UAVID, "gust_protection", map[string]interface{}{
		"param1": float32(data.WindGustSpeed),
		"param2": float32(data.WindDirection),
		"param3": float32(1),
	})

	middleware.Logger.Warn("阵风保护触发",
		zap.Uint64("uav_id", data.UAVID),
		zap.Float64("gust_speed", data.WindGustSpeed),
	)
}

func (s *WeatherService) createAndBroadcastAlert(data *models.WeatherData, alertType models.WeatherAlertType, level models.WeatherAlertLevel, message string, action string) {
	alert := &models.WeatherAlertEvent{
		UAVID:       data.UAVID,
		AlertType:   alertType,
		AlertLevel:  level,
		WindSpeed:   data.WindSpeed,
		GustSpeed:   data.WindGustSpeed,
		Temperature: data.Temperature,
		Message:     message,
		ActionTaken: action,
	}

	if err := s.repo.CreateWeatherAlert(alert); err != nil {
		middleware.Logger.Error("创建气象预警失败", zap.Error(err))
		return
	}

	alertEvent := &models.AlertEvent{
		UAVID: data.UAVID,
		Type:  models.AlertTypeCustom,
		Level: models.AlertLevelCritical,
		Title: string(alertType),
		Message: message,
	}
	_ = s.alertRepo.Create(alertEvent)

	websocket.NewHub().BroadcastWeatherAlert(data.UAVID, map[string]interface{}{
		"alert_id":     alert.ID,
		"uav_id":       data.UAVID,
		"alert_type":   string(alertType),
		"alert_level":  string(level),
		"wind_speed":   data.WindSpeed,
		"gust_speed":   data.WindGustSpeed,
		"temperature":  data.Temperature,
		"message":      message,
		"action_taken": action,
		"timestamp":    time.Now().UnixNano() / 1e6,
	})
}

func (s *WeatherService) broadcastWeatherUpdate(data *models.WeatherData) {
	websocket.NewHub().BroadcastWeatherData(data.UAVID, map[string]interface{}{
		"uav_id":         data.UAVID,
		"source":         data.Source,
		"wind_speed":     data.WindSpeed,
		"wind_direction": data.WindDirection,
		"wind_gust_speed": data.WindGustSpeed,
		"temperature":    data.Temperature,
		"humidity":       data.Humidity,
		"pressure":       data.Pressure,
		"condition":      string(data.Condition),
		"is_thunderstorm": data.IsThunderstorm,
		"timestamp":      time.Now().UnixNano() / 1e6,
	})
}

func (s *WeatherService) updateFlightWeatherLog(data *WeatherSensorData) {
	s.flightLogMu.Lock()
	defer s.flightLogMu.Unlock()

	log, exists := s.flightLogs[data.UAVID]
	if !exists {
		return
	}

	log.SampleCount++
	log.MaxWindSpeed = maxFloat(log.MaxWindSpeed, data.WindSpeed)
	log.MaxGustSpeed = maxFloat(log.MaxGustSpeed, data.WindGustSpeed)
	log.AvgWindSpeed = (log.AvgWindSpeed*float64(log.SampleCount-1) + data.WindSpeed) / float64(log.SampleCount)
	log.AvgTemp = (log.AvgTemp*float64(log.SampleCount-1) + data.Temperature) / float64(log.SampleCount)
	log.MinTemp = minFloat(log.MinTemp, data.Temperature)

	if data.Condition == string(models.WeatherConditionThunderstorm) {
		log.HadThunderstorm = true
	}

	condition := models.WeatherCondition(data.Condition)
	if condition != "" {
		log.Condition = condition
	}

	_ = s.repo.UpdateFlightWeatherLog(log)
}

func (s *WeatherService) StartFlightWeatherLog(uavID, flightID uint64, takeoffWeatherID uint64) error {
	log := &models.FlightWeatherLog{
		FlightID:        flightID,
		UAVID:           uavID,
		TakeoffWeatherID: takeoffWeatherID,
		SampleCount:     0,
		MinTemp:         999.0,
	}

	s.flightLogMu.Lock()
	s.flightLogs[uavID] = log
	s.flightLogMu.Unlock()

	return s.repo.CreateFlightWeatherLog(log)
}

func (s *WeatherService) EndFlightWeatherLog(uavID uint64, landingWeatherID uint64) error {
	s.flightLogMu.Lock()
	log, exists := s.flightLogs[uavID]
	if exists {
		delete(s.flightLogs, uavID)
	}
	s.flightLogMu.Unlock()

	if !exists {
		return fmt.Errorf("飞行气象日志不存在: uav_%d", uavID)
	}

	log.LandingWeatherID = landingWeatherID
	return s.repo.UpdateFlightWeatherLog(log)
}

func (s *WeatherService) GetLatestWeather(uavID uint64) (*models.WeatherData, error) {
	s.latestMu.RLock()
	if data, ok := s.latestData[uavID]; ok {
		s.latestMu.RUnlock()
		return data, nil
	}
	s.latestMu.RUnlock()

	return s.repo.GetLatestByUAVID(uavID)
}

func (s *WeatherService) GetWeatherHistory(uavID uint64, start, end time.Time) ([]models.WeatherData, error) {
	return s.repo.GetWeatherHistory(uavID, start, end)
}

func (s *WeatherService) GetActiveAlerts(uavID uint64) ([]models.WeatherAlertEvent, error) {
	return s.repo.GetActiveAlerts(uavID)
}

func (s *WeatherService) ListAlerts(pagination *utils.Pagination, uavID uint64, alertType, level string) ([]models.WeatherAlertEvent, int64, error) {
	return s.repo.ListAlerts(pagination, uavID, alertType, level)
}

func (s *WeatherService) ResolveAlert(id uint64) error {
	return s.repo.ResolveWeatherAlert(id)
}

func (s *WeatherService) CheckTakeoffWeather(uavID uint64) (*WeatherCheckResult, error) {
	latest, err := s.GetLatestWeather(uavID)
	if err != nil {
		return &WeatherCheckResult{
			CanTakeoff:   true,
			Warnings:     []string{"无气象数据，默认允许起飞"},
			WeatherData:  nil,
		}, nil
	}

	result := &WeatherCheckResult{
		WeatherData:  latest,
		CanTakeoff:   true,
		Warnings:     []string{},
		BlockReasons: []string{},
	}

	if latest.IsThunderstorm && DefaultWeatherThresholds.ThunderstormReject {
		result.CanTakeoff = false
		result.BlockReasons = append(result.BlockReasons, "雷暴天气，禁止起飞")
	}

	if latest.WindSpeed > DefaultWeatherThresholds.WindAdaptMs {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("风速 %.1fm/s 较大，将自动降低飞行速度", latest.WindSpeed))
	}

	if latest.WindSpeed > DefaultWeatherThresholds.WindSpeedReturnMs {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("风速 %.1fm/s 超过安全阈值，建议不要起飞", latest.WindSpeed))
	}

	if latest.WindGustSpeed > DefaultWeatherThresholds.GustProtectMs {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("阵风 %.1fm/s 较强，飞行稳定性受影响", latest.WindGustSpeed))
	}

	if latest.Temperature < DefaultWeatherThresholds.LowTempC {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("温度 %.1f°C 过低，电池性能将下降", latest.Temperature))
	}

	return result, nil
}

type WeatherCheckResult struct {
	CanTakeoff   bool                `json:"can_takeoff"`
	Warnings     []string            `json:"warnings"`
	BlockReasons []string            `json:"block_reasons"`
	WeatherData  *models.WeatherData `json:"weather_data,omitempty"`
}

func (s *WeatherService) GetThresholds() *WeatherThresholds {
	return DefaultWeatherThresholds
}

func (s *WeatherService) UpdateThresholds(t *WeatherThresholds) {
	DefaultWeatherThresholds = t
	windSpeedReturnThreshold = t.WindSpeedReturnMs
	gustSpeedProtectThreshold = t.GustProtectMs
	windAdaptSpeedThreshold = t.WindAdaptMs
	lowTempThreshold = t.LowTempC
	thunderstormRejectTakeoff = t.ThunderstormReject
}

func (s *WeatherService) GetFlightWeatherLog(flightID uint64) (*models.FlightWeatherLog, error) {
	return s.repo.GetFlightWeatherLog(flightID)
}

func mapWeatherCondition(apiResp OpenWeatherResponse) models.WeatherCondition {
	if len(apiResp.Weather) == 0 {
		return models.WeatherConditionClear
	}

	weatherID := apiResp.Weather[0].ID
	switch {
	case weatherID >= 200 && weatherID < 300:
		return models.WeatherConditionThunderstorm
	case weatherID >= 300 && weatherID < 400:
		return models.WeatherConditionRain
	case weatherID >= 500 && weatherID < 600:
		return models.WeatherConditionRain
	case weatherID >= 600 && weatherID < 700:
		return models.WeatherConditionSnow
	case weatherID >= 700 && weatherID < 800:
		return models.WeatherConditionFog
	case weatherID == 800:
		return models.WeatherConditionClear
	default:
		return models.WeatherConditionCloudy
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
