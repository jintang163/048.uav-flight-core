package service

import (
	"bytes"
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
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/signintech/gopdf"
)

const (
	BlackboxHeaderMagic   = 0x424C4B58
	BlackboxLogTypeData   = 0x01
	BlackboxLogTypeEvent  = 0x02
	BlackboxHeaderSize    = 37
	BlackboxEntryHeadSize = 3
)

var eventTypeMap = map[uint32]string{
	1 << 0:  "LOW_BATTERY",
	1 << 1:  "GPS_LOSS",
	1 << 2:  "RC_LOSS",
	1 << 3:  "VOLTAGE_DIP",
	1 << 4:  "MOTOR_FAILURE",
	1 << 5:  "CRASH",
	1 << 6:  "FENCE_BREACH",
	1 << 7:  "ARM",
	1 << 8:  "DISARM",
	1 << 9:  "TAKEOFF",
	1 << 10: "LAND",
	1 << 11: "MODE_CHANGE",
	1 << 12: "FAILSAFE",
}

type BlackboxHeader struct {
	Magic        uint32
	Version      uint8
	FlightID     uint32
	StartTime    uint32
	EndTime      uint32
	TotalEntries uint32
	DataSize     uint32
	SampleRate   uint16
	Reserved     [10]uint8
}

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

func (s *BlackboxService) AutoUpload(uavID uint64, missionID uint64, file *multipart.FileHeader) (*models.BlackboxLog, error) {
	cfg := config.AppConfig.MinIO

	if file == nil || file.Size == 0 {
		return nil, errors.New("empty or missing file")
	}

	uav, err := s.uavRepo.FindByID(uavID)
	if err != nil {
		return nil, errors.New("uav not found")
	}

	flightName := fmt.Sprintf("Flight_%s_%s", uav.Name, time.Now().Format("20060102_150405"))

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("read file failed: %w", err)
	}

	header, err := s.parseBlackboxHeaderFromBytes(fileData)
	if err != nil {
		return nil, fmt.Errorf("parse header failed: %w", err)
	}

	tmpDir := "./data/blackbox"
	os.MkdirAll(tmpDir, 0755)
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".bin"
	}
	fileName := fmt.Sprintf("blackbox_%d_%s%s", uavID, time.Now().Format("20060102_150405"), ext)
	filePath := filepath.Join(tmpDir, fileName)

	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return nil, fmt.Errorf("save file failed: %w", err)
	}

	fileSize := int64(len(fileData))
	objectName := fmt.Sprintf("blackbox/%d/%s", uavID, fileName)

	uploaded := false
	if s.minioClient != nil && cfg.BucketLogs != "" {
		fileReader := bytes.NewReader(fileData)
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

	dataPoints, events, parseErr := s.parseBinaryLog(bytes.NewReader(fileData), nil)
	stats := LogStatistics{}
	if parseErr == nil && len(dataPoints) > 0 {
		stats = s.calculateStatistics(dataPoints, events)
	}

	startTime := time.Now()
	endTime := time.Now()
	duration := 0
	if header.StartTime > 0 && header.EndTime > 0 {
		startTime = time.Unix(int64(header.StartTime), 0)
		endTime = time.Unix(int64(header.EndTime), 0)
		duration = int(header.EndTime - header.StartTime)
	} else if stats.TotalDuration > 0 {
		startTime = time.Now().Add(-time.Duration(stats.TotalDuration) * time.Second)
		duration = int(stats.TotalDuration)
	}

	log := &models.BlackboxLog{
		UAVID:         uavID,
		MissionID:     missionID,
		FlightName:    flightName,
		StartTime:     &startTime,
		EndTime:       &endTime,
		Duration:      duration,
		FileSize:      fileSize,
		FileName:      fileName,
		FileURL:       objectName,
		LogType:       "bin",
		Status:        models.BlackboxStatusUploaded,
		FileHash:      hash,
		UploaderID:    0,
		Notes:         "Auto-uploaded after landing",
		MaxAltitude:   stats.MaxAltitude,
		MaxSpeed:      stats.MaxSpeed,
		Distance:      stats.TotalDistance,
		BatteryUsed:   stats.BatteryUsed,
		CrashDetected: stats.CrashDetected,
	}

	if err := s.blackboxRepo.Create(log); err != nil {
		return nil, err
	}

	go s.analyzeLogAsync(log.ID)

	return log, nil
}

