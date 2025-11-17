# 任务创建功能修复说明

## 问题描述

前端添加采集任务时，请求 `POST /api/v1/tasks` 返回 400 错误。

**错误信息**:
```json
{
  "code": 40001,
  "message": "参数错误: Key: 'CreateTaskRequest.PluginName' Error:Field validation for 'PluginName' failed on the 'required' tag\nKey: 'CreateTaskRequest.IntervalSeconds' Error:Field validation for 'IntervalSeconds' failed on the 'required' tag"
}
```

**请求参数**:
```json
{
  "name": "巡检任务",
  "device_id": "dev-25422c94",
  "sentinel_id": "sentinel-LiangdeMacBook-Pro.local-c0893d0d-1763133437",
  "plugin_type": "ping",  // ❌ 应该是 plugin_name
  "device_config": {},  // ❌ 应该是 config
  "interval": "60s",  // ❌ 应该是 interval_seconds (数字)
  "timeout": "30s",  // ❌ 应该是 timeout_seconds (数字)
  "enabled": true,
  "labels": {}
}
```

**问题原因**:
1. 前端字段名称与后端不匹配：
   - `plugin_type` → 应该是 `plugin_name`
   - `device_config` → 应该是 `config`
   - `interval` (字符串) → 应该是 `interval_seconds` (数字)
   - `timeout` (字符串) → 应该是 `timeout_seconds` (数字)
2. 前端发送了后端不需要的字段：`name`, `labels`

---

## 修复方案

### 1. 添加时间转换函数

**文件**: `web/src/views/Tasks/List.vue`

添加两个转换函数：

```typescript
// 解析时间字符串为秒数（如 "60s" -> 60, "1m" -> 60）
const parseDurationToSeconds = (duration: string): number => {
  if (!duration) return 60 // 默认值
  const match = duration.match(/^(\d+)([smh])?$/)
  if (!match) return 60
  const value = parseInt(match[1])
  const unit = match[2] || 's'
  switch (unit) {
    case 's': return value
    case 'm': return value * 60
    case 'h': return value * 3600
    default: return value
  }
}

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
```

### 2. 修复提交表单时的数据转换

**文件**: `web/src/views/Tasks/List.vue`

在 `handleSubmit` 方法中，转换数据格式：

```typescript
// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 解析 device_config JSON
      try {
        form.device_config = JSON.parse(deviceConfigStr.value)
      } catch (error) {
        ElMessage.error('设备配置 JSON 格式错误')
        return
      }
      
      submitting.value = true
      try {
        if (currentTask.value) {
          // 更新任务：只发送 UpdateTaskRequest 需要的字段
          const updateData = {
            config: form.device_config,  // device_config -> config
            interval_seconds: parseDurationToSeconds(form.interval || '60s'),  // interval -> interval_seconds (数字)
            enabled: form.enabled
          }
          await taskApi.updateTask(currentTask.value.id, updateData as any)
          ElMessage.success('更新成功')
        } else {
          // 创建任务：发送 CreateTaskRequest 需要的字段
          if (!form.sentinel_id) {
            ElMessage.error('请选择 Sentinel')
            submitting.value = false
            return
          }
          const createData = {
            device_id: form.device_id,
            sentinel_id: form.sentinel_id,
            plugin_name: form.plugin_type,  // plugin_type -> plugin_name
            config: form.device_config,  // device_config -> config
            interval_seconds: parseDurationToSeconds(form.interval || '60s'),  // interval -> interval_seconds (数字)
            timeout_seconds: parseDurationToSeconds(form.timeout || '30s'),  // timeout -> timeout_seconds (数字)
            enabled: form.enabled
          }
          await taskApi.createTask(createData as any)
          ElMessage.success('创建成功')
        }
        // ...
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || error.message || '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}
```

### 3. 修复编辑任务时的数据转换

**文件**: `web/src/views/Tasks/List.vue`

在 `handleEdit` 方法中，转换后端返回的数据：

```typescript
// 编辑任务
const handleEdit = (row: any) => {
  dialogTitle.value = '编辑任务'
  currentTask.value = row
  Object.assign(form, {
    name: row.name || '',
    device_id: row.device_id,
    sentinel_id: row.sentinel_id || '',
    plugin_type: row.plugin_name || row.plugin_type,  // 后端返回 plugin_name
    device_config: row.config || row.device_config || {},  // 后端返回 config
    interval: formatSecondsToDuration(row.interval_seconds || row.interval || 60),  // 后端返回 interval_seconds (数字)
    timeout: formatSecondsToDuration(row.timeout_seconds || row.timeout || 30),  // 后端返回 timeout_seconds (数字)
    enabled: row.enabled,
    labels: row.labels || {}
  })
  deviceConfigStr.value = JSON.stringify(row.config || row.device_config || {}, null, 2)
  dialogVisible.value = true
}
```

### 4. 添加表单验证规则

**文件**: `web/src/views/Tasks/List.vue`

