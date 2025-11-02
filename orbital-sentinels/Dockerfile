# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o sentinel \
    ./cmd/sentinel

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    ca-certificates \
    iputils \
    tzdata

# 创建非 root 用户
RUN addgroup -g 1000 sentinel && \
    adduser -D -u 1000 -G sentinel sentinel

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /build/sentinel .
COPY --from=builder /build/config/config.example.yaml ./config/config.yaml
COPY --from=builder /build/plugins ./plugins

# 创建数据目录
RUN mkdir -p /app/data /app/logs && \
    chown -R sentinel:sentinel /app

# 切换到非 root 用户
USER sentinel

# 暴露端口（如果需要）
# EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep -f sentinel || exit 1

# 启动命令
ENTRYPOINT ["./sentinel"]
CMD ["start", "-c", "config/config.yaml"]

