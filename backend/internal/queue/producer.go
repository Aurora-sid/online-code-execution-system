package queue
/*本模块是系统最前沿接收来自于前端 API 的打包网关接口。
将用户通过网页传输过来的评测数据构建为标准的传输格式
向高可用中间件 Redis 下达任务，
起到了削峰以及与后台执行层逻辑分离的解耦作用。*/

// 在backend\internal\api\handler.go中，我们在处理代码执行请求的API路由中，
// 调用了Producer的AddTask方法，将用户提交的代码执行任务添加到Redis队列中。
// 这样，前端API只负责接收请求并将任务放入队列，而不需要等待后台执行层完成代码编译和运行的过程，
// 从而实现了削峰填谷和解耦的设计目标。
import (
	"context"
	"encoding/json"
	"time"
	"github.com/redis/go-redis/v9"
)

// 定义一个任务的消息传递格式
type Task struct {
	ID           string `json:"id"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	Input        string `json:"input"` // 标准输入数据
	UserID       uint   `json:"user_id"`
	SubmissionID uint   `json:"submission_id"`
}

// Producer 负责将任务添加到 Redis 队列中
type Producer struct {

	client *redis.Client
	queue  string
}

func NewProducer(addr string) *Producer {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Producer{
		// rdb 是 Redis 客户端实例，负责与 Redis 服务器进行通信
		client: rdb,
		// 定义一个固定的队列名称，所有任务都放在这个队列里
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
/*这种设计有三个主要目的：
削峰填谷：如果瞬间有 1000 个用户提交代码，Redis 队列可以排队处理，防止后台服务器被瞬间压垮。
解耦：前端 API 只需要负责“扔任务”到 Redis，扔完就返回“提交成功”，不需要等待后台漫长的代码编译和运行过程。
高可用：Redis 是内存数据库，读写极快。即使后台执行层（Worker）暂时挂了，任务也还在 Redis 里存着，不会丢失。*/