<template>
  <div class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)" v-if="show">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Добавить пользователя в чат</h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="mb-3">
            <label class="form-label">Выберите из контактов</label>
            <div class="input-group mb-2">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input v-model="contactSearch" type="text" class="form-control" placeholder="Поиск контактов...">
            </div>
            
            <div v-if="loadingContacts" class="text-center py-3">
              <div class="spinner-border spinner-border-sm" role="status">
                <span class="visually-hidden">Загрузка...</span>
              </div>
            </div>
            
            <div v-else class="contacts-list" style="max-height: 200px; overflow-y: auto;">
              <div v-for="friend in filteredContacts" :key="friend.id" 
                   class="contact-item p-2 border-bottom d-flex align-items-center">
                <div class="form-check me-3">
                  <input class="form-check-input" type="checkbox" 
                         :value="friend.phone" 
                         v-model="selectedContacts"
                         :id="'contact-' + friend.id"
                         :disabled="isAlreadyInChat(friend.id)">
                </div>
                <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                     style="width: 40px; height: 40px;">
                  {{ getFriendInitial(friend) }}
                </div>
                <div class="flex-grow-1">
                  <div class="fw-bold">{{ friend.username || friend.phone }}</div>
                  <div class="small text-muted">{{ friend.phone }}</div>
                  <div v-if="isAlreadyInChat(friend.id)" class="small text-danger">
                    <i class="bi bi-exclamation-circle"></i> Уже в чате
                  </div>
                </div>
              </div>
              
              <div v-if="filteredContacts.length === 0" class="text-center py-3 text-muted">
                Контактов не найдено
              </div>
            </div>
          </div>
          
          <div class="mb-3">
            <label class="form-label">Или введите номер телефона</label>
            <div class="input-group">
              <input v-model="newPhone" type="text" class="form-control" placeholder="7XXXXXXXXXX" @keyup.enter="addPhone">
              <button class="btn btn-outline-secondary" type="button" @click="addPhone">
                <i class="bi bi-plus"></i> Добавить
              </button>
            </div>
            <div class="form-text">Формат: 7XXXXXXXXXX (11 цифр, начинается с 7)</div>
            
            <div v-if="phones.length > 0" class="mt-2">
              <div v-for="(phone, index) in phones" :key="index" class="badge bg-info me-2 mb-1">
                {{ phone }}
                <button type="button" class="btn-close btn-close-white ms-1" @click="removePhone(index)"></button>
              </div>
            </div>
          </div>
          
          <div v-if="error" class="alert alert-danger">{{ error }}</div>
          <div v-if="success" class="alert alert-success">{{ success }}</div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="$emit('close')">Отмена</button>
          <button type="button" class="btn btn-primary" @click="addUsers" :disabled="adding || !canAdd">
            {{ adding ? 'Добавление...' : 'Добавить' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useChatsStore } from '@/stores/chats'
import { useRouter } from 'vue-router'

const router = useRouter()
const emit = defineEmits(['close', 'added'])

const props = defineProps({
  show: Boolean,
  chatId: Number,
  currentMembers: {
    type: Array,
    default: () => []
  }
})

const chatsStore = useChatsStore()

const contactSearch = ref('')
const newPhone = ref('')
const phones = ref([])
const selectedContacts = ref([])
const loadingContacts = ref(false)
const contacts = ref([])
const adding = ref(false)
const error = ref('')
const success = ref('')

const canAdd = computed(() => {
  return selectedContacts.value.length > 0 || phones.value.length > 0
})

const filteredContacts = computed(() => {
  if (!contactSearch.value) return contacts.value
  
  const search = contactSearch.value.toLowerCase()
  return contacts.value.filter(friend => {
    const username = friend.username?.toLowerCase() || ''
    const phone = friend.phone?.toLowerCase() || ''
    return username.includes(search) || phone.includes(search)
  })
})

onMounted(async () => {
  if (props.show) {
    await loadContacts()
  }
})

const loadContacts = async () => {
  loadingContacts.value = true
  try {
    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()
    const currentUserId = authStore.user?.id
    
    if (!currentUserId) {
      contacts.value = []
      return
    }
    
    await chatsStore.fetchChats()
    
    const userContacts = []
    
    chatsStore.chats.forEach(chat => {
      if (chat.type === 'private') {
        chat.members?.forEach(member => {
          if (member.id && member.id !== currentUserId) {
            const exists = userContacts.some(c => c.id === member.id)
            if (!exists) {
              userContacts.push(member)
            }
          }
        })
      }
    })
    
    contacts.value = userContacts
  } catch {
    contacts.value = []
  } finally {
    loadingContacts.value = false
  }
}

const getFriendInitial = (friend) => {
  if (!friend) return '?'
  if (friend.username) return friend.username.charAt(0).toUpperCase()
  return friend.phone ? friend.phone.slice(-1) : '?'
}

const isValidPhone = (phone) => {
  const phoneRegex = /^7\d{10}$/
  return phoneRegex.test(phone.trim())
}

const isAlreadyInChat = (userId) => {
  return props.currentMembers.some(member => member.id === userId)
}

const findContactByPhone = (phone) => {
  const cleanPhone = phone.trim()
  return contacts.value.find(contact => contact.phone === cleanPhone)
}

const checkUserExists = async (phone) => {
  try {
    const { api } = await import('@/services/api')
    const response = await api.get(`/users/search?phone=${phone}`)
    return response.data && response.data.id
  } catch {
    return false
  }
}

const addPhone = async () => {
  const phone = newPhone.value.trim()
  
  if (!phone) {
    error.value = 'Введите номер телефона'
    return
  }
  
  if (!isValidPhone(phone)) {
    error.value = 'Неверный формат номера. Используйте: 7XXXXXXXXXX (11 цифр)'
    return
  }
  
  if (selectedContacts.value.includes(phone)) {
    error.value = 'Этот контакт уже выбран'
    newPhone.value = ''
    return
  }
  
  if (phones.value.includes(phone)) {
    error.value = 'Этот номер уже добавлен'
    return
  }
  
  const existingContact = findContactByPhone(phone)
  if (existingContact) {
    if (isAlreadyInChat(existingContact.id)) {
      error.value = 'Этот пользователь уже в чате'
      return
    }
    
    if (!selectedContacts.value.includes(phone)) {
      selectedContacts.value.push(phone)
      success.value = `Контакт "${existingContact.username || existingContact.phone}" добавлен`
    }
    newPhone.value = ''
    return
  }
  
  const exists = await checkUserExists(phone)
  if (exists) {
    phones.value.push(phone)
    newPhone.value = ''
    error.value = ''
    success.value = 'Номер добавлен'
  } else {
    error.value = 'Пользователь с таким номером не найден'
  }
}

const removePhone = (index) => {
  phones.value.splice(index, 1)
}

const addUsers = async () => {
  if (!props.chatId) {
    error.value = 'Ошибка: не указан чат'
    return
  }

  if (selectedContacts.value.length === 0 && phones.value.length === 0) {
    error.value = 'Выберите контакты или добавьте номера'
    return
  }

  adding.value = true
  error.value = ''
  success.value = ''

  try {
    const allPhones = [...selectedContacts.value, ...phones.value]
    
    let addedCount = 0
    for (const phone of allPhones) {
      const result = await chatsStore.addUserToChat(props.chatId, phone)
      if (result.success) {
        addedCount++
      } else {
        error.value = result.error || 'Ошибка добавления пользователя'
        adding.value = false
        return
      }
    }
    
    success.value = `Успешно добавлено ${addedCount} пользователей`
    
    await chatsStore.fetchChat(props.chatId)
    
    setTimeout(() => {
      emit('added')
      emit('close')
    }, 1000)
  } catch (err) {
    error.value = 'Ошибка добавления пользователей'
  } finally {
    adding.value = false
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
    loadContacts()
  }
})

const resetForm = () => {
  contactSearch.value = ''
  newPhone.value = ''
  phones.value = []
  selectedContacts.value = []
  error.value = ''
  success.value = ''
}
</script>

<style scoped>
.contact-item:hover {
  background-color: #f8f9fa;
}

.contacts-list {
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
}
</style>