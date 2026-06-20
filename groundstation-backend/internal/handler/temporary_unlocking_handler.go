package handler

import (
	"net/http"

	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var unlockingService = service.NewTemporaryUnlockingService()

type ApplyUnlockingRequest struct {
	UAVID       uint64 `json:"uav_id" binding:"required"`
	GeofenceID  uint64 `json:"geofence_id"`
	Title       string `json:"title" binding:"required,max=200"`
	Reason      string `json:"reason" binding:"required"`
	Category    string `json:"category"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	MaxAltitude float64 `json:"max_altitude"`
	MaxDistance float64 `json:"max_distance"`
	CenterLat   float64 `json:"center_lat"`
	CenterLng   float64 `json:"center_lng"`
	Radius      float64 `json:"radius"`
	MissionID   uint64 `json:"mission_id"`
	ContactName string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
}

func ApplyUnlocking(c *gin.Context) {
	var req ApplyUnlockingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	category := models.GeofenceCategoryCustom
	if req.Category != "" {
		category = models.GeofenceCategory(req.Category)
	}

	unlocking := &models.TemporaryUnlocking{
		UAVID:        req.UAVID,
		GeofenceID:   req.GeofenceID,
		Title:        req.Title,
		Reason:       req.Reason,
		Category:     category,
		MaxAltitude:  req.MaxAltitude,
		MaxDistance:  req.MaxDistance,
		CenterLat:    req.CenterLat,
		CenterLng:    req.CenterLng,
		Radius:       req.Radius,
		MissionID:    req.MissionID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
	}

	if req.StartTime != "" {
		if t, err := utils.ParseTime(req.StartTime); err == nil {
			unlocking.StartTime = &t
		}
	}
	if req.EndTime != "" {
		if t, err := utils.ParseTime(req.EndTime); err == nil {
			unlocking.EndTime = &t
		}
	}

	result, err := unlockingService.Apply(unlocking, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "申请成功", result)
}

func GetUnlocking(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的申请ID", nil)
		return
	}

	unlocking, err := unlockingService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "申请不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", unlocking)
}

func ListUnlockings(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	uavID, _ := utils.ParseUint64(c.Query("uav_id"))
	applicantID, _ := utils.ParseUint64(c.Query("applicant_id"))
	status := c.Query("status")
	category := c.Query("category")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	unlockings, total, err := unlockingService.List(pagination, uavID, applicantID, status, category, startTime, endTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", unlockings, total)
}

func ApproveUnlocking(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的申请ID", nil)
		return
	}

	var data struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&data)

	approverID := middleware.GetCurrentUserID(c)

	if err := unlockingService.Approve(id, approverID, data.Remark); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "审批通过", nil)
}

func RejectUnlocking(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的申请ID", nil)
		return
	}

	var data struct {
		Remark string `json:"remark"`
	}
	_ = c.ShouldBindJSON(&data)

	approverID := middleware.GetCurrentUserID(c)

	if err := unlockingService.Reject(id, approverID, data.Remark); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已驳回", nil)
}

func CancelUnlocking(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的申请ID", nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)

	if err := unlockingService.Cancel(id, userID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "已取消", nil)
}

func GetActiveUnlockings(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	category := c.Query("category")

	unlockings, err := unlockingService.GetActiveUnlockings(uavID, category)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", unlockings)
}
