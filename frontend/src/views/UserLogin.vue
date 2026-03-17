<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <div class="logo">
          <el-icon :size="48" color="#409eff"><TeaIcon /></el-icon>
        </div>
        <h1 class="title">评茶员初赛理论题库</h1>
        <p class="subtitle">答题用户登录</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="login-form"
        @keyup.enter="handleLogin"
      >
        <el-form-item prop="name">
          <el-input
            v-model="form.name"
            placeholder="请输入姓名"
            size="large"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
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

    </div>

    <div class="decoration">
      <svg class="tea-leaf leaf-1" viewBox="0 0 1200 1200" preserveAspectRatio="xMidYMid meet">
        <path fill="rgba(64, 158, 255, 0.1)" d="M131.652,839.52c-7.46-17.981-16.747-35.855-19.464-55.253C76.402,534.149,280.466,331.362,465.682,254.339c185.214-77.022,398.471,45.498,524.903-86.019c31.587-35.667,63.736-69.443,111.762-69.694c20.018,0.993,36.688,10.401,45.207,26.999c158.034,323.688-67.753,707.493-322.728,843.865c-161.812,79.98-329.249,97.576-485.976,43.323c-43.582-10.338-78.515-55.918-123.691-56.51c-31.154,17.48-57.029,74.434-78.17,105.797c-24.625,42.738-73.658,54.038-108.937,16.64C-74.847,977.334,138.995,887.333,131.652,839.52z M269.158,798.08c22.099,17.854,53.978,12.793,70.95-5.022C475.799,628.65,659.535,556.522,859.36,561.999c29.444,1.709,50.848-22.136,52.74-47.718c0.592-30.381-22.788-50.927-48.975-52.742c-241.513-13.425-451.044,93.065-598.993,264.964C245.394,749.903,248.796,781.038,269.158,798.08L269.158,798.08z"/>
      </svg>
      <svg class="tea-leaf leaf-2" viewBox="0 0 1200 1200" preserveAspectRatio="xMidYMid meet">
        <path fill="rgba(64, 158, 255, 0.08)" d="M131.652,839.52c-7.46-17.981-16.747-35.855-19.464-55.253C76.402,534.149,280.466,331.362,465.682,254.339c185.214-77.022,398.471,45.498,524.903-86.019c31.587-35.667,63.736-69.443,111.762-69.694c20.018,0.993,36.688,10.401,45.207,26.999c158.034,323.688-67.753,707.493-322.728,843.865c-161.812,79.98-329.249,97.576-485.976,43.323c-43.582-10.338-78.515-55.918-123.691-56.51c-31.154,17.48-57.029,74.434-78.17,105.797c-24.625,42.738-73.658,54.038-108.937,16.64C-74.847,977.334,138.995,887.333,131.652,839.52z M269.158,798.08c22.099,17.854,53.978,12.793,70.95-5.022C475.799,628.65,659.535,556.522,859.36,561.999c29.444,1.709,50.848-22.136,52.74-47.718c0.592-30.381-22.788-50.927-48.975-52.742c-241.513-13.425-451.044,93.065-598.993,264.964C245.394,749.903,248.796,781.038,269.158,798.08L269.158,798.08z"/>
      </svg>
      <svg class="tea-leaf leaf-3" viewBox="0 0 1200 1200" preserveAspectRatio="xMidYMid meet">
        <path fill="rgba(64, 158, 255, 0.1)" d="M131.652,839.52c-7.46-17.981-16.747-35.855-19.464-55.253C76.402,534.149,280.466,331.362,465.682,254.339c185.214-77.022,398.471,45.498,524.903-86.019c31.587-35.667,63.736-69.443,111.762-69.694c20.018,0.993,36.688,10.401,45.207,26.999c158.034,323.688-67.753,707.493-322.728,843.865c-161.812,79.98-329.249,97.576-485.976,43.323c-43.582-10.338-78.515-55.918-123.691-56.51c-31.154,17.48-57.029,74.434-78.17,105.797c-24.625,42.738-73.658,54.038-108.937,16.64C-74.847,977.334,138.995,887.333,131.652,839.52z M269.158,798.08c22.099,17.854,53.978,12.793,70.95-5.022C475.799,628.65,659.535,556.522,859.36,561.999c29.444,1.709,50.848-22.136,52.74-47.718c0.592-30.381-22.788-50.927-48.975-52.742c-241.513-13.425-451.044,93.065-598.993,264.964C245.394,749.903,248.796,781.038,269.158,798.08L269.158,798.08z"/>
      </svg>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { authApi } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref(null)
const loading = ref(false)

const form = reactive({
  name: '',
  password: ''
})

const rules = {
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await authApi.userLogin(form)
      userStore.setToken(res.token)
      userStore.setUserInfo(res.user)

      ElMessage.success('登录成功')
      router.push('/welcome')
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
  background: linear-gradient(135deg, #e8f5e9 0%, #c8e6c9 50%, #a5d6a7 100%);
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
  width: 80px;
  height: 80px;
}

.leaf-1 {
  top: 10%;
  left: 10%;
  animation: float 6s ease-in-out infinite;
}

.leaf-2 {
  top: 20%;
  right: 15%;
  width: 60px;
  height: 60px;
  animation: float 8s ease-in-out infinite 1s;
}

.leaf-3 {
  bottom: 15%;
  left: 20%;
  width: 70px;
  height: 70px;
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

/* 自定义图标 */
:deep(.tea-icon) {
  font-size: 48px;
}
</style>
