-- ClickHouse 初始化脚本

-- 创建数据库
CREATE DATABASE IF NOT EXISTS metrics;

-- 创建指标数据表
CREATE TABLE IF NOT EXISTS metrics.data (
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
SETTINGS index_granularity = 8192;

-- 创建物化视图：按小时聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.data_hourly
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (metric_name, device_id, hour)
AS SELECT
    toStartOfHour(timestamp) as hour,
    metric_name,
    device_id,
    sentinel_id,
    avg(metric_value) as avg_value,
    min(metric_value) as min_value,
    max(metric_value) as max_value,
    count() as count,
    toDate(hour) as date
FROM metrics.data
GROUP BY hour, metric_name, device_id, sentinel_id;

-- 创建物化视图：按天聚合
CREATE MATERIALIZED VIEW IF NOT EXISTS metrics.data_daily
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (metric_name, device_id, date)
AS SELECT
    toDate(timestamp) as date,
    metric_name,
    device_id,
    sentinel_id,
    avg(metric_value) as avg_value,
    min(metric_value) as min_value,
    max(metric_value) as max_value,
    count() as count
FROM metrics.data
GROUP BY date, metric_name, device_id, sentinel_id;

-- 创建索引
-- ALTER TABLE metrics.data ADD INDEX idx_device_id device_id TYPE bloom_filter GRANULARITY 1;
-- ALTER TABLE metrics.data ADD INDEX idx_metric_name metric_name TYPE bloom_filter GRANULARITY 1;

