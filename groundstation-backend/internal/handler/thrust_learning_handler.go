package handler

import (
	"groundstation-backend/internal/service"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var thrustLearningService = service.NewThrustLearningService()

type UpdateThrustCurveRequest struct {
	Points []struct {
		Throttle float64 `json:"throttle" binding:"required"`
		Thrust   float64 `json:"thrust" binding:"required"`
		Rpm      float64 `json:"rpm"`
	} `json:"points" binding:"required"`
}

type UpdatePIDGainsRequest struct {
	ProfileName *string  `json:"profile_name"`
	RollKP      *float64 `json:"roll_kp"`
	RollKI      *float64 `json:"roll_ki"`
	RollKD      *float64 `json:"roll_kd"`
	PitchKP     *float64 `json:"pitch_kp"`
	PitchKI     *float64 `json:"pitch_ki"`
	PitchKD     *float64 `json:"pitch_kd"`
	YawKP       *float64 `json:"yaw_kp"`
	YawKI       *float64 `json:"yaw_ki"`
	YawKD       *float64 `json:"yaw_kd"`
	RateRollKP  *float64 `json:"rate_roll_kp"`
	RateRollKI  *float64 `json:"rate_roll_ki"`
	RateRollKD  *float64 `json:"rate_roll_kd"`
	RatePitchKP *float64 `json:"rate_pitch_kp"`
	RatePitchKI *float64 `json:"rate_pitch_ki"`
	RatePitchKD *float64 `json:"rate_pitch_kd"`
	RateYawKP   *float64 `json:"rate_yaw_kp"`
	RateYawKI   *float64 `json:"rate_yaw_ki"`
	RateYawKD   *float64 `json:"rate_yaw_kd"`
	AltKP       *float64 `json:"alt_kp"`
	AltKI       *float64 `json:"alt_ki"`
	AltKD       *float64 `json:"alt_kd"`
}

func GetThrustLearningStatus(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	status, err := thrustLearningService.GetStatus(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取推力学习状态成功", status)
}

func TriggerThrustLearning(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	if err := thrustLearningService.TriggerLearning(uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	params := map[string]interface{}{
		"param1": float32(1),
	}
	_ = websocket.SendCommandToUAV(uavID, "trigger_thrust_learning", params)

	status, _ := thrustLearningService.GetStatus(uavID)
	websocket.BroadcastThrustLearningStatus(uavID, status)

	utils.SuccessResponse(c, "触发推力学习成功", gin.H{"success": true})
}

func GetThrustCurve(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	points, err := thrustLearningService.GetThrustCurve(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取推力曲线成功", points)
}

func UpdateThrustCurve(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdateThrustCurveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	points := make([]struct {
		Throttle float64
		Thrust   float64
		Rpm      float64
	}, len(req.Points))
	for i, p := range req.Points {
		points[i].Throttle = p.Throttle
		points[i].Thrust = p.Thrust
		points[i].Rpm = p.Rpm
	}

	if err := thrustLearningService.StoreCurvePoints(uavID, points); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	updatedPoints, _ := thrustLearningService.GetThrustCurve(uavID)
	websocket.BroadcastThrustCurveUpdate(uavID, updatedPoints)

	utils.SuccessResponse(c, "更新推力曲线成功", updatedPoints)
}

func GetPIDGains(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	profile, err := thrustLearningService.GetPIDGains(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取PID增益成功", profile)
}

func UpdatePIDGains(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdatePIDGainsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	gains := make(map[string]float64)
	if req.RollKP != nil {
		gains["roll_kp"] = *req.RollKP
	}
	if req.RollKI != nil {
		gains["roll_ki"] = *req.RollKI
	}
	if req.RollKD != nil {
		gains["roll_kd"] = *req.RollKD
	}
	if req.PitchKP != nil {
		gains["pitch_kp"] = *req.PitchKP
	}
	if req.PitchKI != nil {
		gains["pitch_ki"] = *req.PitchKI
	}
	if req.PitchKD != nil {
		gains["pitch_kd"] = *req.PitchKD
	}
	if req.YawKP != nil {
		gains["yaw_kp"] = *req.YawKP
	}
	if req.YawKI != nil {
		gains["yaw_ki"] = *req.YawKI
	}
	if req.YawKD != nil {
		gains["yaw_kd"] = *req.YawKD
	}
	if req.RateRollKP != nil {
		gains["rate_roll_kp"] = *req.RateRollKP
	}
	if req.RateRollKI != nil {
		gains["rate_roll_ki"] = *req.RateRollKI
	}
	if req.RateRollKD != nil {
		gains["rate_roll_kd"] = *req.RateRollKD
	}
	if req.RatePitchKP != nil {
		gains["rate_pitch_kp"] = *req.RatePitchKP
	}
	if req.RatePitchKI != nil {
		gains["rate_pitch_ki"] = *req.RatePitchKI
	}
	if req.RatePitchKD != nil {
		gains["rate_pitch_kd"] = *req.RatePitchKD
	}
	if req.RateYawKP != nil {
		gains["rate_yaw_kp"] = *req.RateYawKP
	}
	if req.RateYawKI != nil {
		gains["rate_yaw_ki"] = *req.RateYawKI
	}
	if req.RateYawKD != nil {
		gains["rate_yaw_kd"] = *req.RateYawKD
	}
	if req.AltKP != nil {
		gains["alt_kp"] = *req.AltKP
	}
	if req.AltKI != nil {
		gains["alt_ki"] = *req.AltKI
	}
	if req.AltKD != nil {
		gains["alt_kd"] = *req.AltKD
	}

	profile, err := thrustLearningService.UpdatePIDGains(uavID, gains)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	params := map[string]interface{}{
		"param1":  float32(profile.RollKP),
		"param2":  float32(profile.RollKI),
		"param3":  float32(profile.RollKD),
		"param4":  float32(profile.PitchKP),
		"param5":  float32(profile.PitchKI),
		"param6":  float32(profile.PitchKD),
		"param7":  float32(profile.YawKP),
		"param8":  float32(profile.YawKI),
		"param9":  float32(profile.YawKD),
		"param10": float32(profile.RateRollKP),
		"param11": float32(profile.RateRollKI),
		"param12": float32(profile.RateRollKD),
		"param13": float32(profile.RatePitchKP),
		"param14": float32(profile.RatePitchKI),
		"param15": float32(profile.RatePitchKD),
		"param16": float32(profile.RateYawKP),
		"param17": float32(profile.RateYawKI),
		"param18": float32(profile.RateYawKD),
		"param19": float32(profile.AltKP),
		"param20": float32(profile.AltKI),
		"param21": float32(profile.AltKD),
	}
	_ = websocket.SendCommandToUAV(uavID, "set_pid_gains", params)

	websocket.BroadcastPIDGainsUpdate(uavID, profile)

	utils.SuccessResponse(c, "更新PID增益成功", profile)
}

func ApplyAutoTunedPID(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	if err := thrustLearningService.ApplyAutoTunedPID(uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	profile, _ := thrustLearningService.GetPIDGains(uavID)
	websocket.BroadcastPIDGainsUpdate(uavID, profile)

	utils.SuccessResponse(c, "应用自动调参PID成功", profile)
}

func GetThrustLearningSamples(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))

	samples, err := thrustLearningService.GetSamples(uavID, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取学习样本成功", samples)
}

func OptimizeThrustModel(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	points, err := thrustLearningService.OptimizeModel(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	status, _ := thrustLearningService.GetStatus(uavID)
	websocket.BroadcastThrustLearningStatus(uavID, status)
	websocket.BroadcastThrustCurveUpdate(uavID, points)

	profile, _ := thrustLearningService.GetPIDGains(uavID)
	if profile != nil {
		websocket.BroadcastPIDGainsUpdate(uavID, profile)
	}

	utils.SuccessResponse(c, "优化推力模型成功", gin.H{
		"curve_points": points,
		"status":       status,
	})
}
