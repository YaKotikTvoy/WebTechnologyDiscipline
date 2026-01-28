<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-8 offset-md-2">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <div>
              <h5 class="mb-0">{{ chatTitle }}</h5>
              <small class="text-muted">
                {{ chatType }} • {{ memberCount }} участников
              </small>
            </div>
            <div>
              <div class="btn-group">
                <button
                  @click="showAddMember = true"
                  class="btn btn-sm btn-outline-primary"
                >
                  Добавить контакт
                </button>
                <button v-if="isAdmin" class="btn btn-sm btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
                  Настройки
                </button>
                <ul v-if="isAdmin" class="dropdown-menu dropdown-menu-end">
                  <li>
                    <button class="dropdown-item" @click="toggleChatVisibility">
                      <svg v-if="currentChat?.is_searchable" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-eye-fill me-2" viewBox="0 0 16 16">
                        <path d="M10.5 8a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0z"/>
                        <path d="M0 8s3-5.5 8-5.5S16 8 16 8s-3 5.5-8 5.5S0 8 0 8zm8 3.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z"/>
                      </svg>
                      <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-eye-slash-fill me-2" viewBox="0 0 16 16">
                        <path d="m10.79 12.912-1.614-1.615a3.5 3.5 0 0 1-4.474-4.474l-2.06-2.06C.938 6.278 0 8 0 8s3 5.5 8 5.5a7.029 7.029 0 0 0 2.79-.588zM5.21 3.088A7.028 7.028 0 0 1 8 2.5c5 0 8 5.5 8 5.5s-.939 1.721-2.641 3.238l-2.062-2.062a3.5 3.5 0 0 0-4.474-4.474L5.21 3.089z"/>
                        <path d="M5.525 7.646a2.5 2.5 0 0 0 2.829 2.829l-2.83-2.829zm4.95.708-2.829-2.83a2.5 2.5 0 0 1 2.829 2.829zm3.171 6-12-12 .708-.708 12 12-.708.708z"/>
                      </svg>
                      {{ currentChat?.is_searchable ? 'Скрыть из поиска' : 'Показать в поиске' }}
                    </button>
                  </li>
                  <li v-if="currentChat?.type === 'group'">
                    <button class="dropdown-item" @click="showJoinRequests = true">
                      Заявки на вступление
                      <span v-if="pendingJoinRequests > 0" class="badge bg-danger ms-1">
                        {{ pendingJoinRequests }}
                      </span>
                    </button>
                  </li>
                </ul>
              </div>
            </div>
          </div>

          <div class="card-body chat-body" ref="chatBody">
            <div v-if="loading" class="text-center py-3">
              <div class="spinner-border" role="status">
                <span class="visually-hidden">Загрузка...</span>
              </div>
            </div>

            <div v-else-if="messages.length === 0" class="text-center py-5">
              <h5>Нет сообщений</h5>
              <p class="text-muted">Начать общение</p>
            </div>

            <div v-else>
              <div
                v-for="message in messages"
                :key="message.id"
                class="message mb-3"
                :class="{ 'text-end': message.sender_id === userId }"
              >
                <div
                  class="message-content d-inline-block p-2 rounded"
                  :class="{
                    'bg-primary text-white': message.sender_id === userId,
                    'bg-light': message.sender_id !== userId
                  }"
                >
                  <div v-if="message.is_deleted" class="text-muted">
                    Удалено
                  </div>
                  <div v-else>
                    <div v-if="message.sender_id !== userId" class="small text-muted mb-1">
                        {{ message.sender.username || message.sender.phone }}
                    </div>
                    <div class="message-text">
                      {{ message.content }}
                    </div>
                      <div
                        v-for="file in message.files"
                        :key="file.id"
                        class="mt-2"
                      >
                        <div v-if="isImage(file)">
                          <a
                            :href="`http://localhost:8080/uploads/${file.filepath}`"
                            target="_blank"
                            class="d-block"
                          >
                            <img
                              :src="`http://localhost:8080/uploads/${file.filepath}`"
                              :alt="file.filename"
                              class="img-thumbnail"
                              style="max-width: 200px; max-height: 200px;"
                              @load="imageLoaded"
                            />
                          </a>
                          <div class="small text-muted mt-1">
                            {{ file.filename }} ({{ formatFileSize(file.filesize) }})
                          </div>
                        </div>
                        <div v-else>
                          <a
                            :href="`http://localhost:8080/uploads/${file.filepath}`"
                            target="_blank"
                            class="d-block p-2 bg-white rounded border text-decoration-none"
                          >
                            <div class="d-flex align-items-center">
                              <div class="me-2">
                                <i class="bi bi-file-earmark"></i>
                              </div>
                              <div class="flex-grow-1">
                                <div class="fw-bold text-dark">{{ file.filename }}</div>
                                <div class="text-muted small">
                                  {{ formatFileSize(file.filesize) }}
                                </div>
                              </div>
                            </div>
                          </a>
                        </div>
                      </div>
                    <div class="small text-muted mt-1">
                      {{ formatTime(message.created_at) }}
                    </div>
                  </div>
                </div>
                <div
                  v-if="message.sender_id === userId && !message.is_deleted"
                  class="mt-1"
                >
                  <button
                    @click="deleteMessage(message.id)"
                    class="btn btn-sm btn-link text-danger p-0"
                  >
                    Удалить
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div class="card-footer">
            <form @submit.prevent="sendMessage" class="d-flex">
              <div class="flex-grow-1 me-2">
                <input
                  v-model="newMessage"
                  type="text"
                  class="form-control"
                  placeholder="Введи сообщение..."
                  :disabled="sending"
                />
              </div>
              <div class="btn-group">
                <input
                  type="file"
                  ref="fileInput"
                  style="display: none"
                  @change="handleFileSelect"
                  multiple
                  accept="image/*,.pdf,.doc,.docx,.txt"
                />
                <button
                  type="button"
                  class="btn btn-outline-secondary"
                  @click="$refs.fileInput.click()"
                  :disabled="sending"
                  title="Прикрепить файл"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-paperclip" viewBox="0 0 16 16">
                    <path d="M4.5 3a2.5 2.5 0 0 1 5 0v9a1.5 1.5 0 0 1-3 0V5a.5.5 0 0 1 1 0v7a.5.5 0 0 0 1 0V3a1.5 1.5 0 1 0-3 0v9a2.5 2.5 0 0 0 5 0V5a.5.5 0 0 1 1 0v7a3.5 3.5 0 1 1-7 0V3z"/>
                  </svg>
                </button>
                <button
                  type="submit"
                  class="btn btn-primary"
                  :disabled="(!newMessage.trim() && selectedFiles.length === 0) || sending"
                >
                  <span v-if="sending">
                    <span class="spinner-border spinner-border-sm" role="status"></span>
                  </span>
                  <span v-else>Отправить</span>
                </button>
              </div>
            </form>
            <div v-if="selectedFiles.length > 0" class="mt-2">
              <div
                v-for="(file, index) in selectedFiles"
                :key="index"
                class="badge bg-info me-2 mb-1"
              >
                {{ file.name }}
                <button
                  type="button"
                  class="btn-close btn-close-white ms-1"
                  @click="removeFile(index)"
                ></button>
              </div>
            </div>
            <div v-if="error" class="alert alert-danger mt-2 mb-0">
              {{ error }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showAddMember" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Добавить контакт</h5>
            <button
              type="button"
              class="btn-close"
              @click="showAddMember = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Номер</label>
              <input
                v-model="newMemberPhone"
                type="text"
                class="form-control"
                placeholder="Введите номер"
              />
            </div>
            <div v-if="addMemberError" class="alert alert-danger">
              {{ addMemberError }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="showAddMember = false"
            >
              Закрыть
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="addMember"
              :disabled="addingMember"
            >
              {{ addingMember ? 'Добавление...' : 'Добавить' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showJoinRequests" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Заявки на вступление в чат</h5>
            <button type="button" class="btn-close" @click="showJoinRequests = false"></button>
          </div>
          <div class="modal-body">
            <div v-if="joinRequests.length === 0" class="text-center py-3">
              Нет заявок на вступление
            </div>
            <div v-else class="list-group">
              <div v-for="request in joinRequests" :key="request.id" class="list-group-item">
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ request.user?.username || request.user?.phone }}</strong>
                    <div class="text-muted small">
                      {{ formatDate(request.created_at) }}
                    </div>
                  </div>
                  <div>
                    <button @click="respondToJoinRequest(request.id, 'accepted')" 
                            class="btn btn-sm btn-success me-2"
                            :disabled="responding">
                      Принять
                    </button>
                    <button @click="respondToJoinRequest(request.id, 'rejected')" 
                            class="btn btn-sm btn-danger"
                            :disabled="responding">
                      Отклонить
                    </button>
                  </div>
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
import { onMounted, ref, computed, watch, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { api } from '@/services/api'

const route = useRoute()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const chatId = ref(parseInt(route.params.id))
const messages = ref([])
const newMessage = ref('')
const selectedFiles = ref([])
const sending = ref(false)
const error = ref('')
const loading = ref(false)
const showAddMember = ref(false)
const newMemberPhone = ref('')
const addingMember = ref(false)
const addMemberError = ref('')
const chatBody = ref(null)
const showJoinRequests = ref(false)
const joinRequests = ref([])
const pendingJoinRequests = ref(0)
const isAdmin = ref(false)
const responding = ref(false)

const userId = computed(() => authStore.user?.id)
const currentChat = computed(() => chatsStore.currentChat)
const chatTitle = computed(() => {
  if (!currentChat.value) return ''
  if (currentChat.value.type === 'group') {
    return currentChat.value.name || 'Групповой чат'
  }
  const otherMember = currentChat.value.members?.find(m => m.id !== userId.value)
  return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
})
const chatType = computed(() => {
  return currentChat.value?.type === 'group' ? 'Групповой' : 'Приватный'
})
const memberCount = computed(() => {
  return currentChat.value?.members?.length || 0
})

const checkIfAdmin = async () => {
  if (!currentChat.value || !userId.value) return
  const members = currentChat.value.members || []
  const member = members.find(m => m.id === userId.value)
  isAdmin.value = member?.is_admin || false
}

const loadJoinRequests = async () => {
  if (!isAdmin.value || !chatId.value) return
  try {
    const response = await api.get(`/chats/${chatId.value}/join-requests`)
    joinRequests.value = response.data
    pendingJoinRequests.value = joinRequests.value.filter(r => r.status === 'pending').length
  } catch (error) {
    console.error('Ошибка загрузки заявок:', error)
  }
}

const respondToJoinRequest = async (requestId, status) => {
  responding.value = true
  try {
    await api.put(`/chat-join-requests/${requestId}`, { status })
    await loadJoinRequests()
  } catch (error) {
    console.error('Ошибка обработки заявки:', error)
  } finally {
    responding.value = false
  }
}

const toggleChatVisibility = async () => {
  if (!currentChat.value) return
  try {
    await api.put(`/chats/${chatId.value}/visibility`, {
      is_searchable: !currentChat.value.is_searchable
    })
    await loadChat()
    alert(currentChat.value.is_searchable ? 'Чат скрыт из поиска' : 'Чат теперь виден в поиске')
  } catch (error) {
    console.error('Ошибка обновления видимости:', error)
    alert('Не удалось изменить видимость чата')
  }
}

onMounted(async () => {
  await loadChat()
  await loadMessages()
  scrollToBottom()
  
  if (chatId.value) {
    try {
      await api.post(`/chats/${chatId.value}/read`)
    } catch (error) {
      console.error('Не удалось отметить сообщения как прочитанные:', error)
    }
  }
})

watch(() => currentChat.value, () => {
  checkIfAdmin()
  if (isAdmin.value) {
    loadJoinRequests()
  }
})

const loadChat = async () => {
  loading.value = true
  await chatsStore.fetchChat(chatId.value)
  loading.value = false
}

const loadMessages = async () => {
  const result = await chatsStore.fetchMessages(chatId.value)
  if (result.success) {
    messages.value = chatsStore.messages
  }
}

const sendMessage = async () => {
  if (!newMessage.value.trim() && selectedFiles.value.length === 0) return

  sending.value = true
  error.value = ''

  const formData = new FormData()
  formData.append('content', newMessage.value)
  
  selectedFiles.value.forEach((file, index) => {
    formData.append(`files[${index}]`, file)
  })

  const result = await chatsStore.sendMessageWithFiles(
    chatId.value,
    newMessage.value,
    selectedFiles.value
  )

  if (result.success) {
    newMessage.value = ''
    selectedFiles.value = []
    messages.value = chatsStore.messages
    scrollToBottom()
  } else {
    error.value = result.error
  }

  sending.value = false
}

const handleFileSelect = (event) => {
  const files = Array.from(event.target.files)
  
  files.forEach(file => {
    if (file.size > 10 * 1024 * 1024) {
      error.value = `Файл "${file.name}" превышает 10 МБ`
      return
    }
    
    selectedFiles.value.push(file)
  })
  
  event.target.value = null
}

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1)
}

const deleteMessage = async (messageId) => {
  if (confirm('Удалить данное сообщение?')) {
    await chatsStore.deleteMessage(messageId)
    const index = messages.value.findIndex(m => m.id === messageId)
    if (index !== -1) {
      messages.value[index].is_deleted = true
    }
  }
}

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const scrollToBottom = () => {
  nextTick(() => {
    if (chatBody.value) {
      chatBody.value.scrollTop = chatBody.value.scrollHeight
    }
  })
}

const addMember = async () => {
  addingMember.value = true
  addMemberError.value = ''

  try {
    await api.post(`/chats/${chatId.value}/members`, {
      phone: newMemberPhone.value
    })
    
    showAddMember.value = false
    newMemberPhone.value = ''
    alert('Приглашение отправлено')
  } catch (error) {
    addMemberError.value = error.response?.data || 'Не удалось отправить приглашение'
  }

  addingMember.value = false
}

watch(
  () => chatsStore.messages,
  (newMessages) => {
    if (chatId.value === chatsStore.currentChat?.id) {
      messages.value = newMessages
      scrollToBottom()
    }
  }
)

const isImage = (file) => {
  const imageTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
  return imageTypes.includes(file.mime_type) || file.filename.match(/\.(jpg|jpeg|png|gif|webp)$/i)
}
</script>

<style scoped>
.chat-body {
  height: 500px;
  overflow-y: auto;
  padding: 1rem;
}

.message-content {
  max-width: 70%;
}

.message-text {
  word-wrap: break-word;
}

.bi-paperclip {
  vertical-align: -0.125em;
}
</style>