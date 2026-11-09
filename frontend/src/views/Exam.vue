<template>
  <div class="exam-container">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <div class="brand">
            <el-icon :size="24" color="#409eff"><Collection /></el-icon>
            <span class="brand-text">{{ examInfo.question_bank_name || (isPractice ? '通用题库文字练习' : '茶叶理论题库') }}</span>
          </div>
          <el-divider direction="vertical" />
          <div class="user-name">
            <el-icon><User /></el-icon>
            <span>{{ userStore.userName }}</span>
          </div>
        </div>

        <div class="header-center">
          <div class="timer">
            <el-icon><Timer /></el-icon>
            <span class="timer-label">已用时</span>
            <span class="timer-value">{{ formattedTime }}</span>
          </div>
        </div>

        <div class="header-right">
          <div class="progress-info">
            <span class="progress-text">
              已完成 {{ answeredCount }} / {{ totalQuestions }} 题
            </span>
            <el-progress
              :percentage="progressPercentage"
              :stroke-width="8"
              :show-text="false"
              class="progress-bar"
            />
          </div>
        </div>
      </div>
    </header>

    <!-- 主内容区域 -->
    <div class="main-layout">
      <!-- 左侧进度网格 -->
      <aside class="sidebar">
        <div class="sidebar-header">
          <el-icon><Grid /></el-icon>
          <span>答题进度</span>
        </div>
        <div class="question-grid">
          <div
            v-for="item in questionStatusList"
            :key="item.question_id"
            :class="[
              'grid-item',
              {
                'answered-correct': item.status === 'correct',
                'answered-wrong': item.status === 'wrong',
                'answered-ungraded': item.status === 'ungraded',
                'current-page': isQuestionInCurrentPage(item.position)
              }
            ]"
            @click="jumpToQuestionPage(item.position, item.question_id)"
          >
            {{ item.question_no }}
          </div>
        </div>
        <div class="legend">
          <div class="legend-item">
            <span class="legend-dot answered-correct"></span>
            <span>答对</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot answered-wrong"></span>
            <span>答错</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot answered-ungraded"></span>
            <span>已答（未评分）</span>
          </div>
          <div class="legend-item">
            <span class="legend-dot unanswered"></span>
            <span>未答</span>
          </div>
        </div>
      </aside>

      <!-- 中间答题区域 -->
      <main class="exam-main">
        <div v-if="loading" class="loading-state">
          <el-skeleton :rows="5" animated />
        </div>

        <div v-else class="questions-list">
          <div
            v-for="question in questions"
            :key="question.id"
            :id="`question-${question.id}`"
            class="question-card"
          >
            <div class="question-header">
              <span class="question-no">第 {{ question.question_no }} 题</span>
              <el-tag :type="typeTagType(question.question_type)" size="small" class="type-tag">{{ questionTypeLabel(question.question_type) }}</el-tag>
              <el-tag v-if="question.has_answered && !question.is_graded" type="warning" size="small">已答（未自动评分）</el-tag>
              <el-tag v-else-if="question.has_answered" :type="question.is_correct ? 'success' : 'danger'" size="small">
                {{ question.is_correct ? '正确' : '错误' }}
              </el-tag>
              <el-tag v-else type="info" size="small">未作答</el-tag>
            </div>

            <div class="question-content">
              <p class="question-text">{{ question.question_text }}</p>

              <!-- 客观题选项 -->
              <div
                v-if="isChoiceType(question.question_type)"
                class="options-list"
                :class="{ 'multiple-choice': question.question_type === 'multiple_choice' }"
              >
                <template v-if="question.question_type === 'multiple_choice'">
                  <el-checkbox-group
                    v-model="question.selectedAnswers"
                    :disabled="question.has_answered"
                    @change="handleMultipleChoiceChange(question)"
                  >
                    <div
                      v-for="option in getOptions(question)"
                      :key="option.key"
                      :class="[
                        'option-item',
                        'checkbox-option',
                        {
                          'correct': question.has_answered && question.correct_answer.includes(option.key),
                          'wrong': question.has_answered && question.user_answer.includes(option.key) && !question.correct_answer.includes(option.key)
                        }
                      ]"
                    >
                      <el-checkbox :value="option.key">
                        <span class="option-key checkbox-key">{{ option.key }}</span>
                        <span class="option-text">{{ option.value }}</span>
                      </el-checkbox>

                      <el-icon v-if="question.has_answered && question.correct_answer.includes(option.key)" class="answer-icon correct-icon">
                        <CircleCheck />
                      </el-icon>
                      <el-icon v-if="question.has_answered && question.user_answer.includes(option.key) && !question.correct_answer.includes(option.key)" class="answer-icon wrong-icon">
                        <CircleClose />
                      </el-icon>
                    </div>
                  </el-checkbox-group>

                  <div v-if="!question.has_answered" class="submit-section">
                    <el-button
                      type="primary"
                      :disabled="!question.selectedAnswers || question.selectedAnswers.length === 0"
                      :loading="question.submitting"
                      @click="submitMultipleChoice(question)"
                    >
                      提交本题
                    </el-button>
                    <span v-if="question.selectedAnswers && question.selectedAnswers.length > 0" class="selected-hint">
                      已选：{{ question.selectedAnswers.slice().sort().join('') }}
                    </span>
                  </div>
                </template>

                <template v-else>
                  <div
                    v-for="option in getOptions(question)"
                    :key="option.key"
                    :class="[
                      'option-item',
                      {
                        'selected': question.user_answer === option.key,
                        'correct': question.has_answered && question.correct_answer === option.key,
                        'wrong': question.has_answered && question.user_answer === option.key && !question.is_correct
                      }
                    ]"
                    @click="selectAnswer(question, option.key)"
                  >
                    <span class="option-key">{{ option.key }}</span>
                    <span class="option-text">{{ option.value }}</span>

                    <el-icon v-if="question.has_answered && question.correct_answer === option.key" class="answer-icon correct-icon">
                      <CircleCheck />
                    </el-icon>
                    <el-icon v-if="question.has_answered && question.user_answer === option.key && !question.is_correct" class="answer-icon wrong-icon">
                      <CircleClose />
                    </el-icon>
                  </div>
                </template>
              </div>

              <!-- 填空题/简答题：文字作答，错误或未提交时保留草稿 -->
              <div v-else class="text-answer">
                <el-input
                  v-if="question.question_type === 'fill_blank'"
                  v-model="question.draftAnswer"
                  type="textarea"
                  :rows="2"
                  :disabled="question.has_answered"
                  placeholder="请输入答案（可换行）"
                />
                <el-input
                  v-else
                  v-model="question.draftAnswer"
                  type="textarea"
                  :rows="4"
                  :disabled="question.has_answered"
                  placeholder="请输入你的回答（提交后仅对照参考答案，不自动评分）"
                />
                <div v-if="!question.has_answered" class="text-submit">
                  <span class="text-hint">{{ runeLength(question.draftAnswer) }} / 4000 个字符</span>
                  <el-button type="primary" :loading="question.submitting" @click="submitTextAnswer(question)">提交本题</el-button>
                </div>
              </div>

              <!-- 答案解析 -->
              <div v-if="question.has_answered" class="answer-analysis">
                <el-alert
                  v-if="!question.is_graded"
                  title="已提交，简答题未自动评分"
                  type="info"
                  :closable="false"
                  show-icon
                >
                  <template #default>
                    <p class="pre-wrap">参考答案：{{ question.correct_answer }}</p>
                    <p class="pre-wrap">你的答案：{{ question.user_answer }}</p>
                  </template>
                </el-alert>
                <el-alert
                  v-else
                  :title="question.is_correct ? '回答正确！' : '回答错误'"
                  :type="question.is_correct ? 'success' : 'error'"
                  :closable="false"
                  show-icon
                >
                  <template #default>
                    <p class="pre-wrap">正确答案：{{ question.correct_answer }}</p>
                    <p v-if="!question.is_correct" class="pre-wrap">你的答案：{{ question.user_answer }}</p>
                  </template>
                </el-alert>
              </div>
            </div>
          </div>
          <el-alert
            v-if="examCompleted"
            class="complete-alert"
            title="本场答题已完成"
            type="success"
            :closable="false"
            show-icon
          >
            <div class="complete-alert-content">
              <span>请先核对上方最后一题的参考答案，再进入结果页。</span>
              <el-button type="primary" @click="goToResult">查看结果</el-button>
            </div>
          </el-alert>
        </div>

        <!-- 分页器 -->
        <div class="pagination-section">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="totalQuestions"
            :page-sizes="[10]"
            layout="total, prev, pager, next, jumper"
            @change="handlePageChange"
          />
        </div>
      </main>
    </div>

    <!-- 未答题提示对话框 -->
    <el-dialog
      v-model="unansweredDialogVisible"
      title="提示"
      width="min(400px, calc(100vw - 32px))"
    >
      <p>您还有 {{ unansweredCount }} 道题未作答，请完成所有题目后系统将自动交卷。</p>
      <template #footer>
        <el-button @click="unansweredDialogVisible = false">继续答题</el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Collection, User, Timer, Grid,
  CircleCheck, CircleClose, CircleCheckFilled
} from '@element-plus/icons-vue'
import { examApi } from '../api'
import { useUserStore } from '../stores/user'
import { entryFromRoute, entryHome, isPracticeEntry } from '../utils/entry'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const entry = computed(() => entryFromRoute(route))
const isPractice = computed(() => isPracticeEntry(entry.value))

