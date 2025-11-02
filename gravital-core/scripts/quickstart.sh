#!/bin/bash

# Gravital Core 快速启动脚本

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "=========================================="
echo -e "${BLUE}  Gravital Core 快速启动${NC}"
echo "=========================================="
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: 未安装 Docker${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}错误: 未安装 Docker Compose${NC}"
    echo "请先安装 Docker Compose"
    exit 1
fi

echo -e "${BLUE}步骤 1: 准备配置文件${NC}"
echo "----------------------------------------"
if [ ! -f "config/config.yaml" ]; then
    cp config/config.example.yaml config/config.yaml
    echo -e "${GREEN}✓ 已创建配置文件${NC}"
else
    echo -e "${YELLOW}! 配置文件已存在，跳过${NC}"
fi
echo ""

echo -e "${BLUE}步骤 2: 启动数据库服务${NC}"
echo "----------------------------------------"
docker-compose up -d postgres redis
echo -e "${GREEN}✓ PostgreSQL 和 Redis 已启动${NC}"
echo ""

echo -e "${BLUE}步骤 3: 等待数据库就绪${NC}"
echo "----------------------------------------"
echo "等待 PostgreSQL 启动..."
for i in {1..30}; do
    if docker exec gravital-postgres pg_isready -U postgres > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PostgreSQL 已就绪${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗ PostgreSQL 启动超时${NC}"
        exit 1
    fi
    sleep 1
done
echo ""

echo -e "${BLUE}步骤 4: 运行数据库迁移${NC}"
echo "----------------------------------------"
# 检查是否安装了 golang-migrate
if command -v migrate &> /dev/null; then
    migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/gravital?sslmode=disable" up
    echo -e "${GREEN}✓ 数据库迁移完成${NC}"
else
    echo -e "${YELLOW}! 未安装 golang-migrate，跳过迁移${NC}"
    echo "请手动运行迁移或安装 golang-migrate"
    echo "安装: brew install golang-migrate (macOS)"
fi
echo ""

echo -e "${BLUE}步骤 5: 启动 VictoriaMetrics 和 Grafana${NC}"
echo "----------------------------------------"
docker-compose up -d victoria-metrics grafana
echo -e "${GREEN}✓ VictoriaMetrics 和 Grafana 已启动${NC}"
echo ""

echo -e "${BLUE}步骤 6: 构建并启动 Gravital Core${NC}"
echo "----------------------------------------"
if [ -f "bin/gravital-core" ]; then
    echo "使用已编译的二进制文件..."
    ./bin/gravital-core -c config/config.yaml &
    CORE_PID=$!
    echo -e "${GREEN}✓ Gravital Core 已启动 (PID: $CORE_PID)${NC}"
else
    echo "编译 Gravital Core..."
    make build
    ./bin/gravital-core -c config/config.yaml &
    CORE_PID=$!
    echo -e "${GREEN}✓ Gravital Core 已启动 (PID: $CORE_PID)${NC}"
fi
echo ""

echo "等待服务启动..."
sleep 5

echo "=========================================="
echo -e "${GREEN}  启动完成！${NC}"
echo "=========================================="
echo ""

echo "📊 服务信息:"
echo "  - Gravital Core API: http://localhost:8080"
echo "  - VictoriaMetrics UI: http://localhost:8428/vmui"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - PostgreSQL: localhost:5432 (postgres/postgres)"
echo "  - Redis: localhost:6379"
echo ""

echo "🔑 默认账号:"
echo "  - 用户名: admin"
echo "  - 密码: admin123"
echo ""

echo "🚀 快速测试:"
echo "  # 健康检查"
echo "  curl http://localhost:8080/health"
echo ""
echo "  # 登录"
echo "  curl -X POST http://localhost:8080/api/v1/auth/login \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"username\":\"admin\",\"password\":\"admin123\"}'"
echo ""

echo "📚 文档:"
echo "  - README: ./README.md"
echo "  - API 文档: ../docs/05-API接口文档.md"
echo ""

echo "🛑 停止服务:"
echo "  kill $CORE_PID"
echo "  docker-compose down"
echo ""

