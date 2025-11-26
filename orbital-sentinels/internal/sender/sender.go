package sender

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/celestial/orbital-sentinels/internal/buffer"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// SendMode 发送模式
type SendMode string

const (
	SendModeCore   SendMode = "core"   // 发送到中心端
	SendModeDirect SendMode = "direct" // 直连数据库
	SendModeHybrid SendMode = "hybrid" // 混合模式
)

// Config 发送器配置
type Config struct {
	Mode          SendMode
	BatchSize     int
	FlushInterval time.Duration
	Timeout       time.Duration
	RetryTimes    int
	RetryInterval time.Duration
}

// Sender 数据发送器
type Sender struct {
	config       *Config
	mode         SendMode
	buffer       buffer.Buffer
	coreSender   *CoreSender
	directSender *DirectSender
	successCount atomic.Int64
	failedCount  atomic.Int64
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewSender 创建发送器
func NewSender(config *Config, buf buffer.Buffer) *Sender {
	return &Sender{
		config: config,
		mode:   config.Mode,
		buffer: buf,
	}
}

// SetCoreSender 设置中心端发送器
func (s *Sender) SetCoreSender(coreSender *CoreSender) {
	s.coreSender = coreSender
}

// GetCoreSender 获取中心端发送器
func (s *Sender) GetCoreSender() *CoreSender {
	return s.coreSender
}

// SetDirectSender 设置直连发送器
func (s *Sender) SetDirectSender(directSender *DirectSender) {
	s.directSender = directSender
}

// Start 启动发送器
func (s *Sender) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	go s.flushLoop()

	logger.Info("Sender started", zap.String("mode", string(s.mode)))
}

// Stop 停止发送器
func (s *Sender) Stop() {
	if s.cancel != nil {
		s.cancel()
	}

	// 最后一次刷新
	s.flush()

	logger.Info("Sender stopped",
		zap.Int64("success_count", s.successCount.Load()),
		zap.Int64("failed_count", s.failedCount.Load()))
}

// flushLoop 刷新循环
func (s *Sender) flushLoop() {
	// 确保 FlushInterval 有效（至少 1 秒）
	flushInterval := s.config.FlushInterval
	if flushInterval <= 0 {
		logger.Warn("FlushInterval is zero or negative, using default 10s",
			zap.Duration("configured", flushInterval))
		flushInterval = 10 * time.Second
	}

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.flush()
		case <-s.ctx.Done():
			return
		}
	}
}

// flush 刷新缓冲区
func (s *Sender) flush() {
	metrics, err := s.buffer.Pop(s.config.BatchSize)
	if err != nil || len(metrics) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.config.Timeout)
	defer cancel()

	switch s.mode {
	case SendModeCore:
		s.sendToCore(ctx, metrics)
	case SendModeDirect:
		s.sendDirect(ctx, metrics)
	case SendModeHybrid:
		s.sendHybrid(ctx, metrics)
	}
}

// sendToCore 发送到中心端
func (s *Sender) sendToCore(ctx context.Context, metrics []*plugin.Metric) {
	if s.coreSender == nil {
		logger.Error("Core sender not configured")
		s.failedCount.Add(int64(len(metrics)))
		return
	}

	if err := s.coreSender.Send(ctx, metrics); err != nil {
		logger.Error("Failed to send to core", zap.Error(err))
		// 重新放回缓冲区
		s.buffer.Push(metrics)
		s.failedCount.Add(int64(len(metrics)))
		return
	}

	s.successCount.Add(int64(len(metrics)))
	logger.Debug("Sent to core", zap.Int("count", len(metrics)))
}

// sendDirect 直连发送
func (s *Sender) sendDirect(ctx context.Context, metrics []*plugin.Metric) {
	if s.directSender == nil {
		logger.Error("Direct sender not configured")
		s.failedCount.Add(int64(len(metrics)))
		return
	}

	if err := s.directSender.Send(ctx, metrics); err != nil {
		logger.Error("Failed to send direct", zap.Error(err))
		s.buffer.Push(metrics)
		s.failedCount.Add(int64(len(metrics)))
		return
	}

	s.successCount.Add(int64(len(metrics)))
	logger.Debug("Sent direct", zap.Int("count", len(metrics)))
}

// sendHybrid 混合模式发送
func (s *Sender) sendHybrid(ctx context.Context, metrics []*plugin.Metric) {
	// 简化实现：所有数据都发送到两个目标
	// 实际可以根据标签或指标名称进行分离

	// 发送到中心端
	go s.sendToCore(ctx, metrics)

	// 发送到直连
	go s.sendDirect(ctx, metrics)
}

// GetStats 获取统计信息
func (s *Sender) GetStats() (success, failed int64) {
	return s.successCount.Load(), s.failedCount.Load()
}
