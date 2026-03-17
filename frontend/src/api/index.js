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
  // 获取进行中的考试
  getInProgressExam: () => request.get('/exam/in-progress'),

  // 开始考试
  startExam: () => request.post('/exam/start'),

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
  // 获取题库统计
  getBankStats: () => request.get('/admin/stats'),

  // 获取考试记录
  getExamRecords: (params) => request.get('/admin/records', { params }),

  // 导入题库
  importQuestions: (data) => {
    const formData = new FormData()
    formData.append('file', data.file)
    formData.append('mode', data.mode || 'replace')
    return request.post('/admin/questions/import', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}
