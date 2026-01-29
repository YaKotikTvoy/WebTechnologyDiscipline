import { defineStore } from 'pinia'
import { api } from '@/services/api'

export const useContactsStore = defineStore('contacts', {
  state: () => ({
    searchResults: []
  }),

  actions: {
    async searchUser(phone) {
      try {
        const response = await api.get(`/users/search?phone=${phone}`)
        this.searchResults = [response.data]
        return { success: true, user: response.data }
      } catch (error) {
        return { success: false, error: 'Пользователь не найден' }
      }
    }
  }
})