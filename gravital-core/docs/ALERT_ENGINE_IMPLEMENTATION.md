# 告警引擎实现说明

## 问题描述

创建了告警规则，并且也有数据在上报，但是没有告警事件生成。

**根本原因**：告警引擎还没有实现，虽然有目录结构 `internal/alert/engine/`，但目录是空的，主程序也没有启动告警引擎。

## 解决方案

实现了完整的告警引擎，包括：

### 1. 告警引擎实现

**文件**：`gravital-core/internal/alert/engine/engine.go`

**核心功能**：

1. **规则评估循环**
   - 定期（默认 30 秒）评估所有启用的告警规则
   - 并发评估多个规则，提高性能

2. **条件解析**
   - 支持格式：`metric_name operator threshold`
   - 示例：`device_status != 0`
   - 支持的运算符：`>`, `>=`, `<`, `<=`, `==`, `!=`

3. **指标查询**
   - 当前实现：从 PostgreSQL 数据库查询设备状态
   - 支持过滤条件（device_id, device_type 等）
   - 未来可扩展：调用 VictoriaMetrics API

4. **告警触发**
   - 检测到满足告警条件时创建告警事件
   - 记录活跃告警，避免重复触发
   - 生成详细的告警消息

5. **告警解决**
   - 条件不再满足时自动解决告警
   - 更新事件状态为 `resolved`
   - 记录解决时间

### 2. 主程序集成

**文件**：`gravital-core/cmd/server/main.go`

**修改内容**：

```go
// 启动告警引擎
logger.Info("Starting alert engine...")

// 从转发器配置中查找 VictoriaMetrics 端点
vmURL := ""
for _, target := range cfg.Forwarder.Targets {
    if target.Type == "victoriametrics" && target.Enabled {
        vmURL = target.Endpoint
        break
    }
}

alertEngine := engine.NewAlertEngine(db, logger.Get(), &engine.Config{
    VMURL:         vmURL,
    CheckInterval: 30 * time.Second, // 每 30 秒检查一次
})
alertEngine.Start()
logger.Info("Alert engine started")
```

**优雅关闭**：

```go
// 停止告警引擎
logger.Info("Stopping alert engine...")
alertEngine.Stop()
```

## 工作流程

```
┌─────────────────────────────────────────────────────────────┐
│ 1. 告警引擎启动                                               │
│    - 初始化活跃告警映射                                       │
│    - 启动评估循环（每 30 秒）                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. 获取所有启用的告警规则                                     │
│    SELECT * FROM alert_rules WHERE enabled = true           │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. 并发评估每个规则                                           │
│    - 解析条件：device_status != 0                            │
│    - 提取：metric_name, operator, threshold                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. 查询指标数据                                               │
│    - 从数据库查询设备状态                                     │
│    - 应用过滤条件（device_id, device_type）                  │
│    - 返回：[{device_id, value}, ...]                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. 评估每个设备                                               │
│    - 检查条件：value != 0                                    │
│    - 满足条件 → 触发告警                                      │
│    - 不满足条件 → 解决告警                                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. 触发告警                                                   │
│    - 检查是否已有活跃告警（避免重复）                         │
│    - 创建告警事件：INSERT INTO alert_events                  │
│    - 记录活跃告警映射                                         │
│    - 记录日志                                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. 解决告警                                                   │
│    - 检查是否有活跃告警                                       │
│    - 更新事件状态：UPDATE alert_events SET status='resolved' │
│    - 从活跃告警映射中移除                                     │
│    - 记录日志                                                 │
└─────────────────────────────────────────────────────────────┘
```

## 数据结构

### AlertEngine

```go
type AlertEngine struct {
    db             *gorm.DB
    logger         *zap.Logger
    alertRepo      repository.AlertRepository
    vmURL          string
    checkInterval  time.Duration
    ctx            context.Context
    cancel         context.CancelFunc
    wg             sync.WaitGroup
    activeAlerts   map[uint]map[string]*ActiveAlert // rule_id -> device_id -> alert
    activeAlertsMu sync.RWMutex
}
```

### ActiveAlert

```go
type ActiveAlert struct {
    RuleID       uint
    DeviceID     string
    EventID      uint
    FirstFiredAt time.Time
    LastFiredAt  time.Time
}
```

## 配置参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `CheckInterval` | 30s | 规则评估间隔 |
| `VMURL` | 从配置读取 | VictoriaMetrics 端点（可选）|

## 支持的告警规则格式

### 条件格式

```
metric_name operator threshold
```

### 示例

```json
{
  "rule_name": "设备离线告警",
  "condition": "device_status == 0",
  "duration": 300,
  "severity": "critical",
  "enabled": true
}
```

```json
{
  "rule_name": "CPU 使用率过高",
  "condition": "cpu_usage > 80",
  "duration": 300,
  "severity": "warning",
  "enabled": true
}
```

### 支持的运算符

- `>` - 大于
- `>=` - 大于等于
- `<` - 小于
- `<=` - 小于等于
- `==` - 等于
- `!=` - 不等于

## 当前限制

### 1. 指标查询

**当前实现**：只支持从数据库查询 `device_status` 指标

```go
// 对于 device_status 指标，从数据库查询
if metricName == "device_status" {
    var devices []model.Device
    query := e.db.Model(&model.Device{})
    // ... 查询设备状态
}
```

**未来扩展**：调用 VictoriaMetrics API 查询任意指标

```go
// TODO: 调用 VictoriaMetrics API
resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", e.vmURL, query))
```

### 2. 持续时间（Duration）

