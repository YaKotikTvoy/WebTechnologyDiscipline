import { defineStore } from 'pinia'
import { useChatsStore } from './chats'
import { useAuthStore } from './auth'
import { api } from '@/services/api'

export const useWebSocketStore = defineStore('websocket', {
  state: () => ({
    ws: null,
    isConnected: false,
    notifications: [],
    reconnectAttempts: 0,
    maxReconnectAttempts: 5
  }),

  getters: {
    unreadNotifications: (state) => {
      return state.notifications.filter(n => !n.read).length
    },
    pendingRequests: (state) => {
      return state.notifications.filter(n => 
        !n.read && (n.type === 'chat_invite' || n.type === 'chat_join_request')
      ).length
    }
  },

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

      if (data.type === 'new_message') {
        const messageData = data.data
        const isFromMe = messageData.sender?.id === authStore.user?.id
        const isInThisChat = chatsStore.currentChat?.id === messageData.chat_id
        
        if (!isFromMe && !isInThisChat) {
          this.updateUnreadCount(messageData.chat_id, messageData.unread_count)
        }
        
        if (chatsStore.currentChat?.id === messageData.chat_id) {
          if (messageData.message) {
            chatsStore.addMessage(messageData.message)
          }
        }
      }

      if (data.type === 'message_read') {
        const readData = data.data
        const chatID = readData.chat_id
        const messageID = readData.message_id
        const readerID = readData.reader_id
        
        if (chatsStore.currentChat?.id === chatID) {
          chatsStore.updateMessageReadOptimistically(messageID, readerID)
        }
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
      
      if (data.type === 'added_to_group') {
        const groupData = data.data
        const chatID = groupData.chat_id
        
        const chatsStore = useChatsStore()
        chatsStore.fetchChats() 
        
        this.addNotification({
          id: Date.now(),
          type: 'added_to_group',
          data: groupData,
          read: false,
          createdAt: new Date().toISOString()
        })
      }
      
      if (data.type === 'chat_created') {
        const chatData = data.data
        console.log('WS: Получен новый чат', chatData)
        
        const chatsStore = useChatsStore()
        
        chatsStore.fetchChats().then(() => {
          if (chatData.unread_count > 0) {
            chatsStore.updateChatUnreadCount(chatData.chat_id, chatData.unread_count)
          }
        })
        
        this.addNotification({
          id: Date.now(),
          type: 'chat_created',
          data: chatData,
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

      if (data.type === 'message_edited') {
        const editData = data.data
        const editedChatID = editData.chat_id
        const editedMessage = editData.message
        
        if (chatsStore.activeChatId === editedChatID) {
          const messages = chatsStore.messagesCache.get(editedChatID) || []
          const messageIndex = messages.findIndex(m => m.id === editedMessage.id)
          
          if (messageIndex !== -1) {
            messages[messageIndex] = {
              ...messages[messageIndex],
              ...editedMessage,
              is_edited: true
            }
            chatsStore.messagesCache.set(editedChatID, [...messages])
          }
        }
      }
    },

    updateUnreadCount(chatId, count) {
      const chatsStore = useChatsStore()
      const chatIndex = chatsStore.chats.findIndex(c => c.id === chatId)
      
      if (chatIndex !== -1) {
        chatsStore.chats[chatIndex].unreadCount = count
      }
    },

    async markMessageAsRead(chatId, messageId) {
      try {
        await api.post(`/chats/${chatId}/messages/${messageId}/read`)
      } catch {
      }
    },

    updateMessageReadStatus(messageId, readerId) {
      const chatsStore = useChatsStore()
      const authStore = useAuthStore()
      
      chatsStore.updateMessageReadOptimistically(messageId, readerId)
    },

    addNotification(notification) {
      this.notifications.unshift(notification)
      localStorage.setItem('notifications', JSON.stringify(this.notifications))
    }
  }
})