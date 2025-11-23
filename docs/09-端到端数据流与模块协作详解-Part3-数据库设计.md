# 端到端数据流与模块协作详解 - Part 3B: 数据库设计与时序库使用

> 本文档详细说明数据库表结构设计、数据流转和时序数据库的使用方式。

## 5. 数据库设计与数据流

### 5.1 PostgreSQL 表结构设计

#### 5.1.1 设备表 (devices)

```sql
CREATE TABLE devices (
    id SERIAL PRIMARY KEY,
    device_id VARCHAR(64) UNIQUE NOT NULL,      -- 业务唯一标识
    name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL,           -- server, switch, router, etc.
    group_id INT REFERENCES device_groups(id),  -- 设备分组
    sentinel_id VARCHAR(128),                   -- 负责采集的 Sentinel
    connection_config JSONB NOT NULL,           -- 连接配置（SSH、SNMP等）
    labels JSONB DEFAULT '{}',                  -- 自定义标签
    status VARCHAR(20) DEFAULT 'unknown',       -- online, offline, unknown
    last_seen TIMESTAMP,                        -- 最后上报时间
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_device_id (device_id),
    INDEX idx_device_type (device_type),
    INDEX idx_sentinel_id (sentinel_id),
    INDEX idx_status (status)
);

-- 示例数据
INSERT INTO devices VALUES (
    1,
    'dev-25422c94',
    '生产服务器-01',
    'server',
    NULL,
    'sentinel-beijing-1-xxx',
    '{"protocol":"ssh","host":"192.168.1.100","port":22,"username":"admin"}',
    '{"env":"production","region":"beijing"}',
    'online',
    '2025-11-20 10:00:00',
    '2025-11-20 09:00:00',
    '2025-11-20 10:00:00'
);
```

**设计要点**：

1. **device_id vs id**：
   - `id`: 数据库自增主键，内部使用
   - `device_id`: 业务唯一标识，对外暴露，便于跨系统引用

2. **connection_config 使用 JSONB**：
   - 不同设备类型的连接方式不同（SSH、SNMP、HTTP）
   - JSONB 支持灵活的 schema
   - PostgreSQL 的 JSONB 支持索引和查询

3. **labels 使用 JSONB**：
   - 支持自定义标签（如环境、地域、业务线）
   - 可以用于查询过滤：`WHERE labels @> '{"env":"production"}'`

#### 5.1.2 采集任务表 (collection_tasks)

```sql
CREATE TABLE collection_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(64) UNIQUE NOT NULL,
    device_id VARCHAR(64) NOT NULL REFERENCES devices(device_id),
    sentinel_id VARCHAR(128) NOT NULL,          -- 执行采集的 Sentinel
    plugin_name VARCHAR(100) NOT NULL,          -- ping, snmp, ssh, etc.
    config JSONB DEFAULT '{}',                  -- 插件配置参数
    interval_seconds INT NOT NULL DEFAULT 60,   -- 采集间隔（秒）
    timeout_seconds INT NOT NULL DEFAULT 30,    -- 超时时间（秒）
    enabled BOOLEAN DEFAULT true,
    priority INT DEFAULT 5,                     -- 优先级 1-10
    retry_count INT DEFAULT 3,
    last_executed_at TIMESTAMP,                 -- 最后执行时间
    next_execution_at TIMESTAMP,                -- 下次执行时间
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_task_id (task_id),
    INDEX idx_device_id (device_id),
    INDEX idx_sentinel_id (sentinel_id),
    INDEX idx_enabled (enabled),
    INDEX idx_next_execution (next_execution_at)
);

-- 示例数据
INSERT INTO collection_tasks VALUES (
    1,
    'task-177bfd59',
    'dev-25422c94',
    'sentinel-beijing-1-xxx',
    'ping',
    '{"count":4,"timeout":"5s"}',
    60,
    30,
    true,
    5,
    3,
    '2025-11-20 10:00:00',
    '2025-11-20 10:01:00',
    '2025-11-20 09:00:00',
    '2025-11-20 09:00:00'
);
```

**设计要点**：

1. **next_execution_at 字段**：
   - 用于查询"即将执行的任务"
   - 支持手动触发（设置为当前时间）
   - 便于任务执行历史追踪

2. **config 使用 JSONB**：
   - 不同插件需要不同的配置参数
   - 灵活性高，易于扩展

3. **索引策略**：
   - `idx_sentinel_id + idx_enabled`: 采集端拉取任务时使用
   - `idx_next_execution`: 中心端查询待执行任务

