<template>
  <div class="h-100 d-flex flex-column">
    
    <ChatHeader 
      :chat-title="chatTitle"
      :chat-color="getChatColor()"
      :chat-initial="getChatInitial()"
      :member-count="memberCount"
      :chat-type="currentChat?.type"
      :show-back-button="true"
      @back="goBack"
    >
      <template #actions>
        <button class="btn btn-outline-secondary btn-sm me-2"
                @click="openChatSettings">
          <i class="bi bi-gear"></i>
        </button>
        
        <button v-if="currentChat?.type === 'group' && isChatCreator" 
                class="btn btn-outline-primary btn-sm"
                @click="openAddUserModal">
          <i class="bi bi-person-plus"></i> Добавить
        </button>
      </template>
    </ChatHeader>
    
    <ChatMessages 
      ref="messagesComponent"
      :messages="messages"
      :userId="userId"
      :loading="loading"
      :chat-type="currentChat?.type"
      :is-chat-admin="isChatCreator"
      @edit-message="handleEditMessage"
      @delete-message="handleDeleteMessage"
    />
    
    <ChatInput 
      ref="inputComponent"
      @send="sendMessage"
    />
    
    <AddUserToChatModal 
      v-if="showAddUserModalVisible"
      :show="showAddUserModalVisible"
      :chatId="chatId"
      :currentMembers="currentChat?.members || []"
      @close="closeAddUserModal"
      @added="handleUserAdded"
    />
    
    <ChatSettingsModal 
      v-if="showChatSettings"
      :show="showChatSettings"
      :chatId="chatId"
      :chatName="chatTitle"
      :chatType="currentChat?.type"
      :members="currentChat?.members || []"
      :creatorId="currentChat?.created_by"
      :currentUserId="userId"
      :chatColor="getChatColor()"
      @close="closeChatSettings"
      @deleted="handleChatDeleted"
      @left="handleLeftChat"
      @member-removed="handleMemberRemoved"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'

import ChatHeader from './ChatHeader.vue'
import ChatMessages from './ChatMessages.vue'
import ChatInput from './ChatInput.vue'
import AddUserToChatModal from '../AddUserToChatModal.vue'
import ChatSettingsModal from './ChatSettingsModal.vue'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const observer = ref(null)
const readMessagesQueue = ref(new Set())
const messagesComponent = ref(null)
const inputComponent = ref(null)
const showAddUserModalVisible = ref(false)
const loading = ref(false)
const showChatSettings = ref(false)
const forceRefresh = ref(false)

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

const isChatCreator = computed(() => {
  return currentChat.value?.created_by === authStore.user?.id
})

const openChatSettings = () => {
  showChatSettings.value = true
}

const closeChatSettings = () => {
  showChatSettings.value = false
}

const handleChatDeleted = async (deletedChatId) => {
  if (deletedChatId === chatId.value) {
    await chatsStore.fetchChats()
    router.push('/')
  }
}

const handleLeftChat = async (leftChatId) => {
  if (leftChatId === chatId.value) {
    await chatsStore.fetchChats()
    router.push('/')
  }
}

const handleMemberRemoved = async (memberId) => {
  await chatsStore.fetchChat(chatId.value)
}

const openAddUserModal = () => {
  showAddUserModalVisible.value = true
}

const closeAddUserModal = () => {
  showAddUserModalVisible.value = false
}

const handleUserAdded = async () => {
  await chatsStore.fetchChat(chatId.value)
  closeAddUserModal()
}

const markMessageAsRead = async (messageId) => {
  if (!chatId.value || !messageId) return
  
  try {
    if (readMessagesQueue.value.has(messageId)) return
    readMessagesQueue.value.add(messageId)
    
    await wsStore.markMessageAsRead(chatId.value, messageId)
    
    readMessagesQueue.value.delete(messageId)
  } catch (error) {
    console.error('Ошибка пометки сообщения как прочитанного:', error)
    readMessagesQueue.value.delete(messageId)
  }
}

