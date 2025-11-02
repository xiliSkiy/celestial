# Prometheus Remote Write 问题解决方案

## 🔍 问题描述

当使用 Sentinel 向 Prometheus 发送数据时，遇到以下错误：

```
Failed to write to Prometheus: failed to send request: Post "http://localhost:9090/api/v1/write": ...
```

## 🎯 根本原因

**Prometheus 默认不接受 Remote Write 请求！**

Prometheus 的 `/api/v1/write` 端点是用于**发送**数据到其他系统（如 VictoriaMetrics），而不是**接收**数据。

要接收远程写入的数据，你需要使用以下方案之一。

## ✅ 解决方案

### 方案 1: 使用 VictoriaMetrics（推荐）⭐

VictoriaMetrics 是 Prometheus 的高性能替代品，完全兼容 Prometheus 查询语言，且原生支持 Remote Write 接收。

#### 1.1 启动 VictoriaMetrics

```bash
# 使用 Docker 启动
docker run -d \
  --name victoria-metrics \
  -p 8428:8428 \
  -v victoria-data:/victoria-metrics-data \
  victoriametrics/victoria-metrics:latest
```

#### 1.2 修改 Sentinel 配置

编辑 `config/config.yaml`:

```yaml
sender:
  mode: "direct"
  flush_interval: 10s
  direct:
    victoria_metrics:
      enabled: true
      url: "http://localhost:8428/api/v1/write"  # 改为 VictoriaMetrics
    prometheus:
      enabled: false  # 禁用 Prometheus
```

#### 1.3 重启 Sentinel

```bash
./bin/sentinel start -c config/config.yaml
```

#### 1.4 验证数据

访问 VictoriaMetrics UI：http://localhost:8428/vmui

查询数据：
```promql
ping_rtt_ms
```

**优势**：
- ✅ 原生支持 Remote Write
- ✅ 完全兼容 Prometheus 查询
- ✅ 性能更高
- ✅ 资源占用更低
- ✅ 内置 UI

---

### 方案 2: 使用 Prometheus + Pushgateway

如果必须使用 Prometheus，可以通过 Pushgateway 作为中转。

#### 2.1 启动 Pushgateway

```bash
docker run -d \
  --name pushgateway \
  -p 9091:9091 \
  prom/pushgateway
```

#### 2.2 配置 Prometheus 抓取 Pushgateway

编辑 Prometheus 配置文件 `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'pushgateway'
    honor_labels: true
    static_configs:
      - targets: ['localhost:9091']
```

重启 Prometheus：
```bash
docker restart prometheus
```

#### 2.3 修改 Sentinel 发送到 Pushgateway

**注意**：这需要修改 Sentinel 代码，因为 Pushgateway 使用不同的 API。

**不推荐此方案**，因为：
- ❌ 需要额外组件
- ❌ 增加复杂度
- ❌ Pushgateway 不适合高频数据

---

### 方案 3: 使用 Prometheus Agent Mode（Prometheus 2.32+）

Prometheus Agent Mode 可以接收 Remote Write。

#### 3.1 启动 Prometheus Agent

```bash
docker run -d \
  --name prometheus-agent \
  -p 9090:9090 \
  -v /tmp/prometheus-agent.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:latest \
  --config.file=/etc/prometheus/prometheus.yml \
  --enable-feature=agent \
  --web.enable-remote-write-receiver
```

**注意**：需要 `--web.enable-remote-write-receiver` 标志。

#### 3.2 验证

```bash
curl -X POST http://localhost:9090/api/v1/write \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @/dev/null

# 如果返回 400 而不是 404，说明端点已启用
```

---

## 🚀 推荐配置（完整示例）

### 使用 VictoriaMetrics + Grafana

#### docker-compose.yml

```yaml
version: '3.8'

services:
  victoria-metrics:
    image: victoriametrics/victoria-metrics:latest
    container_name: victoria-metrics
    ports:
      - "8428:8428"
    volumes:
      - victoria-data:/victoria-metrics-data
    command:
      - '--storageDataPath=/victoria-metrics-data'
      - '--httpListenAddr=:8428'
    restart: always

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    restart: always

volumes:
  victoria-data:
  grafana-data:
```

#### 启动

```bash
docker-compose up -d
```

#### Sentinel 配置

```yaml
sentinel:
  name: "sentinel-standalone"
  region: "local"

sender:
  mode: "direct"
  flush_interval: 10s
  direct:
    victoria_metrics:
      enabled: true
      url: "http://localhost:8428/api/v1/write"

tasks:
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"
    enabled: true
    config:
      host: "192.168.1.1"
      count: 4
```

