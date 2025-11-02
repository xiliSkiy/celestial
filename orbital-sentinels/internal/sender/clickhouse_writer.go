package sender

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"go.uber.org/zap"
)

// ClickHouseWriter ClickHouse 写入器
type ClickHouseWriter struct {
	db        *sql.DB
	tableName string
	batchSize int
}

// ClickHouseConfig ClickHouse 配置
type ClickHouseConfig struct {
	DSN       string
	TableName string
	BatchSize int
}

// NewClickHouseWriter 创建 ClickHouse 写入器
func NewClickHouseWriter(config *ClickHouseConfig) (*ClickHouseWriter, error) {
	// 解析 DSN
	options, err := clickhouse.ParseDSN(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// 创建连接
	db := clickhouse.OpenDB(options)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	writer := &ClickHouseWriter{
		db:        db,
		tableName: config.TableName,
		batchSize: config.BatchSize,
	}

	// 确保表存在
	if err := writer.ensureTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure table: %w", err)
	}

	logger.Info("ClickHouse writer initialized",
		zap.String("table", config.TableName))

	return writer, nil
}

// ensureTable 确保表存在
func (cw *ClickHouseWriter) ensureTable() error {
	// 创建表的 SQL（如果不存在）
	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime64(3),
			metric_name String,
			metric_value Float64,
			metric_type String,
			device_id String,
			labels Map(String, String),
			date Date DEFAULT toDate(timestamp)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (metric_name, device_id, timestamp)
		TTL date + INTERVAL 90 DAY
		SETTINGS index_granularity = 8192
	`, cw.tableName)

	_, err := cw.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Write 写入数据
func (cw *ClickHouseWriter) Write(ctx context.Context, metrics []*plugin.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// 开始事务
	tx, err := cw.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 准备批量插入语句
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (timestamp, metric_name, metric_value, metric_type, device_id, labels)
		VALUES (?, ?, ?, ?, ?, ?)
	`, cw.tableName)

	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 批量插入
	for _, metric := range metrics {
		// 转换时间戳
		timestamp := time.Unix(metric.Timestamp, 0)

		// 获取 device_id
		deviceID := metric.Labels["device_id"]
		if deviceID == "" {
			deviceID = "unknown"
		}

		// 执行插入
		_, err := stmt.ExecContext(ctx,
			timestamp,
			metric.Name,
			metric.Value,
			string(metric.Type),
			deviceID,
			metric.Labels,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Debug("Wrote to ClickHouse",
		zap.Int("metrics", len(metrics)),
		zap.String("table", cw.tableName))

	return nil
}

// WriteBatch 批量写入（优化版本）
func (cw *ClickHouseWriter) WriteBatch(ctx context.Context, metrics []*plugin.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// 构建批量插入 SQL
	var valueStrings []string
	var valueArgs []interface{}

	for _, metric := range metrics {
		timestamp := time.Unix(metric.Timestamp, 0)
		deviceID := metric.Labels["device_id"]
		if deviceID == "" {
			deviceID = "unknown"
		}

		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			timestamp,
			metric.Name,
			metric.Value,
			string(metric.Type),
			deviceID,
			metric.Labels,
		)
	}

	// 执行批量插入
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (timestamp, metric_name, metric_value, metric_type, device_id, labels)
		VALUES %s
	`, cw.tableName, strings.Join(valueStrings, ","))

	_, err := cw.db.ExecContext(ctx, insertSQL, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to batch insert: %w", err)
	}

	logger.Debug("Batch wrote to ClickHouse",
		zap.Int("metrics", len(metrics)),
		zap.String("table", cw.tableName))

	return nil
}

// Close 关闭连接
func (cw *ClickHouseWriter) Close() error {
	if cw.db != nil {
		return cw.db.Close()
	}
	return nil
}

// Query 查询数据（用于验证）
func (cw *ClickHouseWriter) Query(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := cw.db.QueryContext(ctx, query)
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