func (s *BlackboxService) parseBlackboxHeaderFromBytes(data []byte) (*BlackboxHeader, error) {
	if len(data) < 64 {
		return nil, fmt.Errorf("file too small for header: %d bytes", len(data))
	}

	headerBuf := bytes.NewReader(data[0:BlackboxHeaderSize])
	var header BlackboxHeader
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Magic); err != nil {
		return nil, err
	}
	if header.Magic != BlackboxHeaderMagic {
		return nil, fmt.Errorf("invalid magic: 0x%X", header.Magic)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.FlightID); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.StartTime); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.EndTime); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.TotalEntries); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.DataSize); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.SampleRate); err != nil {
		return nil, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Reserved); err != nil {
		return nil, err
	}

	return &header, nil
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
			if _, statErr := os.Stat(log.FileURL); statErr == nil {
				reader, err = os.Open(log.FileURL)
				if err != nil {
					return nil, nil, fmt.Errorf("open local file failed: %w", err)
				}
			} else {
				return nil, nil, fmt.Errorf("file not found in minio or local: %w", err)
			}
		}
	} else {
		if _, statErr := os.Stat(log.FileURL); statErr == nil {
			reader, err = os.Open(log.FileURL)
			if err != nil {
				return nil, nil, fmt.Errorf("open local file failed: %w", err)
			}
		} else {
			return nil, nil, fmt.Errorf("file not found: %s", log.FileURL)
		}
	}
	defer reader.Close()

	return s.parseBinaryLog(reader, log)
}