// 状态
const loading = ref(true)
const questions = ref([])
const totalQuestions = ref(500)
const currentPage = ref(1)
const pageSize = ref(10)
const totalPages = ref(50)

const examId = ref(null)
const examInfo = ref({})
const startTime = ref(null)
const elapsedSeconds = ref(0)
const timerInterval = ref(null)

const questionStatusList = ref([])
const answeredCount = ref(0)

const unansweredDialogVisible = ref(false)
const unansweredCount = ref(0)
const examCompleted = ref(false)

// 计时器
const formattedTime = computed(() => {
  const hours = Math.floor(elapsedSeconds.value / 3600)
  const minutes = Math.floor((elapsedSeconds.value % 3600) / 60)
  const seconds = elapsedSeconds.value % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
})

// 进度百分比
const progressPercentage = computed(() => {
  if (totalQuestions.value === 0) return 0
  return Math.round((answeredCount.value / totalQuestions.value) * 100)
})

const isChoiceType = (type) => ['single_choice', 'multiple_choice', 'true_false'].includes(type)
const isTextType = (type) => ['fill_blank', 'short_answer'].includes(type)
const questionTypeLabel = (type) => ({
  single_choice: '单选题',
  multiple_choice: '多选题',
  true_false: '判断题',
  fill_blank: '填空题',
  short_answer: '简答题'
}[type] || '单选题')
const typeTagType = (type) => ({
  single_choice: 'primary',
  multiple_choice: 'warning',
  true_false: 'success',
  fill_blank: 'info',
  short_answer: 'danger'
}[type] || 'primary')
const runeLength = (value) => Array.from(value || '').length

