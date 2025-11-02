#!/bin/bash

# Prometheus Remote Write 问题快速修复脚本
# 自动启动 VictoriaMetrics 并更新配置

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo -e "${BLUE}  Prometheus Remote Write 问题修复${NC}"
echo "=========================================="
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: 未安装 Docker${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo -e "${BLUE}步骤 1: 停止现有的 VictoriaMetrics（如果存在）${NC}"
echo "----------------------------------------"
docker stop victoria-metrics 2>/dev/null || true
docker rm victoria-metrics 2>/dev/null || true
echo "✓ 已清理"
echo ""

echo -e "${BLUE}步骤 2: 启动 VictoriaMetrics${NC}"
echo "----------------------------------------"
docker run -d \
  --name victoria-metrics \
  -p 8428:8428 \
  -v victoria-data:/victoria-metrics-data \
  victoriametrics/victoria-metrics:latest

sleep 3

# 检查是否启动成功
if curl -s http://localhost:8428/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ VictoriaMetrics 启动成功！${NC}"
else
    echo -e "${RED}✗ VictoriaMetrics 启动失败${NC}"
    exit 1
fi
echo ""

echo -e "${BLUE}步骤 3: 更新 Sentinel 配置${NC}"
echo "----------------------------------------"

# 备份原配置
if [ -f "config/config.yaml" ]; then
    cp config/config.yaml config/config.yaml.backup
    echo "✓ 已备份原配置: config/config.yaml.backup"
fi

# 如果没有 config.yaml，从模板复制
if [ ! -f "config/config.yaml" ]; then
    cp config/config.local-tasks.yaml config/config.yaml
    echo "✓ 已从模板创建配置文件"
fi

# 更新配置文件
cat > config/config.yaml << 'EOF'
# Orbital Sentinel 配置文件 - 使用 VictoriaMetrics

sentinel:
  id: ""
  name: "sentinel-standalone"
  region: "local"
  labels:
    env: production
    mode: standalone

core:
  url: ""
  api_token: ""

heartbeat:
  interval: 30s
  timeout: 10s
  retry_times: 3

collector:
  worker_pool_size: 10
  task_fetch_interval: 60s
  max_execution_time: 300s

buffer:
  type: "memory"
  size: 10000
  flush_interval: 10s
  disk_path: "./data/buffer"

sender:
  mode: "direct"
  batch_size: 1000
  flush_interval: 10s
  timeout: 30s
  retry_times: 3
  retry_interval: 5s
  
  direct:
    prometheus:
      enabled: false
    
    victoria_metrics:
      enabled: true
      url: "http://localhost:8428/api/v1/write"
      username: ""
      password: ""
    
    clickhouse:
      enabled: false

plugins:
  directory: "./plugins"
  auto_reload: false
  reload_interval: 300s

logging:
  level: info
  format: json
  output: both
  file_path: "./logs/sentinel.log"
  max_size: 100
  max_backups: 7
  max_age: 30

# 本地任务配置
tasks:
  - id: "ping-gateway"
    device_id: "192.168.1.1"
    plugin: "ping"
    interval: "60s"
    timeout: "10s"
    enabled: true
    config:
      host: "192.168.1.1"
      count: 4
      interval: "1s"
      timeout: "5s"

  - id: "ping-google-dns"
    device_id: "8.8.8.8"
    plugin: "ping"
    interval: "300s"
    timeout: "10s"
    enabled: true
    config:
      host: "8.8.8.8"
      count: 4
      interval: "1s"
      timeout: "5s"

  - id: "ping-aliyun-dns"
    device_id: "223.5.5.5"
    plugin: "ping"
    interval: "300s"
    timeout: "10s"
    enabled: true
    config:
      host: "223.5.5.5"
      count: 4
      interval: "1s"
      timeout: "5s"

  - id: "ping-cloudflare-dns"
    device_id: "1.1.1.1"
    plugin: "ping"
    interval: "300s"
    timeout: "10s"
    enabled: true
    config:
      host: "1.1.1.1"
      count: 4
      interval: "1s"
      timeout: "5s"
EOF

echo -e "${GREEN}✓ 配置已更新${NC}"
echo ""

echo -e "${BLUE}步骤 4: 验证配置${NC}"
echo "----------------------------------------"
echo "VictoriaMetrics URL: $(grep -A 2 'victoria_metrics:' config/config.yaml | grep 'url:' | awk '{print $2}')"
echo "任务数量: $(grep -c '^  - id:' config/config.yaml)"
echo ""

echo "=========================================="
echo -e "${GREEN}  修复完成！${NC}"
echo "=========================================="
echo ""

echo "📊 服务信息:"
echo "  - VictoriaMetrics UI: http://localhost:8428/vmui"
echo "  - VictoriaMetrics API: http://localhost:8428/api/v1/write"
echo "  - 健康检查: http://localhost:8428/health"
echo ""

echo "🚀 下一步:"
echo "  1. 启动 Sentinel:"
echo "     ./bin/sentinel start -c config/config.yaml"
echo ""
echo "  2. 查看日志:"
echo "     tail -f logs/sentinel.log"
echo ""
echo "  3. 等待 60 秒后查询数据:"
echo "     curl 'http://localhost:8428/api/v1/query?query=ping_rtt_ms'"
echo ""
echo "  4. 或访问 Web UI:"
echo "     open http://localhost:8428/vmui"
echo ""

echo "💡 提示:"
echo "  - 原配置已备份到: config/config.yaml.backup"
echo "  - VictoriaMetrics 数据存储在 Docker volume: victoria-data"
echo "  - 完全兼容 Prometheus 查询语法"
echo ""

echo "📚 相关文档:"
echo "  - docs/PROMETHEUS_REMOTE_WRITE_ISSUE.md"
echo "  - docs/DIRECT_SENDER_GUIDE.md"
echo ""

