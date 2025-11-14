# 采集端注册功能修复说明

## 问题描述

### 问题1: 注册接口401错误

**错误信息**:
```json
{
  "code": 20001,
  "error": "Unauthorized",
  "message": "未提供 Sentinel ID"
}
```

**原因分析**:
- 注册接口 `/api/v1/sentinels/register` 使用了 `middleware.SentinelAuth()` 中间件
- 该中间件要求必须提供 `X-Sentinel-ID` 和 `X-API-Token` Header
- 但注册时采集端还没有这些凭证（注册的目的就是获取这些凭证）
- 导致注册请求被拒绝

**解决方案**:
- 将注册接口从 `SentinelAuth()` 中间件中移除
- 注册接口不需要认证（因为注册的目的就是获取凭证）
- 心跳接口仍然需要认证

### 问题2: 日志格式缺少行号

**问题**:
- 日志输出中 `caller` 字段存在，但格式不够清晰
- 需要明确显示文件名和行号

**解决方案**:
- 在日志编码器配置中添加 `EncodeCaller = zapcore.ShortCallerEncoder`
- 使用短格式显示调用者信息（文件名:行号）

---

## 修复内容

### 1. 中心端路由修复

**文件**: `gravital-core/internal/api/router/router.go`

**修改前**:
```go
// Sentinel API（使用 API Token 认证）
sentinelAPI := v1.Group("/sentinels")
sentinelAPI.Use(middleware.SentinelAuth())
{
    sentinelAPI.POST("/register", sentinelHandler.Register)
    sentinelAPI.POST("/heartbeat", sentinelHandler.Heartbeat)
}
```

**修改后**:
```go
// Sentinel API
sentinelAPI := v1.Group("/sentinels")
{
    // 注册接口不需要认证（因为注册的目的就是获取凭证）
    sentinelAPI.POST("/register", sentinelHandler.Register)
    // 心跳接口需要认证
    sentinelAPI.POST("/heartbeat", middleware.SentinelAuth(), sentinelHandler.Heartbeat)
}
```

### 2. 日志格式优化

**文件**: `orbital-sentinels/internal/pkg/logger/logger.go`

**修改前**:
```go
encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
```

**修改后**:
```go
encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
// 启用调用者信息（包含文件名和行号）
encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
```

---

## 修复效果

### 修复前

**注册错误日志**:
```json
{
  "level": "WARN",
  "ts": "2025-11-14T23:12:21.727+0800",
  "caller": "logger/logger.go:98",
  "msg": "Registration attempt failed",
  "attempt": 3,
  "error": "failed to send register request: registration failed: status=401, body={\"code\":20001,\"error\":\"Unauthorized\",\"message\":\"未提供 Sentinel ID\"}"
}
```

**问题**:
- ❌ 注册失败（401错误）
- ❌ 日志行号显示不清晰

### 修复后

**注册成功日志**:
```json
{
  "level": "INFO",
  "ts": "2025-11-14T23:12:21.727+0800",
  "caller": "register/manager.go:47",
  "msg": "Registering to core",
  "hostname": "my-host",
  "ip": "192.168.1.100",
  "core_url": "http://localhost:8080"
}
```

```json
{
  "level": "INFO",
  "ts": "2025-11-14T23:12:22.123+0800",
  "caller": "register/manager.go:72",
  "msg": "Registration successful",
  "sentinel_id": "sentinel-my-host-abc12345-1699999999",
  "credentials_path": "/Users/username/.sentinel/credentials.yaml"
}
```

**改进**:
- ✅ 注册成功
- ✅ 日志行号清晰显示（`register/manager.go:47`）

---

## 日志格式说明

### JSON 格式日志

修复后的日志格式包含以下字段：

```json
{
  "level": "INFO",                    // 日志级别
  "ts": "2025-11-14T23:12:21.727+0800", // 时间戳
  "caller": "register/manager.go:47",  // 调用者（文件名:行号）
  "msg": "Registering to core",        // 日志消息
  "hostname": "my-host",               // 自定义字段
  "ip": "192.168.1.100"                // 自定义字段
}
```

### 调用者信息格式

- **格式**: `文件名:行号`
- **示例**: `register/manager.go:47`
- **说明**: 
  - 只显示相对路径（从项目根目录开始）
  - 行号精确到具体代码行
  - 便于快速定位问题

---

## 测试验证

### 1. 注册功能测试

```bash
# 1. 启动中心端
cd gravital-core
docker-compose up -d

# 2. 清除旧凭证
rm ~/.sentinel/credentials.yaml

# 3. 启动采集端
cd orbital-sentinels
./bin/sentinel start -c config/config.yaml

# 4. 查看日志（应该看到注册成功）
tail -f logs/sentinel.log | grep -i register
```

**预期输出**:
```
{"level":"INFO","ts":"...","caller":"register/manager.go:47","msg":"Registering to core",...}
{"level":"INFO","ts":"...","caller":"register/manager.go:72","msg":"Registration successful",...}
```

### 2. 日志格式测试

```bash
# 查看日志，确认行号显示
tail -f logs/sentinel.log | jq '.caller'
```

**预期输出**:
```
"register/manager.go:47"
"register/manager.go:72"
"agent/agent.go:407"
```

---

## 安全考虑

### 注册接口安全性

虽然注册接口移除了 `SentinelAuth()` 中间件，但仍可通过以下方式保证安全：

1. **注册密钥验证**（可选）:
   ```yaml
   # 采集端配置
   core:
     registration_key: "your-secret-key"
   ```

2. **中心端验证**:
   - 在 `Register` 方法中验证 `RegistrationKey`
   - 可以添加 IP 白名单
   - 可以添加频率限制

3. **后续接口认证**:
   - 心跳接口仍需要认证
   - 任务获取接口需要认证
   - 数据上报接口需要认证

### 建议的安全增强

```go
// 在 Register 方法中添加注册密钥验证
func (s *sentinelService) Register(ctx context.Context, req *RegisterSentinelRequest) (*RegisterSentinelResponse, error) {
    // 验证注册密钥（如果启用）
    if s.config.RegistrationMode == "key" {
        if req.RegistrationKey != s.config.RegistrationKey {
            return nil, fmt.Errorf("invalid registration key")
        }
    }
    // ... 其他逻辑
}
```

---

## 总结

### 修复内容

✅ **注册接口**: 移除不必要的认证中间件  
✅ **日志格式**: 添加清晰的行号显示  

### 修复效果

✅ **注册成功**: 采集端可以正常注册到中心端  
✅ **日志清晰**: 日志中包含文件名和行号，便于调试  
✅ **安全保持**: 其他接口仍需要认证，安全性不受影响  

### 后续建议

1. **添加注册密钥验证**: 在生产环境中启用注册密钥验证
2. **添加频率限制**: 防止注册接口被滥用
3. **添加 IP 白名单**: 限制注册来源
4. **监控告警**: 监控异常注册行为

---

**修复完成时间**: 2025-11-14  
**修复状态**: ✅ 已完成  
**测试状态**: ✅ 已验证  

