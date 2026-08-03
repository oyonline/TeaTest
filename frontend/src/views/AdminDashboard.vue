<template>
  <div class="admin-container">
    <header class="header">
      <div class="header-content">
        <div class="brand">
          <el-icon :size="24" color="#409eff"><Setting /></el-icon>
          <span class="brand-text">管理控制台</span>
        </div>
        <el-button type="danger" link @click="handleLogout">
          <el-icon><SwitchButton /></el-icon>
          退出登录
        </el-button>
      </div>
    </header>

    <main class="main-content">
      <section class="stats-section">
        <el-card v-for="item in statCards" :key="item.key" class="stat-card" shadow="hover">
          <div class="stat-icon" :class="item.color"><el-icon><component :is="item.icon" /></el-icon></div>
          <div>
            <div class="stat-value">{{ stats[item.key] || 0 }}</div>
            <div class="stat-label">{{ item.label }}</div>
          </div>
        </el-card>
      </section>

      <el-alert
        v-if="stats.unclassified_questions > 0"
        :title="`还有 ${stats.unclassified_questions} 道历史题目未分类`"
        description="未分类题目不会显示在用户端，请在下方“题目归类”中完成分配。"
        type="warning"
        :closable="false"
        show-icon
      />

      <el-card class="section-card" shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-title"><el-icon><Collection /></el-icon><span>题库类型管理</span></div>
            <el-button type="primary" @click="openBankDialog()"><el-icon><Plus /></el-icon>新增具体题库</el-button>
          </div>
        </template>

        <div class="table-wrapper">
          <el-table :data="banks" stripe v-loading="banksLoading">
            <el-table-column prop="category_name" label="一级分类" width="110" />
            <el-table-column prop="name" label="具体题库" min-width="210" />
            <el-table-column prop="question_count" label="题目数" width="90" />
            <el-table-column prop="exam_count" label="考试数" width="90" />
            <el-table-column prop="sort_order" label="排序" width="75" />
            <el-table-column label="状态" width="90">
              <template #default="{ row }">
                <el-switch
                  v-model="row.status"
                  :active-value="1"
                  :inactive-value="0"
                  @change="handleBankStatusChange(row)"
                />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="90" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link @click="openBankDialog(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>

      <el-card class="section-card" shadow="never">
        <template #header>
          <div class="header-title"><el-icon><Upload /></el-icon><span>题库导入</span></div>
        </template>

        <el-alert title="覆盖模式只会清空所选具体题库，不影响其他题库。导入后，进行中的答题会直接使用最新题目。" type="info" :closable="false" show-icon />
        <el-form :model="importForm" label-width="100px" class="form-block">
          <el-form-item label="所属题库" required>
            <el-select v-model="importForm.questionBankId" placeholder="请选择具体题库" class="wide-control">
              <el-option v-for="bank in banks" :key="bank.id" :label="`${bank.category_name} / ${bank.name}`" :value="bank.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="导入模式">
            <el-radio-group v-model="importForm.mode">
              <el-radio value="replace">覆盖所选题库</el-radio>
              <el-radio value="append">追加到所选题库</el-radio>
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
              class="upload-control"
            >
              <el-button type="primary"><el-icon><Upload /></el-icon>选择文件</el-button>
              <template #tip><div class="el-upload__tip">支持现有 CSV / XLSX 题库文件格式</div></template>
            </el-upload>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="importLoading" :disabled="!importForm.file || !importForm.questionBankId" @click="handleImport">
              开始导入
            </el-button>
          </el-form-item>
        </el-form>

        <div v-if="importResult" class="result-box">
          成功 {{ importResult.success_count }} 道，失败 {{ importResult.fail_count }} 道
          <ul v-if="importResult.fail_reasons?.length"><li v-for="reason in importResult.fail_reasons" :key="reason">{{ reason }}</li></ul>
        </div>
      </el-card>

      <el-card class="section-card" shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-title"><el-icon><Sort /></el-icon><span>题目归类</span></div>
            <div class="header-actions">
              <el-select v-model="questionFilter.bank" class="filter-select" @change="resetQuestionPage">
                <el-option label="全部题目" value="" />
                <el-option label="未分类" value="unclassified" />
                <el-option v-for="bank in banks" :key="bank.id" :label="bank.name" :value="String(bank.id)" />
              </el-select>
              <el-input v-model="questionFilter.keyword" placeholder="题号或题干" clearable class="search-input" @keyup.enter="loadQuestions" @clear="loadQuestions" />
              <el-button @click="loadQuestions">查询</el-button>
            </div>
          </div>
        </template>

        <div class="classify-toolbar">
          <span>已选 {{ selectedQuestions.length }} 道</span>
          <el-select v-model="targetBankId" placeholder="选择目标题库" class="target-select">
            <el-option v-for="bank in banks" :key="bank.id" :label="`${bank.category_name} / ${bank.name}`" :value="bank.id" />
          </el-select>
          <el-button type="primary" :disabled="!targetBankId || selectedQuestions.length === 0" :loading="classifyLoading" @click="classifySelected">归类所选题目</el-button>
          <el-button v-if="questionFilter.bank === 'unclassified' && !questionFilter.keyword.trim()" type="warning" :disabled="!targetBankId || questionTotal === 0" :loading="classifyLoading" @click="classifyAllUnclassified">全部未分类归入此题库</el-button>
        </div>

        <div class="table-wrapper">
          <el-table :data="questions" stripe v-loading="questionsLoading" @selection-change="selectedQuestions = $event">
            <el-table-column type="selection" width="48" />
            <el-table-column prop="question_no" label="题号" width="75" />
            <el-table-column label="当前题库" width="180">
              <template #default="{ row }">{{ row.question_bank?.name || '未分类' }}</template>
            </el-table-column>
            <el-table-column prop="question_type_name" label="题型" width="90" />
            <el-table-column prop="question_text" label="题干" min-width="360" show-overflow-tooltip />
          </el-table>
        </div>
        <div class="pagination-wrapper">
          <el-pagination v-model:current-page="questionPage" v-model:page-size="questionPageSize" :total="questionTotal" :page-sizes="[20, 50, 100]" layout="total, sizes, prev, pager, next" @change="loadQuestions" />
        </div>
      </el-card>

      <el-card class="section-card" shadow="never">
        <template #header>
          <div class="card-header">
            <div class="header-title"><el-icon><List /></el-icon><span>考试记录</span></div>
            <div class="header-actions">
              <el-select v-model="recordBankFilter" class="filter-select" @change="resetRecordPage">
                <el-option label="全部题库" value="" />
                <el-option label="历史未分类" value="unclassified" />
                <el-option v-for="bank in banks" :key="bank.id" :label="bank.name" :value="String(bank.id)" />
              </el-select>
              <el-input v-model="searchKeyword" placeholder="答题人姓名" clearable class="search-input" @keyup.enter="loadRecords" @clear="loadRecords" />
              <el-button @click="loadRecords">查询</el-button>
            </div>
          </div>
        </template>

        <div class="table-wrapper records-table">
          <el-table :data="records" stripe v-loading="recordsLoading">
            <el-table-column prop="user_name" label="答题人" width="100" />
            <el-table-column label="题库" min-width="190">
              <template #default="{ row }">
                <div>{{ row.question_bank_name || '历史未分类' }}</div>
                <small v-if="row.question_bank_category" class="muted">{{ categoryName(row.question_bank_category) }}</small>
              </template>
            </el-table-column>
            <el-table-column label="开始时间" width="160"><template #default="{ row }">{{ formatDate(row.start_time) }}</template></el-table-column>
            <el-table-column label="完成时间" width="160"><template #default="{ row }">{{ formatDate(row.end_time) }}</template></el-table-column>
            <el-table-column label="耗时" width="100"><template #default="{ row }">{{ formatDuration(row.duration_seconds) }}</template></el-table-column>
            <el-table-column prop="completed_count" label="完成" width="75" />
            <el-table-column prop="correct_count" label="正确" width="75" />
            <el-table-column prop="wrong_count" label="错误" width="75" />
            <el-table-column label="正确率" width="95"><template #default="{ row }"><el-tag :type="accuracyType(row.accuracy_rate)" size="small">{{ Number(row.accuracy_rate || 0).toFixed(2) }}%</el-tag></template></el-table-column>
            <el-table-column label="状态" width="90"><template #default="{ row }"><el-tag :type="row.status === 'completed' ? 'success' : 'warning'" size="small">{{ row.status === 'completed' ? '已完成' : '进行中' }}</el-tag></template></el-table-column>
          </el-table>
        </div>
        <div class="pagination-wrapper">
          <el-pagination v-model:current-page="recordPage" v-model:page-size="recordPageSize" :total="recordTotal" :page-sizes="[10, 20, 50]" layout="total, sizes, prev, pager, next" @change="loadRecords" />
        </div>
      </el-card>
    </main>

    <el-dialog v-model="bankDialogVisible" :title="editingBankId ? '编辑具体题库' : '新增具体题库'" width="min(480px, calc(100vw - 24px))">
      <el-form :model="bankForm" label-width="90px">
        <el-form-item label="一级分类" required>
          <el-select v-model="bankForm.category_code" class="wide-control">
            <el-option label="竞赛题库" value="competition" />
            <el-option label="考级题库" value="certification" />
          </el-select>
        </el-form-item>
        <el-form-item label="题库名称" required><el-input v-model="bankForm.name" maxlength="100" show-word-limit /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="bankForm.sort_order" :min="0" :max="9999" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="bankForm.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="停用" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bankDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="bankSaving" @click="saveBank">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { computed, markRaw, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  CircleCheck, Collection, Document, List, Plus, Setting,
  Sort, SwitchButton, Timer, Upload, Warning
} from '@element-plus/icons-vue'
import { adminApi } from '../api'
import { useAdminStore } from '../stores/admin'

