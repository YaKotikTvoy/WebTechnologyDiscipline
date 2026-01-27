<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-8 offset-md-2">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <div>
              <h5 class="mb-0">{{ chatTitle }}</h5>
              <small class="text-muted">
                {{ chatType }} • {{ memberCount }} members
              </small>
            </div>
            <div v-if="currentChat?.type === 'group'">
              <button
                @click="showAddMember = true"
                class="btn btn-sm btn-outline-primary"
              >
                Add Member
              </button>
            </div>
          </div>

          <div class="card-body chat-body" ref="chatBody">
            <div v-if="loading" class="text-center py-3">
              <div class="spinner-border" role="status">
                <span class="visually-hidden">Loading...</span>
              </div>
            </div>

            <div v-else-if="messages.length === 0" class="text-center py-5">
              <h5>No messages yet</h5>
              <p class="text-muted">Start the conversation!</p>
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
                    Message deleted
                  </div>
                  <div v-else>
                    <div v-if="message.sender_id !== userId" class="small text-muted mb-1">
                      {{ message.sender.phone }}
                    </div>
                    <div class="message-text">
                      {{ message.content }}
                    </div>
                    <div
                      v-for="file in message.files"
                      :key="file.id"
                      class="mt-2"
                    >
                      <a
                        :href="`http://localhost:8080/uploads/${file.filepath}`"
                        target="_blank"
                        class="d-block p-2 bg-white rounded border"
                      >
                        <div class="d-flex align-items-center">
                          <div class="me-2">
                            <i class="bi bi-file-earmark"></i>
                          </div>
                          <div class="flex-grow-1">
                            <div class="fw-bold">{{ file.filename }}</div>
                            <div class="text-muted small">
                              {{ formatFileSize(file.filesize) }}
                            </div>
                          </div>
                        </div>
                      </a>
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
                    Delete
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
                  placeholder="Type your message..."
                  :disabled="sending"
                />
              </div>
              <div class="btn-group">
                <input
                  type="file"
                  ref="fileInput"
                  style="display: none"
                  @change="handleFileSelect"
                />
                <button
                  type="button"
                  class="btn btn-outline-secondary"
                  @click="$refs.fileInput.click()"
                  :disabled="sending"
                >
                  <i class="bi bi-paperclip"></i>
                </button>
                <button
                  type="submit"
                  class="btn btn-primary"
                  :disabled="!newMessage.trim() && !selectedFile || sending"
                >
                  <span v-if="sending">
                    <span class="spinner-border spinner-border-sm" role="status"></span>
                  </span>
                  <span v-else>Send</span>
                </button>
              </div>
            </form>
            <div v-if="selectedFile" class="mt-2">
              <div class="badge bg-info">
                {{ selectedFile.name }}
                <button
                  type="button"
                  class="btn-close btn-close-white ms-1"
                  @click="selectedFile = null"
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
            <h5 class="modal-title">Add Member</h5>
            <button
              type="button"
              class="btn-close"
              @click="showAddMember = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Friend Phone</label>
              <input
                v-model="newMemberPhone"
                type="text"
                class="form-control"
                placeholder="Enter friend's phone number"
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
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="addMember"
              :disabled="addingMember"
            >
              {{ addingMember ? 'Adding...' : 'Add Member' }}
            </button>
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
import { api } from '@/services/api'

const route = useRoute()
const chatsStore = useChatsStore()
const authStore = useAuthStore()

const chatId = ref(parseInt(route.params.id))
const messages = ref([])
const newMessage = ref('')
const selectedFile = ref(null)
const sending = ref(false)
const error = ref('')
const loading = ref(false)
const showAddMember = ref(false)
const newMemberPhone = ref('')
const addingMember = ref(false)
const addMemberError = ref('')
const chatBody = ref(null)

const userId = computed(() => authStore.user?.id)
const currentChat = computed(() => chatsStore.currentChat)
const chatTitle = computed(() => {
  if (!currentChat.value) return ''
  if (currentChat.value.type === 'group') {
    return currentChat.value.name || 'Group Chat'
  }
  const otherMember = currentChat.value.members?.find(m => m.id !== userId.value)
  return otherMember ? otherMember.phone : 'Private Chat'
})
const chatType = computed(() => {
  return currentChat.value?.type === 'group' ? 'Group' : 'Private'
})
const memberCount = computed(() => {
  return currentChat.value?.members?.length || 0
})

onMounted(async () => {
  await loadChat()
  await loadMessages()
  scrollToBottom()
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
  if (!newMessage.value.trim() && !selectedFile.value) return

  sending.value = true
  error.value = ''

  const result = await chatsStore.sendMessage(
    chatId.value,
    newMessage.value,
    selectedFile.value
  )

  if (result.success) {
    newMessage.value = ''
    selectedFile.value = null
    messages.value = chatsStore.messages
    scrollToBottom()
  } else {
    error.value = result.error
  }

  sending.value = false
}

const deleteMessage = async (messageId) => {
  if (confirm('Are you sure you want to delete this message?')) {
    await chatsStore.deleteMessage(messageId)
    const index = messages.value.findIndex(m => m.id === messageId)
    if (index !== -1) {
      messages.value[index].is_deleted = true
    }
  }
}

const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    if (file.size > 10 * 1024 * 1024) {
      error.value = 'File size must be less than 10MB'
      return
    }
    selectedFile.value = file
  }
}

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], {
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
    await loadChat()
  } catch (error) {
    addMemberError.value = error.response?.data || 'Failed to add member'
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
</style>