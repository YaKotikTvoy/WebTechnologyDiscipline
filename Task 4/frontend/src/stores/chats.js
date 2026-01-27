import { defineStore } from 'pinia'
import { api } from '@/services/api'

export const useChatsStore = defineStore('chats', {
  state: () => ({
    chats: [],
    currentChat: null,
    messages: []
  }),

  actions: {
    async fetchChats() {
      try {
        const response = await api.get('/chats')
        this.chats = response.data
      } catch (error) {
        console.error('Failed to fetch chats:', error)
      }
    },

    async fetchChat(chatId) {
      try {
        const response = await api.get(`/chats/${chatId}`)
        this.currentChat = response.data
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Chat not found' }
      }
    },

    async fetchMessages(chatId) {
      try {
        const response = await api.get(`/chats/${chatId}/messages`)
        this.messages = response.data.reverse()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to fetch messages' }
      }
    },

    async createPrivateChat(phone) {
      try {
        const response = await api.post('/chats', {
          type: 'private',
          member_phones: [phone]
        })
        await this.fetchChats()
        return { success: true, chat: response.data }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to create chat' }
      }
    },

    async createGroupChat(name, memberPhones) {
      try {
        const response = await api.post('/chats', {
          type: 'group',
          name,
          member_phones: memberPhones
        })
        await this.fetchChats()
        return { success: true, chat: response.data }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to create chat' }
      }
    },

    async sendMessage(chatId, content, file) {
      try {
        const formData = new FormData()
        formData.append('content', content)
        if (file) {
          formData.append('file', file)
        }

        const response = await api.post(`/chats/${chatId}/messages`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        
        this.messages.push(response.data)
        return { success: true, message: response.data }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to send message' }
      }
    },

    async deleteMessage(messageId) {
      try {
        await api.delete(`/messages/${messageId}`)
        const index = this.messages.findIndex(m => m.id === messageId)
        if (index !== -1) {
          this.messages[index].is_deleted = true
        }
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to delete message' }
      }
    },

    addMessage(message) {
      if (message.chat_id === this.currentChat?.id) {
        this.messages.push(message)
      }
    },

    updateChats() {
      this.fetchChats()
    }
  }
})