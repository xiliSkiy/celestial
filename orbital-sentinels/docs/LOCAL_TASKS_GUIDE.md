# 本地任务配置指南

本文档说明如何在无中心端的场景下，通过配置文件直接定义采集任务。

## 🎯 适用场景

- ✅ **无中心端部署** - 边缘计算、独立监控
- ✅ **简化部署** - 无需部署和维护中心端
- ✅ **固定任务** - 监控目标相对固定
- ✅ **快速开始** - 配置即用，无需额外开发

## 📝 配置格式

### 基本结构

```yaml
# config.yaml
tasks:
  - id: "任务唯一标识"
    device_id: "设备ID"
    plugin: "插件名称"
    interval: "执行间隔"
    timeout: "超时时间"
    enabled: true/false
    config:
      # 插件特定配置
```

### 完整示例

```yaml
sentinel:
  name: "sentinel-standalone"
  region: "local"

sender:
  mode: "direct"  # 使用 direct 模式
  flush_interval: 10s
  direct:
    prometheus:
      enabled: true
      url: "http://localhost:9090/api/v1/write"

# 本地任务配置
tasks:
  # 任务 1: 监控网关
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"       # 每 60 秒执行一次
    timeout: "10s"        # 超时时间
    enabled: true         # 启用此任务
    config:
      host: "192.168.1.1"
      count: 4
      interval: "1s"
      timeout: "5s"

  # 任务 2: 监控 DNS
  - id: "ping-dns"
    device_id: "8.8.8.8"
    plugin: "ping"
    interval: "300s"      # 每 5 分钟执行一次
    timeout: "10s"
    enabled: true
    config:
      host: "8.8.8.8"
      count: 4

  # 任务 3: 禁用的任务
  - id: "ping-disabled"
    device_id: "192.168.1.100"
    plugin: "ping"
    interval: "60s"
    enabled: false        # 禁用此任务
    config:
      host: "192.168.1.100"
```

## 🔧 字段说明

### 必需字段

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | string | 任务唯一标识符 | "ping-gateway" |
| `device_id` | string | 设备ID，用于标识被监控设备 | "192.168.1.1" |
| `plugin` | string | 使用的插件名称 | "ping" |
| `interval` | string | 执行间隔 | "60s", "5m", "1h" |
| `enabled` | boolean | 是否启用此任务 | true/false |
| `config` | object | 插件特定的配置参数 | 见下文 |

### 可选字段

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `timeout` | string | "30s" | 任务超时时间 |

### 间隔格式

支持以下时间格式：

- `30s` - 30 秒
- `1m` - 1 分钟
- `5m` - 5 分钟
- `1h` - 1 小时
- `24h` - 24 小时

## 🔌 插件配置

### Ping 插件

```yaml
- id: "ping-example"
  device_id: "8.8.8.8"
  plugin: "ping"
  interval: "60s"
  enabled: true
  config:
    host: "8.8.8.8"        # 必需：目标主机
    count: 4               # 可选：Ping 次数，默认 4
    interval: "1s"         # 可选：Ping 间隔，默认 1s
    timeout: "5s"          # 可选：Ping 超时，默认 5s
```

### 未来支持的插件

```yaml
# SNMP 插件（计划中）
- id: "snmp-switch"
  device_id: "192.168.1.100"
  plugin: "snmp"
  interval: "300s"
  enabled: true
  config:
    host: "192.168.1.100"
    community: "public"
    version: "2c"
    oids:
      - "1.3.6.1.2.1.1.1.0"  # sysDescr

# HTTP 插件（计划中）
- id: "http-api"
  device_id: "api.example.com"
  plugin: "http"
  interval: "60s"
  enabled: true
  config:
    url: "https://api.example.com/health"
    method: "GET"
    timeout: "10s"
```

## 🚀 使用步骤

### 1. 复制配置模板

```bash
cp config/config.local-tasks.yaml config/config.yaml
```

### 2. 编辑配置文件

```bash
vim config/config.yaml
```

修改以下部分：

1. **Sentinel 信息**:
```yaml
sentinel:
  name: "your-sentinel-name"
  region: "your-region"
```

2. **发送目标**:
```yaml
sender:
  mode: "direct"
  direct:
    prometheus:
      enabled: true
      url: "http://your-prometheus:9090/api/v1/write"
```

3. **任务列表**:
```yaml
tasks:
  - id: "your-task-1"
    device_id: "your-device-1"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "your-host-1"
```

### 3. 验证配置

```bash
# 使用配置检查脚本
./scripts/check-config.sh config/config.yaml
```

### 4. 启动 Sentinel

```bash
./bin/sentinel start -c config/config.yaml
```

### 5. 查看日志

```bash
# 查看任务加载情况
tail -f logs/sentinel.log | grep "Loaded local task"

# 查看任务执行情况
tail -f logs/sentinel.log | grep "Task succeeded"

# 查看发送情况
tail -f logs/sentinel.log | grep "Sent to"
```

## 📊 验证数据

### 查看 Prometheus

```bash
# 访问 Prometheus
open http://localhost:9090

# 查询数据
ping_rtt_ms{device_id="192.168.1.1"}
ping_packet_loss{device_id="8.8.8.8"}
```

### 查看日志统计

```bash
# 查看成功/失败统计
tail logs/sentinel.log | grep "Sender stopped"

# 输出示例：
# {"msg":"Sender stopped","success_count":120,"failed_count":0}
```

## 💡 最佳实践

### 1. 任务命名规范

使用清晰的命名规则：

```yaml
tasks:
  - id: "ping-gateway-office"      # 功能-设备-位置
  - id: "ping-dns-google"          # 功能-设备-提供商
  - id: "ping-server-web-01"       # 功能-类型-编号
```

