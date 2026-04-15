package main

import (
	"code-exec/config"
	"code-exec/internal/ai"
	"code-exec/internal/api"
	"code-exec/internal/auth"
	"code-exec/internal/docker"
	"code-exec/internal/logbuf"
	"code-exec/internal/model"
	"code-exec/internal/queue"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载 .env 文件 (如果存在)
	_ = godotenv.Load()

	// 加载配置
	cfg := config.Load()

	// 初始化日志环形缓冲区（500 条），挂载到标准 log
	logBuf := logbuf.NewRingBuffer(500)
	log.SetOutput(logbuf.NewTeeWriter(logBuf))
	log.SetFlags(log.Ldate | log.Ltime)

	// 初始化JWT认证模块
	auth.Init(cfg.JWTSecret)

	// 加载语言配置
	if err := config.LoadLanguages("languages.yaml"); err != nil {
		log.Fatalf("加载语言配置失败: %v", err)
	}

	// 连接 MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}

	// 配置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间
	log.Printf("数据库连接池已配置: MaxIdle=%d, MaxOpen=%d", cfg.DBMaxIdleConns, cfg.DBMaxOpenConns)

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.Submission{}, &model.Language{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("数据库已连接并完成迁移。")

	// 初始化语言种子数据
	seedLanguages(db)

	// 初始化管理员账号
	seedAdmin(db, cfg)

	// 设置路由
	r := gin.Default()

	// 设置 CORS 配置 (供 WebSocket 使用)
	api.SetCORSConfig(cfg.AllowedOrigins)

	// 使用 CORS 中间件
	r.Use(api.CORSMiddleware(cfg.AllowedOrigins))

	// Docker 组件 (需要在 API 注册之前初始化)
	sandbox, err := docker.NewSandbox()
	if err != nil {
		log.Fatalf("初始化沙箱失败: %v", err)
	}
	pool := docker.NewPool(sandbox, cfg)

	// 启动时清理残留容器
	ctx := context.Background()
	if removed, err := pool.CleanupStaleContainers(ctx); err != nil {
		log.Printf("[警告] 清理残留容器失败: %v", err)
	} else if removed > 0 {
		log.Printf("[启动] 已清理 %d 个上次运行遗留的容器", removed)
	}

	// 启动弹性调度水位监控协程（清理残留容器后再启动）
	pool.StartMonitor(ctx)

	// API 组
	v1 := r.Group("/api")
	{
		api.RegisterRoutes(v1, db, pool, cfg.RedisAddr)
		aiHandler := ai.RegisterRoutes(v1, cfg) // 注册 AI 分析路由，获取 handler 实例

		// 注册管理员路由（传入 LLM 客户端和日志缓冲以支持管理功能）
		api.RegisterAdminRoutes(v1, db, pool, aiHandler.Client, logBuf)
	}

	// 启动队列消费者
	consumerWorker := queue.NewConsumer(cfg.RedisAddr, pool, db)
	go consumerWorker.StartWorker()
	//log.Println("队列消费者已启动",consumerWorker)
	// 2026/04/13 17:43:48 队列消费者已启动 
	// &{0xc00047ac40 code_execution_queue 0xc000243200 0xc0001ff3e0 127.0.0.1:16379}
	// redis客户端指针+队列名称+docker池指针+数据库指针+redis地址

	// 停机
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("服务器正在启动，端口 %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 给予 5 秒的缓慢停机时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止容器池监控协程
	pool.StopMonitor()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	}

	// 关闭数据库连接
	sqlDB.Close()
	log.Println("服务器已停止")
}

// seedLanguages 初始化语言种子数据
func seedLanguages(db *gorm.DB) {
	languages := []model.Language{
		// {Value: "cpp", Label: "C++ (GCC Latest)", Icon: "C++.png", DisplayOrder: 1, Enabled: true},
		{Value: "cpp", Label: "C++", Icon: "C++.webp", DisplayOrder: 1, Enabled: true},

		// {Value: "c", Label: "C (GCC Latest)", Icon: "C.png", DisplayOrder: 2, Enabled: true},
		{Value: "c", Label: "C", Icon: "C.webp", DisplayOrder: 2, Enabled: true},

		// {Value: "java", Label: "Java 17 (Temurin)", Icon: "java.png", DisplayOrder: 3, Enabled: true},
		{Value: "java", Label: "Java", Icon: "java.webp", DisplayOrder: 3, Enabled: true},

		// {Value: "python", Label: "Python 3.12 (Alpine)", Icon: "Python.png", DisplayOrder: 4, Enabled: true},
		{Value: "python", Label: "Python", Icon: "Python.webp", DisplayOrder: 4, Enabled: true},

		// {Value: "go", Label: "Go (Latest)", Icon: "go.png", DisplayOrder: 6, Enabled: true},
		{Value: "go", Label: "Go", Icon: "go.webp", DisplayOrder: 6, Enabled: true},

		// {Value: "javascript", Label: "Node.js 18", Icon: "node_js.png", DisplayOrder: 7, Enabled: true},
		{Value: "javascript", Label: "JavaScript", Icon: "node_js.webp", DisplayOrder: 7, Enabled: true},

		// {Value: "rust", Label: "Rust (latest)", Icon: "Rust.png", DisplayOrder: 8, Enabled: true},
		{Value: "rust", Label: "Rust", Icon: "Rust.webp", DisplayOrder: 8, Enabled: true},

		// {Value: "csharp", Label: "C# (.NET 8)", Icon: "csharp.png", DisplayOrder: 9, Enabled: true},
		{Value: "csharp", Label: "C#", Icon: "csharp.webp", DisplayOrder: 9, Enabled: true},

		{Value: "typescript", Label: "TypeScript", Icon: "typescript.webp", DisplayOrder: 10, Enabled: true},
	}

	for _, lang := range languages {
		var existing model.Language
		if err := db.Where("value = ?", lang.Value).First(&existing).Error; err == nil {
			// 如果存在，更新 Label 和 Order
			db.Model(&existing).Updates(map[string]interface{}{
				"label":         lang.Label,
				"display_order": lang.DisplayOrder,
			})
		} else {
			// 不存在，创建
			db.Create(&lang)
			log.Printf("Created language: %s", lang.Value)
		}
	}

	// 移除已废弃的语言 (如 pypy)
	if err := db.Where("value = ?", "pypy").Delete(&model.Language{}).Error; err == nil {
		log.Println("Removed deprecated language: pypy")
	}
}

// seedAdmin 初始化管理员账号
// 使用配置文件中的 AdminUsername 和 AdminPassword
func seedAdmin(db *gorm.DB, cfg *config.Config) {
	// 生成管理员密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[Admin] 密码哈希生成失败: %v", err)
		return
	}

	var existing model.User
	if err := db.Where("username = ?", cfg.AdminUsername).First(&existing).Error; err == nil {
		// 管理员已存在，保持现有密码不变，仅确保角色正确
		if existing.Role != "admin" {
			db.Model(&existing).Update("role", "admin")
			log.Println("[Admin] 已将", cfg.AdminUsername, "用户角色更新为管理员")
		}
		return
	}

	// 创建管理员账号
	admin := model.User{
		Username: cfg.AdminUsername,
		Password: string(hashedPassword),
		Role:     "admin",
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Printf("[Admin] 创建管理员账号失败: %v", err)
		return
	}
	log.Println("[Admin] 已创建管理员账号，请立即登录并修改默认密码！")
}
