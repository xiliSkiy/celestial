# 数据转发模块使用指南

## 概述

数据转发模块负责接收来自 Sentinel 的指标数据，并将其转发到各种时序数据库（Prometheus、VictoriaMetrics、ClickHouse）。

## 架构

```
Sentinel 上报数据
    ↓
HTTP API (/api/v1/data/ingest)
    ↓
Forwarder Manager (缓冲 + 批处理)
    ↓
├─→ Prometheus Forwarder (Remote Write)
├─→ VictoriaMetrics Forwarder (Remote Write)
└─→ ClickHouse Forwarder (Native TCP)
```

## 支持的转发器类型

### 1. Prometheus

使用 Prometheus Remote Write 协议转发数据。

**配置示例：**
```yaml
forwarder:
  targets:
    - name: "prometheus-prod"
      type: "prometheus"
      enabled: true
      endpoint: "http://prometheus:9090/api/v1/write"
      username: "admin"
      password: "password"
      timeout: 30s
      batch_size: 1000
```

**注意：** Prometheus 默认不支持 Remote Write 接收，需要配置或使用 VictoriaMetrics。

### 2. VictoriaMetrics

完全兼容 Prometheus Remote Write 协议，推荐使用。

**配置示例：**
```yaml
forwarder:
  targets:
    - name: "victoria-prod"
      type: "victoria-metrics"
      enabled: true
      endpoint: "http://victoria:8428/api/v1/write"
      timeout: 30s
      batch_size: 5000
```

**优势：**
- 原生支持 Remote Write 接收
- 高性能、低资源占用
- 完全兼容 Prometheus 查询语法

### 3. ClickHouse

使用 ClickHouse 原生 TCP 协议，适合大规模数据存储和分析。

**配置示例：**
```yaml
forwarder:
  targets:
    - name: "clickhouse-analytics"
      type: "clickhouse"
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics?username=default&password=password"
      table: "metrics.data"
      timeout: 30s
      batch_size: 10000
```

**数据表结构：**
```sql
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
SETTINGS index_granularity = 8192
```

## 配置说明

### 全局配置

```yaml
forwarder:
  buffer_size: 10000      # 缓冲区大小
  batch_size: 1000        # 批处理大小
  flush_interval: 10s     # 刷新间隔
  max_retries: 3          # 最大重试次数
  retry_interval: 5s      # 重试间隔
  targets: []             # 转发目标列表
```

### 转发目标配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 转发器名称（唯一） |
| type | string | 是 | 类型：prometheus / victoria-metrics / clickhouse |
| enabled | bool | 否 | 是否启用（默认 true） |
| endpoint | string | 条件 | HTTP 端点（Prometheus/VictoriaMetrics） |
| dsn | string | 条件 | 连接字符串（ClickHouse） |
| table | string | 否 | 表名（ClickHouse，默认 metrics.data） |
| username | string | 否 | 用户名 |
| password | string | 否 | 密码 |
| timeout | duration | 否 | 超时时间（默认 30s） |
| batch_size | int | 否 | 批处理大小 |

## API 接口

### 1. 数据接收

**接口：** `POST /api/v1/data/ingest`

**请求头：**
```
X-Sentinel-ID: sentinel-001
Authorization: Bearer <api_token>
```

**请求体：**
```json
{
  "metrics": [
    {
      "name": "cpu_usage",
      "value": 85.5,
      "type": "gauge",
      "labels": {
        "device_id": "server-001",
        "host": "web-server-1"
      },
      "timestamp": 1698765432
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "received": 1
  }
}
```

### 2. 转发器管理

#### 列出转发器
```bash
GET /api/v1/forwarders?enabled=true
```

#### 获取转发器详情
```bash
GET /api/v1/forwarders/{name}
```

#### 创建转发器
```bash
POST /api/v1/forwarders
Content-Type: application/json

{
  "name": "victoria-prod",
  "type": "victoria-metrics",
  "enabled": true,
  "endpoint": "http://victoria:8428/api/v1/write",
  "timeout_seconds": 30,
  "batch_size": 5000
}
```

#### 更新转发器
```bash
PUT /api/v1/forwarders/{name}
Content-Type: application/json

{
  "enabled": false
}
```

#### 删除转发器
```bash
DELETE /api/v1/forwarders/{name}
```

#### 重新加载配置
```bash
POST /api/v1/forwarders/reload
```

### 3. 统计信息

#### 获取单个转发器统计
```bash
GET /api/v1/forwarders/{name}/stats
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "name": "victoria-prod",
    "current": {
      "success_count": 1000,
      "failed_count": 5,
      "total_bytes": 1048576,
      "avg_latency_ms": 15,
      "last_success": "2025-11-02T12:00:00Z"
    },
    "history": [...],
    "buffer_status": {
      "used": 100,
      "capacity": 10000,
      "usage": 1.0
    }
  }
}
```

#### 获取所有转发器统计
```bash
GET /api/v1/forwarders/stats
```

## 使用示例

### 示例 1：配置 VictoriaMetrics 转发

1. **启动 VictoriaMetrics：**
```bash
docker run -d \
  --name victoria \
  -p 8428:8428 \
  victoriametrics/victoria-metrics:latest
```

