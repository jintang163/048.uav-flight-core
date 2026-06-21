package handler

import (
	"net/http"
	"time"

	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var blackboxService = service.NewBlackboxService()
var blackboxRepo = repository.NewBlackboxRepository()

type UploadLogRequest struct {
	UAVID      uint64 `form:"uav_id" binding:"required"`
	MissionID  uint64 `form:"mission_id"`
	LogType    string `form:"log_type"`
	StartTime  string `form:"start_time"`
	EndTime    string `form:"end_time"`
	FlightName string `form:"flight_name"`
	Notes      string `form:"notes"`
}

func UploadBlackbox(c *gin.Context) {
	var req UploadLogRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "请选择文件", nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	serviceReq := &service.UploadLogRequest{
		UAVID:      req.UAVID,
		MissionID:  req.MissionID,
		LogType:    req.LogType,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		FlightName: req.FlightName,
		Notes:      req.Notes,
		File:       file,
	}

	log, err := blackboxService.UploadLog(serviceReq, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "上传成功", log)
}

func GetBlackbox(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	log, err := blackboxService.GetLog(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", log)
}

func ListBlackboxes(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	missionID, _ := utils.ParseUint64(c.Query("mission_id"))
	status := c.Query("status")
	crashStr := c.Query("crash_detected")

	var crashDetected *bool
	if crashStr != "" {
		crash := crashStr == "true" || crashStr == "1"
		crashDetected = &crash
	}

	logs, total, err := blackboxService.ListLogs(pagination, uavID, missionID, models.BlackboxLogStatus(status), crashDetected)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", logs, total)
}

func UpdateBlackbox(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	var data struct {
		Status models.BlackboxLogStatus `json:"status"`
		Notes  string                   `json:"notes"`
		Tags   string                   `json:"tags"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	log, err := blackboxService.GetLog(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	if data.Status != "" {
		log.Status = data.Status
	}
	if data.Notes != "" {
		log.Notes = data.Notes
	}
	if data.Tags != "" {
		log.Tags = data.Tags
	}

	if err := blackboxRepo.Update(log); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", log)
}

func DeleteBlackbox(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	if err := blackboxService.DeleteLog(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func DownloadBlackbox(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	log, err := blackboxService.GetLog(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{"download_url": log.FileURL, "file_name": log.FileName})
}

func GetBlackboxStatistics(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var start, end time.Time
	if startTime != "" {
		start, _ = time.Parse(time.RFC3339, startTime)
	}
	if endTime != "" {
		end, _ = time.Parse(time.RFC3339, endTime)
	}

	stats, err := blackboxService.GetStatistics(uavID, start, end)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}

func ParseBlackboxLog(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	data, err := blackboxService.ParseLog(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "解析失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "解析成功", data)
}

func GetAnalysisReport(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	report, err := blackboxService.GenerateAnalysisReport(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "生成报告失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", report)
}

func ExportBlackboxCSV(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	filePath, err := blackboxService.ExportCSV(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "导出失败: "+err.Error(), nil)
		return
	}

	c.FileAttachment(filePath, "flight_log.csv")
}

func ExportBlackboxReport(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	filePath, err := blackboxService.ExportPDF(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "导出失败: "+err.Error(), nil)
		return
	}

	c.FileAttachment(filePath, "flight_report.txt")
}

func GetBlackboxReports(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	reports, err := blackboxService.GetReports(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", reports)
}

func AnalyzeBlackbox(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的日志ID", nil)
		return
	}

	go func() {
		_, _ = blackboxService.GenerateAnalysisReport(id)
	}()

	utils.SuccessResponse(c, "分析已开始", nil)
}

func AutoUploadBlackbox(c *gin.Context) {
	var req service.AutoUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	log, err := blackboxService.AutoUpload(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "自动上传失败: "+err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "自动上传成功", log)
}
