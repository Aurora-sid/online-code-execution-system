package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"code-exec/internal/docker"
)

type Consumer struct {
	client *redis.Client
	queue  string
	pool   *docker.Pool
}

func NewConsumer(addr string, pool *docker.Pool) *Consumer {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Consumer{
		client: rdb,
		queue:  "code_execution_queue",
		pool:   pool,
	}
}

func (c *Consumer) StartWorker() {
	ctx := context.Background()
	log.Println("Worker started, waiting for tasks...")

	for {
		// BLPop blocks until a task is available
		result, err := c.client.BLPop(ctx, 0, c.queue).Result()
		if err != nil {
			log.Println("Redis BLPop Error:", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// result[0] is queue name, result[1] is value
		var task Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Println("Unmarshal Task Error:", err)
			continue
		}

		fmt.Printf("Processing Task: %v\n", task.ID)
		
		// 1. Get Container
		containerID, err := c.pool.GetContainer(ctx, task.Language)
		if err != nil {
			log.Printf("Failed to get container for %s: %v\n", task.Language, err)
			continue
		}

		// 2. Execute
		output, err := c.pool.Sandbox.Execute(ctx, containerID, task.Language, task.Code)
		if err != nil {
			log.Printf("Execution error: %v\n", err)
			output = fmt.Sprintf("Error: %v", err)
		}

		// 3. Return/Cleanup Container
		c.pool.ReturnContainer(ctx, containerID)

		fmt.Printf("Task %s Finished. Output: %s\n", task.ID, output)
		
		// Publish output
		if err := c.client.Publish(ctx, "task_output:"+task.ID, output).Err(); err != nil {
			log.Println("Redis Publish Error:", err)
		}
	}
}