2. **配置 Gravital Core：**
```yaml
forwarder:
  buffer_size: 10000
  batch_size: 1000
  flush_interval: 10s
  targets:
    - name: "victoria-local"
      type: "victoria-metrics"
      enabled: true
      endpoint: "http://localhost:8428/api/v1/write"
      timeout: 30s
      batch_size: 5000
```

3. **启动 Gravital Core：**
```bash
./bin/gravital-core -c config/config.yaml
```

4. **验证数据写入：**
```bash
# 查询指标
curl 'http://localhost:8428/api/v1/query?query=cpu_usage'
```

### 示例 2：配置 ClickHouse 转发

1. **启动 ClickHouse：**
```bash
docker run -d \
  --name clickhouse \
  -p 9000:9000 \
  -p 8123:8123 \
  clickhouse/clickhouse-server:latest
```

2. **创建数据库和表：**
```sql
CREATE DATABASE IF NOT EXISTS metrics;

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
TTL date + INTERVAL 90 DAY;
```

3. **配置 Gravital Core：**
```yaml
forwarder:
  targets:
    - name: "clickhouse-local"
      type: "clickhouse"
      enabled: true
      dsn: "tcp://localhost:9000/metrics"
      table: "metrics.data"
      timeout: 30s
      batch_size: 10000
```

4. **查询数据：**
```sql
SELECT 
    metric_name,
    avg(metric_value) as avg_value,
    count() as count
FROM metrics.data
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY metric_name
ORDER BY count DESC
LIMIT 10;
```

### 示例 3：多目标转发

同时转发到 VictoriaMetrics 和 ClickHouse：

```yaml
forwarder:
  buffer_size: 10000
  batch_size: 1000
  flush_interval: 10s
  targets:
    # VictoriaMetrics - 用于实时查询
    - name: "victoria-prod"
      type: "victoria-metrics"
      enabled: true
      endpoint: "http://victoria:8428/api/v1/write"
      batch_size: 5000
    
    # ClickHouse - 用于长期存储和分析
    - name: "clickhouse-analytics"
      type: "clickhouse"
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics"
      table: "metrics.data"
      batch_size: 10000
```

## 性能优化

### 1. 批处理配置

- **buffer_size**: 设置足够大的缓冲区，避免数据丢失
- **batch_size**: 根据目标数据库性能调整批次大小
  - VictoriaMetrics: 5000-10000
  - ClickHouse: 10000-50000
  - Prometheus: 1000-5000
- **flush_interval**: 平衡延迟和吞吐量
  - 实时场景: 5-10s
  - 批处理场景: 30-60s

### 2. 网络优化

- 使用内网连接，减少延迟
- 启用 HTTP/2（VictoriaMetrics）
- 使用连接池

### 3. 资源配置

```yaml
# 高吞吐量场景
forwarder:
  buffer_size: 50000
  batch_size: 10000
  flush_interval: 5s
  
# 低延迟场景
forwarder:
  buffer_size: 10000
  batch_size: 1000
  flush_interval: 1s
```

## 监控指标

转发器会自动记录以下指标：

- `success_count`: 成功转发的批次数
- `failed_count`: 失败的批次数
- `total_bytes`: 总传输字节数
- `avg_latency_ms`: 平均延迟（毫秒）
- `buffer_usage`: 缓冲区使用率

可以通过 API 查询这些指标：

```bash
curl http://localhost:8080/api/v1/forwarders/stats
```

## 故障排查

### 1. 数据未写入

**检查转发器状态：**
```bash
curl http://localhost:8080/api/v1/forwarders
```

**查看日志：**
```bash
tail -f logs/gravital-core.log | grep forwarder
```

### 2. 高延迟

- 检查网络连接
- 增加 batch_size
- 增加 timeout
- 检查目标数据库性能

### 3. 数据丢失

- 增加 buffer_size
- 检查 flush_interval
- 启用持久化（未来功能）

### 4. ClickHouse 连接失败

```bash
# 测试连接
clickhouse-client --host localhost --port 9000

# 检查表是否存在
SHOW TABLES FROM metrics;
```

### 5. VictoriaMetrics 写入失败

```bash
# 检查 VictoriaMetrics 状态
curl http://localhost:8428/metrics

# 测试写入
curl -X POST http://localhost:8428/api/v1/write \
  -H 'Content-Type: application/x-protobuf' \
  --data-binary @test.pb
```

## 最佳实践

1. **使用 VictoriaMetrics 作为主要时序数据库**
   - 性能优异，资源占用低
   - 原生支持 Remote Write
   - 完全兼容 Prometheus

2. **ClickHouse 用于长期存储**
   - 90 天以上的历史数据
   - 复杂的聚合查询
   - 数据分析和报表

3. **配置合理的批处理参数**
   - 根据数据量调整 batch_size
   - 根据延迟要求调整 flush_interval
   - 预留足够的 buffer_size

4. **监控转发器状态**
   - 定期检查统计信息
   - 设置告警规则
   - 关注缓冲区使用率

5. **使用多目标转发实现高可用**
   - 主备转发器
   - 不同数据库用于不同用途
   - 数据冗余备份

## 相关文档

- [Prometheus Remote Write 规范](https://prometheus.io/docs/concepts/remote_write_spec/)
- [VictoriaMetrics 文档](https://docs.victoriametrics.com/)
- [ClickHouse 文档](https://clickhouse.com/docs/)
- [Gravital Core API 文档](./API.md)

