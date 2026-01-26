import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token'))
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const setAuth = (newToken, userData) => {
    token.value = newToken
    user.value = userData
    localStorage.setItem('token', newToken)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  const logout = () => {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const login = async (phone, password) => {
    try {
      const response = await api.post('/login', { phone, password })
      setAuth(response.data.token, { 
        id: response.data.user_id, 
        role: response.data.role || 'user' 
      })
      return { success: true }
    } catch (error) {
      return { success: false, error: error.response?.data?.error || 'Ошибка входа' }
    }
  }

  const register = async (phone, password, code) => {
    try {
      const response = await api.post('/register', { phone, password, code })
      setAuth(response.data.token, { 
        id: response.data.user_id, 
        role: response.data.role || 'user' 
      })
      return { success: true }
    } catch (error) {
      return { success: false, error: error.response?.data?.error || 'Ошибка регистрации' }
    }
  }

  return {
    token,
    user,
    setAuth,
    logout,
    login,
    register
  }
})