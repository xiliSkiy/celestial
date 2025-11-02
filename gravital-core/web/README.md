# Gravital Core Web UI

基于 Vue 3 + TypeScript + Element Plus 的现代化监控管理平台前端。

## 技术栈

- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **UI 组件库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **图表库**: Apache ECharts
- **HTTP 客户端**: Axios
- **实时通信**: Socket.io
- **工具库**: VueUse, Day.js

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发环境

```bash
npm run dev
```

访问 http://localhost:5173

### 生产构建

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 项目结构

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

## 功能模块

### 已实现
- [ ] 登录/认证
- [ ] 仪表盘
- [ ] 设备管理
- [ ] Sentinel 管理
- [ ] 任务管理
- [ ] 告警管理
- [ ] 数据转发管理
- [ ] 系统设置

### 计划中
- [ ] 实时数据推送
- [ ] 自定义 Dashboard
- [ ] 移动端适配
- [ ] 国际化支持

## 开发规范

### 命名规范

- 组件: PascalCase (DeviceCard.vue)
- 文件夹: kebab-case (device-management/)
- 变量: camelCase (deviceList)
- 常量: UPPER_SNAKE_CASE (API_BASE_URL)
- CSS 类: kebab-case (device-card)

### Git 提交规范

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

## API 配置

开发环境下，API 请求会自动代理到 `http://localhost:8080`。

如需修改，请编辑 `vite.config.ts` 中的 proxy 配置。

## 环境变量

创建 `.env.local` 文件配置本地环境变量：

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## 浏览器支持

- Chrome >= 87
- Firefox >= 78
- Safari >= 14
- Edge >= 88

## 许可证

MIT

