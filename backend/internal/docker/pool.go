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
	"github.com/docker/docker/api/types/filters"
)

// PoolLabelKey 是用于标识容器池容器的标签键
const PoolLabelKey = "com.docker.compose.project"

// PoolLabelValue 是容器池容器的标签值
const PoolLabelValue = "container-pool"

type Pool struct {
	Sandbox *Sandbox
	// 语言 -> 容器ID列表 的映射
	Available map[string][]string
	// 容器ID -> 语言 的映射（用于追踪）
	containerLang map[string]string
	mu            sync.Mutex
	maxPoolSize   int   // 每种语言最大缓存容器数
	memoryBytes   int64 // 容器内存限制 (bytes)
	nanoCPUs      int64 // 容器 CPU 限制 (纳秒)
}

// NewPool 创建一个新的容器池，使用配置中的资源限制
func NewPool(s *Sandbox, cfg *config.Config) *Pool {
	maxPoolSize := cfg.ContainerPoolSize
	if maxPoolSize <= 0 {
		maxPoolSize = 3 // 默认值
	}

	// 转换配置值
	memoryBytes := int64(cfg.ContainerMemoryMB) * 1024 * 1024
	nanoCPUs := int64(cfg.ContainerCPUCores * 1e9)

	log.Printf("[Pool] 初始化容器池: 池大小=%d, 内存=%dMB, CPU=%.1f核\n",
		maxPoolSize, cfg.ContainerMemoryMB, cfg.ContainerCPUCores)

	return &Pool{
		Sandbox:       s,
		Available:     make(map[string][]string),
		containerLang: make(map[string]string),
		maxPoolSize:   maxPoolSize,
		memoryBytes:   memoryBytes,
		nanoCPUs:      nanoCPUs,
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
				PoolLabelKey: PoolLabelValue, // 自定义标签，避免被 Docker Desktop 误识别为 Compose 项目
			},
			Env: func() []string {
				if language == "go" {
					return []string{"CGO_ENABLED=0", "GOCACHE=/app/.cache"} // 禁用 CGO 加速编译，指定缓存目录
				}
				return nil
			}(),
		},
		// 限制容器资源（使用配置值）
		&container.HostConfig{
			Resources: container.Resources{
				Memory:    p.memoryBytes,
				NanoCPUs:  p.nanoCPUs,
				PidsLimit: func() *int64 { v := int64(100); return &v }(), // 进程数限制
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
		// 池已满，销毁容器（异步执行但记录日志追踪错误）
		log.Printf("[Pool] 池已满，销毁容器 %s (语言: %s)\n", containerID[:12], lang)
		delete(p.containerLang, containerID)
		go func(id string) {
			if err := p.Sandbox.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Printf("[Pool] 异步删除容器 %s 失败: %v\n", id[:12], err)
			}
		}(containerID)
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

// CleanupStaleContainers 清理所有带有容器池标签的残留容器
// 应在程序启动时调用，以清除上次运行遗留的容器
func (p *Pool) CleanupStaleContainers(ctx context.Context) (int, error) {
	// 使用标签过滤器查找所有容器池容器
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", PoolLabelKey+"="+PoolLabelValue)

	containers, err := p.Sandbox.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true, // 包括已停止的容器
		Filters: filterArgs,
	})
	if err != nil {
		return 0, fmt.Errorf("列出容器失败: %w", err)
	}

	removedCount := 0
	for _, c := range containers {
		log.Printf("[Pool] 清理残留容器: %s (状态: %s)\n", c.ID[:12], c.State)
		if err := p.Sandbox.cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Printf("[Pool] 删除容器 %s 失败: %v\n", c.ID[:12], err)
		} else {
			removedCount++
		}
	}

	if removedCount > 0 {
		log.Printf("[Pool] 已清理 %d 个残留容器\n", removedCount)
	}
	return removedCount, nil
}

// ResetPool 重置容器池：清空所有缓存的容器
func (p *Pool) ResetPool(ctx context.Context) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	removedCount := 0

	// 清理所有已缓存的容器
	for lang, ids := range p.Available {
		for _, id := range ids {
			log.Printf("[Pool] 重置: 销毁容器 %s (语言: %s)\n", id[:12], lang)
			if err := p.Sandbox.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Printf("[Pool] 删除容器 %s 失败: %v\n", id[:12], err)
			} else {
				removedCount++
			}
		}
	}

	// 清空池状态
	p.Available = make(map[string][]string)
	p.containerLang = make(map[string]string)

	log.Printf("[Pool] 容器池已重置，共销毁 %d 个容器\n", removedCount)
	return removedCount, nil
}

// PoolStats 容器池统计信息
type PoolStats struct {
	TotalContainers int                      `json:"total"`
	MaxPerLang      int                      `json:"maxPerLang"`
	MemoryMB        int                      `json:"memoryMB"`
	CPUCores        float64                  `json:"cpuCores"`
	LangDetails     map[string]LangPoolStats `json:"details"`
}

// LangPoolStats 每种语言的容器池统计
type LangPoolStats struct {
	Idle   int `json:"idle"`
	Active int `json:"active"`
}

// GetStats 获取容器池实时统计信息
func (p *Pool) GetStats() *PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := &PoolStats{
		MaxPerLang:      p.maxPoolSize,
		MemoryMB:        int(p.memoryBytes / 1024 / 1024),
		CPUCores:        float64(p.nanoCPUs) / 1e9,
		LangDetails:     make(map[string]LangPoolStats),
		TotalContainers: len(p.containerLang),
	}

	// 统计各语言的闲置容器数
	idleCounts := make(map[string]int)
	for lang, ids := range p.Available {
		idleCounts[lang] = len(ids)
	}

	// 统计各语言的总容器数
	totalCounts := make(map[string]int)
	for _, lang := range p.containerLang {
		totalCounts[lang]++
	}

	// 组装详情
	// 遍历所有已知语言（包括有活跃容器但没有闲置容器的）
	allLangs := make(map[string]bool)
	for lang := range idleCounts {
		allLangs[lang] = true
	}
	for lang := range totalCounts {
		allLangs[lang] = true
	}

	for lang := range allLangs {
		total := totalCounts[lang]
		idle := idleCounts[lang]
		active := total - idle
		if active < 0 {
			active = 0 // 理论上不应发生
		}

		stats.LangDetails[lang] = LangPoolStats{
			Idle:   idle,
			Active: active,
		}
	}

	return stats
}
