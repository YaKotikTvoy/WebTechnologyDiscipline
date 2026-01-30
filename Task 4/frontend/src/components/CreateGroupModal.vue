<template>
  <div class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)" v-if="show">
    <div class="modal-dialog modal-dialog-centered modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Создать группу</h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="mb-3">
            <label class="form-label">Название группы</label>
            <input v-model="groupName" type="text" class="form-control" placeholder="Введите название группы">
          </div>
          
          <div class="mb-3">
            <label class="form-label">Видимость группы</label>
            <div class="form-check">
              <input class="form-check-input" type="radio" v-model="isSearchable" :value="true" id="visible-true">
              <label class="form-check-label" for="visible-true">
                Публичная (видимая в поиске)
              </label>
            </div>
            <div class="form-check">
              <input class="form-check-input" type="radio" v-model="isSearchable" :value="false" id="visible-false">
              <label class="form-check-label" for="visible-false">
                Скрытая (только по приглашению)
              </label>
            </div>
          </div>
          
          <div class="mb-3">
            <label class="form-label">Добавить из контактов</label>
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
                         :id="'contact-' + friend.id">
                </div>
                <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                     style="width: 40px; height: 40px;">
                  {{ getFriendInitial(friend) }}
                </div>
                <div class="flex-grow-1">
                  <div class="fw-bold">{{ friend.username || friend.phone }}</div>
                  <div class="small text-muted">{{ friend.phone }}</div>
                </div>
              </div>
              
              <div v-if="filteredContacts.length === 0" class="text-center py-3 text-muted">
                Контактов не найдено
              </div>
            </div>
          </div>
          
          <div class="mb-3">
            <label class="form-label">Или добавить по номеру телефона</label>
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
          <button type="button" class="btn btn-primary" @click="createGroup" :disabled="creating || !canCreate">
            {{ creating ? 'Создание...' : 'Создать' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'

const router = useRouter()
const emit = defineEmits(['close', 'created'])

const props = defineProps({
  show: Boolean
})

const chatsStore = useChatsStore()

const groupName = ref('')
const isSearchable = ref(true)
const contactSearch = ref('')
const newPhone = ref('')
const phones = ref([])
const selectedContacts = ref([])
const loadingContacts = ref(false)
const contacts = ref([])
const creating = ref(false)
const error = ref('')
const success = ref('')



const isPhoneInContacts = computed(() => {
  const phone = newPhone.value.trim()
  if (!phone) return false
  
  return contacts.value.some(contact => contact.phone === phone)
})

const findContactByPhone = (phone) => {
  return contacts.value.find(contact => contact.phone === phone)
}

const canCreate = computed(() => {
  return groupName.value.trim() && 
    (selectedContacts.value.length > 0 || phones.value.length > 0)
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
  await loadContacts()
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
  } catch (error) {
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
  return phoneRegex.test(phone)
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
  
  const existingContact = findContactByPhone(phone)
  if (existingContact) {
    if (!selectedContacts.value.includes(phone)) {
      selectedContacts.value.push(phone)
      success.value = `Контакт "${existingContact.username || existingContact.phone}" добавлен из списка`
    } else {
      error.value = 'Этот контакт уже выбран'
    }
    newPhone.value = ''
    return
  }
  
  if (phones.value.includes(phone) || selectedContacts.value.includes(phone)) {
    error.value = 'Этот номер уже добавлен'
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

const createGroup = async () => {
  if (!groupName.value.trim()) {
    error.value = 'Введите название группы'
    return
  }

  if (selectedContacts.value.length === 0 && phones.value.length === 0) {
    error.value = 'Добавьте хотя бы одного участника'
    return
  }

  creating.value = true
  error.value = ''
  success.value = ''

  try {
    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()
    const currentUserPhone = authStore.user?.phone
    
    const filteredPhones = phones.value.filter(phone => phone !== currentUserPhone)
    
    const allPhones = [
      ...filteredPhones,
      ...selectedContacts.value
    ].filter(Boolean)

    if (allPhones.length === 0) {
      error.value = 'Добавьте хотя бы одного другого участника'
      creating.value = false
      return
    }

    const result = await chatsStore.createGroupChat(
      groupName.value.trim(),
      allPhones,
      isSearchable.value
    )
    
    if (result.success) {
      success.value = 'Группа создана'
      
      setTimeout(() => {
        emit('close')
        
        if (result.chat && result.chat.id) {
          router.push(`/chats/${result.chat.id}`)
        }
        
        emit('created', result.chat)
      }, 1000)
    } else {
      error.value = result.error || 'Ошибка создания группы'
    }
  } catch (err) {
    error.value = 'Ошибка создания группы'
  } finally {
    creating.value = false
  }
}

watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

const resetForm = () => {
  groupName.value = ''
  isSearchable.value = true
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