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
            
            <div v-if="searchResults.length > 0" class="mt-4">
              <h6>Найденные пользователи:</h6>
              <div class="list-group">
                <div v-for="user in searchResults" :key="user.id" class="list-group-item d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ user.username || 'Пользователь' }}</strong>
                    <br>
                    <small class="text-muted">{{ maskPhone(user.phone) }}</small>
                  </div>
                  <button @click="addContact(user.id)" class="btn btn-sm btn-primary">Добавить</button>
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
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../services/api'

const searchPhone = ref('')
const searchResults = ref([])
const error = ref('')
const success = ref('')

const maskPhone = (phone) => {
  if (!phone) return ''
  return phone.replace(/(\d{4})\d+(\d{3})/, '$1***$2')
}

const searchUser = async () => {
  error.value = ''
  success.value = ''
  
  try {
    const response = await api.get(`/users/search?phone=${encodeURIComponent(searchPhone.value)}`)
    searchResults.value = response.data
  } catch (err) {
    error.value = 'Пользователь не найден'
  }
}

const addContact = async (userId) => {
  try {
    await api.post('/contacts', { contact_id: userId })
    success.value = 'Контакт успешно добавлен'
    searchResults.value = []
    searchPhone.value = ''
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка добавления контакта'
  }
}
</script>