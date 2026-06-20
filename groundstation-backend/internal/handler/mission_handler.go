package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var missionService = service.NewMissionService()

type CreateTemplateRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Category    string                   `json:"category"`
	Waypoints   []models.MissionWaypoint `json:"waypoints" binding:"required,min=1"`
}

type UpdateTemplateRequest struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Category    string                   `json:"category"`
	Waypoints   []models.MissionWaypoint `json:"waypoints"`
}

type CreateMissionRequest struct {
	UAVID         uint64                 `json:"uav_id" binding:"required"`
	TemplateID    uint64                 `json:"template_id"`
	Name          string                 `json:"name" binding:"required"`
	Description   string                 `json:"description"`
	PlannedTime   string                 `json:"planned_time"`
	MaxAltitude   float64                `json:"max_altitude"`
	MaxSpeed      float64                `json:"max_speed"`
}

func CreateTemplate(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	template := &models.MissionTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		CreatorID:   userID,
		Waypoints:   req.Waypoints,
	}

	result, err := missionService.CreateTemplate(template)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetTemplate(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的模板ID", nil)
		return
	}

	template, err := missionService.GetTemplate(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "模板不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", template)
}

func ListTemplates(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	category := c.Query("category")
	keyword := c.Query("keyword")

	templates, total, err := missionService.ListTemplates(pagination, category, keyword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", templates, total)
}

func UpdateTemplate(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的模板ID", nil)
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	template := &models.MissionTemplate{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	result, err := missionService.UpdateTemplate(id, template, req.Waypoints)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteTemplate(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的模板ID", nil)
		return
	}

	if err := missionService.DeleteTemplate(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func CreateMission(c *gin.Context) {
	var req CreateMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	mission := &models.FlightMission{
		UAVID:       req.UAVID,
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Description: req.Description,
		CreatorID:   userID,
		MaxAltitude: req.MaxAltitude,
		MaxSpeed:    req.MaxSpeed,
	}

	result, err := missionService.CreateMission(mission)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	mission, err := missionService.GetMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "任务不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", mission)
}

func ListMissions(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	missions, total, err := missionService.ListMissions(pagination, uavID, status, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", missions, total)
}

func StartMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	mission, err := missionService.StartMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已启动", gin.H{"mission": mission, "operator_id": userID})
}

func PauseMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	mission, err := missionService.PauseMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已暂停", mission)
}

func ResumeMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	mission, err := missionService.ResumeMission(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已恢复", mission)
}

func AbortMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	var data struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&data)

	mission, err := missionService.AbortMission(id, data.Reason)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "任务已终止", mission)
}

func UpdateMission(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	var mission models.FlightMission
	if err := c.ShouldBindJSON(&mission); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	result, err := missionService.UpdateMission(id, &mission)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}