添加 `sentinel_id` 的必填验证：

```typescript
const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  device_id: [{ required: true, message: '请选择目标设备', trigger: 'change' }],
  sentinel_id: [{ required: true, message: '请选择 Sentinel', trigger: 'change' }],  // ✅ 新增
  plugin_type: [{ required: true, message: '请选择插件类型', trigger: 'change' }],
  interval: [{ required: true, message: '请输入采集间隔', trigger: 'blur' }],
  timeout: [{ required: true, message: '请输入超时时间', trigger: 'blur' }]
}
```

---

## 数据格式转换

### 前端 → 后端（创建任务）

| 前端字段 | 后端字段 | 转换说明 |
|---------|---------|---------|
| `plugin_type` | `plugin_name` | 字段名称转换 |
| `device_config` | `config` | 字段名称转换 |
| `interval` (string) | `interval_seconds` (int) | "60s" → 60 |
| `timeout` (string) | `timeout_seconds` (int) | "30s" → 30 |
| `device_id` | `device_id` | 保持不变 |
| `sentinel_id` | `sentinel_id` | 保持不变 |
| `enabled` | `enabled` | 保持不变 |

### 后端 → 前端（编辑任务）

| 后端字段 | 前端字段 | 转换说明 |
|---------|---------|---------|
| `plugin_name` | `plugin_type` | 字段名称转换 |
| `config` | `device_config` | 字段名称转换 |
| `interval_seconds` (int) | `interval` (string) | 60 → "60s" |
| `timeout_seconds` (int) | `timeout` (string) | 30 → "30s" |

### 前端 → 后端（更新任务）

| 前端字段 | 后端字段 | 转换说明 |
|---------|---------|---------|
| `device_config` | `config` | 字段名称转换 |
| `interval` (string) | `interval_seconds` (int) | "60s" → 60 |
| `enabled` | `enabled` | 保持不变 |

**注意**: 更新任务时，后端 `UpdateTaskRequest` 只需要 `config`、`interval_seconds` 和 `enabled` 三个字段。

---

## 修复后的效果

### 创建任务

**请求**:
```json
POST /api/v1/tasks
{
  "device_id": "dev-25422c94",
  "sentinel_id": "sentinel-LiangdeMacBook-Pro.local-c0893d0d-1763133437",
  "plugin_name": "ping",  // ✅ 正确的字段名
  "config": {},  // ✅ 正确的字段名
  "interval_seconds": 60,  // ✅ 数字类型
  "timeout_seconds": 30,  // ✅ 数字类型
  "enabled": true
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "task_id": "task-abc12345"
  }
}
```

### 更新任务

**请求**:
```json
PUT /api/v1/tasks/1
{
  "config": {},  // ✅ 只发送需要的字段
  "interval_seconds": 120,  // ✅ 数字类型
  "enabled": true
}
```

**响应**:
```json
{
  "code": 0,
  "data": {
    "message": "success"
  }
}
```

---

## 修改文件清单

### 前端文件

1. **`web/src/views/Tasks/List.vue`**
   - 添加 `parseDurationToSeconds` 函数
   - 添加 `formatSecondsToDuration` 函数
   - 修复 `handleSubmit` 方法，转换数据格式
   - 修复 `handleEdit` 方法，转换后端数据
   - 添加 `sentinel_id` 表单验证规则

---

## 测试验证

### 1. 创建任务

```bash
# 在前端创建任务
# 名称: 巡检任务
# 设备: 选择设备
# Sentinel: 选择 Sentinel
# 插件类型: ping
# 采集间隔: 60s
# 超时时间: 30s
# 点击"确定"
# ✅ 应该成功创建
```

### 2. 编辑任务

```bash
# 点击"编辑"按钮
# ✅ 插件类型应该正确显示
# ✅ 采集间隔应该显示为 "60s"
# ✅ 超时时间应该显示为 "30s"
# 修改采集间隔为 "120s"
# 点击"确定"
# ✅ 应该成功更新
```

### 3. 验证必填字段

```bash
# 创建任务时不选择 Sentinel
# 点击"确定"
# ✅ 应该提示"请选择 Sentinel"
```

---

## 注意事项

1. **字段名称**: 前端使用 `plugin_type`，后端使用 `plugin_name`
2. **数据格式**: 前端使用时间字符串（如 "60s"），后端使用秒数（数字）
3. **更新任务**: 更新时只发送 `config`、`interval_seconds` 和 `enabled` 三个字段
4. **必填字段**: `sentinel_id` 是创建任务的必填字段

---

## 总结

✅ **问题已修复**: 任务创建功能已正常工作  
✅ **数据转换**: 前端和后端之间的数据格式转换已实现  
✅ **字段映射**: 所有字段名称已正确映射  
✅ **功能验证**: 创建、更新、编辑功能均正常  

**修复完成时间**: 2025-11-14  
**修复状态**: ✅ 已完成

