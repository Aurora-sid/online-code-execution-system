package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许所有来源，生产环境应该限制
		return true
	},
}

func main() {
	// 初始化Gin路由
	r := gin.Default()

	// 健康检查接口
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// WebSocket运行代码端点
	r.GET("/ws/run", handleWebSocket)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// 代码执行请求
 type CodeRunRequest struct {
	Type     string `json:"type"`
	Code     string `json:"code"`
	Language string `json:"language"`
}

// handleWebSocket 处理WebSocket连接
func handleWebSocket(c *gin.Context) {
	// 升级HTTP连接为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket连接已建立")

	// 发送欢迎消息
	welcomeMsg := "欢迎使用在线编程系统！\n\n请输入代码，然后按Ctrl+D结束输入...\n"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg)); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// 循环处理WebSocket消息
	for {
		// 读取客户端消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket读取错误: %v", err)
			break
		}

		// 解析JSON消息
		var req CodeRunRequest
		if err := json.Unmarshal(message, &req); err != nil {
			// 如果不是JSON格式，直接回显
			response := fmt.Sprintf("收到消息: %s\n", string(message))
			if err := conn.WriteMessage(messageType, []byte(response)); err != nil {
				log.Printf("WebSocket写入错误: %v", err)
				break
			}
			continue
		}

		// 处理代码执行请求
		if req.Type == "run" {
			log.Printf("执行代码请求: 语言=%s, 代码长度=%d", req.Language, len(req.Code))
			
			// 发送执行开始消息
			if err := conn.WriteMessage(websocket.TextMessage, []byte("开始执行代码...\n")); err != nil {
				log.Printf("发送执行开始消息失败: %v", err)
				break
			}

			// 执行代码
			result, err := runCodeInDocker(req.Code, req.Language)
			if err != nil {
				// 发送错误消息
				errMsg := fmt.Sprintf("执行错误: %v\n", err)
				if err := conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
					log.Printf("发送错误消息失败: %v", err)
					break
				}
			} else {
				// 发送执行结果
				if err := conn.WriteMessage(websocket.TextMessage, []byte("执行结果:\n")); err != nil {
					log.Printf("发送结果消息失败: %v", err)
					break
				}
				if err := conn.WriteMessage(websocket.TextMessage, []byte(result)); err != nil {
					log.Printf("发送执行结果失败: %v", err)
					break
				}
			}

			// 发送执行完成消息
			if err := conn.WriteMessage(websocket.TextMessage, []byte("\n执行完成\n")); err != nil {
				log.Printf("发送执行完成消息失败: %v", err)
				break
			}
		}

		// 如果收到结束消息，关闭连接
		if string(message) == "exit" {
			break
		}
	}

	log.Printf("WebSocket连接已关闭")
}

// runCodeInDocker 使用Docker命令行运行代码
func runCodeInDocker(code string, language string) (string, error) {
	// 创建一个简单的测试，不涉及复杂的引号嵌套
	// 测试Docker是否能正常运行
	dockerCmd := "docker run --rm hello-world"

	// 执行Docker命令（兼容Windows和Linux）
	var cmd *exec.Cmd
	if os.PathSeparator == '\\' {
		// Windows环境
		cmd = exec.Command("cmd.exe", "/c", dockerCmd)
	} else {
		// Linux/Unix环境
		cmd = exec.Command("sh", "-c", dockerCmd)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行Docker命令失败: %w\n输出: %s", err, string(output))
	}

	return "Docker集成测试成功！\n" + string(output), nil
}