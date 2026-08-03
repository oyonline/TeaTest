import request from './request'

// 认证相关
export const authApi = {
  // 用户登录
  userLogin: (data) => request.post('/auth/user/login', data),

  // 管理员登录
  adminLogin: (data) => request.post('/auth/admin/login', data),

  // 获取当前用户信息
  getCurrentUser: () => request.get('/auth/me')
}

// 考试相关
export const examApi = {
  // 获取可选择的具体题库
  getQuestionBanks: () => request.get('/exam/question-banks'),

  // 获取进行中的考试
  getInProgressExam: (questionBankId) => request.get('/exam/in-progress', {
    params: questionBankId ? { question_bank_id: questionBankId } : {}
  }),

  // 开始考试
  startExam: (questionBankId) => request.post('/exam/start', { question_bank_id: questionBankId }),

  // 获取考试信息
  getExamInfo: (examId) => request.get(`/exam/${examId}/info`),

  // 获取题目列表（带进度）
  getQuestions: (params) => request.get('/exam/questions', { params }),

  // 获取所有题目状态
  getAllQuestionStatus: (examId) => request.get('/exam/all-status', { params: { exam_id: examId } }),

  // 提交答案
  submitAnswer: (examId, data) => request.post(`/exam/${examId}/answer`, data),

  // 获取未答题列表
  getUnansweredQuestions: (examId) => request.get(`/exam/${examId}/unanswered`),

  // 获取考试结果
  getExamResult: (examId) => request.get(`/exam/${examId}/result`),

  // 获取考试统计
  getExamStats: () => request.get('/exam/stats')
}

// 管理相关
export const adminApi = {
  // 获取和新增答题人
  getUsers: (params) => request.get('/admin/users', { params }),
  createUser: (data) => request.post('/admin/users', data),

  // 获取题库统计
  getBankStats: () => request.get('/admin/stats'),

  // 获取、创建和更新具体题库
  getQuestionBanks: () => request.get('/admin/question-banks'),
  createQuestionBank: (data) => request.post('/admin/question-banks', data),
  updateQuestionBank: (id, data) => request.put(`/admin/question-banks/${id}`, data),

  // 获取考试记录
  getExamRecords: (params) => request.get('/admin/records', { params }),

  // 获取题目并批量归类
  getQuestions: (params) => request.get('/admin/questions', { params }),
  reclassifyQuestions: (data) => request.post('/admin/questions/reclassify', data),

  // 导入题库
  importQuestions: (data) => {
    const formData = new FormData()
    formData.append('file', data.file)
    formData.append('mode', data.mode || 'replace')
    formData.append('question_bank_id', data.questionBankId)
    return request.post('/admin/questions/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}
