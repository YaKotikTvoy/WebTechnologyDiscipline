<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-4">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h5 class="mb-0">Чаты</h5>
            <div>
              <button
                @click="showJoinChat = true"
                class="btn btn-sm btn-outline-primary me-2"
              >
                Найти групповой чат
              </button>
              <button
                @click="showCreateChat = true"
                class="btn btn-sm btn-primary"
              >
                Новый чат
              </button>
            </div>
          </div>
          <div class="card-body p-0">
            <div v-if="chats.length === 0" class="text-center py-3">
              Нет чатов
            </div>
            <div v-else class="list-group list-group-flush">
              <router-link
                v-for="chat in chats"
                :key="chat.id"
                :to="`/chats/${chat.id}`"
                class="list-group-item list-group-item-action position-relative"
                :class="{ active: chat.id === currentChatId }"
              >
                <div class="d-flex w-100 justify-content-between">
                  <h6 class="mb-1">
                    {{ chat.name || getChatName(chat) }}
                    <span v-if="chat.unreadCount > 0" class="badge bg-danger ms-1">
                      {{ chat.unreadCount }}
                    </span>
                  </h6>
                  <small class="text-muted">
                    {{ formatLastMessageTime(chat) }}
                  </small>
                </div>
                <p class="mb-1 text-truncate">
                  {{ getLastMessage(chat) }}
                </p>
                <small v-if="chat.type === 'group'" class="text-muted">
                  Группа • {{ chat.members.length }} участников
                </small>
                <small v-else class="text-muted">
                  Приватный чат
                </small>
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <div class="col-md-8">
        <div class="card">
          <div class="card-body text-center py-5">
            <h5>Выберите чат</h5>
            <p class="text-muted">
              Выберите чат из списка или создайте новый
            </p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateChat" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Создать чат</h5>
            <button
              type="button"
              class="btn-close"
              @click="showCreateChat = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Тип чата</label>
              <select v-model="newChat.type" class="form-select">
                <option value="private">Приватный</option>
                <option value="group">Групповой</option>
              </select>
            </div>
            <div v-if="newChat.type === 'group'" class="mb-3">
              <label class="form-label">Название чата</label>
              <input
                v-model="newChat.name"
                type="text"
                class="form-control"
                placeholder="Введите название чата"
              />
            </div>
            <div class="mb-3">
              <label class="form-label">
                {{ newChat.type === 'private' ? 'Телефон пользователя' : 'Телефоны пользователей (через запятую)' }}
              </label>
              <input
                v-model="newChat.phoneInput"
                type="text"
                class="form-control"
                :placeholder="newChat.type === 'private' ? 'Введите телефон пользователя' : 'Введите телефоны пользователей через запятую'"
                @keyup.enter="addPhone"
              />
              <small class="text-muted">
                {{ newChat.type === 'private' ? 'Введите номер телефона пользователя для приватного чата' : 'Введите номера телефонов пользователей для добавления в группу' }}
              </small>
            </div>
            <div v-if="newChat.memberPhones.length > 0" class="mb-3">
              <div
                v-for="(phone, index) in newChat.memberPhones"
                :key="index"
                class="badge bg-primary me-2 mb-2"
              >
                {{ phone }}
                <button
                  type="button"
                  class="btn-close btn-close-white ms-1"
                  @click="removePhone(index)"
                ></button>
              </div>
            </div>
            <div v-if="error" class="alert alert-danger">
              {{ error }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="showCreateChat = false"
            >
              Отмена
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="createChat"
              :disabled="creating"
            >
              {{ creating ? 'Создание...' : 'Создать чат' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showJoinChat" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Найти групповой чат</h5>
            <button
              type="button"
              class="btn-close"
              @click="showJoinChat = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Название чата или ID</label>
              <input
                v-model="searchChatQuery"
                type="text"
                class="form-control"
                placeholder="Введите название или ID чата"
              />
            </div>
            <div v-if="searchChatResults.length > 0" class="mt-3">
              <div class="list-group">
                <div
                  v-for="chat in searchChatResults"
                  :key="chat.id"
                  class="list-group-item"
                >
                  <div class="d-flex justify-content-between align-items-center">
                    <div>
                      <h6 class="mb-1">{{ chat.name }}</h6>
                      <small class="text-muted">
                        Группа • {{ chat.members?.length || 0 }} участников
                      </small>
                    </div>
                    <button
                      @click="joinChat(chat.id)"
                      class="btn btn-sm btn-primary"
                      :disabled="joining"
                    >
                      Вступить
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="showJoinChat = false"
            >
              Отмена
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="searchChats"
              :disabled="searchingChats"
            >
              {{ searchingChats ? 'Поиск...' : 'Найти' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { api } from '@/services/api'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

const chats = ref([])
const showCreateChat = ref(false)
const showJoinChat = ref(false)
const creating = ref(false)
const joining = ref(false)
const error = ref('')
const searchChatQuery = ref('')
const searchChatResults = ref([])
const searchingChats = ref(false)

const currentChatId = computed(() => {
  return route.params.id ? parseInt(route.params.id) : null
})

const newChat = ref({
  type: 'private',
  name: '',
  phoneInput: '',
  memberPhones: []
})

onMounted(async () => {
  await loadChats()
})

const loadChats = async () => {
  await chatsStore.fetchChats()
  await calculateUnreadCounts()
  chats.value = chatsStore.chats.map(chat => ({
    ...chat,
    unreadCount: chat.unreadCount || 0
  }))
}

const calculateUnreadCounts = async () => {
  for (const chat of chatsStore.chats) {
    try {
      const response = await api.get(`/chats/${chat.id}/unread`)
      chat.unreadCount = response.data.count
    } catch (error) {
      chat.unreadCount = 0
    }
  }
}

watch(() => route.params.id, async (newChatId) => {
  if (newChatId) {
    try {
      await api.post(`/chats/${newChatId}/read`)
      const chatIndex = chats.value.findIndex(c => c.id === parseInt(newChatId))
      if (chatIndex !== -1) {
        chats.value[chatIndex].unreadCount = 0
      }
    } catch (error) {
      console.error('Failed to mark chat as read:', error)
    }
  }
})

const getChatName = (chat) => {
  if (chat.type === 'private') {
    const otherMember = chat.members.find(m => m.id !== authStore.user?.id)
    return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
  }
  return chat.name || 'Групповой чат'
}

const getLastMessage = (chat) => {
  return 'Начните общение...'
}

const formatLastMessageTime = (chat) => {
  return ''
}

const addPhone = () => {
  const phone = newChat.value.phoneInput.trim()
  if (phone && !newChat.value.memberPhones.includes(phone)) {
    newChat.value.memberPhones.push(phone)
    newChat.value.phoneInput = ''
  }
}

const removePhone = (index) => {
  newChat.value.memberPhones.splice(index, 1)
}

const createChat = async () => {
  creating.value = true
  error.value = ''

  if (newChat.value.type === 'private' && newChat.value.memberPhones.length !== 1) {
    error.value = 'Приватный чат требует ровно одного пользователя'
    creating.value = false
    return
  }

  let result
  if (newChat.value.type === 'private') {
    result = await chatsStore.createPrivateChat(newChat.value.memberPhones[0])
  } else {
    result = await chatsStore.createGroupChat(newChat.value.name, newChat.value.memberPhones)
  }

  if (result.success) {
    showCreateChat.value = false
    newChat.value = {
      type: 'private',
      name: '',
      phoneInput: '',
      memberPhones: []
    }
    await loadChats()
    router.push(`/chats/${result.chat.id}`)
  } else {
    error.value = result.error
  }

  creating.value = false
}

const searchChats = async () => {
  searchingChats.value = true
  try {
    const response = await api.get('/chats')
    searchChatResults.value = response.data.filter(chat => 
      chat.type === 'group' && 
      !chat.members.some(m => m.id === authStore.user?.id)
    )
  } catch (error) {
    console.error('Failed to search chats:', error)
  }
  searchingChats.value = false
}

const joinChat = async (chatId) => {
  joining.value = true
  try {
    await api.post(`/chats/${chatId}/join`)
    await loadChats()
    showJoinChat.value = false
    router.push(`/chats/${chatId}`)
  } catch (error) {
    console.error('Failed to join chat:', error)
  }
  joining.value = false
}

watch(() => wsStore.isConnected, () => {
  if (wsStore.isConnected) {
    loadChats()
  }
})
</script>

<style scoped>
.list-group-item.active {
  background-color: #007bff;
  border-color: #007bff;
}
</style>