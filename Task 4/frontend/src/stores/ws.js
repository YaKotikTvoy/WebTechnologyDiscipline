import { defineStore } from 'pinia'
import { useChatsStore } from './chats'
import { useFriendsStore } from './friends'
import { useAuthStore } from './auth'

export const useWebSocketStore = defineStore('websocket', {
  state: () => ({
    ws: null,
    isConnected: false,
    notifications: [],
    unreadCounts: {}
  }),

  getters: {
    unreadNotifications: (state) => {
      return state.notifications.filter(n => !n.read).length
    }
  },

  actions: {
    connect() {
      const authStore = useAuthStore()
      if (!authStore.token || this.ws) return

      const wsUrl = `ws://localhost:8080/ws?token=${authStore.token}`
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = () => {
        this.isConnected = true
        console.log('WebSocket connected')
      }

      this.ws.onmessage = (event) => {
        const data = JSON.parse(event.data)
        this.handleMessage(data)
      }

      this.ws.onclose = () => {
        this.isConnected = false
        this.ws = null
        setTimeout(() => this.connect(), 3000)
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    },

    disconnect() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
        this.isConnected = false
      }
    },

    handleMessage(data) {
      const chatsStore = useChatsStore()
      const friendsStore = useFriendsStore()
      const authStore = useAuthStore()

      switch (data.type) {
        case 'message':
          chatsStore.addMessage(data.data.message)
          
          if (data.data.message.sender_id !== authStore.user?.id) {
            this.addNotification({
              type: 'chat_message',
              data: {
                chatId: data.data.message.chat_id,
                chatName: data.data.chatName || 'Чат',
                senderName: data.data.message.sender?.username || data.data.message.sender?.phone,
                content: data.data.message.content,
                messageId: data.data.message.id
              },
              read: false,
              createdAt: new Date().toISOString()
            })
          }
          break
          
        case 'friend_request':
          friendsStore.fetchFriendRequests()
          this.addNotification({
            type: 'friend_request',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          break
          
        case 'chat_invite':
          chatsStore.fetchChats()
          this.addNotification({
            type: 'chat_invite',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          break
          
        case 'chat_update':
          chatsStore.fetchChats()
          break
      }
    },

    addNotification(notification) {
      const notificationId = Date.now() + Math.random()
      this.notifications.unshift({
        id: notificationId,
        ...notification
      })
      
      localStorage.setItem('notifications', JSON.stringify(this.notifications))
    },

    markAsRead(notificationId) {
      const index = this.notifications.findIndex(n => n.id === notificationId)
      if (index !== -1) {
        this.notifications[index].read = true
        localStorage.setItem('notifications', JSON.stringify(this.notifications))
      }
    },

    markNotificationAsReadByData(type, dataId) {
      this.notifications.forEach(notification => {
        if (notification.type === type && notification.data.id === dataId) {
          notification.read = true
        }
      })
      localStorage.setItem('notifications', JSON.stringify(this.notifications))
    },

    clearNotifications() {
      this.notifications = []
      localStorage.removeItem('notifications')
    },

    loadNotifications() {
      const saved = localStorage.getItem('notifications')
      if (saved) {
        this.notifications = JSON.parse(saved)
      }
    }
  }
})