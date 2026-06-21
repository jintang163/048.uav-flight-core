package tts

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type TTSEngine interface {
	Synthesize(ctx context.Context, text string, voice string, speed float64, pitch float64, volume int) ([]byte, string, error)
	GetAvailableVoices() []VoiceInfo
}

type VoiceInfo struct {
	Name     string
	Lang     string
	Gender   string
	FullName string
}

type TTSService struct {
	audioDir      string
	audioURLBase  string
	edgeTTSURL    string
	fallbackWAV   bool
	httpClient    *http.Client
}

var defaultTTSService *TTSService

type TTSServiceConfig struct {
	AudioDir     string
	AudioURLBase string
	EdgeTTSURL   string
}

func NewTTSService(config TTSServiceConfig) *TTSService {
	if config.AudioDir == "" {
		config.AudioDir = "./data/tts_audio"
	}
	if config.AudioURLBase == "" {
		config.AudioURLBase = "/api/v1/static/tts"
	}
	if config.EdgeTTSURL == "" {
		config.EdgeTTSURL = ""
	}

	_ = os.MkdirAll(config.AudioDir, 0755)

	svc := &TTSService{
		audioDir:     config.AudioDir,
		audioURLBase: config.AudioURLBase,
		edgeTTSURL:   config.EdgeTTSURL,
		fallbackWAV:  true,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	defaultTTSService = svc
	return svc
}

func GetTTSService() *TTSService {
	if defaultTTSService == nil {
		defaultTTSService = NewTTSService(TTSServiceConfig{})
	}
	return defaultTTSService
}

func (s *TTSService) GetAudioDir() string {
	return s.audioDir
}

func (s *TTSService) GetAudioURLBase() string {
	return s.audioURLBase
}

func (s *TTSService) GetAvailableVoices() []VoiceInfo {
	return []VoiceInfo{
		{Name: "zh-CN-XiaoxiaoNeural", Lang: "zh-CN", Gender: "female", FullName: "晓晓 (女声)"},
		{Name: "zh-CN-YunxiNeural", Lang: "zh-CN", Gender: "male", FullName: "云希 (男声)"},
		{Name: "zh-CN-YunjianNeural", Lang: "zh-CN", Gender: "male", FullName: "云健 (男声)"},
		{Name: "zh-CN-XiaoyiNeural", Lang: "zh-CN", Gender: "female", FullName: "晓伊 (女声)"},
		{Name: "en-US-AriaNeural", Lang: "en-US", Gender: "female", FullName: "Aria (Female)"},
		{Name: "en-US-GuyNeural", Lang: "en-US", Gender: "male", FullName: "Guy (Male)"},
	}
}

func (s *TTSService) Synthesize(ctx context.Context, text string, voice string, speed float64, pitch float64, volume int) ([]byte, string, string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, "", "", errors.New("text is empty")
	}

	if voice == "" {
		voice = "zh-CN-XiaoxiaoNeural"
	}
	if speed <= 0 {
		speed = 1.0
	}
	if pitch == 0 {
		pitch = 0
	}
	if volume <= 0 {
		volume = 80
	}

	var audioData []byte
	var err error
	var format = "wav"

	if s.edgeTTSURL != "" {
		audioData, format, err = s.synthesizeWithEdge(ctx, text, voice, speed, pitch, volume)
		if err == nil && len(audioData) > 0 {
			audioPath, audioURL, saveErr := s.saveAudioFile(audioData, format)
			if saveErr != nil {
				return nil, "", "", saveErr
			}
			return audioData, audioPath, audioURL, nil
		}
	}

	if s.fallbackWAV || s.edgeTTSURL == "" {
		audioData, err = s.synthesizeWithWAV(text, voice, speed, pitch, volume)
		if err != nil {
			return nil, "", "", err
		}
		audioPath, audioURL, saveErr := s.saveAudioFile(audioData, format)
		if saveErr != nil {
			return nil, "", "", saveErr
		}
		return audioData, audioPath, audioURL, nil
	}

	return nil, "", "", errors.New("all TTS engines failed")
}

func (s *TTSService) synthesizeWithEdge(ctx context.Context, text string, voice string, speed float64, pitch float64, volume int) ([]byte, string, error) {
	reqURL := fmt.Sprintf("%s/tts", s.edgeTTSURL)

	formData := url.Values{}
	formData.Set("text", text)
	formData.Set("voice", voice)
	formData.Set("rate", fmt.Sprintf("%+d%%", int((speed-1)*100)))
	formData.Set("pitch", fmt.Sprintf("%+dHz", int(pitch)))
	formData.Set("volume", fmt.Sprintf("%d", volume))

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("edge tts returned status %d", resp.StatusCode)
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	format := "mp3"
	if strings.Contains(contentType, "wav") {
		format = "wav"
	}

	return audioData, format, nil
}

func (s *TTSService) synthesizeWithWAV(text string, voice string, speed float64, pitch float64, volume int) ([]byte, error) {
	sampleRate := 22050
	bitsPerSample := 16
	numChannels := 1

	durationSec := s.estimateDuration(text, speed)
	totalSamples := int(float64(sampleRate) * durationSec)

	data := make([]int16, totalSamples)
	volMultiplier := float64(volume) / 100.0

	s.generateSpeechWaveform(data, text, voice, sampleRate, speed, pitch, volMultiplier)

	return s.encodeWAV(data, sampleRate, bitsPerSample, numChannels), nil
}

