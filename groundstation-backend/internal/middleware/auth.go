package middleware

import (
	"net/http"
	"strings"
	"groundstation-backend/internal/models"
	"groundstation-backend/internal/service"
	"groundstation-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

var authService = service.NewAuthService()

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, 401001, "未提供认证令牌", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.ErrorResponse(c, http.StatusUnauthorized, 401002, "认证令牌格式错误", nil)
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, 401003, "认证令牌无效或已过期", nil)
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleAuth(allowedRoles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(c, http.StatusForbidden, 403001, "用户角色信息不存在", nil)
			c.Abort()
			return
		}

		userRole := role.(models.UserRole)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(c, http.StatusForbidden, 403002, "权限不足", nil)
		c.Abort()
	}
}

func GetCurrentUserID(c *gin.Context) uint64 {
	if userID, exists := c.Get("userID"); exists {
		return userID.(uint64)
	}
	return 0
}

func GetCurrentUserRole(c *gin.Context) models.UserRole {
	if role, exists := c.Get("role"); exists {
		return role.(models.UserRole)
	}
	return models.UserRoleUser
}
