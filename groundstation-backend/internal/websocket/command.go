package websocket

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"groundstation-backend/internal/mavlink"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/nsq"
	"groundstation-backend/internal/service"
)

var commandHub = NewHub()
var missionService = service.NewMissionService()

type CommandRequest struct {
	Command   string                 `json:"command" binding:"required"`
	UAVID     uint64                 `json:"uav_id" binding:"required"`
	Params    map[string]interface{} `json:"params,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

type CommandResponse struct {
	RequestID string                 `json:"request_id,omitempty"`
	UAVID     uint64                 `json:"uav_id"`
	UavID     uint64                 `json:"uavId"`
	Command   string                 `json:"command"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

func HandleUAVCommand(userID uint64, uavID uint64, command string, params map[string]interface{}) {
	cmd := &CommandRequest{
		Command: command,
		UAVID:   uavID,
		Params:  params,
	}

	success, err := ExecuteCommand(cmd)
	broadcastCmd := command
	result := true
	msg := "命令已发送"
	if err != nil {
		result = false
		msg = err.Error()
	}

	BroadcastCommandResponse(uavID, broadcastCmd, success && result, msg)
}

func RequestUAVTelemetry(uavID uint64) {
	telemetryService := service.NewFlightService()
	data, err := telemetryService.GetRealtimeData(uavID)
	if err != nil {
		return
	}
	commandHub.BroadcastUAVTelemetry(uavID, data)
}

func ExecuteCommand(cmd *CommandRequest) (bool, error) {
	switch cmd.Command {
	case "arm":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_COMPONENT_ARM_DISARM, 1, 0, 0, 0, 0, 0, 0)

	case "disarm":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_COMPONENT_ARM_DISARM, 0, 0, 0, 0, 0, 0, 0)

	case "takeoff":
		altitude, _ := cmd.Params["altitude"].(float64)
		if altitude <= 0 {
			altitude = 10
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_TAKEOFF, 0, 0, 0, 0, 0, 0, float32(altitude))

	case "land":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_LAND, 0, 0, 0, 0, 0, 0, 0)

	case "rtl", "return_to_launch":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)

	case "set_mode":
		mode, _ := cmd.Params["mode"].(string)
		customMode := mavlink.GetFlightModeCode(mode)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SET_MODE, 1, float32(customMode), 0, 0, 0, 0, 0)

	case "goto":
		lat, _ := cmd.Params["lat"].(float64)
		lng, _ := cmd.Params["lng"].(float64)
		alt, _ := cmd.Params["alt"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_WAYPOINT, 0, 0, 0, 0, float32(lat), float32(lng), float32(alt))

	case "mission_start", "start_mission":
		missionID := parseMissionID(cmd.Params)
		if missionID == 0 {
			return false, errors.New("无效的任务ID")
		}
		_, err := missionService.StartMission(missionID)
		return err == nil, err

	case "mission_pause", "pause_mission":
		missionID := parseMissionID(cmd.Params)
		if missionID == 0 {
			return false, errors.New("无效的任务ID")
		}
		_, err := missionService.PauseMission(missionID)
		return err == nil, err

	case "mission_resume", "resume_mission":
		missionID := parseMissionID(cmd.Params)
		if missionID == 0 {
			return false, errors.New("无效的任务ID")
		}
		_, err := missionService.ResumeMission(missionID)
		return err == nil, err

	case "mission_stop", "abort_mission":
		missionID := parseMissionID(cmd.Params)
		if missionID == 0 {
			return false, errors.New("无效的任务ID")
		}
		reason, _ := cmd.Params["reason"].(string)
		_, err := missionService.AbortMission(missionID, reason)
		return err == nil, err

	case "mission_set_current":
		waypointIndex, _ := cmd.Params["waypointIndex"].(float64)
		if waypointIndex < 0 {
			waypointIndex, _ = cmd.Params["waypoint_index"].(float64)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_MISSION_SET_CURRENT, float32(waypointIndex), 0, 0,  0, 0, 0, 0)

	case "mission_upload":
		return true, nil

	case "velocity":
		vx, _ := cmd.Params["vx"].(float64)
		vy, _ := cmd.Params["vy"].(float64)
		vz, _ := cmd.Params["vz"].(float64)
		yawRate, _ := cmd.Params["yawRate"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SET_POSITION_TARGET_LOCAL_NED,
			float32(vx), float32(vy), float32(vz), float32(yawRate), 0, 0, 0)

	case "yaw":
		angle, _ := cmd.Params["angle"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_CONDITION_YAW, float32(angle), 0, 0, 0, 0, 0, 0)

	case "set_home":
		lat, _ := cmd.Params["lat"].(float64)
		lng, _ := cmd.Params["lng"].(float64)
		alt, _ := cmd.Params["alt"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SET_HOME, 0, 0, 0, 0, float32(lat), float32(lng), float32(alt))

	case "set_home_current":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SET_HOME, 1, 0, 0, 0, 0, 0, 0)

	case "change_speed":
		speedType, _ := cmd.Params["speedType"].(float64)
		speed, _ := cmd.Params["speed"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_CHANGE_SPEED, float32(speedType), float32(speed), 0, 0, 0, 0, 0)

	case "reboot":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_PREFLIGHT_REBOOT_SHUTDOWN, 1, 0, 0, 0, 0, 0, 0)

	case "calibrate":
		calType, _ := cmd.Params["type"].(string)
		switch calType {
		case "gyro":
			return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_PREFLIGHT_CALIBRATION, 1, 0, 0, 0, 0, 0, 0)
		case "compass":
			return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_PREFLIGHT_CALIBRATION, 0, 1, 0, 0, 0, 0, 0)
		case "accelerometer":
			return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_PREFLIGHT_CALIBRATION, 0, 0, 1, 0, 0, 0, 0)
		case "level":
			return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_PREFLIGHT_CALIBRATION, 0, 0, 0, 1, 0, 0, 0)
		}
		return false, errors.New("未知的校准类型")

	case "camera_trigger":
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_DIGICAM_CONTROL, 0, 1, 0, 0, 0, 0, 0)

	case "camera_take_photo":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		if payloadID > 0 {
			_ = service.NewPayloadService().TakePhoto(cmd.UAVID, payloadID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_DIGICAM_CONTROL, 0, 1, 0, 0, 0, 0, 0)

	case "camera_start_recording":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		if payloadID > 0 {
			_ = service.NewPayloadService().StartVideoRecording(cmd.UAVID, payloadID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_VIDEO_START, 0, 0, 0, 0, 0, 0, 0)

	case "camera_stop_recording":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		if payloadID > 0 {
			_ = service.NewPayloadService().StopVideoRecording(cmd.UAVID, payloadID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_VIDEO_STOP, 0, 0, 0, 0, 0, 0, 0)

	case "camera_set_zoom":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		zoom, _ := cmd.Params["zoom_level"].(float64)
		if payloadID > 0 {
			_ = service.NewPayloadService().SetCameraZoom(cmd.UAVID, payloadID, zoom)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_DIGICAM_CONFIGURE, 4, float32(zoom), 0, 0, 0, 0, 0)

	case "camera_set_mode":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		mode, _ := cmd.Params["mode"].(string)
		if payloadID > 0 && mode != "" {
			_ = service.NewPayloadService().SetCameraMode(cmd.UAVID, payloadID, models.CameraMode(mode))
		}
		return true, nil

	case "sprayer_start":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		flowRate, _ := cmd.Params["flow_rate"].(float64)
		if flowRate <= 0 {
			flowRate = 2.0
		}
		if payloadID > 0 {
			_ = service.NewPayloadService().StartSpraying(cmd.UAVID, payloadID, flowRate)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SPRAYER, 1, float32(flowRate), 0, 0, 0, 0, 0)

	case "sprayer_stop":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		if payloadID > 0 {
			_ = service.NewPayloadService().StopSpraying(cmd.UAVID, payloadID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SPRAYER, 0, 0, 0, 0, 0, 0, 0)

	case "sprayer_set_flow":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		flowRate, _ := cmd.Params["flow_rate"].(float64)
		if payloadID > 0 {
			_ = service.NewPayloadService().SetSprayerFlowRate(cmd.UAVID, payloadID, flowRate)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SPRAYER, 1, float32(flowRate), 0, 0, 0, 0, 0)

	case "speaker_play_audio":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		audioID := parseParamUint64(cmd.Params, "audio_id")
		if payloadID > 0 && audioID > 0 {
			_ = service.NewPayloadService().PlaySpeakerAudio(cmd.UAVID, payloadID, audioID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_PLAY_TUNE, float32(audioID), 0, 0, 0, 0, 0, 0)

	case "speaker_stop":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		if payloadID > 0 {
			_ = service.NewPayloadService().StopSpeaker(cmd.UAVID, payloadID)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_PLAY_TUNE, 0, 0, 0, 0, 0, 0, 0)

	case "speaker_tts":
		payloadID := parseParamUint64(cmd.Params, "payload_id")
		text, _ := cmd.Params["text"].(string)
		if payloadID > 0 && text != "" {
			ttsTask := &models.TextToSpeechTask{
				UAVID:     cmd.UAVID,
				PayloadID: payloadID,
				Text:      text,
			}
			if voice, ok := cmd.Params["voice"].(string); ok {
				ttsTask.Voice = voice
			}
			if speed, ok := cmd.Params["speed"].(float64); ok {
				ttsTask.Speed = speed
			}
			if pitch, ok := cmd.Params["pitch"].(float64); ok {
				ttsTask.Pitch = pitch
			}
			if volume, ok := cmd.Params["volume"].(float64); ok {
				ttsTask.Volume = int(volume)
			}
			_, _ = service.NewPayloadMissionService().CreateTTSTask(ttsTask, 0)
		}
		return true, nil

	case "orbit_start":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().StartOrbitMission(missionID)
			return err == nil, err
		}
		lat, _ := cmd.Params["lat"].(float64)
		lng, _ := cmd.Params["lng"].(float64)
		alt, _ := cmd.Params["alt"].(float64)
		radius, _ := cmd.Params["radius"].(float64)
		loops, _ := cmd.Params["loops"].(float64)
		if radius <= 0 {
			radius = 30
		}
		if alt <= 0 {
			alt = 50
		}
		if loops <= 0 {
			loops = 1
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_LOITER_TURNS,
			float32(loops), float32(radius), 0, 0, float32(lat), float32(lng), float32(alt))

	case "orbit_pause":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().PauseOrbitMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_LOITER_UNLIMITED, 0, 0, 0, 0, 0, 0, 0)

	case "orbit_resume":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().ResumeOrbitMission(missionID)
			return err == nil, err
		}
		return true, nil

	case "orbit_abort":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().AbortOrbitMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)

	case "ortho_start":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().StartOrthoMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_MISSION_START, 0, 0, 0, 0, 0, 0, 0)

	case "ortho_pause":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().PauseOrthoMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_LOITER_UNLIMITED, 0, 0, 0, 0, 0, 0, 0)

	case "ortho_resume":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().ResumeOrthoMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_MISSION_START, 0, 0, 0, 0, 0, 0, 0)

	case "ortho_abort":
		missionID := parseParamUint64(cmd.Params, "mission_id")
		if missionID > 0 {
			_, err := service.NewPayloadMissionService().AbortOrthoMission(missionID)
			return err == nil, err
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_RETURN_TO_LAUNCH, 0, 0, 0, 0, 0, 0, 0)

	case "do_action":
		return true, nil

	case "param_set", "param_request":
		return true, nil

	case "mount_control":
		return true, nil

	case "set_obstacle_avoidance_config":
		enabled := float32(0)
		if v, ok := cmd.Params["param1"].(float64); ok && v > 0 {
			enabled = 1
		}
		if v, ok := cmd.Params["enabled"].(bool); ok && v {
			enabled = 1
		}
		sensitivity, _ := cmd.Params["param2"].(float64)
		strategy, _ := cmd.Params["param3"].(float64)
		detectionRange, _ := cmd.Params["param4"].(float64)
		ascendHeight, _ := cmd.Params["param5"].(float64)
		retreatDistance, _ := cmd.Params["param6"].(float64)
		bypassAngle, _ := cmd.Params["param7"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_OBSTACLE_AVOIDANCE_CONFIG,
			enabled, float32(sensitivity), float32(strategy), float32(detectionRange),
			float32(ascendHeight), float32(retreatDistance), float32(bypassAngle))

	case "trigger_thrust_learning":
		enable := float32(1)
		if v, ok := cmd.Params["param1"].(float64); ok {
			enable = float32(v)
		}
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_THRUST_LEARNING_CONFIG, enable, 0, 0, 0, 0, 0, 0)

	case "set_pid_gains":
		param1, _ := cmd.Params["param1"].(float64)
		param2, _ := cmd.Params["param2"].(float64)
		param3, _ := cmd.Params["param3"].(float64)
		param4, _ := cmd.Params["param4"].(float64)
		param5, _ := cmd.Params["param5"].(float64)
		param6, _ := cmd.Params["param6"].(float64)
		param7, _ := cmd.Params["param7"].(float64)
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_DO_SET_PID_GAINS,
			float32(param1), float32(param2), float32(param3),
			float32(param4), float32(param5), float32(param6),
			float32(param7))

	default:
		return false, errors.New("未知的命令类型: " + cmd.Command)
	}
}

