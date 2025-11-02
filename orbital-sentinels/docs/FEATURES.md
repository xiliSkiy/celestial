# Orbital Sentinels 功能特性

## 📋 目录

- [核心功能](#核心功能)
- [数据采集](#数据采集)
- [数据发送](#数据发送)
- [插件系统](#插件系统)
- [配置管理](#配置管理)
- [监控和日志](#监控和日志)
- [部署方式](#部署方式)

## 核心功能

### ✅ 已实现

#### 1. 插件化架构
- **标准化接口**: 定义了统一的插件接口，便于扩展
- **动态加载**: 支持运行时加载插件
- **插件管理**: 集中管理所有插件的生命周期
- **元数据支持**: 插件可以声明自己的元数据和配置需求

#### 2. 任务调度
- **工作池**: 并发执行采集任务，可配置并发数
- **任务拉取**: 定期从中心端拉取采集任务
- **超时控制**: 单个任务执行时间限制
- **失败重试**: 自动重试失败的任务

#### 3. 数据缓冲
- **内存缓冲**: 高性能的内存缓冲实现
- **溢出保护**: 缓冲区满时自动丢弃旧数据
- **批量处理**: 支持批量读取数据
- **线程安全**: 支持并发读写

#### 4. 数据发送
- **三种模式**: 
  - Core 模式：发送到中心端
  - Direct 模式：直连时序数据库
  - Hybrid 模式：同时发送到两者
- **批量发送**: 减少网络开销
- **失败重试**: 自动重试失败的发送
- **熔断保护**: 防止雪崩效应

#### 5. 心跳管理
- **定时心跳**: 定期向中心端报告状态
- **系统指标**: 采集并上报系统资源使用情况
- **配置更新**: 接收中心端的配置更新
- **健康检查**: 监控 Sentinel 自身健康状态

## 数据采集

### 支持的插件类型

#### ✅ Ping 插件
- ICMP Echo 请求
- RTT 延迟测量
- 丢包率统计
- 连通性检测

#### 🔄 计划中
- **SNMP 插件**: 网络设备监控
- **HTTP 插件**: Web 服务监控
- **Modbus 插件**: 工业设备监控
- **MQTT 插件**: IoT 设备监控
- **Database 插件**: 数据库监控
- **Custom Script 插件**: 自定义脚本执行

### 采集特性

- **并发采集**: 多个任务并发执行
- **超时控制**: 防止任务卡死
- **错误处理**: 完善的错误处理机制
- **指标标准化**: 统一的指标格式

## 数据发送

### 支持的目标

#### ✅ Prometheus
- **协议**: Remote Write
- **压缩**: Snappy
- **认证**: Basic Auth
- **多租户**: 支持 X-Scope-OrgID
- **特性**:
  - 自动格式转换
  - 标签映射
  - 时间戳转换
  - 批量发送

#### ✅ VictoriaMetrics
- **协议**: Remote Write（兼容 Prometheus）
- **压缩**: Snappy
- **认证**: Basic Auth
- **多租户**: 支持 X-Scope-OrgID
- **特性**:
  - 单机版支持
  - 集群版支持
  - 高性能写入
  - 长期存储

#### ✅ ClickHouse
- **协议**: Native TCP
- **压缩**: 可选
- **认证**: 用户名/密码
- **特性**:
  - 自动建表
  - 批量插入
  - 分区管理
  - TTL 支持
  - Map 类型标签存储

#### ✅ Gravital Core（中心端）
- **协议**: HTTP/JSON
- **压缩**: Gzip
- **认证**: API Token
- **特性**:
  - 统一管理
  - 数据转发
  - 告警规则
  - 设备管理

### 发送特性

- **批量发送**: 提高吞吐量
- **数据压缩**: 减少带宽使用
- **失败重试**: 自动重试机制
- **熔断保护**: 防止级联故障
- **并发发送**: 同时发送到多个目标
- **超时控制**: 防止发送阻塞

## 插件系统

### 插件接口

```go
type Plugin interface {
    Meta() PluginMeta
    Schema() PluginSchema
    Init(config map[string]interface{}) error
    ValidateConfig(config map[string]interface{}) error
    TestConnection(config map[string]interface{}) error
    Collect(ctx context.Context, task *CollectionTask) ([]*Metric, error)
    Close() error
}
```

### 插件开发

- **SDK 支持**: 提供 BasePlugin 简化开发
- **配置验证**: 自动验证配置
- **连接测试**: 测试设备连接
- **错误处理**: 统一的错误处理
- **日志记录**: 集成日志系统

### 插件配置

```yaml
# plugin.yaml
name: "ping"
version: "1.0.0"
description: "ICMP Ping 监控插件"
author: "Celestial Team"

device_fields:
  - name: "host"
    type: "string"
    required: true
    description: "目标主机地址"

metrics:
  - name: "rtt_ms"
    type: "gauge"
    unit: "ms"
    description: "往返时延"
```

## 配置管理

### 配置来源

- **配置文件**: YAML 格式
- **环境变量**: 支持环境变量覆盖
- **默认值**: 合理的默认配置

### 配置热更新

- **心跳更新**: 通过心跳接收配置更新
- **插件重载**: 支持插件配置重载（计划中）

### 配置验证

- **格式验证**: 自动验证配置格式
- **值验证**: 验证配置值的合法性
- **依赖检查**: 检查配置依赖关系

## 监控和日志

### 日志系统

- **日志级别**: Debug, Info, Warn, Error
- **日志格式**: Text, JSON
- **日志输出**: Stdout, File, Both
- **日志轮转**: 
  - 按大小轮转
  - 保留数量限制
  - 保留时间限制

### 内部指标

- **采集统计**:
  - 任务执行次数
  - 成功/失败次数
  - 执行耗时

- **发送统计**:
  - 发送成功/失败次数
  - 发送字节数
  - 发送延迟

- **系统指标**:
  - CPU 使用率
  - 内存使用量
  - 磁盘使用量
  - 网络流量

## 部署方式

### ✅ 二进制部署

```bash
# 编译
make build

# 运行
./bin/sentinel start -c config.yaml
```

### ✅ Docker 部署

```bash
# 构建镜像
docker build -t orbital-sentinel:latest .

# 运行容器
docker run -d \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/plugins:/app/plugins \
  orbital-sentinel:latest
```

### ✅ Docker Compose 部署

```bash
# 启动完整监控栈
cd examples
docker-compose up -d
```

### 🔄 Kubernetes 部署（计划中）

- Helm Chart
- DaemonSet 部署
- ConfigMap 配置
- Secret 管理

### 🔄 Systemd 服务（计划中）

```bash
# 安装服务
sudo systemctl enable sentinel
sudo systemctl start sentinel
```

## 性能特性

### 高性能

- **并发采集**: 工作池并发执行
- **批量处理**: 批量读写数据
- **数据压缩**: 减少网络传输
- **内存优化**: 高效的内存使用

### 高可靠

- **数据缓冲**: 本地缓冲防止数据丢失
- **失败重试**: 自动重试机制
- **熔断保护**: 防止级联故障
- **优雅关闭**: 确保数据不丢失

### 可扩展

- **插件化**: 易于添加新功能
- **配置灵活**: 丰富的配置选项
- **水平扩展**: 支持多实例部署

## 安全特性

### 认证

- **API Token**: 中心端认证
- **Basic Auth**: 数据库认证
- **TLS/SSL**: 加密传输（计划中）

### 数据安全

- **配置加密**: 敏感配置加密存储（计划中）
- **传输加密**: HTTPS/TLS 支持
- **访问控制**: 基于角色的访问控制（计划中）

## 兼容性

### 操作系统

- ✅ Linux (amd64, arm64)
- ✅ macOS (amd64, arm64)
- ✅ Windows (amd64)

### Go 版本

- ✅ Go 1.21+
- ✅ Go 1.22+
- ✅ Go 1.23+

### 数据库版本

- ✅ Prometheus 2.25+
- ✅ VictoriaMetrics 1.80+
- ✅ ClickHouse 21.8+

## 未来规划

### v1.2.0
- [ ] 磁盘缓冲（BadgerDB）
- [ ] SNMP 插件
- [ ] HTTP 插件
- [ ] 插件热重载
- [ ] 性能监控面板

### v1.3.0
- [ ] Modbus 插件
- [ ] MQTT 插件
- [ ] 分布式追踪
- [ ] 高可用支持
- [ ] Web 管理界面

### v2.0.0
- [ ] gRPC 插件系统
- [ ] 插件沙箱
- [ ] 配置加密
- [ ] TLS 支持
- [ ] Kubernetes Operator

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

- **代码贡献**: 提交 Pull Request
- **问题报告**: 创建 Issue
- **功能建议**: 创建 Feature Request
- **文档改进**: 改进文档

## 许可证

本项目采用 MIT 许可证。

