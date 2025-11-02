# 直连发送器使用指南

## 概述

Orbital Sentinels 支持将采集的监控数据直接发送到时序数据库，无需通过中心端转发。目前支持以下三种数据库：

- **Prometheus** - 使用 Remote Write 协议
- **VictoriaMetrics** - 兼容 Prometheus Remote Write 协议
- **ClickHouse** - 使用原生 TCP 协议

## 配置模式

### 1. 直连模式 (direct)

所有数据直接发送到配置的时序数据库，不经过中心端。

```yaml
sender:
  mode: "direct"
  batch_size: 1000
  flush_interval: 10s
  timeout: 30s
  retry_times: 3
  retry_interval: 5s
```

### 2. 中心端模式 (core)

所有数据发送到中心端，由中心端转发到时序数据库。

```yaml
sender:
  mode: "core"
```

### 3. 混合模式 (hybrid)

数据同时发送到中心端和时序数据库。

```yaml
sender:
  mode: "hybrid"
```

## Prometheus 配置

### 基本配置

```yaml
sender:
  mode: "direct"
  direct:
    prometheus:
      enabled: true
      url: "http://prometheus:9090/api/v1/write"
```

### 带认证的配置

```yaml
sender:
  direct:
    prometheus:
      enabled: true
      url: "http://prometheus:9090/api/v1/write"
      username: "admin"
      password: "secret"
```

### 多租户配置

```yaml
sender:
  direct:
    prometheus:
      enabled: true
      url: "http://prometheus:9090/api/v1/write"
      headers:
        X-Scope-OrgID: "tenant1"
        X-Custom-Header: "custom-value"
```

### Prometheus 服务端配置

在 Prometheus 配置文件中启用 Remote Write Receiver：

```yaml
# prometheus.yml
remote_write:
  - url: http://localhost:9090/api/v1/write
    
# 启用 Remote Write Receiver (Prometheus 2.25+)
# 启动参数：
# --web.enable-remote-write-receiver
```

## VictoriaMetrics 配置

### 基本配置

```yaml
sender:
  mode: "direct"
  direct:
    victoria_metrics:
      enabled: true
      url: "http://victoria:8428/api/v1/write"
```

### 集群模式配置

```yaml
sender:
  direct:
    victoria_metrics:
      enabled: true
      url: "http://vminsert:8480/insert/0/prometheus/api/v1/write"
      headers:
        X-Scope-OrgID: "tenant1"
```

### 带认证的配置

```yaml
sender:
  direct:
    victoria_metrics:
      enabled: true
      url: "http://victoria:8428/api/v1/write"
      username: "admin"
      password: "secret"
```

### VictoriaMetrics 服务端配置

VictoriaMetrics 默认支持 Prometheus Remote Write 协议，无需额外配置。

单机版启动：
```bash
./victoria-metrics-prod -storageDataPath=/var/lib/victoria-metrics
```

集群版 vminsert 启动：
```bash
./vminsert-prod -storageNode=vmstorage:8400
```

## ClickHouse 配置

### 基本配置

```yaml
sender:
  mode: "direct"
  direct:
    clickhouse:
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics?username=default&password=&compress=true"
      table_name: "metrics"
      batch_size: 1000
```

### 完整 DSN 配置

```yaml
sender:
  direct:
    clickhouse:
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics?username=monitor&password=secret&compress=true&dial_timeout=10s&read_timeout=20s"
      table_name: "sentinel_metrics"
      batch_size: 5000
```

### DSN 参数说明

- `username` - 用户名
- `password` - 密码
- `compress` - 是否启用压缩（推荐 true）
- `dial_timeout` - 连接超时
- `read_timeout` - 读取超时
- `write_timeout` - 写入超时
- `debug` - 是否启用调试模式

### ClickHouse 表结构

Sentinel 会自动创建表，默认结构如下：

```sql
CREATE TABLE IF NOT EXISTS metrics (
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
```

### 自定义表结构

如果需要自定义表结构，请在启动 Sentinel 前手动创建表：

```sql
CREATE TABLE custom_metrics (
    timestamp DateTime64(3),
    metric_name String,
    metric_value Float64,
    metric_type String,
    device_id String,
    labels Map(String, String),
    region String,  -- 自定义字段
    date Date DEFAULT toDate(timestamp)
) ENGINE = MergeTree()
PARTITION BY (toYYYYMM(date), region)
ORDER BY (metric_name, device_id, timestamp)
TTL date + INTERVAL 180 DAY
```

然后在配置中指定表名：

```yaml
sender:
  direct:
    clickhouse:
      table_name: "custom_metrics"
```

## 多目标配置

可以同时启用多个目标，数据会并发发送到所有启用的目标：

```yaml
sender:
  mode: "direct"
  batch_size: 1000
  flush_interval: 10s
  timeout: 30s
  
  direct:
    # 同时发送到 Prometheus
    prometheus:
      enabled: true
      url: "http://prometheus:9090/api/v1/write"
    
    # 和 VictoriaMetrics
    victoria_metrics:
      enabled: true
      url: "http://victoria:8428/api/v1/write"
    
    # 和 ClickHouse
    clickhouse:
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics"
      table_name: "metrics"
```

