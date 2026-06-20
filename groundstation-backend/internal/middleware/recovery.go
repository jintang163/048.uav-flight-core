package middleware

import (
	"net/http"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)
				utils.ErrorResponse(c, http.StatusInternalServerError, 500001, "服务器内部错误", nil)
				c.Abort()
			}
		}()
		c.Next()
	}
}
