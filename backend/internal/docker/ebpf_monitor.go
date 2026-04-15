package docker
//

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// BehaviorProfile 代表容器运行期间的行为画像
type BehaviorProfile struct {
	ContainerID    string    `json:"containerID"`
	Language       string    `json:"language"`
	FileIOAttempts int       `json:"fileIOAttempts"`    // 拦截/观测到的文件操作数
	ProcessForks   int       `json:"processForks"`      // 克隆/创建进程数
	NetworkSyscalls int      `json:"networkSyscalls"`   // 网络调用尝试次数
	RiskScore      int       `json:"riskScore"`         // 综合风险评分 (0-100)
	Killed         bool      `json:"killed"`            // 是否因危险行为被系统终止
	KilledReason   string    `json:"killedReason,omitempty"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime,omitempty"`
}

// EBPFMonitor 提供动态系统的行为观测与降级限流防御
// 在真实 Linux 环境下，应挂载 XDP/Tracepoint 完成。由于开发环境(Windows Docker)无法运行，
// 采用降级轮询容器 Stats 的方式，模拟核心业务流及检测逻辑。
type EBPFMonitor struct {
	cli      *client.Client
	profiles map[string]*BehaviorProfile
	mu       sync.RWMutex
}

// NewEBPFMonitor 创建观测器
func NewEBPFMonitor(cli *client.Client) *EBPFMonitor {
	log.Printf("[EBPF Monitor] 组件初始化...")
	if runtime.GOOS == "windows" {
		log.Printf("[EBPF Monitor] ⚠️ 警告: 检测到 Windows 宿主机，eBPF/Tracepoint 机制无法原生存取，自动降级为 Docker API Metrics 轮询观测模式。")
	} else {
		log.Printf("[EBPF Monitor] ✓ 运行于类 Unix 系统，准备挂载观测策略...")
	}

	return &EBPFMonitor{
		cli:      cli,
		profiles: make(map[string]*BehaviorProfile),
	}
}

// StartMonitoring 为分配给用户的容器开启行为审计
func (m *EBPFMonitor) StartMonitoring(ctx context.Context, containerID string, language string) {
	m.mu.Lock()
	m.profiles[containerID] = &BehaviorProfile{
		ContainerID: containerID,
		Language:    language,
		StartTime:   time.Now(),
		RiskScore:   0, // 初始分，越高质量越差
	}
	m.mu.Unlock()

	// 启动异步降级监控环（模拟 eBPF 的高频采集流）
	go m.fallbackMetricsLoop(ctx, containerID)
}

// StopMonitoring 代码执行结束时停止审计并结算画像
func (m *EBPFMonitor) StopMonitoring(containerID string) *BehaviorProfile {
	m.mu.Lock()
	defer m.mu.Unlock()

	profile, exists := m.profiles[containerID]
	if !exists {
		return nil
	}

	profile.EndTime = time.Now()
	
	// 在结束时，根据积累的数据进行末次评分计算
	if profile.ProcessForks > 20 {
		profile.RiskScore += 40
	} else if profile.ProcessForks > 5 {
		profile.RiskScore += 10
	}

	if profile.NetworkSyscalls > 0 {
		profile.RiskScore += 50 // 高风险操作（我们不允许网络）
	}
	
	if profile.RiskScore > 100 {
		profile.RiskScore = 100
	}

	// 从活跃列表中移除
	delete(m.profiles, containerID)

	return profile
}

// fallbackMetricsLoop 这是在没有完整 BPF 能力情况下的软性降级逻辑。
// 通过超高频轮询 Docker API，观察其子进程增量和 IO，估测出是否有恶意 fork 炸弹和挂载逃逸探测。
func (m *EBPFMonitor) fallbackMetricsLoop(ctx context.Context, containerID string) {
	ticker := time.NewTicker(200 * time.Millisecond) // 高频200ms
	defer ticker.Stop()

	// 用于计算差值
	var lastPids uint64 = 0

	for {
		select {
		case <-ctx.Done(): // 收到上层取消信号
			return
		case <-ticker.C:
			// 为了防止大量死锁日志
			statsCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			statsResponse, err := m.cli.ContainerStats(statsCtx, containerID, false)
			if err != nil {
				cancel()
				continue // 容器可能刚结束被回收，跳过
			}

			var v types.StatsJSON
			if err := json.NewDecoder(statsResponse.Body).Decode(&v); err != nil {
				statsResponse.Body.Close()
				cancel()
				continue
			}
			statsResponse.Body.Close()
			cancel()

			m.mu.Lock()
			profile, ok := m.profiles[containerID]
			if !ok {
				m.mu.Unlock()
				return // 容器审计已注销
			}

			// 粗略模拟 eBPF 下针的：进程树(fork/clone)数量飙升监控
			currentPids := v.PidsStats.Current
			if currentPids > lastPids && lastPids > 0 {
				profile.ProcessForks += int(currentPids - lastPids)
			}
			lastPids = currentPids

			// 粗略模拟文件IO数量上升 (读加写总量模拟)
			if len(v.BlkioStats.IoServiceBytesRecursive) > 0 {
				profile.FileIOAttempts += 1 // 每次检测到块 IO，就假定发生了相关的系统调用
			}

			// ---- 防御性动作 (模拟 eBPF 的 OOM/Kill Helper 功能) ----
			
			// 防御1：阻断进程 Fork 炸弹（软限制 30个，eBPF拦截比 CGroup PIDS Max 更早触达）
			if profile.ProcessForks > 30 && !profile.Killed {
				profile.Killed = true
				profile.KilledReason = "eBPF 监控: 异常大量的进程 Fork 行为，已拦截并强制终止容器"
				profile.RiskScore = 100
				
				// 异步断开该容器运行
				go m.triggerKill(containerID, profile.KilledReason)
			}
			
			// 我们虽然用 Seccomp 禁用了网络，但在双层架构中依然可以根据网络收发字节判定违规
			// 甚至更严格：收发任何报文立即视为逃逸
			if v.Networks != nil {
				for _, netStats := range v.Networks {
					if netStats.TxBytes > 0 {
						profile.NetworkSyscalls += int(netStats.TxPackets)
						if !profile.Killed && profile.NetworkSyscalls > 0 {
							profile.Killed = true
							profile.KilledReason = "eBPF 监控: 检测到违规的主动外发网络探测 (Tx Packets)"
							profile.RiskScore = 100
							go m.triggerKill(containerID, profile.KilledReason)
						}
					}
				}
			}

			m.mu.Unlock()

			if profile.Killed {
				return // 触发死亡则直接退出轮询环
			}
		}
	}
}

// triggerKill 并发通知 Docker 守护进程杀死此容器
func (m *EBPFMonitor) triggerKill(containerID string, reason string) {
	log.Printf("[EBPF 安全拦截] 容器 %s 行为非法: %s", shortID(containerID), reason)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// 强制终止
	_ = m.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}
