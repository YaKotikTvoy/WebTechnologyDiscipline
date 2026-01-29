<template>
  <div class="h-100 d-flex flex-column">
    <div class="p-3 border-bottom bg-white">
      <div class="d-flex align-items-center">
        <button class="btn btn-link text-dark d-md-none me-2" @click="goBack">
          <i class="bi bi-arrow-left"></i>
        </button>
        <div class="d-flex align-items-center flex-grow-1">
          <div class="rounded-circle d-flex align-items-center justify-content-center me-3" 
               :class="getChatColor()"
               style="width: 50px; height: 50px; font-size: 1.2rem;">
            {{ getChatInitial() }}
          </div>
          <div>
            <h5 class="mb-0">{{ chatTitle }}</h5>
            <div class="text-muted small">{{ memberCount }} участников</div>
          </div>
        </div>
      </div>
    </div>

    <div ref="messagesContainer" class="flex-grow-1 overflow-auto p-3 bg-light">
      <div v-if="loading" class="text-center py-4">
        <div class="spinner-border spinner-border-sm" role="status">
          <span class="visually-hidden">Загрузка...</span>
        </div>
        <div class="mt-2 text-muted small">
          Загрузка сообщений... {{ chatsStore.messages.length }} сообщений в памяти
        </div>
      </div>

      <div v-else-if="messages.length === 0" class="text-center py-5">
        <i class="bi bi-chat-dots display-1 text-muted mb-3"></i>
        <p class="text-muted">Нет сообщений</p>
        <p class="text-muted small">
          ChatId: {{ chatId }}, В кеше: {{ chatsStore.messagesCache.has(chatId) ? 'да' : 'нет' }}
        </p>
      </div>
      
      <div v-else>
        <div v-for="message in messages" :key="message.id" 
            class="mb-3"
            :class="{ 'text-end': message.sender_id === userId }">
          <div class="d-inline-block p-3 rounded shadow-sm" 
               :class="message.sender_id === userId ? 'bg-primary text-white' : 'bg-white'"
               style="max-width: 70%;">
            <div v-if="message.sender_id !== userId" class="small mb-1" 
                 :class="message.sender_id === userId ? 'text-white-50' : 'text-muted'">
              {{ message.sender?.username || message.sender?.phone }}
            </div>
            <div v-if="message.is_deleted" class="text-muted">
              <i class="bi bi-trash"></i> Сообщение удалено
            </div>
            <div v-else style="white-space: pre-wrap;">{{ message.content }}</div>
            
            <div v-if="message.files && message.files.length > 0" class="mt-2">
              <div v-for="file in message.files" :key="file.id" class="mb-2">
                <a :href="`http://localhost:8080/uploads/${file.filepath}`" 
                   target="_blank" 
                   class="text-decoration-none">
                  <div v-if="isImage(file)" class="file-preview">
                    <img :src="`http://localhost:8080/uploads/${file.filepath}`" 
                         :alt="file.filename"
                         class="img-thumbnail"
                         style="max-width: 200px; max-height: 200px;">
                    <div class="small text-muted mt-1">{{ file.filename }}</div>
                  </div>
                  <div v-else class="d-flex align-items-center p-2 bg-white rounded border">
                    <i class="bi bi-file-earmark me-2 fs-4"></i>
                    <div>
                      <div class="fw-bold">{{ file.filename }}</div>
                      <div class="small text-muted">{{ formatFileSize(file.filesize) }}</div>
                    </div>
                  </div>
                </a>
              </div>
            </div>
            
            <div class="small mt-1 d-flex align-items-center gap-1" 
                :class="message.sender_id === userId ? 'text-white-50' : 'text-muted'">
              <span>{{ formatTime(message.created_at) }}</span>
              <span v-if="message.sender_id === userId" class="ms-1">
                <i v-if="message.readers && message.readers.length > 0" 
                  class="bi bi-check2-all text-info"></i>
                <i v-else 
                  class="bi bi-check2"></i>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="p-3 border-top bg-white">
      <form @submit.prevent="sendMessage" class="d-flex align-items-end">
        <button type="button" 
                class="btn btn-outline-secondary me-2" 
                @click="attachFile">
          <i class="bi bi-paperclip fs-5"></i>
        </button>
        
        <div class="flex-grow-1 me-2">
          <textarea v-model="newMessage" 
                    class="form-control" 
                    placeholder="Сообщение..." 
                    rows="1"
                    @keydown="handleEnter"
                    @input="autoResize"
                    ref="messageInput"
                    style="resize: none; max-height: 120px;"></textarea>
        </div>
        
        <button type="button" 
                class="btn btn-outline-secondary me-2" 
                @click="toggleEmojiPicker">
          <i class="bi bi-emoji-smile-fill fs-5"></i>
        </button>
        
        <button type="submit" 
                class="btn btn-primary" 
                :disabled="!newMessage.trim() && selectedFiles.length === 0">
          <i class="bi bi-send-fill fs-5"></i>
        </button>
      </form>
      
      <div v-if="selectedFiles.length > 0" class="mt-2">
        <div v-for="(file, index) in selectedFiles" :key="index" class="badge bg-info me-2 mb-1">
          {{ file.name }}
          <button type="button" class="btn-close btn-close-white ms-1" @click="removeFile(index)"></button>
        </div>
      </div>
      
      <div v-if="emojiPickerOpen" class="mt-2">
        <div class="d-flex flex-wrap gap-1">
          <button v-for="emoji in emojis" :key="emoji" 
                  class="btn btn-sm btn-outline-secondary"
                  @click="insertEmoji(emoji)">
            {{ emoji }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch,watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const newMessage = ref('')
const emojiPickerOpen = ref(false)
const messagesContainer = ref(null)
const messageInput = ref(null)
const loading = ref(false)
const selectedFiles = ref([])

const emojis = ['😊', '😂', '❤️', '👍', '🔥', '🎉', '👏', '🙏']

const userId = computed(() => authStore.user?.id)
const chatId = computed(() => parseInt(route.params.id))

const messages = computed(() => {
  return chatsStore.messages
})
const currentChat = computed(() => {
  return chatsStore.chats.find(chat => chat.id === chatId.value) || chatsStore.currentChat
})
const chatTitle = computed(() => {
  if (!currentChat.value) return 'Загрузка...'
  if (currentChat.value.type === 'group') {
    return currentChat.value.name || 'Групповой чат'
  }
  const otherMember = currentChat.value.members?.find(m => m.id !== userId.value)
  return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
})
const memberCount = computed(() => {
  return currentChat.value?.members?.length || 0
})

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const loadChatData = async () => {
  if (!chatId.value) return
  
  loading.value = true
  try {
    console.log('Загружаем данные чата', chatId.value)
    
    await chatsStore.fetchChat(chatId.value)
    
    const result = await chatsStore.fetchMessages(chatId.value)
    console.log('Результат загрузки сообщений:', result.success)
    
    if (result.success) {
      console.log('Сообщений загружено:', chatsStore.messages.length)
      scrollToBottom()
    } else {
      console.error('Ошибка загрузки сообщений:', result.error)
    }
    
    chatsStore.setActiveChat(chatId.value)
  } catch (error) {
    console.error('Ошибка загрузки данных чата:', error)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  if (chatId.value) {
    await loadChatData()
  }
})

watch(
  () => route.params.id,
  async (newId) => {
    if (newId) {
      messages.value = []
      newMessage.value = ''
      selectedFiles.value = []
      emojiPickerOpen.value = false
      await loadChatData()
    }
  }
)

watchEffect(() => {
  const msgCount = chatsStore.messages.length
  console.log('Сообщений в чате:', msgCount)
})

const sendMessage = async () => {
  if (!newMessage.value.trim() && selectedFiles.value.length === 0) return
  
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
  }
}

