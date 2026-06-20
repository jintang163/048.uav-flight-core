package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"groundstation-backend/internal/config"
	"groundstation-backend/internal/models"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type YOLODetectionResult struct {
	Class      int     `json:"class"`
	ClassName  string  `json:"class_name"`
	Confidence float64 `json:"confidence"`
	X1         float64 `json:"x1"`
	Y1         float64 `json:"y1"`
	X2         float64 `json:"x2"`
	Y2         float64 `json:"y2"`
}

type YOLOv8Service struct {
	apiBase    string
	apiKey     string
	httpClient *http.Client
	enabled    bool
}

func NewYOLOv8Service() *YOLOv8Service {
	cfg := config.AppConfig
	enabled := cfg.YOLOv8API != ""
	return &YOLOv8Service{
		apiBase: cfg.YOLOv8API,
		apiKey:  cfg.YOLOv8APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled: enabled,
	}
}

var classMapping = map[int]models.DetectionClass{
	0:  models.DetectionClassPerson,
	1:  models.DetectionClassBicycle,
	2:  models.DetectionClassCar,
	3:  models.DetectionClassMotorcycle,
	5:  models.DetectionClassBus,
	7:  models.DetectionClassTruck,
	16: models.DetectionClassDog,
}

var classNames = map[models.DetectionClass]string{
	models.DetectionClassPerson:     "人员",
	models.DetectionClassCar:        "轿车",
	models.DetectionClassTruck:      "卡车",
	models.DetectionClassBus:        "公交车",
	models.DetectionClassMotorcycle: "摩托车",
	models.DetectionClassBicycle:    "自行车",
	models.DetectionClassDog:        "狗",
	models.DetectionClassUnknown:    "未知",
}

func GetClassDisplayName(cls models.DetectionClass) string {
	if name, ok := classNames[cls]; ok {
		return name
	}
	return string(cls)
}

func (s *YOLOv8Service) DetectFromImage(imageData []byte, frameWidth, frameHeight int) ([]YOLODetectionResult, error) {
	if !s.enabled {
		return s.generateMockDetections(frameWidth, frameHeight), nil
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "frame.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, fmt.Errorf("write image data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", s.apiBase+"/detect", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if s.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("yolo api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("yolo api error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Detections []YOLODetectionResult `json:"detections"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode yolo response: %w", err)
	}

	return apiResp.Detections, nil
}

func (s *YOLOv8Service) ConvertToTargets(uavID uint64, detections []YOLODetectionResult, frameWidth, frameHeight int, lat, lng, alt float64) []*models.DetectionTarget {
	targets := make([]*models.DetectionTarget, 0, len(detections))
	now := time.Now()

	for _, d := range detections {
		cls, ok := classMapping[d.Class]
		if !ok {
			cls = models.DetectionClassUnknown
		}

		width := d.X2 - d.X1
		height := d.Y2 - d.Y1

		targets = append(targets, &models.DetectionTarget{
			UAVID:       uavID,
			Class:       cls,
			ClassName:   GetClassDisplayName(cls),
			Confidence:  d.Confidence,
			BboxX:       d.X1,
			BboxY:       d.Y1,
			BboxWidth:   width,
			BboxHeight:  height,
			FrameWidth:  frameWidth,
			FrameHeight: frameHeight,
			Latitude:    lat,
			Longitude:   lng,
			Altitude:    alt,
			CreatedAt:   now,
		})
	}

	return targets
}

func (s *YOLOv8Service) generateMockDetections(frameWidth, frameHeight int) []YOLODetectionResult {
	if frameWidth <= 0 {
		frameWidth = 1280
	}
	if frameHeight <= 0 {
		frameHeight = 720
	}

	return []YOLODetectionResult{
		{
			Class:      0,
			ClassName:  "person",
			Confidence: 0.92,
			X1:         float64(frameWidth) * 0.3,
			Y1:         float64(frameHeight) * 0.35,
			X2:         float64(frameWidth) * 0.38,
			Y2:         float64(frameHeight) * 0.65,
		},
		{
			Class:      2,
			ClassName:  "car",
			Confidence: 0.87,
			X1:         float64(frameWidth) * 0.55,
			Y1:         float64(frameHeight) * 0.5,
			X2:         float64(frameWidth) * 0.72,
			Y2:         float64(frameHeight) * 0.72,
		},
	}
}
