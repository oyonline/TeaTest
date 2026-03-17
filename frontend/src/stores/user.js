import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref(localStorage.getItem('userToken') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))

  // Getters
  const isLoggedIn = computed(() => !!token.value)
  const userName = computed(() => userInfo.value?.name || '')
  const userId = computed(() => userInfo.value?.id || 0)

  // Actions
  const setToken = (newToken) => {
    token.value = newToken
    localStorage.setItem('userToken', newToken)
  }

  const setUserInfo = (info) => {
    userInfo.value = info
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  const clearUser = () => {
    token.value = ''
    userInfo.value = {}
    localStorage.removeItem('userToken')
    localStorage.removeItem('userInfo')
  }

  const logout = () => {
    clearUser()
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    userName,
    userId,
    setToken,
    setUserInfo,
    clearUser,
    logout
  }
})
