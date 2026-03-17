import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { useAdminStore } from '../stores/admin'

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
    const token = isAdminApi ? (adminStore.token || userStore.token) : (userStore.token || adminStore.token)
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

      // 登录过期
      if (data.code === 401) {
        const userStore = useUserStore()
        const adminStore = useAdminStore()
        userStore.logout()
        adminStore.logout()
        window.location.href = '/'
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
