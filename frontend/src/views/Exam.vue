<template>
  <div class="exam-container">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <div class="brand">
            <el-icon :size="24" color="#409eff"><Collection /></el-icon>
            <span class="brand-text">评茶员初赛理论题库</span>
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
                'current-page': isQuestionInCurrentPage(item.question_no)
              }
            ]"
            @click="jumpToQuestionPage(item.question_no)"
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
            :id="`question-${question.question_no}`"
            class="question-card"
          >
            <div class="question-header">
              <span class="question-no">第 {{ question.question_no }} 题</span>
              <el-tag v-if="question.question_type === 'multiple_choice'" type="warning" size="small" class="type-tag">多选题</el-tag>
              <el-tag v-else-if="question.question_type === 'true_false'" type="success" size="small" class="type-tag">判断题</el-tag>
              <el-tag v-else type="primary" size="small" class="type-tag">单选题</el-tag>
              <el-tag v-if="question.has_answered" :type="question.is_correct ? 'success' : 'danger'" size="small">
                {{ question.is_correct ? '正确' : '错误' }}
              </el-tag>
              <el-tag v-else type="info" size="small">未作答</el-tag>
            </div>

            <div class="question-content">
              <p class="question-text">{{ question.question_text }}</p>

              <!-- 单选题和判断题：点击选择 -->
              <div v-if="question.question_type !== 'multiple_choice'" class="options-list">
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

                  <!-- 答案标记 -->
                  <el-icon v-if="question.has_answered && question.correct_answer === option.key" class="answer-icon correct-icon">
                    <CircleCheck />
                  </el-icon>
                  <el-icon v-if="question.has_answered && question.user_answer === option.key && !question.is_correct" class="answer-icon wrong-icon">
                    <CircleClose />
                  </el-icon>
                </div>
              </div>

              <!-- 多选题：复选框选择 -->
              <div v-else class="options-list multiple-choice">
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

                    <!-- 答案标记 -->
                    <el-icon v-if="question.has_answered && question.correct_answer.includes(option.key)" class="answer-icon correct-icon">
                      <CircleCheck />
                    </el-icon>
                    <el-icon v-if="question.has_answered && question.user_answer.includes(option.key) && !question.correct_answer.includes(option.key)" class="answer-icon wrong-icon">
                      <CircleClose />
                    </el-icon>
                  </div>
                </el-checkbox-group>

                <!-- 提交按钮 -->
                <div v-if="!question.has_answered" class="submit-section">
                  <el-button
                    type="primary"
                    :disabled="!question.selectedAnswers || question.selectedAnswers.length === 0"
                    @click="submitMultipleChoice(question)"
                  >
                    提交本题
                  </el-button>
                  <span v-if="question.selectedAnswers && question.selectedAnswers.length > 0" class="selected-hint">
                    已选：{{ question.selectedAnswers.sort().join('') }}
                  </span>
                </div>
              </div>

              <!-- 答案解析 -->
              <div v-if="question.has_answered" class="answer-analysis">
                <el-alert
                  :title="question.is_correct ? '回答正确！' : '回答错误'"
                  :type="question.is_correct ? 'success' : 'error'"
                  :closable="false"
                  show-icon
                >
                  <template #default>
                    <p>正确答案：{{ question.correct_answer }}</p>
                    <p v-if="question.question_type === 'multiple_choice' && !question.is_correct">
                      你的答案：{{ question.user_answer }}
                    </p>
                  </template>
                </el-alert>
              </div>
            </div>
          </div>
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
      width="400px"
    >
      <p>您还有 {{ unansweredCount }} 道题未作答，请完成所有题目后系统将自动交卷。</p>
      <template #footer>
        <el-button @click="unansweredDialogVisible = false">继续答题</el-button>
      </template>
    </el-dialog>

    <!-- 考试完成对话框 -->
    <el-dialog
      v-model="completeDialogVisible"
      title="考试完成"
      width="400px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="complete-content">
        <el-icon :size="64" color="#67c23a" class="complete-icon"><CircleCheckFilled /></el-icon>
        <h3>恭喜您完成考试！</h3>
        <p>正在为您计算成绩...</p>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Collection, User, Timer, Grid,
  CircleCheck, CircleClose, CircleCheckFilled
} from '@element-plus/icons-vue'
import { examApi } from '../api'
import { useUserStore } from '../stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 状态
const loading = ref(true)
const questions = ref([])
const totalQuestions = ref(500)
const currentPage = ref(1)
const pageSize = ref(10)
const totalPages = ref(50)

