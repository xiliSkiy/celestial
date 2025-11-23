# 告警引擎 VictoriaMetrics 集成说明

## 概述

本文档说明告警引擎如何从 VictoriaMetrics 时序数据库查询指标数据，以及如何配置和测试该功能。

---

## 1. 架构变更

### 1.1 变更前（旧实现）

```
┌──────────────┐
│ Alert Engine │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  PostgreSQL  │  ← 查询 devices.status 字段
│  (devices)   │
└──────────────┘
```

**问题**：
- ❌ 数据不实时（依赖 DeviceMonitor 每分钟更新）
- ❌ 只支持 `device_status` 指标
- ❌ 与时序数据库数据不一致

### 1.2 变更后（新实现）

```
┌──────────────┐
│ Alert Engine │
└──────┬───────┘
       │
       ├──────────────┐
       │              │
       ▼              ▼ (fallback)
┌──────────────┐  ┌──────────────┐
│VictoriaMetrics│  │  PostgreSQL  │
│  (时序数据)   │  │  (devices)   │
└──────────────┘  └──────────────┘
```

**优势**：
- ✅ 实时数据（Sentinel 每 30 秒上报）
- ✅ 支持所有时序指标
- ✅ 数据一致性
- ✅ 自动回退机制（VM 不可用时使用 PostgreSQL）

---

## 2. 实现细节

### 2.1 VictoriaMetrics 客户端

**文件位置**: `gravital-core/internal/alert/engine/vm_client.go`

#### 核心功能

```go
type VMClient struct {
    baseURL string        // VictoriaMetrics URL
    client  *http.Client  // HTTP 客户端
    logger  *zap.Logger   // 日志
}

// Query 执行 PromQL 查询
func (c *VMClient) Query(promQL string) ([]MetricResult, error)

// QueryRange 执行范围查询
func (c *VMClient) QueryRange(promQL string, start, end time.Time, step time.Duration) ([]MetricResult, error)

// Health 检查健康状态
func (c *VMClient) Health() error
```

#### 查询流程

```
1. 构建查询 URL
   GET /api/v1/query?query={promQL}&time={timestamp}

2. 发送 HTTP 请求
   ├─► 成功 ──► 解析 JSON 响应
   └─► 失败 ──► 返回错误

3. 解析响应
   {
     "status": "success",
     "data": {
       "resultType": "vector",
       "result": [
         {
           "metric": {"device_id": "dev-001", ...},
           "value": [timestamp, "1.0"]
         }
       ]
     }
   }

4. 转换为 MetricResult
   [{Labels: {"device_id": "dev-001"}, Value: 1.0}]
```

### 2.2 告警引擎集成

**文件位置**: `gravital-core/internal/alert/engine/engine.go`

#### 初始化

```go
func NewAlertEngine(db *gorm.DB, logger *zap.Logger, cfg *Config) *AlertEngine {
    // 创建 VM 客户端
    vmClient := NewVMClient(cfg.VMURL, logger)
    
    // 健康检查
    if cfg.VMURL != "" {
        if err := vmClient.Health(); err != nil {
            logger.Warn("VictoriaMetrics health check failed")
        } else {
            logger.Info("VictoriaMetrics connection established")
        }
    }
    
    return &AlertEngine{
        vmClient: vmClient,
        // ...
    }
}
```

#### 查询逻辑（带回退机制）

```go
func (e *AlertEngine) queryMetric(query string) ([]MetricResult, error) {
    // 1. 优先使用 VictoriaMetrics
    if e.vmClient != nil && e.vmClient.baseURL != "" {
        results, err := e.vmClient.Query(query)
        if err != nil {
            // VM 查询失败，回退到数据库
            logger.Warn("Failed to query VM, falling back to database")
            return e.queryMetricFromDB(query)
        }
        
        // VM 返回空结果，尝试数据库（兼容性）
        if len(results) == 0 {
            return e.queryMetricFromDB(query)
        }
        
        return results, nil
    }
    
    // 2. 未配置 VM，使用数据库
    return e.queryMetricFromDB(query)
}

func (e *AlertEngine) queryMetricFromDB(query string) ([]MetricResult, error) {
    // 解析查询条件
    // 从 PostgreSQL devices 表查询
    // 仅支持 device_status 指标
}
```

---

## 3. 配置说明

### 3.1 VictoriaMetrics URL 配置

告警引擎从 `Forwarder.Targets` 配置中自动查找 VictoriaMetrics 端点。

**配置文件**: `gravital-core/config/config.yaml`

```yaml
forwarder:
  targets:
    - type: victoriametrics
      enabled: true
      endpoint: http://localhost:8428  # ← VictoriaMetrics URL
      batch_size: 100
      flush_interval: 10s
```

**启动日志**：

```
INFO  Starting alert engine...
INFO  VictoriaMetrics connection established  url=http://localhost:8428
INFO  Alert engine started
```

### 3.2 未配置 VictoriaMetrics

如果未配置或配置为空，告警引擎会自动使用数据库回退模式：

```
WARN  VictoriaMetrics URL not configured, alert engine will use fallback mode
```

---

