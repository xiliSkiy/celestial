# 转发器测试连接功能修复说明

## 问题描述

前端添加转发器后，点击"测试连接"按钮时，请求接口 `http://localhost:5173/api/v1/forwarders/[object%20Object]/test` 报 404 错误。

**问题原因**:
1. 前端调用 `forwarderApi.testConnection(form)` 传递的是整个 `form` 对象
2. 前端 API 定义 `testConnection: (id: number)` 期望的是 `id`，导致 URL 中出现 `[object Object]`
3. 后端没有实现测试连接的接口

---

## 修复方案

### 1. 后端实现测试连接接口

#### 1.1 服务接口定义

**文件**: `internal/service/forwarder_service.go`

添加 `TestConnection` 方法到 `ForwarderService` 接口：

```go
// ForwarderService 转发器服务接口
type ForwarderService interface {
    // ...
    // 测试连接
    TestConnection(ctx context.Context, config *model.ForwarderConfig) (*ForwarderTestConnectionResult, error)
    // ...
}

// ForwarderTestConnectionResult 转发器测试连接结果
type ForwarderTestConnectionResult struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Latency int64  `json:"latency_ms,omitempty"`
}
```

#### 1.2 服务实现

实现 `TestConnection` 方法：

```go
func (s *forwarderService) TestConnection(ctx context.Context, config *model.ForwarderConfig) (*ForwarderTestConnectionResult, error) {
    startTime := time.Now()

    // 转换为转发器配置
    fwdConfig := s.modelToForwarderConfig(config)

    // 创建临时转发器实例进行测试
    var testForwarder forwarder.Forwarder
    var err error

    switch fwdConfig.Type {
    case forwarder.ForwarderTypePrometheus:
        testForwarder, err = forwarder.NewPrometheusForwarder(fwdConfig, s.logger)
    case forwarder.ForwarderTypeVictoriaMetrics:
        testForwarder, err = forwarder.NewVictoriaMetricsForwarder(fwdConfig, s.logger)
    case forwarder.ForwarderTypeClickHouse:
        testForwarder, err = forwarder.NewClickHouseForwarder(fwdConfig, s.logger)
    default:
        return &ForwarderTestConnectionResult{
            Success: false,
            Message: fmt.Sprintf("不支持的转发器类型: %s", config.Type),
        }, nil
    }

    if err != nil {
        return &ForwarderTestConnectionResult{
            Success: false,
            Message: fmt.Sprintf("创建转发器失败: %v", err),
        }, nil
    }
    defer testForwarder.Close()

    // 创建测试指标
    testMetrics := []*forwarder.Metric{
        {
            Name:      "test_connection",
            Value:     1.0,
            Type:      "gauge",
            Labels:    map[string]string{"test": "true"},
            Timestamp: time.Now().Unix(),
        },
    }

    // 尝试写入测试数据
    if err := testForwarder.Write(testMetrics); err != nil {
        latency := time.Since(startTime).Milliseconds()
        return &ForwarderTestConnectionResult{
            Success: false,
            Message: fmt.Sprintf("连接测试失败: %v", err),
            Latency: latency,
        }, nil
    }

    latency := time.Since(startTime).Milliseconds()
    return &ForwarderTestConnectionResult{
        Success: true,
        Message: "连接测试成功",
        Latency: latency,
    }, nil
}
```

**实现逻辑**:
1. 根据配置创建临时转发器实例
2. 创建测试指标数据
3. 尝试写入测试数据到目标系统
4. 返回测试结果（成功/失败、消息、延迟）

#### 1.3 Handler 实现

**文件**: `internal/api/handler/forwarder_handler.go`

添加 `TestConnection` handler：

```go
// TestConnection 测试转发器连接
// @Summary 测试转发器连接
// @Tags forwarder
// @Accept json
// @Produce json
// @Param config body model.ForwarderConfig true "转发器配置"
// @Success 200 {object} Response{data=service.ForwarderTestConnectionResult}
// @Router /api/v1/forwarders/test [post]
func (h *ForwarderHandler) TestConnection(c *gin.Context) {
    var req model.ForwarderConfig
    if err := c.ShouldBindJSON(&req); err != nil {
        ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
        return
    }

    result, err := h.service.TestConnection(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to test connection", zap.Error(err))
        ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
        return
    }

    SuccessResponse(c, result)
}
```

#### 1.4 路由注册

**文件**: `internal/api/router/router.go`

添加测试连接路由：

```go
forwarders.POST("/test", middleware.RequirePermission("admin.config"), forwarderHandler.TestConnection)
```

**路由**: `POST /api/v1/forwarders/test`

---

### 2. 前端修复

#### 2.1 API 定义修复

**文件**: `web/src/api/forwarder.ts`

修改 `testConnection` 方法，接受配置对象而不是 `id`：

```typescript
// 修复前
testConnection: (id: number) => {
  return request.post(`/v1/forwarders/${id}/test`)
}

// 修复后
testConnection: (data: ForwarderForm) => {
  return request.post('/v1/forwarders/test', data)
}
```

#### 2.2 前端调用修复

**文件**: `web/src/views/Forwarders/List.vue`

添加时间字符串解析函数，并修复 `handleTest` 方法：