#### 5.1.3 采集端表 (sentinels)

```sql
CREATE TABLE sentinels (
    id SERIAL PRIMARY KEY,
    sentinel_id VARCHAR(128) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    ip_address VARCHAR(45),
    mac_address VARCHAR(17),
    region VARCHAR(100),
    version VARCHAR(50),
    api_token VARCHAR(255) NOT NULL,            -- 认证 Token
    status VARCHAR(20) DEFAULT 'offline',       -- online, offline
    last_heartbeat_at TIMESTAMP,
    registered_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_sentinel_id (sentinel_id),
    INDEX idx_status (status),
    INDEX idx_last_heartbeat (last_heartbeat_at)
);
```

**设计要点**：

1. **sentinel_id 生成规则**：
   ```
   sentinel-{hostname}-{mac_hash}-{timestamp}
   示例：sentinel-beijing-server-c0893d0d-1763133437
   ```

2. **api_token**：
   - 用于采集端认证
   - 注册时由中心端生成
   - 采集端持久化到本地文件

3. **状态判断**：
   - 通过 `last_heartbeat_at` 判断在线状态
   - 超过 3 分钟无心跳视为离线

#### 5.1.4 告警规则表 (alert_rules)

```sql
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL,
    device_id VARCHAR(64),                      -- NULL 表示全局规则
    metric_name VARCHAR(255) NOT NULL,          -- ping_rtt_avg_ms
    condition VARCHAR(10) NOT NULL,             -- >, <, >=, <=, ==, !=
    threshold FLOAT NOT NULL,
    duration_seconds INT DEFAULT 60,            -- 持续时间
    severity VARCHAR(20) NOT NULL,              -- critical, warning, info
    enabled BOOLEAN DEFAULT true,
    notify_channels JSONB DEFAULT '[]',         -- ["email", "dingtalk"]
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_enabled (enabled),
    INDEX idx_device_id (device_id),
    INDEX idx_metric_name (metric_name)
);

-- 示例数据
INSERT INTO alert_rules VALUES (
    1,
    'CPU 使用率过高',
    'dev-25422c94',
    'cpu_usage_percent',
    '>',
    80.0,
    300,  -- 持续 5 分钟
    'warning',
    true,
    '["email", "dingtalk"]',
    'CPU 使用率超过 80% 持续 5 分钟',
    NOW(),
    NOW()
);
```

#### 5.1.5 告警事件表 (alert_events)

```sql
CREATE TABLE alert_events (
    id SERIAL PRIMARY KEY,
    rule_id INT REFERENCES alert_rules(id),
    rule_name VARCHAR(255) NOT NULL,
    device_id VARCHAR(64),
    metric_name VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    current_value FLOAT NOT NULL,
    threshold FLOAT NOT NULL,
    status VARCHAR(20) NOT NULL,                -- firing, resolved
    message TEXT,
    triggered_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    acknowledged_at TIMESTAMP,
    acknowledged_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_rule_id (rule_id),
    INDEX idx_device_id (device_id),
    INDEX idx_status (status),
    INDEX idx_triggered_at (triggered_at)
);
```

**设计要点**：

1. **告警生命周期**：
   ```
   firing (触发) → acknowledged (确认) → resolved (解决)
   ```

2. **告警去重**：
   - 同一规则的告警，如果未解决，不会重复创建
   - 通过 `rule_id + status='firing'` 查询活跃告警

#### 5.1.6 转发器配置表 (forwarder_configs)

```sql
CREATE TABLE forwarder_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,                  -- prometheus, victoria-metrics, clickhouse
    enabled BOOLEAN DEFAULT true,
    endpoint VARCHAR(500),                      -- HTTP endpoint
    dsn VARCHAR(500),                           -- Database DSN (for ClickHouse)
    auth_config JSONB DEFAULT '{}',             -- 认证配置
    tls_config JSONB DEFAULT '{}',              -- TLS 配置
    batch_size INT DEFAULT 1000,
    flush_interval INT DEFAULT 10,              -- 秒
    retry_times INT DEFAULT 3,
    timeout_seconds INT DEFAULT 30,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    INDEX idx_name (name),
    INDEX idx_enabled (enabled)
);

-- 示例数据
INSERT INTO forwarder_configs VALUES (
    1,
    'vm-prod',
    'victoria-metrics',
    true,
    'http://localhost:8428/api/v1/write',
    NULL,
    '{}',
    '{}',
    1000,
    10,
    3,
    30,
    NOW(),
    NOW()
);
```

