# 快速开始 - 无中心端部署

本指南将帮助你在 5 分钟内完成 Orbital Sentinel 的部署和配置，无需部署中心端。

## 🎯 目标

完成本指南后，你将拥有：

- ✅ 一个运行中的 Sentinel 实例
- ✅ 自动采集网络设备的 Ping 数据
- ✅ 数据直接发送到 Prometheus
- ✅ 可视化监控面板

## 📋 前置条件

- Linux/macOS 系统
- Go 1.21+ （如果从源码构建）
- Prometheus（可选，用于存储数据）

## 🚀 步骤 1: 构建 Sentinel

```bash
# 克隆项目
git clone https://github.com/celestial/orbital-sentinels.git
cd orbital-sentinels

# 构建
make build

# 验证
./bin/sentinel version
```

输出示例：
```
Sentinel version 1.0.0
```

## 📝 步骤 2: 配置 Sentinel

使用提供的本地任务配置模板：

```bash
# 复制配置模板
cp config/config.local-tasks.yaml config/config.yaml

# 创建日志目录
mkdir -p logs
```

编辑配置文件（可选）：

```bash
vim config/config.yaml
```

关键配置项：

```yaml
sentinel:
  name: "sentinel-standalone"  # 修改为你的名称
  region: "local"

sender:
  mode: "direct"
  flush_interval: 10s
  direct:
    prometheus:
      enabled: true
      url: "http://localhost:9090/api/v1/write"  # 修改为你的 Prometheus 地址

tasks:
  # 修改为你要监控的设备
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.1"
```

## 🏃 步骤 3: 启动 Sentinel

```bash
# 前台运行（用于测试）
./bin/sentinel start -c config/config.yaml

# 或后台运行
nohup ./bin/sentinel start -c config/config.yaml > sentinel.log 2>&1 &
```

查看日志确认启动成功：

```bash
tail -f logs/sentinel.log
```

你应该看到类似的输出：

```json
{"level":"INFO","msg":"Starting Sentinel","version":"1.0.0","name":"sentinel-standalone"}
{"level":"INFO","msg":"Registered builtin plugin","name":"ping"}
{"level":"INFO","msg":"Local tasks loaded","success":4,"failed":0,"total":5}
{"level":"INFO","msg":"All components started"}
```

## ✅ 步骤 4: 验证数据采集

### 方法 1: 查看日志

```bash
# 查看发送统计
tail logs/sentinel.log | grep "Sender stopped"

# 输出示例：
# {"msg":"Sender stopped","success_count":120,"failed_count":0}
```

### 方法 2: 查询 Prometheus

如果你配置了 Prometheus，可以查询数据：

```bash
# 访问 Prometheus
open http://localhost:9090

# 在查询框中输入：
ping_rtt_ms{device_id="192.168.1.1"}
```

### 方法 3: 手动触发测试

```bash
# 手动触发一次 Ping 采集
./bin/sentinel trigger ping 8.8.8.8 -n 4

# 输出示例：
# ✓ Ping 采集成功
# 
# 采集到 4 个指标:
# - ping_rtt_ms{host="8.8.8.8"} = 14.5
# - ping_packet_loss{host="8.8.8.8"} = 0.0
```

## 📊 步骤 5: 配置 Grafana（可选）

### 5.1 添加 Prometheus 数据源

1. 访问 Grafana: http://localhost:3000
2. 添加数据源 → Prometheus
3. URL: http://localhost:9090
4. 保存并测试

### 5.2 导入仪表板

创建一个简单的面板：

**查询 1 - Ping RTT**:
```promql
ping_rtt_ms
```

**查询 2 - 丢包率**:
```promql
ping_packet_loss
```

## 🎨 自定义配置

### 添加更多监控目标

编辑 `config/config.yaml`，添加新任务：

```yaml
tasks:
  # 现有任务...
  
  # 新增任务
  - id: "ping-new-server"
    device_id: "192.168.1.100"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.100"
      count: 4
```

