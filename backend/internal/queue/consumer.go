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

/*进行 Redis 连接的建立封装以及提供面对服务器
由于长时间空跑链接意外中断导致获取失效后重连接机制
（含有指数退让避免造成闪断后的雪崩连接冲击）。*/
// 函数名+返回值类型+参数列表
func NewConsumer(addr string, pool *docker.Pool, db *gorm.DB) *Consumer {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,    // Redis 服务器地址
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
	// ctx是一个上下文对象，提供了控制函数执行的生命周期和取消机制。
	//  cancel是一个函数，用于取消上下文。当调用cancel()时，
	// ctx会被标记为已取消，所有使用该ctx的操作都会收到取消信号。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 通过发送PING命令测试连接是否成功。如果连接成功，返回nil；如果连接失败，返回错误信息。
	return c.client.Ping(ctx).Err()  // 返回连接测试结果，供调用者判断是否重连成功
}


// StartWorker 启动消费者工作进程，持续监听 Redis 队列并处理任务
func (c *Consumer) StartWorker() {
	ctx := context.Background()  // 创建一个新的上下文对象，通常用于控制函数执行的生命周期和取消机制。
	log.Println("[Consumer] 工作进程已启动，等待任务...") // 打印日志，表明消费者工作进程已经启动，并显示当前的Redis客户端信息

	consecutiveErrors := 0 // 连续错误计数
	maxRetries := 5        // 最大重试次数

	for {
		// BLPop 阻塞直到有新任务可用
		result, err := c.client.BLPop(ctx, 0, c.queue).Result()
		if err != nil {
			consecutiveErrors++  // 增加连续错误计数

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
// 限定执行时间为 25 秒，允许实时输入输出交互
func (c *Consumer) executeInteractive(ctx context.Context, task Task, containerID string) {
	startTime := time.Now() // 记录执行开始时间

	// 限定生命期为 25 秒，超过则强制终止
	execCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	// 创建 stdin 和 output channels，用于与 Sandbox 进行交互
	stdinChan := make(chan string, 10)
	outputChan := make(chan string, 100)

	// 订阅 stdin 输入频道
	pubsub := c.client.Subscribe(ctx, "task_input:"+task.ID)
	defer pubsub.Close()

	// Goroutine: 从 Redis 接收输入并发送到 stdinChan
	go func() {
		ch := pubsub.Channel() // 获取订阅频道的消息通道
		for {
			select {
			case msg, ok := <-ch:  // 从消息通道接收消息
				if !ok {
					return  // 频道关闭，退出 Goroutine
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
	
	// 等待前端的 WebSocket 成功订阅 Redis 的 task_output 频道
	for i := 0; i < 30; i++ { // 最多等待 30 * 100ms = 3秒
		res, err := c.client.PubSubNumSub(ctx, "task_output:"+task.ID).Result()
		// 如果订阅人数 > 0，说明前端 WebSocket 已经准备好了
		if err == nil && res["task_output:"+task.ID] > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}


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