const router = useRouter()
const adminStore = useAdminStore()
const stats = ref({})
const banks = ref([])
const banksLoading = ref(false)

const statCards = computed(() => [
  { key: 'total_questions', label: '题目总量', icon: markRaw(Document), color: 'blue' },
  { key: 'unclassified_questions', label: '未分类题目', icon: markRaw(Warning), color: 'red' },
  { key: 'active_banks', label: '启用题库', icon: markRaw(Collection), color: 'purple' },
  { key: 'completed_exams', label: '已完成考试', icon: markRaw(CircleCheck), color: 'green' },
  { key: 'in_progress_exams', label: '进行中考试', icon: markRaw(Timer), color: 'orange' }
])

const bankDialogVisible = ref(false)
const editingBankId = ref(null)
const bankSaving = ref(false)
const bankForm = ref({ category_code: 'competition', name: '', sort_order: 0, status: 1 })

const uploadRef = ref(null)
const importLoading = ref(false)
const importResult = ref(null)
const importForm = ref({ questionBankId: null, mode: 'replace', file: null })

const questions = ref([])
const questionsLoading = ref(false)
const classifyLoading = ref(false)
const selectedQuestions = ref([])
const targetBankId = ref(null)
const questionFilter = ref({ bank: 'unclassified', keyword: '' })
const questionPage = ref(1)
const questionPageSize = ref(20)
const questionTotal = ref(0)

