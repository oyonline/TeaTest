import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'
import { useAdminStore } from '../stores/admin'
import { entryFromRoute, entryHome, entryLogin } from '../utils/entry'

const routes = [
  {
    path: '/',
    name: 'UserLogin',
    component: () => import('../views/UserLogin.vue'),
    meta: { guest: true, entry: 'tea' }
  },
  {
    path: '/practice',
    name: 'PracticeLogin',
    component: () => import('../views/UserLogin.vue'),
    meta: { guest: true, entry: 'practice' }
  },
  {
    path: '/welcome',
    name: 'Welcome',
    component: () => import('../views/Welcome.vue'),
    meta: { requiresUser: true, entry: 'tea' }
  },
  {
    path: '/practice/banks',
    name: 'PracticeWelcome',
    component: () => import('../views/Welcome.vue'),
    meta: { requiresUser: true, entry: 'practice' }
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
    redirect: '/admin/question-banks',
    meta: { requiresAdmin: true }
  },
  {
    path: '/admin/question-banks',
    name: 'AdminQuestionBanks',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true, adminSection: 'banks' }
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true, adminSection: 'users' }
  },
  {
    path: '/admin/import',
    name: 'AdminImport',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true, adminSection: 'import' }
  },
  {
    path: '/admin/questions',
    name: 'AdminQuestions',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true, adminSection: 'questions' }
  },
  {
    path: '/admin/records',
    name: 'AdminRecords',
    component: () => import('../views/AdminDashboard.vue'),
    meta: { requiresAdmin: true, adminSection: 'records' }
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
      next(entryLogin(entryFromRoute(to)))
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
    if (userStore.token && (to.path === '/' || to.path === '/practice')) {
      next(entryHome(entryFromRoute(to)))
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
