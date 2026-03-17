<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <div class="logo">
          <el-icon :size="48" color="#409eff"><Setting /></el-icon>
        </div>
        <h1 class="title">评茶员初赛理论题库</h1>
        <p class="subtitle">管理员登录</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入管理员密码"
            size="large"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            class="login-btn"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="login-footer">
        <router-link to="/" class="back-link">
          <el-icon><ArrowLeft /></el-icon>
          返回答题入口
        </router-link>
      </div>
    </div>

    <div class="decoration">
      <div class="tea-leaf leaf-1"></div>
      <div class="tea-leaf leaf-2"></div>
      <div class="tea-leaf leaf-3"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Lock, Setting, ArrowLeft } from '@element-plus/icons-vue'
import { authApi } from '../api'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const adminStore = useAdminStore()
const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  password: ''
})

const rules = {
  password: [
    { required: true, message: '请输入管理员密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await authApi.adminLogin(form)
      adminStore.setToken(res.token)
      adminStore.setAdminInfo({ role: res.role })

      ElMessage.success('登录成功')
      router.push('/admin')
    } catch (error) {
      console.error('登录失败:', error)
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f3e5f5 0%, #e1bee7 50%, #ce93d8 100%);
  position: relative;
  overflow: hidden;
}

.login-box {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  padding: 48px 40px;
  width: 100%;
  max-width: 420px;
  position: relative;
  z-index: 1;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  margin-bottom: 16px;
}

.title {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.subtitle {
  font-size: 14px;
  color: #909399;
}

.login-form {
  margin-top: 24px;
}

.login-btn {
  width: 100%;
  height: 44px;
  font-size: 16px;
  border-radius: 8px;
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #909399;
  font-size: 14px;
  text-decoration: none;
  transition: color 0.3s;
}

.back-link:hover {
  color: #409eff;
}

/* 装饰性元素 */
.decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  overflow: hidden;
}

.tea-leaf {
  position: absolute;
  width: 60px;
  height: 60px;
  background: rgba(156, 39, 176, 0.1);
  border-radius: 50% 0 50% 50%;
  transform: rotate(-45deg);
}

.leaf-1 {
  top: 10%;
  right: 10%;
  animation: float 6s ease-in-out infinite;
}

.leaf-2 {
  top: 20%;
  left: 15%;
  width: 40px;
  height: 40px;
  animation: float 8s ease-in-out infinite 1s;
}

.leaf-3 {
  bottom: 15%;
  right: 20%;
  width: 50px;
  height: 50px;
  animation: float 7s ease-in-out infinite 2s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(-45deg);
  }
  50% {
    transform: translateY(-20px) rotate(-45deg);
  }
}
</style>
