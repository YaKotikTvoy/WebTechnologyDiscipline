<template>
  <div class="container mt-3">
    <div class="row justify-content-center">
      <div class="col-md-8 col-lg-6">
        <div class="card">
          <div class="card-body">
            <h2 class="text-center mb-4">Профиль</h2>
            
            <div v-if="loading" class="text-center">
              <div class="spinner-border"></div>
            </div>
            
            <div v-else>
              <div class="text-center mb-4">
                <div v-if="userProfile.avatar_url" class="mb-3">
                  <img :src="userProfile.avatar_url" alt="Аватар" class="rounded-circle" style="width: 100px; height: 100px; object-fit: cover;">
                </div>
                <h4>{{ userProfile.username || 'Пользователь' }}</h4>
                <small class="text-muted">{{ maskPhone(userProfile.phone) }}</small>
              </div>
              
              <form @submit.prevent="updateProfile">
                <div class="mb-3">
                  <label class="form-label">Имя пользователя</label>
                  <input v-model="profileForm.username" class="form-control">
                </div>
                
                <div class="d-grid gap-2">
                  <button type="submit" class="btn btn-primary" :disabled="updating">
                    Сохранить
                  </button>
                </div>
              </form>
              
              <hr class="my-4">
              
              <div class="mb-3">
                <button @click="showDeleteModal = true" class="btn btn-danger w-100">
                  Удалить аккаунт
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <div v-if="showDeleteModal" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Удаление аккаунта</h5>
            <button @click="showDeleteModal = false" class="btn-close"></button>
          </div>
          <div class="modal-body">
            <div v-if="!deleteCodeSent">
              <p>Для удаления аккаунта введите email для получения кода подтверждения:</p>
              <input v-model="deleteEmail" type="email" class="form-control mb-3" placeholder="Email">
              <button @click="requestDeleteCode" class="btn btn-danger w-100">
                Отправить код
              </button>
            </div>
            
            <div v-else>
              <p>Код отправлен на {{ deleteEmail }}</p>
              <input v-model="deleteCode" type="text" class="form-control mb-3" placeholder="Код подтверждения">
              <button @click="confirmDelete" class="btn btn-danger w-100">
                Подтвердить удаление
              </button>
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
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const authStore = useAuthStore()
const router = useRouter()

const userProfile = ref({})
const profileForm = ref({})
const loading = ref(false)
const updating = ref(false)
const showDeleteModal = ref(false)
const deleteEmail = ref('')
const deleteCode = ref('')
const deleteCodeSent = ref(false)

const maskPhone = (phone) => {
  if (!phone) return ''
  return phone.replace(/(\d{4})\d+(\d{3})/, '$1***$2')
}

onMounted(async () => {
  loading.value = true
  try {
    const response = await api.get('/profile')
    userProfile.value = response.data
    profileForm.value = {
      username: userProfile.value.username || '',
      email: userProfile.value.email || ''
    }
  } catch (error) {
    console.error('Ошибка загрузки профиля:', error)
  }
  loading.value = false
})

const updateProfile = async () => {
  updating.value = true
  try {
    await api.put('/profile', profileForm.value)
    alert('Профиль обновлен')
  } catch (error) {
    console.error('Ошибка обновления профиля:', error)
  }
  updating.value = false
}

const requestDeleteCode = async () => {
  try {
    await api.post('/profile/delete-request', { email: deleteEmail.value })
    deleteCodeSent.value = true
  } catch (error) {
    alert(error.response?.data?.error || 'Ошибка запроса удаления')
  }
}

const confirmDelete = async () => {
  try {
    await api.post('/profile/confirm-delete', {
      email: deleteEmail.value,
      code: deleteCode.value
    })
    
    authStore.logout()
    router.push('/login')
    alert('Аккаунт удален')
  } catch (error) {
    alert(error.response?.data?.error || 'Ошибка подтверждения удаления')
  }
}
</script>