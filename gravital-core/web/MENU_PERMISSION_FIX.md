# 菜单权限控制修复说明

## 问题描述

admin 用户登录后，前端侧边栏只显示"仪表盘"一个菜单项，其他菜单（设备管理、Sentinel 管理等）都没有显示。

## 问题原因

在 `DefaultLayout.vue` 中，菜单项是硬编码的，没有根据用户权限动态显示。虽然 admin 用户拥有 `["*"]` 权限（表示所有权限），但菜单没有使用 `v-if` 指令来检查权限。

## 解决方案

### 1. 添加权限检查方法

在 `DefaultLayout.vue` 的 `<script setup>` 中添加 `hasPermission` 方法：

```typescript
const hasPermission = (permission: string) => {
  return userStore.hasPermission(permission)
}
```

### 2. 为菜单项添加权限控制

为每个菜单项添加 `v-if` 指令，根据权限动态显示：

```vue
<!-- 仪表盘 - 所有人可见 -->
<el-menu-item index="/">
  <el-icon><DataLine /></el-icon>
  <template #title>仪表盘</template>
</el-menu-item>

<!-- 设备管理 - 需要 devices.read 权限 -->
<el-menu-item 
  v-if="hasPermission('devices.read')" 
  index="/devices"
>
  <el-icon><Monitor /></el-icon>
  <template #title>设备管理</template>
</el-menu-item>

<!-- Sentinel 管理 - 需要 sentinels.read 权限 -->
<el-menu-item 
  v-if="hasPermission('sentinels.read')" 
  index="/sentinels"
>
  <el-icon><Compass /></el-icon>
  <template #title>Sentinel 管理</template>
</el-menu-item>

<!-- 任务管理 - 需要 tasks.read 权限 -->
<el-menu-item 
  v-if="hasPermission('tasks.read')" 
  index="/tasks"
>
  <el-icon><List /></el-icon>
  <template #title>任务管理</template>
</el-menu-item>

<!-- 告警管理 - 需要 alerts.read 权限 -->
<el-menu-item 
  v-if="hasPermission('alerts.read')" 
  index="/alerts"
>
  <el-icon><Bell /></el-icon>
  <template #title>告警管理</template>
</el-menu-item>

<!-- 数据转发 - 需要 forwarders.read 权限 -->
<el-menu-item 
  v-if="hasPermission('forwarders.read')" 
  index="/forwarders"
>
  <el-icon><Connection /></el-icon>
  <template #title>数据转发</template>
</el-menu-item>

<!-- 系统设置 - 需要 settings.read 权限 -->
<el-menu-item 
  v-if="hasPermission('settings.read')" 
  index="/settings"
>
  <el-icon><Setting /></el-icon>
  <template #title>系统设置</template>
</el-menu-item>
```

## 权限说明

### Admin 角色权限

Admin 用户拥有 `["*"]` 权限，表示拥有所有权限。`hasPermission` 方法会检查：

```typescript
const hasPermission = (permission: string) => {
  return permissions.value.includes(permission) || permissions.value.includes('*')
}
```

因此，admin 用户可以看到所有菜单项。

### 权限列表

| 菜单项 | 所需权限 | 说明 |
|--------|----------|------|
| 仪表盘 | 无 | 所有用户可见 |
| 设备管理 | `devices.read` | 查看设备 |
| Sentinel 管理 | `sentinels.read` | 查看 Sentinel |
| 任务管理 | `tasks.read` | 查看任务 |
| 告警管理 | `alerts.read` | 查看告警 |
| 数据转发 | `forwarders.read` | 查看转发器 |
| 系统设置 | `settings.read` | 查看设置 |

### 其他角色权限示例

**Operator 角色** (运维人员):
```json
{
  "permissions": [
    "devices.read",
    "devices.write",
    "sentinels.read",
    "tasks.read",
    "alerts.read"
  ]
}
```

**Viewer 角色** (只读用户):
```json
{
  "permissions": [
    "devices.read",
    "sentinels.read",
    "tasks.read",
    "alerts.read"
  ]
}
```

## 权限检查流程

1. **登录时**: 从后端获取用户信息和权限
   ```typescript
   permissions.value = res.user?.role?.permissions || []
   ```

2. **菜单渲染**: 使用 `v-if` 检查权限
   ```vue
   <el-menu-item v-if="hasPermission('devices.read')" index="/devices">
   ```

3. **权限判断**: 检查用户是否拥有特定权限或通配符权限
   ```typescript
   permissions.includes(permission) || permissions.includes('*')
   ```

## 测试

### 1. Admin 用户（应该看到所有菜单）

```bash
# 登录
用户名: admin
密码: admin123

# 应该看到的菜单:
- 仪表盘
- 设备管理
- Sentinel 管理
- 任务管理
- 告警管理
- 数据转发
- 系统设置
```

### 2. Operator 用户（部分菜单）

如果创建了 operator 用户，应该看到：
- 仪表盘
- 设备管理
- Sentinel 管理
- 任务管理
- 告警管理

### 3. Viewer 用户（只读菜单）

如果创建了 viewer 用户，应该看到：
- 仪表盘
- 设备管理（只读）
- Sentinel 管理（只读）
- 任务管理（只读）
- 告警管理（只读）

## 相关文件

- `web/src/layouts/DefaultLayout.vue` - 布局和菜单
- `web/src/stores/user.ts` - 用户状态和权限检查
- `migrations/001_init.up.sql` - 数据库角色定义

## 扩展权限

如果需要添加新的权限，需要：

1. **数据库**: 在 `roles` 表中添加权限
   ```sql
   UPDATE roles SET permissions = '["devices.read", "devices.write", "new.permission"]'::jsonb WHERE name = 'operator';
   ```

2. **菜单**: 在 `DefaultLayout.vue` 中添加权限检查
   ```vue
   <el-menu-item v-if="hasPermission('new.permission')" index="/new-page">
   ```

3. **路由守卫**: 在 `router/index.ts` 中添加路由级别的权限检查（可选）

## 注意事项

1. **前端权限仅用于 UI 显示**: 真正的权限控制在后端 API
2. **通配符权限**: `*` 表示拥有所有权限
3. **权限命名**: 使用 `resource.action` 格式（如 `devices.read`）
4. **菜单刷新**: 权限更新后需要重新登录才能生效

---

**修复日期**: 2025-11-02  
**状态**: ✅ 已修复