// 获取选项列表
const getOptions = (question) => {
  const options = []
  if (question.option_a) options.push({ key: 'A', value: question.option_a })
  if (question.option_b) options.push({ key: 'B', value: question.option_b })
  if (question.option_c) options.push({ key: 'C', value: question.option_c })
  if (question.option_d) options.push({ key: 'D', value: question.option_d })
  if (question.option_e) options.push({ key: 'E', value: question.option_e })
  return options
}

// 判断题目是否在当前页
const isQuestionInCurrentPage = (position) => {
  const pageStart = (currentPage.value - 1) * pageSize.value + 1
  const pageEnd = currentPage.value * pageSize.value
  return position >= pageStart && position <= pageEnd
}

// 加载题目
const loadQuestions = async () => {
  loading.value = true
  try {
    const res = await examApi.getQuestions({
      page: currentPage.value,
      page_size: pageSize.value,
      exam_id: examId.value
    })

    // 初始化多选题选中状态和文字题草稿；已提交后由服务端恢复答案。
    questions.value = res.list.map(q => {
      q.submitting = false
      q.correct_answer = q.correct_answer || ''
      if (q.question_type === 'multiple_choice') {
        q.selectedAnswers = q.has_answered && q.user_answer ? q.user_answer.split('') : []
      }
      if (isTextType(q.question_type)) {
        q.draftAnswer = q.user_answer || ''
      }
      return q
    })
    totalQuestions.value = res.total
    totalPages.value = res.total_pages
  } catch (error) {
    console.error('加载题目失败:', error)
  } finally {
    loading.value = false
  }
}

