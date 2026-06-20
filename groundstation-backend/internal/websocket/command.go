package websocket

import (
	"encoding/json"
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
	Command   string                 `json:"command"`
	Success   bool                   `json:"success"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

func (c *Client) handleCommand(msg ClientMessage) {
	var cmd CommandRequest
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		c.hub.SendToClient(c, "command_error", gin.H{
			"error": "无效的命令格式",
			"request_id": cmd.RequestID,
		})
		return
	}

	cmd.UAVID = msg.UAVID
	cmd.RequestID = msg.Action

	response := &CommandResponse{
		RequestID: cmd.RequestID,
		UAVID:     cmd.UAVID,
		Command:   cmd.Command,
		Timestamp: time.Now().UnixNano() / 1e6,
	}

	success, err := ExecuteCommand(&cmd)
	if err != nil {
		response.Success = false
		response.Message = err.Error()
	} else {
		response.Success = success
		response.Message = "命令已发送"
	}

	c.hub.SendToClient(c, "command_response", response)
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
		return sendMAVLinkCommand(cmd.UAVID, mavlink.CMD_NAV_TAKEOFF, 0, 0, 0, 0, 0, 0, altitude)

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

	case "start_mission":
		missionID, _ := cmd.Params["mission_id"].(uint64)
		_, err := missionService.StartMission(missionID)
		return err == nil, err

	case "pause_mission":
		missionID, _ := cmd.Params["mission_id"].(uint64)
		_, err := missionService.PauseMission(missionID)
		return err == nil, err

	case "resume_mission":
		missionID, _ := cmd.Params["mission_id"].(uint64)
		_, err := missionService.ResumeMission(missionID)
		return err == nil, err

	case "abort_mission":
		missionID, _ := cmd.Params["mission_id"].(uint64)
		reason, _ := cmd.Params["reason"].(string)
		_, err := missionService.AbortMission(missionID, reason)
		return err == nil, err

	default:
		return false, errors.New("未知的命令类型")
	}
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
		Command:   command,
		Success:   success,
		Message:   message,
		Timestamp: time.Now().UnixNano() / 1e6,
	}

	msg := &Message{
		Type: "command_result",
		Data: response,
		Time: response.Timestamp,
	}

	bytes, _ := json.Marshal(msg)
	commandHub.broadcast <- bytes
}

import "time"
import "errors"
import "github.com/gin-gonic/gin"
