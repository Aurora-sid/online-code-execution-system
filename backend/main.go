package main

import (
	"code-exec/config"
	"code-exec/internal/api"
	"code-exec/internal/docker"
	"code-exec/internal/model"
	"code-exec/internal/queue"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 加载语言配置
	if err := config.LoadLanguages("languages.yaml"); err != nil {
		log.Fatalf("加载语言配置失败: %v", err)
	}

	// 连接 MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&model.User{}, &model.Submission{}, &model.Language{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	fmt.Println("数据库已连接并完成迁移。")

	// 初始化语言种子数据
	seedLanguages(db)

	// 设置路由
	r := gin.Default()

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API 组
	v1 := r.Group("/api")
	{
		api.RegisterRoutes(v1, db)
	}

	// Docker 组件
	sandbox, err := docker.NewSandbox()
	if err != nil {
		log.Fatalf("初始化沙箱失败: %v", err)
	}
	pool := docker.NewPool(sandbox)

	// 启动队列消费者
	consumerWorker := queue.NewConsumer(cfg.RedisAddr, pool, db)
	go consumerWorker.StartWorker()

	// 启动服务器
	log.Printf("服务器正在启动，端口 %s", cfg.ServerPort)

	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}

// seedLanguages 初始化语言种子数据
func seedLanguages(db *gorm.DB) {
	languages := []model.Language{
		// {Value: "cpp", Label: "C++ (GCC Latest)", Icon: "C++.png", DisplayOrder: 1, Enabled: true},
		{Value: "cpp", Label: "C++", Icon: "C++.png", DisplayOrder: 1, Enabled: true},

		// {Value: "c", Label: "C (GCC Latest)", Icon: "C.png", DisplayOrder: 2, Enabled: true},
		{Value: "c", Label: "C", Icon: "C.png", DisplayOrder: 2, Enabled: true},

		// {Value: "java", Label: "Java 17 (Temurin)", Icon: "java.png", DisplayOrder: 3, Enabled: true},
		{Value: "java", Label: "Java", Icon: "java.png", DisplayOrder: 3, Enabled: true},

		// {Value: "python", Label: "Python 3.12 (Alpine)", Icon: "Python.png", DisplayOrder: 4, Enabled: true},
		{Value: "python", Label: "Python", Icon: "Python.png", DisplayOrder: 4, Enabled: true},

		// {Value: "pypy", Label: "PyPy 3.7 (7.3.8)", Icon: "pypy.png", DisplayOrder: 5, Enabled: true},
		{Value: "pypy", Label: "PyPy", Icon: "pypy.png", DisplayOrder: 5, Enabled: true},

		// {Value: "go", Label: "Go (Latest)", Icon: "go.png", DisplayOrder: 6, Enabled: true},
		{Value: "go", Label: "Go", Icon: "go.png", DisplayOrder: 6, Enabled: true},

		// {Value: "javascript", Label: "Node.js 18", Icon: "node_js.png", DisplayOrder: 7, Enabled: true},
		{Value: "javascript", Label: "JavaScript", Icon: "node_js.png", DisplayOrder: 7, Enabled: true},
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
}
