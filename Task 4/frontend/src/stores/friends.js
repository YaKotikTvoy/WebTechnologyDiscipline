import { defineStore } from 'pinia'
import { api } from '@/services/api'

export const useFriendsStore = defineStore('friends', {
  state: () => ({
    friends: [],
    friendRequests: [],
    searchResults: []
  }),

  actions: {
    async fetchFriends() {
      try {
        const response = await api.get('/friends')
        this.friends = response.data
      } catch (error) {
        console.error('Failed to fetch friends:', error)
      }
    },

    async fetchFriendRequests() {
      try {
        const response = await api.get('/friends/requests')
        this.friendRequests = response.data
      } catch (error) {
        console.error('Failed to fetch friend requests:', error)
      }
    },

    async sendFriendRequest(phone) {
      try {
        await api.post('/friends/requests', { recipient_phone: phone })
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to send request' }
      }
    },

    async respondToRequest(requestId, status) {
      try {
        await api.put(`/friends/requests/${requestId}`, { status })
        await this.fetchFriendRequests()
        await this.fetchFriends()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to respond' }
      }
    },

    async removeFriend(friendId) {
      try {
        await api.delete(`/friends/${friendId}`)
        await this.fetchFriends()
        return { success: true }
      } catch (error) {
        return { success: false, error: error.response?.data || 'Failed to remove friend' }
      }
    },

    async searchUser(phone) {
      try {
        const response = await api.get(`/users/search?phone=${phone}`)
        this.searchResults = [response.data]
        return { success: true, user: response.data }
      } catch (error) {
        this.searchResults = []
        return { success: false, error: error.response?.data || 'User not found' }
      }
    }
  }
})