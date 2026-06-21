package handler

import (
	"net/http"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"time"
)

var linkService = service.NewLinkService()

type ReportLinkStatusRequest struct {
	UAVID          uint64  `json:"uav_id" binding:"required"`
	ActiveLink     uint8   `json:"active_link" binding:"required,min=1,max=3"`
	RadioRSSI      int8    `json:"radio_rssi"`
	RadioConnected bool    `json:"radio_connected"`
	LteRSSI        int8    `json:"lte_rssi"`
	LteConnected   bool    `json:"lte_connected"`
	LteNetworkType string  `json:"lte_network_type"`
	PacketLoss     float64 `json:"packet_loss"`
	LatencyMs      uint32  `json:"latency_ms"`
}

func ReportLinkStatus(c *gin.Context) {
	var req ReportLinkStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	report := &service.LinkStatusReport{
		ActiveLink:     req.ActiveLink,
		RadioRSSI:      req.RadioRSSI,
		RadioConnected: req.RadioConnected,
		LteRSSI:        req.LteRSSI,
		LteConnected:   req.LteConnected,
		LteNetworkType: req.LteNetworkType,
		PacketLoss:     req.PacketLoss,
		LatencyMs:      req.LatencyMs,
	}

	status, err := linkService.ReportStatus(req.UAVID, report)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "上报成功", status)
}

func GetLinkStatus(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	status, err := linkService.GetLatestByUAVID(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "链路状态不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", status)
}

func GetLinkHistory(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := utils.ParseTime(startTimeStr)
		if err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		t, err := utils.ParseTime(endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	history, total, err := linkService.GetHistory(uavID, pagination.Page, pagination.PageSize, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", history, total)
}

func GetLinkStatistics(c *gin.Context) {
	uavIDStr := c.Query("uav_id")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var uavID *uint64
	if uavIDStr != "" {
		id, err := utils.ParseUint64(uavIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
			return
		}
		uavID = &id
	}

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := utils.ParseTime(startTimeStr)
		if err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		t, err := utils.ParseTime(endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	stats, err := linkService.GetStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}
