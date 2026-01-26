<template>
  <div class="container-fluid mt-3">
    <div class="card">
      <div class="card-header">
        <h5 class="mb-0">Личные сообщения с {{ contactName }}</h5>
      </div>
      <div class="card-body messages-container" ref="messagesContainer">
        <div v-if="loading" class="text-center">
          <div class="spinner-border"></div>
        </div>
        <div v-else-if="messages.length === 0" class="text-center text-muted">
          Нет сообщений
        </div>
        <div v-else>
          <div v-for="message in messages" :key="message.id" class="message mb-3">
            <div class="d-flex align-items-start">
              <div class="flex-grow-1">
                <div class="d-flex justify-content-between align-items-start">
                  <div>
                    <span class="fw-bold">{{ message.sender_username || 'Пользователь' }}</span>
                    <small class="text-muted ms-2">{{ formatTime(message.created_at) }}</small>
                    <span v-if="message.is_edited" class="badge bg-secondary ms-2">изменено</span>
                  </div>
                </div>
                <div class="message-content p-3 rounded mt-2" 
                     :class="{ 'bg-primary text-white': message.sender_id === authStore.user.id, 'bg-light': message.sender_id !== authStore.user.id }">
                  {{ message.content }}
                </div>
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
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const route = useRoute()
const authStore = useAuthStore()
const contactId = route.params.id
const messages = ref([])
const newMessage = ref('')
const contactName = ref('Пользователь')
const loading = ref(false)
const messagesContainer = ref(null)
let intervalId = null

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const loadMessages = async () => {
  try {
    const response = await api.get(`/direct-messages?contact_id=${contactId}`)
    messages.value = response.data.messages
    scrollToBottom()
  } catch (error) {
    console.error('Ошибка загрузки сообщений:', error)
  }
}

const loadContactInfo = async () => {
  try {
    const response = await api.get(`/users/${contactId}`)
    contactName.value = response.data.username || 'Пользователь'
  } catch (error) {
    console.error('Ошибка загрузки информации о контакте:', error)
  }
}

const sendMessage = async () => {
  if (!newMessage.value.trim()) return
  
  try {
    const response = await api.post('/messages', {
      recipient_id: contactId,
      content: newMessage.value
    })
    
    messages.value.push(response.data)
    newMessage.value = ''
    scrollToBottom()
  } catch (error) {
    console.error('Ошибка отправки сообщения:', error)
  }
}

onMounted(() => {
  loadContactInfo()
  loadMessages()
  
  intervalId = setInterval(() => {
    loadMessages()
  }, 3000)
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
})
</script>

<style scoped>
.messages-container {
  height: 60vh;
  overflow-y: auto;
}
.message-content {
  max-width: 100%;
  word-break: break-word;
}
</style>