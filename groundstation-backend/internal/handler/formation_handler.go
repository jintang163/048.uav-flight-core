package handler

import (
	"net/http"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var formationService = service.NewFormationService()

type CreateFormationRequest struct {
	Name        string              `json:"name" binding:"required"`
	Type        models.FormationType `json:"type" binding:"required"`
	Spacing     float64             `json:"spacing"`
	Description string              `json:"description"`
	UAVIDs      []uint64            `json:"uav_ids"`
	LeaderID    uint64              `json:"leader_id"`
}

type UpdateFormationRequest struct {
	Name        string              `json:"name"`
	Type        models.FormationType `json:"type"`
	Spacing     float64             `json:"spacing"`
	Description string              `json:"description"`
}

type AddMemberRequest struct {
	UAVID uint64 `json:"uav_id" binding:"required"`
}

type LightConfigRequest struct {
	Red    uint8            `json:"red"`
	Green  uint8            `json:"green"`
	Blue   uint8            `json:"blue"`
	Effect models.LightEffect `json:"effect"`
}

func CreateFormation(c *gin.Context) {
	var req CreateFormationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.CreateFormationRequest{
		Name:        req.Name,
		Type:        req.Type,
		Spacing:     req.Spacing,
		Description: req.Description,
		UAVIDs:      req.UAVIDs,
		LeaderID:    req.LeaderID,
	}

	result, err := formationService.Create(serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "创建成功", result)
}

func GetFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	formation, err := formationService.GetByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "编队不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", formation)
}

func ListFormations(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)

	formations, total, err := formationService.List(pagination, 0)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", formations, total)
}

func UpdateFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	var req UpdateFormationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	serviceReq := &service.UpdateFormationRequest{
		Name:        req.Name,
		Type:        req.Type,
		Spacing:     req.Spacing,
		Description: req.Description,
	}

	result, err := formationService.Update(id, serviceReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	if err := formationService.Delete(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}

func AddFormationMember(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := formationService.AddMember(id, &service.AddMemberRequest{UAVID: req.UAVID}); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "添加成功", nil)
}

func RemoveFormationMember(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "无效的无人机ID", nil)
		return
	}

	if err := formationService.RemoveMember(id, uavID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "移除成功", nil)
}

func SetFormationLeader(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	uavID, err := utils.ParseUint64(c.Param("uav_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "无效的无人机ID", nil)
		return
	}

	if err := formationService.SetLeader(id, uavID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "设置成功", nil)
}

func GetFormationMembers(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	members, err := formationService.GetMembers(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", members)
}

func StartFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	if err := formationService.Start(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "编队已启动", nil)
}

func PauseFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	if err := formationService.Pause(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "编队已暂停", nil)
}

func ResumeFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	if err := formationService.Resume(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "编队已恢复", nil)
}

func StopFormation(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	if err := formationService.Stop(id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "编队已停止", nil)
}

func GetActiveFormations(c *gin.Context) {
	formations, err := formationService.GetActiveFormations()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", formations)
}

func GetCollisionWarnings(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	pagination := utils.GeneratePaginationFromRequest(c)

	warnings, total, err := formationService.GetCollisionWarnings(id, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", warnings, total)
}

func SetFormationLight(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	var req LightConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := formationService.SetLightConfig(id, &service.LightConfigRequest{
		Red:    req.Red,
		Green:  req.Green,
		Blue:   req.Blue,
		Effect: req.Effect,
	}); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "灯光配置已下发", nil)
}

func SyncFormationWaypoints(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	var req struct {
		MissionID uint64 `json:"mission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	if err := formationService.SyncWaypoints(id, req.MissionID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400003, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "航点同步已下发", nil)
}

func MultiTakeoff(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的编队ID", nil)
		return
	}

	var req struct {
		Altitude float64 `json:"altitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Altitude = 5.0
	}
	if req.Altitude <= 0 {
		req.Altitude = 5.0
	}

	if err := formationService.MultiTakeoff(id, req.Altitude); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "多机起飞指令已下发", nil)
}
