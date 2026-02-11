package api

import (
	"code-exec/internal/auth"
	"code-exec/internal/model"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CORS 配置 (WebSocket 也需要使用)
var (
	allowedOrigins  map[string]bool
	allowAllOrigins bool
	allowedMu       sync.RWMutex
)

// SetCORSConfig 设置 CORS 配置 (供 main.go 和 websocket.go 使用)
func SetCORSConfig(origins []string) {
	allowedMu.Lock()
	defer allowedMu.Unlock()
	allowedOrigins = make(map[string]bool)
	for _, origin := range origins {
		if origin == "*" {
			allowAllOrigins = true
		}
		allowedOrigins[origin] = true
	}
}

// IsOriginAllowed 检查来源是否被允许 (供 WebSocket 使用)
func IsOriginAllowed(origin string) bool {
	allowedMu.RLock()
	defer allowedMu.RUnlock()
	if allowAllOrigins {
		return true
	}
	return allowedOrigins[origin] || origin == ""
}

// CORSMiddleware 返回 CORS 中间件
func CORSMiddleware(origins []string) gin.HandlerFunc {
	originMap := make(map[string]bool)
	allowAll := false
	for _, origin := range origins {
		if origin == "*" {
			allowAll = true
		}
		originMap[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 如果配置了 * 则允许所有来源
		if allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if originMap[origin] || origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(origins) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origins[0])
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 在线用户追踪
var (
	onlineUsers   = make(map[uint]bool)
	onlineUsersMu sync.RWMutex
)

// TrackUserOnline 标记用户上线
func TrackUserOnline(userID uint) {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	onlineUsers[userID] = true
}

// TrackUserOffline 标记用户下线
func TrackUserOffline(userID uint) {
	onlineUsersMu.Lock()
	defer onlineUsersMu.Unlock()
	delete(onlineUsers, userID)
}

// GetOnlineUserCount 获取在线用户数
func GetOnlineUserCount() int {
	onlineUsersMu.RLock()
	defer onlineUsersMu.RUnlock()
	return len(onlineUsers)
}

// IsUserOnline 检查指定用户是否在线
func IsUserOnline(userID uint) bool {
	onlineUsersMu.RLock()
	defer onlineUsersMu.RUnlock()
	return onlineUsers[userID]
}

// db 引用（需要在 RegisterRoutes 时设置）
var dbInstance *gorm.DB

// SetDB 设置数据库实例
func SetDB(db *gorm.DB) {
	dbInstance = db
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 从 "Bearer <token>" 中提取令牌
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 将用户 ID 设置到上下文中以供后续使用
		c.Set("userID", claims.UserID)

		// 追踪用户上线状态
		TrackUserOnline(claims.UserID)

		// 获取用户角色
		if dbInstance != nil {
			var user model.User
			if err := dbInstance.First(&user, claims.UserID).Error; err == nil {
				c.Set("userRole", user.Role)
			}
		}

		c.Next()
	}
}

// GetUserID 从上下文中提取用户 ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	return userID.(uint)
}
