import { defineStore } from 'pinia'
import { api } from '@/services/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    token: localStorage.getItem('token') || null,
    isAuthenticated: !!localStorage.getItem('token')
  }),

  actions: {
    async login(phone, password) {
      try {
        const response = await api.post('/auth/login', { phone, password })
        this.token = response.data.token
        this.isAuthenticated = true
        localStorage.setItem('token', this.token)
        
        await this.fetchUser()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Login failed' }
      }
    },

    async register(phone, password) {
      try {
        const response = await api.post('/auth/register', { phone, password })
        this.token = response.data.token
        this.isAuthenticated = true
        localStorage.setItem('token', this.token)
        
        await this.fetchUser()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Registration failed' }
      }
    },

    async fetchUser() {
      try {
        const response = await api.get('/auth/me')
        this.user = response.data
      } catch (error) {
        this.logout()
      }
    },

    async updateProfile(username) {
      try {
        await api.put('/auth/profile', { username })
        await this.fetchUser()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to update profile' }
      }
    },

    async logout() {
      try {
        await api.post('/auth/logout')
      } catch (error) {
      } finally {
        this.clearAuth()
      }
    },

    async logoutAll() {
      try {
        await api.post('/auth/logout-all')
      } catch (error) {
      } finally {
        this.clearAuth()
      }
    },

    clearAuth() {
      this.user = null
      this.token = null
      this.isAuthenticated = false
      localStorage.removeItem('token')
    }
  }
})