func (s *BlackboxService) parseBinaryLog(reader io.Reader, log *models.BlackboxLog) ([]LogDataPoint, []LogEvent, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("read log data failed: %w", err)
	}

	if len(data) < BlackboxHeaderSize {
		return nil, nil, fmt.Errorf("log file too small: %d bytes", len(data))
	}

	headerBuf := bytes.NewReader(data[:BlackboxHeaderSize])
	var header BlackboxHeader
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Magic); err != nil {
		return nil, nil, fmt.Errorf("read magic failed: %w", err)
	}
	if header.Magic != BlackboxHeaderMagic {
		return nil, nil, fmt.Errorf("invalid magic number: 0x%X, expected 0x%X", header.Magic, BlackboxHeaderMagic)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Version); err != nil {
		return nil, nil, fmt.Errorf("read version failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.FlightID); err != nil {
		return nil, nil, fmt.Errorf("read flight_id failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.StartTime); err != nil {
		return nil, nil, fmt.Errorf("read start_time failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.EndTime); err != nil {
		return nil, nil, fmt.Errorf("read end_time failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.TotalEntries); err != nil {
		return nil, nil, fmt.Errorf("read total_entries failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.DataSize); err != nil {
		return nil, nil, fmt.Errorf("read data_size failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.SampleRate); err != nil {
		return nil, nil, fmt.Errorf("read sample_rate failed: %w", err)
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &header.Reserved); err != nil {
		return nil, nil, fmt.Errorf("read reserved failed: %w", err)
	}

	var dataPoints []LogDataPoint
	var events []LogEvent

	offset := BlackboxHeaderSize
	for offset+BlackboxEntryHeadSize <= len(data) {
		entryType := data[offset]
		entrySize := binary.LittleEndian.Uint16(data[offset+1 : offset+3])
		offset += BlackboxEntryHeadSize

		if offset+int(entrySize) > len(data) {
			break
		}

		entryData := data[offset : offset+int(entrySize)]

		switch entryType {
		case BlackboxLogTypeData:
			dp, err := s.parseDataEntry(entryData)
			if err == nil {
				dataPoints = append(dataPoints, dp)
			}
		case BlackboxLogTypeEvent:
			event, err := s.parseEventEntry(entryData)
			if err == nil {
				events = append(events, event)
			}
		}

		offset += int(entrySize)
	}

	return dataPoints, events, nil
}

func (s *BlackboxService) parseDataEntry(data []byte) (LogDataPoint, error) {
	var dp LogDataPoint
	buf := bytes.NewReader(data)

	var timestamp uint32
	if err := binary.Read(buf, binary.LittleEndian, &timestamp); err != nil {
		return dp, err
	}
	dp.Timestamp = int64(timestamp)

	var lat, lon, alt int32
	if err := binary.Read(buf, binary.LittleEndian, &lat); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &lon); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &alt); err != nil {
		return dp, err
	}
	dp.Latitude = float64(lat) / 1e7
	dp.Longitude = float64(lon) / 1e7
	dp.Altitude = float64(alt) / 1000.0

	var roll, pitch, yaw, vx, vy, vz, voltage, current, throttle float32
	if err := binary.Read(buf, binary.LittleEndian, &roll); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &pitch); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &yaw); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &vx); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &vy); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &vz); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &voltage); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &current); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &throttle); err != nil {
		return dp, err
	}
	dp.Roll = float64(roll)
	dp.Pitch = float64(pitch)
	dp.Yaw = float64(yaw)
	dp.Vx = float64(vx)
	dp.Vy = float64(vy)
	dp.Vz = float64(vz)
	dp.Voltage = float64(voltage)
	dp.Current = float64(current)
	dp.Throttle = float64(throttle)

	rcChannels := make([]uint16, 8)
	for i := 0; i < 8; i++ {
		if err := binary.Read(buf, binary.LittleEndian, &rcChannels[i]); err != nil {
			return dp, err
		}
	}
	dp.RCChannels = make([]int, 8)
	for i := 0; i < 8; i++ {
		dp.RCChannels[i] = int(rcChannels[i])
	}

	motorPWM := make([]uint16, 4)
	for i := 0; i < 4; i++ {
		if err := binary.Read(buf, binary.LittleEndian, &motorPWM[i]); err != nil {
			return dp, err
		}
	}
	dp.MotorPWM = make([]int, 4)
	for i := 0; i < 4; i++ {
		dp.MotorPWM[i] = int(motorPWM[i])
	}

	var flightMode, satellites, gpsFixType, errorFlags uint8
	if err := binary.Read(buf, binary.LittleEndian, &flightMode); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &satellites); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &gpsFixType); err != nil {
		return dp, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &errorFlags); err != nil {
		return dp, err
	}
	dp.FlightMode = int(flightMode)
	dp.Satellites = int(satellites)
	dp.GPSFixType = int(gpsFixType)
	dp.ErrorFlags = int(errorFlags)

	return dp, nil
}

