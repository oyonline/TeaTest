import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useAdminStore } from '../stores/admin'

const routes = [
  {
    path: '/',
    name: 'UserLogin',
    component: () => import('../views/UserLogin.vue'),
    meta: { guest: true }
  },
  {
    path: '/welcome',
    name: 'Welcome',
    component: () => import('../views/Welcome.vue'),
    meta: { requiresUser: true }
  },
  {
    path: '/exam',
    name: 'Exam',
    component: () => import('../views/Exam.vue'),
    meta: { requiresUser: true }
  },
  {
    path: '/result/:examId',
    name: 'Result',
    component: () => import('../views/Result.vue'),
    meta: { requiresUser: true }
  },
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('../views/AdminLogin.vue'),
    meta: { guest: true }
  },
  {
    path: '/admin',
    name: 'AdminDashboard',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const adminStore = useAdminStore()

  // 需要用户登录的页面
  if (to.meta.requiresUser) {
    if (!userStore.token) {
      next('/')
      return
    }
  }

  // 需要管理员登录的页面
  if (to.meta.requiresAdmin) {
    if (!adminStore.token) {
      next('/admin/login')
      return
    }
  }

  // 游客页面（已登录则跳转）
  if (to.meta.guest) {
    if (userStore.token && to.path === '/') {
      next('/welcome')
      return
    }
    if (adminStore.token && to.path === '/admin/login') {
      next('/admin')
      return
    }
  }

  next()
})

export default router
