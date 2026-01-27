import { defineStore } from 'pinia'
import { useChatsStore } from './chats'
import { useFriendsStore } from './friends'
import { useAuthStore } from './auth'

export const useWebSocketStore = defineStore('websocket', {
  state: () => ({
    ws: null,
    isConnected: false,
    notifications: []
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
          break
          
        case 'friend_request':
          friendsStore.fetchFriendRequests()
          this.addNotification({
            id: Date.now(),
            type: 'friend_request',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          break
          
        case 'chat_invite':
          this.addNotification({
            id: Date.now(),
            type: 'chat_invite',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          break
          
        case 'friend_request_accepted':
          this.addNotification({
            id: Date.now(),
            type: 'friend_request_accepted',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          friendsStore.fetchFriends()
          break
          
        case 'friend_request_rejected':
          this.addNotification({
            id: Date.now(),
            type: 'friend_request_rejected',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          break
          
        case 'chat_invite_accepted':
          this.addNotification({
            id: Date.now(),
            type: 'chat_invite_accepted',
            data: data.data,
            read: false,
            createdAt: new Date().toISOString()
          })
          chatsStore.fetchChats()
          break
          
        case 'chat_invite_rejected':
          this.addNotification({
            id: Date.now(),
            type: 'chat_invite_rejected',
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
      this.notifications.unshift(notification)
      this.saveNotifications()
    },

    markAsRead(notificationId) {
      const index = this.notifications.findIndex(n => n.id === notificationId)
      if (index !== -1) {
        this.notifications.splice(index, 1)
        this.saveNotifications()
      }
    },

    saveNotifications() {
      localStorage.setItem('notifications', JSON.stringify(this.notifications))
    },

    loadNotifications() {
      const saved = localStorage.getItem('notifications')
      if (saved) {
        this.notifications = JSON.parse(saved)
      }
    },

    clearNotifications() {
      this.notifications = []
      localStorage.removeItem('notifications')
    }
  }
})