func (s *TTSService) estimateDuration(text string, speed float64) float64 {
	chineseChars := 0
	otherChars := 0

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			chineseChars++
		} else if !unicode.IsSpace(r) {
			otherChars++
		}
	}

	chineseDuration := float64(chineseChars) * 0.18 / speed
	englishDuration := float64(otherChars) * 0.06 / speed

	total := chineseDuration + englishDuration + 0.3
	if total < 0.5 {
		total = 0.5
	}
	return total
}

func (s *TTSService) generateSpeechWaveform(data []int16, text string, voice string, sampleRate int, speed float64, pitch float64, volume float64) {
	gender := "female"
	for _, v := range s.GetAvailableVoices() {
		if v.Name == voice {
			gender = v.Gender
			break
		}
	}

	baseFreq := 220.0
	if gender == "male" {
		baseFreq = 150.0
	}
	baseFreq *= math.Pow(2, pitch/1200.0)

	runes := []rune(text)
	samplesPerRune := float64(len(data)) / float64(len(runes))
	if samplesPerRune < 100 {
		samplesPerRune = 100
	}

	for i, r := range runes {
		startSample := int(float64(i) * samplesPerRune)
		endSample := int(float64(i+1) * samplesPerRune)
		if endSample > len(data) {
			endSample = len(data)
		}

		var freq float64
		isVowel := false
		isPunctuation := false

		switch {
		case unicode.Is(unicode.Han, r):
			freq = baseFreq + s.toneForChar(r)*15
			isVowel = true
		case strings.ContainsRune("aeiouAEIOU", r):
			freq = baseFreq * 1.2
			isVowel = true
		case strings.ContainsRune("bcdfghjklmnpqrstvwxyzBCDFGHJKLMNPQRSTVWXYZ", r):
			freq = baseFreq * 0.9
		case unicode.IsDigit(r):
			freq = baseFreq * 1.1
			isVowel = true
		case unicode.IsPunct(r) || unicode.IsSpace(r):
			isPunctuation = true
		default:
			freq = baseFreq
		}

		numSamples := endSample - startSample
		if numSamples <= 0 {
			continue
		}

		for j := 0; j < numSamples; j++ {
			pos := startSample + j
			if pos >= len(data) {
				break
			}

			t := float64(j) / float64(sampleRate)
			progress := float64(j) / float64(numSamples)

			amplitude := 1.0
			if isPunctuation {
				amplitude = 0.05
			} else {
				envelope := 1.0
				if progress < 0.1 {
					envelope = progress / 0.1
				} else if progress > 0.85 {
					envelope = (1.0 - progress) / 0.15
					if envelope < 0 {
						envelope = 0
					}
				}
				amplitude = envelope * volume
			}

			var sample float64
			if isVowel {
				harmonic1 := math.Sin(2 * math.Pi * freq * t)
				harmonic2 := 0.5 * math.Sin(2*math.Pi*freq*2*t)
				harmonic3 := 0.25 * math.Sin(2*math.Pi*freq*3*t)
				sample = amplitude * 0.5 * (harmonic1 + harmonic2 + harmonic3)
			} else if !isPunctuation {
				noise := (float64(int(math.Sin(float64(j)*12345.0)*10000)%1000) / 1000.0 - 0.5)
				sample = amplitude * 0.15 * noise
			}

			intSample := int16(sample * 32767)
			data[pos] += intSample

			if pos+1 < len(data) && j%3 == 0 {
				prev := int32(data[pos])
				next := int32(data[pos+1])
				data[pos] = int16((prev + next) / 2)
			}
		}
	}

	for i := range data {
		if data[i] > 32767 {
			data[i] = 32767
		} else if data[i] < -32768 {
			data[i] = -32768
		}
	}
}

func (s *TTSService) toneForChar(r rune) int {
	tones := map[rune]int{
		'a': 0, 'e': 2, 'i': 4, 'o': -2, 'u': -1,
		'A': 1, 'B': -1, 'C': 3, 'D': 0, 'E': 2,
	}
	if t, ok := tones[r]; ok {
		return t
	}
	return (int(r) % 11) - 5
}

func (s *TTSService) encodeWAV(samples []int16, sampleRate int, bitsPerSample int, numChannels int) []byte {
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(samples) * bitsPerSample / 8
	fileSize := 44 + dataSize

	var buf bytes.Buffer

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(fileSize))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, uint16(numChannels))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(byteRate))
	binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))
	binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(dataSize))

	for _, sample := range samples {
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	return buf.Bytes()
}

func (s *TTSService) saveAudioFile(audioData []byte, format string) (string, string, error) {
	fileID := uuid.New().String()
	dateDir := time.Now().Format("2006/01/02")
	fullDir := filepath.Join(s.audioDir, dateDir)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", "", err
	}

	fileName := fmt.Sprintf("%s.%s", fileID, format)
	fullPath := filepath.Join(fullDir, fileName)

	if err := os.WriteFile(fullPath, audioData, 0644); err != nil {
		return "", "", err
	}

	relativePath := filepath.Join(dateDir, fileName)
	audioURL := fmt.Sprintf("%s/%s", s.audioURLBase, strings.ReplaceAll(relativePath, "\\", "/"))

	return fullPath, audioURL, nil
}
