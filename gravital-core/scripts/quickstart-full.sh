#!/bin/bash

# Gravital Core 完整环境快速启动脚本
# 包含 PostgreSQL, Redis, VictoriaMetrics, ClickHouse, Grafana

set -e

echo "================================"
echo "Gravital Core 完整环境启动"
echo "================================"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "错误: 未安装 Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "错误: 未安装 Docker Compose"
    exit 1
fi

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}1. 停止并清理现有容器...${NC}"
docker-compose -f docker-compose.full.yaml down -v

echo ""
echo -e "${YELLOW}2. 构建镜像...${NC}"
docker-compose -f docker-compose.full.yaml build

echo ""
echo -e "${YELLOW}3. 启动所有服务...${NC}"
docker-compose -f docker-compose.full.yaml up -d

echo ""
echo -e "${YELLOW}4. 等待服务启动...${NC}"
sleep 10

echo ""
echo -e "${YELLOW}5. 检查服务状态...${NC}"
docker-compose -f docker-compose.full.yaml ps

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}服务已启动！${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

echo "访问地址："
echo "  - Gravital Core API: http://localhost:8080"
echo "  - Grafana:           http://localhost:3000 (admin/admin)"
echo "  - VictoriaMetrics:   http://localhost:8428"
echo "  - ClickHouse HTTP:   http://localhost:8123"
echo ""

echo "健康检查："
echo "  - API:               curl http://localhost:8080/health"
echo "  - VictoriaMetrics:   curl http://localhost:8428/health"
echo "  - ClickHouse:        curl http://localhost:8123/ping"
echo ""

echo "查看日志："
echo "  - 所有服务:          docker-compose -f docker-compose.full.yaml logs -f"
echo "  - Gravital Core:     docker-compose -f docker-compose.full.yaml logs -f gravital-core"
echo "  - VictoriaMetrics:   docker-compose -f docker-compose.full.yaml logs -f victoria"
echo "  - ClickHouse:        docker-compose -f docker-compose.full.yaml logs -f clickhouse"
echo ""

echo "测试数据转发："
echo "  - 发送测试数据:      ./scripts/test-forwarder.sh"
echo ""

echo "停止服务："
echo "  - docker-compose -f docker-compose.full.yaml down"
echo ""

# 等待服务完全启动
echo -e "${YELLOW}等待服务完全启动（30秒）...${NC}"
for i in {1..30}; do
    echo -n "."
    sleep 1
done
echo ""

# 健康检查
echo ""
echo -e "${YELLOW}执行健康检查...${NC}"

check_service() {
    local name=$1
    local url=$2
    
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $name 正常${NC}"
        return 0
    else
        echo -e "${RED}✗ $name 异常${NC}"
        return 1
    fi
}

check_service "Gravital Core" "http://localhost:8080/health"
check_service "VictoriaMetrics" "http://localhost:8428/health"
check_service "ClickHouse" "http://localhost:8123/ping"
check_service "Grafana" "http://localhost:3000/api/health"

echo ""
echo -e "${GREEN}环境已就绪！${NC}"
echo ""
echo "下一步："
echo "1. 访问 Grafana (http://localhost:3000) 配置数据源"
echo "2. 运行测试脚本: ./scripts/test-forwarder.sh"
echo "3. 查看 API 文档: http://localhost:8080/swagger/index.html (如果已配置)"
echo ""

