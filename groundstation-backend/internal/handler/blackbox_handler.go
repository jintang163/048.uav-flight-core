package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/repository"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"time"
)

var blackboxRepo = repository.NewBlackboxRepository()

type UploadLogRequest struct {
	UAVID       uint64                `form:"uav_id" binding:"required"`
	MissionID   uint64                `form:"mission_id"`
	LogType     models.BlackboxType   `form:"log_type" binding:"required"`
	StartTime   string                `form:"start_time"`
	EndTime     string                `form:"end_time"`
	Notes       string                `form:"notes"`
}

func UploadBlackbox(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.PostForm("uav_id"))
	missionID, _ := utils.ParseUint64(c.PostForm("mission_id"))
	logType := c.PostForm("log_type")
	startTime := c.PostForm("start_time")
	endTime := c.PostForm("end_time")
	notes := c.PostForm("notes")

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "请选择文件", nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	objectName := ""
	if file != nil {
		objectName = "blackbox/" + utils.GenerateUUID() + ".ulg"
	}

	log := &models.BlackboxLog{
		UAVID:      uavID,
		MissionID:  missionID,
		LogType:    models.BlackboxType(logType),
		FileURL:    objectName,
		FileSize:   file.Size,
		StartTime:  startTime,
		EndTime:    endTime,
		Notes:      notes,
		UploaderID: userID,
		Status:     models.BlackboxStatusUploaded,
	}

	if err := blackboxRepo.Create(log); err != nil {
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

	log, err := blackboxRepo.FindByID(id)
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
	logType := c.Query("log_type")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	logs, total, err := blackboxRepo.List(pagination, uavID, missionID, logType, status, startTime, endTime)
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
		Status         models.BlackboxStatus `json:"status"`
		AnalysisReport string                `json:"analysis_report"`
		Notes          string                `json:"notes"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	log, err := blackboxRepo.FindByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	if data.Status != "" {
		log.Status = data.Status
	}
	if data.AnalysisReport != "" {
		log.AnalysisReport = data.AnalysisReport
		log.AnalyzedAt = time.Now()
	}
	if data.Notes != "" {
		log.Notes = data.Notes
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

	if err := blackboxRepo.SoftDelete(&models.BlackboxLog{}, id); err != nil {
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

	log, err := blackboxRepo.FindByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "日志不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{"download_url": log.FileURL})
}

func GetBlackboxStatistics(c *gin.Context) {
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	stats, err := blackboxRepo.GetStatistics(uavID, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", stats)
}
