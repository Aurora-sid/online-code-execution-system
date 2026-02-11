package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"code-exec/internal/docker"
	"code-exec/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Consumer struct {
	client    *redis.Client
	queue     string
	pool      *docker.Pool
	db        *gorm.DB
	redisAddr string // 保存 Redis 地址用于重连
}

func NewConsumer(addr string, pool *docker.Pool, db *gorm.DB) *Consumer {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     10,              // 连接池大小
		MinIdleConns: 2,               // 最小空闲连接
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
	})
	return &Consumer{
		client:    rdb,
		queue:     "code_execution_queue",
		pool:      pool,
		db:        db,
		redisAddr: addr,
	}
}

// reconnect 尝试重新连接 Redis
func (c *Consumer) reconnect() error {
	// 关闭旧连接
	if c.client != nil {
		c.client.Close()
	}

	// 创建新连接
	c.client = redis.NewClient(&redis.Options{
		Addr:         c.redisAddr,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.client.Ping(ctx).Err()
}

func (c *Consumer) StartWorker() {
	ctx := context.Background()
	log.Println("[Consumer] 工作进程已启动，等待任务...")

	consecutiveErrors := 0 // 连续错误计数
	maxRetries := 5        // 最大重试次数

	for {
		// BLPop 阻塞直到有新任务可用
		result, err := c.client.BLPop(ctx, 0, c.queue).Result()
		if err != nil {
			consecutiveErrors++

			// 计算指数退避时间: 1s, 2s, 4s, 8s, 16s (最大)
			backoff := time.Duration(1<<min(consecutiveErrors-1, 4)) * time.Second
			log.Printf("[Consumer] Redis BLPop 错误 (第 %d 次): %v，%v 后重试...\n", consecutiveErrors, err, backoff)

			// 如果连续错误达到阈值，尝试重连
			if consecutiveErrors >= maxRetries {
				log.Println("[Consumer] 连续错误次数过多，尝试重新连接 Redis...")
				if reconnErr := c.reconnect(); reconnErr != nil {
					log.Printf("[Consumer] 重连失败: %v\n", reconnErr)
				} else {
					log.Println("[Consumer] Redis 重连成功！")
					consecutiveErrors = 0 // 重置错误计数
				}
			}

			time.Sleep(backoff)
			continue
		}

		// 成功获取任务，重置错误计数
		consecutiveErrors = 0

		// result[0] 是队列名称，result[1] 是值
		var task Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("[Consumer] 反序列化任务错误: %v\n", err)
			continue
		}

		log.Printf("[Consumer] 收到新任务: %s (Lang: %s, User: %d)\n", task.ID, task.Language, task.UserID)

		// 更新状态为正在运行
		if task.SubmissionID > 0 {
			c.db.Model(&model.Submission{}).Where("id = ?", task.SubmissionID).Update("status", "Running")
		}

		// 获取容器
		containerID, err := c.pool.GetContainer(ctx, task.Language)
		if err != nil {
			log.Printf("[Consumer] 获取 %s 容器失败: %v\n", task.Language, err)
			if task.SubmissionID > 0 {
				c.db.Model(&model.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
					"status": "Failed",
					"output": fmt.Sprintf("Failed to get container: %v", err),
				})
			}
			// 发布错误到输出频道
			c.client.Publish(ctx, "task_output:"+task.ID, fmt.Sprintf("Error: %v", err))
			continue
		}

		// 使用交互式执行
		c.executeInteractive(ctx, task, containerID)

		// 归还/清理容器
		c.pool.ReturnContainer(ctx, containerID) // Pool 现在内部处理语言追踪

		log.Printf("[Consumer] 任务 %s 处理完成\n", task.ID)
	}
}

// executeInteractive 使用交互式方式执行代码
func (c *Consumer) executeInteractive(ctx context.Context, task Task, containerID string) {
	startTime := time.Now() // 记录执行开始时间

	execCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	// 创建 stdin 和 output channels
	stdinChan := make(chan string, 10)
	outputChan := make(chan string, 100)

	// 订阅 stdin 输入频道
	pubsub := c.client.Subscribe(ctx, "task_input:"+task.ID)
	defer pubsub.Close()

	// Goroutine: 从 Redis 接收输入并发送到 stdinChan
	go func() {
		ch := pubsub.Channel()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case stdinChan <- msg.Payload:
				case <-execCtx.Done():
					return
				}
			case <-execCtx.Done():
				return
			}
		}
	}()

	// Goroutine: 从 outputChan 读取输出并发布到 Redis
	var outputBuilder strings.Builder
	go func() {
		for {
			select {
			case output, ok := <-outputChan:
				if !ok {
					return
				}
				outputBuilder.WriteString(output)
				// 发布到 Redis 供 WebSocket 转发
				c.client.Publish(ctx, "task_output:"+task.ID, output)
			case <-execCtx.Done():
				return
			}
		}
	}()

	// 执行代码
	exitCode, err := c.pool.Sandbox.ExecuteInteractive(execCtx, containerID, task.Language, task.Code, stdinChan, outputChan)

	// 关闭 channels
	close(stdinChan)
	close(outputChan)

	// 确定状态
	status := "Success"
	finalOutput := outputBuilder.String()

	if err != nil {
		log.Printf("[Consumer] 任务 %s 执行出错: %v\n", task.ID, err)
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "超时") {
			status = "Timeout"
		} else {
			status = "Failed"
		}
		finalOutput = fmt.Sprintf("%s\nError: %v", finalOutput, err)
	} else if exitCode != 0 {
		status = "Failed"
		log.Printf("[Consumer] 任务 %s 执行失败 (ExitCode: %d)\n", task.ID, exitCode)
		if exitCode == 137 {
			finalOutput = fmt.Sprintf("%s\nError: Process killed (OOM?)", finalOutput)
		}
		if strings.Contains(finalOutput, "Resource temporarily unavailable") {
			finalOutput = fmt.Sprintf("%s\n\nSecurity Alert: Detected massive process creation (Fork Bomb Attempt).", finalOutput)
			status = "Security Violation"
		}
		if strings.Contains(finalOutput, "MemoryError") {
			finalOutput = fmt.Sprintf("%s\n\nSecurity Alert: Detected excessive memory allocation (Memory Bomb Attempt).", finalOutput)
			status = "Security Violation"
		}
	}

	// 更新数据库中的提交记录
	if task.SubmissionID > 0 {
		c.db.Model(&model.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
			"status": status,
			"output": finalOutput,
		})
	}

	// 发送完成信号（包含执行时长）
	elapsed := time.Since(startTime)
	c.client.Publish(ctx, "task_output:"+task.ID, fmt.Sprintf("\n[执行完成] 耗时: %.3fs", elapsed.Seconds()))
}
