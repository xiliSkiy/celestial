# 端到端数据流与模块协作详解 - Part 2

> 本文档是《09-端到端数据流与模块协作详解.md》的第二部分，重点讲解数据查询、前端展示、时序数据库使用和告警处理。

## 目录

- [3.5 数据查询与展示流程](#35-数据查询与展示流程)
- [3.6 告警触发与处理流程](#36-告警触发与处理流程)
- [4. 模块职责与边界](#4-模块职责与边界)
- [5. 数据库设计与数据流](#5-数据库设计与数据流)
- [6. 时序数据库使用详解](#6-时序数据库使用详解)
- [7. 前端数据消费场景](#7-前端数据消费场景)
- [8. 关键技术决策说明](#8-关键技术决策说明)

---

## 3.5 数据查询与展示流程

### 3.5.1 业务场景

运维人员需要在监控大盘查看设备的实时状态、历史趋势、性能指标等信息。

### 3.5.2 系统架构中的数据消费路径

```
┌──────────────────────────────────────────────────────────────┐
│                        前端 (Vue 3)                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │ 设备列表   │  │ 监控大盘   │  │ 告警列表   │            │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘            │
└────────┼────────────────┼────────────────┼───────────────────┘
         │                │                │
         │ ①查询元数据    │ ②查询时序数据  │ ③查询告警事件
         │                │                │
         ▼                ▼                ▼
┌──────────────────────────────────────────────────────────────┐
│                    中心端 API (Gin)                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐            │
│  │ Device API │  │Dashboard API│  │ Alert API  │            │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘            │
└────────┼────────────────┼────────────────┼───────────────────┘
         │                │                │
         ▼                │                ▼
   ┌──────────┐          │          ┌──────────┐
   │PostgreSQL│          │          │PostgreSQL│
   │(元数据)  │          │          │(告警事件)│
   └──────────┘          │          └──────────┘
                         │
                         ▼
                  ┌─────────────┐
                  │VictoriaMetrics│
                  │  (时序数据)   │
                  └─────────────┘
```

### 3.5.3 场景 1：设备列表页面

**前端需求**：显示所有设备的基本信息和最新状态

**数据来源**：
- 设备基本信息 → PostgreSQL (`devices` 表)
- 设备在线状态 → Redis (采集端心跳更新)
- 最新采集时间 → PostgreSQL (`collection_tasks.last_executed_at`)

**API 调用**：

```http
GET /api/v1/devices?page=1&size=20&keyword=服务器
Authorization: Bearer <token>
```

**后端处理**：

```go
// gravital-core/internal/api/handler/device_handler.go
func (h *DeviceHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
    keyword := c.Query("keyword")
    
    // 1. 从 PostgreSQL 查询设备列表
    devices, total, err := h.service.ListDevices(c.Request.Context(), &ListDevicesRequest{
        Page:    page,
        Size:    size,
        Keyword: keyword,
    })
    
    // 2. 批量查询设备在线状态（从 Redis）
    for _, device := range devices {
        // 检查是否有采集端在监控此设备
        sentinel, _ := h.sentinelService.GetSentinelByID(c.Request.Context(), device.SentinelID)
        if sentinel != nil && sentinel.Status == "online" {
            device.Status = "online"
        } else {
            device.Status = "offline"
        }
    }
    
    SuccessResponse(c, gin.H{
        "items": devices,
        "total": total,
        "page":  page,
        "size":  size,
    })
}
```

**前端展示**：

```vue
<template>
  <el-table :data="devices" v-loading="loading">
    <el-table-column prop="name" label="设备名称" />
    <el-table-column prop="device_type" label="类型" />
    <el-table-column label="状态">
      <template #default="{ row }">
        <StatusBadge :status="row.status" />
      </template>
    </el-table-column>
    <el-table-column prop="connection_config.host" label="IP地址" />
    <el-table-column prop="last_seen" label="最后上报" />
  </el-table>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { deviceApi } from '@/api/device'

const devices = ref([])
const loading = ref(false)

const fetchDevices = async () => {
  loading.value = true
  try {
    const res = await deviceApi.getDevices({ page: 1, size: 20 })
    devices.value = res.items
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchDevices()
  // 每 30 秒刷新一次
  setInterval(fetchDevices, 30000)
})
</script>
```

### 3.5.4 场景 2：监控大盘 - 实时指标展示

**前端需求**：显示设备的实时 CPU、内存、网络等指标

**数据来源**：VictoriaMetrics (时序数据库)

**关键问题**：前端如何从时序库获取数据？

#### 方案 A：前端直连时序库（推荐用于 Grafana）

```
前端 ──PromQL──> VictoriaMetrics
```

**优点**：
- 查询灵活，可以使用完整的 PromQL 功能
- 减轻中心端压力
- Grafana 原生支持

**缺点**：
- 需要暴露时序库端口（安全风险）
- 前端需要学习 PromQL
- 跨域问题

#### 方案 B：通过中心端代理（当前实现）

```
前端 ──HTTP──> 中心端 API ──PromQL──> VictoriaMetrics
```

**API 设计**：

```http
POST /api/v1/metrics/query
Content-Type: application/json

{
  "query": "ping_rtt_avg_ms{device_id='dev-25422c94'}",
  "start": 1700500000,
  "end": 1700503600,
  "step": "60s"
}
```

**后端实现**：

```go
// gravital-core/internal/api/handler/metrics_handler.go
type MetricsHandler struct {
    vmClient *victoriametrics.Client
    logger   *zap.Logger
}

func (h *MetricsHandler) Query(c *gin.Context) {
    var req struct {
        Query string `json:"query" binding:"required"`
        Start int64  `json:"start"`
        End   int64  `json:"end"`
        Step  string `json:"step"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
        return
    }
    
    // 调用 VictoriaMetrics API
    result, err := h.vmClient.QueryRange(c.Request.Context(), &victoriametrics.QueryRangeRequest{
        Query: req.Query,
        Start: time.Unix(req.Start, 0),
        End:   time.Unix(req.End, 0),
        Step:  req.Step,
    })
    
    if err != nil {
        h.logger.Error("Failed to query metrics", zap.Error(err))
        ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
        return
    }
    
    SuccessResponse(c, result)
}
```

**VictoriaMetrics Client 实现**：

```go
// gravital-core/internal/tsdb/victoriametrics/client.go
type Client struct {
    baseURL string
    client  *http.Client
}

func (c *Client) QueryRange(ctx context.Context, req *QueryRangeRequest) (*QueryResult, error) {
    // 构造查询 URL
    u, _ := url.Parse(c.baseURL + "/api/v1/query_range")
    q := u.Query()
    q.Set("query", req.Query)
    q.Set("start", strconv.FormatInt(req.Start.Unix(), 10))
    q.Set("end", strconv.FormatInt(req.End.Unix(), 10))
    q.Set("step", req.Step)
    u.RawQuery = q.Encode()
    
    // 发送请求
    httpReq, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
    resp, err := c.client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 解析响应
    var result QueryResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

**前端调用**：

```typescript
// src/api/metrics.ts
export const metricsApi = {
  // 查询时间范围内的指标
  queryRange(params: {
    query: string
    start: number
    end: number
    step?: string
  }) {
    return request.post('/v1/metrics/query', params)
  },
  
  // 查询最新值
  queryCurrent(query: string) {
    return request.post('/v1/metrics/query', {
      query,
      start: Math.floor(Date.now() / 1000) - 300, // 最近 5 分钟
      end: Math.floor(Date.now() / 1000),
      step: '60s'
    })
  }
}
```

**前端组件**：

```vue
<template>
  <div class="metric-chart">
    <h3>{{ title }}</h3>
    <div ref="chartRef" style="width: 100%; height: 300px;"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import * as echarts from 'echarts'
import { metricsApi } from '@/api/metrics'

const props = defineProps({
  deviceId: String,
  metricName: String,
  title: String
})

const chartRef = ref(null)
let chart = null

const fetchData = async () => {
  const now = Math.floor(Date.now() / 1000)
  const start = now - 3600 // 最近 1 小时
  
  const res = await metricsApi.queryRange({
    query: `${props.metricName}{device_id="${props.deviceId}"}`,
    start,
    end: now,
    step: '60s'
  })
  
  // 解析 VictoriaMetrics 响应
  if (res.data.result && res.data.result.length > 0) {
    const values = res.data.result[0].values
    const times = values.map(v => new Date(v[0] * 1000))
    const data = values.map(v => parseFloat(v[1]))
    
    // 更新图表
    chart.setOption({
      xAxis: { type: 'time', data: times },
      yAxis: { type: 'value' },
      series: [{
        type: 'line',
        data: data,
        smooth: true
      }]
    })
  }
}

onMounted(() => {
  chart = echarts.init(chartRef.value)
  fetchData()
  
  // 每分钟刷新
  const interval = setInterval(fetchData, 60000)
  onBeforeUnmount(() => clearInterval(interval))
})
</script>
```

### 3.5.5 场景 3：设备详情页 - 综合信息展示

**前端需求**：显示设备的完整信息（基本信息 + 实时指标 + 告警历史）

**数据来源**：
- 设备基本信息 → PostgreSQL
- 实时指标 → VictoriaMetrics
- 告警历史 → PostgreSQL
- 采集任务列表 → PostgreSQL

**API 调用序列**：

```javascript
// 1. 获取设备基本信息
const device = await deviceApi.getDevice(deviceId)

// 2. 获取设备关联的采集任务
const tasks = await taskApi.getTasks({ device_id: deviceId })

// 3. 获取最新指标值（并发查询多个指标）
const [cpuData, memData, pingData] = await Promise.all([
  metricsApi.queryCurrent(`cpu_usage_percent{device_id="${deviceId}"}`),
  metricsApi.queryCurrent(`memory_used_percent{device_id="${deviceId}"}`),
  metricsApi.queryCurrent(`ping_rtt_avg_ms{device_id="${deviceId}"}`)
])

// 4. 获取告警历史
const alerts = await alertApi.getAlerts({ 
  device_id: deviceId,
  page: 1,
  size: 10
})
```

**页面布局**：

```vue
<template>
  <div class="device-detail">
    <!-- 基本信息卡片 -->
    <el-card class="info-card">
      <h2>{{ device.name }}</h2>
      <el-descriptions :column="2">
        <el-descriptions-item label="设备ID">{{ device.device_id }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ device.device_type }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ device.connection_config.host }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <StatusBadge :status="device.status" />
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
    
    <!-- 实时指标卡片 -->
    <el-row :gutter="20" class="metrics-row">
      <el-col :span="8">
        <MetricCard 
          title="CPU 使用率" 
          :value="cpuUsage" 
          unit="%" 
          :trend="cpuTrend" 
        />
      </el-col>
      <el-col :span="8">
        <MetricCard 
          title="内存使用率" 
          :value="memUsage" 
          unit="%" 
          :trend="memTrend" 
        />
      </el-col>
      <el-col :span="8">
        <MetricCard 
          title="Ping 延迟" 
          :value="pingRtt" 
          unit="ms" 
          :trend="pingTrend" 
        />
      </el-col>
    </el-row>
    
    <!-- 历史趋势图表 -->
    <el-card class="chart-card">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="CPU" name="cpu">
          <MetricChart 
            :device-id="deviceId" 
            metric-name="cpu_usage_percent" 
            title="CPU 使用率趋势" 
          />
        </el-tab-pane>
        <el-tab-pane label="内存" name="memory">
          <MetricChart 
            :device-id="deviceId" 
            metric-name="memory_used_percent" 
            title="内存使用率趋势" 
          />
        </el-tab-pane>
        <el-tab-pane label="网络" name="network">
          <MetricChart 
            :device-id="deviceId" 
            metric-name="ping_rtt_avg_ms" 
            title="网络延迟趋势" 
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>
    
    <!-- 告警历史 -->
    <el-card class="alert-card">
      <h3>告警历史</h3>
      <el-table :data="alerts">
        <el-table-column prop="alert_name" label="告警名称" />
        <el-table-column prop="severity" label="级别" />
        <el-table-column prop="triggered_at" label="触发时间" />
        <el-table-column prop="status" label="状态" />
      </el-table>
    </el-card>
    
    <!-- 采集任务列表 -->
    <el-card class="task-card">
      <h3>采集任务</h3>
      <el-table :data="tasks">
        <el-table-column prop="plugin_name" label="插件" />
        <el-table-column prop="interval_seconds" label="间隔" />
        <el-table-column prop="enabled" label="状态" />
        <el-table-column label="操作">
          <template #default="{ row }">
            <el-button size="small" @click="triggerTask(row.id)">立即执行</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
```

---

## 3.6 告警触发与处理流程

### 3.6.1 业务场景

当设备的某个指标超过阈值时，系统自动触发告警，通知运维人员处理。

### 3.6.2 告警架构

```
┌─────────────────────────────────────────────────────────────┐
│                   告警处理流程                               │
└─────────────────────────────────────────────────────────────┘

时序数据 ──> 告警规则引擎 ──> 告警事件 ──> 通知渠道
                  │                │            │
                  │                │            ├──> 邮件
                  │                │            ├──> 钉钉
                  │                │            ├──> 企业微信
                  │                │            └──> Webhook
                  │                │
                  │                └──> PostgreSQL (告警历史)
                  │
                  └──> 定时查询 VictoriaMetrics
```

### 3.6.3 告警规则配置

**数据库表结构**：

```sql
CREATE TABLE alert_rules (
    id SERIAL PRIMARY KEY,
    rule_name VARCHAR(255) NOT NULL,
    device_id VARCHAR(64),           -- NULL 表示全局规则
    metric_name VARCHAR(255) NOT NULL,
    condition VARCHAR(50) NOT NULL,  -- >, <, >=, <=, ==, !=
    threshold FLOAT NOT NULL,
    duration_seconds INT DEFAULT 60, -- 持续时间
    severity VARCHAR(20) NOT NULL,   -- critical, warning, info
    enabled BOOLEAN DEFAULT true,
    notify_channels JSONB,           -- ["email", "dingtalk"]
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

**规则示例**：

```json
{
  "rule_name": "服务器 CPU 过高",
  "device_id": "dev-25422c94",
  "metric_name": "cpu_usage_percent",
  "condition": ">",
  "threshold": 80,
  "duration_seconds": 300,  // 持续 5 分钟
  "severity": "warning",
  "enabled": true,
  "notify_channels": ["email", "dingtalk"]
}
```

### 3.6.4 告警检测流程

#### 方案 A：中心端定时查询（当前实现）

```go
// gravital-core/internal/alert/engine.go
type AlertEngine struct {
    ruleRepo    repository.AlertRuleRepository
    eventRepo   repository.AlertEventRepository
    vmClient    *victoriametrics.Client
    notifier    *Notifier
    logger      *zap.Logger
}

func (e *AlertEngine) Start() {
    // 每分钟检查一次所有规则
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            e.checkAllRules()
        case <-e.ctx.Done():
            return
        }
    }
}

func (e *AlertEngine) checkAllRules() {
    // 1. 获取所有启用的规则
    rules, err := e.ruleRepo.ListEnabled(context.Background())
    if err != nil {
        e.logger.Error("Failed to list rules", zap.Error(err))
        return
    }
    
    // 2. 并发检查每个规则
    var wg sync.WaitGroup
    for _, rule := range rules {
        wg.Add(1)
        go func(r *model.AlertRule) {
            defer wg.Done()
            e.checkRule(r)
        }(rule)
    }
    wg.Wait()
}

func (e *AlertEngine) checkRule(rule *model.AlertRule) {
    // 1. 构造 PromQL 查询
    query := fmt.Sprintf(`%s{device_id="%s"}`, rule.MetricName, rule.DeviceID)
    if rule.DeviceID == "" {
        query = rule.MetricName  // 全局规则
    }
    
    // 2. 查询最近的数据点
    now := time.Now()
    start := now.Add(-time.Duration(rule.DurationSeconds) * time.Second)
    
    result, err := e.vmClient.QueryRange(context.Background(), &QueryRangeRequest{
        Query: query,
        Start: start,
        End:   now,
        Step:  "60s",
    })
    
    if err != nil {
        e.logger.Error("Failed to query metrics", 
            zap.String("rule", rule.RuleName),
            zap.Error(err))
        return
    }
    
    // 3. 检查是否满足告警条件
    if e.evaluateCondition(result, rule) {
        e.triggerAlert(rule, result)
    } else {
        e.resolveAlert(rule)
    }
}

func (e *AlertEngine) evaluateCondition(result *QueryResult, rule *model.AlertRule) bool {
    if len(result.Data.Result) == 0 {
        return false
    }
    
    // 检查所有数据点是否都满足条件
    values := result.Data.Result[0].Values
    if len(values) == 0 {
        return false
    }
    
    // 至少需要 duration_seconds / step 个数据点
    minPoints := rule.DurationSeconds / 60
    if len(values) < minPoints {
        return false
    }
    
    // 检查最近的 N 个点是否都满足条件
    for i := len(values) - minPoints; i < len(values); i++ {
        value := values[i][1].(float64)
        if !e.compareValue(value, rule.Condition, rule.Threshold) {
            return false  // 有一个不满足就不告警
        }
    }
    
    return true
}

func (e *AlertEngine) compareValue(value float64, condition string, threshold float64) bool {
    switch condition {
    case ">":
        return value > threshold
    case ">=":
        return value >= threshold
    case "<":
        return value < threshold
    case "<=":
        return value <= threshold
    case "==":
        return value == threshold
    case "!=":
        return value != threshold
    default:
        return false
    }
}

func (e *AlertEngine) triggerAlert(rule *model.AlertRule, result *QueryResult) {
    // 1. 检查是否已经有未解决的告警
    existing, _ := e.eventRepo.GetActiveByRule(context.Background(), rule.ID)
    if existing != nil {
        return  // 已经告警过了，不重复告警
    }
    
    // 2. 创建告警事件
    currentValue := result.Data.Result[0].Values[len(result.Data.Result[0].Values)-1][1].(float64)
    
    event := &model.AlertEvent{
        RuleID:      rule.ID,
        RuleName:    rule.RuleName,
        DeviceID:    rule.DeviceID,
        MetricName:  rule.MetricName,
        Severity:    rule.Severity,
        CurrentValue: currentValue,
        Threshold:   rule.Threshold,
        Status:      "firing",
        TriggeredAt: time.Now(),
        Message:     fmt.Sprintf("%s: 当前值 %.2f %s 阈值 %.2f", 
            rule.RuleName, currentValue, rule.Condition, rule.Threshold),
    }
    
    if err := e.eventRepo.Create(context.Background(), event); err != nil {
        e.logger.Error("Failed to create alert event", zap.Error(err))
        return
    }
    
    // 3. 发送通知
    e.notifier.Send(event, rule.NotifyChannels)
    
    e.logger.Info("Alert triggered",
        zap.String("rule", rule.RuleName),
        zap.Float64("value", currentValue),
        zap.Float64("threshold", rule.Threshold))
}

func (e *AlertEngine) resolveAlert(rule *model.AlertRule) {
    // 查找未解决的告警
    event, err := e.eventRepo.GetActiveByRule(context.Background(), rule.ID)
    if err != nil || event == nil {
        return  // 没有活跃告警
    }
    
    // 更新告警状态为已解决
    event.Status = "resolved"
    event.ResolvedAt = timePtr(time.Now())
    
    if err := e.eventRepo.Update(context.Background(), event); err != nil {
        e.logger.Error("Failed to resolve alert", zap.Error(err))
        return
    }
    
    // 发送恢复通知
    e.notifier.SendResolved(event, rule.NotifyChannels)
    
    e.logger.Info("Alert resolved",
        zap.String("rule", rule.RuleName))
}
```

#### 方案 B：基于 VictoriaMetrics 的 vmalert（推荐）

VictoriaMetrics 提供了 `vmalert` 组件，专门用于告警规则评估：

```yaml
# vmalert 配置
groups:
  - name: device_alerts
    interval: 1m
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率过高"
          description: "设备 {{ $labels.device_id }} CPU 使用率为 {{ $value }}%"
```

**优势**：
- 性能更好（原生支持）
- 功能更强大（支持复杂的 PromQL 表达式）
- 可靠性更高（成熟的开源组件）

**集成方式**：

```go
// 中心端监听 vmalert 的 webhook
func (h *AlertHandler) ReceiveWebhook(c *gin.Context) {
    var alerts []struct {
        Status      string            `json:"status"`      // firing, resolved
        Labels      map[string]string `json:"labels"`
        Annotations map[string]string `json:"annotations"`
        StartsAt    time.Time         `json:"startsAt"`
        EndsAt      time.Time         `json:"endsAt"`
    }
    
    if err := c.ShouldBindJSON(&alerts); err != nil {
        ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
        return
    }
    
    // 处理每个告警
    for _, alert := range alerts {
        if alert.Status == "firing" {
            h.handleFiringAlert(alert)
        } else if alert.Status == "resolved" {
            h.handleResolvedAlert(alert)
        }
    }
    
    SuccessResponse(c, gin.H{"received": len(alerts)})
}
```

### 3.6.5 告警通知

```go
// gravital-core/internal/alert/notifier.go
type Notifier struct {
    emailSender    *EmailSender
    dingtalkSender *DingtalkSender
    wechatSender   *WechatSender
    webhookSender  *WebhookSender
    logger         *zap.Logger
}

func (n *Notifier) Send(event *model.AlertEvent, channels []string) {
    for _, channel := range channels {
        switch channel {
        case "email":
            n.sendEmail(event)
        case "dingtalk":
            n.sendDingtalk(event)
        case "wechat":
            n.sendWechat(event)
        case "webhook":
            n.sendWebhook(event)
        }
    }
}

func (n *Notifier) sendDingtalk(event *model.AlertEvent) error {
    message := map[string]interface{}{
        "msgtype": "markdown",
        "markdown": map[string]string{
            "title": "告警通知",
            "text": fmt.Sprintf(`
### %s
- **级别**: %s
- **设备**: %s
- **指标**: %s
- **当前值**: %.2f
- **阈值**: %.2f
- **时间**: %s
            `, 
                event.RuleName,
                event.Severity,
                event.DeviceID,
                event.MetricName,
                event.CurrentValue,
                event.Threshold,
                event.TriggeredAt.Format("2006-01-02 15:04:05"),
            ),
        },
    }
    
    data, _ := json.Marshal(message)
    resp, err := http.Post(n.dingtalkWebhook, "application/json", bytes.NewReader(data))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

---


