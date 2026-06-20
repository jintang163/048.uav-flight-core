package handler

import (
	"net/http"
	"groundstation-backend/internal/middleware"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

var authService = service.NewAuthService()

type RegisterRequest struct {
	Username string           `json:"username" binding:"required,min=3,max=50"`
	Password string           `json:"password" binding:"required,min=6,max=50"`
	Email    string           `json:"email" binding:"omitempty,email"`
	Phone    string           `json:"phone" binding:"omitempty"`
	FullName string           `json:"full_name" binding:"omitempty"`
	Role     models.UserRole  `json:"role" binding:"omitempty,oneof=ADMIN OPERATOR USER"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
		FullName: req.FullName,
		Role:     req.Role,
	}
	if user.Role == "" {
		user.Role = models.UserRoleUser
	}

	result, err := authService.Register(user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "注册成功", result)
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	reqSvc := &service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	result, err := authService.Login(reqSvc, ip, userAgent)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, 401001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "登录成功", result)
}

func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	result, err := authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, 401002, "刷新令牌无效", nil)
		return
	}

	utils.SuccessResponse(c, "令牌刷新成功", result)
}

func Logout(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	_ = authService.Logout(userID)
	utils.SuccessResponse(c, "登出成功", nil)
}

func GetCurrentUser(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	user, err := authService.GetUserByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "用户不存在", nil)
		return
	}
	utils.SuccessResponse(c, "获取成功", user)
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "参数错误: "+err.Error(), nil)
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "密码修改成功", nil)
}

func ListUsers(c *gin.Context) {
	pagination := utils.GeneratePaginationFromRequest(c)
	users, total, err := authService.ListUsers(pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponseWithTotal(c, "获取成功", users, total)
}

func GetUser(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的用户ID", nil)
		return
	}

	user, err := authService.GetUserByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, 404001, "用户不存在", nil)
		return
	}

	utils.SuccessResponse(c, "获取成功", user)
}

func UpdateUser(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的用户ID", nil)
		return
	}

	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400002, "参数错误: "+err.Error(), nil)
		return
	}

	user.ID = id
	result, err := authService.UpdateUser(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "更新成功", result)
}

func DeleteUser(c *gin.Context) {
	id, err := utils.ParseUint64(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, 400001, "无效的用户ID", nil)
		return
	}

	if err := authService.DeleteUser(id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, 500001, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, "删除成功", nil)
}
