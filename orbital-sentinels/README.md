# Orbital Sentinels - 采集端

Celestial 监控系统的分布式数据采集端，负责从各种设备和系统中采集监控数据。

## ✨ 特性

- 🔌 **插件化架构**: 标准化插件接口，支持快速扩展
- 🚀 **高性能**: Go 语言实现，低资源占用
- 💪 **高可靠**: 本地缓冲、断线重连、失败重试
- 🌐 **灵活部署**: 支持边缘计算、跨网络采集
- 📊 **多种数据流**: 直连、中转、混合三种模式
- 🔄 **热更新**: 支持插件和配置热更新
- 📝 **本地任务**: 支持配置文件定义任务，无需中心端

## 📦 安装

### 二进制安装

```bash
# 下载最新版本
wget https://releases.celestial.io/sentinel/v1.0.0/sentinel-linux-amd64.tar.gz

# 解压
tar -xzf sentinel-linux-amd64.tar.gz
cd sentinel

# 配置
cp config/config.example.yaml config/config.yaml
vi config/config.yaml

# 运行
./sentinel start
```

### Docker 安装

```bash
# 拉取镜像
docker pull celestial/sentinel:latest

# 运行
docker run -d \
  --name sentinel \
  -v ./config:/app/config \
  -v ./plugins:/app/plugins \
  celestial/sentinel:latest
```

### 从源码构建

```bash
# 克隆项目
git clone https://github.com/celestial/orbital-sentinels.git
cd orbital-sentinels

# 安装依赖
go mod download

# 构建
make build

# 运行
./bin/sentinel start
```

## 🚀 快速开始

### 1. 配置文件

编辑 `config/config.yaml`:

```yaml
sentinel:
  name: "sentinel-office-1"
  region: "office-beijing"

core:
  url: "https://gravital-core.example.com"
  api_token: "your_api_token_here"

collector:
  worker_pool_size: 10
  task_fetch_interval: 60s

sender:
  mode: "core"  # core, direct, hybrid
  batch_size: 1000
```

### 2. 快速测试（无需配置）

```bash
# 手动触发一次 Ping 采集
./bin/sentinel trigger ping 8.8.8.8

# 自定义参数
./bin/sentinel trigger ping 8.8.8.8 -n 10 -i 500ms

# 查看帮助
./bin/sentinel trigger --help
```

### 3. 启动服务

```bash
# 前台运行
./bin/sentinel start -c config/config.yaml

# 后台运行
nohup ./bin/sentinel start -c config/config.yaml > sentinel.log 2>&1 &

# 使用 systemd
sudo systemctl start sentinel
```

### 4. 查看日志

```bash
# 查看实时日志
tail -f logs/sentinel.log

# 使用 systemd
sudo journalctl -u sentinel -f
```

## 📖 配置说明

### 核心配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| sentinel.name | Sentinel 名称 | - |
| sentinel.region | 所属区域 | - |
| core.url | 中心端地址 | - |
| core.api_token | API Token | - |

### 采集器配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| collector.worker_pool_size | 并发采集数 | 10 |
| collector.task_fetch_interval | 任务拉取间隔 | 60s |
| collector.max_execution_time | 单个任务最大执行时间 | 300s |

### 发送器配置

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| sender.mode | 发送模式 | core, direct, hybrid |
| sender.batch_size | 批量大小 | 1000 |
| sender.flush_interval | 刷新间隔 | 10s |

**发送模式说明**:
- `core`: 数据发送到中心端，由中心端转发
- `direct`: 数据直接发送到时序数据库（Prometheus/VictoriaMetrics/ClickHouse）
- `hybrid`: 混合模式，同时发送到中心端和时序数据库

### 本地任务配置

无需中心端，直接在配置文件中定义采集任务：

```yaml
sender:
  mode: "direct"  # 使用直连模式
  flush_interval: 10s
  direct:
    prometheus:
      enabled: true
      url: "http://localhost:9090/api/v1/write"

# 本地任务配置
tasks:
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"       # 每 60 秒执行一次
    timeout: "10s"
    enabled: true
    config:
      host: "192.168.1.1"
      count: 4

  - id: "ping-dns"
    device_id: "8.8.8.8"
    plugin: "ping"
    interval: "300s"      # 每 5 分钟执行一次
    enabled: true
    config:
      host: "8.8.8.8"
      count: 4
```

**适用场景**:
- ✅ 边缘计算、独立监控
- ✅ 监控目标相对固定
- ✅ 无需部署中心端
- ✅ 配置即用，快速开始

详细文档：[本地任务配置指南](docs/LOCAL_TASKS_GUIDE.md)

### 直连配置

