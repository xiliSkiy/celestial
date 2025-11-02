# 独立运行模式

本文档说明如何在中心端（Gravital Core）不可用或未部署时，独立运行 Orbital Sentinels 采集端。

## 概述

Orbital Sentinels 支持三种运行模式：

1. **Core 模式** - 需要中心端，数据发送到中心端
2. **Direct 模式** - 不需要中心端，数据直接发送到时序数据库
3. **Hybrid 模式** - 需要中心端，数据同时发送到中心端和时序数据库

当中心端不可用时，推荐使用 **Direct 模式**。

## 配置独立运行

### 方式一：使用 Direct 模式（推荐）

这是最简单的方式，完全不依赖中心端。

#### 1. 修改配置文件

编辑 `config.yaml`：

```yaml
sentinel:
  id: ""
  name: "sentinel-standalone-1"
  region: "local"

# 心跳配置（可选，中心端不可用时会自动跳过）
heartbeat:
  interval: 30s
  timeout: 10s
  retry_times: 3

# 采集器配置
collector:
  worker_pool_size: 10
  task_fetch_interval: 60s
  max_execution_time: 300s

# 缓冲配置
buffer:
  type: "memory"
  size: 10000
  flush_interval: 10s

# 发送器配置 - 使用 Direct 模式
sender:
  mode: "direct"              # 关键：使用 direct 模式
  batch_size: 1000
  flush_interval: 10s         # 必须配置
  timeout: 30s
  retry_times: 3
  retry_interval: 5s
  
  # 直连配置 - 至少启用一个
  direct:
    prometheus:
      enabled: true
      url: "http://localhost:9090/api/v1/write"
    
    victoria_metrics:
      enabled: false
      url: "http://localhost:8428/api/v1/write"
    
    clickhouse:
      enabled: false
      dsn: "tcp://localhost:9000/metrics?username=default&password=&compress=true"
      table_name: "metrics"
      batch_size: 1000

# 插件配置
plugins:
  directory: "./plugins"
  auto_reload: false

# 日志配置
logging:
  level: info
  format: json
  output: both
  file_path: "./logs/sentinel.log"
  max_size: 100
  max_backups: 7
  max_age: 30
```

#### 2. 启动 Sentinel

```bash
./bin/sentinel start -c config.yaml
```

#### 3. 验证运行

检查日志：

```bash
tail -f logs/sentinel.log
```

应该看到类似的输出：

```json
{"level":"INFO","msg":"Sender started","mode":"direct"}
{"level":"INFO","msg":"Scheduler started"}
{"level":"WARN","msg":"Failed to send heartbeat (core may be unavailable)"}
```

**注意**：心跳失败的警告是正常的，不会影响采集端运行。

### 方式二：Core 模式（中心端不可用时）

如果你暂时使用 Core 模式但中心端不可用，Sentinel 仍然可以启动，但数据会缓存在本地。

#### 配置示例

```yaml
sender:
  mode: "core"
  batch_size: 1000
  flush_interval: 10s         # 必须配置
  timeout: 30s
  retry_times: 3
  retry_interval: 5s

core:
  url: "http://gravital-core:8080"  # 即使不可用也没关系
  api_token: "dummy-token"
  insecure_skip_verify: true
```

#### 行为说明

- ✅ Sentinel 可以正常启动
- ✅ 采集任务可以正常执行
- ✅ 数据会缓存在本地缓冲区
- ⚠️ 心跳会失败（仅记录警告）
- ⚠️ 数据发送会失败（会重试）
- ⚠️ 缓冲区满时会丢弃旧数据

当中心端恢复后，新数据会自动发送。

## 常见问题

### Q1: 配置文件缺少 flush_interval 导致启动失败

**错误信息**：
```
panic: non-positive interval for NewTicker
```

**解决方案**：
在配置文件的 `sender` 部分添加 `flush_interval`：

```yaml
sender:
  mode: "direct"
  batch_size: 1000
  flush_interval: 10s  # 添加这一行
  timeout: 30s
```

### Q2: 心跳一直失败

**日志信息**：
```json
{"level":"WARN","msg":"Failed to send heartbeat (core may be unavailable)"}
```

**说明**：
这是正常的。当使用 Direct 模式或中心端不可用时，心跳会失败，但不会影响采集端的正常运行。

