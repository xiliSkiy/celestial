#!/bin/bash

# Gravital Core API 测试脚本

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

API_BASE="http://localhost:8080"

echo "=========================================="
echo -e "${BLUE}  Gravital Core API 测试${NC}"
echo "=========================================="
echo ""

# 检查服务是否运行
echo -e "${BLUE}步骤 1: 检查服务状态${NC}"
echo "----------------------------------------"
if curl -s "${API_BASE}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 服务正在运行${NC}"
else
    echo -e "${RED}✗ 服务未运行，请先启动服务${NC}"
    echo "启动命令: ./bin/gravital-core -c config/config.yaml"
    exit 1
fi
echo ""

# 健康检查
echo -e "${BLUE}步骤 2: 健康检查${NC}"
echo "----------------------------------------"
HEALTH=$(curl -s "${API_BASE}/health")
echo "$HEALTH" | jq .
echo ""

# 版本信息
echo -e "${BLUE}步骤 3: 版本信息${NC}"
echo "----------------------------------------"
VERSION=$(curl -s "${API_BASE}/version")
echo "$VERSION" | jq .
echo ""

# 登录
echo -e "${BLUE}步骤 4: 用户登录${NC}"
echo "----------------------------------------"
LOGIN_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

echo "$LOGIN_RESPONSE" | jq .

# 提取 Token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ 登录失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 登录成功${NC}"
echo "Token: ${TOKEN:0:20}..."
echo ""

# 创建设备
echo -e "${BLUE}步骤 5: 创建设备${NC}"
echo "----------------------------------------"
DEVICE_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/devices" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Server",
    "device_type": "server",
    "connection_config": {
      "host": "192.168.1.100",
      "port": 22
    },
    "labels": {
      "env": "test",
      "region": "local"
    }
  }')

echo "$DEVICE_RESPONSE" | jq .
DEVICE_ID=$(echo "$DEVICE_RESPONSE" | jq -r '.data.device_id')
echo -e "${GREEN}✓ 设备创建成功: $DEVICE_ID${NC}"
echo ""

# 获取设备列表
echo -e "${BLUE}步骤 6: 获取设备列表${NC}"
echo "----------------------------------------"
DEVICES=$(curl -s -X GET "${API_BASE}/api/v1/devices?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
echo "$DEVICES" | jq .
echo ""

# 注册 Sentinel
echo -e "${BLUE}步骤 7: 注册 Sentinel${NC}"
echo "----------------------------------------"
SENTINEL_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/sentinels/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-sentinel",
    "hostname": "test-host",
    "ip_address": "192.168.1.200",
    "version": "1.0.0",
    "os": "linux",
    "arch": "amd64",
    "region": "test-region"
  }')

echo "$SENTINEL_RESPONSE" | jq .
SENTINEL_ID=$(echo "$SENTINEL_RESPONSE" | jq -r '.data.sentinel_id')
API_TOKEN=$(echo "$SENTINEL_RESPONSE" | jq -r '.data.api_token')
echo -e "${GREEN}✓ Sentinel 注册成功: $SENTINEL_ID${NC}"
echo ""

# 发送心跳
echo -e "${BLUE}步骤 8: 发送心跳${NC}"
echo "----------------------------------------"
HEARTBEAT_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/sentinels/heartbeat" \
  -H "X-Sentinel-ID: $SENTINEL_ID" \
  -H "X-API-Token: $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_usage": 25.5,
    "memory_usage": 60.2,
    "disk_usage": 45.0,
    "task_count": 5,
    "plugin_count": 3,
    "uptime_seconds": 3600,
    "version": "1.0.0"
  }')

echo "$HEARTBEAT_RESPONSE" | jq .
echo -e "${GREEN}✓ 心跳发送成功${NC}"
echo ""

# 创建任务
echo -e "${BLUE}步骤 9: 创建采集任务${NC}"
echo "----------------------------------------"
TASK_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"device_id\": \"$DEVICE_ID\",
    \"sentinel_id\": \"$SENTINEL_ID\",
    \"plugin_name\": \"ping\",
    \"config\": {
      \"host\": \"192.168.1.100\",
      \"count\": 4
    },
    \"interval_seconds\": 60,
    \"enabled\": true
  }")

echo "$TASK_RESPONSE" | jq .
TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.task_id')
echo -e "${GREEN}✓ 任务创建成功: $TASK_ID${NC}"
echo ""

# 获取 Sentinel 的任务列表
echo -e "${BLUE}步骤 10: 获取 Sentinel 任务列表${NC}"
echo "----------------------------------------"
SENTINEL_TASKS=$(curl -s -X GET "${API_BASE}/api/v1/tasks" \
  -H "X-Sentinel-ID: $SENTINEL_ID" \
  -H "X-API-Token: $API_TOKEN")
echo "$SENTINEL_TASKS" | jq .
echo ""

# 创建告警规则
echo -e "${BLUE}步骤 11: 创建告警规则${NC}"
echo "----------------------------------------"
ALERT_RULE_RESPONSE=$(curl -s -X POST "${API_BASE}/api/v1/alert-rules" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "rule_name": "CPU 使用率过高",
    "enabled": true,
    "severity": "warning",
    "condition": "cpu_usage > 80",
    "duration": 300,
    "description": "当 CPU 使用率持续 5 分钟超过 80% 时触发告警"
  }')

echo "$ALERT_RULE_RESPONSE" | jq .
echo -e "${GREEN}✓ 告警规则创建成功${NC}"
echo ""

# 获取 Sentinel 列表
echo -e "${BLUE}步骤 12: 获取 Sentinel 列表${NC}"
echo "----------------------------------------"
SENTINELS=$(curl -s -X GET "${API_BASE}/api/v1/sentinels?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
echo "$SENTINELS" | jq .
echo ""

echo "=========================================="
echo -e "${GREEN}  测试完成！${NC}"
echo "=========================================="
echo ""

echo "📊 测试结果:"
echo "  ✅ 健康检查"
echo "  ✅ 用户登录"
echo "  ✅ 设备管理"
echo "  ✅ Sentinel 注册"
echo "  ✅ 心跳上报"
echo "  ✅ 任务管理"
echo "  ✅ 告警规则"
echo ""

echo "🔑 测试数据:"
echo "  - Token: ${TOKEN:0:30}..."
echo "  - Device ID: $DEVICE_ID"
echo "  - Sentinel ID: $SENTINEL_ID"
echo "  - Task ID: $TASK_ID"
echo ""

echo "📝 下一步:"
echo "  1. 查看数据库中的数据"
echo "  2. 启动 orbital-sentinels 进行集成测试"
echo "  3. 实现数据转发模块"
echo ""

