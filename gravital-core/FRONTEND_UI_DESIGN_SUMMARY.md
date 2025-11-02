# Gravital Core 前端 UI 设计总结

## 📋 完成内容

我已经为 Gravital Core 创建了一套完整的前端 UI 设计方案，包括：

### 1. 设计文档

#### ✅ 详细设计方案 (`docs/07-前端UI设计方案.md`)
- **设计理念**: 简洁高效、数据可视化、响应式设计、暗色主题
- **技术栈**: Vue 3 + TypeScript + Element Plus + ECharts
- **整体布局**: Header + Sidebar + Main Content
- **配色方案**: 暗色/亮色主题切换
- **页面设计**: 11 个核心页面的详细设计
  - 仪表盘
  - 设备管理（列表、详情、添加/编辑）
  - Sentinel 管理（列表、详情）
  - 任务管理（列表、创建/编辑）
  - 告警管理（规则、事件、创建规则）
  - 数据转发管理（列表、详情）
  - 系统设置（基本设置、用户管理）
- **组件设计**: 通用组件 + 业务组件
- **状态管理**: Pinia Store 结构
- **API 接口封装**: 完整的 API 模块设计
- **实时通信**: WebSocket 集成方案
- **响应式设计**: 移动端适配
- **性能优化**: 7 大优化策略
- **部署方案**: Docker + Nginx
- **开发规范**: 目录结构、命名规范、Git 规范

#### ✅ 快速启动指南 (`docs/08-前端UI快速启动指南.md`)
- UI 预览和说明
- 快速开始步骤
- 项目结构详解
- 主题配置
- API 集成示例
- 图表集成示例
- WebSocket 集成示例
- 响应式设计示例
- 权限控制示例
- 部署指南
- 常见问题解答

### 2. 项目脚手架

#### ✅ 基础配置文件
- `web/package.json` - 项目依赖和脚本
- `web/vite.config.ts` - Vite 构建配置
- `web/tsconfig.json` - TypeScript 配置
- `web/index.html` - HTML 模板
- `web/README.md` - 项目说明

## 🎨 设计亮点

### 1. 现代化技术栈
- **Vue 3**: 最新的 Vue 框架，性能优异
- **TypeScript**: 类型安全，提高代码质量
- **Element Plus**: 成熟的企业级 UI 组件库
- **ECharts**: 强大的数据可视化能力
- **Vite**: 极速的开发体验

### 2. 用户体验优化
- **暗色主题**: 适合长时间监控，减少视觉疲劳
- **实时更新**: WebSocket 推送，数据实时刷新
- **响应式设计**: 支持桌面端和移动端
- **直观的可视化**: 丰富的图表和数据展示

### 3. 开发体验
- **组件化开发**: 通用组件 + 业务组件
- **自动导入**: Element Plus 组件自动导入
- **类型安全**: 完整的 TypeScript 类型定义
- **代码规范**: ESLint + Prettier

### 4. 性能优化
- **懒加载**: 路由和组件按需加载
- **虚拟滚动**: 大列表性能优化
- **代码分割**: 第三方库分离打包
- **缓存策略**: API 响应缓存

## 📊 页面功能概览

### 仪表盘
- 4 个统计卡片（设备总数、在线设备、活跃告警、任务数）
- 4 个图表区域（设备状态分布、告警趋势、Sentinel 状态、数据转发统计）
- 最近活动时间线

### 设备管理
- 设备列表（搜索、筛选、分页）
- 设备分组树
- 设备详情（基本信息、监控指标、采集任务、告警规则）
- 添加/编辑设备对话框

### Sentinel 管理
- Sentinel 卡片网格
- Sentinel 详情（系统信息、性能监控、设备列表、任务列表）

### 任务管理
- 任务列表
- 任务执行统计
- 创建/编辑任务对话框

### 告警管理
- 告警规则列表
- 告警事件时间线
- 告警统计图表
- 创建告警规则对话框

### 数据转发管理
- 转发器卡片
- 转发器详情（统计信息、配置、日志）
- 缓冲区状态

### 系统设置
- 基本设置
- 用户管理
- 角色权限
- 系统日志

## 🚀 快速开始

### 1. 安装依赖