**解决方案**：
- 如果不需要中心端，忽略这个警告即可
- 如果需要中心端，请确保中心端正常运行并且 `core.url` 配置正确

### Q3: 数据发送失败

**日志信息**：
```json
{"level":"ERROR","msg":"Failed to send to core"}
```

**原因**：
- Core 模式下中心端不可用
- Direct 模式下时序数据库不可用

**解决方案**：
1. 检查目标服务是否运行
2. 检查网络连接
3. 检查配置是否正确
4. 查看详细错误信息

### Q4: 缓冲区满了怎么办

**说明**：
当发送失败且缓冲区满时，会自动丢弃最旧的数据。

**解决方案**：
1. 增加缓冲区大小：
```yaml
buffer:
  size: 50000  # 增加到 50000
```

2. 减少刷新间隔：
```yaml
sender:
  flush_interval: 5s  # 从 10s 减少到 5s
```

3. 修复发送问题，让数据能够正常发送

## 部署示例

### 使用 Docker Compose（仅时序数据库）

如果只需要时序数据库而不需要中心端：

```yaml
version: '3.8'

services:
  # VictoriaMetrics
  victoria-metrics:
    image: victoriametrics/victoria-metrics:latest
    container_name: victoria-metrics
    command:
      - '--storageDataPath=/victoria-metrics-data'
      - '--httpListenAddr=:8428'
      - '--retentionPeriod=12'
    volumes:
      - victoria-data:/victoria-metrics-data
    ports:
      - "8428:8428"
    restart: unless-stopped

  # Grafana
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    ports:
      - "3000:3000"
    depends_on:
      - victoria-metrics
    restart: unless-stopped

volumes:
  victoria-data:
  grafana-data:
```

### 配置 Sentinel

```yaml
sender:
  mode: "direct"
  batch_size: 1000
  flush_interval: 10s
  timeout: 30s
  
  direct:
    victoria_metrics:
      enabled: true
      url: "http://localhost:8428/api/v1/write"
```

### 启动服务

```bash
# 1. 启动时序数据库和 Grafana
docker-compose up -d

# 2. 启动 Sentinel
./bin/sentinel start -c config.yaml

# 3. 访问 Grafana
open http://localhost:3000
```

## 最佳实践

### 1. 选择合适的模式

| 场景 | 推荐模式 | 说明 |
|------|---------|------|
| 开发测试 | Direct | 简单快速 |
| 边缘计算 | Direct | 减少依赖 |
| 小型部署 | Direct | 降低复杂度 |
| 大型部署 | Core/Hybrid | 统一管理 |
| 高可用 | Hybrid | 双重保障 |

### 2. 配置检查清单

在启动前检查：

- [ ] `sender.flush_interval` 已配置
- [ ] 至少启用一个数据目标
- [ ] 目标服务正常运行
- [ ] 网络连接正常
- [ ] 配置文件语法正确

### 3. 监控和告警

建议监控以下指标：

- Sentinel 进程状态
- 缓冲区使用率
- 发送成功率
- 发送延迟
- 系统资源使用

### 4. 日志管理

```yaml
logging:
  level: info          # 生产环境使用 info
  format: json         # JSON 格式便于解析
  output: both         # 同时输出到文件和标准输出
  file_path: "./logs/sentinel.log"
  max_size: 100        # 单文件最大 100MB
  max_backups: 7       # 保留 7 个备份
  max_age: 30          # 保留 30 天
```

## 故障排查

### 启动失败

1. 检查配置文件语法
2. 检查必需字段是否配置
3. 查看详细错误信息
4. 检查日志文件

### 数据未写入

1. 检查目标服务是否运行
2. 验证连接配置
3. 查看发送日志
4. 检查认证信息

### 性能问题

1. 增加批量大小
2. 调整刷新间隔
3. 增加工作池大小
4. 启用数据压缩

## 总结

Orbital Sentinels 可以在中心端不可用时独立运行：

✅ **推荐方式**：使用 Direct 模式直连时序数据库
✅ **关键配置**：必须设置 `sender.flush_interval`
✅ **容错机制**：心跳失败不影响采集和发送
✅ **数据保护**：本地缓冲防止数据丢失

如有问题，请查看日志文件或提交 Issue。

