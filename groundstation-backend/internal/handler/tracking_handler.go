package handler

import (
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var trackingService = service.NewTrackingService()

type LockTargetRequest struct {
	UAVID        uint64                `json:"uav_id" binding:"required"`
	BboxX        float64               `json:"bbox_x" binding:"required"`
	BboxY        float64               `json:"bbox_y" binding:"required"`
	BboxWidth    float64               `json:"bbox_width" binding:"required"`
	BboxHeight   float64               `json:"bbox_height" binding:"required"`
	FrameWidth   int                   `json:"frame_width"`
	FrameHeight  int                   `json:"frame_height"`
	TargetClass  models.DetectionClass `json:"target_class"`
	Name         string                `json:"name"`
	SearchRadius float64               `json:"search_radius"`
	MaxRadius    float64               `json:"max_radius"`
}

func LockTarget(c *gin.Context) {
	var req LockTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	if req.FrameWidth <= 0 {
		req.FrameWidth = 1280
	}
	if req.FrameHeight <= 0 {
		req.FrameHeight = 720
	}

	userID, _ := c.Get("user_id")
	var uid uint64
	if v, ok := userID.(uint64); ok {
		uid = v
	}

	task, err := trackingService.LockTarget(&service.LockTargetRequest{
		UAVID:        req.UAVID,
		BboxX:        req.BboxX,
		BboxY:        req.BboxY,
		BboxWidth:    req.BboxWidth,
		BboxHeight:   req.BboxHeight,
		FrameWidth:   req.FrameWidth,
		FrameHeight:  req.FrameHeight,
		TargetClass:  req.TargetClass,
		Name:         req.Name,
		CreatedBy:    uid,
		SearchRadius: req.SearchRadius,
		MaxRadius:    req.MaxRadius,
	})
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "目标锁定成功", task)
}

func StopTracking(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	if err := trackingService.StopTracking(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "追踪已停止", nil)
}

func GetTrackingTask(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的任务ID", nil)
		return
	}

	task, err := trackingService.GetTask(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "任务不存在", nil)
		return
	}

	utils.SuccessResponse(c, "ok", task)
}

func ListTrackingTasks(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)

	var uavID *uint64
	if uavStr := c.Query("uav_id"); uavStr != "" {
		if id, err := strconv.ParseUint(uavStr, 10, 64); err == nil {
			uavID = &id
		}
	}

	var status *models.TrackingStatus
	if s := c.Query("status"); s != "" {
		ts := models.TrackingStatus(s)
		status = &ts
	}

	tasks, total, err := trackingService.ListTasks(uavID, status, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "ok", gin.H{
		"list":  tasks,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func GetActiveTracking(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	task, err := trackingService.GetActiveTask(uavID)
	if err != nil {
		utils.SuccessResponse(c, "ok", nil)
		return
	}

	utils.SuccessResponse(c, "ok", task)
}

func ListDetections(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}
	page, pageSize := utils.ParsePagination(c)

	targets, total, err := trackingService.ListDetections(uavID, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "ok", gin.H{
		"list":  targets,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

func DetectImage(c *gin.Context) {
	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的无人机ID", nil)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "请上传图片", nil)
		return
	}

	f, err := file.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, "读取图片失败", nil)
		return
	}
	defer f.Close()

	buf := make([]byte, file.Size)
	if _, err := f.Read(buf); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400004, "读取图片失败", nil)
		return
	}

	frameWidth := 1280
	frameHeight := 720
	if w := c.Query("width"); w != "" {
		if v, e := strconv.Atoi(w); e == nil {
			frameWidth = v
		}
	}
	if h := c.Query("height"); h != "" {
		if v, e := strconv.Atoi(h); e == nil {
			frameHeight = v
		}
	}

	targets, err := trackingService.DetectAndTrack(uavID, buf, frameWidth, frameHeight)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "ok", targets)
}
