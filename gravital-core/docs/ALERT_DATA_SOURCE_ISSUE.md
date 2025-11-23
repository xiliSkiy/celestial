# 告警引擎数据源不一致问题

## 问题描述

用户报告：设备 `dev-25422c94` 在时序数据库中显示一直在线（`device_status = 1`），但告警引擎却报告设备离线（`当前值 0.00 != 阈值 1.00`），并且告警立即被自动解决。

## 根本原因

**告警引擎和时序数据库使用了不同的数据源！**

### 1. 时序数据库（VictoriaMetrics）

- **数据来源**：Sentinel 实时采集并上报
- **更新频率**：每 30 秒（采集任务间隔）
- **数据特点**：实时反映设备当前状态
- **示例数据**：
  ```
  device_status{device_id="dev-25422c94", device_type="router"} = 1.0
  ```

### 2. PostgreSQL 数据库（devices 表）

- **数据来源**：
  1. 设备心跳上报时更新 `last_seen` 字段
  2. DeviceMonitor 定时检查（每 1 分钟）
  3. 如果 `last_seen` 超过 5 分钟，则将 `status` 标记为 `offline`
  
- **更新频率**：
  - 心跳上报：不定期（取决于设备实现）
  - 定时检查：每 1 分钟
  
- **数据特点**：可能存在延迟，不一定实时

### 3. 告警引擎当前实现

**位置**：`gravital-core/internal/alert/engine/engine.go:231-265`

```go
func (e *AlertEngine) queryMetric(query string) ([]MetricResult, error) {
    // TODO: 这里应该调用 VictoriaMetrics API
    // 目前简化实现：从数据库查询设备状态
    
    if metricName == "device_status" {
        var devices []model.Device
        query := e.db.Model(&model.Device{})
        
        // 从 PostgreSQL 查询设备状态
        if err := query.Find(&devices).Error; err != nil {
            return nil, err
        }

        results := make([]MetricResult, 0, len(devices))
        for _, device := range devices {
            value := 0.0
            if device.Status == "online" {  // ❌ 这里读取的是 PostgreSQL 中的状态
                value = 1.0
            }
            results = append(results, MetricResult{...})
        }
        return results, nil
    }
}
```

**问题**：告警引擎查询的是 PostgreSQL 的 `devices.status` 字段，而不是时序数据库的实时数据！

## 问题场景

### 场景 1：设备实际在线，但数据库显示离线

1. 设备正常采集，Sentinel 持续上报 `device_status = 1` 到时序数据库 ✅
2. 但设备没有发送心跳到 Gravital Core（或心跳间隔 > 5 分钟）
3. DeviceMonitor 检测到 `last_seen` 超时，将 PostgreSQL 中的 `status` 设为 `offline` ❌
4. 告警引擎查询 PostgreSQL，获取到 `status = offline`，转换为 `value = 0.0`
5. 告警条件 `device_status != 1`，即 `0.0 != 1`，触发告警 ❌
6. 下一次评估时，如果 PostgreSQL 状态仍是 `offline`，条件仍然满足，告警继续触发
7. 如果 PostgreSQL 状态变为 `online`，条件不满足，告警被自动解决

### 场景 2：设备实际离线，但数据库显示在线

1. 设备停止采集，Sentinel 停止上报 `device_status` 到时序数据库
2. 但设备最近发送过心跳（< 5 分钟前）
3. PostgreSQL 中的 `status` 仍为 `online` ✅
4. 告警引擎查询 PostgreSQL，获取到 `status = online`，转换为 `value = 1.0`
5. 告警条件 `device_status != 1`，即 `1.0 != 1`，不满足，不触发告警 ❌（漏报）

## 解决方案

### 方案 1：告警引擎改为查询时序数据库（推荐）

修改 `queryMetric` 函数，调用 VictoriaMetrics API 查询实时数据：

```go
func (e *AlertEngine) queryMetric(query string) ([]MetricResult, error) {
    // 解析查询条件
    metricName, filters := e.parseQuery(query)
    
    // 构建 PromQL 查询
    promQL := metricName
    if len(filters) > 0 {
        labels := []string{}
        for k, v := range filters {
            labels = append(labels, fmt.Sprintf("%s=\"%s\"", k, v))
        }
        promQL = fmt.Sprintf("%s{%s}", metricName, strings.Join(labels, ","))
    }
    
    // 查询 VictoriaMetrics
    resp, err := http.Get(fmt.Sprintf("%s/api/v1/query?query=%s", e.config.VMURL, url.QueryEscape(promQL)))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 解析响应
    var vmResp struct {
        Status string `json:"status"`
        Data   struct {
            ResultType string `json:"resultType"`
            Result     []struct {
                Metric map[string]string `json:"metric"`
                Value  []interface{}     `json:"value"`
            } `json:"result"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&vmResp); err != nil {
        return nil, err
    }
    
    // 转换为 MetricResult
    results := make([]MetricResult, 0, len(vmResp.Data.Result))
    for _, item := range vmResp.Data.Result {
        value, _ := strconv.ParseFloat(item.Value[1].(string), 64)
        results = append(results, MetricResult{
            Labels: item.Metric,
            Value:  value,
        })
    }
    
    return results, nil
}
```

**优点**：
- ✅ 使用实时数据，准确反映设备当前状态
- ✅ 支持所有时序指标，不仅限于 `device_status`
- ✅ 符合告警引擎的设计初衷

**缺点**：
- ❌ 需要实现 VictoriaMetrics API 调用
- ❌ 依赖时序数据库的可用性

### 方案 2：同步 PostgreSQL 状态（不推荐）

让 Sentinel 在上报时序数据的同时，也发送心跳更新 PostgreSQL 的 `devices.last_seen`。

**优点**：
- ✅ 保持当前告警引擎实现不变

**缺点**：
- ❌ 增加系统复杂度
- ❌ 仍然存在延迟
- ❌ 无法支持其他时序指标

### 方案 3：混合方案

- 对于 `device_status` 指标，查询时序数据库
- 对于其他指标，也查询时序数据库
- PostgreSQL 的 `devices.status` 仅用于前端展示

## 实施步骤

### 第 1 步：实现 VictoriaMetrics 查询客户端

创建 `gravital-core/internal/alert/engine/vm_client.go`：

```go
package engine

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
)