### 5.2 数据流转路径

#### 5.2.1 设备数据的完整生命周期

```
1. 设备录入
   ├─> devices 表 (PostgreSQL)
   └─> 状态: unknown

2. 创建采集任务
   ├─> collection_tasks 表 (PostgreSQL)
   └─> 关联: device_id

3. 采集端拉取任务
   ├─> 查询: SELECT * FROM collection_tasks WHERE sentinel_id=? AND enabled=true
   └─> 合并: device.connection_config + task.config

4. 执行采集
   ├─> 插件采集数据
   └─> 生成指标: [{name, value, labels, timestamp}, ...]

5. 数据上报
   ├─> 采集端 Buffer (内存)
   ├─> 批量发送: POST /api/v1/data/ingest (gzip 压缩)
   └─> 中心端接收

6. 数据转发
   ├─> Forwarder Manager Buffer (内存)
   ├─> 批量转发: 并发发送到多个转发器
   └─> 时序库存储: VictoriaMetrics / Prometheus / ClickHouse

7. 数据查询
   ├─> 前端请求: POST /api/v1/metrics/query
   ├─> 中心端代理: 查询 VictoriaMetrics
   └─> 返回结果: {timestamps, values}

8. 告警检测
   ├─> 告警引擎: 定时查询时序库
   ├─> 评估规则: 判断是否满足条件
   ├─> 创建事件: INSERT INTO alert_events
   └─> 发送通知: Email / 钉钉 / 企业微信

9. 前端展示
   ├─> 设备列表: 查询 devices 表
   ├─> 实时指标: 查询时序库
   ├─> 告警历史: 查询 alert_events 表
   └─> 任务状态: 查询 collection_tasks 表
```

#### 5.2.2 数据查询的优化策略

**场景 1：设备列表页面**

```sql
-- ❌ 低效：N+1 查询
SELECT * FROM devices LIMIT 20;
-- 然后对每个设备查询任务数量
SELECT COUNT(*) FROM collection_tasks WHERE device_id = ?;

-- ✅ 高效：使用 JOIN
SELECT 
    d.*,
    COUNT(t.id) as task_count,
    MAX(t.last_executed_at) as last_collection_time
FROM devices d
LEFT JOIN collection_tasks t ON d.device_id = t.device_id
GROUP BY d.id
LIMIT 20;
```

**场景 2：设备详情页面**

```go
// 并发查询多个数据源
var (
    device  *model.Device
    tasks   []*model.Task
    alerts  []*model.Alert
    metrics map[string]float64
)

// 并发执行
var wg sync.WaitGroup
wg.Add(4)

go func() {
    defer wg.Done()
    device, _ = deviceRepo.GetByID(ctx, deviceID)
}()

go func() {
    defer wg.Done()
    tasks, _ = taskRepo.GetByDeviceID(ctx, deviceID)
}()

go func() {
    defer wg.Done()
    alerts, _ = alertRepo.GetByDeviceID(ctx, deviceID)
}()

go func() {
    defer wg.Done()
    metrics = queryLatestMetrics(deviceID)
}()

wg.Wait()
```

---

## 6. 时序数据库使用详解

### 6.1 VictoriaMetrics 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    VictoriaMetrics                           │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Write Path (写入路径)                              │    │
│  │                                                      │    │
│  │  HTTP API ──> Parser ──> Deduplication ──> Storage │    │
│  │  (Remote Write)  (解析)    (去重)        (存储)    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Storage Layer (存储层)                             │    │
│  │                                                      │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐         │    │
│  │  │  Part 1  │  │  Part 2  │  │  Part 3  │  ...    │    │
│  │  │(1个月数据)│  │(1个月数据)│  │(1个月数据)│         │    │
│  │  └──────────┘  └──────────┘  └──────────┘         │    │
│  │                                                      │    │
│  │  - 自动压缩（10:1 压缩率）                          │    │
│  │  - 自动合并（小文件合并为大文件）                   │    │
│  │  - 自动降采样（可选）                               │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Query Path (查询路径)                              │    │
│  │                                                      │    │
│  │  HTTP API ──> PromQL Parser ──> Query Engine       │    │
│  │              (解析查询)        (执行查询)           │    │
│  │                                  │                   │    │
│  │                                  ▼                   │    │
│  │                            Index Lookup              │    │
│  │                            (索引查找)                │    │
│  │                                  │                   │    │
│  │                                  ▼                   │    │
│  │                            Data Scan                 │    │
│  │                            (数据扫描)                │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 数据存储格式

