package config

import (
	"fmt"
	"os"
)

type Config struct {
	MySQLDSN    string
	RedisAddr   string
	ServerPort  string
	DockerHost  string
	JWTSecret   string
}

func Load() *Config {
	// Default configuration
	cfg := &Config{
		MySQLDSN:   "root:rootpassword@tcp(127.0.0.1:13306)/code_exec?charset=utf8mb4&parseTime=True&loc=Local",
		RedisAddr:  "127.0.0.1:16379",
		ServerPort: ":8080",
		DockerHost: "unix:///var/run/docker.sock", // Default for Linux
		JWTSecret:  "your-secret-key",
	}

	// Override with simple environment variables if needed
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
	
	// For Windows, DockerHost is often npipe:////./pipe/docker_engine
    // check OS
    if os.Getenv("OS") == "Windows_NT" {
        cfg.DockerHost = "npipe:////./pipe/docker_engine"
    }

	return cfg
}
