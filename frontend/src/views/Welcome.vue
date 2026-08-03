<template>
  <div class="welcome-container">
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <el-icon :size="28" color="#409eff"><Collection /></el-icon>
          <span class="brand-text">茶叶理论题库</span>
        </div>
        <div class="user-info">
          <el-icon><User /></el-icon>
          <span class="user-name">{{ userStore.userName }}</span>
          <el-button type="danger" link @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出
          </el-button>
        </div>
      </div>
    </header>

    <main class="main-content">
      <section class="welcome-header">
        <div>
          <h1>选择题库开始答题</h1>
          <p>每个题库会单独保存一场进行中的答题记录</p>
        </div>
        <div class="summary-card">
          <el-icon><CircleCheck /></el-icon>
          <div>
            <strong>{{ completedExams }}</strong>
            <span>已完成考试</span>
          </div>
        </div>
      </section>

      <el-alert
        v-if="legacyExam"
        class="legacy-alert"
        title="您还有一场历史未分类考试"
        type="warning"
        :closable="false"
        show-icon
      >
        <template #default>
          <div class="legacy-content">
            <span>已答 {{ legacyExam.answered_count || 0 }} / {{ legacyExam.total_questions || 0 }} 题</span>
            <el-button type="warning" size="small" @click="continueExam(legacyExam.exam_id)">继续历史考试</el-button>
          </div>
        </template>
      </el-alert>

      <div v-loading="loading" class="bank-sections">
        <section v-for="group in bankGroups" :key="group.code" class="bank-section">
          <div class="section-title">
            <div class="section-icon" :class="group.code">
              <el-icon><Trophy v-if="group.code === 'competition'" /><Medal v-else /></el-icon>
            </div>
            <div>
              <h2>{{ group.name }}</h2>
              <p>{{ group.code === 'competition' ? '竞赛专项练习' : '职业技能考级练习' }}</p>
            </div>
          </div>

          <div class="bank-grid">
            <article v-for="bank in group.banks" :key="bank.id" class="bank-card">
              <div class="bank-card-top">
                <h3>{{ bank.name }}</h3>
                <el-tag v-if="bank.status !== 1" type="info" size="small">已停用</el-tag>
                <el-tag v-else-if="bank.has_in_progress" type="warning" size="small">答题中</el-tag>
                <el-tag v-else type="success" size="small">可答题</el-tag>
              </div>

              <div class="bank-meta">
                <span><el-icon><Document /></el-icon>{{ bank.question_count }} 道题</span>
                <span v-if="bank.has_in_progress"><el-icon><Timer /></el-icon>已答 {{ bank.answered_count }} 题</span>
              </div>

              <el-progress
                v-if="bank.has_in_progress && bank.question_count > 0"
                :percentage="progressOf(bank)"
                :stroke-width="8"
                class="bank-progress"
              />

              <el-button
                v-if="bank.has_in_progress"
                type="primary"
                class="bank-action"
                @click="continueExam(bank.current_exam_id)"
              >
                继续答题
              </el-button>
              <el-button
                v-else
                type="primary"
                class="bank-action"
                :disabled="bank.status !== 1 || bank.question_count === 0"
                @click="startNewExam(bank)"
              >
                {{ bank.question_count === 0 ? '暂无题目' : '开始答题' }}
              </el-button>
            </article>
          </div>
        </section>

        <el-empty v-if="!loading && banks.length === 0" description="暂无可用题库，请联系管理员" />
      </div>

      <section class="rules-section">
        <h3><el-icon><InfoFilled /></el-icon>答题规则</h3>
        <div class="rules-grid">
          <span><el-icon><Check /></el-icon>进入题库后回答该题库的全部题目</span>
          <span><el-icon><Check /></el-icon>不限时，可随时中断并继续</span>
          <span><el-icon><Check /></el-icon>答案提交后立即锁定</span>
          <span><el-icon><Check /></el-icon>答完全部题目后自动结算</span>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Check, CircleCheck, Collection, Document, InfoFilled,
  Medal, SwitchButton, Timer, Trophy, User
} from '@element-plus/icons-vue'
import { examApi } from '../api'
import { useUserStore } from '../stores/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(true)
const banks = ref([])
const completedExams = ref(0)
const legacyExam = ref(null)

const bankGroups = computed(() => [
  {
    code: 'competition',
    name: '竞赛题库',
    banks: banks.value.filter(bank => bank.category_code === 'competition')
  },
  {
    code: 'certification',
    name: '考级题库',
    banks: banks.value.filter(bank => bank.category_code === 'certification')
  }
].filter(group => group.banks.length > 0))

const progressOf = (bank) => {
  if (!bank.question_count) return 0
  return Math.min(100, Math.round((bank.answered_count / bank.question_count) * 100))
}

