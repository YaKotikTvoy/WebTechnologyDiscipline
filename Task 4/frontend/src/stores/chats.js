import { defineStore } from 'pinia'
import { api } from '@/services/api'

export const useChatsStore = defineStore('chats', {
  state: () => ({
    chats: [],
    currentChat: null,
    messagesCache: new Map(),
    activeChatId: null,
    isLoadingMessages: false
  }),

  getters: {
    messages: (state) => {
      if (!state.activeChatId) {
        console.log('Нет активного чата ID')
        return []
      }
      const cached = state.messagesCache.get(state.activeChatId)
      console.log('Получаем сообщения для чата', state.activeChatId, 
                  'в кеше:', cached ? cached.length + ' сообщений' : 'нет данных')
      return cached || []
    }
  },

  actions: {
    async fetchChats() {
      try {
        const response = await api.get('/chats')
        this.chats = response.data.map(chat => ({
          ...chat,
          unreadCount: chat.unreadCount || 0
        }))
      } catch {
        console.error('Ошибка загрузки чатов')
      }
    },

    async fetchChat(chatId) {
      try {
        console.log('Загружаем информацию о чате', chatId)
        const response = await api.get(`/chats/${chatId}`)
        this.currentChat = response.data
        this.activeChatId = chatId
        
        console.log('Чат загружен, проверяем кеш сообщений')
        if (!this.messagesCache.has(chatId)) {
          console.log('Сообщений нет в кеше, загружаем...')
          await this.fetchMessages(chatId)
        } else {
          console.log('Сообщения уже в кеше:', this.messagesCache.get(chatId).length)
        }
        
        await this.markSpecificChatAsRead(chatId)
        
        return { success: true }
      } catch (error) {
        console.error('Ошибка загрузки чата:', error)
        return { success: false, error: 'Чат не найден' }
      }
    },

    async fetchMessages(chatId) {
      this.isLoadingMessages = true
      try {
        const response = await api.get(`/chats/${chatId}/messages`)
        this.messagesCache.set(chatId, [...response.data])
        return { success: true }
      } catch {
        return { success: false, error: 'Ошибка загрузки сообщений' }
      } finally {
        this.isLoadingMessages = false
      }
    },

    async markSpecificChatAsRead(chatId) {
      try {
        await api.post(`/chats/${chatId}/read`)
        
        const chatIndex = this.chats.findIndex(c => c.id === chatId)
        if (chatIndex !== -1) {
          this.chats[chatIndex].unreadCount = 0
        }
      } catch {
        console.error('Ошибка пометки чата как прочитанного')
      }
    },

    async markChatAsRead(chatId) {
      try {
        await api.post(`/chats/${chatId}/read`)
      } catch {
        console.error('Ошибка пометки чата как прочитанного')
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
      } catch {
        return { success: false, error: 'Ошибка создания чата' }
      }
    },

    async createGroupChat(name, memberPhones, isSearchable = true) {
      try {
        const response = await api.post('/chats', {
          type: 'group',
          name,
          member_phones: memberPhones,
          is_searchable: isSearchable
        })
        await this.fetchChats()
        return { success: true, chat: response.data }
      } catch {
        return { success: false, error: 'Ошибка создания группы' }
      }
    },

    async sendMessageWithFiles(chatId, content, files) {
      try {
        const formData = new FormData()
        formData.append('content', content)
        
        files.forEach((file, index) => {
          formData.append(`file_${index}`, file)
        })

        const response = await api.post(`/chats/${chatId}/messages`, formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        
        const messages = this.messagesCache.get(chatId) || []
        messages.push(response.data)
        this.messagesCache.set(chatId, messages)
        
        return { success: true, message: response.data }
      } catch {
        return { success: false, error: 'Ошибка отправки сообщения' }
      }
    },

    async deleteMessage(messageId) {
      try {
        await api.delete(`/messages/${messageId}`)
        const messages = this.messagesCache.get(this.activeChatId) || []
        const index = messages.findIndex(m => m.id === messageId)
        if (index !== -1) {
          messages[index].is_deleted = true
          this.messagesCache.set(this.activeChatId, messages)
        }
        return { success: true }
      } catch {
        return { success: false, error: 'Ошибка удаления сообщения' }
      }
    },

    addMessage(message) {
      if (message.chat_id === this.activeChatId) {
        const messages = this.messagesCache.get(this.activeChatId) || []
        messages.push(message)
        this.messagesCache.set(this.activeChatId, messages)
      }
    },

    updateMessageReadOptimistically(messageId, readerId) {
      const messages = this.messagesCache.get(this.activeChatId) || []
      const messageIndex = messages.findIndex(m => m.id === messageId)
      if (messageIndex !== -1) {
        if (!messages[messageIndex].readers) {
          messages[messageIndex].readers = []
        }
        
        const readerExists = messages[messageIndex].readers.some(
          r => r.id === readerId
        )
        
        if (!readerExists) {
          messages[messageIndex].readers.push({
            id: readerId,
            read_at: new Date().toISOString()
          })
        }
        
        this.messagesCache.set(this.activeChatId, messages)
      }
    },

    updateChatUnreadCount(chatId, count) {
      const chatIndex = this.chats.findIndex(c => c.id === chatId)
      if (chatIndex !== -1) {
        this.chats[chatIndex].unreadCount = count
      }
    },

    setActiveChat(chatId) {
      this.activeChatId = chatId
      if (chatId) {
        this.currentChat = this.chats.find(c => c.id === chatId) || null
      } else {
        this.currentChat = null
      }
    },

    async refreshUnreadCounts() {
      for (const chat of this.chats) {
        try {
          const response = await api.get(`/chats/${chat.id}/unread`)
          chat.unreadCount = response.data.count || 0
        } catch {
          chat.unreadCount = 0
        }
      }
    }
  }
})