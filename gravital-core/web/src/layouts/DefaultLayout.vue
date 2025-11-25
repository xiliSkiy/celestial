<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '200px'" class="layout-aside">
      <div class="logo-container">
        <el-icon v-if="!isCollapse" :size="32" class="logo-icon">
          <Monitor />
        </el-icon>
        <span v-if="!isCollapse" class="logo-text">Gravital Core</span>
      </div>
      
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapse"
        :unique-opened="true"
        router
        class="sidebar-menu"
      >
        <el-menu-item index="/">
          <el-icon><DataLine /></el-icon>
          <template #title>仪表盘</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('devices.read')" 
          index="/devices"
        >
          <el-icon><Monitor /></el-icon>
          <template #title>设备管理</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('sentinels.read')" 
          index="/sentinels"
        >
          <el-icon><Compass /></el-icon>
          <template #title>Sentinel 管理</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('tasks.read')" 
          index="/tasks"
        >
          <el-icon><List /></el-icon>
          <template #title>任务管理</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('alerts.read')" 
          index="/alerts"
        >
          <el-icon><Bell /></el-icon>
          <template #title>告警管理</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('topology.read')" 
          index="/topology"
        >
          <el-icon><Share /></el-icon>
          <template #title>网络拓扑</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('forwarders.read')" 
          index="/forwarders"
        >
          <el-icon><Connection /></el-icon>
          <template #title>数据转发</template>
        </el-menu-item>
        
        <el-menu-item 
          v-if="hasPermission('settings.read')" 
          index="/settings"
        >
          <el-icon><Setting /></el-icon>
          <template #title>系统设置</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-icon class="collapse-icon" @click="toggleCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="route.meta.title">
              {{ route.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        
        <div class="header-right">
          <el-icon class="header-icon" @click="toggleTheme">
            <Sunny v-if="isDark" />
            <Moon v-else />
          </el-icon>
          
          <el-badge :value="unreadCount" :hidden="unreadCount === 0" class="header-badge">
            <el-icon class="header-icon">
              <Bell />
            </el-icon>
          </el-badge>
          
          <el-dropdown @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32">
                {{ userStore.userInfo?.username?.charAt(0).toUpperCase() }}
              </el-avatar>
              <span class="username">{{ userStore.userInfo?.username }}</span>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容 -->
      <el-main class="layout-main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessageBox } from 'element-plus'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const isCollapse = ref(false)
// 默认使用日间模式（light mode）
const isDark = ref(false)
const unreadCount = ref(0)

const activeMenu = computed(() => {
  return route.path
})

const hasPermission = (permission: string) => {
  return userStore.hasPermission(permission)
}

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// 初始化主题
const initTheme = () => {
  // 从 localStorage 读取保存的主题偏好
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    isDark.value = savedTheme === 'dark'
  } else {
    // 如果没有保存的主题，默认使用日间模式
    isDark.value = false
  }
  // 应用主题
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  const theme = isDark.value ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', theme)
  // 保存主题偏好到 localStorage
  localStorage.setItem('theme', theme)
}

// 组件挂载时初始化主题
onMounted(() => {
  initTheme()
})

const handleCommand = (command: string) => {
  if (command === 'logout') {
    ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }).then(() => {
      userStore.logout()
    })
  } else if (command === 'profile') {
    router.push('/settings')
  }
}
</script>

<style scoped lang="scss">
.layout-container {
  width: 100%;
  height: 100vh;
}

.layout-aside {
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  transition: width 0.3s;
  overflow-x: hidden;

  .logo-container {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 20px;
    border-bottom: 1px solid var(--border-color);

    .logo-icon {
      margin-right: 10px;
      color: var(--color-primary);
    }

    .logo-text {
      font-size: 18px;
      font-weight: bold;
      color: var(--text-primary);
    }
  }

  .sidebar-menu {
    border-right: none;
    background: var(--bg-secondary);
  }
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  padding: 0 20px;

  .header-left {
    display: flex;
    align-items: center;
    gap: 20px;

    .collapse-icon {
      font-size: 20px;
      cursor: pointer;
      
      &:hover {
        color: var(--color-primary);
      }
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .header-icon {
      font-size: 20px;
      cursor: pointer;
      
      &:hover {
        color: var(--color-primary);
      }
    }

    .header-badge {
      cursor: pointer;
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 10px;
      cursor: pointer;

      .username {
        color: var(--text-primary);
      }
    }
  }
}

.layout-main {
  background: var(--bg-primary);
  padding: 20px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

