.PHONY: build clean test run docker-build

# 变量定义
BINARY_NAME=sentinel
VERSION?=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# 默认目标
all: build

# 构建二进制
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/sentinel
	@echo "Build complete: bin/$(BINARY_NAME)"

# 构建所有平台
build-all:
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/sentinel
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/sentinel
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/sentinel
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/sentinel
	@echo "Build complete for all platforms"

# 构建插件
build-plugins:
	@echo "Building plugins..."
	@cd plugins/ping && go build -buildmode=plugin -o ping.so ping.go
	@echo "Plugins built"

# 运行
run: build
	@./bin/$(BINARY_NAME) start -c config/config.example.yaml

# 测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@find plugins -name "*.so" -delete
	@echo "Clean complete"

# Docker 构建
docker-build:
	@echo "Building Docker image..."
	@docker build -t celestial/sentinel:$(VERSION) .
	@docker tag celestial/sentinel:$(VERSION) celestial/sentinel:latest
	@echo "Docker image built: celestial/sentinel:$(VERSION)"

# Docker 运行
docker-run:
	@docker run -d --name sentinel \
		-v $(PWD)/config:/app/config \
		-v $(PWD)/plugins:/app/plugins \
		celestial/sentinel:latest

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# 代码检查
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

# 生成文档
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

# 帮助
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for all platforms"
	@echo "  build-plugins - Build plugins"
	@echo "  run           - Run the sentinel"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  deps          - Install dependencies"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  docs          - Generate documentation"
	@echo "  help          - Show this help message"

