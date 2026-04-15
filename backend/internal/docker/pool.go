package docker

import (
	"code-exec/config"
	"context"
	"encoding/json"
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

// shortID 安全截断容器 ID 用于日志显示
func shortID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// Pool 弹性容器池 —— 基于双水位线的弹性调度
// LWM (Low Water Mark):  闲置数低于此值时，后台自动预热补充
// HWM (High Water Mark): 闲置数超过此值时，归还的容器直接销毁
type Pool struct {
	Sandbox *Sandbox
	// 语言 -> 容器ID列表 的映射（LIFO 栈）
	Available map[string][]string
	// 容器ID -> 语言 的映射（用于追踪）
	containerLang map[string]string
	mu            sync.Mutex

	// 双水位线配置与安全观测
	predictor   *TrafficPredictor
	hwm         int // High Water Mark: 最多闲置数
	ebpfMonitor *EBPFMonitor

	// 监控协程
	monitorInterval time.Duration
	stopMonitor     chan struct{}
	monitorOnce     sync.Once // 确保 stopMonitor 只关闭一次

	// 资源限制
	memoryBytes int64 // 容器内存限制 (bytes)
	nanoCPUs    int64 // 容器 CPU 限制 (纳秒)

	// 已知语言列表（用于预热所有语言的容器）
	languages []string
}

// NewPool 创建一个新的弹性容器池
func NewPool(s *Sandbox, cfg *config.Config) *Pool {
	lwmMin := cfg.LWMMin
	if lwmMin < 1 {
		lwmMin = 1
	}
	lwmMax := cfg.LWMMax
	if lwmMax < lwmMin {
		lwmMax = lwmMin + 5
	}
	hwm := cfg.ContainerHWM
	if hwm <= lwmMax {
		hwm = lwmMax + 5
	}

	monitorSec := cfg.MonitorInterval
	if monitorSec <= 0 {
		monitorSec = 3
	}

	// 转换配置值
	memoryBytes := int64(cfg.ContainerMemoryMB) * 1024 * 1024
	nanoCPUs := int64(cfg.ContainerCPUCores * 1e9)

	// 从语言配置中获取所有已注册的语言
	var languages []string
	for id := range config.Languages {
		languages = append(languages, id)
	}

	log.Printf("[Pool] 弹性容器池初始化: EMA(α=%.2f, LWM=%d~%d), HWM=%d, 监控间隔=%ds, 内存=%dMB, CPU=%.1f核, 语言数=%d\n",
		cfg.EMAAlpha, lwmMin, lwmMax, hwm, monitorSec, cfg.ContainerMemoryMB, cfg.ContainerCPUCores, len(languages))

	return &Pool{
		Sandbox:         s,
		Available:       make(map[string][]string),
		containerLang:   make(map[string]string),
		predictor:       NewTrafficPredictor(cfg.EMAAlpha, lwmMin, lwmMax),
		hwm:             hwm,
		ebpfMonitor:     NewEBPFMonitor(s.cli),
		monitorInterval: time.Duration(monitorSec) * time.Second,
		stopMonitor:     make(chan struct{}),
		memoryBytes:     memoryBytes,
		nanoCPUs:        nanoCPUs,
		languages:       languages,
	}
}

// ============================================================
//  阶段一：基于低水位的自动预热 (Auto Pre-warming)
// ============================================================

// StartMonitor 启动后台水位监控协程
// 每隔 monitorInterval 秒检查一次所有语言的闲置容器水位
// 如果低于 LWM，自动异步创建容器补充
func (p *Pool) StartMonitor(ctx context.Context) {
	log.Printf("[Pool] 水位监控协程已启动 (间隔: %v, EMA LWM: %d~%d, HWM: %d)\n", p.monitorInterval, p.predictor.lwmMin, p.predictor.lwmMax, p.hwm)
	ticker := time.NewTicker(p.monitorInterval)

	// 启动 EMA 预测器，通知回调触发紧急预热
	p.predictor.Start(1*time.Second, func() {
		log.Printf("[Pool] ⚡ 检测到突发流量，触发激进预热\n")
		p.checkAndPrewarm(ctx)
	})

	go func() {
		defer ticker.Stop()
		// 启动时立即执行一次预热检查
		p.checkAndPrewarm(ctx)

		for {
			select {
			case <-ticker.C:
				p.checkAndPrewarm(ctx)
			case <-p.stopMonitor:
				log.Println("[Pool] 水位监控协程已停止")
				return
			}
		}
	}()
}

// StopMonitor 停止监控协程
func (p *Pool) StopMonitor() {
	p.monitorOnce.Do(func() {
		close(p.stopMonitor)
		p.predictor.Stop()
		log.Println("[Pool] 正在停止水位监控协程...")
	})
}

// checkAndPrewarm 检查所有语言的水位，低于 LWM 时异步预热
func (p *Pool) checkAndPrewarm(ctx context.Context) {
	p.mu.Lock()
	type warmupJob struct {
		lang string
		diff int
	}
	var jobs []warmupJob

	currentLWM := p.predictor.GetDynamicLWM()

	for _, lang := range p.languages {
		idle := len(p.Available[lang])
		if idle < currentLWM {
			diff := currentLWM - idle
			jobs = append(jobs, warmupJob{lang: lang, diff: diff})
		}
	}
	p.mu.Unlock()

	// 异步补充，信号量控制最大并发数（防止 Docker daemon 过载）
	const maxConcurrent = 4
	sem := make(chan struct{}, maxConcurrent)
	for _, job := range jobs {
		log.Printf("[Pool] 预热: %s 水位不足 (当前: %d, 动态 LWM: %d), 补充 %d 个容器\n",
			job.lang, currentLWM-job.diff, currentLWM, job.diff)
		for i := 0; i < job.diff; i++ {
			sem <- struct{}{}
			go func(lang string) {
				defer func() { <-sem }()
				p.prewarmContainer(ctx, lang)
			}(job.lang)
		}
	}
}

// prewarmContainer 预热创建一个容器并放入池中
func (p *Pool) prewarmContainer(ctx context.Context, language string) {
	id, err := p.createContainer(ctx, language)
	if err != nil {
		log.Printf("[Pool] 预热创建 %s 容器失败: %v\n", language, err)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// 严格控制：预热只补充到动态 LWM，避免过度预热
	currentLWM := p.predictor.GetDynamicLWM()

	if len(p.Available[language]) < currentLWM {
		p.Available[language] = append(p.Available[language], id)
		log.Printf("[Pool] 预热完成: %s 容器 %s 已加入池中 (闲置: %d/%d 动态 LWM)\n",
			language, shortID(id), len(p.Available[language]), currentLWM)
	} else {
		// 预热期间水位已达标，销毁多余容器
		log.Printf("[Pool] 预热完成但水位已达动态 LWM=%d，销毁多余 %s 容器 %s\n",
			currentLWM, language, shortID(id))
		delete(p.containerLang, id)
		go func(cid string) {
			p.Sandbox.cli.ContainerRemove(context.Background(), cid, types.ContainerRemoveOptions{Force: true})
		}(id)
	}
}

// ============================================================
//  阶段二：穿透式爆发获取策略 (Burst Allocation)
// ============================================================

// GetContainer 为指定语言返回一个容器ID
// 策略：LIFO 优先获取热容器 → 池空时爆发穿透直接创建
func (p *Pool) GetContainer(ctx context.Context, language string) (string, error) {
	p.predictor.RecordRequest() // 记录请求用于 EMA 预测

	// 循环实现替代递归，防止大量死容器导致栈溢出
	const maxRetries = 10
	for retry := 0; retry < maxRetries; retry++ {
		p.mu.Lock()
		if ids, ok := p.Available[language]; ok && len(ids) > 0 {
			// LIFO 弹出：栈顶容器 CPU 缓存最热
			id := ids[len(ids)-1]
			p.Available[language] = ids[:len(ids)-1]
			currentIdle := len(p.Available[language])
			p.mu.Unlock()

			// 检查容器是否存活
			if _, err := p.Sandbox.cli.ContainerInspect(ctx, id); err == nil {
				// 获取成功，启动 eBPF 安全监控
				p.ebpfMonitor.StartMonitoring(ctx, id, language)

				log.Printf("[Pool] 缓存命中 %s (语言: %s, 剩余闲置: %d)\n", shortID(id), language, currentIdle)
				return id, nil
			}

			// 容器已死：清除追踪记录，继续尝试下一个
			log.Printf("[Pool] 缓存容器 %s 已失效，清除追踪并重试\n", shortID(id))
			p.mu.Lock()
			delete(p.containerLang, id)
			p.mu.Unlock()
			continue
		}
		p.mu.Unlock()
		break // 池为空，跳出循环走爆发模式
	}

	// 爆发穿透模式：池为空或全部失效，直接创建新容器
	log.Printf("[Pool] ⚡ 爆发模式: %s 池为空，直接创建新容器\n", language)
	newID, err := p.createContainer(ctx, language)
	if err == nil {
		// 获取成功，启动 eBPF 安全监控
		p.ebpfMonitor.StartMonitoring(ctx, newID, language)
	}
	return newID, err
}

// createContainer 创建一个新的 Docker 容器
func (p *Pool) createContainer(ctx context.Context, language string) (string, error) {
	langCfg, err := config.GetLanguageConfig(language)
	if err != nil {
		return "", err
	}
	imageName := langCfg.Image
	log.Printf("[Pool] 正在创建新容器 (语言: %s, 镜像: %s)\n", language, imageName)

	resp, err := p.Sandbox.cli.ContainerCreate(ctx,
		&container.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "infinity"},
			Tty:   false,
			Labels: map[string]string{
				PoolLabelKey: PoolLabelValue,
			},
			Env: func() []string {
				if language == "go" {
					return []string{"CGO_ENABLED=0", "GOCACHE=/app/.cache"}
				}
				return nil
			}(),
		},
		// 限制容器资源 + 安全沙盒
		&container.HostConfig{
			Resources: container.Resources{
				Memory:    p.memoryBytes,
				NanoCPUs:  p.nanoCPUs,
				PidsLimit: func() *int64 { v := int64(100); return &v }(),
			},
			NetworkMode: "none", // 安全层1：无网络访问
			SecurityOpt: func() []string {
				// 安全层3：Seccomp 白名单
				if profile := getSeccompProfile(); profile != "" {
					return []string{"seccomp=" + profile}
				}
				return nil // 降级为 Docker 默认 Seccomp 策略
			}(),
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

	// 针对 Go 语言进行预热 (Warm-up) - 直接 exec 构建缓存
	if language == "go" {
		log.Printf("[Pool] 正在预热 Go 容器 %s (生成 build cache)...\n", shortID(resp.ID))
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

	log.Printf("[Pool] 新容器已创建并启动 %s (语言: %s)\n", shortID(resp.ID), language)
	return resp.ID, nil
}

// ============================================================
//  阶段三：基于高水位的弹性缩容 (Elastic Shrinking)
// ============================================================

// ReturnContainer 归还容器：基于 HWM 决定留用或销毁
// - N_idle < HWM → 清理环境后放回池中复用
// - N_idle >= HWM → 溢出资源，异步销毁
func (p *Pool) ReturnContainer(ctx context.Context, containerID string) {
	// 停止 eBPF 安全监控并获取收尾画像
	if profile := p.ebpfMonitor.StopMonitoring(containerID); profile != nil {
		profileJson, _ := json.Marshal(profile)
		log.Printf("[EBPF 审计结果] 容器 %s 运行画像: %s\n", shortID(containerID), string(profileJson))
	}

	// 查找容器语言
	p.mu.Lock()
	lang, exists := p.containerLang[containerID]
	p.mu.Unlock()

	if !exists {
		// 未知容器，直接销毁
		log.Printf("[Pool] 未知容器 %s，直接销毁\n", shortID(containerID))
		p.Sandbox.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
		return
	}

	// 1. 环境清理：rm -rf /app/* 重置工作区
	if err := p.cleanContainer(ctx, containerID); err != nil {
		log.Printf("[Pool] 清理容器 %s 失败: %v，销毁容器\n", shortID(containerID), err)
		p.destroyContainer(ctx, containerID)
		return
	}

	// 2. 水位判定 + 决策分流
	p.mu.Lock()
	defer p.mu.Unlock()

	currentIdle := len(p.Available[lang])

	if currentIdle < p.hwm {
		// 情况 A（留用）：池未满，推回池中复用
		p.Available[lang] = append(p.Available[lang], containerID)
		log.Printf("[Pool] 容器 %s 已归还 (语言: %s, 闲置: %d/%d HWM)\n",
			shortID(containerID), lang, len(p.Available[lang]), p.hwm)
	} else {
		// 情况 B（销毁）：水位已满（洪峰退潮），异步销毁溢出资源
		log.Printf("[Pool] ♻️ 弹性缩容: %s 水位已满 (闲置: %d >= HWM: %d)，销毁容器 %s\n",
			lang, currentIdle, p.hwm, shortID(containerID))
		delete(p.containerLang, containerID)
		go func(id string) {
			if err := p.Sandbox.cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Printf("[Pool] 异步删除容器 %s 失败: %v\n", shortID(id), err)
			}
		}(containerID)
	}
}

// cleanContainer 清理容器工作区
func (p *Pool) cleanContainer(ctx context.Context, containerID string) error {
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

// warmupGoContainer 通过直接 exec 运行空程序生成编译缓存
// 不走 Sandbox.Execute，避免潜在副作用
func (p *Pool) warmupGoContainer(ctx context.Context, containerID string) error {
	execCfg := types.ExecConfig{
		Cmd: []string{"sh", "-c",
			"cd /app && echo 'package main\nfunc main(){}' > _warmup.go && go build -o /dev/null _warmup.go && rm -f _warmup.go"},
	}
	resp, err := p.Sandbox.cli.ContainerExecCreate(ctx, containerID, execCfg)
	if err != nil {
		return err
	}
	return p.Sandbox.cli.ContainerExecStart(ctx, resp.ID, types.ExecStartCheck{})
}

// ============================================================
//  运维管理
// ============================================================

// CleanupStaleContainers 清理所有带有容器池标签的残留容器
// 应在程序启动时调用，以清除上次运行遗留的容器
func (p *Pool) CleanupStaleContainers(ctx context.Context) (int, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", PoolLabelKey+"="+PoolLabelValue)

	containers, err := p.Sandbox.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return 0, fmt.Errorf("列出容器失败: %w", err)
	}

	removedCount := 0
	for _, c := range containers {
		log.Printf("[Pool] 清理残留容器: %s (状态: %s)\n", shortID(c.ID), c.State)
		if err := p.Sandbox.cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Printf("[Pool] 删除容器 %s 失败: %v\n", shortID(c.ID), err)
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
// 先在锁内收集 ID 列表并清空状态，解锁后再批量删除（避免锁内网络调用）
func (p *Pool) ResetPool(ctx context.Context) (int, error) {
	p.mu.Lock()
	// 收集所有待删除的容器 ID
	var toRemove []string
	for _, ids := range p.Available {
		toRemove = append(toRemove, ids...)
	}
	// 立即清空池状态
	p.Available = make(map[string][]string)
	p.containerLang = make(map[string]string)
	p.mu.Unlock()

	// 锁外批量删除（不阻塞其他操作）
	removedCount := 0
	for _, id := range toRemove {
		log.Printf("[Pool] 重置: 销毁容器 %s\n", shortID(id))
		if err := p.Sandbox.cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{Force: true}); err != nil {
			log.Printf("[Pool] 删除容器 %s 失败: %v\n", shortID(id), err)
		} else {
			removedCount++
		}
	}

	log.Printf("[Pool] 容器池已重置，共销毁 %d 个容器\n", removedCount)
	return removedCount, nil
}

// ============================================================
//  统计信息
// ============================================================

// PoolStats 容器池统计信息
type PoolStats struct {
	TotalContainers int                      `json:"total"`
	DynamicLWM      int                      `json:"dynamicLWM"` // 当前动态低水位线
	HWM             int                      `json:"hwm"`        // 高水位线
	MemoryMB        int                      `json:"memoryMB"`
	CPUCores        float64                  `json:"cpuCores"`
	MonitorInterval int                      `json:"monitorInterval"` // 监控间隔（秒）
	PredictorStats  PredictorStats           `json:"predictorStats"`  // EMA 预测器状态
	LangDetails     map[string]LangPoolStats `json:"details"`
}

// LangPoolStats 每种语言的容器池统计
type LangPoolStats struct {
	Idle   int    `json:"idle"`
	Active int    `json:"active"`
	State  string `json:"state"` // 水位状态：steady / prewarm / burst / recovery
}

// GetStats 获取容器池实时统计信息
func (p *Pool) GetStats() *PoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := &PoolStats{
		DynamicLWM:      p.predictor.GetDynamicLWM(),
		HWM:             p.hwm,
		MemoryMB:        int(p.memoryBytes / 1024 / 1024),
		CPUCores:        float64(p.nanoCPUs) / 1e9,
		MonitorInterval: int(p.monitorInterval / time.Second),
		PredictorStats:  p.predictor.GetStats(),
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

	// 组装详情（包括所有已知语言）
	allLangs := make(map[string]bool)
	for lang := range idleCounts {
		allLangs[lang] = true
	}
	for lang := range totalCounts {
		allLangs[lang] = true
	}
	for _, lang := range p.languages {
		allLangs[lang] = true
	}

	for lang := range allLangs {
		total := totalCounts[lang]
		idle := idleCounts[lang]
		active := total - idle
		if active < 0 {
			active = 0
		}

		// 判定水位状态
		state := "steady" // 平稳态：LWM <= N_idle <= HWM
		if idle == 0 && active > 0 {
			state = "burst" // 洪水态：池空但有活跃容器
		} else if idle < stats.DynamicLWM {
			state = "prewarm" // 枯水态：低于 LWM
		} else if idle >= p.hwm {
			state = "recovery" // 退水态：达到或超过 HWM
		}

		stats.LangDetails[lang] = LangPoolStats{
			Idle:   idle,
			Active: active,
			State:  state,
		}
	}

	return stats
}
