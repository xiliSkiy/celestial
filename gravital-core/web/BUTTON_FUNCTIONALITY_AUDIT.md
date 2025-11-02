# 前端按钮功能实现检查报告

## 📋 检查概览

检查日期：2025-11-02  
检查范围：所有前端页面的按钮功能实现

## ✅ 已完全实现的功能

### 1. Dashboard（仪表盘）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 刷新按钮 | ✅ 已实现 | 各图表的刷新功能 |
| 卡片点击跳转 | ✅ 已实现 | 点击统计卡片跳转到对应页面 |

**实现细节**：
- 统计卡片可点击跳转（设备、告警）
- 图表刷新按钮调用对应的 fetch 函数
- 最近活动刷新功能

**待优化**：
- 刷新函数目前只是 `console.log`，需要调用真实 API

### 2. Devices/List（设备列表）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 添加设备 | ✅ 已实现 | 打开对话框，表单验证，API 调用 |
| 批量导入 | ⚠️ 占位符 | 显示"功能开发中"提示 |
| 导出 | ⚠️ 占位符 | 显示"功能开发中"提示 |
| 搜索 | ✅ 已实现 | 关键词搜索，调用 API |
| 状态筛选 | ✅ 已实现 | 按状态筛选设备 |
| 类型筛选 | ✅ 已实现 | 按设备类型筛选 |
| 详情 | ✅ 已实现 | 跳转到设备详情页 |
| 编辑 | ✅ 已实现 | 打开对话框，填充数据，API 调用 |
| 删除 | ✅ 已实现 | 确认对话框，API 调用 |
| 分页 | ✅ 已实现 | 页码和每页数量切换 |

**实现细节**：
```typescript
// 使用 deviceStore 管理状态
const fetchDevices = () => {
  deviceStore.fetchDevices(query)
}

// 表单验证
const rules: FormRules = {
  name: [{ required: true, message: '请输入设备名称', trigger: 'blur' }],
  device_type: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
  'connection_config.host': [{ required: true, message: '请输入IP地址', trigger: 'blur' }],
  'connection_config.port': [{ required: true, message: '请输入端口', trigger: 'blur' }]
}

// 删除确认
ElMessageBox.confirm('确定要删除该设备吗？', '提示', {
  confirmButtonText: '确定',
  cancelButtonText: '取消',
  type: 'warning'
}).then(async () => {
  await deviceStore.deleteDevice(row.id)
  fetchDevices()
})
```

### 3. Sentinels/List（Sentinel 列表）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 搜索 | ✅ 已实现 | 关键词搜索，API 调用 |
| 状态筛选 | ✅ 已实现 | 按状态筛选 |
| 详情 | ✅ 已实现 | 跳转到详情页 |
| 删除 | ✅ 已实现 | 确认对话框，API 调用 |
| 分页 | ✅ 已实现 | 页码和每页数量切换 |

**实现细节**：
```typescript
const fetchSentinels = async () => {
  try {
    const res: any = await sentinelApi.getSentinels(query)
    sentinels.value = res.items
    total.value = res.total
  } catch (error) {
    ElMessage.error('获取 Sentinel 列表失败')
  }
}
```

### 4. Tasks/List（任务列表）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 创建任务 | ❌ 未实现 | 只有 console.log |
| 搜索 | ❌ 未实现 | UI 存在但无功能 |
| 编辑 | ❌ 未实现 | 只有 console.log |
| 删除 | ❌ 未实现 | 只有 console.log |

**当前代码**：
```typescript
const handleCreate = () => {
  console.log('Create task')
}

const handleEdit = (row: any) => {
  console.log('Edit task', row)
}

const handleDelete = (row: any) {
  console.log('Delete task', row)
}
```

**需要实现**：
- 创建任务对话框和表单
- 调用任务 API
- 数据加载和刷新

