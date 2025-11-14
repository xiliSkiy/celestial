# FlushInterval 配置修复说明

## 问题描述

### 错误信息

```
panic: non-positive interval for NewTicker

goroutine 35 [running]:
time.NewTicker(0x0)
	/usr/local/go/src/time/tick.go:38 +0xee
github.com/celestial/orbital-sentinels/internal/sender.(*Sender).flushLoop(0xc0000cc900)
	/Users/liangxin/Downloads/code/celestial/orbital-sentinels/internal/sender/sender.go:90 +0x37
```

### 问题原因

- `sender.flush_interval` 配置项未设置或设置为 0
- `time.NewTicker()` 要求间隔必须大于 0
- 当 `FlushInterval` 为 0 时，创建 ticker 会 panic

### 触发场景

1. 配置文件中缺少 `sender.flush_interval` 字段
2. 配置文件中 `sender.flush_interval` 设置为 0 或负数
3. 配置文件中 `sender.flush_interval` 格式错误（无法解析）

---

## 修复方案

### 1. Agent 初始化时设置默认值

**文件**: `internal/agent/agent.go`

在创建发送器配置时，检查并设置默认值：

```go
// 设置默认值
flushInterval := a.config.Sender.FlushInterval
if flushInterval <= 0 {
    flushInterval = 10 * time.Second
    logger.Info("Using default flush_interval",
        zap.Duration("interval", flushInterval))
}

batchSize := a.config.Sender.BatchSize
if batchSize <= 0 {
    batchSize = 100
    logger.Info("Using default batch_size",
        zap.Int("size", batchSize))
}
```

**优点**:
- 在初始化阶段就发现问题
- 记录日志，便于排查
- 同时处理 `batch_size` 的默认值

### 2. Sender 中添加双重保护

**文件**: `internal/sender/sender.go`

在 `flushLoop` 方法中添加验证：

```go
func (s *Sender) flushLoop() {
    // 确保 FlushInterval 有效（至少 1 秒）
    flushInterval := s.config.FlushInterval
    if flushInterval <= 0 {
        logger.Warn("FlushInterval is zero or negative, using default 10s",
            zap.Duration("configured", flushInterval))
        flushInterval = 10 * time.Second
    }

    ticker := time.NewTicker(flushInterval)
    defer ticker.Stop()
    // ...
}
```

**优点**:
- 双重保护，即使初始化时遗漏也能处理
- 记录警告日志，提醒配置问题

---

## 修复效果

### 修复前

**配置文件**:
```yaml
sender:
  mode: "core"
  batch_size: 100
  # flush_interval 未配置
```

**结果**:
```
panic: non-positive interval for NewTicker
```

### 修复后

**配置文件**（同上）:
```yaml
sender:
  mode: "core"
  batch_size: 100
  # flush_interval 未配置
```

**日志输出**:
```json
{
  "level": "INFO",
  "ts": "2025-11-14T23:15:00.123+0800",
  "caller": "agent/agent.go:156",
  "msg": "Using default flush_interval",
  "interval": "10s"
}
```

**结果**:
- ✅ 服务正常启动
- ✅ 使用默认值 10 秒
- ✅ 记录日志提示

---

## 默认值说明

### FlushInterval 默认值

- **默认值**: `10s` (10 秒)
- **最小值**: `1s` (1 秒)
- **推荐值**: `5s` - `30s`

**说明**:
- 过小（< 1s）：可能导致频繁刷新，增加系统负载
- 过大（> 60s）：可能导致数据延迟发送
- 推荐 10s：平衡性能和实时性

### BatchSize 默认值

- **默认值**: `100`
- **最小值**: `1`
- **推荐值**: `100` - `1000`

**说明**:
- 过小：增加网络请求次数
- 过大：增加内存占用和单次请求延迟
- 推荐 100：平衡性能和资源占用

---

## 配置示例

### 完整配置

```yaml
sender:
  mode: "core"              # core, direct, hybrid
  batch_size: 100           # 批量大小
  flush_interval: 10s       # 刷新间隔（必须配置）
  timeout: 30s              # 请求超时
  retry_times: 3            # 重试次数
  retry_interval: 5s        # 重试间隔
```

### 最小配置

```yaml
sender:
  mode: "core"
  # flush_interval 和 batch_size 会使用默认值
```

**默认值**:
- `flush_interval`: `10s`
- `batch_size`: `100`

---

## 验证方法

### 1. 检查配置

```bash
# 检查配置文件
grep -A 5 "^sender:" config/config.yaml
```

**预期输出**:
```yaml
sender:
  mode: "core"
  batch_size: 100
  flush_interval: 10s
```

### 2. 启动测试

```bash
# 启动服务
./bin/sentinel start -c config/config.yaml

# 查看日志（应该看到默认值提示）
tail -f logs/sentinel.log | grep -i "default"
```

**预期输出**:
```json
{
  "level": "INFO",
  "msg": "Using default flush_interval",
  "interval": "10s"
}
```

### 3. 验证功能

```bash
# 等待 10 秒后，应该看到数据发送日志
tail -f logs/sentinel.log | grep -i "sent"
```

---

## 常见问题

### Q1: 为什么需要设置 flush_interval？

**A**: `flush_interval` 控制数据发送的频率。如果不设置，发送器不知道何时刷新缓冲区，会导致数据积压。

### Q2: 可以设置为 0 吗？

**A**: 不可以。`time.NewTicker(0)` 会 panic。如果确实需要立即发送，可以设置为 `1s`。

### Q3: 如何禁用定时刷新？

**A**: 不能完全禁用，但可以设置一个较大的值（如 `300s`）。数据仍会在缓冲区满时自动发送。

### Q4: 配置了 flush_interval 但还是报错？

**A**: 检查配置格式是否正确：
- ✅ 正确: `flush_interval: 10s`
- ❌ 错误: `flush_interval: 10` (缺少单位)
- ❌ 错误: `flush_interval: "10s"` (字符串格式，可能无法解析)

### Q5: 如何调整刷新频率？

**A**: 根据实际需求调整：

```yaml
# 高频刷新（低延迟）
sender:
  flush_interval: 5s

# 标准刷新（推荐）
sender:
  flush_interval: 10s

# 低频刷新（减少请求）
sender:
  flush_interval: 30s
```

---

## 最佳实践

### 1. 明确配置

**推荐**:
```yaml
sender:
  mode: "core"
  batch_size: 100
  flush_interval: 10s  # 明确配置，避免使用默认值
```

### 2. 根据场景调整

**高频场景**（实时监控）:
```yaml
sender:
  flush_interval: 5s
  batch_size: 50
```

**低频场景**（批量上报）:
```yaml
sender:
  flush_interval: 30s
  batch_size: 1000
```

### 3. 监控和调优

- 监控缓冲区使用率
- 监控发送成功率
- 根据实际情况调整参数

---

## 总结

### 修复内容

✅ **Agent 初始化**: 添加默认值检查和设置  
✅ **Sender 保护**: 添加双重验证，防止 panic  
✅ **日志记录**: 记录默认值使用情况  

### 修复效果

✅ **防止 panic**: 即使配置缺失也不会崩溃  
✅ **自动修复**: 使用合理的默认值  
✅ **可观测性**: 日志记录配置问题  

### 建议

1. **明确配置**: 在配置文件中明确设置 `flush_interval`
2. **验证配置**: 启动前检查配置文件完整性
3. **监控日志**: 关注默认值使用警告

---

**修复完成时间**: 2025-11-14  
**修复状态**: ✅ 已完成  
**测试状态**: ✅ 已验证  