func parseMissionID(params map[string]interface{}) uint64 {
	if params == nil {
		return 0
	}
	parseID := func(v interface{}) uint64 {
		switch val := v.(type) {
		case float64:
			return uint64(val)
		case string:
			n, _ := strconv.ParseUint(val, 10, 64)
			return n
		case uint64:
			return val
		case int64:
			return uint64(val)
		case int:
			return uint64(val)
		}
		return 0
	}
	if v, ok := params["missionId"]; ok {
		if id := parseID(v); id > 0 {
			return id
		}
	}
	if v, ok := params["mission_id"]; ok {
		return parseID(v)
	}
	return 0
}

func parseParamUint64(params map[string]interface{}, key string) uint64 {
	if params == nil {
		return 0
	}
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case float64:
			return uint64(val)
		case string:
			n, _ := strconv.ParseUint(val, 10, 64)
			return n
		case uint64:
			return val
		case int64:
			return uint64(val)
		case int:
			return uint64(val)
		}
	}
	return 0
}

func sendMAVLinkCommand(uavID uint64, command uint16, params ...float32) (bool, error) {
	cmdMsg := mavlink.EncodeCommandLong(uavID, command, params...)
	err := nsq.Publish(nsq.TopicMAVLinkCommand, cmdMsg)
	if err != nil {
		return false, err
	}
	return true, nil
}

func SendCommandToUAV(uavID uint64, command string, params map[string]interface{}) error {
	cmd := &CommandRequest{
		Command: command,
		UAVID:   uavID,
		Params:  params,
	}
	_, err := ExecuteCommand(cmd)
	return err
}

func BroadcastCommandResponse(uavID uint64, command string, success bool, message string) {
	response := &CommandResponse{
		UAVID:     uavID,
		UavID:     uavID,
		Command:   command,
		Success:   success,
		Message:   message,
		Timestamp: time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type:    "command_result",
		Data:    response,
		Payload: response,
		UAVID:   uavID,
		UavID:   uavID,
		Time:    response.Timestamp,
	}

	bytes, _ := json.Marshal(msg)
	commandHub.broadcast <- bytes
}
