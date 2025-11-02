#!/bin/bash

# Orbital Sentinels 配置检查脚本
# 用于检查配置文件的常见问题

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

CONFIG_FILE="${1:-config/config.yaml}"

echo "=========================================="
echo "Orbital Sentinels 配置检查"
echo "=========================================="
echo ""
echo "检查配置文件: $CONFIG_FILE"
echo ""

# 检查文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}✗ 配置文件不存在: $CONFIG_FILE${NC}"
    echo ""
    echo "建议："
    echo "  cp config/config.example.yaml config/config.yaml"
    exit 1
fi

echo -e "${GREEN}✓${NC} 配置文件存在"
echo ""

# 检查必需字段
echo "检查必需字段..."
echo ""

ERRORS=0
WARNINGS=0

# 检查 sender.flush_interval
FLUSH_INTERVAL=$(grep -A 15 "^sender:" "$CONFIG_FILE" | grep "flush_interval:" | head -1 | awk '{print $2}')
if [ -z "$FLUSH_INTERVAL" ]; then
    echo -e "${RED}✗ 缺少 sender.flush_interval 配置${NC}"
    echo "  这是导致 'non-positive interval for NewTicker' 错误的常见原因"
    echo ""
    echo "  修复方法："
    echo "  在 sender 部分添加："
    echo "    flush_interval: 10s"
    echo ""
    ERRORS=$((ERRORS + 1))
elif [ "$FLUSH_INTERVAL" = "0" ] || [ "$FLUSH_INTERVAL" = "0s" ]; then
    echo -e "${RED}✗ sender.flush_interval 值无效: $FLUSH_INTERVAL${NC}"
    echo "  必须是正数，例如: 10s, 30s, 1m"
    echo ""
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓${NC} sender.flush_interval: $FLUSH_INTERVAL"
fi

# 检查 sender.mode
SENDER_MODE=$(grep -A 1 "^sender:" "$CONFIG_FILE" | grep "mode:" | awk '{print $2}' | tr -d '"')
if [ -z "$SENDER_MODE" ]; then
    echo -e "${RED}✗ 缺少 sender.mode 配置${NC}"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓${NC} sender.mode: $SENDER_MODE"
    
    # 检查模式是否有效
    if [ "$SENDER_MODE" != "core" ] && [ "$SENDER_MODE" != "direct" ] && [ "$SENDER_MODE" != "hybrid" ]; then
        echo -e "${RED}✗ sender.mode 值无效: $SENDER_MODE${NC}"
        echo "  有效值: core, direct, hybrid"
        echo ""
        ERRORS=$((ERRORS + 1))
    fi
fi

# 检查 direct 模式配置
if [ "$SENDER_MODE" = "direct" ] || [ "$SENDER_MODE" = "hybrid" ]; then
    echo ""
    echo "检查 direct 模式配置..."
    
    PROM_ENABLED=$(grep -A 20 "direct:" "$CONFIG_FILE" | grep -A 2 "prometheus:" | grep "enabled:" | awk '{print $2}')
    VM_ENABLED=$(grep -A 20 "direct:" "$CONFIG_FILE" | grep -A 2 "victoria_metrics:" | grep "enabled:" | awk '{print $2}')
    CH_ENABLED=$(grep -A 20 "direct:" "$CONFIG_FILE" | grep -A 2 "clickhouse:" | grep "enabled:" | awk '{print $2}')
    
    if [ "$PROM_ENABLED" != "true" ] && [ "$VM_ENABLED" != "true" ] && [ "$CH_ENABLED" != "true" ]; then
        echo -e "${YELLOW}⚠${NC} 警告: direct 模式下没有启用任何数据库"
        echo "  至少启用一个: prometheus, victoria_metrics, clickhouse"
        echo ""
        WARNINGS=$((WARNINGS + 1))
    else
        echo -e "${GREEN}✓${NC} 已启用的数据库:"
        [ "$PROM_ENABLED" = "true" ] && echo "  - Prometheus"
        [ "$VM_ENABLED" = "true" ] && echo "  - VictoriaMetrics"
        [ "$CH_ENABLED" = "true" ] && echo "  - ClickHouse"
    fi
