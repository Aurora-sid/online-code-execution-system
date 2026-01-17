package main

import (
	"code-exec/config"
	"code-exec/internal/api"
	"code-exec/internal/model"
	"code-exec/internal/docker"
	"code-exec/internal/queue"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load Config
	cfg := config.Load()

	// Connect to MySQL
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	// Auto Migrate
	err = db.AutoMigrate(&model.User{}, &model.Submission{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Database connected and migrated.")


	// Setup Router
	r := gin.Default()
	
	// CORS Middleware
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
	
	// API Group
	v1 := r.Group("/api")
	{
		api.RegisterRoutes(v1, db)
	}

	// Docker Components
	sandbox, err := docker.NewSandbox()
	if err != nil {
		log.Fatalf("Failed to init sandbox: %v", err)
	}
	pool := docker.NewPool(sandbox)
	
	// Start Queue Consumer
	consumerWorker := queue.NewConsumer(cfg.RedisAddr, pool)
	go consumerWorker.StartWorker()

	// Start Server
	log.Printf("Server starting on %s", cfg.ServerPort)

	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
