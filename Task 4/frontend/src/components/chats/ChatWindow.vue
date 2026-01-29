<template>
  <div class="h-100 d-flex flex-column">
    <div class="p-3 border-bottom">
      <div class="d-flex align-items-center">
        <div class="d-flex align-items-center flex-grow-1">
          <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
               style="width: 50px; height: 50px;">
            {{ getChatInitial() }}
          </div>
          <div>
            <h5 class="mb-0">{{ chatTitle }}</h5>
            <div class="text-muted small">{{ memberCount }} участников</div>
          </div>
        </div>
        <div class="btn-group">
          <button class="btn btn-sm btn-outline-secondary" @click="toggleEmojiPicker">
            <i class="bi bi-emoji-smile"></i>
          </button>
          <button class="btn btn-sm btn-outline-secondary" @click="attachFile">
            <i class="bi bi-paperclip"></i>
          </button>
        </div>
      </div>
    </div>

    <div class="flex-grow-1 overflow-auto p-3" ref="messagesContainer">
      <div v-for="message in messages" :key="message.id" 
           class="mb-3"
           :class="{ 'text-end': message.sender_id === userId }">
        <div class="d-inline-block p-3 rounded" 
             :class="message.sender_id === userId ? 'bg-primary text-white' : 'bg-light'">
          <div v-if="message.sender_id !== userId" class="small text-muted mb-1">
            {{ message.sender?.username || message.sender?.phone }}
          </div>
          <div style="white-space: pre-wrap;">{{ message.content }}</div>
          <div class="small mt-1" :class="message.sender_id === userId ? 'text-white-50' : 'text-muted'">
            {{ formatTime(message.created_at) }}
          </div>
        </div>
      </div>
    </div>

    <div class="p-3 border-top">
      <div class="input-group">
        <textarea v-model="newMessage" 
                  class="form-control" 
                  placeholder="Сообщение..." 
                  rows="1"
                  @keydown.enter.prevent="handleEnter"
                  ref="messageInput"></textarea>
        <button class="btn btn-primary" 
                @click="sendMessage" 
                :disabled="!newMessage.trim()">
          <i class="bi bi-send"></i>
        </button>
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
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'

const route = useRoute()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const messages = ref([])
const newMessage = ref('')
const emojiPickerOpen = ref(false)
const messagesContainer = ref(null)
const messageInput = ref(null)

const emojis = ['😊', '😂', '😍', '👍', '❤️', '🔥', '🎉', '🙏']

const userId = computed(() => authStore.user?.id)
const chatId = computed(() => parseInt(route.params.id))
const currentChat = computed(() => chatsStore.currentChat)
const chatTitle = computed(() => {
  if (!currentChat.value) return ''
  if (currentChat.value.type === 'group') {
    return currentChat.value.name || 'Групповой чат'
  }
  const otherMember = currentChat.value.members?.find(m => m.id !== userId.value)
  return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
})
const memberCount = computed(() => {
  return currentChat.value?.members?.length || 0
})

onMounted(async () => {
  if (chatId.value) {
    await loadChat()
    await loadMessages()
    scrollToBottom()
  }
})

const loadChat = async () => {
  await chatsStore.fetchChat(chatId.value)
}

const loadMessages = async () => {
  const result = await chatsStore.fetchMessages(chatId.value)
  if (result.success) {
    messages.value = chatsStore.messages
  }
}

const sendMessage = async () => {
  if (!newMessage.value.trim()) return
  
  const result = await chatsStore.sendMessageWithFiles(
    chatId.value,
    newMessage.value,
    []
  )
  
  if (result.success) {
    newMessage.value = ''
    messages.value = chatsStore.messages
    scrollToBottom()
  }
}

const handleEnter = (e) => {
  if (e.shiftKey) {
    newMessage.value += '\n'
    adjustTextareaHeight()
  } else if (!e.shiftKey && newMessage.value.trim()) {
    e.preventDefault()
    sendMessage()
  }
}

const adjustTextareaHeight = () => {
  if (messageInput.value) {
    messageInput.value.style.height = 'auto'
    messageInput.value.style.height = Math.min(messageInput.value.scrollHeight, 100) + 'px'
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const toggleEmojiPicker = () => {
  emojiPickerOpen.value = !emojiPickerOpen.value
}

const insertEmoji = (emoji) => {
  if (messageInput.value) {
    const cursorPos = messageInput.value.selectionStart
    const textBefore = newMessage.value.substring(0, cursorPos)
    const textAfter = newMessage.value.substring(cursorPos)
    newMessage.value = textBefore + emoji + textAfter
    
    nextTick(() => {
      messageInput.value.focus()
      adjustTextareaHeight()
    })
  }
  emojiPickerOpen.value = false
}

const attachFile = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.onchange = async (e) => {
    const files = Array.from(e.target.files)
    if (files.length > 0) {
      const result = await chatsStore.sendMessageWithFiles(
        chatId.value,
        newMessage.value,
        files
      )
      if (result.success) {
        newMessage.value = ''
        messages.value = chatsStore.messages
        scrollToBottom()
      }
    }
  }
  input.click()
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
</script>