const examId = ref(null)
const startTime = ref(null)
const elapsedSeconds = ref(0)
const timerInterval = ref(null)

const questionStatusList = ref([])
const answeredCount = ref(0)

const unansweredDialogVisible = ref(false)
const unansweredCount = ref(0)
const completeDialogVisible = ref(false)

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
const isQuestionInCurrentPage = (questionNo) => {
  const pageStart = (currentPage.value - 1) * pageSize.value + 1
  const pageEnd = currentPage.value * pageSize.value
  return questionNo >= pageStart && questionNo <= pageEnd
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

    // 为多选题初始化选中状态
    questions.value = res.list.map(q => {
      if (q.question_type === 'multiple_choice') {
        // 如果已作答，解析用户答案为数组
        if (q.has_answered && q.user_answer) {
          q.selectedAnswers = q.user_answer.split('')
        } else {
          q.selectedAnswers = []
        }
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

// 选择答案（单选题）
const selectAnswer = async (question, answer) => {
  // 已锁定则不能修改
  if (question.has_answered) {
    ElMessage.warning('本题已锁定，无法修改答案')
    return
  }

  try {
    const res = await examApi.submitAnswer(examId.value, {
      question_id: question.id,
      answer: answer
    })

    // 更新题目状态
    question.has_answered = true
    question.user_answer = answer
    question.is_correct = res.is_correct
    question.is_locked = true

    // 刷新状态列表
    await loadQuestionStatus()

    // 检查是否完成
    if (res.is_completed) {
      completeDialogVisible.value = true
      setTimeout(() => {
        router.push(`/result/${examId.value}`)
      }, 2000)
      return
    }

    // 显示反馈
    if (res.is_correct) {
      ElMessage.success('回答正确！')
    } else {
      ElMessage.error('回答错误')
    }

    // 检查是否是当前页最后一题，是则自动跳转下一页
    const currentPageQuestions = questions.value
    const currentQuestionIndex = currentPageQuestions.findIndex(q => q.id === question.id)
    if (currentQuestionIndex === currentPageQuestions.length - 1 && currentPage.value < totalPages.value) {
      setTimeout(() => {
        currentPage.value++
        handlePageChange()
      }, 1000)
    }
  } catch (error) {
    console.error('提交答案失败:', error)
  }
}

// 多选题选择变化处理
const handleMultipleChoiceChange = (question) => {
  // 将选中的答案排序并转为字符串
  if (question.selectedAnswers && question.selectedAnswers.length > 0) {
    question.selectedAnswers.sort()
  }
}

// 提交多选题答案
const submitMultipleChoice = async (question) => {
  // 已锁定则不能修改
  if (question.has_answered) {
    ElMessage.warning('本题已锁定，无法修改答案')
    return
  }

  // 检查是否选择了答案
  if (!question.selectedAnswers || question.selectedAnswers.length === 0) {
    ElMessage.warning('请至少选择一个选项')
    return
  }

  // 将选中的答案排序并转为字符串（如 ['A', 'C', 'B'] -> 'ABC'）
  const answer = question.selectedAnswers.slice().sort().join('')

  try {
    const res = await examApi.submitAnswer(examId.value, {
      question_id: question.id,
      answer: answer
    })

    // 更新题目状态
    question.has_answered = true
    question.user_answer = answer
    question.is_correct = res.is_correct
    question.is_locked = true

    // 刷新状态列表
    await loadQuestionStatus()

    // 检查是否完成
    if (res.is_completed) {
      completeDialogVisible.value = true
      setTimeout(() => {
        router.push(`/result/${examId.value}`)
      }, 2000)
      return
    }

    // 显示反馈
    if (res.is_correct) {
      ElMessage.success('回答正确！')
    } else {
      ElMessage.error('回答错误')
    }
  } catch (error) {
    console.error('提交答案失败:', error)
  }
}

// 跳转到题目所在页
const jumpToQuestionPage = (questionNo) => {
  const targetPage = Math.ceil(questionNo / pageSize.value)
  if (targetPage !== currentPage.value) {
    currentPage.value = targetPage
    handlePageChange()
  }

  // 滚动到题目位置
  setTimeout(() => {
    const element = document.getElementById(`question-${questionNo}`)
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
    router.push('/welcome')
    return
  }

  // 获取考试信息
  try {
    const examInfo = await examApi.getInProgressExam()
    if (examInfo.has_in_progress) {
      startTime.value = examInfo.start_time
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
  font-size: 14px;
  color: #606266;
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
</style>
