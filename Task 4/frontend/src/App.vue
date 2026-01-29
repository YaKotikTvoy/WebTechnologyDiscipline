<template>
  <div id="app">
    <template v-if="authStore.isAuthenticated">
      <div class="container-fluid h-100 p-0">
        <div class="row h-100 g-0">
          <div class="col-4 col-md-3 h-100 border-end bg-light d-flex flex-column">
            <div class="p-3 border-bottom bg-white">
              <div class="dropdown">
                <button class="btn p-0 d-flex align-items-center w-100 text-start" 
                        type="button" 
                        data-bs-toggle="dropdown">
                  <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                       style="width: 45px; height: 45px;">
                    {{ getUserInitial() }}
                  </div>
                  <div class="flex-grow-1">
                    <div class="fw-bold">{{ authStore.user?.username || authStore.user?.phone }}</div>
                    <div class="small text-muted">online</div>
                  </div>
                  <i class="bi bi-chevron-down"></i>
                </button>
                <ul class="dropdown-menu w-100">
                  <li><button class="dropdown-item" @click="goToProfile">Профиль</button></li>
                  <li><button class="dropdown-item" @click="showAddContactModal">Добавить контакт</button></li>
                  <li><button class="dropdown-item" @click="showNewGroupModal">Новая группа</button></li>
                  <li><hr class="dropdown-divider"></li>
                  <li><button class="dropdown-item text-danger" @click="logout">Выйти</button></li>
                </ul>
              </div>
            </div>

            <div class="p-3 border-bottom bg-white">
              <div class="input-group">
                <span class="input-group-text bg-transparent border-end-0"><i class="bi bi-search"></i></span>
                <input type="text" 
                       v-model="searchQuery" 
                       class="form-control border-start-0" 
                       placeholder="Поиск">
              </div>
            </div>

            <div class="flex-grow-1 overflow-auto">
              <div v-if="loading" class="text-center py-4">
                <div class="spinner-border spinner-border-sm" role="status">
                  <span class="visually-hidden">Загрузка...</span>
                </div>
              </div>
              
              <div v-else-if="filteredChats.length === 0" class="text-center py-4">
                <i class="bi bi-chat-dots display-6 text-muted mb-3"></i>
                <p class="text-muted">Нет чатов</p>
              </div>
              
              <div v-else>
                <div v-for="chat in filteredChats" :key="chat.id" 
                    class="chat-item px-3 py-2 border-bottom"
                    :class="{ 'active-chat': activeChatId === chat.id }"
                    @click="openChat(chat.id)">
                  <div class="d-flex align-items-center">
                    <div class="rounded-circle d-flex align-items-center justify-content-center me-2 me-md-3" 
                        :class="getChatColor(chat.id)"
                        style="width: 40px; height: 40px; font-size: 0.9rem;">
                      {{ getChatInitial(chat) }}
                    </div>
                    
                    <div class="flex-grow-1 d-none d-md-block">
                      <div class="d-flex justify-content-between align-items-center">
                        <div class="fw-bold text-truncate" style="max-width: 120px;">
                          {{ getChatName(chat) }}
                        </div>
                        <small class="text-muted">{{ formatTime(chat.updated_at) }}</small>
                      </div>
                      <div class="d-flex justify-content-between align-items-center mt-1">
                        <div class="text-truncate text-muted small" style="max-width: 150px;">
                          {{ getLastMessage(chat) }}
                        </div>
                        <span v-if="chat.unreadCount > 0" 
                              class="badge bg-danger rounded-pill">
                          {{ chat.unreadCount }}
                        </span>
                      </div>
                    </div>
                    
                    <div v-if="chat.unreadCount > 0" class="d-md-none ms-auto">
                      <span class="badge bg-danger rounded-pill">{{ chat.unreadCount }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="col-8 col-md-9 h-100">
            <router-view />
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <router-view />
    </template>

    <div v-if="showContactModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Добавить контакт</h5>
            <button type="button" class="btn-close" @click="closeContactModal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Номер телефона</label>
              <input v-model="newContactPhone" type="text" class="form-control" placeholder="7XXXXXXXXXX">
            </div>
            <div v-if="contactError" class="alert alert-danger">{{ contactError }}</div>
            <div v-if="contactSuccess" class="alert alert-success">{{ contactSuccess }}</div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeContactModal">Отмена</button>
            <button type="button" class="btn btn-primary" @click="addContact" :disabled="addingContact">
              {{ addingContact ? 'Добавление...' : 'Добавить' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showGroupModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Новая группа</h5>
            <button type="button" class="btn-close" @click="closeGroupModal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Название группы</label>
              <input v-model="newGroupName" type="text" class="form-control" placeholder="Введите название">
            </div>
            <div class="mb-3">
              <label class="form-label">Видимость группы</label>
              <div class="form-check">
                <input class="form-check-input" type="radio" v-model="newGroupVisible" value="true" id="visible-true">
                <label class="form-check-label" for="visible-true">
                  Видимая в поиске
                </label>
              </div>
              <div class="form-check">
                <input class="form-check-input" type="radio" v-model="newGroupVisible" value="false" id="visible-false">
                <label class="form-check-label" for="visible-false">
                  Скрытая (только по приглашению)
                </label>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">Добавить участников</label>
              <div class="input-group mb-2">
                <input v-model="newGroupPhone" type="text" class="form-control" placeholder="Номер телефона">
                <button class="btn btn-outline-secondary" type="button" @click="addGroupParticipant">
                  <i class="bi bi-plus"></i>
                </button>
              </div>
              <div v-if="groupParticipants.length > 0">
                <div v-for="(phone, index) in groupParticipants" :key="index" class="badge bg-primary me-1 mb-1">
                  {{ phone }}
                  <button type="button" class="btn-close btn-close-white ms-1" @click="removeGroupParticipant(index)"></button>
                </div>
              </div>
            </div>
            <div v-if="groupError" class="alert alert-danger">{{ groupError }}</div>
            <div v-if="groupSuccess" class="alert alert-success">{{ groupSuccess }}</div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeGroupModal">Отмена</button>
            <button type="button" class="btn btn-primary" @click="createNewGroup" :disabled="creatingGroup">
              {{ creatingGroup ? 'Создание...' : 'Создать' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { useChatsStore } from '@/stores/chats'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()
const chatsStore = useChatsStore()

const searchQuery = ref('')
const loading = ref(false)
const activeChatId = computed(() => chatsStore.activeChatId)
const chats = ref([])

const showContactModal = ref(false)
const newContactPhone = ref('')
const addingContact = ref(false)
const contactError = ref('')
const contactSuccess = ref('')

const showGroupModal = ref(false)
const newGroupName = ref('')
const newGroupVisible = ref('true')
const newGroupPhone = ref('')
const groupParticipants = ref([])
const creatingGroup = ref(false)
const groupError = ref('')
const groupSuccess = ref('')

const refreshTimer = ref(null)

onMounted(async () => {
  if (authStore.isAuthenticated) {
    await authStore.fetchUser()
    await loadChats()
    wsStore.connect()
    
    refreshTimer.value = setInterval(async () => {
      if (authStore.isAuthenticated) {
        await chatsStore.refreshUnreadCounts()
        chats.value = chatsStore.chats
      }
    }, 30000)
  }
})

onUnmounted(() => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
})

watch(() => route.params.id, (newId) => {
  if (newId) {
    const chatId = parseInt(newId)
    chatsStore.setActiveChat(chatId)
    markChatAsRead(chatId)
  }
})

const filteredChats = computed(() => {
  if (!searchQuery.value) return chats.value
  return chats.value.filter(chat => {
    const name = getChatName(chat).toLowerCase()
    return name.includes(searchQuery.value.toLowerCase())
  })
})
const loadChats = async () => {
  if (loading.value) return
  
  loading.value = true
  try {
    await chatsStore.fetchChats()
    chats.value = chatsStore.chats
  } finally {
    loading.value = false
  }
}

const openChat = async (chatId) => {
  chatsStore.setActiveChat(chatId)
  await chatsStore.markChatAsRead(chatId)
  const chatIndex = chats.value.findIndex(c => c.id === chatId)
  if (chatIndex !== -1) {
    chats.value[chatIndex].unreadCount = 0
  }
  router.push(`/chats/${chatId}`)
}

const markChatAsRead = async (chatId) => {
  try {
    await chatsStore.markChatAsRead(chatId)
  } catch (error) {
    console.error('Ошибка пометки чата как прочитанного:', error)
  }
}

const goToProfile = () => {
  router.push('/profile')
}

const showAddContactModal = () => {
  showContactModal.value = true
  newContactPhone.value = ''
  contactError.value = ''
  contactSuccess.value = ''
}

const closeContactModal = () => {
  showContactModal.value = false
  newContactPhone.value = ''
  contactError.value = ''
  contactSuccess.value = ''
}

const addContact = async () => {
  if (!newContactPhone.value) {
    contactError.value = 'Введите номер телефона'
    return
  }

  addingContact.value = true
  contactError.value = ''
  contactSuccess.value = ''

  try {
    const result = await chatsStore.createPrivateChat(newContactPhone.value)
    
    if (result.success) {
      contactSuccess.value = 'Контакт добавлен'
      await loadChats()
      
      if (result.chat) {
        setTimeout(() => {
          closeContactModal()
          openChat(result.chat.id)
        }, 1000)
      }
    } else {
      contactError.value = result.error || 'Ошибка добавления контакта'
    }
  } catch (error) {
    contactError.value = 'Ошибка добавления контакта'
  } finally {
    addingContact.value = false
  }
}

const showNewGroupModal = () => {
  showGroupModal.value = true
  newGroupName.value = ''
  newGroupPhone.value = ''
  groupParticipants.value = []
  groupError.value = ''
  groupSuccess.value = ''
}

const closeGroupModal = () => {
  showGroupModal.value = false
  newGroupName.value = ''
  newGroupPhone.value = ''
  groupParticipants.value = []
  groupError.value = ''
  groupSuccess.value = ''
}

const addGroupParticipant = () => {
  if (newGroupPhone.value && !groupParticipants.value.includes(newGroupPhone.value)) {
    groupParticipants.value.push(newGroupPhone.value)
    newGroupPhone.value = ''
  }
}

const removeGroupParticipant = (index) => {
  groupParticipants.value.splice(index, 1)
}

const createNewGroup = async () => {
  if (!newGroupName.value) {
    groupError.value = 'Введите название группы'
    return
  }

  if (groupParticipants.value.length === 0) {
    groupError.value = 'Добавьте хотя бы одного участника'
    return
  }

  creatingGroup.value = true
  groupError.value = ''
  groupSuccess.value = ''

  try {
    const isSearchable = newGroupVisible.value === 'true'
    const result = await chatsStore.createGroupChat(
      newGroupName.value, 
      groupParticipants.value,
      isSearchable
    )
    
    if (result.success) {
      groupSuccess.value = 'Группа создана'
      await loadChats()
      
      if (result.chat) {
        setTimeout(() => {
          closeGroupModal()
          openChat(result.chat.id)
        }, 1000)
      }
    } else {
      groupError.value = result.error || 'Ошибка создания группы'
    }
  } catch (error) {
    groupError.value = 'Ошибка создания группы'
  } finally {
    creatingGroup.value = false
  }
}

const getChatColor = (chatId) => {
  const colors = ['bg-primary', 'bg-success', 'bg-warning', 'bg-danger', 'bg-info', 'bg-secondary']
  const index = chatId % colors.length
  return colors[index]
}

const getChatName = (chat) => {
  if (chat.type === 'private') {
    const otherMember = chat.members?.find(m => m.id !== authStore.user?.id)
    return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
  }
  return chat.name || 'Групповой чат'
}

const getChatInitial = (chat) => {
  if (!chat) return '?'
  const name = getChatName(chat)
  return name.charAt(0).toUpperCase()
}

const getUserInitial = () => {
  if (!authStore.user) return '?'
  if (authStore.user.username) return authStore.user.username.charAt(0).toUpperCase()
  return authStore.user.phone ? authStore.user.phone.slice(-1) : '?'
}

const getLastMessage = () => {
  return 'Начните общение...'
}

const formatTime = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date
  
  if (diff < 3600000) return 'только что'
  if (diff < 86400000) return 'сегодня'
  return date.toLocaleDateString()
}

const logout = () => {
  authStore.logout()
  wsStore.disconnect()
  router.push('/login')
}
</script>

<style>
.chat-item {
  cursor: pointer;
  transition: background-color 0.2s;
}

.chat-item:hover {
  background-color: #f8f9fa !important;
}

.active-chat {
  background-color: #0d6efd !important;
  color: white !important;
}

.active-chat:hover {
  background-color: #0d6efd !important;
}

.active-chat .text-muted {
  color: rgba(255, 255, 255, 0.8) !important;
}
</style>