```bash
cd gravital-core/web
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 3. 构建生产版本

```bash
npm run build
```

## 📁 项目结构

```
web/
├── src/
│   ├── api/                 # API 接口
│   ├── assets/              # 静态资源
│   ├── components/          # 组件
│   │   ├── common/          # 通用组件
│   │   └── business/        # 业务组件
│   ├── composables/         # 组合式函数
│   ├── layouts/             # 布局组件
│   ├── router/              # 路由配置
│   ├── stores/              # 状态管理
│   ├── types/               # TypeScript 类型
│   ├── utils/               # 工具函数
│   ├── views/               # 页面组件
│   ├── App.vue
│   └── main.ts
├── public/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md
```

## 🎯 核心组件

### 通用组件
- **StatCard**: 统计卡片
- **ChartCard**: 图表卡片
- **DataTable**: 数据表格
- **StatusBadge**: 状态徽章

### 业务组件
- **DeviceCard**: 设备卡片
- **SentinelCard**: Sentinel 卡片
- **AlertTimeline**: 告警时间线
- **TaskList**: 任务列表

## 🔌 集成方案

### API 集成
```typescript
// 统一的 Axios 封装
import request from '@/api/request'

export const deviceApi = {
  getDevices: (params) => request.get('/api/v1/devices', { params }),
  createDevice: (data) => request.post('/api/v1/devices', data)
}
```

### WebSocket 集成
```typescript
// 实时数据推送
import { io } from 'socket.io-client'

const socket = io(VITE_WS_URL)
socket.on('device_status_change', handleDeviceChange)
socket.on('alert_triggered', handleAlert)
```

### 图表集成
```typescript
// ECharts 图表
import * as echarts from 'echarts'

const chart = echarts.init(chartRef.value, 'dark')
chart.setOption(option)
```

## 📱 响应式设计

### 断点定义
- **xs**: 0-576px (手机)
- **sm**: 576-768px (平板竖屏)
- **md**: 768-992px (平板横屏)
- **lg**: 992-1200px (笔记本)
- **xl**: 1200-1600px (桌面)
- **xxl**: 1600px+ (大屏)

### 移动端适配
- 侧边栏折叠
- 统计卡片单列显示
- 图表自适应
- 表格横向滚动

## 🎨 主题系统

### 暗色主题（默认）
- 背景: #1a1a1a, #2d2d2d, #3d3d3d
- 文本: #ffffff, #b3b3b3
- 主题色: #409eff

### 亮色主题
- 背景: #ffffff, #f5f7fa
- 文本: #303133, #606266
- 主题色: #409eff

## 🔐 权限控制

### 路由权限
```typescript
router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else {
    next()
  }
})
```

### 按钮权限
```vue
<el-button v-if="hasPermission('devices.write')">
  创建设备
</el-button>
```

## 🚢 部署方案

### Docker 部署
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
```

### Nginx 配置
```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api/ {
        proxy_pass http://gravital-core:8080;
    }
}
```

## 📈 性能优化

1. **路由懒加载**: 按需加载页面组件
2. **组件懒加载**: 大组件异步加载
3. **虚拟滚动**: 大列表性能优化
4. **防抖节流**: 搜索、resize 事件
5. **API 缓存**: 减少重复请求
6. **代码分割**: 第三方库分离
7. **图片优化**: WebP 格式、懒加载

## 🔄 开发流程

### 1. 创建新页面
```bash
# 1. 创建页面组件
src/views/NewPage.vue

# 2. 添加路由
src/router/index.ts

# 3. 创建 API 接口
src/api/newpage.ts

# 4. 创建 Store（如需要）
src/stores/newpage.ts
```

### 2. 创建新组件
```bash
# 通用组件
src/components/common/NewComponent.vue

# 业务组件
src/components/business/NewComponent.vue
```

### 3. 添加新 API
```typescript
// src/api/newapi.ts
export const newApi = {
  getData: () => request.get('/api/v1/data'),
  createData: (data) => request.post('/api/v1/data', data)
}
```

## 📚 文档索引

- [详细设计方案](../docs/07-前端UI设计方案.md) - 完整的 UI 设计文档
- [快速启动指南](../docs/08-前端UI快速启动指南.md) - 快速上手指南
- [项目 README](web/README.md) - 项目基本信息

## 🎯 下一步计划

### Phase 1: 基础搭建（1-2 周）
- [ ] 搭建项目框架
- [ ] 实现登录/认证
- [ ] 实现主布局
- [ ] 实现仪表盘

### Phase 2: 核心功能（2-3 周）
- [ ] 实现设备管理
- [ ] 实现 Sentinel 管理
- [ ] 实现任务管理
- [ ] 实现告警管理

### Phase 3: 高级功能（1-2 周）
- [ ] 实现数据转发管理
- [ ] 实现系统设置
- [ ] 实现实时通信
- [ ] 性能优化

### Phase 4: 完善优化（1 周）
- [ ] 移动端适配
- [ ] 国际化支持
- [ ] 单元测试
- [ ] 文档完善

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

### 提交规范
```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
perf: 性能优化
test: 测试相关
chore: 构建/工具链相关
```

## 📄 许可证

MIT License

---

**设计完成日期**: 2025-11-02  
**设计版本**: v1.0.0  
**设计者**: AI Assistant

**🎉 前端 UI 设计方案已完成，可以开始开发了！**

