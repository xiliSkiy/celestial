package forwarder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Manager 转发管理器
type Manager struct {
	forwarders map[string]Forwarder
	buffer     chan *Metric
	batchSize  int
	flushInterval time.Duration
	logger     *zap.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// ManagerConfig 管理器配置
type ManagerConfig struct {
	BufferSize    int
	BatchSize     int
	FlushInterval time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// NewManager 创建转发管理器
func NewManager(config *ManagerConfig, logger *zap.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	if config.BufferSize == 0 {
		config.BufferSize = 10000
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = 10 * time.Second
	}

	return &Manager{
		forwarders:    make(map[string]Forwarder),
		buffer:        make(chan *Metric, config.BufferSize),
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// AddForwarder 添加转发器
func (m *Manager) AddForwarder(forwarder Forwarder) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := forwarder.Name()
	if _, exists := m.forwarders[name]; exists {
		return fmt.Errorf("forwarder %s already exists", name)
	}

	m.forwarders[name] = forwarder
	m.logger.Info("Added forwarder",
		zap.String("name", name),
		zap.String("type", string(forwarder.Type())))

	return nil
}

// RemoveForwarder 移除转发器
func (m *Manager) RemoveForwarder(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	forwarder, exists := m.forwarders[name]
	if !exists {
		return fmt.Errorf("forwarder %s not found", name)
	}

	if err := forwarder.Close(); err != nil {
		m.logger.Error("Failed to close forwarder",
			zap.String("name", name),
			zap.Error(err))
	}

	delete(m.forwarders, name)
	m.logger.Info("Removed forwarder", zap.String("name", name))

	return nil
}

// GetForwarder 获取转发器
func (m *Manager) GetForwarder(name string) (Forwarder, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	forwarder, exists := m.forwarders[name]
	return forwarder, exists
}

// ListForwarders 列出所有转发器
func (m *Manager) ListForwarders() []Forwarder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	forwarders := make([]Forwarder, 0, len(m.forwarders))
	for _, f := range m.forwarders {
		forwarders = append(forwarders, f)
	}
	return forwarders
}

// Forward 转发单个指标
func (m *Manager) Forward(metric *Metric) error {
	select {
	case m.buffer <- metric:
		return nil
	case <-m.ctx.Done():
		return fmt.Errorf("manager is stopped")
	default:
		return fmt.Errorf("buffer is full")
	}
}

// ForwardBatch 批量转发指标
func (m *Manager) ForwardBatch(metrics []*Metric) error {
	for _, metric := range metrics {
		if err := m.Forward(metric); err != nil {
			return err
		}
	}
	return nil
}

// Start 启动管理器
func (m *Manager) Start() {
	m.wg.Add(1)
	go m.processLoop()
	m.logger.Info("Forwarder manager started",
		zap.Int("buffer_size", cap(m.buffer)),
		zap.Int("batch_size", m.batchSize),
		zap.Duration("flush_interval", m.flushInterval))
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()

	// 关闭所有转发器
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, forwarder := range m.forwarders {
		if err := forwarder.Close(); err != nil {
			m.logger.Error("Failed to close forwarder",
				zap.String("name", name),
				zap.Error(err))
		}
	}

	m.logger.Info("Forwarder manager stopped")
}

// processLoop 处理循环
func (m *Manager) processLoop() {
	defer m.wg.Done()

	batch := make([]*Metric, 0, m.batchSize)
	ticker := time.NewTicker(m.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case metric := <-m.buffer:
			batch = append(batch, metric)
			if len(batch) >= m.batchSize {
				m.flush(batch)
				batch = make([]*Metric, 0, m.batchSize)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				m.flush(batch)
				batch = make([]*Metric, 0, m.batchSize)
			}

		case <-m.ctx.Done():
			// 刷新剩余数据
			if len(batch) > 0 {
				m.flush(batch)
			}
			return
		}
	}
}

// flush 刷新批次数据
func (m *Manager) flush(batch []*Metric) {
	if len(batch) == 0 {
		return
	}

	m.mu.RLock()
	forwarders := make([]Forwarder, 0, len(m.forwarders))
	for _, f := range m.forwarders {
		if f.IsEnabled() {
			forwarders = append(forwarders, f)
		}
	}
	m.mu.RUnlock()

	// 并发转发到所有启用的转发器
	var wg sync.WaitGroup
	for _, forwarder := range forwarders {
		wg.Add(1)
		go func(f Forwarder) {
			defer wg.Done()

			if err := f.Write(batch); err != nil {
				m.logger.Error("Failed to forward metrics",
					zap.String("forwarder", f.Name()),
					zap.String("type", string(f.Type())),
					zap.Int("metrics", len(batch)),
					zap.Error(err))
			}
		}(forwarder)
	}

	wg.Wait()

	m.logger.Debug("Flushed metrics batch",
		zap.Int("metrics", len(batch)),
		zap.Int("forwarders", len(forwarders)))
}

// GetStats 获取所有转发器的统计信息
func (m *Manager) GetStats() map[string]Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]Stats)
	for name, forwarder := range m.forwarders {
		// 使用类型断言获取统计信息
		switch f := forwarder.(type) {
		case *PrometheusForwarder:
			stats[name] = f.GetStats()
		case *VictoriaMetricsForwarder:
			stats[name] = f.GetStats()
		case *ClickHouseForwarder:
			stats[name] = f.GetStats()
		}
	}
	return stats
}

// GetBufferStatus 获取缓冲区状态
func (m *Manager) GetBufferStatus() (used int, capacity int) {
	return len(m.buffer), cap(m.buffer)
}

