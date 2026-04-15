package docker

import (
	"sync"
	"sync/atomic"
	"time"
)

// TrafficPredictor 流量预测器，基于 EMA (指数移动平均) 动态计算安全和经济的水位线
type TrafficPredictor struct {
	// 配置参数
	alpha   float64 // 平滑系数 (0~1)
	lwmMin  int     // 最低低水位
	lwmMax  int     // 最高低水位
	redundancy float64 // 冗余系数 K

	// 状态变量
	requestsPerSecond atomic.Int32 // 当前秒的新增请求数
	ema               float64      // 指数移动平均值
	acceleration      float64      // 流量加速度（导数）
	currentLWM        atomic.Int32 /* uint32 不支持直接原子负数，统一用 Int32 */

	// 控制信号
	mu       sync.Mutex
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewTrafficPredictor 创建一个新的流量预测器
func NewTrafficPredictor(alpha float64, lwmMin, lwmMax int) *TrafficPredictor {
	// 参数校验与默认值
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3
	}
	if lwmMin < 1 {
		lwmMin = 1
	}
	if lwmMax < lwmMin {
		lwmMax = lwmMin + 5 // 至少留点空间
	}

	p := &TrafficPredictor{
		alpha:      alpha,
		lwmMin:     lwmMin,
		lwmMax:     lwmMax,
		redundancy: 1.5, // 默认冗余 1.5 倍
		stopChan:   make(chan struct{}),
	}
	p.currentLWM.Store(int32(lwmMin)) // 初始为最小值

	return p
}

// Start 启动后台预测协程
// interval: 计算间隔，通常为 1 秒
// notifyPrewarm: 当检测到需要激进扩容时，主动通知回调执行预热（防抖）
func (p *TrafficPredictor) Start(interval time.Duration, notifyPrewarm func()) {
	if interval <= 0 {
		interval = 1 * time.Second
	}
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.calculateEMA(notifyPrewarm)
			case <-p.stopChan:
				return
			}
		}
	}()
}

// Stop 停止预测协程
func (p *TrafficPredictor) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
}

// RecordRequest 记录一次新的请求（原子操作，极低开销）
func (p *TrafficPredictor) RecordRequest() {
	p.requestsPerSecond.Add(1)
}

// GetDynamicLWM 获取当前计算出的动态低水位线
func (p *TrafficPredictor) GetDynamicLWM() int {
	return int(p.currentLWM.Load())
}

// GetStats 获取预测器当前状态指标，用于展示或测试
type PredictorStats struct {
	CurrentEMA   float64 `json:"currentEMA"`
	Acceleration float64 `json:"acceleration"`
	DynamicLWM   int     `json:"dynamicLWM"`
	LastRPS      int     `json:"lastRPS"`
}

func (p *TrafficPredictor) GetStats() PredictorStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return PredictorStats{
		CurrentEMA:   p.ema,
		Acceleration: p.acceleration,
		DynamicLWM:   int(p.currentLWM.Load()),
		LastRPS:      int(p.requestsPerSecond.Load()), // 注意：这里获取的是当前正在累加的值，上一秒的在 calculateEMA 里清零了
	}
}

// calculateEMA 核心预测算法
func (p *TrafficPredictor) calculateEMA(notifyPrewarm func()) {
	// 获取并原子重置本周期的真实请求数 (R_t)
	rt := float64(p.requestsPerSecond.Swap(0))

	p.mu.Lock()
	emaPrev := p.ema

	// 计算 EMA
	if emaPrev == 0 && rt > 0{
		p.ema = rt // 初始化
	} else {
		// 公式: EMA_t = α * R_t + (1 - α) * EMA_{t-1}
		p.ema = p.alpha*rt + (1-p.alpha)*emaPrev
	}

	// 计算加速度 (一阶导数)
	p.acceleration = p.ema - emaPrev
	
	currentEma := p.ema
	currentAcc := p.acceleration
	p.mu.Unlock()

	// 根据加速度调整冗余系数 (K)
	// 如果流量在急剧上升 (加速度很大)，说明系统可能即将扛一波洪峰，把冗余倍数调高
	k := p.redundancy
	triggerAggressivePrewarm := false

	if currentAcc > 5.0 {
		// 加速度极高，流量暴增
		k = 2.5
		triggerAggressivePrewarm = true
	} else if currentAcc > 2.0 {
		k = 2.0
		triggerAggressivePrewarm = true
	} else if currentAcc < -2.0 {
		// 流量在快速下降，收紧冗余以回收资源
		k = 1.0
	} else {
		// 平稳期或微小波动
		k = 1.2
	}

	// 计算出目标水位: ceil(EMA * K)
	predicted := int(currentEma * k)
	
	// 为了防止空闲期 EMA 长期低于 1 导致归零，遇到任何 > 0 的迹象都保底 1 个
	if rt > 0 && predicted == 0 {
	    predicted = 1
	}

	// 限制上下限 (LWM_min <= predicted <= LWM_max)
	if predicted < p.lwmMin {
		predicted = p.lwmMin
	}
	if predicted > p.lwmMax {
		predicted = p.lwmMax
	}

	// 原子存入
	oldLWM := p.currentLWM.Swap(int32(predicted))

	// 判断是否需要向调度器发送激进预热信号
	// 只有在预测值显著抬升，且确实触发了激进条件时才发送信号，打破原本依赖定时器的预热周期
	if triggerAggressivePrewarm && predicted > int(oldLWM) && notifyPrewarm != nil {
		notifyPrewarm()
	}
}
