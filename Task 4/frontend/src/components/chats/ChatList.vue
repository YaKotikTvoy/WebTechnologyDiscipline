<template>
  <div class="container-fluid mt-3">
    <div class="card">
      <div class="card-header d-flex justify-content-between align-items-center">
        <h5 class="mb-0">Чаты</h5>
        <button
          @click="showCreateChat = true"
          class="btn btn-sm btn-primary"
        >
          <i class="bi bi-plus-lg"></i> Новый чат
        </button>
      </div>
      <div class="card-body p-0">
        <div v-if="chats.length === 0" class="text-center py-5">
          <i class="bi bi-chat-dots display-1 text-muted mb-3"></i>
          <h5>Нет чатов</h5>
          <p class="text-muted">Создайте новый чат или добавьте друга</p>
          <button @click="showCreateChat = true" class="btn btn-primary">
            Создать чат
          </button>
        </div>
        <div v-else class="list-group list-group-flush">
          <router-link
            v-for="chat in chats"
            :key="chat.id"
            :to="`/chats/${chat.id}`"
            class="list-group-item list-group-item-action"
          >
            <div class="d-flex align-items-center">
              <div class="avatar-circle bg-primary text-white me-3">
                {{ getChatInitial(chat) }}
              </div>
              <div class="flex-grow-1">
                <div class="d-flex justify-content-between">
                  <strong>{{ getChatName(chat) }}</strong>
                  <small class="text-muted">{{ formatLastMessageTime(chat) }}</small>
                </div>
                <div class="d-flex justify-content-between">
                  <p class="mb-0 text-truncate" style="max-width: 200px;">
                    {{ getLastMessage(chat) }}
                  </p>
                  <span v-if="chat.unreadCount > 0" class="badge bg-danger">
                    {{ chat.unreadCount }}
                  </span>
                </div>
              </div>
            </div>
          </router-link>
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
              <div class="btn-group w-100" role="group">
                <input
                  type="radio"
                  class="btn-check"
                  v-model="newChat.type"
                  value="private"
                  id="type-private"
                >
                <label class="btn btn-outline-primary" for="type-private">
                  Приватный
                </label>
                <input
                  type="radio"
                  class="btn-check"
                  v-model="newChat.type"
                  value="group"
                  id="type-group"
                >
                <label class="btn btn-outline-primary" for="type-group">
                  Групповой
                </label>
              </div>
            </div>
            
            <div v-if="newChat.type === 'group'" class="mb-3">
              <label class="form-label">Название чата</label>
              <input
                v-model="newChat.name"
                type="text"
                class="form-control"
                placeholder="Введите название"
                required
              />
            </div>
            
            <div class="mb-3">
              <label class="form-label">
                {{ newChat.type === 'private' ? 'Номер друга' : 'Добавить участников' }}
              </label>
              <input
                v-model="newChat.phoneInput"
                type="text"
                class="form-control"
                placeholder="Введите номер телефона"
                @keyup.enter="addPhone"
              />
              <button @click="addPhone" class="btn btn-outline-secondary mt-2 w-100">
                Добавить
              </button>
            </div>
            
            <div v-if="newChat.memberPhones.length > 0" class="mb-3">
              <h6>Добавленные:</h6>
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
              :disabled="creating"
            >
              {{ creating ? 'Создание...' : 'Создать' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { useFriendsStore } from '@/stores/friends'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()
const friendsStore = useFriendsStore()

const chats = ref([])
const showCreateChat = ref(false)
const creating = ref(false)
const error = ref('')

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
  chats.value = chatsStore.chats.map(chat => ({
    ...chat,
    unreadCount: chat.unreadCount || 0
  }))
}

const getChatName = (chat) => {
  if (chat.type === 'private') {
    const otherMember = chat.members.find(m => m.id !== authStore.user?.id)
    return otherMember ? (otherMember.username || otherMember.phone) : 'Приватный чат'
  }
  return chat.name || 'Групповой чат'
}

const getChatInitial = (chat) => {
  if (chat.type === 'private') {
    const otherMember = chat.members.find(m => m.id !== authStore.user?.id)
    if (otherMember?.username) return otherMember.username.charAt(0).toUpperCase()
    if (otherMember?.phone) return otherMember.phone.slice(-1)
  }
  if (chat.name) return chat.name.charAt(0).toUpperCase()
  return '?'
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
    let result
    if (newChat.value.type === 'private') {
      if (newChat.value.memberPhones.length !== 1) {
        throw new Error('Приватный чат требует ровно одного участника')
      }
      result = await chatsStore.createPrivateChat(newChat.value.memberPhones[0])
    } else {
      result = await chatsStore.createGroupChat(
        newChat.value.name, 
        newChat.value.memberPhones,
        true
      )
    }

    if (result.success) {
      showCreateChat.value = false
      resetForm()
      await loadChats()
      if (result.chat) {
        router.push(`/chats/${result.chat.id}`)
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
    type: 'private',
    name: '',
    phoneInput: '',
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
.avatar-circle {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 1.2rem;
}
</style>