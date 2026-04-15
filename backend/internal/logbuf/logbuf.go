package logbuf

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogEntry 单条日志记录
type LogEntry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
}

// RingBuffer 线程安全的环形日志缓冲区
// 保留最近 capacity 条日志，溢出时覆盖最旧的记录
type RingBuffer struct {
	mu       sync.RWMutex
	entries  []LogEntry
	capacity int
	head     int  // 下一个写入位置
	full     bool // 缓冲区是否已满（开始环形覆盖）
}

// NewRingBuffer 创建一个新的环形缓冲区
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 500
	}
	return &RingBuffer{
		entries:  make([]LogEntry, capacity),
		capacity: capacity,
	}
}

// Write 实现 io.Writer 接口，供 log.SetOutput 使用
// 每次 Write 调用对应一条日志
func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	msg := string(p)
	// 去掉尾部换行符
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	entry := LogEntry{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Message: msg,
	}

	rb.mu.Lock()
	rb.entries[rb.head] = entry
	rb.head = (rb.head + 1) % rb.capacity
	if rb.head == 0 && !rb.full {
		rb.full = true
	}
	rb.mu.Unlock()

	return len(p), nil
}

// GetRecent 返回最近 n 条日志（按时间正序）
func (rb *RingBuffer) GetRecent(n int) []LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	// 计算实际存储的日志数量
	total := rb.head
	if rb.full {
		total = rb.capacity
	}

	if n <= 0 || n > total {
		n = total
	}

	result := make([]LogEntry, n)

	// 从最旧的开始读，取最后 n 条
	start := rb.head - n
	if start < 0 {
		if rb.full {
			start += rb.capacity
		} else {
			start = 0
			n = total
			result = make([]LogEntry, n)
		}
	}

	for i := 0; i < n; i++ {
		idx := (start + i) % rb.capacity
		result[i] = rb.entries[idx]
	}

	return result
}

// Count 返回当前缓冲区中的日志数量
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if rb.full {
		return rb.capacity
	}
	return rb.head
}

// Clear 清空缓冲区
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.full = false
}

// NewTeeWriter 创建一个同时写入 stdout 和环形缓冲区的 Writer
// 这样日志既能在终端/journalctl 中看到，也能通过 API 查询
func NewTeeWriter(buf *RingBuffer) io.Writer {
	return io.MultiWriter(os.Stdout, buf)
}

// Fprintf 格式化写入缓冲区（用于不想走标准 log 的场景）
func (rb *RingBuffer) Fprintf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	rb.Write([]byte(msg + "\n"))
}
