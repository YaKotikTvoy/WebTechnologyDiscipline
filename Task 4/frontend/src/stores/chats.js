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
        return []
      }
      const cached = state.messagesCache.get(state.activeChatId)
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
    saveScrollPosition(chatId, position) {
      this.scrollPositions.set(chatId, position)
    },
    
    getScrollPosition(chatId) {
      return this.scrollPositions.get(chatId) || 0
    },
    
    removeChatSynchronously(chatId) {
      this.chats = this.chats.filter(chat => chat.id !== chatId)
      this.messagesCache.delete(chatId)
      this.messagesCacheTime.delete(chatId)
      this.scrollPositions.delete(chatId)
      
      if (this.activeChatId === chatId) {
          this.activeChatId = null
          this.currentChat = null
      }
    },
    
    updateChatUnreadCountFromWS(chatId, count) {
        const chatIndex = this.chats.findIndex(c => c.id === chatId)
        if (chatIndex !== -1) {
            this.chats[chatIndex] = {
              ...this.chats[chatIndex],
              unread_count: count || 0
            }
        } else {
            setTimeout(() => {
                this.fetchChats()
            }, 300)
        }
    },
    
    async fetchChats() {
      try {
        const response = await api.get('/chats')
        
        this.chats = response.data.map(chat => ({
          ...chat,
          unread_count: chat.unread_count || 0,
          last_message: chat.last_message || chat.lastMessage || null,
          members_count: chat.members?.length || 0
        }))
        
        return { success: true, chats: this.chats }
      } catch (error) {
        console.error('Ошибка загрузки чатов:', error)
        return { success: false, error: 'Ошибка загрузки чатов' }
      }
    },
    
    setActiveChat(chatId) {
        if (this.activeChatId && this.activeChatId !== chatId) {
        }
        
        this.activeChatId = chatId
        if (chatId) {
            const chat = this.chats.find(c => c.id === chatId)
            if (chat) {
                this.currentChat = {...chat}
            } else {
                this.currentChat = null
                this.activeChatId = null
            }
        } else {
            this.currentChat = null
        }
    },
    
    async fetchChat(chatId) {
      try {
        const response = await api.get(`/chats/${chatId}`)
        this.currentChat = {...response.data}
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
        const cachedMessages = this.messagesCache.get(chatId)
        return { success: true, cached: true, messages: cachedMessages }
      }
      
      this.isLoadingMessages = true
      try {
        const response = await api.get(`/chats/${chatId}/messages?limit=200`)
        this.messagesCache.set(chatId, [...response.data])
        this.messagesCacheTime.set(chatId, Date.now())
        return { success: true, messages: response.data }
      } catch {
        return { success: false, error: 'Ошибка загрузки сообщений' }
      } finally {
        this.isLoadingMessages = false
      }
    },

    async getMessages(chatId, forceRefresh = false) {
      const cachedMessages = this.messagesCache.get(chatId)
      const shouldRefresh = this.shouldRefreshMessages(chatId)
      
      if (forceRefresh || !cachedMessages || shouldRefresh) {
        const result = await this.fetchMessages(chatId, forceRefresh)
        if (result.success) {
          return this.messagesCache.get(chatId) || []
        }
      }
      
      return cachedMessages || []
    },

    async markSpecificChatAsRead(chatId) {
      try {
        const chat = this.chats.find(c => c.id === chatId)
        if (!chat) {
          return { success: false, error: 'Чат не найден' }
        }
        
        await api.post(`/chats/${chatId}/read`)
        
        const chatIndex = this.chats.findIndex(c => c.id === chatId)
        if (chatIndex !== -1) {
          this.chats[chatIndex] = {
            ...this.chats[chatIndex],
            unread_count: 0
          }
        }
        
        return { success: true }
      } catch (error) {
        console.error('Ошибка пометки чата как прочитанного:', error)
        return { success: false, error: error.message }
      }
    },

    async markChatAsRead(chatId) {
      try {
        await api.post(`/chats/${chatId}/read`)
        return { success: true }
      } catch (error) {
        console.error('Ошибка пометки чата как прочитанного:', error)
        return { success: false, error: error.message }
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
            
            const message = response.data
            
            const currentMessages = this.messagesCache.get(chatId) || []
            const existingIndex = currentMessages.findIndex(m => m.id === message.id)
            
            if (existingIndex === -1) {
                const authStore = useAuthStore()
                const userId = authStore.user?.id
                
                const newMessage = {...message}
                
                if (userId) {
                    if (!newMessage.readers) {
                        newMessage.readers = []
                    }
                    
                    const alreadyRead = newMessage.readers.some(r => r.id === userId)
                    if (!alreadyRead) {
                        newMessage.readers = [...newMessage.readers, {
                            id: userId,
                            read_at: new Date().toISOString()
                        }]
                    }
                }
                
                const updatedMessages = [...currentMessages, newMessage]
                this.messagesCache.set(chatId, updatedMessages)
            }
            
            return { success: true, message: message }
        } catch (error) {
            console.error('Ошибка отправки сообщения:', error)
            return { success: false, error: error.response?.data || 'Ошибка отправки сообщения' }
        }
    },

    async deleteMessage(chatId, messageId) {
      try {
        const response = await api.delete(`/chats/${chatId}/messages/${messageId}`)
        
        const messages = this.messagesCache.get(chatId) || []
        const messageIndex = messages.findIndex(m => m.id === messageId)
        
        if (messageIndex !== -1) {
          const updatedMessages = [...messages]
          updatedMessages[messageIndex] = {
            ...updatedMessages[messageIndex],
            is_deleted: true,
            content: '[Сообщение удалено]'
          }
          this.messagesCache.set(chatId, updatedMessages)
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
          const updatedMessages = [...messages]
          updatedMessages[messageIndex] = {
            ...updatedMessages[messageIndex],
            content: content,
            is_edited: true,
            updated_at: new Date().toISOString()
          }
          this.messagesCache.set(chatId, updatedMessages)
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
      const chatId = message.chat_id || message.chatId
      if (chatId) {
        const messages = this.messagesCache.get(chatId) || []
        const existingIndex = messages.findIndex(m => m.id === message.id)
        
        if (existingIndex === -1) {
          const updatedMessages = [...messages, {...message}]
          this.messagesCache.set(chatId, updatedMessages)
        } else {
          const updatedMessages = [...messages]
          updatedMessages[existingIndex] = {...message}
          this.messagesCache.set(chatId, updatedMessages)
        }
      }
    },

    updateMessageReadOptimistically(messageId, readerId) {
      const authStore = useAuthStore()
      const messages = this.messagesCache.get(this.activeChatId) || []
      const messageIndex = messages.findIndex(m => m.id === messageId)
      
      if (messageIndex !== -1) {
        const message = {...messages[messageIndex]}
        
        if (!message.readers) {
          message.readers = []
        }
        
        const readerExists = message.readers.some(r => r.id === readerId)
        
        if (!readerExists) {
          message.readers = [...message.readers, {
            id: readerId,
            read_at: new Date().toISOString()
          }]
          
          const updatedMessages = [...messages]
          updatedMessages[messageIndex] = message
          this.messagesCache.set(this.activeChatId, updatedMessages)
        }
      }
    },

    updateChatUnreadCount(chatId, count) {
        const chatIndex = this.chats.findIndex(c => c.id === chatId)
        if (chatIndex !== -1) {
            this.chats[chatIndex] = {
              ...this.chats[chatIndex],
              unread_count: count
            }
        }
    },
    
    updateMessageReaders(messageId, readers) {
      for (const [chatId, messages] of this.messagesCache) {
        const messageIndex = messages.findIndex(m => m.id === messageId)
        if (messageIndex !== -1) {
          const updatedMessages = [...messages]
          updatedMessages[messageIndex] = {
            ...updatedMessages[messageIndex],
            readers: [...(readers || [])]
          }
          this.messagesCache.set(chatId, updatedMessages)
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
          error: error.response?.data
        }
      }
    }
  }
})