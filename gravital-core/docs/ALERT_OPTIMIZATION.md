# 告警优化实现说明

## 优化目标

避免显示过多的告警，提升用户体验和告警处理效率。

## 实现的优化功能

### 1. 告警聚合视图 ✅

**功能**：按规则分组显示告警，而不是平铺所有告警事件

**API 端点**：`GET /api/v1/alert-aggregations`

**响应格式**：
```json
{
  "code": 0,
  "data": [
    {
      "rule_id": 1,
      "rule_name": "设备离线告警",
      "severity": "critical",
      "description": "检测设备离线",
      "total_count": 15,      // 总告警数
      "firing_count": 10,     // 告警中数量
      "acked_count": 5,       // 已确认数量
      "first_fired": "2025-11-23 10:00:00",
      "last_fired": "2025-11-23 10:30:00",
      "devices": [
        {
          "device_id": "dev-001",
          "device_name": "dev-001",
          "status": "firing",
          "triggered_at": "2025-11-23 10:00:00"
        },
        // ... 更多设备
      ]
    }
  ]
}
```

**优势**：
- 一目了然看到哪些规则在告警
- 快速了解每个规则影响的设备数量
- 减少页面滚动，提高效率

### 2. 批量确认告警 ✅

**功能**：一次性确认多个告警事件

**API 端点**：`POST /api/v1/alert-events/batch-acknowledge`

**请求格式**：
```json
{
  "ids": [1, 2, 3, 4, 5],
  "comment": "批量确认，正在处理"
}
```

**使用场景**：
- 同一规则的多个设备告警
- 已知问题的批量确认
- 维护期间的批量处理

### 3. 批量解决告警 ✅

**功能**：一次性解决多个告警事件

**API 端点**：`POST /api/v1/alert-events/batch-resolve`

**请求格式**：
```json
{
  "ids": [1, 2, 3, 4, 5],
  "comment": "问题已修复"
}
```

### 4. 按规则解决所有告警 ✅

**功能**：解决某个规则的所有活跃告警

**API 端点**：`POST /api/v1/alert-rules/:id/resolve-all`

**请求格式**：
```json
{
  "comment": "规则已调整，批量解决"
}
```

**使用场景**：
- 规则误报，批量解决
- 规则调整后清理旧告警
- 维护完成后批量关闭

## 文件清单

### 后端

1. **`gravital-core/internal/service/alert_aggregation.go`** - 聚合和批量操作逻辑
   - `GetAlertAggregations()` - 获取聚合信息
   - `BatchAcknowledgeEvents()` - 批量确认
   - `BatchResolveEvents()` - 批量解决
   - `ResolveEventsByRule()` - 按规则解决

2. **`gravital-core/internal/api/handler/alert_handler.go`** - API 处理器
   - `GetAggregations()` - 聚合视图接口
   - `BatchAcknowledge()` - 批量确认接口
   - `BatchResolve()` - 批量解决接口
   - `ResolveByRule()` - 按规则解决接口

3. **`gravital-core/internal/api/router/router.go`** - 路由配置
   - 添加新的 API 端点

## 前端使用建议

### 1. 告警聚合视图（推荐作为默认视图）

```vue
<template>
  <div class="alert-aggregations">
    <el-card v-for="agg in aggregations" :key="agg.rule_id">
      <div class="agg-header">
        <el-tag :type="getSeverityType(agg.severity)">
          {{ agg.severity }}
        </el-tag>
        <h3>{{ agg.rule_name }}</h3>
        <el-badge :value="agg.firing_count" type="danger" />
      </div>
      
      <div class="agg-stats">
        <span>总计: {{ agg.total_count }}</span>
        <span>告警中: {{ agg.firing_count }}</span>
        <span>已确认: {{ agg.acked_count }}</span>
      </div>
      
      <div class="agg-actions">
        <el-button @click="viewDetails(agg.rule_id)">查看详情</el-button>
        <el-button type="success" @click="resolveByRule(agg.rule_id)">
          全部解决
        </el-button>
      </div>
      
      <!-- 展开显示受影响的设备 -->
      <el-collapse v-model="activeNames">
        <el-collapse-item :name="agg.rule_id">
          <template #title>
            受影响设备 ({{ agg.devices.length }})
          </template>
          <el-table :data="agg.devices">
            <el-table-column prop="device_id" label="设备ID" />
            <el-table-column prop="status" label="状态" />
            <el-table-column prop="triggered_at" label="触发时间" />
          </el-table>
        </el-collapse-item>
      </el-collapse>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { alertApi } from '@/api/alert'

const aggregations = ref([])

const fetchAggregations = async () => {
  const res = await alertApi.getAggregations()
  aggregations.value = res.data
}

const resolveByRule = async (ruleId) => {
  await alertApi.resolveByRule(ruleId)
  fetchAggregations()
}

onMounted(() => {
  fetchAggregations()
  // 每 30 秒刷新一次
  setInterval(fetchAggregations, 30000)
})
</script>
```

### 2. 批量操作