const records = ref([])
const recordsLoading = ref(false)
const recordPage = ref(1)
const recordPageSize = ref(10)
const recordTotal = ref(0)
const recordBankFilter = ref('')
const searchKeyword = ref('')

const categoryName = code => code === 'competition' ? '竞赛题库' : code === 'certification' ? '考级题库' : '历史未分类'

const loadStats = async () => {
  try { stats.value = await adminApi.getBankStats() } catch (error) { console.error('加载统计失败:', error) }
}

const loadBanks = async () => {
  banksLoading.value = true
  try { banks.value = await adminApi.getQuestionBanks() } catch (error) { console.error('加载题库失败:', error) } finally { banksLoading.value = false }
}

const openBankDialog = (bank = null) => {
  editingBankId.value = bank?.id || null
  bankForm.value = bank ? {
    category_code: bank.category_code,
    name: bank.name,
    sort_order: bank.sort_order,
    status: bank.status
  } : { category_code: 'competition', name: '', sort_order: 0, status: 1 }
  bankDialogVisible.value = true
}

const saveBank = async () => {
  if (!bankForm.value.name.trim()) {
    ElMessage.warning('请输入题库名称')
    return
  }
  bankSaving.value = true
  try {
    if (editingBankId.value) await adminApi.updateQuestionBank(editingBankId.value, bankForm.value)
    else await adminApi.createQuestionBank(bankForm.value)
    ElMessage.success(editingBankId.value ? '题库已更新' : '题库已创建')
    bankDialogVisible.value = false
    await Promise.all([loadBanks(), loadStats()])
  } catch (error) {
    console.error('保存题库失败:', error)
  } finally {
    bankSaving.value = false
  }
}

