<template>
  <div class="chat-container">
    <div class="card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <div>
          <h5 class="mb-0">{{ chat.name }}</h5>
          <small v-if="chat.description" class="text-muted">{{ chat.description }}</small>
        </div>
        <div v-if="isCreator" class="dropdown">
          <button class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
            Управление чатом
          </button>
          <ul class="dropdown-menu">
            <li><button @click="showEditChatModal = true" class="dropdown-item">Изменить чат</button></li>
            <li><button @click="showInviteModal = true" class="dropdown-item">Пригласить пользователя</button></li>
            <li><button @click="showMembersModal = true" class="dropdown-item">Участники</button></li>
          </ul>
        </div>
      </div>
      <div class="card-body messages-container" ref="messagesContainer">
        <div v-for="message in chatStore.messages" :key="message.id" class="message mb-3">
          <div class="d-flex align-items-start">
            <div v-if="message.sender_avatar" class="me-2">
              <img :src="message.sender_avatar" alt="Аватар" class="rounded-circle" style="width: 40px; height: 40px; object-fit: cover;">
            </div>
            <div class="flex-grow-1">
              <div class="d-flex justify-content-between align-items-start">
                <div>
                  <span class="fw-bold">{{ message.sender_username || 'Пользователь' }}</span>
                  <small class="text-muted ms-2">{{ formatTime(message.created_at) }}</small>
                  <span v-if="message.is_edited" class="badge bg-secondary ms-2">изменено</span>
                </div>
                <div v-if="message.sender_id === authStore.user.id || isCreator" class="dropdown">
                  <button class="btn btn-sm btn-link dropdown-toggle" type="button" data-bs-toggle="dropdown">
                    ⋮
                  </button>
                  <ul class="dropdown-menu">
                    <li v-if="message.sender_id === authStore.user.id">
                      <button @click="startEditMessage(message)" class="dropdown-item">Редактировать</button>
                    </li>
                    <li>
                      <button @click="deleteMessage(message.id)" class="dropdown-item text-danger">Удалить</button>
                    </li>
                  </ul>
                </div>
              </div>
              
              <div v-if="editingMessageId === message.id" class="mt-2">
                <textarea v-model="editMessageContent" class="form-control mb-2" rows="3"></textarea>
                <div class="d-flex gap-2">
                  <button @click="saveEditMessage(message.id)" class="btn btn-sm btn-primary">Сохранить</button>
                  <button @click="cancelEditMessage" class="btn btn-sm btn-secondary">Отмена</button>
                </div>
              </div>
              <div v-else class="message-content p-3 rounded mt-2" 
                   :class="{ 'bg-primary text-white': message.sender_id === authStore.user.id, 'bg-light': message.sender_id !== authStore.user.id }">
                <div class="message-text">{{ message.content }}</div>
                
                <div v-if="message.files?.length" class="mt-3">
                  <div v-for="file in message.files" :key="file.id" class="file-item mb-2">
                    <div v-if="isImage(file.mime_type)" class="mb-2">
                      <a :href="file.url" target="_blank" class="d-block">
                        <img :src="file.url" :alt="file.name" class="img-thumbnail" style="max-width: 200px; max-height: 200px;">
                      </a>
                    </div>
                    <div v-else class="d-flex align-items-center">
                      <span class="me-2">📎</span>
                      <a :href="file.url" target="_blank" class="text-decoration-none">
                        {{ file.name }} ({{ formatFileSize(file.size) }})
                      </a>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="card-footer">
        <form @submit.prevent="sendMessage" class="d-flex align-items-center">
          <div class="me-2">
            <label class="btn btn-outline-secondary btn-sm mb-0" title="Добавить файл">
              <input type="file" @change="handleFileUpload" ref="fileInput" class="d-none">
              📎
            </label>
            <label class="btn btn-outline-secondary btn-sm mb-0 ms-1" title="Добавить изображение">
              <input type="file" @change="handleImageUpload" ref="imageInput" class="d-none" accept="image/*">
              🖼️
            </label>
          </div>
          <div v-if="filesToUpload.length > 0" class="me-2">
            <div class="d-flex align-items-center">
              <span class="badge bg-info me-2">{{ filesToUpload.length }} файл(ов)</span>
              <button @click="filesToUpload = []" type="button" class="btn-close btn-close-sm"></button>
            </div>
          </div>
          <input v-model="newMessage" type="text" class="form-control me-2" placeholder="Введите сообщение...">
          <button type="submit" class="btn btn-primary">Отправить</button>
        </form>
      </div>
    </div>

    <div v-if="showEditChatModal" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Изменить чат</h5>
            <button @click="showEditChatModal = false" class="btn-close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="updateChat">
              <div class="mb-3">
                <label class="form-label">Название чата</label>
                <input v-model="editChatForm.name" class="form-control" required>
              </div>
              <div class="mb-3">
                <label class="form-label">Описание</label>
                <textarea v-model="editChatForm.description" class="form-control"></textarea>
              </div>
              <div class="mb-3 form-check">
                <input v-model="editChatForm.is_public" type="checkbox" class="form-check-input">
                <label class="form-check-label">Публичный чат</label>
              </div>
              <div class="mb-3 form-check">
                <input v-model="editChatForm.only_admin_invite" type="checkbox" class="form-check-input">
                <label class="form-check-label">Только админы могут приглашать</label>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-primary">Сохранить изменения</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showInviteModal" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Пригласить пользователя</h5>
            <button @click="showInviteModal = false" class="btn-close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="inviteUser">
              <div class="mb-3">
                <label class="form-label">Телефон пользователя</label>
                <input v-model="invitePhone" type="tel" class="form-control" placeholder="+79123456789" required>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-primary">Пригласить</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showMembersModal" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Участники чата</h5>
            <button @click="showMembersModal = false" class="btn-close"></button>
          </div>
          <div class="modal-body">
            <div class="list-group">
              <div v-for="member in chatMembers" :key="member.user_id" class="list-group-item d-flex justify-content-between align-items-center">
                <div>
                  <strong>{{ member.username || 'Пользователь' }}</strong>
                  <br>
                  <small class="text-muted">{{ member.role_name }}</small>
                </div>
                <div v-if="isCreator && member.user_id !== authStore.user.id">
                  <button @click="removeMember(member.user_id)" class="btn btn-sm btn-danger">Удалить</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const route = useRoute()