#### 配置 Grafana 数据源

1. 访问 Grafana: http://localhost:3000 (admin/admin)
2. 添加数据源 → Prometheus
3. URL: http://victoria-metrics:8428
4. 保存并测试

---

## 🔍 验证数据写入

### 方法 1: 查看 Sentinel 日志

```bash
# 查看发送统计
tail -f logs/sentinel.log | grep "Sender stopped"

# 应该看到：
# {"msg":"Sender stopped","success_count":12,"failed_count":0}
```

### 方法 2: 查询 VictoriaMetrics

```bash
# 查询所有指标
curl 'http://localhost:8428/api/v1/query?query=ping_rtt_ms'

# 查询特定设备
curl 'http://localhost:8428/api/v1/query?query=ping_rtt_ms{device_id="192.168.1.1"}'
```

### 方法 3: 使用 VictoriaMetrics UI

访问：http://localhost:8428/vmui

输入查询：
```promql
ping_rtt_ms
ping_packet_loss
```

---

## 🐛 常见问题

### Q1: VictoriaMetrics 启动失败

**错误**：端口被占用

**解决**：
```bash
# 检查端口占用
lsof -i :8428

# 停止占用进程或使用其他端口
docker run -d -p 8429:8428 victoriametrics/victoria-metrics:latest
```

### Q2: 数据写入成功但查询不到

**原因**：时间戳问题

**检查**：
```bash
# 查看最近 5 分钟的数据
curl 'http://localhost:8428/api/v1/query?query=ping_rtt_ms&time='$(date +%s)
```

### Q3: Sentinel 报错 "connection refused"

**检查**：
```bash
# 1. VictoriaMetrics 是否运行
docker ps | grep victoria

# 2. 端口是否正确
curl http://localhost:8428/health

# 3. URL 配置是否正确
grep "url:" config/config.yaml
```

---

## 📊 性能对比

| 特性 | Prometheus | VictoriaMetrics |
|------|-----------|-----------------|
| Remote Write 接收 | ❌ 需要特殊配置 | ✅ 原生支持 |
| 查询性能 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 存储效率 | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| 内存占用 | 高 | 低 |
| 学习曲线 | 低 | 低（兼容 PromQL） |
| 社区支持 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

---

## 🎯 快速测试脚本

创建测试脚本 `test-remote-write.sh`:

```bash
#!/bin/bash

echo "=== 测试 Remote Write ==="
echo ""

# 1. 启动 VictoriaMetrics
echo "1. 启动 VictoriaMetrics..."
docker run -d --name victoria-test -p 8428:8428 victoriametrics/victoria-metrics:latest
sleep 3

# 2. 测试写入
echo "2. 测试数据写入..."
./bin/sentinel trigger ping 8.8.8.8 -n 1

# 3. 修改配置
echo "3. 修改配置..."
sed -i.bak 's|http://localhost:9090|http://localhost:8428|g' config/config.yaml

# 4. 启动 Sentinel
echo "4. 启动 Sentinel..."
./bin/sentinel start -c config/config.yaml &
SENTINEL_PID=$!
sleep 65

# 5. 查询数据
echo "5. 查询数据..."
curl -s 'http://localhost:8428/api/v1/query?query=ping_rtt_ms' | jq .

# 6. 清理
echo "6. 清理..."
kill $SENTINEL_PID
docker stop victoria-test
docker rm victoria-test

echo ""
echo "=== 测试完成 ==="
```

---

## 📚 相关文档

- [VictoriaMetrics 官方文档](https://docs.victoriametrics.com/)
- [Prometheus Remote Write 规范](https://prometheus.io/docs/concepts/remote_write_spec/)
- [Sentinel 直连发送器指南](DIRECT_SENDER_GUIDE.md)

---

## 💡 总结

**推荐方案**：使用 VictoriaMetrics

```bash
# 1. 启动 VictoriaMetrics
docker run -d -p 8428:8428 --name victoria victoriametrics/victoria-metrics:latest

# 2. 修改配置
vim config/config.yaml
# 将 url 改为: http://localhost:8428/api/v1/write

# 3. 启动 Sentinel
./bin/sentinel start -c config/config.yaml

# 4. 查看数据
open http://localhost:8428/vmui
```

**为什么选择 VictoriaMetrics**：
- ✅ 开箱即用，无需额外配置
- ✅ 完全兼容 Prometheus
- ✅ 性能更好，资源占用更低
- ✅ 内置美观的 UI
- ✅ 适合生产环境

---

**更新日期**: 2025-11-01  
**适用版本**: Sentinel v1.0.0+

