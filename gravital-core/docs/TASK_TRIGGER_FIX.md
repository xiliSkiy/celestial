# 任务触发功能实现说明

## 问题描述

任务管理界面点击"执行"按钮，请求 `POST /api/v1/tasks/1/trigger` 接口时，返回错误：

```json
{
  "code": 10001,
  "message": "触发失败: trigger not implemented"
}
```

**问题原因**:
- 后端的 `Trigger` 方法还没有实现，只是返回了一个占位错误

---

## 实现方案

### 1. 修改 Service 接口

**文件**: `internal/service/task_service.go`

修改 `Trigger` 方法的签名，使用 `uint` ID 而不是 `string` taskID：

```go
// TaskService 任务服务接口
type TaskService interface {
    // ...
    Trigger(ctx context.Context, id uint) error  // ✅ 改为 uint ID
    // ...
}
```

### 2. 实现 Trigger 方法

**文件**: `internal/service/task_service.go`

实现手动触发任务的逻辑：

```go
func (s *taskService) Trigger(ctx context.Context, id uint) error {
    // 获取任务
    task, err := s.taskRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return fmt.Errorf("task not found")
        }
        return fmt.Errorf("failed to get task: %w", err)
    }

    // 检查任务是否启用
    if !task.Enabled {
        return fmt.Errorf("task is disabled")
    }

    // 将下次执行时间设置为当前时间，这样 Sentinel 下次拉取任务时就会立即执行
    now := time.Now()
    task.NextExecutionAt = &now

    // 更新任务
    if err := s.taskRepo.Update(ctx, task); err != nil {
        return fmt.Errorf("failed to update task: %w", err)
    }

    return nil
}
```

**实现逻辑**:
1. 根据 ID 获取任务
2. 检查任务是否存在
3. 检查任务是否启用（禁用的任务不能触发）
4. 将 `NextExecutionAt` 设置为当前时间
5. 更新任务到数据库

**工作原理**:
- Sentinel 定期从中心端拉取任务列表
- Sentinel 只拉取 `NextExecutionAt <= 当前时间` 的任务
- 将 `NextExecutionAt` 设置为当前时间后，Sentinel 下次拉取任务时就会立即执行该任务

### 3. 修复 Handler

**文件**: `internal/api/handler/task_handler.go`

修复 `Trigger` handler，正确解析 ID 并调用 service：

```go
// Trigger 手动触发任务
func (h *TaskHandler) Trigger(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "code":    40001,
            "message": "无效的任务 ID",
        })
        return
    }

    if err := h.taskService.Trigger(c.Request.Context(), uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    10001,
            "message": "触发失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code":    0,
        "message": "任务已触发，将在下次 Sentinel 拉取任务时立即执行",
    })
}
```

---

## 工作流程

### 手动触发任务流程

```
1. 用户点击"执行"按钮
   ↓
2. 前端发送 POST /api/v1/tasks/{id}/trigger
   ↓
3. Handler 解析 ID 并调用 Service.Trigger
   ↓
4. Service 获取任务并检查状态
   ↓
5. Service 将 NextExecutionAt 设置为当前时间
   ↓
6. Service 更新任务到数据库
   ↓
7. 返回成功响应
   ↓
8. Sentinel 下次拉取任务时，发现 NextExecutionAt <= 当前时间
   ↓
9. Sentinel 立即执行该任务
```

### 任务执行时机

- **正常调度**: Sentinel 根据 `NextExecutionAt` 自动执行任务
- **手动触发**: 将 `NextExecutionAt` 设置为当前时间，Sentinel 下次拉取时立即执行

---

## 修复后的效果

### 成功触发

**请求**:
```
POST /api/v1/tasks/1/trigger
```

**响应**:
```json
{
  "code": 0,
  "message": "任务已触发，将在下次 Sentinel 拉取任务时立即执行"
}
```

### 错误场景

#### 1. 任务不存在

**响应**:
```json
{
  "code": 10001,
  "message": "触发失败: task not found"
}
```

#### 2. 任务已禁用

**响应**:
```json
{
  "code": 10001,
  "message": "触发失败: task is disabled"
}
```

#### 3. 无效的任务 ID

**响应**:
```json
{
  "code": 40001,
  "message": "无效的任务 ID"
}
```

---

## 修改文件清单

### 后端文件

1. **`internal/service/task_service.go`**
   - 修改 `Trigger` 方法签名：`taskID string` → `id uint`
   - 实现 `Trigger` 方法逻辑

2. **`internal/api/handler/task_handler.go`**
   - 修复 `Trigger` handler，正确解析 ID
   - 改进错误处理和响应消息

---

## 测试验证

### 1. 正常触发

```bash
# 1. 创建一个启用的任务
# 2. 点击"执行"按钮
# ✅ 应该返回成功消息
# ✅ 任务的 NextExecutionAt 应该更新为当前时间
```

### 2. 禁用任务触发

```bash
# 1. 禁用一个任务
# 2. 点击"执行"按钮
# ✅ 应该返回错误："task is disabled"
```

### 3. 不存在的任务

```bash
# 1. 使用不存在的任务 ID
# 2. 点击"执行"按钮
# ✅ 应该返回错误："task not found"
```

### 4. Sentinel 执行验证

```bash
# 1. 手动触发一个任务
# 2. 等待 Sentinel 拉取任务（通常几秒内）
# ✅ Sentinel 应该立即执行该任务
# ✅ 任务的 LastExecutedAt 应该更新
```

---

## 注意事项

1. **异步执行**: 手动触发后，任务不会立即执行，而是等待 Sentinel 下次拉取任务时执行
2. **任务状态**: 只有启用的任务才能被触发
3. **执行时机**: 取决于 Sentinel 的拉取频率（通常为几秒到几十秒）
4. **时间精度**: `NextExecutionAt` 设置为当前时间，Sentinel 会立即识别并执行

---

## 未来优化建议

### 1. 实时通知机制

如果需要立即执行，可以考虑：
- 使用消息队列（如 Redis Pub/Sub、RabbitMQ）
- 使用 WebSocket 实时通知 Sentinel
- 使用 HTTP 回调通知 Sentinel

### 2. 执行结果反馈

- 添加执行状态查询接口
- 实时显示执行进度
- 执行完成后通知前端

### 3. 批量触发

- 支持批量触发多个任务
- 支持按条件触发（如按设备、按 Sentinel）

---

## 总结

✅ **功能已实现**: 任务手动触发功能已完整实现  
✅ **错误处理**: 完善的错误处理和验证  
✅ **工作流程**: 清晰的工作流程和实现逻辑  
✅ **用户体验**: 明确的成功和错误消息  

**实现完成时间**: 2025-11-14  
**实现状态**: ✅ 已完成