### 2. 合理设置间隔

根据监控需求设置不同的间隔：

```yaml
tasks:
  # 关键设备 - 高频监控
  - id: "ping-gateway"
    interval: "30s"
    
  # 普通设备 - 中频监控
  - id: "ping-server"
    interval: "60s"
    
  # 外部服务 - 低频监控
  - id: "ping-public-dns"
    interval: "300s"
```

### 3. 分组管理

使用注释分组管理任务：

```yaml
tasks:
  # ========== 网络设备 ==========
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.1"

  - id: "ping-switch"
    device_id: "192.168.1.100"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.100"

  # ========== 服务器 ==========
  - id: "ping-web-server"
    device_id: "192.168.1.10"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.10"

  # ========== 外部服务 ==========
  - id: "ping-google-dns"
    device_id: "8.8.8.8"
    plugin: "ping"
    interval: "300s"
    enabled: true
    config:
      host: "8.8.8.8"
```

### 4. 禁用而非删除

暂时不需要的任务设为 disabled 而不是删除：

```yaml
- id: "ping-old-server"
  device_id: "192.168.1.99"
  plugin: "ping"
  interval: "60s"
  enabled: false  # 暂时禁用，保留配置
  config:
    host: "192.168.1.99"
```

## 🔍 故障排查

### 问题 1: 任务未加载

**症状**:
```
{"msg":"Local tasks loaded","success":0,"failed":5}
```

**可能原因**:
1. YAML 格式错误
2. 必需字段缺失
3. interval 格式错误

**解决方案**:
```bash
# 1. 检查 YAML 格式
yamllint config/config.yaml

# 2. 查看详细错误
tail -f logs/sentinel.log | grep "ERROR"

# 3. 验证配置
./scripts/check-config.sh config/config.yaml
```

### 问题 2: 插件未找到

**症状**:
```
{"level":"ERROR","msg":"Plugin not found","plugin":"ping"}
```

**原因**: 插件未正确注册

**解决方案**:
```bash
# 查看插件注册日志
tail logs/sentinel.log | grep "Registered builtin plugin"

# 应该看到：
# {"msg":"Registered builtin plugin","name":"ping"}
```

### 问题 3: 任务不执行

**症状**: 没有采集数据

**检查步骤**:

1. 确认任务已启用：
```yaml
enabled: true  # 不是 false
```

2. 查看任务加载：
```bash
tail logs/sentinel.log | grep "Loaded local task"
```

3. 查看执行日志：
```bash
tail -f logs/sentinel.log | grep -E "(Task|metrics)"
```

## 📚 示例配置

### 示例 1: 小型办公室

```yaml
tasks:
  # 网关
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "192.168.1.1"
      count: 4

  # 交换机
  - id: "ping-switch"
    device_id: "192.168.1.100"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.100"
      count: 4

  # 打印机
  - id: "ping-printer"
    device_id: "192.168.1.200"
    plugin: "ping"
    interval: "300s"
    enabled: true
    config:
      host: "192.168.1.200"
      count: 2
```

### 示例 2: 数据中心

```yaml
tasks:
  # Web 服务器集群
  - id: "ping-web-01"
    device_id: "10.0.1.10"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "10.0.1.10"

  - id: "ping-web-02"
    device_id: "10.0.1.11"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "10.0.1.11"

  # 数据库服务器
  - id: "ping-db-master"
    device_id: "10.0.2.10"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "10.0.2.10"

  - id: "ping-db-slave"
    device_id: "10.0.2.11"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "10.0.2.11"
```

### 示例 3: 混合环境

```yaml
tasks:
  # 内网设备 - 高频
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "30s"
    enabled: true
    config:
      host: "192.168.1.1"
      count: 4

  # 外网服务 - 中频
  - id: "ping-website"
    device_id: "example.com"
    plugin: "ping"
    interval: "300s"
    enabled: true
    config:
      host: "example.com"
      count: 3

  # 公共 DNS - 低频
  - id: "ping-google-dns"
    device_id: "8.8.8.8"
    plugin: "ping"
    interval: "600s"
    enabled: true
    config:
      host: "8.8.8.8"
      count: 3
```

## 🆚 对比：本地任务 vs 中心端任务

| 特性 | 本地任务配置 | 中心端任务 |
|------|-------------|-----------|
| 部署复杂度 | ⭐ 简单 | ⭐⭐⭐ 复杂 |
| 配置方式 | 配置文件 | Web 界面/API |
| 动态更新 | ❌ 需要重启 | ✅ 实时更新 |
| 集中管理 | ❌ 分散管理 | ✅ 统一管理 |
| 适用规模 | 小型（<50 设备） | 大型（>50 设备） |
| 网络依赖 | ❌ 无依赖 | ✅ 需要网络 |

## 🔮 未来改进

计划支持的功能：

- [ ] 热重载配置（无需重启）
- [ ] 任务模板
- [ ] 条件执行
- [ ] 任务依赖
- [ ] 从文件导入任务列表

## 📖 参考

- [配置文件示例](../config/config.local-tasks.yaml)
- [任务获取与采集流程](./TASK_COLLECTION_FLOW.md)
- [Ping 插件文档](../plugins/ping/README.md)

## 总结

本地任务配置是无中心端场景下的最佳选择：

✅ **简单** - 配置文件即可定义任务  
✅ **灵活** - 支持多种插件和配置  
✅ **可靠** - 无需依赖外部服务  
✅ **高效** - 直接发送到时序数据库  

立即开始使用：

```bash
# 1. 复制配置
cp config/config.local-tasks.yaml config/config.yaml

# 2. 编辑任务
vim config/config.yaml

# 3. 启动服务
./bin/sentinel start -c config/config.yaml
```