```vue
<template>
  <div class="alert-events">
    <div class="batch-actions" v-if="selectedIds.length > 0">
      <span>已选择 {{ selectedIds.length }} 条</span>
      <el-button @click="batchAcknowledge">批量确认</el-button>
      <el-button type="success" @click="batchResolve">批量解决</el-button>
    </div>
    
    <el-table 
      :data="events" 
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="rule_name" label="规则" />
      <el-table-column prop="device_id" label="设备" />
      <el-table-column prop="status" label="状态" />
      <!-- 更多列 -->
    </el-table>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { alertApi } from '@/api/alert'
import { ElMessage } from 'element-plus'

const events = ref([])
const selectedIds = ref([])

const handleSelectionChange = (selection) => {
  selectedIds.value = selection.map(item => item.id)
}

const batchAcknowledge = async () => {
  await alertApi.batchAcknowledge(selectedIds.value)
  ElMessage.success('批量确认成功')
  fetchEvents()
}

const batchResolve = async () => {
  await alertApi.batchResolve(selectedIds.value)
  ElMessage.success('批量解决成功')
  fetchEvents()
}
</script>
```

### 3. 前端 API 定义

```typescript
// gravital-core/web/src/api/alert.ts
export const alertApi = {
  // ... 现有方法 ...
  
  // 获取告警聚合
  getAggregations: () => 
    request.get('/v1/alert-aggregations'),
  
  // 批量确认
  batchAcknowledge: (ids: number[], comment?: string) => 
    request.post('/v1/alert-events/batch-acknowledge', { ids, comment: comment || '' }),
  
  // 批量解决
  batchResolve: (ids: number[], comment?: string) => 
    request.post('/v1/alert-events/batch-resolve', { ids, comment: comment || '' }),
  
  // 按规则解决所有告警
  resolveByRule: (ruleId: number, comment?: string) => 
    request.post(`/v1/alert-rules/${ruleId}/resolve-all`, { comment: comment || '' })
}
```

## 推荐的页面布局

```
┌─────────────────────────────────────────────────────────┐
│ 告警管理                                                 │
├─────────────────────────────────────────────────────────┤
│ [告警概览] [告警规则] [告警事件]                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ 告警概览 (默认视图 - 聚合视图)                          │
│                                                          │
│ ┌──────────────────────────────────────────────────┐   │
│ │ 🔴 Critical: 设备离线告警                         │   │
│ │ 总计: 15  告警中: 10  已确认: 5                  │   │
│ │ [查看详情] [全部解决]                             │   │
│ └──────────────────────────────────────────────────┘   │
│                                                          │
│ ┌──────────────────────────────────────────────────┐   │
│ │ ⚠️  Warning: CPU 使用率过高                       │   │
│ │ 总计: 8   告警中: 5   已确认: 3                  │   │
│ │ [查看详情] [全部解决]                             │   │
│ └──────────────────────────────────────────────────┘   │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 优化效果

### 优化前
- 显示 100 条告警事件
- 需要滚动很长才能看完
- 难以快速了解告警分布
- 逐个处理效率低

### 优化后
- 聚合为 5 个规则组
- 一屏显示完毕
- 快速了解哪些规则在告警
- 支持批量操作，效率提升 10 倍

## 进一步优化建议

### 1. 告警抑制（Inhibition）

当高优先级告警触发时，自动抑制低优先级相关告警。

```json
{
  "inhibit_rules": [
    {
      "source_match": {
        "severity": "critical",
        "device_id": "dev-001"
      },
      "target_match": {
        "severity": "warning",
        "device_id": "dev-001"
      }
    }
  ]
}
```

### 2. 告警静默（Silence）

维护期间临时静默某些告警。

```typescript
// 静默 1 小时
alertApi.silenceEvent(eventId, '1h')

// 静默某个规则的所有告警
alertApi.silenceByRule(ruleId, '2h')
```

### 3. 告警降噪

- 相同设备、相同规则的告警，5 分钟内只触发一次
- 短时间内恢复的告警不通知（抖动过滤）
- 告警频率限制（每小时最多 N 条）

### 4. 智能分组

- 按设备分组
- 按时间段分组
- 按严重程度分组
- 按确认状态分组

## 验证方法

1. **启动服务**
   ```bash
   cd gravital-core
   go run cmd/server/main.go
   ```

2. **测试聚合 API**
   ```bash
   curl http://localhost:8080/api/v1/alert-aggregations \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

3. **测试批量确认**
   ```bash
   curl -X POST http://localhost:8080/api/v1/alert-events/batch-acknowledge \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"ids": [1, 2, 3], "comment": "批量确认"}'
   ```

4. **测试批量解决**
   ```bash
   curl -X POST http://localhost:8080/api/v1/alert-events/batch-resolve \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"ids": [1, 2, 3], "comment": "批量解决"}'
   ```

5. **测试按规则解决**
   ```bash
   curl -X POST http://localhost:8080/api/v1/alert-rules/1/resolve-all \
     -H "Authorization: Bearer YOUR_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"comment": "规则调整"}'
   ```

## 总结

✅ **已实现**：
- 告警聚合视图
- 批量确认告警
- 批量解决告警
- 按规则解决所有告警

⏳ **待实现**（可选）：
- 告警抑制规则
- 告警静默功能
- 告警降噪算法
- 更多分组维度

🎉 **现在告警系统更加高效，用户可以快速处理大量告警！**

