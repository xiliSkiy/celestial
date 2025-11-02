import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'devices',
        name: 'Devices',
        component: () => import('@/views/Devices/List.vue'),
        meta: { title: '设备管理' }
      },
      {
        path: 'devices/:id',
        name: 'DeviceDetail',
        component: () => import('@/views/Devices/Detail.vue'),
        meta: { title: '设备详情' }
      },
      {
        path: 'sentinels',
        name: 'Sentinels',
        component: () => import('@/views/Sentinels/List.vue'),
        meta: { title: 'Sentinel 管理' }
      },
      {
        path: 'sentinels/:id',
        name: 'SentinelDetail',
        component: () => import('@/views/Sentinels/Detail.vue'),
        meta: { title: 'Sentinel 详情' }
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('@/views/Tasks/List.vue'),
        meta: { title: '任务管理' }
      },
      {
        path: 'alerts',
        name: 'Alerts',
        component: () => import('@/views/Alerts/Index.vue'),
        meta: { title: '告警管理' }
      },
      {
        path: 'forwarders',
        name: 'Forwarders',
        component: () => import('@/views/Forwarders/List.vue'),
        meta: { title: '数据转发' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings/Index.vue'),
        meta: { title: '系统设置' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const requiresAuth = to.meta.requiresAuth !== false

  if (requiresAuth && !userStore.token) {
    next('/login')
  } else if (to.path === '/login' && userStore.token) {
    next('/')
  } else {
    next()
  }
})

export default router

