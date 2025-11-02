package forwarder

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// ClickHouseForwarder ClickHouse 转发器
type ClickHouseForwarder struct {
	config    *ForwarderConfig
	db        *sql.DB
	tableName string
	logger    *zap.Logger
	stats     Stats
	mu        sync.RWMutex
}

// NewClickHouseForwarder 创建 ClickHouse 转发器
func NewClickHouseForwarder(config *ForwarderConfig, logger *zap.Logger) (*ClickHouseForwarder, error) {
	if config.DSN == "" {
		return nil, fmt.Errorf("clickhouse DSN is required")
	}

	if config.Table == "" {
		config.Table = "metrics.data"
	}

	// 解析 DSN
	options, err := clickhouse.ParseDSN(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// 创建连接
	db := clickhouse.OpenDB(options)

	// 设置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	forwarder := &ClickHouseForwarder{
		config:    config,
		db:        db,
		tableName: config.Table,
		logger:    logger,
	}

	// 确保表存在
	if err := forwarder.ensureTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}

	logger.Info("ClickHouse forwarder initialized",
		zap.String("name", config.Name),
		zap.String("table", config.Table))

	return forwarder, nil
}

// ensureTable 确保表存在
func (f *ClickHouseForwarder) ensureTable() error {
	// 创建数据库（如果不存在）
	createDBSQL := `CREATE DATABASE IF NOT EXISTS metrics`
	if _, err := f.db.Exec(createDBSQL); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// 创建表（如果不存在）
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime64(3),
			metric_name String,
			metric_value Float64,
			metric_type String,
			device_id String,
			sentinel_id String,
			labels Map(String, String),
			date Date DEFAULT toDate(timestamp)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (metric_name, device_id, timestamp)
		TTL date + INTERVAL 90 DAY
		SETTINGS index_granularity = 8192
	`, f.tableName)

	if _, err := f.db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Write 写入指标数据
func (f *ClickHouseForwarder) Write(metrics []*Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), f.config.Timeout)
	defer cancel()

	// 开始事务
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		f.recordError()
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 准备批量插入语句
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (timestamp, metric_name, metric_value, metric_type, device_id, sentinel_id, labels)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, f.tableName)

	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		f.recordError()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 批量插入
	totalBytes := int64(0)
	for _, metric := range metrics {
		// 转换时间戳
		timestamp := time.Unix(metric.Timestamp, 0)
		if metric.Timestamp == 0 {
			timestamp = time.Now()
		}

		// 获取 device_id 和 sentinel_id
		deviceID := metric.Labels["device_id"]
		if deviceID == "" {
			deviceID = "unknown"
		}

		sentinelID := metric.Labels["sentinel_id"]
		if sentinelID == "" {
			sentinelID = "unknown"
		}

		// 执行插入
		_, err := stmt.ExecContext(ctx,
			timestamp,
			metric.Name,
			metric.Value,
			metric.Type,
			deviceID,
			sentinelID,
			metric.Labels,
		)
		if err != nil {
			f.recordError()
			return fmt.Errorf("failed to insert metric: %w", err)
		}

		// 估算数据大小
		totalBytes += int64(len(metric.Name) + 8 + 8) // name + value + timestamp
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		f.recordError()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 记录成功
	latency := time.Since(startTime).Milliseconds()
	f.recordSuccess(totalBytes, latency)

	f.logger.Debug("Wrote to ClickHouse",
		zap.String("forwarder", f.config.Name),
		zap.Int("metrics", len(metrics)),
		zap.Int64("bytes", totalBytes),
		zap.Int64("latency_ms", latency))

	return nil
}

// Close 关闭转发器
func (f *ClickHouseForwarder) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}

// Name 获取转发器名称
func (f *ClickHouseForwarder) Name() string {
	return f.config.Name
}

// Type 获取转发器类型
func (f *ClickHouseForwarder) Type() ForwarderType {
	return ForwarderTypeClickHouse
}

// IsEnabled 是否启用
func (f *ClickHouseForwarder) IsEnabled() bool {
	return f.config.Enabled
}

// GetStats 获取统计信息
func (f *ClickHouseForwarder) GetStats() Stats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.stats
}

// Query 查询数据（用于验证和调试）
func (f *ClickHouseForwarder) Query(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := f.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// 读取结果
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

// recordSuccess 记录成功
func (f *ClickHouseForwarder) recordSuccess(bytes int64, latencyMs int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.stats.SuccessCount++
	f.stats.TotalBytes += bytes
	f.stats.LastSuccess = time.Now()

	// 计算平均延迟
	if f.stats.AvgLatencyMs == 0 {
		f.stats.AvgLatencyMs = latencyMs
	} else {
		f.stats.AvgLatencyMs = (f.stats.AvgLatencyMs + latencyMs) / 2
	}
}

// recordError 记录错误
func (f *ClickHouseForwarder) recordError() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stats.FailedCount++
}