const handleBankStatusChange = async (bank) => {
  try {
    await adminApi.updateQuestionBank(bank.id, {
      category_code: bank.category_code,
      name: bank.name,
      sort_order: bank.sort_order,
      status: bank.status
    })
    ElMessage.success(bank.status ? '题库已启用' : '题库已停用')
    loadStats()
  } catch (error) {
    bank.status = bank.status ? 0 : 1
    console.error('更新状态失败:', error)
  }
}

const handleFileChange = file => { importForm.value.file = file.raw; importResult.value = null }
const handleFileRemove = () => { importForm.value.file = null; importResult.value = null }

const handleImport = async () => {
  if (!importForm.value.questionBankId || !importForm.value.file) {
    ElMessage.warning('请选择所属题库和文件')
    return
  }
  const bank = banks.value.find(item => item.id === importForm.value.questionBankId)
  const message = importForm.value.mode === 'replace'
    ? `将清空“${bank?.name}”现有题目后导入，其他题库不受影响。确定继续吗？`
    : `确定将题目追加到“${bank?.name}”吗？`
  try {
    await ElMessageBox.confirm(message, '确认导入', { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' })
  } catch { return }
  importLoading.value = true
  try {
    importResult.value = await adminApi.importQuestions({
      file: importForm.value.file,
      mode: importForm.value.mode,
      questionBankId: importForm.value.questionBankId
    })
    ElMessage.success(`成功导入 ${importResult.value.success_count} 道题`)
    uploadRef.value?.clearFiles()
    importForm.value.file = null
    await Promise.all([loadStats(), loadBanks(), loadQuestions()])
  } catch (error) {
    console.error('导入失败:', error)
  } finally { importLoading.value = false }
}

const loadQuestions = async () => {
  questionsLoading.value = true
  try {
    const res = await adminApi.getQuestions({
      page: questionPage.value,
      page_size: questionPageSize.value,
      keyword: questionFilter.value.keyword,
      question_bank_id: questionFilter.value.bank
    })
    questions.value = res.list || []
    questionTotal.value = res.total || 0
    selectedQuestions.value = []
  } catch (error) { console.error('加载题目失败:', error) } finally { questionsLoading.value = false }
}

const resetQuestionPage = () => { questionPage.value = 1; loadQuestions() }

const classifySelected = async () => {
  await classifyQuestions({
    question_ids: selectedQuestions.value.map(item => item.id),
    question_bank_id: targetBankId.value,
    all_unclassified: false
  })
}

const classifyAllUnclassified = async () => {
  const bank = banks.value.find(item => item.id === targetBankId.value)
  try {
    await ElMessageBox.confirm(`确定将全部 ${questionTotal.value} 道未分类题目归入“${bank?.name}”吗？`, '批量归类', { type: 'warning' })
  } catch { return }
  await classifyQuestions({ question_ids: [], question_bank_id: targetBankId.value, all_unclassified: true })
}

const classifyQuestions = async data => {
  classifyLoading.value = true
  try {
    const res = await adminApi.reclassifyQuestions(data)
    ElMessage.success(`已归类 ${res.updated_count} 道题`)
    await Promise.all([loadQuestions(), loadStats(), loadBanks()])
  } catch (error) { console.error('题目归类失败:', error) } finally { classifyLoading.value = false }
}

const loadRecords = async () => {
  recordsLoading.value = true
  try {
    const res = await adminApi.getExamRecords({
      page: recordPage.value,
      page_size: recordPageSize.value,
      keyword: searchKeyword.value,
      question_bank_id: recordBankFilter.value
    })
    records.value = res.list || []
    recordTotal.value = res.total || 0
  } catch (error) { console.error('加载记录失败:', error) } finally { recordsLoading.value = false }
}

const resetRecordPage = () => { recordPage.value = 1; loadRecords() }

const formatDate = value => value ? new Date(value).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : '-'
const formatDuration = seconds => seconds ? `${Math.floor(seconds / 60)}分${seconds % 60}秒` : '-'
const accuracyType = rate => rate >= 90 ? 'success' : rate >= 60 ? 'warning' : 'danger'

const handleLogout = () => {
  adminStore.logout()
  ElMessage.success('已退出登录')
  router.push('/admin/login')
}

onMounted(async () => {
  await Promise.all([loadStats(), loadBanks()])
  await Promise.all([loadQuestions(), loadRecords()])
})
</script>

<style scoped>
.admin-container { min-height: 100vh; background: #f5f7fa; }
.header { margin-bottom: 24px; background: #fff; box-shadow: 0 2px 8px rgba(0, 0, 0, .06); }
.header-content { display: flex; align-items: center; justify-content: space-between; max-width: 1280px; margin: 0 auto; padding: 16px 24px; }
.brand, .header-title { display: flex; align-items: center; gap: 9px; }
.brand-text, .header-title { color: #303133; font-weight: 600; }
.brand-text { font-size: 18px; }
.main-content { display: flex; flex-direction: column; gap: 24px; max-width: 1280px; margin: 0 auto; padding: 0 24px 40px; }
.stats-section { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 14px; }
.stat-card :deep(.el-card__body) { display: flex; align-items: center; gap: 13px; padding: 18px; }
.stat-icon { display: flex; flex-shrink: 0; align-items: center; justify-content: center; width: 44px; height: 44px; border-radius: 11px; font-size: 22px; }
.stat-icon.blue { color: #2196f3; background: #e3f2fd; }
.stat-icon.red { color: #f56c6c; background: #fef0f0; }
.stat-icon.green { color: #4caf50; background: #e8f5e9; }
.stat-icon.orange { color: #ff9800; background: #fff3e0; }
.stat-icon.purple { color: #9c27b0; background: #f3e5f5; }
.stat-value { color: #303133; font-size: 23px; font-weight: 600; }
.stat-label { margin-top: 3px; color: #909399; font-size: 13px; white-space: nowrap; }
.section-card { border-radius: 12px; }
.card-header { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.header-actions, .classify-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 10px; }
.form-block { max-width: 720px; margin-top: 24px; }
.wide-control { width: 100%; }
.upload-control { width: 100%; }
.result-box { padding: 14px 18px; color: #67c23a; background: #f0f9eb; border-radius: 8px; }
.result-box ul { margin: 8px 0 0; color: #e6a23c; }
.classify-toolbar { margin-bottom: 16px; padding: 14px; background: #f5f7fa; border-radius: 8px; }
.target-select { width: 260px; }
.filter-select { width: 180px; }
.search-input { width: 190px; }
.table-wrapper { width: 100%; overflow-x: auto; overscroll-behavior-x: contain; -webkit-overflow-scrolling: touch; }
.table-wrapper :deep(.el-table) { min-width: 820px; }
.records-table :deep(.el-table) { min-width: 1190px; }
.pagination-wrapper { display: flex; justify-content: center; margin-top: 22px; overflow-x: auto; }
.muted { color: #909399; }

@media (max-width: 1000px) {
  .stats-section { grid-template-columns: repeat(3, minmax(0, 1fr)); }
}

@media (max-width: 768px) {
  .header { margin-bottom: 16px; }
  .header-content { padding: 12px 16px; }
  .main-content { gap: 16px; padding: 0 16px 32px; }
  .stats-section { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .card-header { align-items: stretch; flex-direction: column; }
  .header-actions { align-items: stretch; flex-direction: column; }
  .filter-select, .search-input { width: 100%; }
  .section-card :deep(.el-card__body) { padding: 16px; }
  .classify-toolbar { align-items: stretch; flex-direction: column; }
  .target-select { width: 100%; }
  .pagination-wrapper { justify-content: flex-start; }
}

@media (max-width: 480px) {
  .header-content { padding-right: 12px; padding-left: 12px; }
  .brand-text { font-size: 16px; }
  .main-content { gap: 12px; padding: 0 12px 24px; }
  .stats-section { grid-template-columns: 1fr; }
  .stat-card :deep(.el-card__body) { min-height: 72px; }
  .section-card :deep(.el-card__header) { padding: 14px 16px; }
  .form-block :deep(.el-form-item) { display: block; margin-bottom: 20px; }
  .form-block :deep(.el-form-item__label) { width: auto !important; height: auto; margin-bottom: 8px; line-height: 1.4; }
  .form-block :deep(.el-form-item__content) { margin-left: 0 !important; }
  .upload-control, .upload-control :deep(.el-upload), .upload-control :deep(.el-button) { width: 100%; }
  .pagination-wrapper :deep(.el-pagination__total), .pagination-wrapper :deep(.el-pagination__sizes) { display: none; }
  .pagination-wrapper :deep(.el-pagination) { --el-pagination-button-width: 28px; --el-pagination-button-height: 32px; }
}
</style>
