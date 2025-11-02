#!/bin/bash

# 本地任务功能演示脚本
# 用途：演示无中心端场景下的本地任务配置功能

set -e

echo "=========================================="
echo "  Orbital Sentinel - 本地任务功能演示"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查二进制文件
if [ ! -f "./bin/sentinel" ]; then
    echo -e "${YELLOW}未找到 sentinel 二进制文件，正在构建...${NC}"
    make build
    echo ""
fi

# 创建日志目录
mkdir -p logs

echo -e "${BLUE}步骤 1: 检查配置文件${NC}"
echo "----------------------------------------"
if [ ! -f "config/config.local-tasks.yaml" ]; then
    echo -e "${YELLOW}错误: 配置文件不存在${NC}"
    exit 1
fi

echo "✓ 配置文件存在: config/config.local-tasks.yaml"
echo ""

echo -e "${BLUE}步骤 2: 查看配置的任务${NC}"
echo "----------------------------------------"
echo "配置的任务列表:"
grep -A 3 "^  - id:" config/config.local-tasks.yaml | grep -E "(id:|device_id:|interval:|enabled:)" | head -20
echo ""

echo -e "${BLUE}步骤 3: 启动 Sentinel${NC}"
echo "----------------------------------------"
echo "启动命令: ./bin/sentinel start -c config/config.local-tasks.yaml"
echo ""

# 清空日志
> logs/sentinel.log

# 后台启动
./bin/sentinel start -c config/config.local-tasks.yaml > /dev/null 2>&1 &
SENTINEL_PID=$!

echo "✓ Sentinel 已启动 (PID: $SENTINEL_PID)"
echo ""

# 等待启动
echo -e "${BLUE}步骤 4: 等待初始化 (3秒)${NC}"
echo "----------------------------------------"
sleep 3

echo -e "${BLUE}步骤 5: 检查启动日志${NC}"
echo "----------------------------------------"
echo ""

echo "1. 插件注册:"
grep "Registered builtin plugin" logs/sentinel.log | tail -1
echo ""

echo "2. 任务加载统计:"
grep "Local tasks loaded" logs/sentinel.log | tail -1
echo ""

echo "3. 已加载的任务:"
grep "Loaded local task" logs/sentinel.log | while read line; do
    echo "   $line"
done
echo ""

echo "4. 组件启动状态:"
grep "All components started" logs/sentinel.log | tail -1
echo ""

echo -e "${BLUE}步骤 6: 等待任务执行 (65秒)${NC}"
echo "----------------------------------------"
echo "等待第一个任务执行（interval: 60s）..."
echo ""

# 显示进度条
for i in {1..65}; do
    printf "\r进度: [%-65s] %d/65秒" $(printf '#%.0s' $(seq 1 $i)) $i
    sleep 1
done
echo ""
echo ""

echo -e "${BLUE}步骤 7: 停止 Sentinel${NC}"
echo "----------------------------------------"
kill $SENTINEL_PID 2>/dev/null || true
sleep 2
echo "✓ Sentinel 已停止"
echo ""

echo -e "${BLUE}步骤 8: 查看执行结果${NC}"
echo "----------------------------------------"
echo ""

echo "1. 发送统计:"
grep "Sender stopped" logs/sentinel.log | tail -1
echo ""

echo "2. 最后 10 条日志:"
tail -10 logs/sentinel.log | while read line; do
    echo "   $line"
done
echo ""

echo "=========================================="
echo -e "${GREEN}  演示完成！${NC}"
echo "=========================================="
echo ""

echo "📊 结果总结:"
echo ""

# 提取统计信息
SUCCESS_COUNT=$(grep "Sender stopped" logs/sentinel.log | tail -1 | grep -o '"success_count":[0-9]*' | cut -d: -f2)
FAILED_COUNT=$(grep "Sender stopped" logs/sentinel.log | tail -1 | grep -o '"failed_count":[0-9]*' | cut -d: -f2)
TASKS_LOADED=$(grep "Local tasks loaded" logs/sentinel.log | tail -1 | grep -o '"success":[0-9]*' | cut -d: -f2)

if [ -n "$SUCCESS_COUNT" ] && [ "$SUCCESS_COUNT" -gt 0 ]; then
    echo -e "  ${GREEN}✓${NC} 任务加载成功: $TASKS_LOADED 个"
    echo -e "  ${GREEN}✓${NC} 数据采集成功: $SUCCESS_COUNT 个指标"
    echo -e "  ${GREEN}✓${NC} 发送失败: $FAILED_COUNT 个 (Prometheus 未运行)"
    echo ""
    echo -e "${GREEN}本地任务功能运行正常！${NC}"
else
    echo -e "  ${YELLOW}⚠${NC} 未检测到数据采集"
    echo "  请检查日志: tail -f logs/sentinel.log"
fi

echo ""
echo "📚 相关文档:"
echo "  - 本地任务配置指南: docs/LOCAL_TASKS_GUIDE.md"
echo "  - 快速开始: QUICKSTART.md"
echo "  - 功能总结: LOCAL_TASKS_SUMMARY.md"
echo ""

echo "🔧 下一步:"
echo "  1. 修改配置: vim config/config.local-tasks.yaml"
echo "  2. 启动 Prometheus: docker run -p 9090:9090 prom/prometheus"
echo "  3. 重新启动: ./bin/sentinel start -c config/config.local-tasks.yaml"
echo "  4. 查看数据: http://localhost:9090"
echo ""

