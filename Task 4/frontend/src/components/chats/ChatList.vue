<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-4">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h5 class="mb-0">Чаты</h5>
            <div>
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
            <h5 class="modal-title">Создать групповой чат</h5>
            <button
              type="button"
              class="btn-close"
              @click="showCreateChat = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Название чата</label>
              <input
                v-model="newChat.name"
                type="text"
                class="form-control"
                placeholder="Введите название чата"
                required
              />
            </div>
            
            <div class="mb-3">
              <label class="form-label">Добавить участников</label>
              <div class="mb-2">
                <div class="form-check" v-for="friend in friends" :key="friend.id">
                  <input
                    class="form-check-input"
                    type="checkbox"
                    :value="friend.friend.phone"
                    v-model="newChat.selectedFriends"
                    :id="'friend-' + friend.id"
                  >
                  <label class="form-check-label" :for="'friend-' + friend.id">
                    {{ friend.friend.username || friend.friend.phone }}
                  </label>
                </div>
              </div>
              
              <div class="mt-3">
                <label class="form-label">Или добавьте по номеру телефона</label>
                <div class="input-group">
                  <input
                    v-model="newChat.phoneInput"
                    type="text"
                    class="form-control"
                    placeholder="Введите номер телефона"
                    @keyup.enter="addPhone"
                  />
                  <button @click="addPhone" class="btn btn-outline-secondary" type="button">
                    Добавить
                  </button>
                </div>
              </div>
            </div>
            
            <div v-if="newChat.memberPhones.length > 0" class="mb-3">
              <h6>Добавленные участники:</h6>
              <div
                v-for="(phone, index) in newChat.memberPhones"
                :key="index"
                class="badge bg-primary me-2 mb-2 p-2"
              >
                {{ phone }}
                <button
                  type="button"
                  class="btn-close btn-close-white ms-1"
                  @click="removePhone(index)"
                ></button>
              </div>
            </div>
            
            <div v-if="error" class="alert alert-danger mt-2">
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
              :disabled="creating || !isFormValid"
            >
              {{ creating ? 'Создание...' : 'Создать чат' }}
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
import { useFriendsStore } from '@/stores/friends'
import { api } from '@/services/api'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()
const friendsStore = useFriendsStore()

const chats = ref([])
const friends = ref([])
const showCreateChat = ref(false)
const creating = ref(false)
const error = ref('')

const currentChatId = computed(() => {
  return route.params.id ? parseInt(route.params.id) : null
})

const newChat = ref({
  name: '',
  phoneInput: '',
  selectedFriends: [],
  memberPhones: []
})

const isFormValid = computed(() => {
  return newChat.value.name.trim() !== '' && 
         (newChat.value.selectedFriends.length > 0 || newChat.value.memberPhones.length > 0)
})

onMounted(async () => {
  await loadChats()
  await loadFriends()
})

const loadFriends = async () => {
  await friendsStore.fetchFriends()
  friends.value = friendsStore.friends
}

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
      chat.unreadCount = response.data.count || 0
    } catch (error) {
      console.error('Не удалось загрузить количество непрочитанных:', error)
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
      console.error('Не удалось отметить чат как прочитанный:', error)
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

  try {
    const memberPhones = [...new Set([
      ...newChat.value.selectedFriends,
      ...newChat.value.memberPhones
    ])]
    
    if (memberPhones.length === 0) {
      throw new Error('Добавьте хотя бы одного участника')
    }

    const result = await chatsStore.createGroupChat(newChat.value.name, memberPhones)

    if (result.success) {
      showCreateChat.value = false
      resetForm()
      await loadChats()
      if (result.chat) {
        router.push(`/chats/${result.chat.id}`)
      } else {
        error.value = result.message || 'Чат создан'
      }
    } else {
      error.value = result.error
    }
  } catch (err) {
    error.value = err.message
  }

  creating.value = false
}

const resetForm = () => {
  newChat.value = {
    name: '',
    phoneInput: '',
    selectedFriends: [],
    memberPhones: []
  }
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