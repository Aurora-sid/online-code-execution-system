package api

import (
	"net/http"
	"fmt"
	"time"
	"code-exec/internal/queue"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	authHandler := NewAuthHandler(db)

	// Health Check
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Auth Routes
	rg.POST("/register", authHandler.Register)
	rg.POST("/login", authHandler.Login)

	// WebSocket
	rg.GET("/ws", HandleWebSocket)

	// Code Execution Routes
	// Init producer (should be singleton ideally, but here for simplicity)
	producer := queue.NewProducer("code_exec_redis:6379") // Use Docker network address

	rg.POST("/run", func(c *gin.Context) {
		var input struct {
			Language string `json:"language" binding:"required"`
			Code     string `json:"code" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate Task ID
		taskID := "task-" + fmt.Sprintf("%d", time.Now().UnixNano())

		// Push to Queue
		err := producer.AddTask(queue.Task{
			ID:       taskID,
			Language: input.Language,
			Code:     input.Code,
			UserID:   1, // Mock UserID for now
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue task"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "queued", "taskId": taskID})
	})
}
