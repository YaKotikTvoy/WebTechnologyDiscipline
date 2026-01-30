<template>
  <div class="h-100 d-flex flex-column">
    
    <ChatHeader 
      :chat-title="chatTitle"
      :chat-color="getChatColor()"
      :chat-initial="getChatInitial()"
      :member-count="memberCount"
      :show-back-button="true"
      @back="goBack"
    >
      <template #actions>
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
      :loading-message="`Загрузка... ${chatsStore.messages.length} сообщений в памяти`"
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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'

import ChatHeader from './ChatHeader.vue'
import ChatMessages from './ChatMessages.vue'
import ChatInput from './ChatInput.vue'
import AddUserToChatModal from '../AddUserToChatModal.vue'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const messagesComponent = ref(null)
const inputComponent = ref(null)
const showAddUserModalVisible = ref(false)
const loading = ref(false)

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

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesComponent.value) {
      messagesComponent.value.scrollToBottom()
    }
  })
}

const sendMessage = async ({ text, files }) => {
  if (!text.trim() && files.length === 0) return
  
  console.log('Отправка сообщения:', { text, filesCount: files.length })
  
  const result = await chatsStore.sendMessageWithFiles(
    chatId.value,
    text,
    files
  )
  
  if (result.success) {
    console.log('Сообщение отправлено успешно')
    if (inputComponent.value) {
      inputComponent.value.resetForm()
    }
    scrollToBottom()
  } else {
    console.error('Ошибка отправки сообщения:', result.error)
    alert('Ошибка отправки сообщения: ' + result.error)
  }
}

const handleEditMessage = async ({ messageId, content }) => {
  if (!chatId.value) {
    console.error('Нет chatId для редактирования сообщения')
    return
  }
  
  console.log('Вызов handleEditMessage:', { messageId, content, chatId: chatId.value })
  
  try {
    const result = await chatsStore.editMessage(chatId.value, messageId, content)
    
    if (result.success) {
      console.log('Сообщение успешно отредактировано')
    } else {
      console.error('Ошибка редактирования сообщения:', result.error)
      alert(result.error || 'Ошибка редактирования сообщения')
    }
  } catch (error) {
    console.error('Ошибка редактирования сообщения:', error)
    alert('Ошибка редактирования сообщения')
  }
}

const handleDeleteMessage = async (messageId) => {
  if (!chatId.value) {
    console.error('Нет chatId для удаления сообщения')
    return
  }
  
  console.log('Вызов handleDeleteMessage:', { messageId, chatId: chatId.value })
  
  try {
    const result = await chatsStore.deleteMessage(chatId.value, messageId)
    
    if (result.success) {
      console.log('Сообщение успешно удалено')
    } else {
      console.error('Ошибка удаления сообщения:', result.error)
      alert(result.error || 'Ошибка удаления сообщения')
    }
  } catch (error) {
    console.error('Ошибка удаления сообщения:', error)
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

onMounted(async () => {
  if (chatId.value) {
    await loadChatData()
  }
})

watch(
  () => route.params.id,
  async (newId) => {
    if (newId) {
      if (inputComponent.value) {
        inputComponent.value.resetForm()
      }
      await loadChatData()
    }
  }
)

watch(() => chatsStore.messages, (newMessages) => {
  if (newMessages.length > 0) {
    scrollToBottom()
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