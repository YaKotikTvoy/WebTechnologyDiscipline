<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-12">
        <div class="card">
          <div class="card-header">
            <div class="d-flex justify-content-between align-items-center">
              <h5 class="mb-0">Чаты для вступления</h5>
              <div class="w-50">
                <div class="input-group">
                  <input v-model="searchQuery" type="text" class="form-control" 
                         placeholder="Поиск чатов по названию" @input="searchChats">
                  <button class="btn btn-outline-secondary" type="button" @click="searchChats">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-search" viewBox="0 0 16 16">
                      <path d="M11.742 10.344a6.5 6.5 0 1 0-1.397 1.398h-.001c.03.04.062.078.098.115l3.85 3.85a1 1 0 0 0 1.415-1.414l-3.85-3.85a1.007 1.007 0 0 0-.115-.1zM12 6.5a5.5 5.5 0 1 1-11 0 5.5 5.5 0 0 1 11 0z"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div class="card-body">
            <div v-if="loading" class="text-center py-3">
              <div class="spinner-border" role="status">
                <span class="visually-hidden">Загрузка...</span>
              </div>
            </div>
            
            <div v-else-if="chats.length === 0" class="text-center py-3">
              {{ searchQuery ? 'Чаты не найдены' : 'Нет доступных чатов' }}
            </div>
            
            <div v-else class="list-group">
              <div v-for="chat in chats" :key="chat.id" class="list-group-item">
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <h6 class="mb-1">{{ chat.name }}</h6>
                    <small class="text-muted">
                      Группа • {{ chat.members?.length || 0 }} участников
                    </small>
                  </div>
                  <div>
                    <button v-if="!isMember(chat)" 
                            @click="sendJoinRequest(chat.id)"
                            class="btn btn-sm btn-primary"
                            :disabled="joinRequestStatus[chat.id]">
                      {{ getJoinButtonText(chat.id) }}
                    </button>
                    <span v-else class="badge bg-success">
                      Участник
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/services/api'

const authStore = useAuthStore()
const chats = ref([])
const searchQuery = ref('')
const loading = ref(false)
const joinRequestStatus = ref({})

let searchTimeout

onMounted(async () => {
  await loadChats()
})

const loadChats = async () => {
  loading.value = true
  try {
    const response = await api.get('/chats/search', {
      params: { search: searchQuery.value }
    })
    chats.value = response.data
    await loadJoinRequestStatuses()
  } catch (error) {
    console.error('Ошибка загрузки чатов:', error)
  }
  loading.value = false
}

const loadJoinRequestStatuses = async () => {
  try {
    const response = await api.get('/user/chat-join-requests')
    const requests = response.data
    requests.forEach(request => {
      joinRequestStatus.value[request.chat_id] = request.status
    })
  } catch (error) {
    console.error('Ошибка загрузки статусов заявок:', error)
  }
}

const searchChats = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(loadChats, 300)
}

const isMember = (chat) => {
  return chat.members?.some(member => member.id === authStore.user?.id)
}

const getJoinButtonText = (chatId) => {
  const status = joinRequestStatus.value[chatId]
  if (status === 'pending') return 'Заявка отправлена'
  if (status === 'accepted') return 'Принят'
  if (status === 'rejected') return 'Отклонено'
  return 'Вступить'
}

const sendJoinRequest = async (chatId) => {
  try {
    joinRequestStatus.value[chatId] = 'pending'
    await api.post(`/chats/${chatId}/join-request`)
  } catch (error) {
    console.error('Ошибка отправки заявки:', error)
    delete joinRequestStatus.value[chatId]
  }
}
</script>