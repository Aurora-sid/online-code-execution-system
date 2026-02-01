package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Task struct {
	ID           string `json:"id"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	Input        string `json:"input"` // 标准输入数据
	UserID       uint   `json:"user_id"`
	SubmissionID uint   `json:"submission_id"`
}

type Producer struct {
	client *redis.Client
	queue  string
}

func NewProducer(addr string) *Producer {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Producer{
		client: rdb,
		queue:  "code_execution_queue",
	}
}

func (p *Producer) AddTask(task Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return p.client.RPush(ctx, p.queue, data).Err()
}
