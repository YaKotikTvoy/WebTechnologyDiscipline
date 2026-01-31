<template>
  <div ref="container" 
       class="flex-grow-1 overflow-auto p-3 bg-light"
       @scroll="handleScroll">
    
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
      <div v-for="(message, index) in messages" 
           :key="message.id"
           :data-message-id="message.id"
           class="message-item"
           :class="{ 'read': isMessageRead(message) }"
           @mouseenter="handleMessageHover(message.id)">
        
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
import { ref, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { useChatsStore } from '@/stores/chats'
import { debounce } from 'lodash-es'
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

const emit = defineEmits(['edit-message', 'delete-message', 'scrolled', 'message-hover'])

const container = ref(null)
const chatsStore = useChatsStore()
const lastScrollTop = ref(0)
const currentChatId = ref(null)

const formatTime = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const isMessageRead = (message) => {
  if (!message.readers || !Array.isArray(message.readers)) return false
  return message.readers.some(reader => reader.id === props.userId)
}

const handleEditMessage = (data) => {
  emit('edit-message', data)
}

const handleDeleteMessage = (messageId) => {
  emit('delete-message', messageId)
}

const handleMessageHover = (messageId) => {
  emit('message-hover', messageId)
}

const handleScroll = debounce(() => {
  if (container.value) {
    const scrollTop = container.value.scrollTop
    lastScrollTop.value = scrollTop
    
    emit('scrolled', scrollTop)
  }
}, 100)

const restoreScrollPosition = () => {
  if (!container.value || !chatsStore.activeChatId) return
  
  const savedPosition = chatsStore.getScrollPosition(chatsStore.activeChatId)
  
  if (savedPosition > 0) {
    nextTick(() => {
      setTimeout(() => {
        if (container.value) {
          container.value.scrollTop = savedPosition
        }
      }, 100)
    })
  } else {
    nextTick(() => {
      setTimeout(() => {
        if (container.value) {
          container.value.scrollTop = container.value.scrollHeight
        }
      }, 100)
    })
  }
}

onMounted(() => {
  setTimeout(() => {
    restoreScrollPosition()
  }, 200)
})

onUnmounted(() => {
  if (container.value && chatsStore.activeChatId) {
    const scrollTop = container.value.scrollTop
    chatsStore.saveScrollPosition(chatsStore.activeChatId, scrollTop)
  }
})

watch(() => chatsStore.activeChatId, (newChatId, oldChatId) => {
  if (oldChatId && container.value) {
    const scrollTop = container.value.scrollTop
    chatsStore.saveScrollPosition(oldChatId, scrollTop)
  }
  
  currentChatId.value = newChatId
  
  if (newChatId) {
    nextTick(() => {
      setTimeout(() => {
        restoreScrollPosition()
      }, 100)
    })
  }
}, { immediate: true })

watch(() => props.messages, (newMessages) => {
  if (newMessages.length > 0) {
    nextTick(() => {
      setTimeout(() => {
        if (container.value && chatsStore.activeChatId === currentChatId.value) {
          const savedPosition = chatsStore.getScrollPosition(chatsStore.activeChatId)
          if (savedPosition > 0 && Math.abs(container.value.scrollTop - savedPosition) > 50) {
            container.value.scrollTop = savedPosition
          }
        }
      }, 150)
    })
  }
}, { deep: true })

defineExpose({
  scrollToBottom: () => {
    if (container.value) {
      container.value.scrollTop = container.value.scrollHeight
    }
  },
  getContainer: () => container.value
})
</script>

<style scoped>
.message-item.read {
  opacity: 0.9;
}

.message-item:not(.read) {
  background-color: rgba(13, 110, 253, 0.05);
  border-radius: 4px;
  margin: 2px 0;
}

.message-item {
  transition: background-color 0.2s;
}

.message-item:hover {
  background-color: rgba(13, 110, 253, 0.1);
}
</style>