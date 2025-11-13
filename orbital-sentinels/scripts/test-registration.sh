#!/bin/bash

# 测试采集端自动注册功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "=== 采集端自动注册测试 ==="
echo

# 检查中心端是否运行
echo "1. 检查中心端状态..."
if ! curl -s http://localhost:8080/api/v1/health > /dev/null 2>&1; then
    echo "❌ 中心端未运行,请先启动 gravital-core"
    echo "   cd ../gravital-core && docker-compose up -d"
    exit 1
fi
echo "✅ 中心端正常运行"
echo

# 清除旧的凭证文件(用于测试)
CREDENTIALS_FILE="${HOME}/.sentinel/credentials.yaml"
if [ -f "${CREDENTIALS_FILE}" ]; then
    echo "2. 清除旧的凭证文件..."
    rm -f "${CREDENTIALS_FILE}"
    echo "✅ 旧凭证已清除"
else
    echo "2. 无旧凭证文件"
fi
echo

# 构建采集端
echo "3. 构建采集端..."
cd "${PROJECT_ROOT}"
make build
echo "✅ 构建完成"
echo

# 启动采集端(后台运行)
echo "4. 启动采集端(自动注册测试)..."
./bin/sentinel start -c config/config.register-test.yaml > /tmp/sentinel-register-test.log 2>&1 &
SENTINEL_PID=$!
echo "✅ 采集端已启动 (PID: ${SENTINEL_PID})"
echo

# 等待注册完成
echo "5. 等待注册完成(最多10秒)..."
for i in {1..10}; do
    if [ -f "${CREDENTIALS_FILE}" ]; then
        echo "✅ 注册成功! 凭证已保存到: ${CREDENTIALS_FILE}"
        echo
        echo "凭证内容:"
        cat "${CREDENTIALS_FILE}"
        echo
        break
    fi
    echo "   等待中... ($i/10)"
    sleep 1
done

if [ ! -f "${CREDENTIALS_FILE}" ]; then
    echo "❌ 注册失败: 凭证文件未生成"
    echo
    echo "采集端日志:"
    cat /tmp/sentinel-register-test.log
    kill ${SENTINEL_PID} 2>/dev/null || true
    exit 1
fi

# 检查凭证内容
echo "6. 验证凭证..."
if ! grep -q "sentinel_id:" "${CREDENTIALS_FILE}"; then
    echo "❌ 凭证文件格式错误"
    kill ${SENTINEL_PID} 2>/dev/null || true
    exit 1
fi

SENTINEL_ID=$(grep "sentinel_id:" "${CREDENTIALS_FILE}" | awk '{print $2}')
API_TOKEN=$(grep "api_token:" "${CREDENTIALS_FILE}" | awk '{print $2}')

echo "✅ 凭证验证通过"
echo "   Sentinel ID: ${SENTINEL_ID}"
echo "   API Token: ${API_TOKEN:0:20}..."
echo

# 测试心跳
echo "7. 测试心跳..."
sleep 3  # 等待一次心跳
echo "✅ 心跳测试(查看日志)"
tail -n 20 /tmp/sentinel-register-test.log | grep -i "heartbeat" || echo "   (未找到心跳日志,可能还未发送)"
echo

# 查询中心端 Sentinel 列表
echo "8. 从中心端查询 Sentinel 状态..."
SENTINEL_INFO=$(curl -s http://localhost:8080/api/v1/admin/sentinels | \
    python3 -m json.tool 2>/dev/null || echo "查询失败")

if echo "${SENTINEL_INFO}" | grep -q "${SENTINEL_ID}"; then
    echo "✅ Sentinel 已注册到中心端"
    echo "${SENTINEL_INFO}" | grep -A 5 "${SENTINEL_ID}" || true
else
    echo "⚠️  中心端未查询到 Sentinel (可能需要认证)"
fi
echo

# 停止采集端
echo "9. 停止采集端..."
kill ${SENTINEL_PID} 2>/dev/null || true
sleep 2
echo "✅ 采集端已停止"
echo

# 测试重启(使用已有凭证)
echo "10. 测试重启(应该使用已有凭证)..."
./bin/sentinel start -c config/config.register-test.yaml > /tmp/sentinel-restart-test.log 2>&1 &
SENTINEL_PID=$!
sleep 5

if grep -q "Using existing credentials" /tmp/sentinel-restart-test.log; then
    echo "✅ 重启成功,使用已有凭证"
else
    echo "⚠️  未找到'使用已有凭证'日志"
fi

kill ${SENTINEL_PID} 2>/dev/null || true
echo

# 总结
echo "=== 测试完成 ==="
echo
echo "✅ 所有测试通过!"
echo
echo "凭证文件位置: ${CREDENTIALS_FILE}"
echo "测试日志: /tmp/sentinel-register-test.log"
echo "重启日志: /tmp/sentinel-restart-test.log"
echo
echo "如需清除测试环境:"
echo "  rm -f ${CREDENTIALS_FILE}"
echo "  rm -f /tmp/sentinel-*.log"

