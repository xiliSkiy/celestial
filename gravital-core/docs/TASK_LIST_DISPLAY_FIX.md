# 任务列表显示修复说明

## 问题描述

任务管理界面请求接口后，返回了数据，但是以下字段没有显示：
- 任务名称
- 插件类型
- 采集间隔
- 最后执行

**问题原因**:
1. 后端返回的字段名称与前端表格列期望的字段名称不匹配：
   - 后端返回 `plugin_name`，前端期望 `plugin_type`
   - 后端返回 `interval_seconds` (数字)，前端期望 `interval` (字符串)
   - 后端返回 `last_executed_at`，前端期望 `last_execution`
2. 后端返回的数据中没有 `name` 字段（任务名称）
3. 日期时间需要格式化显示

---

## 修复方案

### 1. 添加数据转换函数

**文件**: `web/src/views/Tasks/List.vue`

在 `fetchTasks` 方法之前添加转换函数：

```typescript
// 将秒数转换为时间字符串（如 60 -> "60s", 120 -> "2m"）
const formatSecondsToDuration = (seconds: number | string): string => {
  if (typeof seconds === 'string') {
    return seconds
  }
  if (!seconds || seconds <= 0) return '60s'
  
  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    return `${minutes}m`
  } else {
    const hours = Math.floor(seconds / 3600)
    return `${hours}h`
  }
}

// 格式化日期时间
const formatDateTime = (date: string | null | undefined): string => {
  if (!date) return '-'
  try {
    const d = new Date(date)
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}
```

### 2. 修复 fetchTasks 方法

**文件**: `web/src/views/Tasks/List.vue`

在获取任务列表后，对数据进行转换：

```typescript
// 获取任务列表
const fetchTasks = async () => {
  loading.value = true
  try {
    const res: any = await taskApi.getTasks(query)
    // 转换数据格式以匹配前端表格列
    tasks.value = (res.items || []).map((task: any) => ({
      ...task,
      // 任务名称：使用设备名称 + 插件类型，或使用 task_id
      name: task.device?.name 
        ? `${task.device.name} - ${task.plugin_name || '未知'}`
        : `任务 ${task.task_id}`,
      // 插件类型：后端返回 plugin_name，前端期望 plugin_type
      plugin_type: task.plugin_name || task.plugin_type || '-',
      // 采集间隔：后端返回 interval_seconds (数字)，前端期望 interval (字符串)
      interval: formatSecondsToDuration(task.interval_seconds || task.interval || 60),
      // 最后执行：后端返回 last_executed_at，前端期望 last_execution (格式化)
      last_execution: formatDateTime(task.last_executed_at),
      // 目标设备：显示设备名称
      device_name: task.device?.name || task.device_id || '-'
    }))
    total.value = res.total || 0
  } catch (error) {
    ElMessage.error('获取任务列表失败')
  } finally {
    loading.value = false
  }
}
```

### 3. 更新表格列定义

**文件**: `web/src/views/Tasks/List.vue`

使用模板来显示转换后的数据：

```vue
<el-table-column prop="name" label="任务名称" width="200" show-overflow-tooltip />
<el-table-column label="目标设备" width="180">
  <template #default="{ row }">
    {{ row.device?.name || row.device_id || '-' }}
  </template>
</el-table-column>
<el-table-column prop="sentinel_id" label="Sentinel" width="180" show-overflow-tooltip />
<el-table-column prop="plugin_type" label="插件类型" width="120">
  <template #default="{ row }">
    {{ row.plugin_type || '-' }}
  </template>
</el-table-column>
<el-table-column prop="interval" label="采集间隔" width="120">
  <template #default="{ row }">
    {{ row.interval || '-' }}
  </template>
</el-table-column>
<el-table-column prop="last_execution" label="最后执行" width="180">
  <template #default="{ row }">
    {{ row.last_execution || '-' }}
  </template>
</el-table-column>
```

---

## 数据转换映射

### 后端 → 前端

| 后端字段 | 前端字段 | 转换说明 |
|---------|---------|---------|
| `plugin_name` | `plugin_type` | 字段名称转换 |
| `interval_seconds` (int) | `interval` (string) | 60 → "60s" |
| `last_executed_at` (ISO string) | `last_execution` (formatted) | "2025-11-15T13:49:15Z" → "2025/11/15 13:49:15" |
| `device.name` + `plugin_name` | `name` | 生成任务名称 |

### 任务名称生成规则

1. 如果有设备名称：`设备名称 - 插件类型`
   - 例如：`本地服务器 - ping`
2. 如果没有设备名称：`任务 task_id`
   - 例如：`任务 task-177bfd59`

---

## 修复后的效果

### 任务列表显示

**修复前**:
- ❌ 任务名称：空白
- ❌ 插件类型：空白
- ❌ 采集间隔：空白
- ❌ 最后执行：空白

**修复后**:
- ✅ 任务名称：`本地服务器 - ping`
- ✅ 插件类型：`ping`
- ✅ 采集间隔：`60s`
- ✅ 最后执行：`2025/11/15 13:49:15` 或 `-`（如果未执行）

### 数据示例

**后端返回**:
```json
{
  "plugin_name": "ping",
  "interval_seconds": 60,
  "last_executed_at": null,
  "device": {
    "name": "本地服务器"
  }
}
```

**前端显示**:
```json
{
  "name": "本地服务器 - ping",
  "plugin_type": "ping",
  "interval": "60s",
  "last_execution": "-"
}
```

---

## 修改文件清单

### 前端文件

1. **`web/src/views/Tasks/List.vue`**
   - 添加 `formatSecondsToDuration` 函数
   - 添加 `formatDateTime` 函数
   - 修复 `fetchTasks` 方法，添加数据转换逻辑
   - 更新表格列定义，使用模板显示转换后的数据

---

## 测试验证

### 1. 任务列表显示

```bash
# 访问任务管理页面
# ✅ 任务名称应该显示为 "设备名称 - 插件类型"
# ✅ 插件类型应该正确显示
# ✅ 采集间隔应该显示为 "60s" 格式
# ✅ 最后执行应该显示格式化后的日期时间，或 "-"（如果未执行）
```

### 2. 数据转换验证

```bash
# 检查不同场景：
# 1. 有设备名称的任务：应该显示 "设备名称 - 插件类型"
# 2. 没有设备名称的任务：应该显示 "任务 task_id"
# 3. 采集间隔为 60 秒：应该显示 "60s"
# 4. 采集间隔为 120 秒：应该显示 "2m"
# 5. 最后执行时间为 null：应该显示 "-"
# 6. 最后执行时间有值：应该显示格式化后的日期时间
```

---

## 注意事项

1. **任务名称**: 如果没有设备名称，使用 `task_id` 作为后备
2. **日期格式化**: 使用中文本地化格式显示日期时间
3. **空值处理**: 所有空值都显示为 `-`
4. **数据转换**: 在 `fetchTasks` 中进行一次性转换，避免在模板中重复转换

---

## 总结

✅ **问题已修复**: 任务列表所有字段已正确显示  
✅ **数据转换**: 后端字段已正确映射到前端显示字段  
✅ **格式化**: 日期时间和时间间隔已正确格式化  
✅ **空值处理**: 空值已正确处理并显示为 `-`  

**修复完成时间**: 2025-11-14  
**修复状态**: ✅ 已完成

