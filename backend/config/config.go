/*
定义系统的全局配置结构体，
集中管理了系统在运行期间所需要的所有配置参数，
例如数据库连接配置、Redis 配置、容器资源限制（内存/CPU）、弹性调度水位线、
大语言模型 (LLM) 接口和管理员固定配置等。同时提供了一个统一的加载入口，
支持默认值和环境变量覆盖。
*/
package config

import (
	"fmt"     // 用于格式化字符串
	"os"      // 用于读取环境变量
	"strconv" // 用于字符串与数字的转换
	"strings"
)

type Config struct {
	MySQLDSN   string // MySQL 数据库连接字符串
	RedisAddr  string // 服务器监听地址
	ServerPort string // 服务器监听端口
	DockerHost string // Docker 守护进程地址
	JWTSecret  string // JWT 密钥

	// 数据库连接池配置
	DBMaxIdleConns int
	DBMaxOpenConns int

	// 容器资源配置
	ContainerMemoryMB int     // 容器内存限制 (MB)
	ContainerCPUCores float64 // 容器 CPU 核心数

	// 弹性容器调度配置
	ContainerLWM    int // 废弃，被动态下限替代
	ContainerHWM    int // High Water Mark: 每种语言最多缓存的闲置容器数
	MonitorInterval int // 监控协程检查间隔（秒）

	// 动态水位(EMA)配置
	EMAAlpha float64
	LWMMin   int
	LWMMax   int

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
		// 默认值可以根据被环境变量覆盖
		MySQLDSN:   "root:rootpassword@tcp(127.0.0.1:3306)/code_exec?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:  "127.0.0.1:6379",
		ServerPort: ":8080",
		DockerHost: "unix:///var/run/docker.sock", // Linux 默认
		JWTSecret:  "", // JWT 密钥默认值留空，必须通过环境变量设置，否则会在 auth.Init() 中使用临时密钥并发出警告

		// 数据库连接池默认值，当用户需要使用数据库时使用，如用户注册、提交记录等
		DBMaxIdleConns: 10,
		DBMaxOpenConns: 100,

		// 容器资源默认值
		ContainerMemoryMB: 2048, // 默认 2GB
		ContainerCPUCores: 2.0,  // 默认 2 核

		// 弹性调度默认值
		ContainerLWM:    2, // 废弃
		ContainerHWM:    8, // 高水位：每种语言最多缓存 8 个闲置容器
		MonitorInterval: 3, // 每 3 秒检查一次水位

		// EMA 默认值 ，
		EMAAlpha: 0.3,
		LWMMin:   1,
		LWMMax:   20,

		// CORS 默认允许的源（开发环境）
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
		},

		// LLM 默认配置 (Zhipu AI)
		// glm-4-flashx 比 glm-4-flash 更快，适合代码分析场景
		LLMAPIKey: "", // 通过环境变量设置
		LLMAPIURL: "https://open.bigmodel.cn/api/paas/v4/chat/completions",
		LLMModel:  "codegeex-4",
		//LLMModel:  "glm-4-flashx",

		// 管理员账号 (生产环境必须修改!)
		AdminUsername: "admin",
		AdminPassword: "admin123",
	}

	// 如果需要，使用简单的环境变量覆盖
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		port := "3306"
		if p := os.Getenv("MYSQL_PORT"); p != "" {
			port = p // 允许覆盖默认端口
		}
		//把主机地址、端口、用户名、密码等信息组装成数据库驱动能听懂的格式，以便程序连接数据库
		cfg.MySQLDSN = fmt.Sprintf("root:rootpassword@tcp(%s:%s)/code_exec?charset=utf8mb4&parseTime=True&loc=Local", host, port)
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
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

	// 弹性调度 - 低水位线，废弃，使用动态下限替代
	if lwm := os.Getenv("CONTAINER_LWM"); lwm != "" {
		if val, err := strconv.Atoi(lwm); err == nil && val >= 0 {
			cfg.ContainerLWM = val
		}
	}

	// 弹性调度 - 高水位线，控制系统为每种语言最多缓存多少个闲置容器，以平衡资源使用和响应速度
	if hwm := os.Getenv("CONTAINER_HWM"); hwm != "" {
		if val, err := strconv.Atoi(hwm); err == nil && val > 0 {
			cfg.ContainerHWM = val
		}
	}

	// 监控协程检查间隔（秒），控制系统检查容器水位的频率，过短会增加系统负担，过长会导致响应变慢
	if interval := os.Getenv("MONITOR_INTERVAL"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil && val > 0 {
			cfg.MonitorInterval = val
		}
	}
	// 指数移动平均 (EMA)是一种常用的时间序列数据平滑技术，可以帮助系统更稳定地调整容器下限，
	// 适应不同的负载情况。通过环境变量配置 EMA 的 alpha 参数和动态下限的最小/最大值，
	// 可以让系统在负载变化时快速响应，同时避免过度波动。
	// 动态水位参数，使用指数移动平均 (EMA) 来平滑请求率的波动，从而动态调整容器的下限，适应不同的负载情况

	// alpha 参数控制 EMA 的平滑程度，值越小表示历史数据权重越大，响应越平滑但反应越慢；值越大表示当前数据权重越大，响应更快但可能更波动。
	if alpha := os.Getenv("EMA_ALPHA"); alpha != "" {
		if val, err := strconv.ParseFloat(alpha, 64); err == nil && val > 0 && val <= 1 {
			cfg.EMAAlpha = val // 0.3 表示当前请求率占 30%，历史请求率占 70%，适度平滑波动，快速响应负载变化
		}
	}

	// LWMMin 和 LWMMax 分别设置动态下限的最小值和最大值，确保系统在低负载时也能保持一定的预热容器，避免冷启动，
	// 同时在高负载时限制预热过多容器，防止资源过度消耗。
	if lwmMin := os.Getenv("LWM_MIN"); lwmMin != "" {
		if val, err := strconv.Atoi(lwmMin); err == nil && val >= 1 {
			cfg.LWMMin = val
		}
	}
	if lwmMax := os.Getenv("LWM_MAX"); lwmMax != "" {
		if val, err := strconv.Atoi(lwmMax); err == nil && val > cfg.LWMMin {
			cfg.LWMMax = val
		}
	}

	// CORS 允许的源
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		// 简单按逗号分隔
		cfg.AllowedOrigins = splitAndTrim(origins)
	}

	// Windows 下，DockerHost 通常是 npipe:////./pipe/docker_engine
	// 检查操作系统,如果是 Windows 则覆盖默认的 DockerHost 配置，以确保程序能够正确连接到 Docker 守护进程
	if os.Getenv("OS") == "Windows_NT" {
		cfg.DockerHost = "npipe:////./pipe/docker_engine"
	}

	// LLM 配置加载,支持通过环境变量覆盖默认值，
	// 方便在不同环境下使用不同的 LLM 服务或模型，
	// 例如开发环境使用免费或本地模型，生产环境使用更强大的云服务模型。
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
// 使用标准库实现,防止用户输入时多加空格导致配置错误，
// 例如 "http://localhost:5173, http://localhost:3000" 这种常见的格式
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
