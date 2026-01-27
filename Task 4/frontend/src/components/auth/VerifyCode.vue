<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Подтверждение регистрации</h4>
          </div>
          <div class="card-body">
            <form @submit.prevent="verifyCode">
              <div class="mb-3">
                <label class="form-label">Код подтверждения</label>
                <input
                  v-model="code"
                  type="text"
                  class="form-control"
                  placeholder="Введите 6-значный код"
                  maxlength="6"
                  required
                />
                <div class="form-text">
                  Код отправлен в консоль браузера. Проверьте Developer Tools (F12)
                </div>
              </div>
              
              <div v-if="error" class="alert alert-danger">
                {{ error }}
              </div>
              
              <div class="d-flex gap-2">
                <button type="submit" class="btn btn-primary" :disabled="loading">
                  {{ loading ? 'Проверка...' : 'Подтвердить' }}
                </button>
                <button type="button" class="btn btn-outline-secondary" @click="resendCode" :disabled="resending">
                  {{ resending ? 'Отправка...' : 'Отправить снова' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/services/api'

const router = useRouter()
const authStore = useAuthStore()

const code = ref('')
const loading = ref(false)
const resending = ref(false)
const error = ref('')
const phone = ref('')

onMounted(() => {
  phone.value = localStorage.getItem('pendingRegistration')
  if (!phone.value) {
    router.push('/register')
  }
})

const verifyCode = async () => {
  loading.value = true
  error.value = ''

  try {
    const response = await api.post('/auth/verify', {
      phone: phone.value,
      code: code.value
    })
    
    authStore.token = response.data.token
    authStore.isAuthenticated = true
    localStorage.setItem('token', authStore.token)
    localStorage.removeItem('pendingRegistration')
    
    await authStore.fetchUser()
    router.push('/')
  } catch (err) {
    error.value = err.response?.data || 'Неверный код'
  }
  
  loading.value = false
}

const resendCode = async () => {
  resending.value = true
  error.value = ''
  
  try {
    await api.post('/auth/resend-code', { phone: phone.value })
    error.value = 'Новый код отправлен в консоль'
  } catch (err) {
    error.value = err.response?.data || 'Ошибка отправки кода'
  }
  
  resending.value = false
}
</script>