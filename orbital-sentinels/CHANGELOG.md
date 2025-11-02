# 更新日志

## [1.1.0] - 2025-11-01

### ✨ 新增功能

#### 直连发送器
- **Prometheus 写入器**: 支持通过 Remote Write 协议直接发送数据到 Prometheus
  - 支持基本认证（用户名/密码）
  - 支持自定义请求头
  - 自动 Snappy 压缩
  - 完整的错误处理和重试机制

- **VictoriaMetrics 写入器**: 支持直接发送数据到 VictoriaMetrics
  - 兼容 Prometheus Remote Write 协议
  - 支持单机版和集群版
  - 支持多租户配置（X-Scope-OrgID）
  - 支持基本认证和自定义请求头

- **ClickHouse 写入器**: 支持直接发送数据到 ClickHouse
  - 使用原生 TCP 协议
  - 自动创建表结构
  - 支持批量插入优化
  - 支持数据压缩
  - 自动分区和 TTL 管理
  - Map 类型存储标签

#### 发送模式
- **直连模式 (direct)**: 数据直接发送到时序数据库
- **中心端模式 (core)**: 数据发送到中心端转发（已有）
- **混合模式 (hybrid)**: 同时发送到中心端和时序数据库

#### 配置增强
- 扩展配置结构，支持多数据库配置
- 支持每个数据库独立的认证配置
- 支持自定义请求头
- 支持 ClickHouse 表名和批量大小配置

### 📚 文档

- 新增 `docs/DIRECT_SENDER_GUIDE.md` - 直连发送器完整使用指南
  - 详细的配置说明
  - 每个数据库的配置示例
  - 性能优化建议
  - 故障排查指南
  - 最佳实践

- 新增 `examples/` 目录 - 完整的部署示例
  - Docker Compose 配置
  - Prometheus 配置示例
  - Grafana 集成示例
  - 快速开始指南

- 更新 `README.md` - 添加直连发送器说明

### 🧪 测试

- 新增 `internal/sender/direct_sender_test.go`
  - Prometheus 格式转换测试
  - VictoriaMetrics 格式转换测试
  - 认证配置测试
  - 自定义请求头测试
  - 8 个测试用例，全部通过

### 📦 依赖

新增以下依赖：
- `github.com/prometheus/prometheus/prompb` - Prometheus Remote Write 协议
- `github.com/golang/snappy` - Snappy 压缩
- `github.com/ClickHouse/clickhouse-go/v2` - ClickHouse Go 驱动

### 🔧 改进

- 优化 `DirectSender` 实现，支持并发发送到多个目标
- 增强错误处理和日志记录
- 改进配置验证和默认值处理
- 添加熔断保护机制

### 📊 性能

- 支持批量发送，减少网络开销
- 自动压缩，降低带宽使用
- 并发发送到多个目标
- 可配置的缓冲区和刷新策略

---

## [1.0.0] - 2025-11-01

### ✨ 初始版本

#### 核心功能
- 插件化架构
- 任务调度系统
- 数据缓冲机制
- 心跳管理
- 中心端发送器

#### 插件系统
- 标准化插件接口
- 插件管理器
- 动态加载支持
- Ping 示例插件

#### 基础设施
- 配置管理（Viper）
- 日志系统（Zap）
- 工作池
- 内存缓冲

#### 测试
- 内存缓冲测试
- 工作池测试
- 插件管理器测试
- 11 个测试用例

#### 文档
- README
- 快速开始指南
- 配置示例
- 插件开发指南

#### 部署
- Makefile
- Dockerfile
- systemd 服务文件

---

## 版本说明

版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **主版本号**: 不兼容的 API 修改
- **次版本号**: 向下兼容的功能性新增
- **修订号**: 向下兼容的问题修正

## 未来计划

### v1.2.0
- [ ] 磁盘缓冲实现（BadgerDB）
- [ ] 更多插件（SNMP, HTTP, Modbus）
- [ ] 插件热重载
- [ ] 性能监控和指标

### v1.3.0
- [ ] 分布式追踪
- [ ] 高可用支持
- [ ] 插件市场
- [ ] Web 管理界面

### v2.0.0
- [ ] gRPC 插件系统
- [ ] 插件沙箱
- [ ] 增强的安全特性
- [ ] 云原生部署

