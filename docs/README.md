# Celestial 监控系统 - 文档索引

> 一个现代化、分布式、插件化的监控系统

## 📚 文档导航

### 快速开始
- [00-产品需求分析与优化建议](./00-产品需求分析与优化建议.md) - 了解项目背景、需求分析和设计思路
- [01-系统整体架构设计](./01-系统整体架构设计.md) - 了解系统整体架构和技术选型

### 核心设计
- [02-中心端详细设计](./02-中心端详细设计.md) - Gravital Core 中心端详细设计
- [03-采集端详细设计](./03-采集端详细设计.md) - Orbital Sentinels 采集端详细设计

### 开发指南
- [04-插件开发指南](./04-插件开发指南.md) - 学习如何开发采集插件
- [05-API接口文档](./05-API接口文档.md) - 完整的 REST API 接口文档

### 部署运维
- [06-部署运维手册](./06-部署运维手册.md) - 部署、配置、监控和故障排查指南

## 🎯 项目概述

### 系统架构

```
┌─────────────────────────────────────────────┐
│          用户界面 (Web UI / Grafana)         │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│         中心端 (Gravital Core)              │
│  - 设备管理  - 告警管理  - 任务调度         │
│  - 数据转发  - 用户认证  - 配置管理         │
└─────────────────┬───────────────────────────┘
                  │
         ┌────────┴─────────┐
         │                  │
┌────────▼────────┐  ┌──────▼────────────┐
│ Prometheus/VM/  │  │  Orbital Sentinels│
│   ClickHouse    │  │   (采集端集群)     │
└─────────────────┘  └──────┬────────────┘
                            │
                     ┌──────┴──────┐
                     │  监控设备    │
                     └─────────────┘
```

### 核心特性

#### 🎨 中心端 (Gravital Core)
- ✅ **设备管理**: 支持分组、标签、批量导入，管理大规模设备
- ✅ **告警管理**: 灵活的告警规则、多通知渠道、静默和抑制
- ✅ **数据转发**: 支持 Prometheus、VictoriaMetrics、ClickHouse
- ✅ **数据展示**: Grafana 集成 + 自定义 Dashboard
- ✅ **任务调度**: 智能分配采集任务，支持多种调度策略
- ✅ **用户认证**: JWT + RBAC 权限控制
- ✅ **API 网关**: RESTful API + WebSocket

#### 🛰️ 采集端 (Orbital Sentinels)
- ✅ **插件化架构**: 标准化插件接口，支持快速扩展
- ✅ **灵活部署**: 支持边缘计算、跨网络采集
- ✅ **多种数据流**: 直连、中转、混合三种模式
- ✅ **高可靠性**: 本地缓冲、断线重连、失败重试
- ✅ **低资源占用**: Go 语言实现，内存 < 500MB，CPU < 30%
- ✅ **热更新**: 支持插件和配置热更新

#### 🔌 插件生态
- **网络协议**: SNMP, Modbus, OPC-UA, MQTT
- **系统监控**: SSH, WMI, HTTP, Ping
- **数据库**: MySQL, PostgreSQL, Redis, MongoDB
- **云服务**: AWS, Azure, Prometheus
- **自定义**: 提供 SDK，轻松开发专属插件

## 📖 阅读指南

### 我是产品经理/架构师
1. 阅读 [00-产品需求分析与优化建议](./00-产品需求分析与优化建议.md)
2. 阅读 [01-系统整体架构设计](./01-系统整体架构设计.md)
3. 了解技术实现细节：[02-中心端详细设计](./02-中心端详细设计.md) 和 [03-采集端详细设计](./03-采集端详细设计.md)

### 我是后端开发工程师
1. 快速了解 [01-系统整体架构设计](./01-系统整体架构设计.md)
2. 深入学习 [02-中心端详细设计](./02-中心端详细设计.md)
3. 参考 [05-API接口文档](./05-API接口文档.md) 实现 API

### 我是前端开发工程师
1. 了解 [01-系统整体架构设计](./01-系统整体架构设计.md) 中的前端部分
2. 阅读 [05-API接口文档](./05-API接口文档.md)
3. 参考中心端设计中的 Dashboard 设计

### 我是插件开发者
1. 阅读 [04-插件开发指南](./04-插件开发指南.md)
2. 参考示例插件快速上手
3. 加入社区交流插件开发经验

