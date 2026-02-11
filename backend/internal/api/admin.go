package api

import (
	"code-exec/internal/docker"
	"code-exec/internal/model"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminHandler 管理员 API 处理器
type AdminHandler struct {
	DB   *gorm.DB
	Pool *docker.Pool
}

// NewAdminHandler 创建管理员处理器实例
func NewAdminHandler(db *gorm.DB, pool *docker.Pool) *AdminHandler {
	return &AdminHandler{DB: db, Pool: pool}
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// StatsResponse 系统统计响应结构
type StatsResponse struct {
	OnlineUsers       int               `json:"onlineUsers"`
	TotalUsers        int64             `json:"totalUsers"`
	TodaySubmissions  int64             `json:"todaySubmissions"`
	FailedSubmissions int64             `json:"failedSubmissions"`
	SuccessRate       float64           `json:"successRate"`
	PoolStats         *docker.PoolStats `json:"poolStats,omitempty"`
}

// GetStats 获取系统统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	var stats StatsResponse

	// 总用户数
	h.DB.Model(&model.User{}).Count(&stats.TotalUsers)

	// 今日提交数
	today := time.Now().Truncate(24 * time.Hour)
	h.DB.Model(&model.Submission{}).Where("created_at >= ?", today).Count(&stats.TodaySubmissions)

	// 失败/超时提交数
	h.DB.Model(&model.Submission{}).Where("status IN ?", []string{"Failed", "Timeout"}).Count(&stats.FailedSubmissions)

	// 成功率计算
	var totalSubmissions int64
	h.DB.Model(&model.Submission{}).Count(&totalSubmissions)
	if totalSubmissions > 0 {
		successCount := totalSubmissions - stats.FailedSubmissions
		stats.SuccessRate = float64(successCount) / float64(totalSubmissions) * 100
	}

	// 在线用户数 (从 WebSocket 连接池获取，暂时返回占位值)
	stats.OnlineUsers = GetOnlineUserCount()

	// 获取容器池统计
	if h.Pool != nil {
		stats.PoolStats = h.Pool.GetStats()
	}

	c.JSON(http.StatusOK, stats)
}

// UserResponse 用户列表响应结构
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IsOnline  bool      `json:"isOnline"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListUsers 获取用户列表
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	h.DB.Model(&model.User{}).Count(&total)
	h.DB.Order("id ASC").Limit(pageSize).Offset(offset).Find(&users) // 按 ID 升序排序

	response := make([]UserResponse, len(users))
	for i, u := range users {
		response[i] = UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			IsOnline:  IsUserOnline(u.ID),
			CreatedAt: u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": response,
		"total": total,
		"page":  page,
	})
}

// CreateUserInput 创建用户请求结构
type CreateUserInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

// CreateUser 管理员创建用户
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认角色
	if input.Role == "" {
		input.Role = "user"
	}

	// 验证角色值
	if input.Role != "user" && input.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色，必须是 'user' 或 'admin'"})
		return
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := model.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Role:     input.Role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败，用户名可能已存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户创建成功",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// DeleteUser 删除用户
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// 检查是否试图删除自己
	currentUserID := GetUserID(c)
	targetUserID, _ := strconv.ParseUint(id, 10, 32)
	if uint(targetUserID) == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		return
	}

	// 删除用户及其提交记录
	tx := h.DB.Begin()

	if err := tx.Where("user_id = ?", id).Delete(&model.Submission{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user submissions"})
		return
	}

	if err := tx.Delete(&model.User{}, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// SubmissionDetail 提交详情响应结构
type SubmissionDetail struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	Username  string    `json:"username"`
	Language  string    `json:"language"`
	Code      string    `json:"code"`
	Input     string    `json:"input"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListSubmissions 获取所有提交记录（管理员）
func (h *AdminHandler) ListSubmissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	userID := c.Query("userId")
	offset := (page - 1) * pageSize

	query := h.DB.Model(&model.Submission{})

	// 筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Count(&total)

	var submissions []model.Submission
	query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&submissions)

	// 获取用户信息
	userIDs := make([]uint, len(submissions))
	for i, s := range submissions {
		userIDs[i] = s.UserID
	}
	var users []model.User
	h.DB.Where("id IN ?", userIDs).Find(&users)
	userMap := make(map[uint]string)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 组装响应
	response := make([]SubmissionDetail, len(submissions))
	for i, s := range submissions {
		response[i] = SubmissionDetail{
			ID:        s.ID,
			UserID:    s.UserID,
			Username:  userMap[s.UserID],
			Language:  s.Language,
			Code:      s.Code,
			Input:     s.Input,
			Status:    s.Status,
			Output:    s.Output,
			CreatedAt: s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"submissions": response,
		"total":       total,
		"page":        page,
	})
}

// WeeklyStatsResponse 周提交统计响应
type WeeklyStatsResponse struct {
	Labels []string `json:"labels"`
	Data   []int64  `json:"data"`
}

// GetWeeklySubmissionStats 获取本周每日提交量统计
func (h *AdminHandler) GetWeeklySubmissionStats(c *gin.Context) {
	now := time.Now()

	// 计算本周一 00:00:00
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -int(weekday-time.Monday))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())

	labels := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	data := make([]int64, 7)

	for i := 0; i < 7; i++ {
		dayStart := monday.AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)

		var count int64
		h.DB.Model(&model.Submission{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&count)
		data[i] = count
	}

	c.JSON(http.StatusOK, WeeklyStatsResponse{
		Labels: labels,
		Data:   data,
	})
}

// RegisterAdminRoutes 注册管理员路由
func RegisterAdminRoutes(rg *gin.RouterGroup, db *gorm.DB, pool *docker.Pool) {
	adminHandler := NewAdminHandler(db, pool)

	admin := rg.Group("/admin")
	admin.Use(AuthMiddleware(), AdminMiddleware())
	{
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/stats/weekly", adminHandler.GetWeeklySubmissionStats)
		admin.GET("/users", adminHandler.ListUsers)
		admin.POST("/users", adminHandler.CreateUser)
		admin.DELETE("/users/:id", adminHandler.DeleteUser)
		admin.GET("/submissions", adminHandler.ListSubmissions)

		// 容器池管理
		admin.POST("/pool/reset", adminHandler.ResetPool)
	}
}

// ResetPool 重置容器池
func (h *AdminHandler) ResetPool(c *gin.Context) {
	if h.Pool == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Container pool not available"})
		return
	}

	ctx := context.Background()
	removed, err := h.Pool.ResetPool(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "容器池已重置",
		"removed": removed,
	})
}