const chatStore = useChatStore()
const authStore = useAuthStore()
const newMessage = ref('')
const chat = ref({})
const messagesContainer = ref(null)
const fileInput = ref(null)
const imageInput = ref(null)
const filesToUpload = ref([])
const editingMessageId = ref(null)
const editMessageContent = ref('')
const showEditChatModal = ref(false)
const showInviteModal = ref(false)
const showMembersModal = ref(false)
const chatMembers = ref([])
const invitePhone = ref('')
let intervalId = null

const editChatForm = ref({
  name: '',
  description: '',
  is_public: false,
  only_admin_invite: false
})

const isCreator = computed(() => {
  return chat.value.creator_id === authStore.user.id
})

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const isImage = (mimeType) => {
  return mimeType.startsWith('image/')
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const handleFileUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    filesToUpload.value.push(file)
  }
  event.target.value = ''
}

const handleImageUpload = (event) => {
  const file = event.target.files[0]
  if (file) {
    filesToUpload.value.push(file)
  }
  event.target.value = ''
}

const uploadFile = async (messageId, file) => {
  const formData = new FormData()
  formData.append('message_id', messageId)
  formData.append('file', file)

  try {
    const response = await api.post('/messages/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return response.data
  } catch (error) {
    console.error('Ошибка загрузки файла:', error)
    return null
  }
}

const uploadFiles = async (messageId) => {
  if (filesToUpload.value.length === 0) return

  const uploadedFiles = []
  for (const file of filesToUpload.value) {
    const uploadedFile = await uploadFile(messageId, file)
    if (uploadedFile) {
      uploadedFiles.push(uploadedFile)
    }
  }
  
  filesToUpload.value = []
  return uploadedFiles
}

const loadChatMembers = async () => {
  try {
    const response = await api.get(`/chats/${route.params.id}/members`)
    chatMembers.value = response.data
  } catch (error) {
    console.error('Ошибка загрузки участников:', error)
  }
}

const startEditMessage = (message) => {
  editingMessageId.value = message.id
  editMessageContent.value = message.content
}

const cancelEditMessage = () => {
  editingMessageId.value = null
  editMessageContent.value = ''
}

const saveEditMessage = async (messageId) => {
  try {
    await api.put(`/messages/${messageId}`, {
      content: editMessageContent.value
    })
    
    const messageIndex = chatStore.messages.findIndex(m => m.id === messageId)
    if (messageIndex !== -1) {
      chatStore.messages[messageIndex].content = editMessageContent.value
      chatStore.messages[messageIndex].is_edited = true
    }
    
    editingMessageId.value = null
    editMessageContent.value = ''
  } catch (error) {
    console.error('Ошибка редактирования сообщения:', error)
  }
}

const deleteMessage = async (messageId) => {
  if (!confirm('Удалить сообщение?')) return
  
  try {
    await api.delete(`/messages/${messageId}`)
    chatStore.messages = chatStore.messages.filter(m => m.id !== messageId)
  } catch (error) {
    console.error('Ошибка удаления сообщения:', error)
  }
}

const updateChat = async () => {
  try {
    await api.put(`/chats/${route.params.id}`, editChatForm.value)
    
    chat.value.name = editChatForm.value.name
    chat.value.description = editChatForm.value.description
    chat.value.is_public = editChatForm.value.is_public
    chat.value.only_admin_invite = editChatForm.value.only_admin_invite
    
    showEditChatModal.value = false
    alert('Чат обновлен')
  } catch (error) {
    console.error('Ошибка обновления чата:', error)
  }
}

const inviteUser = async () => {
  try {
    await api.post(`/chats/${route.params.id}/invite`, {
      expires_in_hours: 24
    })
    
    showInviteModal.value = false
    invitePhone.value = ''
    alert('Приглашение создано')
  } catch (error) {
    console.error('Ошибка создания приглашения:', error)
  }
}

const removeMember = async (userId) => {
  if (!confirm('Удалить пользователя из чата?')) return
  
  try {
    await api.delete(`/chats/${route.params.id}/members/${userId}`)
    chatMembers.value = chatMembers.value.filter(m => m.user_id !== userId)
  } catch (error) {
    console.error('Ошибка удаления пользователя:', error)
  }
}

onMounted(() => {
  loadChat()
  loadMessages()
  
  intervalId = setInterval(() => {
    if (route.params.id) {
      loadMessages()
    }
  }, 3000)
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
    
    editChatForm.value = {
      name: chat.value.name,
      description: chat.value.description || '',
      is_public: chat.value.is_public,
      only_admin_invite: chat.value.only_admin_invite
    }
  } catch (error) {
    console.error('Ошибка загрузки чата:', error)
  }
}

const loadMessages = () => {
  chatStore.fetchMessages(route.params.id)
  scrollToBottom()
}

const sendMessage = async () => {
  if (!newMessage.value.trim() && filesToUpload.value.length === 0) return
  
  try {
    let messageData = {
      chat_id: route.params.id,
      content: newMessage.value.trim() || '(файл)'
    }
    
    const response = await api.post('/messages', messageData)
    
    const newMsg = response.data
    
    if (filesToUpload.value.length > 0) {
      const uploadedFiles = await uploadFiles(response.data.id)
      if (uploadedFiles && uploadedFiles.length > 0) {
        newMsg.files = uploadedFiles
      }
    }
    
    chatStore.messages.push(newMsg)
    newMessage.value = ''
    
    scrollToBottom()
  } catch (error) {
    console.error('Ошибка отправки сообщения:', error)
  }
}
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