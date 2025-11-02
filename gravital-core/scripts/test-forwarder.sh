#!/bin/bash

# 数据转发模块测试脚本

set -e

BASE_URL="http://localhost:8080/api/v1"
SENTINEL_ID="test-sentinel-001"
API_TOKEN="test-token"

echo "================================"
echo "数据转发模块测试脚本"
echo "================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_api() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local headers=$5
    
    echo -e "${YELLOW}测试: ${name}${NC}"
    
    if [ -z "$data" ]; then
        response=$(curl -s -X ${method} "${BASE_URL}${endpoint}" ${headers})
    else
        response=$(curl -s -X ${method} "${BASE_URL}${endpoint}" \
            -H "Content-Type: application/json" \
            ${headers} \
            -d "${data}")
    fi
    
    echo "响应: $response"
    
    # 检查响应是否包含 "code":0
    if echo "$response" | grep -q '"code":0'; then
        echo -e "${GREEN}✓ 测试通过${NC}"
    else
        echo -e "${RED}✗ 测试失败${NC}"
    fi
    echo ""
}

# 1. 测试数据接收接口
echo "=== 1. 测试数据接收 ==="
test_api "接收指标数据" "POST" "/data/ingest" \
'{
  "metrics": [
    {
      "name": "cpu_usage",
      "value": 85.5,
      "type": "gauge",
      "labels": {
        "device_id": "server-001",
        "host": "test-server"
      },
      "timestamp": '$(date +%s)'
    },
    {
      "name": "memory_usage",
      "value": 70.2,
      "type": "gauge",
      "labels": {
        "device_id": "server-001",
        "host": "test-server"
      },
      "timestamp": '$(date +%s)'
    }
  ]
}' \
"-H 'X-Sentinel-ID: ${SENTINEL_ID}'"

# 2. 测试转发器列表
echo "=== 2. 测试转发器管理 ==="
test_api "列出所有转发器" "GET" "/forwarders"

# 3. 测试创建转发器
echo "=== 3. 测试创建转发器 ==="
test_api "创建 VictoriaMetrics 转发器" "POST" "/forwarders" \
'{
  "name": "victoria-test",
  "type": "victoria-metrics",
  "enabled": true,
  "endpoint": "http://victoria:8428/api/v1/write",
  "timeout_seconds": 30,
  "batch_size": 5000,
  "flush_interval": 10,
  "retry_times": 3
}'

# 4. 测试获取转发器详情
echo "=== 4. 测试获取转发器详情 ==="
test_api "获取转发器详情" "GET" "/forwarders/victoria-test"

# 5. 测试更新转发器
echo "=== 5. 测试更新转发器 ==="
test_api "更新转发器配置" "PUT" "/forwarders/victoria-test" \
'{
  "name": "victoria-test",
  "type": "victoria-metrics",
  "enabled": false,
  "endpoint": "http://victoria:8428/api/v1/write",
  "timeout_seconds": 60,
  "batch_size": 10000
}'

# 6. 测试统计信息
echo "=== 6. 测试统计信息 ==="
test_api "获取所有转发器统计" "GET" "/forwarders/stats"

# 7. 测试删除转发器
echo "=== 7. 测试删除转发器 ==="
test_api "删除转发器" "DELETE" "/forwarders/victoria-test"

# 8. 批量发送数据测试
echo "=== 8. 批量发送数据测试 ==="
echo "发送 100 条指标数据..."

metrics='['
for i in {1..100}; do
    if [ $i -gt 1 ]; then
        metrics+=','
    fi
    metrics+='{
      "name": "test_metric_'$i'",
      "value": '$((RANDOM % 100))',
      "type": "gauge",
      "labels": {
        "device_id": "server-001",
        "index": "'$i'"
      },
      "timestamp": '$(date +%s)'
    }'
done
metrics+=']'

test_api "批量发送 100 条数据" "POST" "/data/ingest" \
"{\"metrics\": ${metrics}}" \
"-H 'X-Sentinel-ID: ${SENTINEL_ID}'"

# 9. 再次查看统计信息
echo "=== 9. 查看更新后的统计信息 ==="
test_api "获取所有转发器统计" "GET" "/forwarders/stats"

echo ""
echo "================================"
echo "测试完成！"
echo "================================"
echo ""
echo "提示："
echo "1. 确保 Gravital Core 服务正在运行"
echo "2. 确保已配置至少一个转发器"
echo "3. 查看日志了解详细信息: tail -f logs/gravital-core.log"
echo ""

