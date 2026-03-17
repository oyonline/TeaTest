<template>
  <div class="admin-container">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <el-icon :size="24" color="#409eff"><Setting /></el-icon>
          <span class="brand-text">管理控制台</span>
        </div>
        <div class="header-right">
          <el-button type="danger" link @click="handleLogout">
            <el-icon><SwitchButton /></el-icon>
            退出登录
          </el-button>
        </div>
      </div>
    </header>

    <!-- 主内容 -->
    <main class="main-content">
      <!-- 统计卡片 -->
      <div class="stats-section">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon blue">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total_questions || 0 }}</div>
            <div class="stat-label">题库总量</div>
          </div>
        </el-card>

        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon green">
            <el-icon><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.completed_exams || 0 }}</div>
            <div class="stat-label">已完成考试</div>
          </div>
        </el-card>

        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon orange">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.in_progress_exams || 0 }}</div>
            <div class="stat-label">进行中考试</div>
          </div>
        </el-card>

        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon purple">
            <el-icon><User /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total_exams || 0 }}</div>
            <div class="stat-label">考试总数</div>
          </div>
        </el-card>
      </div>

      <!-- 题库导入 -->
      <el-card class="import-section" shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <el-icon><Upload /></el-icon>
              <span>题库导入</span>
            </div>
          </div>
        </template>

        <div class="import-content">
          <el-alert
            title="导入说明"
            type="info"
            :closable="false"
            show-icon
            class="import-tips"
          >
            <template #default>
              <ul>
                <li>支持 .xlsx 和 .csv 格式文件</li>
                <li>自动识别 UTF-8 和 GBK 编码</li>
                <li>覆盖模式：清空旧题库后导入新题库</li>
                <li>追加模式：将新题追加到现有题库</li>
              </ul>
            </template>
          </el-alert>

          <div class="import-form">
            <el-form :model="importForm" label-width="100px">
              <el-form-item label="导入模式">
                <el-radio-group v-model="importForm.mode">
                  <el-radio value="replace">覆盖模式</el-radio>
                  <el-radio value="append">追加模式</el-radio>
                </el-radio-group>
              </el-form-item>

              <el-form-item label="选择文件">
                <el-upload
                  ref="uploadRef"
                  action=""
                  :auto-upload="false"
                  :on-change="handleFileChange"
                  :on-remove="handleFileRemove"
                  :limit="1"
                  accept=".csv,.xlsx"
                  class="upload-demo"
                >
                  <el-button type="primary">
                    <el-icon><Upload /></el-icon>
                    选择文件
                  </el-button>
                  <template #tip>
                    <div class="el-upload__tip">
                      支持 .csv 和 .xlsx 格式文件
                    </div>
                  </template>
                </el-upload>
              </el-form-item>

              <el-form-item>
                <el-button
                  type="primary"
                  :loading="importLoading"
                  :disabled="!importForm.file"
                  @click="handleImport"
                >
                  <el-icon><Check /></el-icon>
                  开始导入
                </el-button>
              </el-form-item>
            </el-form>
          </div>

          <!-- 导入结果 -->
          <div v-if="importResult" class="import-result">
            <el-divider />
            <h4>导入结果</h4>
            <el-row :gutter="20">
              <el-col :span="12">
                <el-statistic title="成功数量" :value="importResult.success_count" />
              </el-col>
              <el-col :span="12">
                <el-statistic
                  title="失败数量"
                  :value="importResult.fail_count"
                  value-style="color: #f56c6c"
                />
              </el-col>
            </el-row>

            <div v-if="importResult.fail_reasons && importResult.fail_reasons.length > 0" class="fail-reasons">
              <el-alert
                title="失败原因"
                type="warning"
                :closable="false"
                show-icon
              >
                <ul>
                  <li v-for="(reason, index) in importResult.fail_reasons" :key="index">
                    {{ reason }}
                  </li>
                </ul>
              </el-alert>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 考试记录 -->
      <el-card class="records-section" shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-title">
              <el-icon><List /></el-icon>
              <span>考试记录</span>
            </div>
            <el-input
              v-model="searchKeyword"
              placeholder="搜索答题人姓名"
              clearable
              style="width: 200px"
              @clear="loadRecords"
              @keyup.enter="loadRecords"
            >
              <template #suffix>
                <el-icon @click="loadRecords"><Search /></el-icon>
              </template>
            </el-input>
          </div>
        </template>

        <el-table :data="records" stripe style="width: 100%">
          <el-table-column prop="user_name" label="答题人" width="100" />
          <el-table-column prop="start_time" label="开始时间" width="160">
            <template #default="{ row }">
              {{ formatDate(row.start_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="end_time" label="完成时间" width="160">
            <template #default="{ row }">
              {{ row.end_time ? formatDate(row.end_time) : '-' }}
            </template>
          </el-table-column>
          <el-table-column prop="duration_seconds" label="总耗时" width="100">
            <template #default="{ row }">
              {{ formatDuration(row.duration_seconds) }}
            </template>
          </el-table-column>
          <el-table-column prop="completed_count" label="完成题数" width="90" />
          <el-table-column prop="correct_count" label="正确数" width="80">
            <template #default="{ row }">
              <span style="color: #67c23a">{{ row.correct_count }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="wrong_count" label="错误数" width="80">
            <template #default="{ row }">
              <span style="color: #f56c6c">{{ row.wrong_count }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="total_score" label="总分" width="80">
            <template #default="{ row }">
              <strong>{{ row.total_score }}</strong>
            </template>
          </el-table-column>
          <el-table-column prop="accuracy_rate" label="正确率" width="90">
            <template #default="{ row }">
              <el-tag :type="getAccuracyTagType(row.accuracy_rate)" size="small">
                {{ row.accuracy_rate?.toFixed(2) }}%
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="90">
            <template #default="{ row }">
              <el-tag :type="row.status === 'completed' ? 'success' : 'warning'" size="small">
                {{ row.status === 'completed' ? '已完成' : '进行中' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="handleSizeChange"
            @current-change="handleCurrentChange"
          />
        </div>
      </el-card>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Setting, SwitchButton, Document, CircleCheck, Timer,
  User, Upload, Check, List, Search
} from '@element-plus/icons-vue'
import { adminApi } from '../api'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const adminStore = useAdminStore()

// 统计
const stats = ref({})

// 导入
const uploadRef = ref(null)
const importLoading = ref(false)
const importForm = ref({
  mode: 'replace',
  file: null
})
const importResult = ref(null)

// 考试记录
const records = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchKeyword = ref('')

// 加载统计
const loadStats = async () => {
  try {
    const res = await adminApi.getBankStats()
    stats.value = res
  } catch (error) {
    console.error('加载统计失败:', error)
  }
}

// 加载考试记录
const loadRecords = async () => {
  try {
    const res = await adminApi.getExamRecords({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchKeyword.value
    })
    records.value = res.list
    total.value = res.total
  } catch (error) {
    console.error('加载考试记录失败:', error)
  }
}

// 文件变更
const handleFileChange = (file) => {
  importForm.value.file = file.raw
  importResult.value = null
}

// 文件移除
const handleFileRemove = () => {
  importForm.value.file = null
  importResult.value = null
}

// 导入题库
const handleImport = async () => {
  if (!importForm.value.file) {
    ElMessage.warning('请选择文件')
    return
  }

  // 确认对话框
  const confirmText = importForm.value.mode === 'replace'
    ? '覆盖模式将清空原有题库，确定要继续吗？'
    : '确定要导入新题目吗？'

  try {
    await ElMessageBox.confirm(confirmText, '确认导入', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }

  importLoading.value = true
  try {
    const res = await adminApi.importQuestions({
      file: importForm.value.file,
      mode: importForm.value.mode
    })

    importResult.value = res
    ElMessage.success(`导入成功：${res.success_count} 道题`)

    // 清空文件
    uploadRef.value?.clearFiles()
    importForm.value.file = null

    // 刷新统计
    loadStats()
  } catch (error) {
    console.error('导入失败:', error)
  } finally {
    importLoading.value = false
  }
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  loadRecords()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadRecords()
}

// 格式化日期
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 格式化时长
const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}分${secs}秒`
}

// 正确率标签类型
const getAccuracyTagType = (rate) => {
  if (rate >= 90) return 'success'
  if (rate >= 60) return 'warning'
  return 'danger'
}

// 退出登录
const handleLogout = () => {
  adminStore.logout()
  ElMessage.success('已退出登录')
  router.push('/admin/login')
}

onMounted(() => {
  loadStats()
  loadRecords()
})
</script>

<style scoped>
.admin-container {
  min-height: 100vh;
  background: #f5f7fa;
}

/* 顶部导航 */
.header {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  margin-bottom: 24px;
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

/* 主内容 */
.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px 40px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 统计卡片 */
.stats-section {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
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
  color: #2196f3;
}

.stat-icon.green {
  background: #e8f5e9;
  color: #4caf50;
}

.stat-icon.orange {
  background: #fff3e0;
  color: #ff9800;
}

.stat-icon.purple {
  background: #f3e5f5;
  color: #9c27b0;
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

/* 导入区域 */
.import-section {
  border-radius: 12px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #303133;
}

.import-tips :deep(ul) {
  margin: 8px 0;
  padding-left: 20px;
}

.import-tips :deep(li) {
  margin: 4px 0;
}

.import-form {
  margin-top: 24px;
}

.import-result {
  margin-top: 24px;
}

.import-result h4 {
  margin-bottom: 16px;
  color: #303133;
}

.fail-reasons {
  margin-top: 16px;
}

.fail-reasons :deep(ul) {
  margin: 8px 0;
  padding-left: 20px;
}

.fail-reasons :deep(li) {
  margin: 4px 0;
  color: #e6a23c;
}

/* 考试记录 */
.records-section {
  border-radius: 12px;
}

.pagination-wrapper {
  margin-top: 24px;
  display: flex;
  justify-content: center;
}

@media (max-width: 768px) {
  .stats-section {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
