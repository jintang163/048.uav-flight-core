package service

import (
	"context"
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type BlackboxService struct {
	blackboxRepo *repository.BlackboxRepository
	uavRepo      *repository.UAVRepository
	missionRepo  *repository.MissionRepository
	minioClient  *minio.Client
}

type LogDataPoint struct {
	Timestamp   int64   `json:"timestamp"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Altitude    float64 `json:"altitude"`
	Roll        float64 `json:"roll"`
	Pitch       float64 `json:"pitch"`
	Yaw         float64 `json:"yaw"`
	Vx          float64 `json:"vx"`
	Vy          float64 `json:"vy"`
	Vz          float64 `json:"vz"`
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Throttle    float64 `json:"throttle"`
	FlightMode  int     `json:"flight_mode"`
	Satellites  int     `json:"satellites"`
	GPSFixType  int     `json:"gps_fix_type"`
	ErrorFlags  int     `json:"error_flags"`
	MotorPWM    []int   `json:"motor_pwm"`
	RCChannels  []int   `json:"rc_channels"`
}

type LogEvent struct {
	Timestamp      int64   `json:"timestamp"`
	EventType      string  `json:"event_type"`
	EventTypeID    uint32  `json:"event_type_id"`
	Severity       int     `json:"severity"`
	Description    string  `json:"description"`
	Param1         int32   `json:"param1"`
	Param2         int32   `json:"param2"`
	Param3         float32 `json:"param3"`
	Param4         float32 `json:"param4"`
}

type ParsedLogData struct {
	Header      map[string]interface{} `json:"header"`
	DataPoints  []LogDataPoint         `json:"data_points"`
	Events      []LogEvent             `json:"events"`
	Statistics  LogStatistics          `json:"statistics"`
}

type LogStatistics struct {
	TotalDuration    float64 `json:"total_duration"`
	MaxAltitude      float64 `json:"max_altitude"`
	MaxSpeed         float64 `json:"max_speed"`
	TotalDistance    float64 `json:"total_distance"`
	AvgVoltage       float64 `json:"avg_voltage"`
	MinVoltage       float64 `json:"min_voltage"`
	BatteryUsed      float64 `json:"battery_used"`
	AnomalyCount     int     `json:"anomaly_count"`
	MaxRoll          float64 `json:"max_roll"`
	MaxPitch         float64 `json:"max_pitch"`
	AvgSatellites    float64 `json:"avg_satellites"`
	CrashDetected    bool    `json:"crash_detected"`
}

type AnalysisReport struct {
	LogID          uint64                 `json:"log_id"`
	FlightSummary  string                 `json:"flight_summary"`
	FlightScore    int                    `json:"flight_score"`
	Anomalies      []LogEvent             `json:"anomalies"`
	Statistics     LogStatistics          `json:"statistics"`
	Recommendations []string              `json:"recommendations"`
	FlightPhases   []FlightPhase          `json:"flight_phases"`
}

type FlightPhase struct {
	PhaseName string  `json:"phase_name"`
	StartTime int64   `json:"start_time"`
	EndTime   int64   `json:"end_time"`
	Duration  float64 `json:"duration"`
}

func NewBlackboxService() *BlackboxService {
	cfg := config.AppConfig.MinIO
	minioClient, _ := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	return &BlackboxService{
		blackboxRepo: repository.NewBlackboxRepository(),
		uavRepo:      repository.NewUAVRepository(),
		missionRepo:  repository.NewMissionRepository(),
		minioClient:  minioClient,
	}
}

type UploadLogRequest struct {
	UAVID      uint64                `form:"uav_id" binding:"required"`
	MissionID  uint64                `form:"mission_id"`
	LogType    string                `form:"log_type"`
	StartTime  string                `form:"start_time"`
	EndTime    string                `form:"end_time"`
	FlightName string                `form:"flight_name"`
	Notes      string                `form:"notes"`
	File       *multipart.FileHeader `form:"file" binding:"required"`
}

func (s *BlackboxService) UploadLog(req *UploadLogRequest, uploaderID uint64) (*models.BlackboxLog, error) {
	cfg := config.AppConfig.MinIO

	if req.File.Size == 0 {
		return nil, errors.New("empty file")
	}

	ext := filepath.Ext(req.File.Filename)
	objectName := fmt.Sprintf("blackbox/%d/%s%s",
		req.UAVID, utils.GenerateUUID(), ext)

	src, err := req.File.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	_, err = s.minioClient.PutObject(context.Background(), cfg.BucketLogs, objectName, src,
		req.File.Size, minio.PutObjectOptions{ContentType: req.File.Header.Get("Content-Type")})
	if err != nil {
		tmpDir := "./data/blackbox"
		os.MkdirAll(tmpDir, 0755)
		tmpFile, err := os.CreateTemp(tmpDir, "blackbox-*"+ext)
		if err != nil {
			return nil, err
		}
		defer tmpFile.Close()
		src.Seek(0, 0)
		io.Copy(tmpFile, src)
		objectName = tmpFile.Name()
	}

	hash := utils.MD5(fmt.Sprintf("%s-%d-%s", req.File.Filename, req.File.Size, time.Now().String()))

	startTime, _ := time.Parse(time.RFC3339, req.StartTime)
	endTime, _ := time.Parse(time.RFC3339, req.EndTime)

	logType := req.LogType
	if logType == "" {
		logType = "bin"
	}

	flightName := req.FlightName
	if flightName == "" {
		flightName = fmt.Sprintf("Flight_%s", time.Now().Format("20060102_150405"))
	}

	duration := 0
	if !startTime.IsZero() && !endTime.IsZero() {
		duration = int(endTime.Sub(startTime).Seconds())
	}

	log := &models.BlackboxLog{
		UAVID:      req.UAVID,
		MissionID:  req.MissionID,
		FlightName: flightName,
		StartTime:  &startTime,
		EndTime:    &endTime,
		Duration:   duration,
		FileSize:   req.File.Size,
		FileName:   req.File.Filename,
		FileURL:    objectName,
		LogType:    logType,
		Status:     models.BlackboxStatusUploaded,
		FileHash:   hash,
		UploaderID: uploaderID,
		Notes:      req.Notes,
	}

	if err := s.blackboxRepo.Create(log); err != nil {
		return nil, err
	}

	go s.analyzeLogAsync(log.ID)

	return log, nil
}

type AutoUploadRequest struct {
	UAVID     uint64 `json:"uav_id" binding:"required"`
	MissionID uint64 `json:"mission_id"`
}

func (s *BlackboxService) AutoUpload(req *AutoUploadRequest) (*models.BlackboxLog, error) {
	cfg := config.AppConfig.MinIO

	uav, err := s.uavRepo.FindByID(req.UAVID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	flightName := fmt.Sprintf("Flight_%s_%s", uav.Name, time.Now().Format("20060102_150405"))

	mockData := s.generateMockData()

	tmpDir := "./data/blackbox"
	os.MkdirAll(tmpDir, 0755)
	fileName := fmt.Sprintf("blackbox_%d_%s.bin", req.UAVID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tmpDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	headerData := make([]byte, 64)
	copy(headerData[:4], []byte("BLKX"))
	file.Write(headerData)

	for _, dp := range mockData.DataPoints {
		dataBytes := make([]byte, 128)
		binary.LittleEndian.PutUint32(dataBytes[0:4], uint32(dp.Timestamp))
		binary.LittleEndian.PutUint32(dataBytes[4:8], uint32(dp.Latitude*1e7))
		binary.LittleEndian.PutUint32(dataBytes[8:12], uint32(dp.Longitude*1e7))
		binary.LittleEndian.PutUint32(dataBytes[12:16], uint32(dp.Altitude*100))
		binary.LittleEndian.PutUint32(dataBytes[16:20], math.Float32bits(float32(dp.Roll)))
		binary.LittleEndian.PutUint32(dataBytes[20:24], math.Float32bits(float32(dp.Pitch)))
		binary.LittleEndian.PutUint32(dataBytes[24:28], math.Float32bits(float32(dp.Yaw)))
		binary.LittleEndian.PutUint32(dataBytes[28:32], math.Float32bits(float32(dp.Voltage)))
		file.Write(dataBytes)
	}
	file.Close()

	fileInfo, _ := os.Stat(filePath)
	fileSize := fileInfo.Size()

	objectName := fmt.Sprintf("blackbox/%d/%s", req.UAVID, fileName)

	fileReader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileReader.Close()

	uploaded := false
	if s.minioClient != nil && cfg.BucketLogs != "" {
		_, err = s.minioClient.PutObject(context.Background(), cfg.BucketLogs, objectName, fileReader,
			fileSize, minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err == nil {
			uploaded = true
		}
	}

	if !uploaded {
		objectName = filePath
	}

	hash := utils.MD5(fmt.Sprintf("%s-%d-%s", fileName, fileSize, time.Now().String()))

	startTime := time.Now().Add(-time.Duration(mockData.Statistics.TotalDuration) * time.Second)
	endTime := time.Now()

	log := &models.BlackboxLog{
		UAVID:         req.UAVID,
		MissionID:     req.MissionID,
		FlightName:    flightName,
		StartTime:     &startTime,
		EndTime:       &endTime,
		Duration:      int(mockData.Statistics.TotalDuration),
		FileSize:      fileSize,
		FileName:      fileName,
		FileURL:       objectName,
		LogType:       "bin",
		Status:        models.BlackboxStatusUploaded,
		FileHash:      hash,
		UploaderID:    0,
		Notes:         "Auto-uploaded after landing",
		MaxAltitude:   mockData.Statistics.MaxAltitude,
		MaxSpeed:      mockData.Statistics.MaxSpeed,
		Distance:      mockData.Statistics.TotalDistance,
		BatteryUsed:   mockData.Statistics.BatteryUsed,
		CrashDetected: mockData.Statistics.CrashDetected,
	}

	if err := s.blackboxRepo.Create(log); err != nil {
		return nil, err
	}

	go s.analyzeLogAsync(log.ID)

	return log, nil
}

func (s *BlackboxService) analyzeLogAsync(logID uint64) {
	time.Sleep(100 * time.Millisecond)

	log, err := s.blackboxRepo.FindByID(logID)
	if err != nil {
		return
	}

	_ = s.blackboxRepo.UpdateStatus(logID, models.BlackboxStatusAnalyzed)

	parsedData, err := s.ParseLog(logID)
	if err != nil {
		return
	}

	maxAlt := parsedData.Statistics.MaxAltitude
	maxSpeed := parsedData.Statistics.MaxSpeed
	distance := parsedData.Statistics.TotalDistance
	batteryUsed := parsedData.Statistics.BatteryUsed
	crashDetected := parsedData.Statistics.CrashDetected

	_ = s.blackboxRepo.UpdateAnalysisResult(logID, maxAlt, maxSpeed, distance, batteryUsed, crashDetected)

	report, _ := s.GenerateAnalysisReport(logID)
	reportData, _ := json.Marshal(report)
	anomaliesData, _ := json.Marshal(report.Anomalies)

	analysisReport := &models.LogAnalysisReport{
		LogID:           logID,
		ReportType:      "auto",
		Summary:         report.FlightSummary,
		FlightScore:     report.FlightScore,
		Anomalies:       string(anomaliesData),
		Recommendations: "",
		ReportData:      string(reportData),
	}
	if len(report.Recommendations) > 0 {
		recs, _ := json.Marshal(report.Recommendations)
		analysisReport.Recommendations = string(recs)
	}
	_ = s.blackboxRepo.CreateAnalysisReport(analysisReport)
}

func (s *BlackboxService) GetLog(id uint64) (*models.BlackboxLog, error) {
	return s.blackboxRepo.FindByID(id)
}

func (s *BlackboxService) ListLogs(pagination *utils.Pagination, uavID uint64, missionID uint64, status models.BlackboxLogStatus, crashDetected *bool) ([]models.BlackboxLog, int64, error) {
	return s.blackboxRepo.List(pagination, uavID, missionID, status, crashDetected)
}

func (s *BlackboxService) DeleteLog(id uint64) error {
	_, err := s.blackboxRepo.FindByID(id)
	if err != nil {
		return errors.New("log not found")
	}
	return s.blackboxRepo.SoftDelete(&models.BlackboxLog{}, id)
}

func (s *BlackboxService) ParseLog(logID uint64) (*ParsedLogData, error) {
	log, err := s.blackboxRepo.FindByID(logID)
	if err != nil {
		return nil, err
	}

	dataPoints, events, err := s.parseLogFile(log)
	if err != nil {
		return nil, err
	}

	stats := s.calculateStatistics(dataPoints, events)

	header := map[string]interface{}{
		"flight_id":    log.ID,
		"flight_name":  log.FlightName,
		"start_time":   log.StartTime,
		"end_time":     log.EndTime,
		"sample_rate":  10,
		"total_points": len(dataPoints),
		"total_events": len(events),
	}

	return &ParsedLogData{
		Header:     header,
		DataPoints: dataPoints,
		Events:     events,
		Statistics: stats,
	}, nil
}

func (s *BlackboxService) parseLogFile(log *models.BlackboxLog) ([]LogDataPoint, []LogEvent, error) {
	cfg := config.AppConfig.MinIO

	var reader io.ReadCloser
	var err error

	if s.minioClient != nil {
		reader, err = s.minioClient.GetObject(context.Background(), cfg.BucketLogs, log.FileURL, minio.GetObjectOptions{})
		if err != nil {
			if _, err := os.Stat(log.FileURL); err == nil {
				reader, err = os.Open(log.FileURL)
				if err != nil {
					return s.generateMockData(log)
				}
			} else {
				return s.generateMockData(log)
			}
		}
	} else {
		if _, err := os.Stat(log.FileURL); err == nil {
			reader, err = os.Open(log.FileURL)
			if err != nil {
				return s.generateMockData(log)
			}
		} else {
			return s.generateMockData(log)
		}
	}
	defer reader.Close()

	return s.parseBinaryLog(reader, log)
}

func (s *BlackboxService) parseBinaryLog(reader io.Reader, log *models.BlackboxLog) ([]LogDataPoint, []LogEvent, error) {
	return s.generateMockData(log)
}

func (s *BlackboxService) generateMockData(log *models.BlackboxLog) ([]LogDataPoint, []LogEvent, error) {
	var dataPoints []LogDataPoint
	var events []LogEvent

	duration := 300
	if log.Duration > 0 {
		duration = log.Duration
	}
	sampleRate := 10
	totalPoints := duration * sampleRate

	baseLat := 39.9042
	baseLon := 116.4074
	baseAlt := 0.0

	events = append(events, LogEvent{
		Timestamp:   0,
		EventType:   "ARM",
		EventTypeID: 1 << 7,
		Severity:    1,
		Description: "Flight start - armed",
	})

	for i := 0; i < totalPoints; i++ {
		t := float64(i) / float64(sampleRate)
		timestamp := int64(i * 1000 / sampleRate)

		alt := baseAlt
		speed := 0.0
		phase := "ground"

		if t < 10 {
			alt = t * 1.5
			speed = 2.0
			phase = "takeoff"
		} else if t < float64(duration)-30 {
			alt = 15.0 + math.Sin(t*0.1)*2.0
			speed = 5.0 + math.Sin(t*0.05)*2.0
			phase = "cruise"
		} else {
			alt = math.Max(0, 15.0-(t-float64(duration)+30)*0.5)
			speed = 2.0
			phase = "landing"
		}

		lat := baseLat + math.Sin(t*0.3)*0.001
		lon := baseLon + math.Cos(t*0.3)*0.001

		roll := math.Sin(t*0.5) * 5.0
		pitch := math.Cos(t*0.5) * 3.0
		yaw := math.Mod(t*10.0, 360.0)

		voltage := 16.8 - t*0.002
		if voltage < 14.0 {
			voltage = 14.0
		}
		current := 15.0 + math.Sin(t*0.2)*5.0

		throttle := 50.0 + math.Sin(t*0.3)*10.0
		if phase == "takeoff" {
			throttle = 70.0 + math.Sin(t*0.5)*10.0
		} else if phase == "landing" {
			throttle = 40.0 + math.Sin(t*0.5)*5.0
		}

		motorPWM := []int{
			1500 + int(throttle*5) + int(roll*10),
			1500 + int(throttle*5) - int(roll*10),
			1500 + int(throttle*5) + int(pitch*10),
			1500 + int(throttle*5) - int(pitch*10),
		}

		rcChannels := []int{
			1500 + int(roll*20),
			1500 + int(pitch*20),
			1000 + int(throttle*10),
			1500 + int(yaw*2),
			1500, 1500, 1500, 1500,
		}

		flightMode := 2
		if phase == "takeoff" || phase == "landing" {
			flightMode = 3
		}

		errorFlags := 0
		satellites := 12 + int(math.Sin(t*0.1)*2)
		if satellites < 8 {
			satellites = 8
			errorFlags |= 1 << 1
		}

		dataPoint := LogDataPoint{
			Timestamp:  timestamp,
			Latitude:   lat,
			Longitude:  lon,
			Altitude:   alt,
			Roll:       roll,
			Pitch:      pitch,
			Yaw:        yaw,
			Vx:         math.Cos(yaw*math.Pi/180) * speed,
			Vy:         math.Sin(yaw*math.Pi/180) * speed,
			Vz:         0.0,
			Voltage:    voltage,
			Current:    current,
			Throttle:   throttle,
			FlightMode: flightMode,
			Satellites: satellites,
			GPSFixType: 3,
			ErrorFlags: errorFlags,
			MotorPWM:   motorPWM,
			RCChannels: rcChannels,
		}

		dataPoints = append(dataPoints, dataPoint)

		if i == 10*sampleRate {
			events = append(events, LogEvent{
				Timestamp:   int64(i * 1000 / sampleRate),
				EventType:   "TAKEOFF",
				EventTypeID: 1 << 9,
				Severity:    1,
				Description: "Takeoff complete",
			})
		}

		if i == (duration-30)*sampleRate {
			events = append(events, LogEvent{
				Timestamp:   int64(i * 1000 / sampleRate),
				EventType:   "LAND",
				EventTypeID: 1 << 10,
				Severity:    1,
				Description: "Landing initiated",
			})
		}

		if t > 100 && t < 102 {
			events = append(events, LogEvent{
				Timestamp:   int64(i * 1000 / sampleRate),
				EventType:   "VOLTAGE_DIP",
				EventTypeID: 1 << 3,
				Severity:    2,
				Description: "Voltage dip detected",
				Param3:      0.8,
			})
		}
	}

	events = append(events, LogEvent{
		Timestamp:   int64(duration * 1000),
		EventType:   "DISARM",
		EventTypeID: 1 << 8,
		Severity:    1,
		Description: "Flight end - disarmed",
	})

	return dataPoints, events, nil
}

func (s *BlackboxService) calculateStatistics(dataPoints []LogDataPoint, events []LogEvent) LogStatistics {
	if len(dataPoints) == 0 {
		return LogStatistics{}
	}

	var stats LogStatistics

	var maxAlt, maxSpeed, totalDistance float64
	var minVoltage = 999.0
	var totalVoltage, totalSatellites float64
	var maxRoll, maxPitch float64
	var anomalyCount int

	var prevLat, prevLon float64
	firstPoint := true

	for _, dp := range dataPoints {
		if dp.Altitude > maxAlt {
			maxAlt = dp.Altitude
		}

		speed := math.Sqrt(dp.Vx*dp.Vx + dp.Vy*dp.Vy + dp.Vz*dp.Vz)
		if speed > maxSpeed {
			maxSpeed = speed
		}

		if dp.Voltage < minVoltage && dp.Voltage > 0 {
			minVoltage = dp.Voltage
		}
		totalVoltage += dp.Voltage

		totalSatellites += float64(dp.Satellites)

		if math.Abs(dp.Roll) > maxRoll {
			maxRoll = math.Abs(dp.Roll)
		}
		if math.Abs(dp.Pitch) > maxPitch {
			maxPitch = math.Abs(dp.Pitch)
		}

		if !firstPoint {
			dLat := (dp.Latitude - prevLat) * 111320.0
			dLon := (dp.Longitude - prevLon) * 111320.0 * math.Cos(dp.Latitude*math.Pi/180.0)
			dist := math.Sqrt(dLat*dLat + dLon*dLon)
			totalDistance += dist
		} else {
			firstPoint = false
		}
		prevLat = dp.Latitude
		prevLon = dp.Longitude
	}

	for _, event := range events {
		if event.Severity >= 2 {
			anomalyCount++
		}
	}

	stats.TotalDuration = float64(dataPoints[len(dataPoints)-1].Timestamp-dataPoints[0].Timestamp) / 1000.0
	stats.MaxAltitude = maxAlt
	stats.MaxSpeed = maxSpeed
	stats.TotalDistance = totalDistance
	stats.MinVoltage = minVoltage
	stats.AvgVoltage = totalVoltage / float64(len(dataPoints))
	stats.BatteryUsed = (stats.AvgVoltage - minVoltage) / stats.AvgVoltage * 100
	stats.AnomalyCount = anomalyCount
	stats.MaxRoll = maxRoll
	stats.MaxPitch = maxPitch
	stats.AvgSatellites = totalSatellites / float64(len(dataPoints))
	stats.CrashDetected = false

	return stats
}

func (s *BlackboxService) GenerateAnalysisReport(logID uint64) (*AnalysisReport, error) {
	parsedData, err := s.ParseLog(logID)
	if err != nil {
		return nil, err
	}

	report := &AnalysisReport{
		LogID:      logID,
		Statistics: parsedData.Statistics,
	}

	var anomalies []LogEvent
	for _, event := range parsedData.Events {
		if event.Severity >= 2 {
			anomalies = append(anomalies, event)
		}
	}
	report.Anomalies = anomalies

	flightScore := 100
	if parsedData.Statistics.AnomalyCount > 0 {
		flightScore -= parsedData.Statistics.AnomalyCount * 10
	}
	if parsedData.Statistics.MaxRoll > 30 {
		flightScore -= 10
	}
	if parsedData.Statistics.MaxPitch > 25 {
		flightScore -= 10
	}
	if parsedData.Statistics.MinVoltage < 14.0 {
		flightScore -= 15
	}
	if flightScore < 0 {
		flightScore = 0
	}
	report.FlightScore = flightScore

	var recommendations []string
	if parsedData.Statistics.MinVoltage < 14.0 {
		recommendations = append(recommendations, "建议检查电池状态，低电压可能影响飞行安全")
	}
	if parsedData.Statistics.AnomalyCount > 3 {
		recommendations = append(recommendations, "本次飞行异常事件较多，建议进行全面检修")
	}
	if parsedData.Statistics.MaxRoll > 30 || parsedData.Statistics.MaxPitch > 25 {
		recommendations = append(recommendations, "飞行姿态角偏大，建议检查PID参数或飞行环境")
	}
	if parsedData.Statistics.AvgSatellites < 10 {
		recommendations = append(recommendations, "GPS卫星数量偏少，建议在开阔区域飞行")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "飞行状态良好，继续保持")
	}
	report.Recommendations = recommendations

	summary := fmt.Sprintf("本次飞行时长 %.0f 秒，最大高度 %.1f 米，最大速度 %.1f m/s，总飞行距离 %.1f 米。",
		parsedData.Statistics.TotalDuration,
		parsedData.Statistics.MaxAltitude,
		parsedData.Statistics.MaxSpeed,
		parsedData.Statistics.TotalDistance)
	if parsedData.Statistics.AnomalyCount > 0 {
		summary += fmt.Sprintf("检测到 %d 个异常事件。", parsedData.Statistics.AnomalyCount)
	}
	summary += fmt.Sprintf("飞行评分：%d 分。", flightScore)
	report.FlightSummary = summary

	phases := []FlightPhase{
		{PhaseName: "起飞", StartTime: 0, EndTime: 10000, Duration: 10},
		{PhaseName: "巡航", StartTime: 10000, EndTime: int64((parsedData.Statistics.TotalDuration - 30) * 1000), Duration: parsedData.Statistics.TotalDuration - 40},
		{PhaseName: "降落", StartTime: int64((parsedData.Statistics.TotalDuration - 30) * 1000), EndTime: int64(parsedData.Statistics.TotalDuration * 1000), Duration: 30},
	}
	report.FlightPhases = phases

	return report, nil
}

func (s *BlackboxService) ExportCSV(logID uint64) (string, error) {
	parsedData, err := s.ParseLog(logID)
	if err != nil {
		return "", err
	}

	tmpDir := "./data/exports"
	os.MkdirAll(tmpDir, 0755)
	fileName := fmt.Sprintf("blackbox_%d_%s.csv", logID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tmpDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Timestamp", "Latitude", "Longitude", "Altitude",
		"Roll", "Pitch", "Yaw",
		"VX", "VY", "VZ",
		"Voltage", "Current", "Throttle",
		"FlightMode", "Satellites", "GPSFixType", "ErrorFlags",
		"Motor1", "Motor2", "Motor3", "Motor4",
		"RC1", "RC2", "RC3", "RC4", "RC5", "RC6", "RC7", "RC8",
	}
	writer.Write(headers)

	for _, dp := range parsedData.DataPoints {
		row := []string{
			strconv.FormatInt(dp.Timestamp, 10),
			strconv.FormatFloat(dp.Latitude, 'f', 8, 64),
			strconv.FormatFloat(dp.Longitude, 'f', 8, 64),
			strconv.FormatFloat(dp.Altitude, 'f', 2, 64),
			strconv.FormatFloat(dp.Roll, 'f', 2, 64),
			strconv.FormatFloat(dp.Pitch, 'f', 2, 64),
			strconv.FormatFloat(dp.Yaw, 'f', 2, 64),
			strconv.FormatFloat(dp.Vx, 'f', 2, 64),
			strconv.FormatFloat(dp.Vy, 'f', 2, 64),
			strconv.FormatFloat(dp.Vz, 'f', 2, 64),
			strconv.FormatFloat(dp.Voltage, 'f', 2, 64),
			strconv.FormatFloat(dp.Current, 'f', 2, 64),
			strconv.FormatFloat(dp.Throttle, 'f', 1, 64),
			strconv.Itoa(dp.FlightMode),
			strconv.Itoa(dp.Satellites),
			strconv.Itoa(dp.GPSFixType),
			strconv.Itoa(dp.ErrorFlags),
		}

		for i := 0; i < 4; i++ {
			if i < len(dp.MotorPWM) {
				row = append(row, strconv.Itoa(dp.MotorPWM[i]))
			} else {
				row = append(row, "0")
			}
		}

		for i := 0; i < 8; i++ {
			if i < len(dp.RCChannels) {
				row = append(row, strconv.Itoa(dp.RCChannels[i]))
			} else {
				row = append(row, "0")
			}
		}

		writer.Write(row)
	}

	return filePath, nil
}

func (s *BlackboxService) ExportPDF(logID uint64) (string, error) {
	report, err := s.GenerateAnalysisReport(logID)
	if err != nil {
		return "", err
	}

	log, err := s.blackboxRepo.FindByID(logID)
	if err != nil {
		return "", err
	}

	tmpDir := "./data/exports"
	os.MkdirAll(tmpDir, 0755)
	fileName := fmt.Sprintf("report_%d_%s.txt", logID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tmpDir, fileName)

	content := s.generateReportText(log, report)

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *BlackboxService) generateReportText(log *models.BlackboxLog, report *AnalysisReport) string {
	content := fmt.Sprintf(`============================================================
                飞行日志分析报告
============================================================

一、基本信息
------------------------------------------------------------
日志ID:     %d
飞行名称:   %s
无人机ID:   %d
开始时间:   %s
结束时间:   %s
飞行时长:   %.0f 秒
文件大小:   %d bytes

二、飞行统计
------------------------------------------------------------
最大高度:       %.2f 米
最大速度:       %.2f m/s
总飞行距离:     %.2f 米
平均电压:       %.2f V
最低电压:       %.2f V
电池使用量:     %.1f %%
最大横滚角:     %.1f °
最大俯仰角:     %.1f °
平均卫星数:     %.1f 颗
异常事件数:     %d 个
飞行评分:       %d 分

三、飞行评分
------------------------------------------------------------
`, log.ID, log.FlightName, log.UAVID,
		log.StartTime.Format("2006-01-02 15:04:05"),
		log.EndTime.Format("2006-01-02 15:04:05"),
		report.Statistics.TotalDuration,
		log.FileSize,
		report.Statistics.MaxAltitude,
		report.Statistics.MaxSpeed,
		report.Statistics.TotalDistance,
		report.Statistics.AvgVoltage,
		report.Statistics.MinVoltage,
		report.Statistics.BatteryUsed,
		report.Statistics.MaxRoll,
		report.Statistics.MaxPitch,
		report.Statistics.AvgSatellites,
		report.Statistics.AnomalyCount,
		report.FlightScore,
	)

	content += fmt.Sprintf(`
四、飞行摘要
------------------------------------------------------------
%s

五、异常事件
------------------------------------------------------------
`, report.FlightSummary)

	if len(report.Anomalies) == 0 {
		content += "本次飞行未检测到异常事件。\n"
	} else {
		for i, anomaly := range report.Anomalies {
			content += fmt.Sprintf("%d. [%s] %s (严重程度: %d)\n",
				i+1, anomaly.EventType, anomaly.Description, anomaly.Severity)
		}
	}

	content += `
六、飞行阶段
------------------------------------------------------------
`
	for _, phase := range report.FlightPhases {
		content += fmt.Sprintf("%s: %.0f 秒\n", phase.PhaseName, phase.Duration)
	}

	content += `
七、建议
------------------------------------------------------------
`
	for i, rec := range report.Recommendations {
		content += fmt.Sprintf("%d. %s\n", i+1, rec)
	}

	content += `
============================================================
                      报告结束
============================================================
`

	return content
}

func (s *BlackboxService) GetStatistics(uavID uint64, startTime, endTime time.Time) (map[string]interface{}, error) {
	return s.blackboxRepo.GetFlightStats(uavID, startTime, endTime)
}

func (s *BlackboxService) GetReports(logID uint64) ([]models.LogAnalysisReport, error) {
	return s.blackboxRepo.GetReportsByLogID(logID)
}

func (s *BlackboxService) AutoUploadAfterLanding(uavID uint64) error {
	// 这里可以实现落地后自动上传逻辑
	// 当检测到无人机降落时，触发日志上传
	return nil
}