重启 Sentinel：

```bash
# 停止
pkill -f "sentinel start"

# 启动
./bin/sentinel start -c config/config.yaml
```

### 调整采集频率

修改 `interval` 字段：

```yaml
tasks:
  - id: "ping-critical"
    interval: "30s"    # 高频：每 30 秒
    
  - id: "ping-normal"
    interval: "60s"    # 中频：每 1 分钟
    
  - id: "ping-low"
    interval: "300s"   # 低频：每 5 分钟
```

## 🔧 常见问题

### Q1: Sentinel 启动失败

**错误**: `Failed to initialize logger: open ./logs/sentinel.log: no such file or directory`

**解决**:
```bash
mkdir -p logs
```

### Q2: 插件未找到

**错误**: `Plugin not found: ping`

**解决**: 确保插件已注册，查看日志：
```bash
tail logs/sentinel.log | grep "Registered builtin plugin"
```

### Q3: 数据未发送到 Prometheus

**错误**: `Failed to write to Prometheus: connection refused`

**检查**:
1. Prometheus 是否运行：`curl http://localhost:9090/-/healthy`
2. URL 是否正确：检查 `config/config.yaml` 中的 `sender.direct.prometheus.url`
3. 网络是否可达：`ping localhost`

### Q4: 任务未执行

**检查**:
1. 任务是否启用：`enabled: true`
2. 查看任务加载日志：
```bash
tail logs/sentinel.log | grep "Loaded local task"
```

## 📈 监控 Sentinel 自身

Sentinel 会自动采集自身的运行指标：

```promql
# 运行时间
sentinel_uptime_seconds

# 任务统计
sentinel_tasks_total
sentinel_tasks_success_total
sentinel_tasks_failed_total

# 资源使用
sentinel_cpu_usage_percent
sentinel_memory_usage_bytes
```

## 🔄 生产环境部署

### 使用 Systemd

创建服务文件：

```bash
sudo vim /etc/systemd/system/sentinel.service
```

内容：

```ini
[Unit]
Description=Orbital Sentinel
After=network.target

[Service]
Type=simple
User=sentinel
WorkingDirectory=/opt/sentinel
ExecStart=/opt/sentinel/bin/sentinel start -c /opt/sentinel/config/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable sentinel
sudo systemctl start sentinel
sudo systemctl status sentinel
```

### 使用 Docker

```bash
# 构建镜像
docker build -t sentinel:latest .

# 运行容器
docker run -d \
  --name sentinel \
  --restart always \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  sentinel:latest
```

## 🎓 下一步

现在你已经有了一个运行中的 Sentinel，可以：

1. **添加更多插件** - 查看 [插件开发指南](../../docs/04-插件开发指南.md)
2. **配置直连数据库** - 查看 [直连发送器指南](docs/DIRECT_SENDER_GUIDE.md)
3. **优化性能** - 调整 `worker_pool_size`、`batch_size` 等参数
4. **部署中心端** - 实现集中管理和动态任务分发

## 📚 相关文档

- [本地任务配置指南](docs/LOCAL_TASKS_GUIDE.md) - 详细的任务配置说明
- [独立运行模式](docs/STANDALONE_MODE.md) - 无中心端运行的详细说明
- [手动触发指南](docs/TRIGGER_GUIDE.md) - 测试和调试工具
- [故障排查](docs/troubleshooting.md) - 常见问题解决

## 💬 获取帮助

- 📖 查看文档：[docs/](docs/)
- 🐛 报告问题：[GitHub Issues](https://github.com/celestial/orbital-sentinels/issues)
- 💬 讨论交流：[GitHub Discussions](https://github.com/celestial/orbital-sentinels/discussions)

---

🎉 恭喜！你已经成功部署了 Orbital Sentinel！

现在可以开始监控你的网络设备了。如有任何问题，请查看上述文档或联系我们。
