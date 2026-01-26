<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-4">
        <div class="card">
          <div class="card-body">
            <h2 class="text-center mb-4">Регистрация</h2>
            
            <div v-if="step === 1">
              <div class="mb-3">
                <label class="form-label">Телефон (+7XXXXXXXXXX)</label>
                <input v-model="phone" type="tel" class="form-control" placeholder="+79123456789" required>
              </div>
              <div class="d-grid gap-2">
                <button @click="sendCode" class="btn btn-primary" :disabled="loading">
                  <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
                  Получить код
                </button>
              </div>
            </div>

            <div v-else-if="step === 2">
              <div class="alert alert-info">
                Код отправлен на {{ phone }}
                <br>
                <small>Для тестирования код можно посмотреть в консоли сервера</small>
              </div>
              <div class="mb-3">
                <label class="form-label">Код из SMS</label>
                <input v-model="smsCode" type="text" class="form-control" maxlength="6" placeholder="123456">
              </div>
              <div class="d-grid gap-2 mb-2">
                <button @click="verifyCode" class="btn btn-primary" :disabled="loading">
                  Проверить код
                </button>
              </div>
              <button @click="step = 1" class="btn btn-link w-100">
                Изменить номер
              </button>
            </div>

            <div v-else>
              <div class="mb-3">
                <label class="form-label">Пароль</label>
                <input v-model="password" type="password" class="form-control" required minlength="6">
              </div>
              <div class="mb-3">
                <label class="form-label">Подтвердите пароль</label>
                <input v-model="confirmPassword" type="password" class="form-control" required>
              </div>
              <div class="d-grid gap-2">
                <button @click="completeRegistration" class="btn btn-success" :disabled="loading">
                  <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
                  Завершить регистрацию
                </button>
              </div>
            </div>

            <div v-if="error" class="alert alert-danger mt-3">
              {{ error }}
            </div>
            
            <div class="mt-3 text-center">
              <router-link to="/login">Уже есть аккаунт? Войдите</router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import api from '../services/api'

const phone = ref('')
const smsCode = ref('')
const password = ref('')
const confirmPassword = ref('')
const step = ref(1)
const loading = ref(false)
const error = ref('')
const router = useRouter()
const authStore = useAuthStore()

const sendCode = async () => {
  if (!phone.value.match(/^\+7\d{10}$/)) {
    error.value = 'Неверный формат телефона. Используйте +7XXXXXXXXXX'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    const response = await api.post('/send-registration-code', { phone: phone.value })
    step.value = 2
    console.log('Код для тестирования:', response.data.code)
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка отправки кода'
  }
  
  loading.value = false
}

const verifyCode = async () => {
  if (!smsCode.value.match(/^\d{6}$/)) {
    error.value = 'Код должен быть 6 цифр'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    await api.post('/verify-registration-code', {
      phone: phone.value,
      code: smsCode.value
    })
    step.value = 3
  } catch (err) {
    error.value = err.response?.data?.error || 'Неверный код'
  }
  
  loading.value = false
}

const completeRegistration = async () => {
  if (password.value !== confirmPassword.value) {
    error.value = 'Пароли не совпадают'
    return
  }
  
  if (password.value.length < 6) {
    error.value = 'Пароль должен быть не менее 6 символов'
    return
  }
  
  loading.value = true
  error.value = ''
  
  try {
    const response = await api.post('/register', {
      phone: phone.value,
      password: password.value,
      code: smsCode.value
    })
    
    authStore.setAuth(response.data.token, { 
      id: response.data.user_id, 
      role: response.data.role || 'user' 
    })
    
    router.push('/chats')
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка регистрации'
  }
  
  loading.value = false
}
</script>