const markAllMessagesAsRead = async () => {
    if (!chatId.value || !currentChat.value?.members) return
    
    try {
        await chatsStore.markSpecificChatAsRead(chatId.value)
        
        const chatIndex = chatsStore.chats.findIndex(c => c.id === chatId.value)
        if (chatIndex !== -1) {
            chatsStore.chats[chatIndex].unreadCount = 0
        }
    } catch (error) {
        console.error('Ошибка пометки чата как прочитанного:', error)
    }
}

const setupMessageObserver = () => {
  if (!messagesComponent.value || !messagesComponent.value.getContainer) return
  
  const container = messagesComponent.value.getContainer()
  if (!container) return
  
  if (observer.value) {
    observer.value.disconnect()
    observer.value = null
  }
  
  observer.value = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const messageId = parseInt(entry.target.dataset.messageId)
        if (messageId && !isNaN(messageId)) {
          markMessageAsRead(messageId)
        }
      }
    })
  }, {
    root: container,
    rootMargin: '0px',
    threshold: 0.5
  })
  
  const messageElements = container.querySelectorAll('.message-item[data-message-id]')
  messageElements.forEach(element => {
    observer.value.observe(element)
  })
}

const loadChatData = async () => {
    if (!chatId.value) return
    
    loading.value = true
    try {
        const chatExists = chatsStore.chats.some(chat => chat.id === chatId.value)
        if (!chatExists) {
            router.push('/')
            return
        }
        
        await chatsStore.fetchChat(chatId.value)
        const messages = await chatsStore.getMessages(chatId.value, forceRefresh.value)
        
        scrollToBottom()
        
        nextTick(() => {
            markAllMessagesAsRead()
            setupMessageObserver()
        })
        
        chatsStore.setActiveChat(chatId.value)
    } catch (error) {
        console.error('Ошибка загрузки данных чата:', error)
        router.push('/')
    } finally {
        loading.value = false
    }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesComponent.value) {
      messagesComponent.value.scrollToBottom()
    }
  })
}


const sendMessage = async ({ text, files }) => {
  if (!text.trim() && files.length === 0) return
  
  const result = await chatsStore.sendMessageWithFiles(
    chatId.value,
    text,
    files
  )
  
  if (result.success) {
    if (inputComponent.value) {
      inputComponent.value.resetForm()
    }
    scrollToBottom()
  } else {
    alert('Ошибка отправки сообщения: ' + result.error)
  }
}

const handleEditMessage = async ({ messageId, content }) => {
  if (!chatId.value) return
  
  try {
    const result = await chatsStore.editMessage(chatId.value, messageId, content)
    
    if (!result.success) {
      alert(result.error || 'Ошибка редактирования сообщения')
    }
  } catch (error) {
    alert('Ошибка редактирования сообщения')
  }
}

const handleDeleteMessage = async (messageId) => {
  if (!chatId.value) return
  
  try {
    const result = await chatsStore.deleteMessage(chatId.value, messageId)
    
    if (!result.success) {
      alert(result.error || 'Ошибка удаления сообщения')
    }
  } catch (error) {
    alert('Ошибка удаления сообщения')
  }
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

const goBack = () => {
  router.push('/')
}

const handleVisibilityChange = () => {
  if (!document.hidden && chatId.value) {
    markAllMessagesAsRead()
  }
}

onMounted(async () => {
  if (chatId.value) {
    await loadChatData()
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  if (observer.value) {
    observer.value.disconnect()
    observer.value = null
  }
  
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

watch(
  () => route.params.id,
  async (newId) => {
    if (newId) {
      const chatIdNum = parseInt(newId)
      console.log('Смена чата через URL:', chatIdNum)

      setTimeout(() => {
        chatsStore.setActiveChat(chatIdNum)
      }, 50)
      
      await loadChatData()
    }
  },
  { immediate: true }
)

watch(() => chatsStore.messages, (newMessages) => {
  if (newMessages.length > 0) {
    scrollToBottom()
    nextTick(() => {
      setupMessageObserver()
    })
  }
}, { deep: true })

watch(() => wsStore.notifications, (notifications) => {
  const chatNotifications = notifications.filter(n => 
    (n.type === 'new_message' && n.data.chatId === chatId.value) ||
    (n.type === 'message_read' && n.data.chat_id === chatId.value)
  )
  if (chatNotifications.length > 0) {
    loadChatData()
  }
}, { deep: true })
</script>

<style scoped>
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