// 加载所有题目状态
const loadQuestionStatus = async () => {
  try {
    const res = await examApi.getAllQuestionStatus(examId.value)
    questionStatusList.value = res

    // 计算已答题数
    answeredCount.value = res.filter(item => item.status !== 'unanswered').length
  } catch (error) {
    console.error('加载题目状态失败:', error)
  }
}

// 提交成功后的统一状态更新；最后一题也先展示参考答案，不立即跳转。
const applySubmitResult = async (question, answer, res) => {
  question.has_answered = true
  question.user_answer = res.user_answer ?? answer
  question.is_correct = res.is_correct === true
  question.is_graded = res.is_graded !== false
  question.is_locked = true
  question.correct_answer = res.correct_answer || ''
  question.draftAnswer = question.user_answer

  await loadQuestionStatus()

  if (res.is_completed) {
    examCompleted.value = true
    stopTimer()
    return
  }

  if (question.is_graded) {
    if (res.is_correct) {
      ElMessage.success('回答正确！')
    } else {
      ElMessage.error('回答错误')
    }
  } else {
    ElMessage.info('已提交，简答题不自动评分')
  }

  const currentQuestionIndex = questions.value.findIndex(q => q.id === question.id)
  // 文字题需要停留当前页对照参考答案，不自动翻页；旧客观题保留原行为。
  if (!isTextType(question.question_type) && currentQuestionIndex === questions.value.length - 1 && currentPage.value < totalPages.value) {
    setTimeout(() => {
      currentPage.value++
      handlePageChange()
    }, 1000)
  }
}

// 选择答案（单选题/判断题）
const selectAnswer = async (question, answer) => {
  if (question.has_answered || question.submitting) {
    ElMessage.warning('本题已锁定，无法修改答案')
    return
  }
  question.submitting = true
  try {
    const res = await examApi.submitAnswer(examId.value, {
      question_id: question.id,
      answer: answer
    })
    await applySubmitResult(question, answer, res)
  } catch (error) {
    console.error('提交答案失败:', error)
  } finally {
    question.submitting = false
  }
}

// 多选题选择变化处理
const handleMultipleChoiceChange = (question) => {
  if (question.selectedAnswers && question.selectedAnswers.length > 0) {
    question.selectedAnswers.sort()
  }
}

// 提交多选题答案
const submitMultipleChoice = async (question) => {
  if (question.has_answered || question.submitting) {
    ElMessage.warning('本题已锁定，无法修改答案')
    return
  }
  if (!question.selectedAnswers || question.selectedAnswers.length === 0) {
    ElMessage.warning('请至少选择一个选项')
    return
  }

  const answer = question.selectedAnswers.slice().sort().join('')
  question.submitting = true
  try {
    const res = await examApi.submitAnswer(examId.value, {
      question_id: question.id,
      answer: answer
    })
    await applySubmitResult(question, answer, res)
  } catch (error) {
    console.error('提交答案失败:', error)
  } finally {
    question.submitting = false
  }
}

