package api

import (
	"code-exec/internal/model"
	"code-exec/internal/queue"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	authHandler := NewAuthHandler(db)

	// 健康检查
	rg.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 认证路由（公开）
	rg.POST("/register", authHandler.Register)
	rg.POST("/login", authHandler.Login)
	rg.GET("/verify", authHandler.Verify)

	// WebSocket（目前公开，可以加保护）
	rg.GET("/ws", HandleWebSocket)

	// 语言相关路由（公开） - 获取支持的编程语言列表
	rg.GET("/languages", func(c *gin.Context) {
		var languages []model.Language
		if err := db.Where("enabled = ?", true).Order("display_order ASC").Find(&languages).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch languages"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"languages": languages})
	})

	// 代码执行路由（受保护）
	// 初始化生产者（理想情况下应当是单例，但为了简单起见在此处初始化）
	producer := queue.NewProducer("code_exec_redis:6379") // 使用 Docker 网络地址

	// 受保护的路由组
	protected := rg.Group("/")
	protected.Use(AuthMiddleware())
	{
		protected.POST("/run", func(c *gin.Context) {
			var input struct {
				Language string `json:"language" binding:"required"`
				Code     string `json:"code" binding:"required"`
				Input    string `json:"input"` // 标准输入数据（可选）
			}

			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			userID := GetUserID(c)

			// 生成任务 ID
			taskID := "task-" + fmt.Sprintf("%d", time.Now().UnixNano())

			// 创建提交记录
			submission := model.Submission{
				UserID:   userID,
				Language: input.Language,
				Code:     input.Code,
				Input:    input.Input,
				Status:   "Pending",
			}
			if err := db.Create(&submission).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
				return
			}

			// 推送到队列
			err := producer.AddTask(queue.Task{
				ID:           taskID,
				Language:     input.Language,
				Code:         input.Code,
				Input:        input.Input,
				UserID:       userID,
				SubmissionID: submission.ID,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue task"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "queued", "taskId": taskID})
		})

		// 获取用户的提交历史
		protected.GET("/submissions", func(c *gin.Context) {
			userID := GetUserID(c)

			var submissions []model.Submission
			if err := db.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&submissions).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
				return
			}

			// 转换为响应格式
			type SubmissionResponse struct {
				ID        uint      `json:"id"`
				Language  string    `json:"language"`
				Code      string    `json:"code"`
				Status    string    `json:"status"`
				Output    string    `json:"output"`
				CreatedAt time.Time `json:"createdAt"`
			}

			response := make([]SubmissionResponse, len(submissions))
			for i, s := range submissions {
				response[i] = SubmissionResponse{
					ID:        s.ID,
					Language:  s.Language,
					Code:      s.Code,
					Status:    s.Status,
					Output:    s.Output,
					CreatedAt: s.CreatedAt,
				}
			}

			c.JSON(http.StatusOK, gin.H{"submissions": response})
		})
	}
}
