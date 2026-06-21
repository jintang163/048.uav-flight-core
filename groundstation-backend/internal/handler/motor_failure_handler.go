package handler

import (
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func GetMotorStatuses(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	statuses := motorService.GetMotorStatuses(uavID)
	if statuses == nil {
		statuses = []*models.MotorStatus{}
	}

	utils.SuccessResponse(c, "获取成功", statuses)
}

func GetMotorFailureState(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	state := motorService.GetFailureState(uavID)
	if state == nil {
		utils.SuccessResponse(c, "获取成功", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{
		"uav_id":        state.UAVID,
		"failed_motors": state.FailedMotors,
		"pid_adjusted":  state.PIDAdjusted,
		"rth_triggered": state.RTHTriggered,
		"start_time":    state.StartTime,
		"last_update":   state.LastUpdateTime,
		"motor_count":   motorService.GetMotorCount(uavID),
	})
}

type ManualPIDRequest struct {
	PGain float64 `json:"p_gain"`
	IGain float64 `json:"i_gain"`
	DGain float64 `json:"d_gain"`
}

func ManualPIDAdjustment(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req ManualPIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "invalid request body", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	params := map[string]float64{
		"0": req.PGain,
		"1": req.IGain,
		"2": req.DGain,
	}

	if err := motorService.SendManualPIDAdjustment(uavID, params); err != nil {
		middleware.Logger.Error("Manual PID adjustment failed", zap.Uint64("uav_id", uavID), zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "PID调整指令发送失败", nil)
		return
	}

	utils.SuccessResponse(c, "PID调整指令已发送", nil)
}

func EmergencyRTH(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	if err := motorService.TriggerManualRTH(uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "紧急返航指令发送失败", nil)
		return
	}

	utils.SuccessResponse(c, "紧急返航指令已发送", nil)
}

func EmergencyLand(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	if err := motorService.TriggerLand(uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "紧急降落指令发送失败", nil)
		return
	}

	utils.SuccessResponse(c, "紧急降落指令已发送", nil)
}

func ResolveMotorFailure(c *gin.Context) {
	uavIDStr := c.Param("uav_id")
	uavID, err := strconv.ParseUint(uavIDStr, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	motorIndexStr := c.Param("motor_index")
	motorIndex, err := strconv.Atoi(motorIndexStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid motor_index", nil)
		return
	}

	motorService := service.NewMotorFailureService()
	if err := motorService.ResolveFailure(uavID, motorIndex); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "解除失败", nil)
		return
	}

	utils.SuccessResponse(c, "已解除电机失效状态", nil)
}
