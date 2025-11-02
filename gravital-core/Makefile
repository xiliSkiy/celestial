.PHONY: all build run test clean docker migrate help

# 变量定义
APP_NAME=gravital-core
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(shell go version | awk '{print $$3}')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GoVersion=$(GO_VERSION)"

# 默认目标
all: build

# 编译
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server
	@echo "Build complete: bin/$(APP_NAME)"

# 运行
run:
	@echo "Running $(APP_NAME)..."
	go run ./cmd/server -c config/config.yaml

# 开发模式（热重载）
dev:
	@echo "Running in development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# 测试
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 清理
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# 安装依赖
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 代码格式化
fmt:
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

# 代码检查
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# 数据库迁移
migrate-up:
	@echo "Running migrations..."
	@which migrate > /dev/null || (echo "Please install golang-migrate: https://github.com/golang-migrate/migrate" && exit 1)
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path migrations -database "$(DB_URL)" down

migrate-create:
	@echo "Creating migration: $(NAME)"
	@which migrate > /dev/null || (echo "Please install golang-migrate" && exit 1)
	migrate create -ext sql -dir migrations -seq $(NAME)

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Docker 运行
docker-run:
	@echo "Running Docker container..."
	docker run -d \
		--name $(APP_NAME) \
		-p 8080:8080 \
		-v $(PWD)/config:/app/config \
		$(APP_NAME):latest

# Docker Compose
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down:
	@echo "Stopping services..."
	docker-compose down

# 生成 API 文档
api-docs:
	@echo "Generating API documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/server/main.go -o docs/swagger

# 版本信息
version:
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

# 帮助信息
help:
	@echo "Gravital Core - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build              编译项目"
	@echo "  run                运行项目"
	@echo "  dev                开发模式（热重载）"
	@echo "  test               运行测试"
	@echo "  test-coverage      运行测试并生成覆盖率报告"
	@echo "  clean              清理构建文件"
	@echo "  deps               下载依赖"
	@echo "  fmt                格式化代码"
	@echo "  lint               代码检查"
	@echo "  migrate-up         运行数据库迁移"
	@echo "  migrate-down       回滚数据库迁移"
	@echo "  migrate-create     创建新的迁移文件 (NAME=xxx)"
	@echo "  docker-build       构建 Docker 镜像"
	@echo "  docker-run         运行 Docker 容器"
	@echo "  docker-compose-up  启动所有服务"
	@echo "  docker-compose-down 停止所有服务"
	@echo "  api-docs           生成 API 文档"
	@echo "  version            显示版本信息"
	@echo "  help               显示帮助信息"

