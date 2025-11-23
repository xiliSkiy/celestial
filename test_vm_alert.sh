#!/bin/bash

# 测试 VictoriaMetrics 告警查询功能

echo "=========================================="
echo "测试 VictoriaMetrics 告警查询功能"
echo "=========================================="

# 配置
GRAVITAL_URL="http://localhost:8080"
VM_URL="http://localhost:8428"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "1. 检查 VictoriaMetrics 健康状态"
echo "----------------------------------------"
VM_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" ${VM_URL}/health)
if [ "$VM_HEALTH" == "200" ]; then
    echo -e "${GREEN}✓ VictoriaMetrics 运行正常${NC}"
else
    echo -e "${RED}✗ VictoriaMetrics 不可用 (HTTP $VM_HEALTH)${NC}"
    echo -e "${YELLOW}请确保 VictoriaMetrics 正在运行：docker-compose up -d victoriametrics${NC}"
    exit 1
fi

echo ""
echo "2. 检查时序数据库中的 device_status 指标"
echo "----------------------------------------"
QUERY="device_status"
VM_RESPONSE=$(curl -s "${VM_URL}/api/v1/query?query=${QUERY}")
RESULT_COUNT=$(echo $VM_RESPONSE | jq '.data.result | length')

echo "查询: $QUERY"
echo "结果数量: $RESULT_COUNT"

