# Logo 文件修复说明

## 问题描述

前端启动时报错：
```
Failed to resolve import "/logo.svg" from "src/layouts/DefaultLayout.vue". Does the file exist?
```

## 问题原因

`DefaultLayout.vue` 中引用了 `/logo.svg` 文件，但该文件不存在。

## 解决方案

### 方案 1: 使用 Element Plus 图标（已实施）

将 logo 图片改为使用 Element Plus 的图标组件：

```vue
<!-- 修改前 -->
<img v-if="!isCollapse" src="/logo.svg" alt="Logo" class="logo" />

<!-- 修改后 -->
<el-icon v-if="!isCollapse" :size="32" class="logo-icon">
  <Monitor />
</el-icon>
```

**优点**:
- 无需额外的图片文件
- 图标可以动态改变颜色
- 与 Element Plus 风格统一

### 方案 2: 创建 SVG Logo（已提供）

在 `public/logo.svg` 创建了一个简单的 SVG logo：

**设计说明**:
- 中心圆形代表 "Gravital Core"（引力核心）
- 外圈代表引力场
- 4个小圆点代表 "Orbital Sentinels"（轨道哨兵）
- 连接线表示数据流

**使用方式**:
```vue
<img src="/logo.svg" alt="Gravital Core" class="logo" />
```

## 文件位置

```
web/
├── public/
│   ├── logo.svg          # SVG logo（新增）
│   └── favicon.ico       # 网站图标
└── src/
    └── layouts/
        └── DefaultLayout.vue  # 已修改
```

## Logo 设计元素

### 颜色方案
- **主色**: #409eff (蓝色) - 代表科技感
- **辅色**: #67c23a (绿色) - 代表在线状态
- **警告色**: #e6a23c (橙色) - 代表警告
- **危险色**: #f56c6c (红色) - 代表告警

### 图形含义
1. **中心核心** - Gravital Core 中心端
2. **外圈** - 引力场/监控范围
3. **轨道点** - Orbital Sentinels 采集端
4. **连接线** - 数据流/心跳连接

## 自定义 Logo

如果需要使用自己的 logo，有以下选项：

### 1. 替换 SVG 文件
将你的 logo 文件放到 `public/logo.svg`

### 2. 使用 PNG/JPG
```vue
<img src="/logo.png" alt="Logo" class="logo" />
```
将图片文件放到 `public/` 目录

### 3. 使用其他图标
```vue
<el-icon :size="32">
  <YourIconName />
</el-icon>
```

查看可用图标: https://element-plus.org/zh-CN/component/icon.html

## 相关文件

- `web/src/layouts/DefaultLayout.vue` - 布局组件
- `web/public/logo.svg` - SVG logo
- `web/public/favicon.ico` - 网站图标

## 建议

1. **生产环境**: 建议使用专业设计的 logo
2. **尺寸**: 建议 logo 尺寸为 32x32 或 64x64 像素
3. **格式**: SVG 格式最佳（可缩放，文件小）
4. **颜色**: 确保在暗色和亮色主题下都清晰可见

---

**修复日期**: 2025-11-02  
**状态**: ✅ 已修复