支持直接发送数据到以下时序数据库：

| 数据库 | 协议 | 说明 |
|--------|------|------|
| Prometheus | Remote Write | 需启用 Remote Write Receiver |
| VictoriaMetrics | Remote Write | 兼容 Prometheus 协议 |
| ClickHouse | Native TCP | 高性能列式存储 |

详细配置说明请参考：
- [配置文档](config/config.example.yaml)
- [任务获取与采集流程](docs/TASK_COLLECTION_FLOW.md) ⭐
- [直连发送器指南](docs/DIRECT_SENDER_GUIDE.md)
- [独立运行模式](docs/STANDALONE_MODE.md)
- [手动触发采集](docs/TRIGGER_GUIDE.md)

## 🔌 插件开发

### 创建插件

```go
package main

import (
    "context"
    "github.com/celestial/orbital-sentinels/internal/plugin"
    "github.com/celestial/orbital-sentinels/sdk"
)

type MyPlugin struct {
    sdk.BasePlugin
}

func (p *MyPlugin) Meta() plugin.PluginMeta {
    return plugin.PluginMeta{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Description: "My awesome plugin",
    }
}

func (p *MyPlugin) Collect(ctx context.Context, task *plugin.CollectionTask) ([]*plugin.Metric, error) {
    // 实现采集逻辑
    return []*plugin.Metric{
        {
            Name:  "my_metric",
            Value: 42.0,
            Labels: map[string]string{
                "device_id": task.DeviceID,
            },
        },
    }, nil
}

func NewPlugin() plugin.Plugin {
    return &MyPlugin{}
}
```

详细插件开发指南请参考 [插件开发文档](../../docs/04-插件开发指南.md)。

## 📊 内置插件

| 插件 | 说明 | 状态 |
|------|------|------|
| ping | ICMP Ping 连通性检测 | ✅ |
| snmp | SNMP 协议采集 | 🚧 |
| http | HTTP/HTTPS 监控 | 🚧 |
| modbus | Modbus 协议采集 | 🚧 |
| mqtt | MQTT 消息监控 | 🚧 |

## 🛠️ 命令行

```bash
# 启动服务
sentinel start [-c config.yaml]

# 查看版本
sentinel version

# 测试连接
sentinel test-connection --plugin ping --host 192.168.1.1

# 列出插件
sentinel list-plugins

# 验证配置
sentinel validate-config
```

## 📈 监控指标

Sentinel 自身导出以下监控指标：

```
sentinel_uptime_seconds           # 运行时间
sentinel_tasks_total              # 任务总数
sentinel_tasks_success_total      # 成功任务数
sentinel_tasks_failed_total       # 失败任务数
sentinel_plugins_loaded           # 已加载插件数
sentinel_buffer_size              # 缓冲区大小
sentinel_sent_metrics_total       # 已发送指标数
sentinel_cpu_usage_percent        # CPU 使用率
sentinel_memory_usage_bytes       # 内存使用量
```

## 🐛 故障排查

### 问题：无法连接中心端

**解决方案**:
1. 检查网络连通性: `curl https://gravital-core.example.com/health`
2. 检查 API Token 是否正确
3. 查看日志: `tail -f logs/sentinel.log`

### 问题：插件加载失败

**解决方案**:
1. 检查插件目录: `ls -la plugins/`
2. 检查插件配置: `cat plugins/*/plugin.yaml`
3. 查看错误日志

### 问题：数据未采集

**解决方案**:
1. 检查任务是否分配: 查看日志中的 "Added task"
2. 检查插件是否正常: `sentinel list-plugins`
3. 手动测试插件: `sentinel test-connection --plugin ping --host <host>`

更多故障排查指南请参考 [故障排查文档](docs/troubleshooting.md)。

## 📚 文档

- [本地任务配置指南](docs/LOCAL_TASKS_GUIDE.md) - 无中心端场景配置
- [任务获取与采集流程](docs/TASK_COLLECTION_FLOW.md) - 任务执行流程
- [直连发送器指南](docs/DIRECT_SENDER_GUIDE.md) - 直连数据库配置
- [独立运行模式](docs/STANDALONE_MODE.md) - 无中心端运行
- [手动触发指南](docs/TRIGGER_GUIDE.md) - 测试和调试
- [配置文档](docs/configuration.md)
- [插件开发指南](../../docs/04-插件开发指南.md)
- [API 文档](docs/api.md)
- [故障排查](docs/troubleshooting.md)

## 🤝 贡献

欢迎贡献代码、报告问题或提出建议！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📄 许可证

Apache 2.0 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

感谢所有贡献者的支持！

---

Made with ❤️ by Celestial Team

