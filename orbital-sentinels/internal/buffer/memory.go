package buffer

import (
	"fmt"
	"sync"

	"github.com/celestial/orbital-sentinels/internal/plugin"
)

// MemoryBuffer 内存缓冲实现
type MemoryBuffer struct {
	queue    []*plugin.Metric
	maxSize  int
	mu       sync.Mutex
	notEmpty *sync.Cond
	closed   bool
}

// NewMemoryBuffer 创建内存缓冲
func NewMemoryBuffer(maxSize int) *MemoryBuffer {
	mb := &MemoryBuffer{
		queue:   make([]*plugin.Metric, 0, maxSize),
		maxSize: maxSize,
	}
	mb.notEmpty = sync.NewCond(&mb.mu)
	return mb
}

// Push 推入数据
func (mb *MemoryBuffer) Push(metrics []*plugin.Metric) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.closed {
		return fmt.Errorf("buffer is closed")
	}

	// 检查容量
	if len(mb.queue)+len(metrics) > mb.maxSize {
		// 策略：丢弃最旧的数据以腾出空间
		overflow := len(mb.queue) + len(metrics) - mb.maxSize
		if overflow > 0 {
			if overflow >= len(mb.queue) {
				// 如果溢出量大于等于当前队列长度，清空队列
				mb.queue = mb.queue[:0]
			} else {
				// 否则删除最旧的数据
				mb.queue = mb.queue[overflow:]
			}
		}
	}

	mb.queue = append(mb.queue, metrics...)
	mb.notEmpty.Signal()

	return nil
}

// Pop 弹出数据
func (mb *MemoryBuffer) Pop(count int) ([]*plugin.Metric, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// 如果没有数据且未关闭，等待
	for len(mb.queue) == 0 && !mb.closed {
		mb.notEmpty.Wait()
	}

	if mb.closed && len(mb.queue) == 0 {
		return nil, fmt.Errorf("buffer is closed")
	}

	if count > len(mb.queue) {
		count = len(mb.queue)
	}

	if count == 0 {
		return []*plugin.Metric{}, nil
	}

	metrics := make([]*plugin.Metric, count)
	copy(metrics, mb.queue[:count])
	mb.queue = mb.queue[count:]

	return metrics, nil
}

// Size 获取当前大小
func (mb *MemoryBuffer) Size() int {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return len(mb.queue)
}

// Close 关闭缓冲区
func (mb *MemoryBuffer) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.closed = true
	mb.notEmpty.Broadcast()

	return nil
}
