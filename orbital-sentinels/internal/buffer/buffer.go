package buffer

import (
	"github.com/celestial/orbital-sentinels/internal/plugin"
)

// Buffer 缓冲队列接口
type Buffer interface {
	// Push 推入数据
	Push(metrics []*plugin.Metric) error

	// Pop 弹出数据
	Pop(count int) ([]*plugin.Metric, error)

	// Size 获取当前大小
	Size() int

	// Close 关闭缓冲区
	Close() error
}