### 我是运维工程师
1. 阅读 [06-部署运维手册](./06-部署运维手册.md)
2. 了解系统要求和部署方式
3. 学习监控、备份和故障排查

## 🚀 快速开始

### 使用 Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/celestial/celestial.git
cd celestial

# 配置环境变量
cp .env.example .env
vi .env

# 启动服务
docker-compose up -d

# 访问 Web UI
open http://localhost:8080
```

### 手动部署

```bash
# 1. 安装依赖
# PostgreSQL, Redis, Prometheus/VictoriaMetrics

# 2. 下载二进制
wget https://releases.celestial.io/v1.0.0/gravital-core-linux-amd64.tar.gz
tar -xzf gravital-core-linux-amd64.tar.gz

# 3. 配置
cp config/config.example.yaml config/config.yaml
vi config/config.yaml

# 4. 初始化数据库
./gravital-core migrate up

# 5. 创建管理员
./gravital-core user create --username admin --password Admin@123 --role admin

# 6. 启动服务
./gravital-core start
```

详细步骤请参考 [部署运维手册](./06-部署运维手册.md)。

## 📊 系统要求

### 中心端
```yaml
最小配置:
  CPU: 2 核
  内存: 4GB
  磁盘: 50GB SSD
  适用: < 100 设备

推荐配置:
  CPU: 4 核
  内存: 8GB
  磁盘: 200GB SSD
  适用: 100-1000 设备

大规模配置:
  CPU: 8+ 核
  内存: 16GB+
  磁盘: 500GB+ SSD
  适用: > 1000 设备
```

### 采集端
```yaml
最小配置:
  CPU: 1 核
  内存: 512MB
  磁盘: 10GB
  
推荐配置:
  CPU: 2 核
  内存: 2GB
  磁盘: 20GB
```

## 🛠️ 技术栈

### 中心端
- **后端**: Go 1.21+, Gin/Echo
- **数据库**: PostgreSQL 15, Redis 7
- **前端**: Vue 3, TypeScript, Element Plus
- **时序数据库**: Prometheus / VictoriaMetrics / ClickHouse
- **可视化**: Grafana 10.x

### 采集端
- **语言**: Go 1.21+
- **插件系统**: Go Plugin / gRPC
- **配置**: YAML
- **日志**: Zap

## 📈 路线图

### Phase 1: MVP ✅
- [x] 中心端基础框架
- [x] 采集端基础框架
- [x] 2-3 个基础插件
- [x] 基础 Web UI
- [x] 数据直连 Prometheus

### Phase 2: 核心功能 🚧
- [ ] 告警规则引擎
- [ ] 数据转发模块
- [ ] Grafana 集成
- [ ] 更多插件
- [ ] 完善 Web UI

### Phase 3: 高级功能 📋
- [ ] 集群部署支持
- [ ] 自定义 Dashboard
- [ ] 插件市场
- [ ] 移动端 App

### Phase 4: 智能化 🔮
- [ ] AI 智能告警
- [ ] 拓扑自动发现
- [ ] ChatOps 集成
- [ ] 多租户支持

## 🤝 参与贡献

我们欢迎任何形式的贡献：

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔌 开发插件
- 💻 提交代码

### 贡献指南
1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📮 联系我们

- **官网**: https://celestial.io
- **文档**: https://docs.celestial.io
- **GitHub**: https://github.com/celestial/celestial
- **社区**: https://community.celestial.io
- **邮箱**: hello@celestial.io

## 📄 许可证

本项目采用 Apache 2.0 许可证。详情请参阅 [LICENSE](../LICENSE) 文件。

## 🙏 致谢

感谢以下开源项目：

- [Prometheus](https://prometheus.io/) - 监控和告警工具
- [VictoriaMetrics](https://victoriametrics.com/) - 时序数据库
- [Grafana](https://grafana.com/) - 可视化平台
- [PostgreSQL](https://www.postgresql.org/) - 关系数据库
- [Redis](https://redis.io/) - 内存数据库
- [Gin](https://gin-gonic.com/) - Go Web 框架

---

**注意**: 本项目目前处于设计阶段，文档已完成，代码实现正在进行中。欢迎 Star ⭐ 关注项目进展！

## 📚 文档版本

- **版本**: v1.0.0
- **更新日期**: 2025-11-01
- **状态**: 设计完成，待实现

---

Made with ❤️ by Celestial Team

