# Orbital Sentinels 示例

本目录包含了 Orbital Sentinels 的完整部署示例，展示如何使用直连模式将数据发送到 Prometheus、VictoriaMetrics 和 ClickHouse。

## 📦 包含的服务

| 服务 | 端口 | 说明 |
|------|------|------|
| Sentinel | - | 数据采集端 |
| Prometheus | 9090 | 时序数据库 + 查询界面 |
| VictoriaMetrics | 8428 | 高性能时序数据库 |
| ClickHouse | 8123, 9000 | 列式数据库 |
| Grafana | 3000 | 数据可视化 |

## 🚀 快速开始

### 1. 准备环境

确保已安装：
- Docker 20.10+
- Docker Compose 2.0+

### 2. 配置环境变量

创建 `.env` 文件：

```bash
# API Token（如果使用中心端）
API_TOKEN=your-api-token-here

# ClickHouse 密码
CH_PASSWORD=clickhouse-password

# Grafana 管理员密码
GRAFANA_PASSWORD=admin
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f sentinel

# 查看服务状态
docker-compose ps
```

### 4. 访问服务

- **Prometheus**: http://localhost:9090
- **VictoriaMetrics**: http://localhost:8428
- **ClickHouse**: http://localhost:8123/play
- **Grafana**: http://localhost:3000 (admin/admin)

## 📊 验证数据

### Prometheus

访问 http://localhost:9090/graph，执行查询：

```promql
# 查看所有指标
{__name__=~".+"}

# 查看 Ping 指标
ping_rtt_ms

# 查看特定主机
ping_rtt_ms{host="8.8.8.8"}
```

### VictoriaMetrics

访问 http://localhost:8428/vmui，执行相同的 PromQL 查询。

### ClickHouse

访问 http://localhost:8123/play，执行 SQL 查询：

```sql
-- 查看最近的数据
SELECT 
    timestamp,
    metric_name,
    metric_value,
    device_id,
    labels
FROM metrics
ORDER BY timestamp DESC
LIMIT 10;

-- 统计指标数量
SELECT 
    metric_name,
    count() as count,
    avg(metric_value) as avg_value,
    max(metric_value) as max_value
FROM metrics
WHERE timestamp >= now() - INTERVAL 1 HOUR
GROUP BY metric_name
ORDER BY count DESC;

-- 查看设备列表
SELECT DISTINCT device_id
FROM metrics
ORDER BY device_id;
```

### Grafana

1. 访问 http://localhost:3000
2. 登录（admin/admin）
3. 添加数据源：
   - Prometheus: http://prometheus:9090
   - VictoriaMetrics: http://victoria-metrics:8428
   - ClickHouse: clickhouse:9000
4. 创建仪表板

## 🔧 配置说明

### 切换发送模式

编辑 `config.yaml`：

```yaml
sender:
  mode: "direct"  # 直连模式
  # mode: "core"    # 中心端模式
  # mode: "hybrid"  # 混合模式
```

### 启用/禁用特定数据库

```yaml
sender:
  direct:
    prometheus:
      enabled: true  # 启用 Prometheus
    
    victoria_metrics:
      enabled: false  # 禁用 VictoriaMetrics
    
    clickhouse:
      enabled: true  # 启用 ClickHouse
```

### 性能调优

```yaml
sender:
  batch_size: 5000        # 增加批量大小
  flush_interval: 5s      # 减少刷新间隔

buffer:
  size: 50000             # 增加缓冲区大小

collector:
  worker_pool_size: 20    # 增加并发数
```

## 📝 添加采集任务

Sentinel 会定期从中心端拉取任务。在直连模式下，你可以手动配置任务（未来版本支持）。

当前可以通过修改插件配置来添加采集目标。例如，编辑 `plugins/ping/plugin.yaml`。

## 🐛 故障排查

### 查看 Sentinel 日志

```bash
docker-compose logs -f sentinel
```

### 查看数据库日志

```bash
# Prometheus
docker-compose logs -f prometheus

# VictoriaMetrics
docker-compose logs -f victoria-metrics

# ClickHouse
docker-compose logs -f clickhouse
```

### 检查连接

```bash
# 进入 Sentinel 容器
docker-compose exec sentinel sh

# 测试 Prometheus 连接
wget -O- http://prometheus:9090/-/healthy

# 测试 VictoriaMetrics 连接
wget -O- http://victoria-metrics:8428/health

# 测试 ClickHouse 连接
wget -O- http://clickhouse:8123/ping
```

### 重启服务

```bash
# 重启 Sentinel
docker-compose restart sentinel

# 重启所有服务
docker-compose restart

# 重新构建并启动
docker-compose up -d --build
```

## 🧹 清理

### 停止服务

```bash
docker-compose down
```

### 删除数据

```bash
# 停止并删除所有数据
docker-compose down -v
```

## 📚 进一步阅读

- [直连发送器指南](../docs/DIRECT_SENDER_GUIDE.md)
- [插件开发指南](../plugins/README.md)
- [配置参考](../config/config.example.yaml)

## 💡 提示

1. **开发环境**: 使用 `direct` 模式，简化部署
2. **生产环境**: 使用 `core` 或 `hybrid` 模式
3. **数据保留**: 
   - Prometheus: 默认 15 天
   - VictoriaMetrics: 配置为 12 个月
   - ClickHouse: TTL 设置为 90 天
4. **资源限制**: 生产环境建议为每个服务设置资源限制
5. **备份**: 定期备份数据卷

## 🤝 贡献

欢迎提交问题和改进建议！

