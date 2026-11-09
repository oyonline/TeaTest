import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { useAdminStore } from '../stores/admin'
import { entryFromLocation, entryLogin } from '../utils/entry'

// 创建 axios 实例
const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    // 获取 token
    const userStore = useUserStore()
    const adminStore = useAdminStore()

    // 根据请求路径判断使用哪个 token
    // 管理后台接口优先使用管理员 token
    const isAdminApi = config.url?.startsWith('/admin')
    // 用户接口绝不借用 admin token，避免固定管理员 ID 被当成答题人。
    const token = isAdminApi ? adminStore.token : userStore.token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    const { data } = response

    // 业务错误
    if (data.code !== 0) {
      ElMessage.error(data.message || '操作失败')

      // 登录接口失败只显示错误并保留表单；只有已持有身份的接口才清理对应 Token。
      if (data.code === 401) {
        const url = response.config?.url || ''
        const isLoginRequest = url.includes('/auth/user/login') || url.includes('/auth/admin/login')
        if (!isLoginRequest) {
          const userStore = useUserStore()
          const adminStore = useAdminStore()
          if (url.startsWith('/admin')) {
            adminStore.logout()
            if (!window.location.pathname.startsWith('/admin/login')) {
              window.location.href = '/admin/login'
            }
          } else {
            userStore.logout()
            window.location.href = entryLogin(entryFromLocation())
          }
        }
      }

      return Promise.reject(new Error(data.message))
    }

    return data.data
  },
  (error) => {
    const message = error.response?.data?.message || '网络错误'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request
