package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"context"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for dev
	},
}

func HandleWebSocket(c *gin.Context) {
	taskId := c.Query("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS Upgrade Error:", err)
		return
	}
	defer conn.Close()

	// Subscribe to Redis
	// Ideally inject Redis client. For now create new or reuse global if available.
	// In a real app, pass *redis.Client to the handler via closure or struct.
	// Creating new client here is costly but functional for MVP.
	rdb := redis.NewClient(&redis.Options{
		Addr: "code_exec_redis:6379", // Use Docker network address
	})
	defer rdb.Close() // Close redis connection for this WS session

	ctx := context.Background()
	pubsub := rdb.Subscribe(ctx, "task_output:"+taskId)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Println("WS Write Error:", err)
			break
		}
	}
}
