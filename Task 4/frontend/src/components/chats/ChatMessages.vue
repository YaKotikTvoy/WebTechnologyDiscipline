<template>
  <div ref="container" 
       class="flex-grow-1 overflow-auto p-3 bg-light"
       @scroll="throttledSaveScroll">
    
    <div v-if="loading" class="text-center py-4">
      <div class="spinner-border spinner-border-sm" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
      <div v-if="loadingMessage" class="mt-2 text-muted small">
        {{ loadingMessage }}
      </div>
    </div>
    
    <div v-else-if="messages.length === 0" class="text-center py-5">
      <i class="bi bi-chat-dots display-1 text-muted mb-3"></i>
      <p class="text-muted">Нет сообщений</p>
    </div>
    
    <div v-else>
      <div v-for="message in messages" 
           :key="message.id"
           :data-message-id="message.id"
           class="message-item">
        
        <div v-if="message.type && message.type.startsWith('system_')" 
             class="text-center my-2">
          <div class="badge bg-secondary text-white py-2 px-3 fw-normal opacity-75">
            <i v-if="message.type === 'system_user_added'" class="bi bi-person-plus me-1"></i>
            <i v-else-if="message.type === 'system_user_removed'" class="bi bi-person-dash me-1"></i>
            <i v-else-if="message.type === 'system_chat_created'" class="bi bi-chat-left-dots me-1"></i>
            <i v-else class="bi bi-info-circle me-1"></i>
            {{ message.content }}
          </div>
          <div class="text-muted small mt-1">
            {{ formatTime(message.created_at) }}
          </div>
        </div>
        
        <ChatMessage 
          v-else
          :message="message"
          :userId="userId"
          :chat-type="chatType"
          :is-chat-admin="isChatAdmin"
          @edit="handleEditMessage"
          @delete="handleDeleteMessage"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { useChatsStore } from '@/stores/chats'
import { throttle } from 'lodash-es'
import ChatMessage from './ChatMessage.vue'

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  },
  userId: {
    type: Number,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  loadingMessage: {
    type: String,
    default: ''
  },
  chatType: {
    type: String,
    default: 'private'
  },
  isChatAdmin: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['edit-message', 'delete-message'])

const container = ref(null)
const chatsStore = useChatsStore()

const formatTime = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const handleEditMessage = (data) => {
  emit('edit-message', data)
}

const handleDeleteMessage = (messageId) => {
  emit('delete-message', messageId)
}

const restoreScrollPosition = () => {
  if (!container.value || !chatsStore.activeChatId) return
  
  const savedPosition = chatsStore.getScrollPosition(chatsStore.activeChatId)
  
  if (savedPosition > 0) {
    container.value.scrollTop = savedPosition
  } else {
    container.value.scrollTop = container.value.scrollHeight
  }
}

const saveScrollPosition = () => {
  if (!container.value || !chatsStore.activeChatId) return
  
  const position = container.value.scrollTop
  chatsStore.saveScrollPosition(chatsStore.activeChatId, position)
}

const throttledSaveScroll = throttle(saveScrollPosition, 500)

onMounted(() => {
  nextTick(() => {
    restoreScrollPosition()
  })
})

watch(() => props.messages, () => {
  nextTick(() => {
    restoreScrollPosition()
  })
}, { deep: true })

watch(() => chatsStore.activeChatId, () => {
  nextTick(() => {
    restoreScrollPosition()
  })
})

defineExpose({
  scrollToBottom: () => {
    if (container.value) {
      container.value.scrollTop = container.value.scrollHeight
    }
  },
  getContainer: () => container.value
})
</script>