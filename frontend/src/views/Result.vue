<template>
  <div class="result-container">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <el-icon :size="24" color="#409eff"><Collection /></el-icon>
          <span class="brand-text">评茶员初赛理论题库</span>
        </div>
        <div class="user-info">
          <el-icon><User /></el-icon>
          <span>{{ userStore.userName }}</span>
          <el-button type="danger" link @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出
          </el-button>
        </div>
      </div>
    </header>

    <!-- 结果内容 -->
    <main class="result-content">
      <div class="result-card">
        <div class="result-header">
          <div class="success-icon">
            <el-icon :size="64" color="#67c23a"><CircleCheckFilled /></el-icon>
          </div>
          <h1 class="result-title">考试完成</h1>
          <p class="result-subtitle">您已完成本次考试，以下是您的成绩</p>
        </div>

        <!-- 成绩展示 -->
        <div class="score-section">
          <div class="score-circle">
            <div class="score-value">{{ result.total_score }}</div>
            <div class="score-total">/ {{ totalScore }}</div>
          </div>
          <div class="accuracy-rate">
            <span class="rate-label">正确率</span>
            <span class="rate-value">{{ result.accuracy_rate?.toFixed(2) }}%</span>
          </div>
        </div>

        <!-- 详细统计 -->
        <div class="stats-grid">
          <div class="stat-item correct">
            <div class="stat-icon">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ result.correct_count }}</div>
              <div class="stat-label">正确数量</div>
            </div>
          </div>

          <div class="stat-item wrong">
            <div class="stat-icon">
              <el-icon><CircleClose /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ result.wrong_count }}</div>
              <div class="stat-label">错误数量</div>
            </div>
          </div>

          <div class="stat-item time">
            <div class="stat-icon">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ formattedDuration }}</div>
              <div class="stat-label">总耗时</div>
            </div>
          </div>

          <div class="stat-item date">
            <div class="stat-icon">
              <el-icon><Calendar /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ formattedDate }}</div>
              <div class="stat-label">完成时间</div>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-section">
          <el-button
            type="primary"
            size="large"
            class="action-btn"
            @click="backToWelcome"
          >
            <el-icon><HomeFilled /></el-icon>
            返回首页
          </el-button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Collection, User, SwitchButton, CircleCheckFilled,
  CircleCheck, CircleClose, Timer, Calendar, HomeFilled
} from '@element-plus/icons-vue'
import { examApi } from '../api'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const result = ref({})
const totalScore = ref(500)

// 格式化时长
const formattedDuration = computed(() => {
  const seconds = result.value.duration_seconds || 0
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}小时${minutes}分${secs}秒`
  } else if (minutes > 0) {
    return `${minutes}分${secs}秒`
  } else {
    return `${secs}秒`
  }
})

// 格式化日期
const formattedDate = computed(() => {
  if (!result.value.end_time) return '-'
  const date = new Date(result.value.end_time)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
})

// 获取考试结果
const loadResult = async () => {
  const examId = route.params.examId
  if (!examId) {
    ElMessage.error('考试ID无效')
    router.push('/welcome')
    return
  }

  try {
    const res = await examApi.getExamResult(examId)
    result.value = res
    totalScore.value = res.total_score + res.wrong_count
  } catch (error) {
    console.error('获取考试结果失败:', error)
  }
}

// 返回首页
const backToWelcome = () => {
  router.push('/welcome')
}

// 退出登录
const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(() => {
  loadResult()
})
</script>

<style scoped>
.result-container {
  min-height: 100vh;
  background: #f5f7fa;
}

/* 顶部导航 */
.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.brand-text {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #606266;
}

/* 结果内容 */
.result-content {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 24px;
}

.result-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  padding: 48px;
}

.result-header {
  text-align: center;
  margin-bottom: 40px;
}

.success-icon {
  margin-bottom: 16px;
}

.result-title {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.result-subtitle {
  font-size: 16px;
  color: #909399;
}

/* 成绩展示 */
.score-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 40px;
}

.score-circle {
  width: 160px;
  height: 160px;
  border-radius: 50%;
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(103, 194, 58, 0.3);
  margin-bottom: 16px;
}

.score-value {
  font-size: 48px;
  font-weight: 700;
  color: #fff;
  line-height: 1;
}

.score-total {
  font-size: 20px;
  color: rgba(255, 255, 255, 0.8);
}

.accuracy-rate {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.rate-label {
  font-size: 14px;
  color: #909399;
}

.rate-value {
  font-size: 24px;
  font-weight: 600;
  color: #67c23a;
}

/* 详细统计 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 40px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 12px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-item.correct .stat-icon {
  background: #e8f5e9;
  color: #4caf50;
}

.stat-item.wrong .stat-icon {
  background: #ffebee;
  color: #f44336;
}

.stat-item.time .stat-icon {
  background: #e3f2fd;
  color: #2196f3;
}

.stat-item.date .stat-icon {
  background: #f3e5f5;
  color: #9c27b0;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

/* 操作按钮 */
.action-section {
  text-align: center;
}

.action-btn {
  width: 200px;
  height: 52px;
  font-size: 18px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}
</style>