const loadData = async () => {
  loading.value = true
  try {
    const [bankList, stats] = await Promise.all([
      examApi.getQuestionBanks(),
      examApi.getExamStats()
    ])
    banks.value = bankList || []
    completedExams.value = stats.completed_exams || 0
    legacyExam.value = stats.legacy_in_progress || null
  } catch (error) {
    console.error('加载题库失败:', error)
  } finally {
    loading.value = false
  }
}

const startNewExam = async (bank) => {
  try {
    await ElMessageBox.confirm(
      `确定开始“${bank.name}”吗？本题库共 ${bank.question_count} 道题。`,
      '开始答题',
      { confirmButtonText: '开始', cancelButtonText: '取消', type: 'info' }
    )
    const res = await examApi.startExam(bank.id)
    ElMessage.success('考试开始')
    continueExam(res.exam_id)
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') {
      console.error('开始考试失败:', error)
    }
  }
}

const continueExam = (examId) => {
  router.push({ path: '/exam', query: { exam_id: examId } })
}

const handleLogout = () => {
  userStore.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

onMounted(loadData)
</script>

<style scoped>
.welcome-container {
  min-height: 100vh;
  background: #f5f7fa;
}

.header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px;
}

.brand,
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-text {
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.user-info {
  color: #606266;
}

.main-content {
  max-width: 1100px;
  margin: 0 auto;
  padding: 36px 24px 48px;
}

.welcome-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
}

.welcome-header h1 {
  margin: 0 0 8px;
  color: #303133;
  font-size: 30px;
}

.welcome-header p,
.section-title p {
  margin: 0;
  color: #909399;
}

.summary-card {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 160px;
  padding: 16px 20px;
  color: #67c23a;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.summary-card > .el-icon {
  font-size: 30px;
}

.summary-card strong,
.summary-card span {
  display: block;
}

.summary-card strong {
  font-size: 24px;
}

.summary-card span {
  color: #909399;
  font-size: 13px;
}

.legacy-alert {
  margin-bottom: 24px;
}

.legacy-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.bank-sections {
  min-height: 200px;
}

.bank-section {
  margin-bottom: 30px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
}

.section-title h2 {
  margin: 0 0 3px;
  color: #303133;
  font-size: 21px;
}

.section-title p {
  font-size: 13px;
}

.section-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  font-size: 23px;
}

.section-icon.competition {
  color: #e6a23c;
  background: #fdf6ec;
}

.section-icon.certification {
  color: #409eff;
  background: #ecf5ff;
}

.bank-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.bank-card {
  display: flex;
  flex-direction: column;
  min-height: 188px;
  padding: 22px;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 14px;
  box-shadow: 0 3px 12px rgba(0, 0, 0, 0.04);
}

.bank-card-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.bank-card h3 {
  margin: 0;
  color: #303133;
  font-size: 18px;
  line-height: 1.45;
}

.bank-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin: 18px 0;
  color: #909399;
  font-size: 14px;
}

.bank-meta span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.bank-progress {
  margin: -2px 0 16px;
}

.bank-action {
  width: 100%;
  min-height: 42px;
  margin: auto 0 0;
}

.rules-section {
  padding: 24px;
  background: #fff;
  border-radius: 14px;
}

.rules-section h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 18px;
  color: #303133;
  font-size: 17px;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 20px;
  color: #606266;
  font-size: 14px;
}

.rules-grid span {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.6;
}

.rules-grid .el-icon {
  flex-shrink: 0;
  margin-top: 4px;
  color: #67c23a;
}

@media (max-width: 768px) {
  .header-content {
    padding: 12px 16px;
  }

  .main-content {
    padding: 24px 16px 36px;
  }

  .welcome-header h1 {
    font-size: 25px;
  }

  .bank-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .header-content {
    padding-right: 12px;
    padding-left: 12px;
  }

  .brand-text {
    font-size: 16px;
  }

  .user-info > .el-icon,
  .user-name {
    display: none;
  }

  .main-content {
    padding: 20px 12px 28px;
  }

  .welcome-header {
    align-items: stretch;
    flex-direction: column;
    gap: 14px;
  }

  .welcome-header h1 {
    font-size: 22px;
  }

  .summary-card {
    min-width: 0;
  }

  .legacy-content {
    align-items: flex-start;
    flex-direction: column;
  }

  .legacy-content .el-button {
    width: 100%;
  }

  .bank-card {
    min-height: 176px;
    padding: 18px 16px;
  }

  .bank-card h3 {
    font-size: 16px;
  }

  .rules-section {
    padding: 20px 16px;
  }

  .rules-grid {
    grid-template-columns: 1fr;
  }
}
</style>