### 5. Alerts/Index（告警管理）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 创建规则 | ❌ 未实现 | 只有 console.log |
| 搜索规则 | ❌ 未实现 | UI 存在但无功能 |
| 编辑规则 | ❌ 未实现 | 只有 console.log |
| 删除规则 | ❌ 未实现 | 只有 console.log |
| 级别筛选 | ❌ 未实现 | UI 存在但无功能 |
| 状态筛选 | ❌ 未实现 | UI 存在但无功能 |
| 确认告警 | ❌ 未实现 | 只有 console.log |
| 解决告警 | ❌ 未实现 | 只有 console.log |

**当前代码**：
```typescript
const handleCreateRule = () => {
  console.log('Create rule')
}

const handleEditRule = (row: any) => {
  console.log('Edit rule', row)
}

const handleDeleteRule = (row: any) => {
  console.log('Delete rule', row)
}

const handleAcknowledge = (event: any) => {
  console.log('Acknowledge', event)
}

const handleResolve = (event: any) => {
  console.log('Resolve', event)
}
```

### 6. Forwarders/List（数据转发）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 添加转发器 | ❌ 未实现 | 只有 console.log |
| 重新加载配置 | ❌ 未实现 | 只有 console.log |
| 详情 | ❌ 未实现 | 只有 console.log |
| 编辑 | ❌ 未实现 | 只有 console.log |
| 删除 | ❌ 未实现 | 只有 console.log |

**当前代码**：
```typescript
const handleCreate = () => {
  console.log('Create forwarder')
}

const handleReload = () => {
  console.log('Reload configuration')
}

const handleView = (forwarder: any) => {
  console.log('View forwarder', forwarder)
}

const handleEdit = (forwarder: any) => {
  console.log('Edit forwarder', forwarder)
}

const handleDelete = (forwarder: any) => {
  console.log('Delete forwarder', forwarder)
}
```

### 7. Settings/Index（系统设置）
| 按钮/功能 | 实现状态 | 说明 |
|----------|---------|------|
| 保存设置 | ❌ 未实现 | 只有 console.log |
| 添加用户 | ❌ 未实现 | 只有 console.log |
| 编辑用户 | ❌ 未实现 | 只有 console.log |
| 删除用户 | ❌ 未实现 | 只有 console.log |

## 📊 统计总结

### 实现状态统计

| 页面 | 已实现 | 部分实现 | 未实现 | 完成度 |
|------|-------|---------|--------|--------|
| Dashboard | 3 | 0 | 0 | 100% |
| Devices/List | 8 | 2 | 0 | 80% |
| Sentinels/List | 5 | 0 | 0 | 100% |
| Tasks/List | 0 | 0 | 4 | 0% |
| Alerts/Index | 0 | 0 | 8 | 0% |
| Forwarders/List | 0 | 0 | 5 | 0% |
| Settings/Index | 0 | 0 | 4 | 0% |
| **总计** | **16** | **2** | **21** | **41%** |

### 功能分类统计

| 功能类型 | 已实现 | 未实现 |
|---------|-------|--------|
| CRUD 操作 | 6 | 15 |
| 搜索/筛选 | 6 | 4 |
| 数据展示 | 4 | 0 |
| 导航跳转 | 3 | 0 |

## 🔧 需要实现的优先级

### P0 - 高优先级（核心功能）

1. **任务管理** - 完全未实现
   - 创建任务（表单 + API）
   - 编辑任务
   - 删除任务
   - 数据加载

2. **告警管理** - 完全未实现
   - 创建规则（表单 + API）
   - 编辑规则
   - 删除规则
   - 确认/解决告警事件
   - 数据加载和筛选

3. **数据转发** - 完全未实现
   - 添加转发器（表单 + API）
   - 编辑转发器
   - 删除转发器
   - 重新加载配置
   - 数据加载

### P1 - 中优先级（增强功能）

4. **Dashboard 数据加载**
   - 替换 mock 数据为真实 API 调用
   - 实现图表刷新逻辑

5. **设备批量操作**
   - 批量导入功能
   - 导出功能

6. **系统设置**
   - 保存设置
   - 用户管理 CRUD