fi

# 检查 core 模式配置
if [ "$SENDER_MODE" = "core" ] || [ "$SENDER_MODE" = "hybrid" ]; then
    echo ""
    echo "检查 core 模式配置..."
    
    CORE_URL=$(grep -A 3 "^core:" "$CONFIG_FILE" | grep "url:" | awk '{print $2}')
    if [ -z "$CORE_URL" ]; then
        echo -e "${RED}✗ 缺少 core.url 配置${NC}"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "${GREEN}✓${NC} core.url: $CORE_URL"
        
        # 提示中心端可能不可用
        echo -e "${BLUE}ℹ${NC} 提示: 如果中心端不可用，建议使用 direct 模式"
        echo "  参考文档: docs/STANDALONE_MODE.md"
    fi
fi

# 检查其他重要配置
echo ""
echo "检查其他配置..."
echo ""

# 检查 buffer
BUFFER_SIZE=$(grep -A 3 "^buffer:" "$CONFIG_FILE" | grep "size:" | awk '{print $2}')
if [ -z "$BUFFER_SIZE" ] || [ "$BUFFER_SIZE" -le 0 ]; then
    echo -e "${YELLOW}⚠${NC} buffer.size 未配置或无效"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✓${NC} buffer.size: $BUFFER_SIZE"
fi

# 检查 collector
WORKER_POOL=$(grep -A 3 "^collector:" "$CONFIG_FILE" | grep "worker_pool_size:" | awk '{print $2}')
if [ -z "$WORKER_POOL" ] || [ "$WORKER_POOL" -le 0 ]; then
    echo -e "${YELLOW}⚠${NC} collector.worker_pool_size 未配置或无效"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✓${NC} collector.worker_pool_size: $WORKER_POOL"
fi

# 检查 plugins
PLUGIN_DIR=$(grep -A 3 "^plugins:" "$CONFIG_FILE" | grep "directory:" | awk '{print $2}' | tr -d '"')
if [ -z "$PLUGIN_DIR" ]; then
    echo -e "${YELLOW}⚠${NC} plugins.directory 未配置"
    WARNINGS=$((WARNINGS + 1))
else
    echo -e "${GREEN}✓${NC} plugins.directory: $PLUGIN_DIR"
    
    if [ ! -d "$PLUGIN_DIR" ]; then
        echo -e "${YELLOW}⚠${NC} 插件目录不存在: $PLUGIN_DIR"
        WARNINGS=$((WARNINGS + 1))
    fi
fi

# 总结
echo ""
echo "=========================================="
if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ 配置检查通过！${NC}"
    echo ""
    echo "可以启动 Sentinel:"
    echo "  ./bin/sentinel start -c $CONFIG_FILE"
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ 配置检查完成，有 $WARNINGS 个警告${NC}"
    echo ""
    echo "警告不会阻止启动，但建议修复"
    echo ""
    echo "可以启动 Sentinel:"
    echo "  ./bin/sentinel start -c $CONFIG_FILE"
else
    echo -e "${RED}✗ 配置检查失败，有 $ERRORS 个错误${NC}"
    echo ""
    echo "请修复错误后再启动 Sentinel"
    exit 1
fi
echo "=========================================="
echo ""

# 额外建议
if [ "$SENDER_MODE" = "core" ]; then
    echo "💡 提示："
    echo ""
    echo "如果中心端不可用，可以使用 direct 模式："
    echo "  1. 修改配置: sender.mode: \"direct\""
    echo "  2. 启用至少一个数据库 (prometheus/victoria_metrics/clickhouse)"
    echo "  3. 参考文档: docs/STANDALONE_MODE.md"
    echo ""
fi

