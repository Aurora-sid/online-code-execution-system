package docker

import (
	"code-exec/config"
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

type Pool struct {
	Sandbox *Sandbox
	// 语言 -> 容器ID列表 的映射
	Available map[string][]string
	// 容器ID -> 语言 的映射（用于追踪）
	containerLang map[string]string
	mu            sync.Mutex
	maxPoolSize   int // 每种语言最大缓存容器数
}

func NewPool(s *Sandbox) *Pool {
	return &Pool{
		Sandbox:       s,
		Available:     make(map[string][]string),
		containerLang: make(map[string]string),
		maxPoolSize:   3, // 每种语言最多保留3个空闲容器
	}
}

// GetContainer 为指定语言返回一个容器ID
func (p *Pool) GetContainer(ctx context.Context, language string) (string, error) {
	p.mu.Lock()
	// 1. 检查是否有可用容器
	if ids, ok := p.Available[language]; ok && len(ids) > 0 {
		// 从栈弹出 (LIFO)
		id := ids[len(ids)-1]
		p.Available[language] = ids[:len(ids)-1]
		p.mu.Unlock()

		// 检查容器是否存活
		if _, err := p.Sandbox.cli.ContainerInspect(ctx, id); err == nil {
			log.Printf("[Pool] 从缓存获取容器 %s (语言: %s)\n", id[:12], language)
			return id, nil
		}

		// 容器已死，清理并递归获取
		log.Printf("[Pool] 缓存容器 %s 已失效，丢弃并重新获取\n", id[:12])
		return p.GetContainer(ctx, language)
	}
	p.mu.Unlock()

	// 2. 创建新容器
	return p.createContainer(ctx, language)
}

func (p *Pool) createContainer(ctx context.Context, language string) (string, error) {
	// 将内部语言名称映射到 Docker 镜像标签
	langCfg, err := config.GetLanguageConfig(language)
	if err != nil {
		return "", err
	}
	imageName := langCfg.Image
	log.Printf("[Pool] 池中无可用容器，正在创建新容器 (语言: %s, 镜像: %s)\n", language, imageName)

	resp, err := p.Sandbox.cli.ContainerCreate(ctx,
		&container.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "infinity"},
			Tty:   false,
			Labels: map[string]string{
				"com.docker.compose.project": "container-pool", // 让容器在 Docker Desktop 中归入 "container-pool" 项目组
			},
			Env: func() []string {
				if language == "go" {
					return []string{"CGO_ENABLED=0", "GOCACHE=/app/.cache"} // 禁用 CGO 加速编译，指定缓存目录
				}
				return nil
			}(),
		},
		// 限制容器资源
		&container.HostConfig{
			Resources: container.Resources{
				Memory:    2 * 1024 * 1024 * 1024,                         // 2GB（为 Go 编译器增加）
				NanoCPUs:  2000000000,                                     // 2.0 CPU
				PidsLimit: func() *int64 { v := int64(100); return &v }(), // 增加进程数限制
			},
			NetworkMode: "none", // 安全：无网络
		},
		nil, nil, "")

	if err != nil {
		log.Printf("[Pool] 创建容器失败: %v\n", err)
		return "", err
	}

	if err := p.Sandbox.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Printf("[Pool] 启动容器失败: %v\n", err)
		return "", err
	}

	// 针对 Go 语言进行预热 (Warm-up) - 同步执行确保预热完成
	if language == "go" {
		log.Printf("[Pool] 正在预热 Go 容器 %s (生成 build cache)...\n", resp.ID[:12])
		if err := p.warmupGoContainer(ctx, resp.ID); err != nil {
			log.Printf("[Pool] Go 容器预热失败: %v\n", err)
		} else {
			log.Printf("[Pool] Go 容器预热完成，已准备好秒级响应\n")
		}
	}

	// 注册追踪
	p.mu.Lock()
	p.containerLang[resp.ID] = language
	p.mu.Unlock()

	log.Printf("[Pool] 新容器已创建并启动 %s (语言: %s)\n", resp.ID[:12], language)

	return resp.ID, nil
}

// ReturnContainer 将容器放回或销毁
func (p *Pool) ReturnContainer(ctx context.Context, containerID string) {
	// 查找容器语言
	p.mu.Lock()
	lang, exists := p.containerLang[containerID]
	p.mu.Unlock()

	if !exists {
		// 未知容器，直接销毁
		log.Printf("[Pool] 未知容器 %s，直接销毁\n", containerID[:12])
		p.Sandbox.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
		return
	}

	// 尝试清理容器
	if err := p.cleanContainer(ctx, containerID); err != nil {
		log.Printf("[Pool] 清理容器 %s 失败: %v，销毁容器\n", containerID[:12], err)
		p.destroyContainer(ctx, containerID)
		return
	}

	// 尝试放回池中
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查池是否已满
	if len(p.Available[lang]) < p.maxPoolSize {
		// 池未满，放回池中
		p.Available[lang] = append(p.Available[lang], containerID)
		log.Printf("[Pool] 容器 %s 已归还到池中 (语言: %s, 池大小: %d/%d)\n", containerID[:12], lang, len(p.Available[lang]), p.maxPoolSize)
	} else {
		// 池已满，销毁容器
		log.Printf("[Pool] 池已满，销毁容器 %s (语言: %s)\n", containerID[:12], lang)
		go p.Sandbox.cli.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{Force: true})
		delete(p.containerLang, containerID)
	}
}

// cleanContainer 清理容器工作区
func (p *Pool) cleanContainer(ctx context.Context, containerID string) error {
	// 使用命令删除 /app 下的所有文件
	execConfig := types.ExecConfig{
		Cmd: []string{"sh", "-c", "rm -rf /app/*"},
	}

	resp, err := p.Sandbox.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	if err := p.Sandbox.cli.ContainerExecStart(ctx, resp.ID, types.ExecStartCheck{}); err != nil {
		return err
	}

	// 等待清理完成
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	inspectTicker := time.NewTicker(100 * time.Millisecond)
	defer inspectTicker.Stop()

	for {
		select {
		case <-timer.C:
			return fmt.Errorf("cleanup timeout")
		case <-inspectTicker.C:
			inspect, err := p.Sandbox.cli.ContainerExecInspect(ctx, resp.ID)
			if err != nil {
				return err
			}
			if !inspect.Running {
				if inspect.ExitCode != 0 {
					return fmt.Errorf("cleanup failed with exit code %d", inspect.ExitCode)
				}
				return nil
			}
		}
	}
}

func (p *Pool) destroyContainer(ctx context.Context, containerID string) {
	p.Sandbox.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
	p.mu.Lock()
	delete(p.containerLang, containerID)
	p.mu.Unlock()
}

// warmupGoContainer 运行一个空的 Go 程序以预热编译缓存
func (p *Pool) warmupGoContainer(ctx context.Context, containerID string) error {
	warmupCode := `package main
func main() {}`
	_, _, err := p.Sandbox.Execute(ctx, containerID, "go", warmupCode, "")
	return err
}
