<template>
  <div class="welcome-container">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <el-icon :size="28" color="#409eff"><Collection /></el-icon>
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

    <!-- 主内容 -->
    <main class="main-content">
      <div class="welcome-card">
        <div class="welcome-header">
          <h1 class="welcome-title">
            <span v-if="hasInProgress">欢迎回来，{{ userStore.userName }}</span>
            <span v-else>欢迎，{{ userStore.userName }}</span>
          </h1>
          <p class="welcome-subtitle">
            <span v-if="hasInProgress">您有一场正在进行的考试，点击"继续答题"继续</span>
            <span v-else>准备好开始您的考试了吗？</span>
          </p>
        </div>

        <!-- 考试统计 -->
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon blue">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ totalQuestions }}</div>
              <div class="stat-label">题库总量</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon green">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ completedExams }}</div>
              <div class="stat-label">已完成考试</div>
            </div>
          </div>

          <div v-if="hasInProgress" class="stat-card">
            <div class="stat-icon orange">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ answeredCount }}</div>
              <div class="stat-label">已答题数</div>
            </div>
          </div>
        </div>

        <!-- 操作按钮 -->
        <div class="action-section">
          <el-button
            v-if="hasInProgress"
            type="primary"
            size="large"
            class="action-btn"
            @click="continueExam"
          >
            <el-icon><VideoPlay /></el-icon>
            继续答题
          </el-button>

          <el-button
            v-else
            type="primary"
            size="large"
            class="action-btn"
            @click="startNewExam"
          >
            <el-icon><VideoPlay /></el-icon>
            开始答题
          </el-button>
        </div>

        <!-- 考试规则说明 -->
        <div class="rules-section">
          <h3 class="rules-title">
            <el-icon><InfoFilled /></el-icon>
            考试规则
          </h3>
          <div class="rules-list">
            <div class="rule-item">
              <el-icon class="rule-icon"><Check /></el-icon>
              <span>共 {{ totalQuestions }} 道单选题，每题 1 分，满分 {{ totalQuestions }} 分</span>
            </div>
            <div class="rule-item">
              <el-icon class="rule-icon"><Check /></el-icon>
              <span>不限制总答题时长，可随时中断并继续</span>
            </div>
            <div class="rule-item">
              <el-icon class="rule-icon"><Check /></el-icon>
              <span>一旦选择答案，本题立即锁定，不能再修改</span>
            </div>
            <div class="rule-item">
              <el-icon class="rule-icon"><Check /></el-icon>
              <span>答完所有题目后自动提交，无需手动交卷</span>
            </div>
            <div class="rule-item">
              <el-icon class="rule-icon"><Check /></el-icon>
              <span>支持分页查看和跳转到指定页面</span>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Collection, User, SwitchButton, Document,
  CircleCheck, Timer, VideoPlay, InfoFilled, Check
} from '@element-plus/icons-vue'
import { examApi } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()

const hasInProgress = ref(false)
const totalQuestions = ref(500)
const completedExams = ref(0)
const answeredCount = ref(0)
const currentExamId = ref(null)

// 获取考试统计
const loadStats = async () => {
  try {
    const res = await examApi.getExamStats()
    totalQuestions.value = res.total_questions || 500
    completedExams.value = res.completed_exams || 0
    hasInProgress.value = res.has_in_progress || false

    if (hasInProgress.value) {
      answeredCount.value = res.answered_count || 0
      currentExamId.value = res.current_exam_id
    }
  } catch (error) {
    console.error('获取统计失败:', error)
  }
}

// 开始新考试
const startNewExam = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要开始新的考试吗？',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    const res = await examApi.startExam()
    ElMessage.success('考试开始')
    router.push({
      path: '/exam',
      query: { exam_id: res.exam_id }
    })
  } catch (error) {
    if (error !== 'cancel') {
      console.error('开始考试失败:', error)
    }
  }
}

// 继续考试
const continueExam = () => {
  router.push({
    path: '/exam',
    query: { exam_id: currentExamId.value }
  })
}

// 退出登录
const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.welcome-container {
  min-height: 100vh;
  background: #f5f7fa;
}

/* 顶部导航 */
.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 100;
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

/* 主内容 */
.main-content {
  max-width: 800px;
  margin: 0 auto;
  padding: 40px 24px;
}

.welcome-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  padding: 48px;
}

.welcome-header {
  text-align: center;
  margin-bottom: 40px;
}

.welcome-title {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.welcome-subtitle {
  font-size: 16px;
  color: #606266;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card {
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

.stat-icon.blue {
  background: #e3f2fd;
  color: #1976d2;
}

.stat-icon.green {
  background: #e8f5e9;
  color: #388e3c;
}

.stat-icon.orange {
  background: #fff3e0;
  color: #f57c00;
}

.stat-value {
  font-size: 24px;
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
  margin-bottom: 40px;
}

.action-btn {
  width: 200px;
  height: 52px;
  font-size: 18px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
  transition: all 0.3s ease;
}

.action-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(64, 158, 255, 0.4);
}

/* 规则说明 */
.rules-section {
  background: #f5f7fa;
  border-radius: 12px;
  padding: 24px;
}

.rules-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.rule-item {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #606266;
  font-size: 14px;
}

.rule-icon {
  color: #67c23a;
  font-size: 16px;
}
</style>