```typescript
// 解析时间字符串为秒数（如 "10s" -> 10, "1m" -> 60）
const parseDurationToSeconds = (duration: string): number => {
  if (!duration) return 10 // 默认值
  const match = duration.match(/^(\d+)([smh])?$/)
  if (!match) return 10
  const value = parseInt(match[1])
  const unit = match[2] || 's'
  switch (unit) {
    case 's': return value
    case 'm': return value * 60
    case 'h': return value * 3600
    default: return value
  }
}

// 测试连接
const handleTest = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 解析配置
      try {
        form.auth_config = JSON.parse(authConfigStr.value)
        form.tls_config = JSON.parse(tlsConfigStr.value)
      } catch (error) {
        ElMessage.error('配置 JSON 格式错误')
        return
      }
      
      testing.value = true
      try {
        // 准备测试数据（转换格式以匹配后端）
        const testData = {
          name: form.name || 'test',
          type: form.type,
          endpoint: form.endpoint,
          enabled: form.enabled,
          batch_size: form.batch_size || 1000,
          flush_interval: parseDurationToSeconds(form.flush_interval || '10s'),
          retry_times: 3,
          timeout_seconds: 30,
          auth_config: form.auth_config || {}
        }
        
        const result = await forwarderApi.testConnection(testData as any)
        if (result.success) {
          ElMessage.success(result.message || '连接测试成功')
        } else {
          ElMessage.error(result.message || '连接测试失败')
        }
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || error.message || '连接测试失败')
      } finally {
        testing.value = false
      }
    }
  })
}
```

**关键修复**:
1. 添加 `parseDurationToSeconds` 函数，将前端的时间字符串（如 "10s"）转换为后端期望的秒数
2. 准备测试数据时，转换数据格式以匹配后端 `model.ForwarderConfig` 结构
3. 处理响应结果，根据 `success` 字段显示成功或失败消息

---

## 修复后的效果

### 请求流程

1. **前端调用**:
   ```typescript
   await forwarderApi.testConnection(testData)
   ```

2. **API 请求**:
   ```
   POST /api/v1/forwarders/test
   Content-Type: application/json
   
   {
     "name": "test",
     "type": "victoria-metrics",
     "endpoint": "http://localhost:8428/api/v1/write",
     "enabled": true,
     "batch_size": 1000,
     "flush_interval": 10,
     "retry_times": 3,
     "timeout_seconds": 30,
     "auth_config": {}
   }
   ```

3. **后端处理**:
   - 创建临时转发器实例
   - 尝试写入测试指标
   - 返回测试结果

4. **响应格式**:
   ```json
   {
     "code": 0,
     "data": {
       "success": true,
       "message": "连接测试成功",
       "latency_ms": 45
     }
   }
   ```

### 成功场景

- ✅ 前端正确发送配置数据
- ✅ 后端创建临时转发器实例
- ✅ 成功写入测试数据
- ✅ 返回成功结果和延迟时间
- ✅ 前端显示成功消息

### 失败场景

- ✅ 配置错误时返回详细错误信息
- ✅ 连接失败时返回失败原因
- ✅ 前端显示错误消息

---

## 测试验证

### 1. 测试 Prometheus 连接

```bash
# 启动 Prometheus
docker-compose up -d prometheus

# 在前端添加转发器
# 类型: Prometheus
# 端点: http://localhost:9090/api/v1/write
# 点击"测试连接"
```

### 2. 测试 VictoriaMetrics 连接

```bash
# 启动 VictoriaMetrics
docker-compose up -d victoriametrics

# 在前端添加转发器
# 类型: VictoriaMetrics
# 端点: http://localhost:8428/api/v1/write
# 点击"测试连接"
```

### 3. 测试 ClickHouse 连接

```bash
# 启动 ClickHouse
docker-compose up -d clickhouse

# 在前端添加转发器
# 类型: ClickHouse
# 端点: tcp://localhost:9000/default
# 点击"测试连接"
```

### 4. 测试错误场景

- 无效的端点地址
- 错误的认证信息
- 不支持的类型

---

## 修改文件清单

### 后端文件

1. `internal/service/forwarder_service.go`
   - 添加 `TestConnection` 方法到接口
   - 实现 `TestConnection` 方法
   - 定义 `ForwarderTestConnectionResult` 结构体

2. `internal/api/handler/forwarder_handler.go`
   - 添加 `TestConnection` handler 方法

3. `internal/api/router/router.go`
   - 注册测试连接路由 `POST /api/v1/forwarders/test`

### 前端文件

1. `web/src/api/forwarder.ts`
   - 修改 `testConnection` 方法签名，接受配置对象

2. `web/src/views/Forwarders/List.vue`
   - 添加 `parseDurationToSeconds` 函数
   - 修复 `handleTest` 方法，正确准备和发送测试数据

---

## 注意事项

1. **权限要求**: 测试连接接口需要 `admin.config` 权限
2. **数据格式**: 前端需要将时间字符串转换为秒数
3. **临时实例**: 测试时创建临时转发器实例，测试完成后自动关闭
4. **错误处理**: 所有错误都会返回详细的错误信息

---

## 总结

✅ **问题已修复**: 转发器测试连接功能已完整实现  
✅ **后端实现**: 添加了测试连接接口和服务方法  
✅ **前端修复**: 修复了 API 调用和数据格式转换  
✅ **功能验证**: 支持 Prometheus、VictoriaMetrics、ClickHouse 三种类型  

**修复完成时间**: 2025-11-14  
**修复状态**: ✅ 已完成

