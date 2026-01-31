import { defineStore } from 'pinia'
import { useChatsStore } from './chats'
import { useAuthStore } from './auth'
import { api } from '@/services/api'

if (!window.messageTracker) {
    window.messageTracker = new Map()
}

export const useWebSocketStore = defineStore('websocket', {
  state: () => ({
    ws: null,
    isConnected: false,
    notifications: [],
    reconnectAttempts: 0,
    maxReconnectAttempts: 5
  }),

  actions: {
    connect() {
      const authStore = useAuthStore()
      if (!authStore.token) return

      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        return
      }

      if (this.ws) {
        this.ws.close()
      }

      const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://'
      const wsUrl = `${protocol}localhost:8080/ws?token=${authStore.token}`
      
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.isConnected = true
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleMessage(data)
        } catch {
        }
      }

      this.ws.onclose = () => {
        this.isConnected = false
        this.ws = null
        
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.reconnectAttempts++
          setTimeout(() => {
            if (authStore.isAuthenticated) {
              this.connect()
            }
          }, 3000 * this.reconnectAttempts)
        }
      }

      this.ws.onerror = () => {
      }
    },

    disconnect() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.isConnected = false
      this.reconnectAttempts = 0
    },

    handleMessage(data) {
      const chatsStore = useChatsStore()
      const authStore = useAuthStore()

      const now = Date.now()
      const messageKey = `${data.type}_${data.data?.chat_id}_${data.data?.message?.id || data.data?.message_id}`
      
      const lastTime = window.messageTracker.get(messageKey) || 0
      if (now - lastTime < 100) {
        return
      }
      
      window.messageTracker.set(messageKey, now)
      setTimeout(() => {
        window.messageTracker.delete(messageKey)
      }, 5000)

      if (data.type === 'new_message') {
        const messageData = data.data
        const currentUserId = this.getCurrentUserID()
        
        const messageId = messageData.message?.id || messageData.message_id
        if (messageId && window.processedMessageIds?.has(messageId)) {
          return
        }
        
        if (messageId) {
          if (!window.processedMessageIds) {
            window.processedMessageIds = new Set()
          }
          window.processedMessageIds.add(messageId)
          setTimeout(() => {
            window.processedMessageIds.delete(messageId)
          }, 30000)
        }
        
        const isFromMe = messageData.sender?.id === currentUserId
        const isInActiveChat = chatsStore.activeChatId === messageData.chat_id
        
        if (!isFromMe && !isInActiveChat) {
          const chatIndex = chatsStore.chats.findIndex(c => c.id === messageData.chat_id)
          if (chatIndex !== -1) {
            const currentCount = chatsStore.chats[chatIndex].unreadCount || 0
            chatsStore.chats[chatIndex].unreadCount = currentCount + 1
            
            if (messageData.message) {
              chatsStore.chats[chatIndex].lastMessage = messageData.message
              chatsStore.chats[chatIndex].updated_at = new Date().toISOString()
            }
          }
        }
        
        if (isInActiveChat && messageData.message) {
          const messages = chatsStore.messagesCache.get(messageData.chat_id) || []
          const existingIndex = messages.findIndex(m => m.id === messageData.message.id)
          
          if (existingIndex === -1) {
            messages.push(messageData.message)
            chatsStore.messagesCache.set(messageData.chat_id, messages)
          }
        }
        
        return
      }

      if (data.type === 'chat_created') {
          const chatData = data.data
          
          setTimeout(async () => {
              await chatsStore.fetchChats()
          }, 500)
          
          this.addNotification({
              id: Date.now(),
              type: 'chat_created',
              data: chatData,
              read: false,
              createdAt: new Date().toISOString()
          })
          
          return
      }
      
      if (data.type === 'message_read') {
        const readData = data.data
        const messageID = readData.message_id
        const readerID = readData.reader_id
        
        this.updateMessageReadStatus(messageID, readerID)
      }

      if (data.type === 'chat_invite') {
        this.addNotification({
          id: Date.now(),
          type: 'chat_invite',
          data: data.data,
          read: false,
          createdAt: new Date().toISOString()
        })
      }

      if (data.type === 'chat_join_request') {
        this.addNotification({
          id: Date.now(),
          type: 'chat_join_request',
          data: data.data,
          read: false,
          createdAt: new Date().toISOString()
        })
      }
      
      if (data.type === 'message_deleted') {
        const deleteData = data.data
        const deletedChatID = deleteData.chat_id
        const deletedMessageID = deleteData.message_id
        
        if (chatsStore.activeChatId === deletedChatID) {
          const messages = chatsStore.messagesCache.get(deletedChatID) || []
          const messageIndex = messages.findIndex(m => m.id === deletedMessageID)
          
          if (messageIndex !== -1) {
            messages[messageIndex] = {
              ...messages[messageIndex],
              is_deleted: true,
              content: '[Сообщение удалено]'
            }
            chatsStore.messagesCache.set(deletedChatID, [...messages])
          }
        }
      }
      
      if (data.type === 'chat_deleted') {
        const deleteData = data.data
        const deletedChatId = deleteData.chat_id
        
        chatsStore.removeChatSynchronously(deletedChatId)
        
        this.addNotification({
          id: Date.now(),
          type: 'chat_deleted',
          data: deleteData,
          read: false,
          createdAt: new Date().toISOString()
        })
      }
    },

    updateMessageReadStatus(messageId, readerId) {
      const chatsStore = useChatsStore()
      
      for (const [chatId, messages] of chatsStore.messagesCache) {
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
            
            const updatedMessages = [...messages]
            updatedMessages[messageIndex] = { ...message }
            chatsStore.messagesCache.set(chatId, updatedMessages)
          }
        }
      }
    },

    async markMessageAsRead(chatId, messageId) {
      try {
        await api.post(`/chats/${chatId}/messages/${messageId}/read`)
      } catch (error) {
      }
    },

    addNotification(notification) {
      this.notifications.unshift(notification)
      localStorage.setItem('notifications', JSON.stringify(this.notifications))
    },
    
    getCurrentUserID() {
      const authStore = useAuthStore()
      return authStore.user?.id
    }
  }
})