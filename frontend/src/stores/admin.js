import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAdminStore = defineStore('admin', () => {
  // State
  const token = ref(localStorage.getItem('adminToken') || '')
  const adminInfo = ref(JSON.parse(localStorage.getItem('adminInfo') || '{}'))

  // Getters
  const isLoggedIn = computed(() => !!token.value)

  // Actions
  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('adminToken', newToken)
  }

  const setAdminInfo = (info) => {
    adminInfo.value = info
    localStorage.setItem('adminInfo', JSON.stringify(info))
  }

  const clearAdmin = () => {
    token.value = ''
    adminInfo.value = {}
    localStorage.removeItem('adminToken')
    localStorage.removeItem('adminInfo')
  }

  const logout = () => {
    clearAdmin()
  }

  return {
    token,
    adminInfo,
    isLoggedIn,
    setToken,
    setAdminInfo,
    clearAdmin,
    logout
  }
})