#### 6.2.1 指标命名规范

```
<metric_name>{<label1>="<value1>", <label2>="<value2>", ...}

示例：
ping_rtt_avg_ms{
    device_id="dev-25422c94",
    host="192.168.1.100",
    task_id="task-177bfd59",
    sentinel_id="sentinel-beijing-1-xxx"
} 12.5 1700500000000
```

#### 6.2.2 标签设计原则

**好的标签设计**：

```
✅ 基数可控的标签（Cardinality < 10000）
- device_id        (设备数量：几百到几千)
- device_type      (类型：server, switch, router 等，< 10)
- env              (环境：prod, test, dev，< 5)
- region           (地域：beijing, shanghai 等，< 20)
- sentinel_id      (采集端数量：几十到几百)

✅ 有查询价值的标签
- 经常用于过滤：device_id, env, region
- 用于聚合：device_type, region
```

**不好的标签设计**：

```
❌ 高基数标签（会导致性能问题）
- timestamp        (每个数据点都不同)
- request_id       (每个请求都不同)
- user_id          (用户数量可能很大)
- ip_address       (IP 数量很大)

❌ 无查询价值的标签
- description      (描述性文本)
- full_config      (完整配置)
```

**基数计算**：

```
总时间序列数 = metric_name 数量 × 所有标签值的笛卡尔积

示例：
- 10 个指标
- device_id: 1000 个设备
- env: 3 个环境
- region: 5 个地域

总时间序列 = 10 × 1000 × 3 × 5 = 150,000

VictoriaMetrics 可以轻松处理百万级时间序列。
```

### 6.3 PromQL 查询示例

#### 6.3.1 基础查询

```promql
# 1. 查询最新值
ping_rtt_avg_ms{device_id="dev-25422c94"}

# 2. 查询时间范围
ping_rtt_avg_ms{device_id="dev-25422c94"}[5m]

# 3. 多条件过滤
cpu_usage_percent{device_type="server", env="production"}

# 4. 正则匹配
memory_used_bytes{device_id=~"dev-.*"}

# 5. 排除条件
network_rx_bytes{device_type!="switch"}
```

#### 6.3.2 聚合查询

```promql
# 1. 平均值
avg(cpu_usage_percent{env="production"})

# 2. 求和
sum(network_rx_bytes) by (device_type)

# 3. 最大值
max(memory_used_percent) by (region)

# 4. 分位数
quantile(0.95, http_request_duration_seconds)

# 5. 计数
count(up{job="sentinel"} == 1)
```

#### 6.3.3 速率计算

```promql
# 1. 每秒速率（适用于 counter 类型）
rate(network_rx_bytes[5m])

# 2. 每秒增长（适用于 counter 类型）
irate(http_requests_total[5m])

# 3. 增量
increase(http_requests_total[1h])

# 4. 变化率
delta(temperature_celsius[10m])
```

#### 6.3.4 复杂查询

```promql
# 1. CPU 使用率 Top 10
topk(10, cpu_usage_percent)

# 2. 内存使用率超过 80% 的设备数量
count(memory_used_percent > 80)

# 3. 网络流量趋势（5 分钟平均）
avg_over_time(rate(network_rx_bytes[1m])[5m:])

# 4. 告警条件：CPU 持续 5 分钟超过 80%
cpu_usage_percent > 80 and 
avg_over_time(cpu_usage_percent[5m]) > 80

# 5. 设备可用性（最近 1 小时）
avg_over_time(up[1h]) * 100
```

### 6.4 数据保留策略

#### 6.4.1 VictoriaMetrics 配置

```yaml
# victoria-metrics 启动参数
./victoria-metrics-prod \
  -storageDataPath=/var/lib/victoria-metrics \
  -retentionPeriod=90d \              # 保留 90 天
  -dedup.minScrapeInterval=60s \      # 去重间隔
  -search.maxQueryDuration=60s \      # 最大查询时长
  -memory.allowedPercent=80            # 允许使用 80% 内存
```

#### 6.4.2 降采样策略（可选）

