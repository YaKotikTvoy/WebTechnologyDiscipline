<template>
  <div class="chat-container">
    <div class="card">
      <div class="card-header">
        <h5 class="mb-0">{{ chat.name }}</h5>
        <small v-if="chat.description">{{ chat.description }}</small>
      </div>
      <div class="card-body messages-container">
        <div v-for="message in chatStore.messages" :key="message.id" class="message mb-2">
          <div class="d-flex">
            <div class="message-content p-2 rounded" 
                 :class="{ 'bg-primary text-white': message.sender_id === authStore.user.id, 'bg-light': message.sender_id !== authStore.user.id }">
              <div class="message-header mb-1">
                <small class="fw-bold">{{ message.sender_username || 'Пользователь' }}</small>
                <small class="text-muted ms-2">{{ formatTime(message.created_at) }}</small>
                <span v-if="message.is_edited" class="badge bg-secondary ms-2">изменено</span>
              </div>
              <div class="message-text">{{ message.content }}</div>
              <div v-if="message.files?.length" class="mt-2">
                <a v-for="file in message.files" :key="file.id" :href="file.url" target="_blank" class="d-block">
                  📎 {{ file.name }}
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="card-footer">
        <form @submit.prevent="sendMessage" class="d-flex">
          <input v-model="newMessage" type="text" class="form-control me-2" placeholder="Введите сообщение...">
          <button type="submit" class="btn btn-primary">Отправить</button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const route = useRoute()
const chatStore = useChatStore()
const authStore = useAuthStore()
const newMessage = ref('')
const chat = ref({})
let intervalId = null

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  loadChat()
  loadMessages()
  
  intervalId = setInterval(() => {
    if (route.params.id) {
      loadMessages()
    }
  }, 5000)
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})

watch(() => route.params.id, () => {
  if (route.params.id) {
    loadChat()
    loadMessages()
  }
})

const loadChat = async () => {
  try {
    const response = await api.get(`/chats/${route.params.id}`)
    chat.value = response.data
  } catch (error) {
    console.error('Ошибка загрузки чата:', error)
  }
}

const loadMessages = () => {
  chatStore.fetchMessages(route.params.id)
}

const sendMessage = async () => {
  if (!newMessage.value.trim()) return
  
  const result = await chatStore.sendMessage(route.params.id, newMessage.value)
  
  if (result.success) {
    newMessage.value = ''
  }
}
</script>

<style scoped>
.messages-container {
  height: 60vh;
  overflow-y: auto;
}
.message-content {
  max-width: 70%;
}
</style>