**当前实现**：暂未实现持续时间检查，只要当前值满足条件就触发告警

**未来实现**：需要在持续时间内所有数据点都满足条件才触发

```go
// 需要查询时间范围内的数据
start := now.Add(-time.Duration(rule.Duration) * time.Second)
result := queryRange(metricName, start, now)

// 检查所有数据点
for _, point := range result {
    if !checkCondition(point.Value, operator, threshold) {
        return false  // 有一个不满足就不告警
    }
}
```

### 3. 通知功能

**当前实现**：只创建告警事件，不发送通知

**未来实现**：集成通知服务（邮件、Webhook、钉钉等）

## 验证方法

### 1. 启动服务

```bash
cd gravital-core
go run cmd/server/main.go -c config/config.yaml
```

**日志输出**：

```
2025-11-23T10:00:00.000+0800    INFO    Starting alert engine...
2025-11-23T10:00:00.001+0800    INFO    Alert engine started
```

### 2. 创建告警规则

```bash
curl -X POST http://localhost:8080/api/v1/alert-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "rule_name": "设备离线告警",
    "description": "检测设备离线",
    "severity": "critical",
    "condition": "device_status == 0",
    "duration": 300,
    "enabled": true,
    "filters": {},
    "notification_config": {}
  }'
```

### 3. 模拟设备离线

```sql
-- 将设备状态设置为 offline
UPDATE devices SET status = 'offline' WHERE device_id = 'dev-25422c94';
```

### 4. 等待评估（最多 30 秒）

**日志输出**：

```
2025-11-23T10:00:30.000+0800    INFO    Alert triggered
    rule: 设备离线告警
    device_id: dev-25422c94
    value: 0.000000
    threshold: 0.000000
```

### 5. 查询告警事件

```bash
curl http://localhost:8080/api/v1/alert-events \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**响应**：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 1,
        "alert_id": "alert-设备离线告警-dev-25422c94-1700712030",
        "rule_id": 1,
        "device_id": "dev-25422c94",
        "severity": "critical",
        "status": "firing",
        "message": "设备离线告警: 当前值 0.00 == 阈值 0.00",
        "triggered_at": "2025-11-23T10:00:30Z"
      }
    ]
  }
}
```

### 6. 恢复设备在线

```sql
-- 将设备状态设置为 online
UPDATE devices SET status = 'online' WHERE device_id = 'dev-25422c94';
```

### 7. 等待评估（最多 30 秒）

**日志输出**：

```
2025-11-23T10:01:00.000+0800    INFO    Alert resolved
    rule: 设备离线告警
    device_id: dev-25422c94
```

### 8. 再次查询告警事件

```json
{
  "status": "resolved",
  "resolved_at": "2025-11-23T10:01:00Z"
}
```

## 性能优化

### 1. 并发评估

使用 goroutine 并发评估多个规则：

```go
var wg sync.WaitGroup
for _, rule := range rules {
    wg.Add(1)
    go func(r *model.AlertRule) {
        defer wg.Done()
        e.evaluateRule(r)
    }(rule)
}
wg.Wait()
```

### 2. 活跃告警缓存

使用内存映射缓存活跃告警，避免重复查询数据库：

```go
activeAlerts map[uint]map[string]*ActiveAlert // rule_id -> device_id -> alert
```

### 3. 读写锁

使用 `sync.RWMutex` 保护活跃告警映射：

```go
e.activeAlertsMu.RLock()
alert, exists := e.activeAlerts[ruleID][deviceID]
e.activeAlertsMu.RUnlock()
```

## 未来扩展

### 1. VictoriaMetrics 集成

```go
func (e *AlertEngine) queryVictoriaMetrics(query string) ([]MetricResult, error) {
    url := fmt.Sprintf("%s/api/v1/query?query=%s", e.vmURL, url.QueryEscape(query))
    resp, err := http.Get(url)
    // ... 解析响应
}
```

### 2. 持续时间支持

```go
func (e *AlertEngine) checkDuration(rule *model.AlertRule, deviceID string) bool {
    // 查询时间范围内的数据
    start := time.Now().Add(-time.Duration(rule.Duration) * time.Second)
    results := e.queryRange(rule.Condition, start, time.Now())
    
    // 检查所有数据点
    for _, result := range results {
        if !e.checkCondition(result.Value, operator, threshold) {
            return false
        }
    }
    return true
}
```

### 3. 通知集成

```go
func (e *AlertEngine) sendNotification(event *model.AlertEvent, rule *model.AlertRule) {
    // 根据 notification_config 发送通知
    if channels, ok := rule.NotificationConfig["channels"].([]string); ok {
        for _, channel := range channels {
            switch channel {
            case "email":
                e.sendEmail(event, rule)
            case "webhook":
                e.sendWebhook(event, rule)
            case "dingtalk":
                e.sendDingTalk(event, rule)
            }
        }
    }
}
```

### 4. 告警抑制和静默

```go
// 检查告警是否被抑制
func (e *AlertEngine) isInhibited(rule *model.AlertRule) bool {
    // 检查 inhibit_rules
}

// 检查告警是否在静默期
func (e *AlertEngine) isSilenced(rule *model.AlertRule) bool {
    // 检查 mute_periods
}
```

## 总结

✅ **已实现**：
- 告警引擎核心逻辑
- 规则评估循环
- 条件解析和检查
- 告警触发和解决
- 活跃告警管理
- 主程序集成

⏳ **待实现**：
- VictoriaMetrics API 集成
- 持续时间检查
- 通知功能
- 告警抑制
- 告警静默

🎉 **现在告警功能已经可以正常工作了！**