## 性能优化

### 批量大小

调整 `batch_size` 以平衡延迟和吞吐量：

```yaml
sender:
  batch_size: 5000  # 增加批量大小可以提高吞吐量，但会增加延迟
```

### 刷新间隔

调整 `flush_interval` 以控制数据发送频率：

```yaml
sender:
  flush_interval: 5s  # 更频繁的刷新可以降低延迟
```

### ClickHouse 批量大小

ClickHouse 支持单独配置批量大小：

```yaml
sender:
  direct:
    clickhouse:
      batch_size: 10000  # ClickHouse 可以处理更大的批量
```

### 压缩

启用压缩可以减少网络传输：

```yaml
sender:
  direct:
    clickhouse:
      dsn: "tcp://clickhouse:9000/metrics?compress=true"
```

## 监控和调试

### 查看发送统计

Sentinel 会记录发送成功和失败的指标数量：

```bash
# 查看日志
tail -f logs/sentinel.log | grep "Sent to"
```

### 调试模式

启用调试日志以查看详细的发送信息：

```yaml
logging:
  level: debug
```

### ClickHouse 查询验证

验证数据是否成功写入 ClickHouse：

```sql
-- 查看最近的数据
SELECT 
    timestamp,
    metric_name,
    metric_value,
    device_id,
    labels
FROM metrics
ORDER BY timestamp DESC
LIMIT 10;

-- 统计指标数量
SELECT 
    metric_name,
    count() as count
FROM metrics
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY metric_name
ORDER BY count DESC;

-- 按设备统计
SELECT 
    device_id,
    count() as count
FROM metrics
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY device_id
ORDER BY count DESC;
```

### Prometheus 查询验证

使用 PromQL 查询验证数据：

```promql
# 查看所有指标
{__name__=~".+"}

# 查看特定指标
cpu_usage{host="server1"}

# 查看最近的数据
rate(cpu_usage[5m])
```

## 故障排查

### 连接失败

**问题**: 无法连接到数据库

**解决方案**:
1. 检查网络连接
2. 验证 URL/DSN 是否正确
3. 检查防火墙规则
4. 验证认证信息

### 数据未写入

**问题**: 数据发送成功但查询不到

**解决方案**:
1. 检查表名是否正确
2. 验证时间戳是否正确
3. 检查数据库权限
4. 查看数据库日志

### 性能问题

**问题**: 发送速度慢或延迟高

**解决方案**:
1. 增加 `batch_size`
2. 减少 `flush_interval`
3. 启用压缩
4. 检查网络带宽
5. 优化数据库配置

### 认证失败

**问题**: 401 Unauthorized 错误

**解决方案**:
1. 验证用户名和密码
2. 检查 API Token
3. 确认权限配置

## 最佳实践

### 1. 选择合适的模式

- **开发环境**: 使用 `direct` 模式，简化部署
- **生产环境**: 使用 `core` 模式，便于集中管理
- **高可用场景**: 使用 `hybrid` 模式，双重保障

### 2. 合理配置批量大小

- 小批量（100-500）: 低延迟场景
- 中批量（1000-2000）: 平衡场景
- 大批量（5000+）: 高吞吐场景

### 3. 启用压缩

对于网络带宽有限的环境，建议启用压缩：
- Prometheus/VictoriaMetrics: 自动使用 Snappy 压缩
- ClickHouse: 在 DSN 中添加 `compress=true`

### 4. 设置合理的超时

```yaml
sender:
  timeout: 30s        # 发送超时
  retry_times: 3      # 重试次数
  retry_interval: 5s  # 重试间隔
```

### 5. 监控发送状态

定期检查发送统计，及时发现问题：
- 成功率
- 失败率
- 延迟
- 吞吐量

## 示例配置

### 完整的生产环境配置

```yaml
sentinel:
  id: ""
  name: "sentinel-prod-1"
  region: "us-west"

core:
  url: "https://gravital-core.example.com"
  api_token: "${API_TOKEN}"

sender:
  mode: "hybrid"  # 混合模式，双重保障
  batch_size: 2000
  flush_interval: 10s
  timeout: 30s
  retry_times: 3
  retry_interval: 5s
  
  direct:
    victoria_metrics:
      enabled: true
      url: "http://vminsert:8480/insert/0/prometheus/api/v1/write"
      headers:
        X-Scope-OrgID: "production"
    
    clickhouse:
      enabled: true
      dsn: "tcp://clickhouse:9000/metrics?username=monitor&password=${CH_PASSWORD}&compress=true"
      table_name: "sentinel_metrics"
      batch_size: 5000

buffer:
  type: "memory"
  size: 50000
  flush_interval: 10s

logging:
  level: info
  format: json
  output: both
  file_path: "./logs/sentinel.log"
```

## 参考资料

- [Prometheus Remote Write Specification](https://prometheus.io/docs/concepts/remote_write_spec/)
- [VictoriaMetrics Documentation](https://docs.victoriametrics.com/)
- [ClickHouse Documentation](https://clickhouse.com/docs/)

