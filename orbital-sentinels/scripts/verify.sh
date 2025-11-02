#!/bin/bash

# Orbital Sentinels 验证脚本
# 用于验证项目的完整性和功能

set -e

echo "=========================================="
echo "Orbital Sentinels 验证脚本"
echo "=========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查函数
check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
        exit 1
    fi
}

# 1. 检查 Go 版本
echo "1. 检查 Go 版本..."
go version
check "Go 已安装"
echo ""

# 2. 检查依赖
echo "2. 检查项目依赖..."
go mod verify
check "依赖验证通过"
echo ""

# 3. 编译项目
echo "3. 编译项目..."
go build -o bin/sentinel ./cmd/sentinel/
check "编译成功"
echo ""

# 4. 运行测试
echo "4. 运行单元测试..."
go test -v ./... > /tmp/test_output.txt 2>&1
check "单元测试通过"
echo ""

# 5. 统计测试结果
echo "5. 测试统计..."
TOTAL_TESTS=$(grep -c "^=== RUN" /tmp/test_output.txt || echo "0")
PASSED_TESTS=$(grep -c "^--- PASS:" /tmp/test_output.txt || echo "0")
FAILED_TESTS=$(grep -c "^--- FAIL:" /tmp/test_output.txt || echo "0")

echo "   总测试数: $TOTAL_TESTS"
echo "   通过: ${GREEN}$PASSED_TESTS${NC}"
echo "   失败: ${RED}$FAILED_TESTS${NC}"
echo ""

# 6. 检查代码格式
echo "6. 检查代码格式..."
gofmt -l . > /tmp/fmt_output.txt
if [ -s /tmp/fmt_output.txt ]; then
    echo -e "${YELLOW}⚠${NC} 以下文件需要格式化:"
    cat /tmp/fmt_output.txt
else
    check "代码格式正确"
fi
echo ""

# 7. 检查文件结构
echo "7. 检查项目结构..."
REQUIRED_DIRS=(
    "cmd/sentinel"
    "internal/agent"
    "internal/plugin"
    "internal/scheduler"
    "internal/buffer"
    "internal/sender"
    "internal/heartbeat"
    "internal/pkg/config"
    "internal/pkg/logger"
    "sdk"
    "plugins/ping"
    "config"
    "docs"
    "examples"
)

for dir in "${REQUIRED_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "   ${GREEN}✓${NC} $dir"
    else
        echo -e "   ${RED}✗${NC} $dir (缺失)"
        exit 1
    fi
done
echo ""

# 8. 检查关键文件
echo "8. 检查关键文件..."
REQUIRED_FILES=(
    "go.mod"
    "go.sum"
    "Makefile"
    "Dockerfile"
    "README.md"
    "config/config.example.yaml"
    "cmd/sentinel/main.go"
    "internal/agent/agent.go"
)

for file in "${REQUIRED_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "   ${GREEN}✓${NC} $file"
    else
        echo -e "   ${RED}✗${NC} $file (缺失)"
        exit 1
    fi
done
echo ""

# 9. 检查文档
echo "9. 检查文档..."
REQUIRED_DOCS=(
    "README.md"
    "QUICKSTART.md"
    "CHANGELOG.md"
    "SUMMARY.md"
    "docs/DIRECT_SENDER_GUIDE.md"
    "docs/FEATURES.md"
    "examples/README.md"
)

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "   ${GREEN}✓${NC} $doc"
    else
        echo -e "   ${YELLOW}⚠${NC} $doc (缺失)"
    fi
done
echo ""

# 10. 检查版本
echo "10. 检查版本信息..."
./bin/sentinel version
check "版本命令可用"
echo ""

# 11. 检查配置文件
echo "11. 验证配置文件..."
if [ -f "config/config.example.yaml" ]; then
    echo -e "   ${GREEN}✓${NC} 配置示例文件存在"
else
    echo -e "   ${RED}✗${NC} 配置示例文件缺失"
    exit 1
fi
echo ""

# 12. 统计代码行数
echo "12. 代码统计..."
echo "   Go 源文件:"
find . -name "*.go" -not -path "./vendor/*" | wc -l | xargs echo "     文件数:"
find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} + | tail -1 | awk '{print "     代码行数: " $1}'

echo "   测试文件:"
find . -name "*_test.go" | wc -l | xargs echo "     文件数:"

echo "   文档文件:"
find . -name "*.md" | wc -l | xargs echo "     文件数:"
echo ""

# 13. 检查 Docker 文件
echo "13. 检查 Docker 支持..."
if [ -f "Dockerfile" ]; then
    echo -e "   ${GREEN}✓${NC} Dockerfile 存在"
fi
if [ -f "examples/docker-compose.yml" ]; then
    echo -e "   ${GREEN}✓${NC} docker-compose.yml 存在"
fi
echo ""

# 14. 检查示例配置
echo "14. 检查示例配置..."
if [ -f "examples/config.yaml" ]; then
    echo -e "   ${GREEN}✓${NC} 示例配置存在"
fi
if [ -f "examples/prometheus.yml" ]; then
    echo -e "   ${GREEN}✓${NC} Prometheus 配置存在"
fi
echo ""

# 总结
echo "=========================================="
echo -e "${GREEN}✓ 所有验证通过！${NC}"
echo "=========================================="
echo ""
echo "项目状态:"
echo "  - 编译: ✓"
echo "  - 测试: ✓ ($PASSED_TESTS/$TOTAL_TESTS)"
echo "  - 文档: ✓"
echo "  - 结构: ✓"
echo ""
echo "可以开始使用 Orbital Sentinels！"
echo ""
echo "快速开始:"
echo "  1. 复制配置: cp config/config.example.yaml config.yaml"
echo "  2. 编辑配置: vim config.yaml"
echo "  3. 启动服务: ./bin/sentinel start -c config.yaml"
echo ""
echo "或使用 Docker Compose:"
echo "  cd examples && docker-compose up -d"
echo ""

