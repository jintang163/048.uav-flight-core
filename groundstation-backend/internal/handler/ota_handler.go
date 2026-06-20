package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var otaService = service.NewOTAService()

func UploadFirmware(c *gin.Context) {
	name := c.PostForm("name")
	fwType := c.PostForm("type")
	version := c.PostForm("version")
	buildNumber := c.PostForm("build_number")
	hardware := c.PostForm("hardware")
	description := c.PostForm("description")
	changelog := c.PostForm("changelog")
	isMandatory := c.PostForm("is_mandatory") == "true"
	minVersion := c.PostForm("min_version")

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "请选择文件", nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	req := &service.UploadFirmwareRequest{
		Name:        name,
		Type:        models.FirmwareType(fwType),
		Version:     version,
		BuildNumber: buildNumber,
		Hardware:    hardware,
		Description: description,
		Changelog:   changelog,
		IsMandatory: isMandatory,
		MinVersion:  minVersion,
		File:        file,
	}

	firmware, err := otaService.UploadFirmware(req, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "上传成功", firmware)
}

func GetFirmware(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的固件ID", nil)
		return
	}

	firmware, err := otaService.GetFirmware(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "固件不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", firmware)
}

func ListFirmwares(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	fwType := models.FirmwareType(c.Query("type"))
	status := models.FirmwareStatus(c.Query("status"))
	hardware := c.Query("hardware")

	firmwares, total, err := otaService.ListFirmwares(pagination, fwType, status, hardware)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", firmwares, total)
}

func UpdateFirmwareStatus(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的固件ID", nil)
		return
	}

	var data struct {
		Status models.FirmwareStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	firmware, err := otaService.UpdateFirmwareStatus(id, data.Status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "状态已更新", firmware)
}

func DeleteFirmware(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的固件ID", nil)
		return
	}

	if err := otaService.DeleteFirmware(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func GetLatestFirmware(c *gin.Context) {
	fwType := models.FirmwareType(c.Query("type"))
	hardware := c.Query("hardware")

	firmware, err := otaService.GetLatestFirmware(fwType, hardware)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "未找到固件", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", firmware)
}

func StartUpdate(c *gin.Context) {
	var req service.StartUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	req.OperatorID = middleware.GetCurrentUserID(c)
	update, err := otaService.StartUpdate(&req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "升级已启动", update)
}

func GetUpdate(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的升级ID", nil)
		return
	}

	update, err := otaService.GetUpdate(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "升级记录不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", update)
}

func ListUpdates(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	status := c.Query("status")

	updates, total, err := otaService.ListUpdates(pagination, uavID, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", updates, total)
}

func GetActiveUpdate(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	update, err := otaService.GetActiveUpdate(uavID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", update)
}

func DownloadFirmware(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的固件ID", nil)
		return
	}

	fileURL, err := otaService.DownloadFirmware(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", gin.H{"download_url": fileURL})
}
