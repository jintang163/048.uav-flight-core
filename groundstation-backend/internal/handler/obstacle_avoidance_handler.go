package handler

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/internal/websocket"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var obstacleAvoidanceService = service.NewObstacleAvoidanceService()

type UpdateAvoidanceConfigRequest struct {
	Enabled         *bool                       `json:"enabled"`
	Sensitivity     *models.AvoidanceSensitivity `json:"sensitivity"`
	Strategy        *models.AvoidanceStrategy   `json:"strategy"`
	SensorType      *models.ObstacleSensorType  `json:"sensor_type"`
	DetectionRange  *float64                    `json:"detection_range"`
	AscendHeight    *float64                    `json:"ascend_height"`
	RetreatDistance  *float64                   `json:"retreat_distance"`
	BypassAngle     *float64                    `json:"bypass_angle"`
}

func GetObstacleAvoidanceConfig(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	config, err := obstacleAvoidanceService.GetConfig(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取避障配置成功", config)
}

func UpdateObstacleAvoidanceConfig(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req UpdateAvoidanceConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Sensitivity != nil {
		updates["sensitivity"] = *req.Sensitivity
	}
	if req.Strategy != nil {
		updates["strategy"] = *req.Strategy
	}
	if req.SensorType != nil {
		updates["sensor_type"] = *req.SensorType
	}
	if req.DetectionRange != nil {
		updates["detection_range"] = *req.DetectionRange
	}
	if req.AscendHeight != nil {
		updates["ascend_height"] = *req.AscendHeight
	}
	if req.RetreatDistance != nil {
		updates["retreat_distance"] = *req.RetreatDistance
	}
	if req.BypassAngle != nil {
		updates["bypass_angle"] = *req.BypassAngle
	}

	config, err := obstacleAvoidanceService.UpdateConfig(uavID, updates)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	sendObstacleAvoidanceConfigToFC(uavID, config)

	utils.SuccessResponse(c, "更新避障配置成功", config)
}

func sendObstacleAvoidanceConfigToFC(uavID uint64, cfg *models.ObstacleAvoidanceConfig) {
	enabled := 0
	if cfg.Enabled {
		enabled = 1
	}
	sensitivity := 1
	switch cfg.Sensitivity {
	case models.SensitivityFar:
		sensitivity = 0
	case models.SensitivityMedium:
		sensitivity = 1
	case models.SensitivityNear:
		sensitivity = 2
	}
	strategy := 0
	switch cfg.Strategy {
	case models.StrategyHover:
		strategy = 0
	case models.StrategyAscendBypass:
		strategy = 1
	case models.StrategyRetreatBypass:
		strategy = 2
	}

	params := map[string]interface{}{
		"param1": float32(enabled),
		"param2": float32(sensitivity),
		"param3": float32(strategy),
		"param4": float32(cfg.DetectionRange),
		"param5": float32(cfg.AscendHeight),
		"param6": float32(cfg.RetreatDistance),
		"param7": float32(cfg.BypassAngle),
	}

	_ = websocket.SendCommandToUAV(uavID, "obstacle_avoidance_config", params)
}

func GetObstacleHeatmap(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uavId"), 10, 64)

	var startTime, endTime time.Time
	if st := c.Query("startTime"); st != "" {
		if ts, err := strconv.ParseInt(st, 10, 64); err == nil {
			startTime = time.Unix(ts/1000, 0)
		}
	}
	if et := c.Query("endTime"); et != "" {
		if ts, err := strconv.ParseInt(et, 10, 64); err == nil {
			endTime = time.Unix(ts/1000, 0)
		}
	}

	points, err := obstacleAvoidanceService.GetHeatmap(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取热力图数据成功", points)
}

func GetObstacleAvoidanceLogs(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uavId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	var startTime, endTime time.Time
	if st := c.Query("startTime"); st != "" {
		if ts, err := strconv.ParseInt(st, 10, 64); err == nil {
			startTime = time.Unix(ts/1000, 0)
		}
	}
	if et := c.Query("endTime"); et != "" {
		if ts, err := strconv.ParseInt(et, 10, 64); err == nil {
			endTime = time.Unix(ts/1000, 0)
		}
	}

	logs, total, err := obstacleAvoidanceService.ListEvents(uavID, c.Query("status"), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取避障日志成功", gin.H{
		"list":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func GetObstacleAvoidanceStatistics(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uavId"), 10, 64)

	var startTime, endTime time.Time
	if st := c.Query("startTime"); st != "" {
		if ts, err := strconv.ParseInt(st, 10, 64); err == nil {
			startTime = time.Unix(ts/1000, 0)
		}
	}
	if et := c.Query("endTime"); et != "" {
		if ts, err := strconv.ParseInt(et, 10, 64); err == nil {
			endTime = time.Unix(ts/1000, 0)
		}
	}

	stats, err := obstacleAvoidanceService.GetStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取避障统计成功", stats)
}

func GetAvoidanceEvents(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uavId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	events, total, err := obstacleAvoidanceService.ListEvents(uavID, c.Query("status"), page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取避障事件成功", gin.H{
		"list":     events,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func GetAvoidanceEventDetail(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid event id", nil)
		return
	}

	event, err := obstacleAvoidanceService.GetEvent(eventID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "事件不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取避障事件详情成功", event)
}

func TriggerAvoidanceTest(c *gin.Context) {
	uavID, err := strconv.ParseUint(c.Param("uav_id"), 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "invalid uav_id", nil)
		return
	}

	var req struct {
		Strategy models.AvoidanceStrategy `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误", nil)
		return
	}

	if err := obstacleAvoidanceService.TriggerAvoidanceTest(uavID, req.Strategy); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "触发避障测试成功", gin.H{"success": true, "message": "避障测试已触发"})
}

func ClearObstacleHeatmap(c *gin.Context) {
	uavID, _ := strconv.ParseUint(c.Query("uavId"), 10, 64)

	if err := obstacleAvoidanceService.ClearHeatmap(uavID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "清除热力图数据成功", nil)
}