// 提交填空题/简答题；失败时保留草稿。
const submitTextAnswer = async (question) => {
  if (question.has_answered || question.submitting) {
    ElMessage.warning('本题已锁定，无法修改答案')
    return
  }
  const answer = question.draftAnswer || ''
  if (!answer.trim()) {
    ElMessage.warning('请输入答案')
    return
  }
  if (runeLength(answer) > 4000) {
    ElMessage.warning('答案不能超过4000个字符')
    return
  }
  question.submitting = true
  try {
    const res = await examApi.submitAnswer(examId.value, {
      question_id: question.id,
      answer: answer
    })
    await applySubmitResult(question, answer, res)
  } catch (error) {
    // 草稿保留在 question.draftAnswer 中，便于修改后重试。
    console.error('提交文字答案失败:', error)
  } finally {
    question.submitting = false
  }
}

const goToResult = () => {
  router.push({ path: `/result/${examId.value}`, query: { entry: entry.value } })
}

// 跳转到题目所在页
const jumpToQuestionPage = (position, questionId) => {
  const targetPage = Math.ceil(position / pageSize.value)
  if (targetPage !== currentPage.value) {
    currentPage.value = targetPage
    handlePageChange()
  }

  // 滚动到题目位置
  setTimeout(() => {
    const element = document.getElementById(`question-${questionId}`)
    if (element) {
      element.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, 100)
}

// 分页切换
const handlePageChange = () => {
  loadQuestions()
}

// 检查未答题
const checkUnanswered = async () => {
  try {
    const res = await examApi.getUnansweredQuestions(examId.value)
    unansweredCount.value = res.unanswered_count

    if (unansweredCount.value > 0) {
      unansweredDialogVisible.value = true
    }
  } catch (error) {
    console.error('检查未答题失败:', error)
  }
}

// 启动计时器
const startTimer = () => {
  if (startTime.value) {
    elapsedSeconds.value = Math.floor((Date.now() - new Date(startTime.value).getTime()) / 1000)
  }

  timerInterval.value = setInterval(() => {
    elapsedSeconds.value++
  }, 1000)
}

// 停止计时器
const stopTimer = () => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value)
    timerInterval.value = null
  }
}

// 页面加载
onMounted(async () => {
  examId.value = route.query.exam_id
  if (!examId.value) {
    ElMessage.error('考试ID无效')
    router.push(entryHome(entry.value))
    return
  }

  // 获取当前考试信息
  try {
    const info = await examApi.getExamInfo(examId.value)
    examInfo.value = info
    const completed = info.status === 'completed'
    if (completed) {
      // 已完成后刷新仍加载锁定题目和参考答案，由用户自行进入结果，不自动跳走。
      examCompleted.value = true
      stopTimer()
    }
    if (info.start_time) {
      startTime.value = info.start_time
    }
    if (!completed && info.start_time) {
      startTimer()
    }
  } catch (error) {
    console.error('获取考试信息失败:', error)
  }

  // 加载题目和状态
  await Promise.all([
    loadQuestions(),
    loadQuestionStatus()
  ])
})

// 页面卸载
onUnmounted(() => {
  stopTimer()
})
</script>