type VMClient struct {
    baseURL string
    client  *http.Client
}

func NewVMClient(baseURL string) *VMClient {
    return &VMClient{
        baseURL: baseURL,
        client:  &http.Client{},
    }
}

func (c *VMClient) Query(promQL string) ([]MetricResult, error) {
    queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", c.baseURL, url.QueryEscape(promQL))
    
    resp, err := c.client.Get(queryURL)
    if err != nil {
        return nil, fmt.Errorf("failed to query VM: %w", err)
    }
    defer resp.Body.Close()
    
    var vmResp struct {
        Status string `json:"status"`
        Data   struct {
            ResultType string `json:"resultType"`
            Result     []struct {
                Metric map[string]string `json:"metric"`
                Value  []interface{}     `json:"value"`
            } `json:"result"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&vmResp); err != nil {
        return nil, fmt.Errorf("failed to decode VM response: %w", err)
    }
    
    if vmResp.Status != "success" {
        return nil, fmt.Errorf("VM query failed: %s", vmResp.Status)
    }
    
    results := make([]MetricResult, 0, len(vmResp.Data.Result))
    for _, item := range vmResp.Data.Result {
        valueStr, ok := item.Value[1].(string)
        if !ok {
            continue
        }
        value, err := strconv.ParseFloat(valueStr, 64)
        if err != nil {
            continue
        }
        
        results = append(results, MetricResult{
            Labels: item.Metric,
            Value:  value,
        })
    }
    
    return results, nil
}
```

### 第 2 步：修改告警引擎

修改 `gravital-core/internal/alert/engine/engine.go`：

```go
type AlertEngine struct {
    db           *gorm.DB
    logger       *zap.Logger
    config       *Config
    vmClient     *VMClient  // 新增
    ticker       *time.Ticker
    done         chan struct{}
    activeAlerts map[string]*model.AlertEvent
    mu           sync.RWMutex
}

func NewAlertEngine(db *gorm.DB, logger *zap.Logger, config *Config) *AlertEngine {
    return &AlertEngine{
        db:           db,
        logger:       logger,
        config:       config,
        vmClient:     NewVMClient(config.VMURL),  // 新增
        done:         make(chan struct{}),
        activeAlerts: make(map[string]*model.AlertEvent),
    }
}

func (e *AlertEngine) queryMetric(query string) ([]MetricResult, error) {
    // 解析查询条件
    metricName, filters := e.parseQuery(query)
    
    // 构建 PromQL
    promQL := metricName
    if len(filters) > 0 {
        labels := []string{}
        for k, v := range filters {
            labels = append(labels, fmt.Sprintf("%s=\"%s\"", k, v))
        }
        promQL = fmt.Sprintf("%s{%s}", metricName, strings.Join(labels, ","))
    }
    
    // 查询 VictoriaMetrics
    return e.vmClient.Query(promQL)
}
```

### 第 3 步：测试验证

1. 确认 VictoriaMetrics 正在运行
2. 重启 Gravital Core
3. 创建告警规则
4. 观察告警事件是否正确触发和解决

## 临时解决方案

在实施完整方案之前，可以通过以下方式临时解决：

### 选项 1：让 Sentinel 发送心跳

修改 Sentinel，在上报指标后立即发送心跳：

```go
// orbital-sentinels/internal/scheduler/scheduler.go
func (s *Scheduler) executeTask(task *model.CollectionTask) {
    // ... 采集数据 ...
    
    // 上报指标
    s.forwarder.Forward(metrics)
    
    // 发送心跳（临时方案）
    s.sendHeartbeat(task.DeviceID)
}
```

### 选项 2：调整 DeviceMonitor 超时时间

将离线超时时间从 5 分钟调整为更长（如 10 分钟），减少误报：

```go
// gravital-core/cmd/server/main.go
deviceMonitor := service.NewDeviceMonitor(db, logger.Get(), &service.DeviceMonitorConfig{
    CheckInterval:  1 * time.Minute,
    OfflineTimeout: 10 * time.Minute,  // 从 5 分钟改为 10 分钟
})
```

## 总结

- **根本问题**：告警引擎查询 PostgreSQL，而不是时序数据库
- **推荐方案**：修改告警引擎，改为查询 VictoriaMetrics
- **临时方案**：调整 DeviceMonitor 超时时间或让 Sentinel 发送心跳

