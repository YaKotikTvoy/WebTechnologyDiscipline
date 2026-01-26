<template>
  <div class="container mt-3">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h5 class="mb-0">Добавить контакт</h5>
          </div>
          <div class="card-body">
            <form @submit.prevent="searchUser">
              <div class="mb-3">
                <label class="form-label">Телефон пользователя</label>
                <div class="input-group">
                  <input v-model="searchPhone" type="tel" class="form-control" placeholder="+79123456789" required>
                  <button type="submit" class="btn btn-primary">Найти</button>
                </div>
              </div>
            </form>
            
            <div v-if="searchResult" class="mt-4">
              <h6>Найденный пользователь:</h6>
              <div class="card">
                <div class="card-body">
                  <div class="d-flex justify-content-between align-items-center">
                    <div>
                      <h6 class="mb-1">{{ searchResult.username || 'Пользователь' }}</h6>
                      <small class="text-muted">{{ maskPhone(searchResult.phone) }}</small>
                    </div>
                    <div class="d-flex gap-2">
                      <button @click="addContact(searchResult.id)" class="btn btn-sm btn-primary">
                        Добавить в контакты
                      </button>
                      <button @click="startChat(searchResult.id)" class="btn btn-sm btn-success">
                        Написать сообщение
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <div v-if="error" class="alert alert-danger mt-3">
              {{ error }}
            </div>
            
            <div v-if="success" class="alert alert-success mt-3">
              {{ success }}
            </div>
          </div>
        </div>
        
        <div class="card mt-3">
          <div class="card-header">
            <h5 class="mb-0">Мои контакты</h5>
          </div>
          <div class="card-body">
            <div v-if="loadingContacts" class="text-center">
              <div class="spinner-border spinner-border-sm"></div>
            </div>
            <div v-else-if="!contacts || contacts.length === 0" class="text-muted text-center">
              Контактов нет
            </div>
            <div v-else class="list-group">
              <div v-for="contact in contacts" :key="contact.id" 
                   class="list-group-item d-flex justify-content-between align-items-center">
                <div>
                  <strong>{{ contact.username || 'Пользователь' }}</strong>
                  <br>
                  <small class="text-muted">{{ maskPhone(contact.phone) }}</small>
                </div>
                <div class="d-flex gap-2">
                  <button @click="startChat(contact.id)" class="btn btn-sm btn-success">
                    Чат
                  </button>
                  <button @click="removeContact(contact.id)" class="btn btn-sm btn-danger">
                    Удалить
                  </button>
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
import { useRouter } from 'vue-router'
import api from '../services/api'

const router = useRouter()
const searchPhone = ref('')
const searchResult = ref(null)
const contacts = ref([])
const loadingContacts = ref(false)
const error = ref('')
const success = ref('')

const maskPhone = (phone) => {
  if (!phone) return ''
  return phone.replace(/(\d{4})\d+(\d{3})/, '$1***$2')
}

const loadContacts = async () => {
  loadingContacts.value = true
  try {
    const response = await api.get('/contacts')
    contacts.value = response.data
  } catch (err) {
    console.error('Ошибка загрузки контактов:', err)
  }
  loadingContacts.value = false
}

const searchUser = async () => {
  error.value = ''
  success.value = ''
  searchResult.value = null
  
  if (!searchPhone.value.trim()) {
    error.value = 'Введите номер телефона'
    return
  }
  
  try {
    const response = await api.get(`/users/search?phone=${encodeURIComponent(searchPhone.value)}`)
    searchResult.value = response.data
  } catch (err) {
    error.value = err.response?.data?.error || 'Пользователь не найден'
  }
}

const addContact = async (userId) => {
  try {
    await api.post('/contacts', { contact_id: userId })
    success.value = 'Контакт успешно добавлен'
    searchResult.value = null
    searchPhone.value = ''
    await loadContacts()
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка добавления контакта'
  }
}

const removeContact = async (userId) => {
  if (!confirm('Удалить контакт?')) return
  
  try {
    await api.delete(`/contacts/${userId}`)
    success.value = 'Контакт удален'
    await loadContacts()
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка удаления контакта'
  }
}

const startChat = async (contactId) => {
  try {
    const response = await api.post('/start-direct-chat', { contact_id: contactId })
    router.push(`/direct-chat/${contactId}`)
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка создания чата'
  }
}

onMounted(() => {
  loadContacts()
})
</script>