func (s *BlackboxService) parseEventEntry(data []byte) (LogEvent, error) {
	var event LogEvent
	buf := bytes.NewReader(data)

	var timestamp uint32
	if err := binary.Read(buf, binary.LittleEndian, &timestamp); err != nil {
		return event, err
	}
	event.Timestamp = int64(timestamp)

	var eventType uint32
	if err := binary.Read(buf, binary.LittleEndian, &eventType); err != nil {
		return event, err
	}
	event.EventTypeID = eventType
	if name, ok := eventTypeMap[eventType]; ok {
		event.EventType = name
	} else {
		event.EventType = fmt.Sprintf("UNKNOWN(0x%X)", eventType)
	}

	var param1, param2 int32
	if err := binary.Read(buf, binary.LittleEndian, &param1); err != nil {
		return event, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &param2); err != nil {
		return event, err
	}
	event.Param1 = param1
	event.Param2 = param2

	var param3, param4 float32
	if err := binary.Read(buf, binary.LittleEndian, &param3); err != nil {
		return event, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &param4); err != nil {
		return event, err
	}
	event.Param3 = param3
	event.Param4 = param4

	var severity uint8
	if err := binary.Read(buf, binary.LittleEndian, &severity); err != nil {
		return event, err
	}
	event.Severity = int(severity)

	descBytes := make([]byte, 64)
	if err := binary.Read(buf, binary.LittleEndian, &descBytes); err != nil {
		return event, err
	}
	event.Description = string(bytes.TrimRight(descBytes, "\x00"))

	return event, nil
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
	fileName := fmt.Sprintf("report_%d_%s.pdf", logID, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(tmpDir, fileName)

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	title := "Flight Log Analysis Report"
	pdf.SetFont("Helvetica", "B", 20)
	titleWidth, _ := pdf.MeasureTextWidth(title)
	pdf.SetX((gopdf.PageSizeA4.W - titleWidth) / 2)
	pdf.Cell(nil, title)
	pdf.Br(30)

	s.pdfAddSectionTitle(&pdf, "1. Basic Information")
	s.pdfAddBasicInfoTable(&pdf, log, report)
	pdf.Br(10)

	s.pdfAddSectionTitle(&pdf, "2. Flight Statistics")
	s.pdfAddStatisticsTable(&pdf, report)
	pdf.Br(10)

	s.pdfAddSectionTitle(&pdf, "3. Anomaly Events")
	s.pdfAddAnomalyList(&pdf, report)
	pdf.Br(10)

	s.pdfAddSectionTitle(&pdf, "4. Recommendations")
	s.pdfAddRecommendations(&pdf, report)
	pdf.Br(10)

	s.pdfAddSectionTitle(&pdf, "5. Flight Summary")
	pdf.SetFont("Helvetica", "", 11)
	summary := report.FlightSummary
	s.pdfDrawTextWrap(&pdf, summary, 40, gopdf.PageSizeA4.W-80)

	err = pdf.WritePdf(filePath)
	if err != nil {
		return "", fmt.Errorf("write pdf failed: %w", err)
	}

	return filePath, nil
}

func (s *BlackboxService) pdfAddSectionTitle(pdf **gopdf.GoPdf, title string) {
	(*pdf).SetFont("Helvetica", "B", 14)
	(*pdf).SetTextColor(0, 51, 102)
	(*pdf).Cell(nil, title)
	(*pdf).Br(12)
	(*pdf).SetTextColor(0, 0, 0)
}

func (s *BlackboxService) pdfAddBasicInfoTable(pdf **gopdf.GoPdf, log *models.BlackboxLog, report *AnalysisReport) {
	startTimeStr := "-"
	endTimeStr := "-"
	if log.StartTime != nil {
		startTimeStr = log.StartTime.Format("2006-01-02 15:04:05")
	}
	if log.EndTime != nil {
		endTimeStr = log.EndTime.Format("2006-01-02 15:04:05")
	}

	basicInfo := [][]string{
		{"Log ID", fmt.Sprintf("%d", log.ID)},
		{"Flight Name", log.FlightName},
		{"UAV ID", fmt.Sprintf("%d", log.UAVID)},
		{"Mission ID", fmt.Sprintf("%d", log.MissionID)},
		{"Start Time", startTimeStr},
		{"End Time", endTimeStr},
		{"Duration (s)", fmt.Sprintf("%.0f", report.Statistics.TotalDuration)},
		{"File Size (bytes)", fmt.Sprintf("%d", log.FileSize)},
		{"Flight Score", fmt.Sprintf("%d", report.FlightScore)},
	}

	s.pdfDrawTable(pdf, basicInfo, []float64{80, 200})
}

func (s *BlackboxService) pdfAddStatisticsTable(pdf **gopdf.GoPdf, report *AnalysisReport) {
	stats := [][]string{
		{"Max Altitude (m)", fmt.Sprintf("%.2f", report.Statistics.MaxAltitude)},
		{"Max Speed (m/s)", fmt.Sprintf("%.2f", report.Statistics.MaxSpeed)},
		{"Total Distance (m)", fmt.Sprintf("%.2f", report.Statistics.TotalDistance)},
		{"Avg Voltage (V)", fmt.Sprintf("%.2f", report.Statistics.AvgVoltage)},
		{"Min Voltage (V)", fmt.Sprintf("%.2f", report.Statistics.MinVoltage)},
		{"Battery Used (%)", fmt.Sprintf("%.1f", report.Statistics.BatteryUsed)},
		{"Max Roll (deg)", fmt.Sprintf("%.1f", report.Statistics.MaxRoll)},
		{"Max Pitch (deg)", fmt.Sprintf("%.1f", report.Statistics.MaxPitch)},
		{"Avg Satellites", fmt.Sprintf("%.1f", report.Statistics.AvgSatellites)},
		{"Anomaly Count", fmt.Sprintf("%d", report.Statistics.AnomalyCount)},
		{"Crash Detected", fmt.Sprintf("%t", report.Statistics.CrashDetected)},
	}

	s.pdfDrawTable(pdf, stats, []float64{120, 180})
}

func (s *BlackboxService) pdfAddAnomalyList(pdf **gopdf.GoPdf, report *AnalysisReport) {
	(*pdf).SetFont("Helvetica", "", 11)
	if len(report.Anomalies) == 0 {
		(*pdf).Cell(nil, "No anomaly events detected during this flight.")
		(*pdf).Br(8)
		return
	}

	for i, anomaly := range report.Anomalies {
		line := fmt.Sprintf("%d. [%s] %s (Severity: %d)", i+1, anomaly.EventType, anomaly.Description, anomaly.Severity)
		s.pdfDrawTextWrap(pdf, line, 40, gopdf.PageSizeA4.W-80)
	}
}

func (s *BlackboxService) pdfAddRecommendations(pdf **gopdf.GoPdf, report *AnalysisReport) {
	(*pdf).SetFont("Helvetica", "", 11)
	for i, rec := range report.Recommendations {
		line := fmt.Sprintf("%d. %s", i+1, rec)
		s.pdfDrawTextWrap(pdf, line, 40, gopdf.PageSizeA4.W-80)
	}
}

func (s *BlackboxService) pdfDrawTable(pdf **gopdf.GoPdf, rows [][]string, colWidths []float64) {
	(*pdf).SetFont("Helvetica", "", 11)
	startX := (*pdf).GetX()
	startY := (*pdf).GetY()
	cellHeight := 18.0

	for rowIdx, row := range rows {
		x := startX
		for colIdx, cell := range row {
			w := colWidths[colIdx]
			(*pdf).SetX(x)
			(*pdf).SetY(startY + float64(rowIdx)*cellHeight)

			(*pdf).Rect(x, startY+float64(rowIdx)*cellHeight, w, cellHeight, "D")

			if rowIdx == 0 {
				(*pdf).SetFont("Helvetica", "B", 11)
			} else {
				(*pdf).SetFont("Helvetica", "", 11)
			}

			(*pdf).SetX(x + 5)
			(*pdf).SetY(startY + float64(rowIdx)*cellHeight + 5)
			(*pdf).Cell(nil, cell)

			x += w
		}
	}

	(*pdf).SetY(startY + float64(len(rows))*cellHeight + 5)
	(*pdf).SetX(startX)
}

func (s *BlackboxService) pdfDrawTextWrap(pdf **gopdf.GoPdf, text string, x, maxWidth float64) {
	(*pdf).SetFont("Helvetica", "", 11)
	(*pdf).SetX(x)

	words := strings.Fields(text)
	if len(words) == 0 {
		(*pdf).Br(14)
		return
	}

	currentLine := ""
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		lineWidth, _ := (*pdf).MeasureTextWidth(testLine)
		if lineWidth > maxWidth && currentLine != "" {
			(*pdf).Cell(nil, currentLine)
			(*pdf).Br(14)
			(*pdf).SetX(x)
			currentLine = word
		} else {
			currentLine = testLine
		}
	}
	if currentLine != "" {
		(*pdf).Cell(nil, currentLine)
		(*pdf).Br(14)
	}
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
