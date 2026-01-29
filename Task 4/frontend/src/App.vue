<template>
  <div id="app">
    <template v-if="authStore.isAuthenticated">
      <div class="container-fluid h-100">
        <div class="row h-100">
          <div class="col-4 col-md-3 h-100 bg-light border-end">
            <div class="p-3 border-bottom">
              <div class="d-flex align-items-center">
                <div class="dropdown">
                  <button class="btn p-0 d-flex align-items-center w-100 text-start" 
                          type="button" 
                          data-bs-toggle="dropdown">
                    <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-2" 
                         style="width: 40px; height: 40px;">
                      {{ getUserInitial() }}
                    </div>
                    <div class="flex-grow-1">
                      <div class="fw-bold">{{ authStore.user?.username || authStore.user?.phone }}</div>
                      <div class="small text-muted">online</div>
                    </div>
                    <i class="bi bi-chevron-down"></i>
                  </button>
                  <ul class="dropdown-menu">
                    <li><button class="dropdown-item" @click="goToProfile">Профиль</button></li>
                    <li><button class="dropdown-item" @click="createNewChat">Новый чат</button></li>
                    <li><button class="dropdown-item" @click="createNewGroup">Новая группа</button></li>
                    <li><button class="dropdown-item" @click="goToContacts">Контакты</button></li>
                    <li><hr class="dropdown-divider"></li>
                    <li><button class="dropdown-item text-danger" @click="logout">Выйти</button></li>
                  </ul>
                </div>
              </div>
            </div>

            <div class="p-3 border-bottom">
              <div class="input-group input-group-sm">
                <span class="input-group-text"><i class="bi bi-search"></i></span>
                <input type="text" 
                       v-model="searchQuery" 
                       class="form-control" 
                       placeholder="Поиск">
              </div>
            </div>

            <div class="chat-list overflow-auto" style="height: calc(100vh - 150px)">
              <div v-for="chat in filteredChats" :key="chat.id" 
                   class="chat-item list-group-item list-group-item-action"
                   :class="{ active: activeChatId === chat.id }"
                   @click="openChat(chat.id)">
                <div class="d-flex align-items-center">
                  <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-2" 
                       style="width: 50px; height: 50px;">
                    {{ getChatInitial(chat) }}
                  </div>
                  <div class="flex-grow-1">
                    <div class="d-flex justify-content-between">
                      <div class="fw-bold">{{ getChatName(chat) }}</div>
                      <small class="text-muted">{{ formatTime(chat.updated_at) }}</small>
                    </div>
                    <div class="d-flex justify-content-between">
                      <div class="text-truncate" style="max-width: 150px;">{{ getLastMessage(chat) }}</div>
                      <span v-if="chat.unreadCount > 0" class="badge bg-danger rounded-pill">{{ chat.unreadCount }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="col-8 col-md-9 h-100 p-0">
            <router-view />
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <router-view />
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
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
const activeChatId = ref(null)
const chats = ref([])

onMounted(async () => {
  if (authStore.isAuthenticated) {
    await loadChats()
    wsStore.connect()
  }
})

const loadChats = async () => {
  await chatsStore.fetchChats()
  chats.value = chatsStore.chats.map(chat => ({
    ...chat,
    unreadCount: chat.unreadCount || 0
  }))
}

watch(() => route.params.id, (newId) => {
  if (newId) {
    activeChatId.value = parseInt(newId)
  }
})

const filteredChats = computed(() => {
  if (!searchQuery.value) return chats.value
  return chats.value.filter(chat => {
    const name = getChatName(chat).toLowerCase()
    return name.includes(searchQuery.value.toLowerCase())
  })
})

const openChat = (chatId) => {
  activeChatId.value = chatId
  router.push(`/chats/${chatId}`)
}

const goToProfile = () => {
  router.push('/profile')
}

const createNewChat = () => {
  const phone = prompt('Введите номер телефона:')
  if (phone) {
    chatsStore.createPrivateChat(phone).then(result => {
      if (result.success && result.chat) {
        openChat(result.chat.id)
        loadChats()
      }
    })
  }
}

const createNewGroup = () => {
  const name = prompt('Название группы:')
  if (name) {
    const members = prompt('Номера участников через запятую:')
    if (members) {
      const phones = members.split(',').map(p => p.trim())
      chatsStore.createGroupChat(name, phones, true).then(result => {
        if (result.success && result.chat) {
          openChat(result.chat.id)
          loadChats()
        }
      })
    }
  }
}

const goToContacts = () => {
  router.push('/contacts')
}

const logout = () => {
  authStore.logout()
  wsStore.disconnect()
  router.push('/login')
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
</script>