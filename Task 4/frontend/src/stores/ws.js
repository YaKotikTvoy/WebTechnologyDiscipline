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

      switch (data.type) {
        case 'message':
          chatsStore.addMessage(data.data.message)
          break
        case 'friend_request':
          friendsStore.fetchFriendRequests()
          this.notifications.push({
            id: Date.now(),
            type: 'friend_request',
            data: data.data,
            read: false
          })
          break
        case 'chat_invite':
          chatsStore.fetchChats()
          this.notifications.push({
            id: Date.now(),
            type: 'chat_invite',
            data: data.data,
            read: false
          })
          break
      }
    },

    markAsRead(notificationId) {
      const index = this.notifications.findIndex(n => n.id === notificationId)
      if (index !== -1) {
        this.notifications[index].read = true
      }
    },

    clearNotifications() {
      this.notifications = []
    }
  }
})