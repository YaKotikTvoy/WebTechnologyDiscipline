import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../services/api'

export const useChatStore = defineStore('chat', () => {
  const chats = ref([])
  const currentChat = ref(null)
  const messages = ref([])

  const fetchChats = async () => {
    try {
      const response = await api.get('/chats')
      chats.value = response.data
    } catch (error) {
      console.error('Ошибка загрузки чатов:', error)
    }
  }

  const fetchPublicChats = async () => {
    try {
      const response = await api.get('/public-chats')
      return response.data
    } catch (error) {
      console.error('Ошибка загрузки публичных чатов:', error)
      return []
    }
  }

  const createChat = async (chatData) => {
    try {
      const response = await api.post('/chats', chatData)
      chats.value.push(response.data)
      return { success: true, data: response.data }
    } catch (error) {
      return { success: false, error: error.response?.data?.error || 'Ошибка создания чата' }
    }
  }

  const fetchMessages = async (chatId) => {
    try {
      const response = await api.get(`/messages?chat_id=${chatId}`)
      messages.value = response.data.messages
    } catch (error) {
      console.error('Ошибка загрузки сообщений:', error)
    }
  }

  const sendMessage = async (chatId, content) => {
    try {
      const response = await api.post('/messages', {
        chat_id: chatId,
        content
      })
      messages.value.push(response.data)
      return { success: true }
    } catch (error) {
      return { success: false, error: error.response?.data?.error || 'Ошибка отправки' }
    }
  }

  return {
    chats,
    currentChat,
    messages,
    fetchChats,
    fetchPublicChats,
    createChat,
    fetchMessages,
    sendMessage
  }
})