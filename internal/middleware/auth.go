package middleware

import (
	"net/http"
	"strings"

	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserAuth 用户API Key认证中间件
func UserAuth(userAPIKeyRepo repository.UserAPIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取API Key
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// 支持 "Bearer <api_key>" 格式
		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key format"})
			c.Abort()
			return
		}

		// 验证API Key
		userAPIKey, err := userAPIKeyRepo.FindByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			logger.Warn("Invalid API key", zap.String("api_key", apiKey[:8]+"..."), zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// 检查Key是否激活
		if !userAPIKey.IsActive {
			logger.Warn("Inactive API key", zap.String("api_key_id", userAPIKey.ID))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is inactive"})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("user_id", userAPIKey.UserID)
		c.Set("api_key_id", userAPIKey.ID)

		logger.Debug("User authenticated", zap.String("user_id", userAPIKey.UserID))
		c.Next()
	}
}