```yaml
# 使用 vmagent 进行降采样
# 原始数据：60s 间隔，保留 7 天
# 降采样数据：5m 间隔，保留 90 天

vmagent:
  remoteWrite:
    - url: http://vm-raw:8428/api/v1/write
      # 原始数据
      
    - url: http://vm-downsampled:8428/api/v1/write
      # 降采样数据
      relabelConfigs:
        - sourceLabels: [__name__]
          regex: '.*'
          action: aggregate
          params:
            interval: 5m
            func: avg
```

### 6.5 查询性能优化

#### 6.5.1 索引优化

```promql
# ✅ 高效：使用精确标签过滤
cpu_usage_percent{device_id="dev-25422c94"}

# ❌ 低效：不使用标签过滤（全表扫描）
cpu_usage_percent

# ✅ 高效：使用多个标签缩小范围
cpu_usage_percent{device_type="server", env="production"}

# ❌ 低效：使用正则匹配（索引效率低）
cpu_usage_percent{device_id=~".*"}
```

#### 6.5.2 查询范围优化

```promql
# ✅ 高效：查询较短时间范围
rate(network_rx_bytes[5m])

# ❌ 低效：查询过长时间范围
rate(network_rx_bytes[24h])

# ✅ 高效：使用降采样数据查询历史
avg_over_time(cpu_usage_percent[5m])[30d:5m]

# ❌ 低效：直接查询 30 天原始数据
cpu_usage_percent[30d]
```

#### 6.5.3 聚合优化

```promql
# ✅ 高效：先聚合再计算
sum(rate(network_rx_bytes[5m])) by (device_type)

# ❌ 低效：先计算再聚合
sum(network_rx_bytes) by (device_type) / 300

# ✅ 高效：使用 recording rules 预计算
# 创建 recording rule
- record: device:network_rx_rate:5m
  expr: rate(network_rx_bytes[5m])

# 查询时使用预计算结果
sum(device:network_rx_rate:5m) by (device_type)
```

---

## 6.6 前端如何使用时序数据

### 6.6.1 实时监控大盘

```typescript
// 查询最近 1 小时的 CPU 使用率
const fetchCPUData = async (deviceId: string) => {
  const now = Math.floor(Date.now() / 1000)
  const start = now - 3600  // 1 小时前
  
  const res = await metricsApi.queryRange({
    query: `cpu_usage_percent{device_id="${deviceId}"}`,
    start,
    end: now,
    step: '60s'  // 1 分钟间隔
  })
  
  // 解析响应
  const result = res.data.result[0]
  const dataPoints = result.values.map(v => ({
    time: new Date(v[0] * 1000),
    value: parseFloat(v[1])
  }))
  
  // 更新图表
  updateChart(dataPoints)
}

// 每分钟刷新一次
setInterval(() => fetchCPUData(deviceId), 60000)
```

### 6.6.2 历史趋势分析

```typescript
// 查询最近 30 天的平均值（使用降采样）
const fetchHistoricalTrend = async (deviceId: string) => {
  const now = Math.floor(Date.now() / 1000)
  const start = now - 30 * 24 * 3600  // 30 天前
  
  const res = await metricsApi.queryRange({
    query: `avg_over_time(cpu_usage_percent{device_id="${deviceId}"}[1h])`,
    start,
    end: now,
    step: '1h'  // 1 小时间隔
  })
  
  // 渲染趋势图
  renderTrendChart(res.data.result[0].values)
}
```

### 6.6.3 设备对比

```typescript
// 对比多个设备的性能
const compareDevices = async (deviceIds: string[]) => {
  const queries = deviceIds.map(id => 
    `avg_over_time(cpu_usage_percent{device_id="${id}"}[5m])`
  )
  
  const results = await Promise.all(
    queries.map(query => metricsApi.queryCurrent(query))
  )
  
  // 渲染对比图表
  renderComparisonChart(results)
}
```

---

## 总结

### 数据流转关键点

1. **元数据（PostgreSQL）**：
   - 设备信息、任务配置、告警规则
   - 支持复杂查询和事务
   - 数据量小，查询快

2. **时序数据（VictoriaMetrics）**：
   - 采集指标、性能数据
   - 高压缩率、高查询性能
   - 数据量大，按时间查询

3. **缓存数据（Redis）**：
   - 会话、在线状态、分布式锁
   - 高频访问，可丢失
   - TTL 自动过期

### 设计原则

1. **数据分类存储**：根据数据特性选择合适的存储
2. **索引优化**：为常用查询条件建立索引
3. **标签设计**：控制基数，提高查询效率
4. **降采样**：历史数据降采样，节省存储空间
5. **并发查询**：多数据源并发查询，提高响应速度