## 4. 查询示例

### 4.1 简单查询

**告警规则条件**：
```
device_status != 1
```

**生成的 PromQL**：
```
device_status
```

**VictoriaMetrics 查询**：
```bash
curl "http://localhost:8428/api/v1/query?query=device_status"
```

**响应**：
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "__name__": "device_status",
          "device_id": "dev-001",
          "device_type": "router",
          "plugin": "ping",
          "sentinel_id": "sentinel-xxx",
          "task_id": "task-xxx"
        },
        "value": [1732348800, "1"]
      }
    ]
  }
}
```

### 4.2 带过滤条件的查询

**告警规则**：
```json
{
  "condition": "device_status != 1",
  "filters": {
    "device_type": "router"
  }
}
```

**生成的 PromQL**：
```
device_status{device_type="router"}
```

**VictoriaMetrics 查询**：
```bash
curl "http://localhost:8428/api/v1/query?query=device_status%7Bdevice_type%3D%22router%22%7D"
```

### 4.3 多个过滤条件

**告警规则**：
```json
{
  "condition": "device_status != 1",
  "filters": {
    "device_type": "router",
    "device_id": "dev-001"
  }
}
```

**生成的 PromQL**：
```
device_status{device_type="router",device_id="dev-001"}
```

---

## 5. 回退机制

### 5.1 触发回退的场景

1. **VictoriaMetrics 不可用**
   ```
   WARN  Failed to query VictoriaMetrics, falling back to database
         query=device_status error="connection refused"
   ```

2. **VictoriaMetrics 返回空结果**
   ```
   DEBUG VictoriaMetrics returned no results, trying database fallback
         query=device_status
   ```

3. **未配置 VictoriaMetrics**
   ```
   DEBUG VictoriaMetrics not configured, using database fallback
         query=device_status
   ```

### 5.2 回退限制

数据库回退模式**仅支持** `device_status` 指标：

```go
if metricName == "device_status" {
    // 从 PostgreSQL devices 表查询
    // 转换 status 字段：online -> 1.0, 其他 -> 0.0
} else {
    return nil, fmt.Errorf("unsupported metric for database fallback: %s", metricName)
}
```

对于其他指标（如 `cpu_usage`、`memory_usage`），**必须**使用 VictoriaMetrics。

---

## 6. 测试验证

### 6.1 使用测试脚本

我们提供了一个完整的测试脚本：

```bash
cd /Users/liangxin/Downloads/code/celestial
./test_vm_alert.sh
```

**测试内容**：
1. ✅ VictoriaMetrics 健康检查
2. ✅ 查询时序数据库中的 `device_status` 指标
3. ✅ 查询特定设备的状态
4. ✅ 检查告警规则配置
5. ✅ 检查告警事件生成
6. ✅ 检查告警聚合

### 6.2 手动测试步骤

#### 步骤 1：检查 VictoriaMetrics

```bash
# 健康检查
curl http://localhost:8428/health

# 查询所有 device_status 指标
curl "http://localhost:8428/api/v1/query?query=device_status"

# 查询特定设备
curl "http://localhost:8428/api/v1/query?query=device_status{device_id=\"dev-001\"}"
```

#### 步骤 2：创建告警规则

```bash
# 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 创建设备离线告警规则
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rule_name": "设备离线告警",
    "description": "监控设备在线状态",
    "severity": "critical",
    "condition": "device_status != 1",
    "duration": 300,
    "enabled": true
  }'
```

#### 步骤 3：等待告警评估

告警引擎每 30 秒评估一次规则，等待 30-60 秒后检查告警事件。

#### 步骤 4：查看告警事件

```bash
# 查看告警事件列表
curl -X GET "http://localhost:8080/api/v1/alert-events?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq

# 查看告警聚合
curl -X GET "http://localhost:8080/api/v1/alert-aggregations" \
  -H "Authorization: Bearer $TOKEN" | jq
```

#### 步骤 5：查看日志

```bash
# 查看 Gravital Core 日志
docker-compose logs -f gravital-core | grep -i alert