const handleEnter = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}




const toggleEmojiPicker = () => {
  emojiPickerOpen.value = !emojiPickerOpen.value
}

const insertEmoji = (emoji) => {
  if (messageInput.value) {
    newMessage.value += emoji
    nextTick(() => {
      messageInput.value.focus()
    })
  }
  emojiPickerOpen.value = false
}

const attachFile = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.onchange = (e) => {
    const files = Array.from(e.target.files)
    files.forEach(file => {
      if (file.size <= 10 * 1024 * 1024) {
        selectedFiles.value.push(file)
      }
    })
  }
  input.click()
}

const removeFile = (index) => {
  selectedFiles.value.splice(index, 1)
}

const isImage = (file) => {
  const imageTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
  return imageTypes.includes(file.mime_type) || 
         file.filename.match(/\.(jpg|jpeg|png|gif|webp)$/i)
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const goBack = () => {
  router.push('/')
}

const getChatColor = () => {
  const colors = ['bg-primary', 'bg-success', 'bg-warning', 'bg-danger', 'bg-info', 'bg-secondary']
  const index = chatId.value % colors.length
  return colors[index]
}

const getChatInitial = () => {
  if (!currentChat.value) return '?'
  return chatTitle.value.charAt(0).toUpperCase()
}

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const autoResize = () => {
  nextTick(() => {
    if (messageInput.value) {
      messageInput.value.style.height = 'auto'
      const newHeight = Math.min(messageInput.value.scrollHeight, 120)
      messageInput.value.style.height = newHeight + 'px'
    }
  })
}
watch(() => chatsStore.messages, (newMessages, oldMessages) => {
  if (newMessages.length > oldMessages.length) {
    scrollToBottom()
  }
}, { deep: true })

onMounted(() => {
  scrollToBottom()
})

watch(() => wsStore.notifications, (notifications) => {
  const chatNotifications = notifications.filter(n => 
    n.type === 'new_message' && n.data.chatId === chatId.value
  )
  if (chatNotifications.length > 0) {
    loadChatData()
  }
}, { deep: true })
</script>

<style>
.h-100 {
  height: 100vh !important;
}

.flex-grow-1 {
  flex: 1 1 0% !important;
}

.overflow-auto {
  overflow-y: auto !important;
}
</style>