if [ "$RESULT_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 找到 $RESULT_COUNT 个时序数据${NC}"
    echo ""
    echo "示例数据:"
    echo $VM_RESPONSE | jq '.data.result[0:3]'
else
    echo -e "${YELLOW}⚠ 未找到 device_status 指标数据${NC}"
    echo -e "${YELLOW}请确保 Sentinel 正在采集并上报数据${NC}"
fi

echo ""
echo "3. 检查特定设备的 device_status"
echo "----------------------------------------"
# 获取第一个设备 ID
DEVICE_ID=$(echo $VM_RESPONSE | jq -r '.data.result[0].metric.device_id // empty')

if [ -n "$DEVICE_ID" ]; then
    echo "查询设备: $DEVICE_ID"
    DEVICE_QUERY="device_status{device_id=\"${DEVICE_ID}\"}"
    DEVICE_RESPONSE=$(curl -s "${VM_URL}/api/v1/query?query=${DEVICE_QUERY}")
    
    echo "查询: $DEVICE_QUERY"
    echo "结果:"
    echo $DEVICE_RESPONSE | jq '.data.result[0]'
    
    DEVICE_VALUE=$(echo $DEVICE_RESPONSE | jq -r '.data.result[0].value[1]')
    if [ "$DEVICE_VALUE" == "1" ]; then
        echo -e "${GREEN}✓ 设备在线 (value = $DEVICE_VALUE)${NC}"
    else
        echo -e "${RED}✗ 设备离线 (value = $DEVICE_VALUE)${NC}"
    fi
else
    echo -e "${YELLOW}⚠ 无法获取设备 ID${NC}"
fi

echo ""
echo "4. 检查告警规则"
echo "----------------------------------------"
# 需要先登录获取 token
echo "登录..."
LOGIN_RESPONSE=$(curl -s -X POST "${GRAVITAL_URL}/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ 登录失败${NC}"
    echo "响应: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ 登录成功${NC}"

# 获取告警规则列表
echo ""
echo "获取告警规则列表..."
RULES_RESPONSE=$(curl -s -X GET "${GRAVITAL_URL}/api/v1/alert-rules" \
    -H "Authorization: Bearer ${TOKEN}")

RULES_COUNT=$(echo $RULES_RESPONSE | jq '.data.items | length')
echo "告警规则数量: $RULES_COUNT"

if [ "$RULES_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 找到 $RULES_COUNT 个告警规则${NC}"
    echo ""
    echo "规则列表:"
    echo $RULES_RESPONSE | jq '.data.items[] | {id, rule_name, enabled, condition}'
else
    echo -e "${YELLOW}⚠ 未找到告警规则${NC}"
    echo -e "${YELLOW}提示：可以通过前端创建告警规则${NC}"
fi

echo ""
echo "5. 检查告警事件"
echo "----------------------------------------"
EVENTS_RESPONSE=$(curl -s -X GET "${GRAVITAL_URL}/api/v1/alert-events?page=1&page_size=10" \
    -H "Authorization: Bearer ${TOKEN}")

EVENTS_COUNT=$(echo $EVENTS_RESPONSE | jq '.data.total')
echo "告警事件总数: $EVENTS_COUNT"

if [ "$EVENTS_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 找到 $EVENTS_COUNT 个告警事件${NC}"
    echo ""
    echo "最近的告警事件:"
    echo $EVENTS_RESPONSE | jq '.data.items[0:3] | .[] | {id, rule_id, device_id, status, message, triggered_at}'
else
    echo -e "${YELLOW}⚠ 未找到告警事件${NC}"
    echo -e "${YELLOW}提示：告警引擎每 30 秒评估一次规则${NC}"
fi

echo ""
echo "6. 检查告警聚合"
echo "----------------------------------------"
AGG_RESPONSE=$(curl -s -X GET "${GRAVITAL_URL}/api/v1/alert-aggregations" \
    -H "Authorization: Bearer ${TOKEN}")

AGG_COUNT=$(echo $AGG_RESPONSE | jq '.data | length')
echo "活跃告警聚合数量: $AGG_COUNT"

if [ "$AGG_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 找到 $AGG_COUNT 个活跃告警聚合${NC}"
    echo ""
    echo "告警聚合:"
    echo $AGG_RESPONSE | jq '.data[] | {rule_name, severity, total_count, firing_count, acked_count}'
else
    echo -e "${YELLOW}⚠ 未找到活跃告警聚合${NC}"
    echo -e "${YELLOW}提示：只有 firing 或 acknowledged 状态的告警才会显示在聚合中${NC}"
fi

echo ""
echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""
echo "总结:"
echo "  - VictoriaMetrics: $([ "$VM_HEALTH" == "200" ] && echo -e "${GREEN}正常${NC}" || echo -e "${RED}异常${NC}")"
echo "  - 时序数据: $([ "$RESULT_COUNT" -gt 0 ] && echo -e "${GREEN}$RESULT_COUNT 条${NC}" || echo -e "${YELLOW}无数据${NC}")"
echo "  - 告警规则: $([ "$RULES_COUNT" -gt 0 ] && echo -e "${GREEN}$RULES_COUNT 个${NC}" || echo -e "${YELLOW}未配置${NC}")"
echo "  - 告警事件: $([ "$EVENTS_COUNT" -gt 0 ] && echo -e "${GREEN}$EVENTS_COUNT 个${NC}" || echo -e "${YELLOW}无事件${NC}")"
echo "  - 活跃告警: $([ "$AGG_COUNT" -gt 0 ] && echo -e "${GREEN}$AGG_COUNT 个${NC}" || echo -e "${YELLOW}无活跃告警${NC}")"
echo ""

# 提供下一步建议
if [ "$RESULT_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}建议：${NC}"
    echo "  1. 确保 Sentinel 正在运行并采集数据"
    echo "  2. 检查 Sentinel 的转发配置是否正确"
    echo "  3. 查看 Sentinel 日志：docker-compose logs -f sentinel"
fi

if [ "$RULES_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}建议：${NC}"
    echo "  1. 通过前端创建告警规则"
    echo "  2. 或使用 API 创建规则："
    echo "     curl -X POST ${GRAVITAL_URL}/api/v1/alert-rules \\"
    echo "       -H 'Authorization: Bearer ${TOKEN}' \\"
    echo "       -H 'Content-Type: application/json' \\"
    echo "       -d '{\"rule_name\":\"设备离线告警\",\"severity\":\"critical\",\"condition\":\"device_status != 1\",\"duration\":300,\"enabled\":true}'"
fi

if [ "$EVENTS_COUNT" -eq 0 ] && [ "$RULES_COUNT" -gt 0 ]; then
    echo -e "${YELLOW}建议：${NC}"
    echo "  1. 等待告警引擎评估（每 30 秒一次）"
    echo "  2. 检查告警规则的条件是否正确"
    echo "  3. 查看 Gravital Core 日志：docker-compose logs -f gravital-core"
fi

echo ""

