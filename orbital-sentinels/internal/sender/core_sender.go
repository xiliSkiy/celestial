package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// CoreSender 中心端发送器
type CoreSender struct {
	client  *http.Client
	url     string
	token   string
	breaker *CircuitBreaker
}

// NewCoreSender 创建中心端发送器
func NewCoreSender(url, token string, timeout time.Duration) *CoreSender {
	return &CoreSender{
		client: &http.Client{
			Timeout: timeout,
		},
		url:     url,
		token:   token,
		breaker: NewCircuitBreaker(5, 30*time.Second),
	}
}

// Send 发送数据
func (cs *CoreSender) Send(ctx context.Context, metrics []*plugin.Metric) error {
	// 熔断检查
	if !cs.breaker.Allow() {
		return fmt.Errorf("circuit breaker open")
	}

	// 序列化
	data, err := json.Marshal(map[string]interface{}{
		"metrics": metrics,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// 压缩
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(data); err != nil {
		return fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", cs.url+"/api/v1/data/ingest", &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-API-Token", cs.token)

	// 发送请求
	resp, err := cs.client.Do(req)
	if err != nil {
		cs.breaker.RecordFailure()
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		cs.breaker.RecordFailure()
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	cs.breaker.RecordSuccess()
	logger.Debug("Sent to core successfully", zap.Int("metrics", len(metrics)))

	return nil
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	state        CircuitState
	failures     int
	lastFailTime time.Time
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// Allow 是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否可以尝试恢复
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		logger.Warn("Circuit breaker opened", zap.Int("failures", cb.failures))
	}
}
