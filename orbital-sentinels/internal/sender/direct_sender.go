package sender

import (
	"context"
	"fmt"
	"sync"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// DirectSender 直连发送器
type DirectSender struct {
	prometheus *PrometheusWriter
	victoria   *VictoriaMetricsWriter
	clickhouse *ClickHouseWriter
}

// NewDirectSender 创建直连发送器
func NewDirectSender() *DirectSender {
	return &DirectSender{}
}

// SetPrometheusWriter 设置 Prometheus 写入器
func (ds *DirectSender) SetPrometheusWriter(writer *PrometheusWriter) {
	ds.prometheus = writer
}

// SetVictoriaMetricsWriter 设置 VictoriaMetrics 写入器
func (ds *DirectSender) SetVictoriaMetricsWriter(writer *VictoriaMetricsWriter) {
	ds.victoria = writer
}

// SetClickHouseWriter 设置 ClickHouse 写入器
func (ds *DirectSender) SetClickHouseWriter(writer *ClickHouseWriter) {
	ds.clickhouse = writer
}

// Send 发送数据
func (ds *DirectSender) Send(ctx context.Context, metrics []*plugin.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	// 并发发送到各个目标
	if ds.prometheus != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ds.prometheus.Write(ctx, metrics); err != nil {
				logger.Error("Failed to write to Prometheus", zap.Error(err))
				errChan <- fmt.Errorf("prometheus: %w", err)
			}
		}()
	}

	if ds.victoria != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ds.victoria.Write(ctx, metrics); err != nil {
				logger.Error("Failed to write to VictoriaMetrics", zap.Error(err))
				errChan <- fmt.Errorf("victoria-metrics: %w", err)
			}
		}()
	}

	if ds.clickhouse != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ds.clickhouse.Write(ctx, metrics); err != nil {
				logger.Error("Failed to write to ClickHouse", zap.Error(err))
				errChan <- fmt.Errorf("clickhouse: %w", err)
			}
		}()
	}

	// 等待所有发送完成
	wg.Wait()
	close(errChan)

	// 收集错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// 如果有错误，返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}

	logger.Debug("Sent to direct targets",
		zap.Int("metrics", len(metrics)),
		zap.Bool("prometheus", ds.prometheus != nil),
		zap.Bool("victoria", ds.victoria != nil),
		zap.Bool("clickhouse", ds.clickhouse != nil))

	return nil
}

// Close 关闭所有写入器
func (ds *DirectSender) Close() error {
	if ds.clickhouse != nil {
		return ds.clickhouse.Close()
	}
	return nil
}
