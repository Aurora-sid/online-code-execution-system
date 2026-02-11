package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// 使用与 CORS 中间件相同的来源检查逻辑
		return IsOriginAllowed(origin)
	},
}

// redisAddr WebSocket 使用的 Redis 地址（由 RegisterRoutes 设置）
var wsRedisAddr string

// wsRedisClient WebSocket 使用的 Redis 客户端单例
var wsRedisClient *redis.Client
var wsRedisOnce sync.Once

// SetWSRedisAddr 设置 WebSocket 使用的 Redis 地址
func SetWSRedisAddr(addr string) {
	wsRedisAddr = addr
}

// getWSRedisClient 获取 WebSocket 使用的 Redis 客户端（单例）
func getWSRedisClient() *redis.Client {
	wsRedisOnce.Do(func() {
		wsRedisClient = redis.NewClient(&redis.Options{
			Addr:     wsRedisAddr,
			PoolSize: 50, // 连接池大小
		})
	})
	return wsRedisClient
}

// WebSocket 消息结构
type WSMessage struct {
	Type string `json:"type"` // "stdin" 或 "stdout"
	Data string `json:"data"`
}

func HandleWebSocket(c *gin.Context) {
	taskId := c.Query("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS 升级错误:", err)
		return
	}
	defer conn.Close()

	// 使用 Redis 客户端单例（不再每次连接都创建）
	rdb := getWSRedisClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 订阅任务输出
	pubsub := rdb.Subscribe(ctx, "task_output:"+taskId)
	defer pubsub.Close()

	// Goroutine: 从 Redis 读取输出并发送给客户端
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			wsMsg := WSMessage{
				Type: "stdout",
				Data: msg.Payload,
			}
			msgBytes, _ := json.Marshal(wsMsg)
			if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
				log.Println("WS 写入错误:", err)
				cancel()
				return
			}
		}
	}()

	// 主循环: 从客户端读取输入并发布到 Redis
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("WS 读取错误:", err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Println("WS 消息解析错误:", err)
			continue
		}

		// 如果是 stdin 类型，发布到 Redis
		if wsMsg.Type == "stdin" {
			if err := rdb.Publish(ctx, "task_input:"+taskId, wsMsg.Data).Err(); err != nil {
				log.Println("Redis 发布错误:", err)
			}
		}
	}
}
