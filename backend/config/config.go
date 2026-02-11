package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	MySQLDSN   string
	RedisAddr  string
	ServerPort string
	DockerHost string
	JWTSecret  string

	// 数据库连接池配置
	DBMaxIdleConns int
	DBMaxOpenConns int

	// 容器池配置
	ContainerPoolSize int
	ContainerMemoryMB int     // 容器内存限制 (MB)
	ContainerCPUCores float64 // 容器 CPU 核心数

	// CORS 配置
	AllowedOrigins []string

	// LLM 配置
	LLMAPIKey string
	LLMAPIURL string
	LLMModel  string

	// 管理员账号配置
	AdminUsername string
	AdminPassword string
}

func Load() *Config {
	// 默认配置
	cfg := &Config{
		MySQLDSN:   "root:rootpassword@tcp(127.0.0.1:13306)/code_exec?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:  "127.0.0.1:16379",
		ServerPort: ":8080",
		DockerHost: "unix:///var/run/docker.sock", // Linux 默认
		JWTSecret:  "your-secret-key",

		// 数据库连接池默认值
		DBMaxIdleConns: 10,
		DBMaxOpenConns: 100,

		// 容器池默认值
		ContainerPoolSize: 3,
		ContainerMemoryMB: 2048, // 默认 2GB
		ContainerCPUCores: 2.0,  // 默认 2 核

		// CORS 默认允许的源（开发环境）
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
		},

		// LLM 默认配置 (Zhipu AI)
		// glm-4-flashx 比 glm-4-flash 更快，适合代码分析场景
		LLMAPIKey: "", // 需要通过环境变量设置
		LLMAPIURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		LLMModel:  "codegeex-4",
		//LLMModel:  "glm-4-flashx",

		// 管理员账号 (生产环境必须修改!)
		AdminUsername: "admin",
		AdminPassword: "admin123",
	}

	// 如果需要，使用简单的环境变量覆盖
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		port := "13306"
		if p := os.Getenv("MYSQL_PORT"); p != "" {
			port = p
		}
		cfg.MySQLDSN = fmt.Sprintf("root:rootpassword@tcp(%s:%s)/code_exec?charset=utf8mb4&parseTime=True&loc=Local", host, port)
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}

	// 容器池大小
	if poolSize := os.Getenv("CONTAINER_POOL_SIZE"); poolSize != "" {
		if size, err := strconv.Atoi(poolSize); err == nil && size > 0 {
			cfg.ContainerPoolSize = size
		}
	}

	// 容器内存限制 (MB)
	if memMB := os.Getenv("CONTAINER_MEMORY_MB"); memMB != "" {
		if mem, err := strconv.Atoi(memMB); err == nil && mem > 0 {
			cfg.ContainerMemoryMB = mem
		}
	}

	// 容器 CPU 核心数
	if cpuCores := os.Getenv("CONTAINER_CPU_CORES"); cpuCores != "" {
		if cpu, err := strconv.ParseFloat(cpuCores, 64); err == nil && cpu > 0 {
			cfg.ContainerCPUCores = cpu
		}
	}

	// CORS 允许的源
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		// 简单按逗号分隔
		cfg.AllowedOrigins = splitAndTrim(origins)
	}

	// Windows 下，DockerHost 通常是 npipe:////./pipe/docker_engine
	// 检查操作系统
	if os.Getenv("OS") == "Windows_NT" {
		cfg.DockerHost = "npipe:////./pipe/docker_engine"
	}

	// LLM 配置加载
	if key := os.Getenv("LLM_API_KEY"); key != "" {
		cfg.LLMAPIKey = key
	}
	if url := os.Getenv("LLM_API_URL"); url != "" {
		cfg.LLMAPIURL = url
	}
	if model := os.Getenv("LLM_MODEL"); model != "" {
		cfg.LLMModel = model
	}

	// 管理员账号配置
	if adminUser := os.Getenv("ADMIN_USERNAME"); adminUser != "" {
		cfg.AdminUsername = adminUser
	}
	if adminPass := os.Getenv("ADMIN_PASSWORD"); adminPass != "" {
		cfg.AdminPassword = adminPass
	}

	return cfg
}

// splitAndTrim 按逗号分隔并去除空格
// 使用标准库实现
func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