# 关键日志示例：
# INFO  Starting alert engine...
# INFO  VictoriaMetrics connection established  url=http://victoriametrics:8428
# DEBUG Querying VictoriaMetrics  query=device_status
# DEBUG VictoriaMetrics query result  query=device_status result_count=5
# INFO  Alert triggered  rule=设备离线告警 device_id=dev-001
```

---

## 7. 故障排查

### 7.1 VictoriaMetrics 连接失败

**症状**：
```
WARN  VictoriaMetrics health check failed  url=http://localhost:8428 error="connection refused"
```

**解决方案**：
1. 检查 VictoriaMetrics 是否运行：
   ```bash
   docker-compose ps victoriametrics
   ```

2. 检查网络连接：
   ```bash
   curl http://localhost:8428/health
   ```

3. 检查配置文件中的 URL 是否正确

### 7.2 查询返回空结果

**症状**：
```
DEBUG VictoriaMetrics returned no results, trying database fallback  query=device_status
```

**原因**：
- Sentinel 未采集数据
- Sentinel 未上报到 VictoriaMetrics
- 查询条件不匹配

**解决方案**：
1. 检查 Sentinel 是否运行：
   ```bash
   docker-compose ps sentinel
   ```

2. 检查 Sentinel 日志：
   ```bash
   docker-compose logs -f sentinel | grep -i forward
   ```

3. 直接查询 VictoriaMetrics：
   ```bash
   curl "http://localhost:8428/api/v1/query?query=device_status"
   ```

### 7.3 告警未触发

**症状**：
- 规则已创建且启用
- 时序数据存在
- 但没有告警事件生成

**排查步骤**：

1. **检查告警规则条件**：
   ```sql
   SELECT id, rule_name, condition, enabled FROM alert_rules;
   ```

2. **检查时序数据的值**：
   ```bash
   curl "http://localhost:8428/api/v1/query?query=device_status"
   ```

3. **检查告警引擎日志**：
   ```bash
   docker-compose logs gravital-core | grep "Evaluating alert rules"
   ```

4. **手动测试条件**：
   - 如果规则是 `device_status != 1`
   - 时序数据显示 `value = 1`（在线）
   - 则条件 `1 != 1` 为 `false`，不会触发告警 ✅

5. **常见错误**：
   - ❌ 规则条件写反了（如：想监控离线，但写成了 `!= 1`，而设备实际在线）
   - ❌ 过滤条件不匹配（如：`device_type="router"`，但实际是 `switch`）

### 7.4 告警立即自动解决

**症状**：
```
INFO  Alert triggered  rule=设备离线告警 device_id=dev-001
INFO  Alert resolved   rule=设备离线告警 device_id=dev-001
```

**原因**：
- 告警引擎在下一次评估时发现条件不再满足
- 可能是数据波动或条件配置错误

**解决方案**：
1. 增加 `duration`（持续时间），避免瞬时波动触发告警
2. 检查条件是否正确
3. 查看时序数据是否稳定

---

## 8. 性能优化

### 8.1 查询优化

**当前实现**：
- 每次评估都查询 VictoriaMetrics
- 每个规则独立查询

**优化建议**：

1. **查询缓存**（未实现）：
   ```go
   // 缓存查询结果 10 秒
   type QueryCache struct {
       cache map[string]CacheEntry
       mu    sync.RWMutex
   }
   ```

2. **批量查询**（未实现）：
   ```go
   // 合并相同指标的查询
   queries := []string{
       "device_status{device_type=\"router\"}",
       "device_status{device_type=\"switch\"}",
   }
   // 合并为：device_status{device_type=~"router|switch"}
   ```

### 8.2 并发控制

**当前实现**：
- 所有规则并发评估
- 无并发限制

**优化建议**：

```go
// 使用 worker pool 限制并发数
semaphore := make(chan struct{}, 10) // 最多 10 个并发
for _, rule := range rules {
    semaphore <- struct{}{}
    go func(r *model.AlertRule) {
        defer func() { <-semaphore }()
        e.evaluateRule(r)
    }(rule)
}
```

### 8.3 超时控制

**当前实现**：
- HTTP 客户端超时：10 秒

**优化建议**：
- 根据实际情况调整超时时间
- 添加重试机制

---

## 9. 未来增强

### 9.1 支持更多 PromQL 功能

**当前限制**：
- 仅支持简单的 `metric_name{label="value"}` 查询
- 不支持聚合函数（如 `avg()`, `sum()`）
- 不支持运算符（如 `rate()`, `increase()`）

**规划**：
```
# 支持聚合
avg(device_status) < 0.8

# 支持速率计算
rate(network_bytes_sent[5m]) > 1000000

# 支持复杂表达式
(cpu_usage > 80) and (memory_usage > 90)
```

### 9.2 支持范围查询

**当前实现**：
- 仅支持即时查询（`/api/v1/query`）

**规划**：
- 支持范围查询（`/api/v1/query_range`）
- 用于判断"持续 N 分钟"的条件

### 9.3 查询结果可视化

**规划**：
- 在告警事件详情中显示指标趋势图
- 帮助用户理解告警触发原因

---

## 10. 总结

### 10.1 实现的功能

✅ VictoriaMetrics 客户端
✅ PromQL 查询支持
✅ 自动回退机制
✅ 健康检查
✅ 完整的日志记录
✅ 测试脚本

### 10.2 优势

- **实时性**：直接查询时序数据库，数据实时准确
- **扩展性**：支持所有时序指标，不仅限于 `device_status`
- **可靠性**：自动回退到数据库，保证告警功能可用
- **兼容性**：完全兼容现有告警规则

### 10.3 注意事项

⚠️ 确保 VictoriaMetrics 正常运行
⚠️ 确保 Sentinel 正常采集并上报数据
⚠️ 数据库回退模式仅支持 `device_status` 指标
⚠️ 告警引擎每 30 秒评估一次，不是实时的

---

**文档版本**: v1.0
**最后更新**: 2025-11-23
**相关文档**:
- [告警模块详细设计](../../docs/13-告警模块详细设计.md)
- [告警数据源不一致问题](./ALERT_DATA_SOURCE_ISSUE.md)

