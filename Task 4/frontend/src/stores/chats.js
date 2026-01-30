import { defineStore } from 'pinia'
import { api } from '@/services/api'
import { useAuthStore } from './auth'

export const useChatsStore = defineStore('chats', {
  state: () => ({
    chats: [],
    currentChat: null,
    messagesCache: new Map(),
    messagesCacheTime: new Map(),
    activeChatId: null,
    isLoadingMessages: false,
    scrollPositions: new Map()
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
    },
    
    shouldRefreshMessages: (state) => (chatId) => {
      const lastLoadTime = state.messagesCacheTime.get(chatId)
      if (!lastLoadTime) return true
      const fiveMinutesAgo = Date.now() - 5 * 60 * 1000
      return lastLoadTime < fiveMinutesAgo
    },
    
    user: () => {
      const authStore = useAuthStore()
      return authStore.user
    }
  },

  actions: {
    async fetchChats() {
      try {
        const response = await api.get('/chats')
        console.log('Ответ от API /chats:', response.data)
        
        this.chats = response.data.map(chat => ({
          ...chat,
          unreadCount: chat.unreadCount || 0,
          lastMessage: chat.last_message || chat.lastMessage || null
        }))
        
        console.log('Чаты сохранены в хранилище:', this.chats.length)
      } catch (error) {
        console.error('Ошибка загрузки чатов:', error)
      }
    },
    
    setActiveChat(chatId) {
      if (this.activeChatId && this.activeChatId !== chatId) {
        this.saveCurrentScrollPosition()
      }
      
      this.activeChatId = chatId
      if (chatId) {
        this.currentChat = this.chats.find(c => c.id === chatId) || null
      } else {
        this.currentChat = null
      }
    },
    
    async fetchChat(chatId) {
      try {
        const response = await api.get(`/chats/${chatId}`)
        this.currentChat = response.data
        this.activeChatId = chatId
        
        const chatIndex = this.chats.findIndex(c => c.id === chatId)
        if (chatIndex !== -1) {
          this.chats[chatIndex] = { ...this.chats[chatIndex], ...response.data }
        }
        
        return { success: true }
      } catch (error) {
        return { success: false, error: 'Чат не найден' }
      }
    },

    async fetchMessages(chatId, forceRefresh = false) {
      if (!forceRefresh && !this.shouldRefreshMessages(chatId)) {
        return { success: true, cached: true }
      }
      
      this.isLoadingMessages = true
      try {
        const response = await api.get(`/chats/${chatId}/messages`)
        this.messagesCache.set(chatId, [...response.data])
        this.messagesCacheTime.set(chatId, Date.now())
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

    async deleteMessage(chatId, messageId) {
      try {
        const response = await api.delete(`/chats/${chatId}/messages/${messageId}`)
        
        const messages = this.messagesCache.get(chatId) || []
        const messageIndex = messages.findIndex(m => m.id === messageId)
        
        if (messageIndex !== -1) {
          messages[messageIndex] = {
            ...messages[messageIndex],
            is_deleted: true,
            content: '[Сообщение удалено]'
          }
          this.messagesCache.set(chatId, [...messages])
        }
        
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data || 'Ошибка удаления сообщения' 
        }
      }
    },

    async editMessage(chatId, messageId, content) {
      try {
        const response = await api.put(`/chats/${chatId}/messages/${messageId}`, { content })
        
        const messages = this.messagesCache.get(chatId) || []
        const messageIndex = messages.findIndex(m => m.id === messageId)
        
        if (messageIndex !== -1) {
          messages[messageIndex] = {
            ...messages[messageIndex],
            content: content,
            is_edited: true,
            updated_at: new Date().toISOString()
          }
          this.messagesCache.set(chatId, [...messages])
        }
        
        return { success: true, message: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data || 'Ошибка редактирования сообщения' 
        }
      }
    },

    clearChatMessages(chatId) {
      this.messagesCache.delete(chatId)
    },

    addMessage(message) {
      if (message.chat_id === this.activeChatId) {
        const messages = this.messagesCache.get(this.activeChatId) || []
        const existingIndex = messages.findIndex(m => m.id === message.id)
        
        if (existingIndex === -1) {
          messages.push(message)
          this.messagesCache.set(this.activeChatId, messages)
        } else {
          messages[existingIndex] = message
          this.messagesCache.set(this.activeChatId, messages)
        }
      }
    },

    updateMessageReadOptimistically(messageId, readerId) {
      const authStore = useAuthStore()
      const messages = this.messagesCache.get(this.activeChatId) || []
      const messageIndex = messages.findIndex(m => m.id === messageId)
      
      if (messageIndex !== -1) {
        const message = messages[messageIndex]
        
        if (!message.readers) {
          message.readers = []
        }
        
        const readerExists = message.readers.some(r => r.id === readerId)
        
        if (!readerExists) {
          message.readers.push({
            id: readerId,
            read_at: new Date().toISOString()
          })
          
          messages[messageIndex] = { ...message }
          this.messagesCache.set(this.activeChatId, messages)
        }
      }
    },

    updateChatUnreadCount(chatId, count) {
      const chatIndex = this.chats.findIndex(c => c.id === chatId)
      if (chatIndex !== -1) {
        this.chats[chatIndex].unreadCount = count
      } else {
        this.fetchChats()
      }
    },
    
    updateMessageReaders(messageId, readers) {
      for (const [chatId, messages] of this.messagesCache) {
        const messageIndex = messages.findIndex(m => m.id === messageId)
        if (messageIndex !== -1) {
          const updatedMessages = [...messages]
          updatedMessages[messageIndex] = {
            ...updatedMessages[messageIndex],
            readers: readers || []
          }
          this.messagesCache.set(chatId, updatedMessages)
        }
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
    },

    async addUserToChat(chatId, phone) {
      try {
        const response = await api.post(`/chats/${chatId}/members`, { phone })
        return { success: true, data: response.data }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Ошибка добавления пользователя' }
      }
    },

    async leaveChat(chatId, permanent = false) {
      try {
        const url = permanent 
          ? `/chats/${chatId}/leave?permanent=true`
          : `/chats/${chatId}/leave`
        
        const response = await api.delete(url)
        this.chats = this.chats.filter(chat => chat.id !== chatId)
        this.messagesCache.delete(chatId)
        if (this.activeChatId === chatId) {
          this.setActiveChat(null)
        }
        
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data || 'Ошибка выхода из чата' 
        }
      }
    },

    async deleteChatForAll(chatId) {
      try {
        const response = await api.delete(`/chats/${chatId}?forAll=true`)
        this.chats = this.chats.filter(chat => chat.id !== chatId)
        this.messagesCache.delete(chatId)
        if (this.activeChatId === chatId) {
          this.setActiveChat(null)
        }
        
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data || 'Ошибка удаления чата' 
        }
      }
    },

    async removeChatMember(chatId, memberId) {
      try {
        const response = await api.delete(`/chats/${chatId}/members/${memberId}`)
        
        await this.fetchChat(chatId)
        
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data || 'Ошибка удаления участника' 
        }
      }
    },

    async reloadChatIfNeeded(chatId) {
      const existingChat = this.chats.find(chat => chat.id === chatId)
      if (!existingChat) {
        try {
          await this.fetchChat(chatId)
        } catch (error) {
          console.error('Ошибка перезагрузки чата:', error)
        }
      }
    },
    
    saveCurrentScrollPosition() {},
    
    saveScrollPosition(chatId, position) {
      this.scrollPositions.set(chatId, position)
      this.saveToLocalStorage()
    },
    
    getScrollPosition(chatId) {
      let position = this.scrollPositions.get(chatId) || 0
      if (position === 0) {
        position = this.loadFromLocalStorage()?.scrollPositions?.[chatId] || 0
      }
      return position
    },
    
    clearScrollPosition(chatId) {
      this.scrollPositions.delete(chatId)
      this.saveToLocalStorage()
    },
    
    saveToLocalStorage() {
      try {
        const data = {
          scrollPositions: Object.fromEntries(this.scrollPositions),
          messagesCacheTime: Object.fromEntries(this.messagesCacheTime)
        }
        localStorage.setItem('chatsCache', JSON.stringify(data))
      } catch (error) {
        console.error('Ошибка сохранения в localStorage:', error)
      }
    },
    
    loadFromLocalStorage() {
      try {
        const data = localStorage.getItem('chatsCache')
        return data ? JSON.parse(data) : null
      } catch (error) {
        console.error('Ошибка загрузки из localStorage:', error)
        return null
      }
    },
    
    initializeFromStorage() {
      const data = this.loadFromLocalStorage()
      if (data) {
        if (data.scrollPositions) {
          this.scrollPositions = new Map(Object.entries(data.scrollPositions))
        }
        if (data.messagesCacheTime) {
          this.messagesCacheTime = new Map(Object.entries(data.messagesCacheTime))
        }
        console.log('Данные восстановлены из localStorage')
      }
    },
    
    async preloadMessages() {
      console.log('Предзагрузка сообщений для активных чатов...')
      
      const chatsToPreload = this.chats.slice(0, 3)
      
      for (const chat of chatsToPreload) {
        if (this.shouldRefreshMessages(chat.id)) {
          console.log('Предзагружаем сообщения для чата', chat.id)
          await this.fetchMessages(chat.id)
        }
      }
    },
    
    async getMessages(chatId, forceRefresh = false) {
      const cachedMessages = this.messagesCache.get(chatId)
      const shouldRefresh = this.shouldRefreshMessages(chatId)
      
      if (cachedMessages && !shouldRefresh && !forceRefresh) {
        console.log('Возвращаем кешированные сообщения для чата', chatId)
        return cachedMessages
      }
      
      const result = await this.fetchMessages(chatId, forceRefresh)
      if (result.success) {
        return this.messagesCache.get(chatId) || []
      }
      
      return []
    },
    
    clearChatCache(chatId) {
      this.messagesCache.delete(chatId)
      this.messagesCacheTime.delete(chatId)
      this.scrollPositions.delete(chatId)
      console.log('Кеш очищен для чата', chatId)
    }
  }
})