<style scoped>
.exam-container {
  min-height: 100vh;
  background: #f5f7fa;
  display: flex;
  flex-direction: column;
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.brand-text {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #606266;
  font-size: 14px;
}

.header-center {
  display: flex;
  align-items: center;
}

.timer {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f5f7fa;
  padding: 8px 16px;
  border-radius: 20px;
}

.timer-label {
  font-size: 14px;
  color: #909399;
}

.timer-value {
  font-size: 16px;
  font-weight: 600;
  color: #409eff;
  font-family: monospace;
}

.header-right {
  display: flex;
  align-items: center;
}

.progress-info {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.progress-text {
  font-size: 14px;
  color: #606266;
}

.progress-bar {
  width: 160px;
}

/* 主布局 */
.main-layout {
  display: flex;
  flex: 1;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  padding: 24px;
  gap: 24px;
}

/* 侧边栏 */
.sidebar {
  width: 280px;
  flex-shrink: 0;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 20px;
  position: sticky;
  top: 80px;
  height: fit-content;
  max-height: calc(100vh - 104px);
  overflow-y: auto;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.question-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 16px;
}

.grid-item {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  background: #f5f7fa;
  color: #606266;
  border: 2px solid transparent;
}

.grid-item:hover {
  background: #e4e7ed;
}

.grid-item.answered-correct {
  background: #e8f5e9;
  color: #4caf50;
}

.grid-item.answered-wrong {
  background: #ffebee;
  color: #f44336;
}

.grid-item.answered-ungraded {
  background: #fdf6ec;
  color: #e6a23c;
}

.grid-item.current-page {
  border-color: #409eff;
}

.legend {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #e4e7ed;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: #606266;
}

.legend-dot {
  width: 16px;
  height: 16px;
  border-radius: 4px;
}

.legend-dot.answered-correct {
  background: #e8f5e9;
  border: 1px solid #4caf50;
}

.legend-dot.answered-wrong {
  background: #ffebee;
  border: 1px solid #f44336;
}

.legend-dot.answered-ungraded {
  background: #fdf6ec;
  border: 1px solid #e6a23c;
}

.legend-dot.unanswered {
  background: #f5f7fa;
  border: 1px solid #dcdfe6;
}

/* 主内容区 */
.exam-main {
  flex: 1;
  min-width: 0;
}

.loading-state {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
}

.questions-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.question-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  padding: 24px;
}

.question-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.type-tag {
  font-weight: 500;
}

.question-no {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.question-text {
  font-size: 15px;
  line-height: 1.6;
  color: #303133;
  margin-bottom: 20px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.option-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.option-item:hover {
  background: #e4e7ed;
}

.option-item.selected {
  border-color: #409eff;
  background: #e3f2fd;
}

.option-item.correct {
  border-color: #67c23a;
  background: #e8f5e9;
}

.option-item.wrong {
  border-color: #f56c6c;
  background: #ffebee;
}

.option-key {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  border-radius: 50%;
  font-size: 14px;
  font-weight: 600;
  color: #606266;
  flex-shrink: 0;
}

.option-item.correct .option-key {
  background: #67c23a;
  color: #fff;
}

.option-item.wrong .option-key {
  background: #f56c6c;
  color: #fff;
}

.option-text {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  color: #606266;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.answer-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.correct-icon {
  color: #67c23a;
}

.wrong-icon {
  color: #f56c6c;
}

.answer-analysis {
  margin-top: 16px;
}

.text-answer {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.text-submit {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.text-hint {
  color: #909399;
  font-size: 13px;
}

.pre-wrap {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.complete-alert {
  margin-top: 16px;
}

.complete-alert-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

/* 多选题样式 */
.multiple-choice .options-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.multiple-choice :deep(.el-checkbox-group) {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.multiple-choice :deep(.el-checkbox) {
  height: auto;
  margin-right: 0;
}

.multiple-choice :deep(.el-checkbox__label) {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-left: 8px;
}

.checkbox-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.checkbox-option :deep(.el-checkbox__input) {
  margin-top: 0;
}

.checkbox-key {
  margin-right: 4px;
}

.submit-section {
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.selected-hint {
  color: #909399;
  font-size: 14px;
}

/* 分页 */
.pagination-section {
  margin-top: 24px;
  display: flex;
  justify-content: center;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

/* 完成对话框 */
.complete-content {
  text-align: center;
  padding: 24px;
}

.complete-icon {
  margin-bottom: 16px;
}

.complete-content h3 {
  font-size: 20px;
  color: #303133;
  margin-bottom: 8px;
}

.complete-content p {
  color: #909399;
}

@media (max-width: 900px) {
  .header-content {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 10px 16px;
    padding: 10px 16px;
  }

  .header-left {
    min-width: 0;
  }

  .brand {
    min-width: 0;
  }

  .brand-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .header-right {
    grid-column: 1 / -1;
  }

  .progress-info {
    width: 100%;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .progress-bar {
    width: min(240px, 45vw);
  }

  .main-layout {
    flex-direction: column;
    padding: 16px;
    gap: 16px;
  }

  .sidebar {
    position: static;
    width: 100%;
    max-height: none;
    padding: 16px;
    overflow: visible;
  }

  .sidebar-header {
    margin-bottom: 12px;
    padding-bottom: 10px;
  }

  .question-grid {
    grid-template-rows: repeat(2, 40px);
    grid-template-columns: none;
    grid-auto-flow: column;
    grid-auto-columns: 40px;
    gap: 8px;
    padding-bottom: 8px;
    overflow-x: auto;
    overscroll-behavior-x: contain;
    scrollbar-width: thin;
    -webkit-overflow-scrolling: touch;
  }

  .grid-item {
    width: 40px;
    min-height: 40px;
    aspect-ratio: auto;
  }

  .legend {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 8px 16px;
  }
}

@media (max-width: 640px) {
  .header-content {
    padding: 10px 12px;
    gap: 8px 12px;
  }

  .header-left {
    gap: 0;
  }

  .header-left :deep(.el-divider),
  .user-name {
    display: none;
  }

  .brand-text {
    font-size: 14px;
  }

  .timer {
    gap: 6px;
    padding: 7px 10px;
  }

  .timer-label {
    display: none;
  }

  .timer-value {
    font-size: 14px;
  }

  .progress-text {
    font-size: 12px;
    white-space: nowrap;
  }

  .progress-bar {
    width: min(160px, 44vw);
  }

  .main-layout {
    padding: 12px;
    gap: 12px;
  }

  .sidebar {
    padding: 14px 12px;
    border-radius: 10px;
  }

  .question-grid {
    margin-bottom: 10px;
  }

  .question-card {
    padding: 16px;
    border-radius: 10px;
  }

  .question-header {
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 14px;
  }

  .question-no {
    margin-right: auto;
    font-size: 15px;
  }

  .question-text {
    font-size: 15px;
    line-height: 1.7;
  }

  .option-item {
    align-items: flex-start;
    min-height: 52px;
    padding: 12px;
  }

  .answer-icon {
    margin-top: 4px;
  }

  .multiple-choice :deep(.el-checkbox) {
    align-items: flex-start;
    white-space: normal;
  }

  .multiple-choice :deep(.el-checkbox__input) {
    margin-top: 6px;
  }

  .multiple-choice :deep(.el-checkbox__label) {
    align-items: flex-start;
    min-width: 0;
    padding-left: 6px;
    white-space: normal;
  }

  .submit-section,
  .text-submit,
  .complete-alert-content {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
  }

  .submit-section .el-button {
    width: 100%;
    min-height: 44px;
  }

  .pagination-section {
    margin-top: 16px;
    padding: 12px 4px;
    overflow: hidden;
  }

  .pagination-section :deep(.el-pagination) {
    max-width: 100%;
    --el-pagination-button-width: 28px;
    --el-pagination-button-height: 32px;
    --el-pagination-button-disabled-bg-color: transparent;
  }

  .pagination-section :deep(.el-pagination__total),
  .pagination-section :deep(.el-pagination__jump) {
    display: none;
  }

  .loading-state {
    padding: 24px 16px;
  }

  .complete-content {
    padding: 12px 0;
  }
}

@media (max-width: 360px) {
  .brand-text {
    max-width: 150px;
  }

  .progress-info {
    gap: 8px;
  }

  .progress-bar {
    width: 42vw;
  }

  .pagination-section :deep(.el-pager li:not(.is-active):not(.more)) {
    display: none;
  }
}
</style>
