<template>
  <div class="container mt-3">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Профиль</h4>
          </div>
          <div class="card-body">
            <div v-if="authStore.user" class="mb-4">
              <div class="row mb-3">
                <div class="col-md-4 fw-bold">Телефон:</div>
                <div class="col-md-8">{{ authStore.user.phone }}</div>
              </div>
              <div class="row mb-3">
                <div class="col-md-4 fw-bold">Имя пользователя:</div>
                <div class="col-md-8">
                  {{ authStore.user.username || 'Не установлено' }}
                </div>
              </div>
              <div class="row mb-3">
                <div class="col-md-4 fw-bold">Зарегистрирован:</div>
                <div class="col-md-8">
                  {{ formatDate(authStore.user.created_at) }}
                </div>
              </div>
            </div>

            <form @submit.prevent="updateProfile">
              <div class="mb-3">
                <label for="username" class="form-label">Имя пользователя</label>
                <input
                  v-model="form.username"
                  type="text"
                  class="form-control"
                  id="username"
                  placeholder="Введите отображаемое имя"
                  maxlength="50"
                />
                <div class="form-text">
                  Это имя будет отображаться вашим друзьям
                </div>
              </div>
              
              <div v-if="error" class="alert alert-danger">
                {{ error }}
              </div>
              <div v-if="success" class="alert alert-success">
                Профиль успешно обновлен
              </div>
              
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? 'Обновление...' : 'Обновить профиль' }}
              </button>
            </form>
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

const form = ref({
  username: ''
})
const loading = ref(false)
const error = ref('')
const success = ref(false)

onMounted(() => {
  if (authStore.user) {
    form.value.username = authStore.user.username || ''
  }
})

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const updateProfile = async () => {
  loading.value = true
  error.value = ''
  success.value = false

  try {
    await api.put('/auth/profile', {
      username: form.value.username.trim()
    })
    
    success.value = true
    await authStore.fetchUser()
    
    setTimeout(() => {
      success.value = false
    }, 3000)
  } catch (err) {
    error.value = err.response?.data || 'Не удалось обновить профиль'
  } finally {
    loading.value = false
  }
}
</script>