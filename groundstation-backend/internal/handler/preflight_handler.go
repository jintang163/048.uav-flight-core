package handler

import (
	"net/http"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var preflightService = service.NewPreflightService()

type RunPreflightRequest struct {
	UAVID              uint64   `json:"uav_id" binding:"required"`
	MinSatellites      *int     `json:"min_satellites"`
	MinGPSFixType      *int     `json:"min_gps_fix_type"`
	MaxHDOP            *float64 `json:"max_hdop"`
	MinVoltage         *float64 `json:"min_voltage"`
	MinVoltagePerCell  *float64 `json:"min_voltage_per_cell"`
	CellCount          *int     `json:"cell_count"`
	MinStorageSpaceMB  *int64   `json:"min_storage_space_mb"`
	MinSignalStrength  *int     `json:"min_signal_strength"`
	MinLinkQuality     *int     `json:"min_link_quality"`
}

func buildThresholdsFromRequest(req *RunPreflightRequest) *service.PreflightCheckThresholds {
	t := service.DefaultThresholds

	if req.MinSatellites != nil {
		t.MinSatellites = *req.MinSatellites
	}
	if req.MinGPSFixType != nil {
		t.MinGPSFixType = *req.MinGPSFixType
	}
	if req.MaxHDOP != nil {
		t.MaxHDOP = *req.MaxHDOP
	}
	if req.MinVoltage != nil {
		t.MinVoltage = *req.MinVoltage
	}
	if req.MinVoltagePerCell != nil {
		t.MinVoltagePerCell = *req.MinVoltagePerCell
	}
	if req.CellCount != nil {
		t.CellCount = *req.CellCount
	}
	if req.MinStorageSpaceMB != nil {
		t.MinStorageSpaceMB = *req.MinStorageSpaceMB
	}
	if req.MinSignalStrength != nil {
		t.MinSignalStrength = *req.MinSignalStrength
	}
	if req.MinLinkQuality != nil {
		t.MinLinkQuality = *req.MinLinkQuality
	}

	return t
}

func RunPreflightCheck(c *gin.Context) {
	var req RunPreflightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	thresholds := buildThresholdsFromRequest(&req)
	result, err := preflightService.RunPreflightCheck(req.UAVID, thresholds)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "自检失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "自检完成", result)
}

func GetPreflightThresholds(c *gin.Context) {
	utils.SuccessResponse(c, "获取成功", service.DefaultThresholds)
}

func BatchRunPreflightCheck(c *gin.Context) {
	type BatchRequest struct {
		UAVIDs []uint64 `json:"uav_ids" binding:"required"`
	}

	var req BatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	results := make([]*service.PreflightCheckResult, 0, len(req.UAVIDs))
	errors := make([]map[string]interface{}, 0)

	for _, uavID := range req.UAVIDs {
		result, err := preflightService.RunPreflightCheck(uavID, service.DefaultThresholds)
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"uav_id": uavID,
				"error":  err.Error(),
			})
			continue
		}
		results = append(results, result)
	}

	utils.SuccessResponse(c, "批量自检完成", map[string]interface{}{
		"results": results,
		"errors":  errors,
		"total":   len(req.UAVIDs),
		"success": len(results),
		"failed":  len(errors),
	})
}