### P2 - 低优先级（优化功能）

7. **实时更新**
   - WebSocket 连接
   - 实时数据推送

8. **高级搜索**
   - 多条件组合搜索
   - 保存搜索条件

## 🎯 实现建议

### 1. 任务管理实现示例

```typescript
// src/api/task.ts
export const taskApi = {
  getTasks: (params: any) => request.get('/api/v1/tasks', { params }),
  getTask: (id: number) => request.get(`/api/v1/tasks/${id}`),
  createTask: (data: any) => request.post('/api/v1/tasks', data),
  updateTask: (id: number, data: any) => request.put(`/api/v1/tasks/${id}`, data),
  deleteTask: (id: number) => request.delete(`/api/v1/tasks/${id}`)
}

// src/views/Tasks/List.vue
const handleCreate = () => {
  dialogTitle.value = '创建任务'
  currentTask.value = null
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (currentTask.value) {
          await taskApi.updateTask(currentTask.value.id, form)
          ElMessage.success('更新成功')
        } else {
          await taskApi.createTask(form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchTasks()
      } catch (error) {
        ElMessage.error('操作失败')
      }
    }
  })
}
```

### 2. 告警管理实现示例

```typescript
// src/api/alert.ts
export const alertApi = {
  getRules: (params: any) => request.get('/api/v1/alert-rules', { params }),
  createRule: (data: any) => request.post('/api/v1/alert-rules', data),
  updateRule: (id: number, data: any) => request.put(`/api/v1/alert-rules/${id}`, data),
  deleteRule: (id: number) => request.delete(`/api/v1/alert-rules/${id}`),
  
  getEvents: (params: any) => request.get('/api/v1/alert-events', { params }),
  acknowledgeEvent: (id: number) => request.post(`/api/v1/alert-events/${id}/acknowledge`),
  resolveEvent: (id: number) => request.post(`/api/v1/alert-events/${id}/resolve`)
}
```

### 3. 数据转发实现示例

```typescript
// src/api/forwarder.ts
export const forwarderApi = {
  getForwarders: () => request.get('/api/v1/forwarders'),
  createForwarder: (data: any) => request.post('/api/v1/forwarders', data),
  updateForwarder: (id: number, data: any) => request.put(`/api/v1/forwarders/${id}`, data),
  deleteForwarder: (id: number) => request.delete(`/api/v1/forwarders/${id}`),
  reloadConfig: () => request.post('/api/v1/forwarders/reload')
}
```

## 📝 实现步骤

### 阶段 1：核心 CRUD（1-2天）
1. 创建 API 文件（task.ts, alert.ts, forwarder.ts）
2. 实现任务管理 CRUD
3. 实现告警管理 CRUD
4. 实现数据转发 CRUD

### 阶段 2：数据加载和展示（1天）
1. 替换 mock 数据
2. 实现数据加载逻辑
3. 添加 loading 状态
4. 错误处理

### 阶段 3：高级功能（1-2天）
1. 批量操作
2. 高级搜索
3. 数据导出
4. 系统设置

### 阶段 4：优化和测试（1天）
1. 性能优化
2. 用户体验优化
3. 全面测试
4. Bug 修复

## 🚀 快速实现指南

### 使用代码生成器

可以基于 `Devices/List.vue` 作为模板，快速生成其他页面：

1. 复制设备管理的实现模式
2. 修改 API 调用
3. 调整表单字段
4. 更新数据模型

### 统一的实现模式

所有 CRUD 页面应遵循相同的模式：
- 工具栏（操作按钮 + 搜索筛选）
- 数据表格/卡片
- 分页组件
- 创建/编辑对话框
- 删除确认对话框

## 📚 相关文件

- `src/api/` - API 调用定义
- `src/views/` - 页面组件
- `src/stores/` - 状态管理
- `src/types/` - TypeScript 类型定义

---

**报告生成日期**: 2025-11-02  
**检查人员**: AI Assistant  
**下一步**: 根据优先级逐步实现未完成的功能

