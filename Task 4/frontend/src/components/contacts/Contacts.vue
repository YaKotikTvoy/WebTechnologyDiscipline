<template>
  <div class="h-100 d-flex flex-column">
    <div class="p-3 border-bottom bg-white">
      <div class="d-flex align-items-center">
        <button class="btn btn-link text-dark me-2" @click="$router.push('/')">
          <i class="bi bi-arrow-left"></i>
        </button>
        <h5 class="mb-0">Контакты</h5>
        <button class="btn btn-primary btn-sm ms-auto" @click="showAddModal">
          <i class="bi bi-person-plus"></i> Добавить
        </button>
      </div>
    </div>

    <div class="p-3 border-bottom bg-white">
      <div class="input-group">
        <span class="input-group-text"><i class="bi bi-search"></i></span>
        <input type="text" 
               v-model="searchQuery" 
               class="form-control" 
               placeholder="Поиск контактов">
      </div>
    </div>

    <div class="flex-grow-1 overflow-auto">
      <div v-if="loading" class="text-center py-4">
        <div class="spinner-border spinner-border-sm" role="status">
          <span class="visually-hidden">Загрузка...</span>
        </div>
      </div>
      
      <div v-else-if="filteredFriends.length === 0" class="text-center py-5">
        <i class="bi bi-people display-1 text-muted mb-3"></i>
        <p class="text-muted">Контактов пока нет</p>
      </div>
      
      <div v-else class="list-group">
        <div v-for="friend in filteredFriends" :key="friend.id" 
             class="list-group-item list-group-item-action border-0">
          <div class="d-flex align-items-center">
            <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                 style="width: 50px; height: 50px;">
              {{ getFriendInitial(friend.friend) }}
            </div>
            <div class="flex-grow-1">
              <div class="fw-bold">{{ friend.friend.username || friend.friend.phone }}</div>
              <div class="text-muted small">{{ friend.friend.phone }}</div>
            </div>
            <button class="btn btn-sm btn-outline-primary" 
                    @click="startChat(friend.friend.phone)">
              <i class="bi bi-chat"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Добавить контакт</h5>
            <button type="button" class="btn-close" @click="closeModal"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Номер телефона</label>
              <input v-model="newContactPhone" type="text" class="form-control" placeholder="7XXXXXXXXXX">
            </div>
            
            <div v-if="searching" class="text-center py-3">
              <div class="spinner-border spinner-border-sm" role="status">
                <span class="visually-hidden">Поиск...</span>
              </div>
            </div>
            
            <div v-if="searchResult" class="mb-3">
              <div class="card">
                <div class="card-body">
                  <div class="d-flex align-items-center">
                    <div class="rounded-circle bg-secondary text-white d-flex align-items-center justify-content-center me-3" 
                         style="width: 40px; height: 40px;">
                      {{ getFriendInitial(searchResult) }}
                    </div>
                    <div>
                      <div class="fw-bold">{{ searchResult.username || searchResult.phone }}</div>
                      <div class="text-muted small">{{ searchResult.phone }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <div v-if="searchError" class="alert alert-danger">{{ searchError }}</div>
            <div v-if="successMessage" class="alert alert-success">{{ successMessage }}</div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModal">Отмена</button>
            <button type="button" class="btn btn-primary" @click="searchUser" :disabled="!newContactPhone">
              Найти
            </button>
            <button v-if="searchResult" type="button" class="btn btn-success" @click="addContact" :disabled="adding">
              {{ adding ? 'Добавление...' : 'Добавить' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFriendsStore } from '@/stores/friends'
import { useChatsStore } from '@/stores/chats'

const router = useRouter()
const friendsStore = useFriendsStore()
const chatsStore = useChatsStore()

const friends = ref([])
const searchQuery = ref('')
const loading = ref(false)
const showModal = ref(false)
const newContactPhone = ref('')
const searching = ref(false)
const searchResult = ref(null)
const searchError = ref('')
const successMessage = ref('')
const adding = ref(false)

onMounted(async () => {
  await loadFriends()
})

const loadFriends = async () => {
  loading.value = true
  try {
    await friendsStore.fetchFriends()
    friends.value = friendsStore.friends
  } finally {
    loading.value = false
  }
}

const filteredFriends = computed(() => {
  if (!searchQuery.value) return friends.value
  const query = searchQuery.value.toLowerCase()
  return friends.value.filter(friend => {
    const username = friend.friend.username?.toLowerCase() || ''
    const phone = friend.friend.phone?.toLowerCase() || ''
    return username.includes(query) || phone.includes(query)
  })
})

const startChat = async (phone) => {
  const result = await chatsStore.createPrivateChat(phone)
  if (result.success && result.chat) {
    router.push(`/chats/${result.chat.id}`)
  }
}

const showAddModal = () => {
  showModal.value = true
  newContactPhone.value = ''
  searchResult.value = null
  searchError.value = ''
  successMessage.value = ''
}

const closeModal = () => {
  showModal.value = false
  newContactPhone.value = ''
  searchResult.value = null
  searchError.value = ''
  successMessage.value = ''
}

const searchUser = async () => {
  if (!newContactPhone.value) return
  
  searching.value = true
  searchError.value = ''
  searchResult.value = null
  
  try {
    const result = await friendsStore.searchUser(newContactPhone.value)
    
    if (result.success) {
      searchResult.value = friendsStore.searchResults[0] || null
      if (!searchResult.value) {
        searchError.value = 'Пользователь не найден'
      }
    } else {
      searchError.value = result.error || 'Пользователь не найден'
    }
  } catch (error) {
    searchError.value = 'Ошибка поиска'
  } finally {
    searching.value = false
  }
}

const addContact = async () => {
  if (!searchResult.value) return
  
  adding.value = true
  searchError.value = ''
  successMessage.value = ''
  
  try {
    const result = await friendsStore.sendFriendRequest(searchResult.value.phone)
    
    if (result.success) {
      successMessage.value = 'Запрос отправлен'
      await loadFriends()
      
      setTimeout(() => {
        closeModal()
      }, 1500)
    } else {
      searchError.value = result.error || 'Ошибка отправки запроса'
    }
  } catch (error) {
    searchError.value = 'Ошибка отправки запроса'
  } finally {
    adding.value = false
  }
}

const getFriendInitial = (friend) => {
  if (!friend) return '?'
  if (friend.username) return friend.username.charAt(0).toUpperCase()
  return friend.phone ? friend.phone.slice(-1) : '?'
}
</script>