<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-4">
        <div class="card">
          <div class="card-body">
            <h2 class="text-center mb-4">Вход</h2>
            <form @submit.prevent="handleLogin">
              <div class="mb-3">
                <label class="form-label">Телефон (+7XXXXXXXXXX)</label>
                <input v-model="phone" type="tel" class="form-control" placeholder="+79123456789" required>
              </div>
              <div class="mb-3">
                <label class="form-label">Пароль</label>
                <input v-model="password" type="password" class="form-control" required>
              </div>
              <div class="d-grid gap-2">
                <button type="submit" class="btn btn-primary" :disabled="loading">
                  <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
                  Войти
                </button>
              </div>
              <div class="mt-3 text-center">
                <router-link to="/register">Нет аккаунта? Зарегистрируйтесь</router-link>
              </div>
              <div class="mt-2 text-center">
                <router-link to="/public-chat" class="btn btn-link">Войти как гость</router-link>
              </div>
            </form>
            <div v-if="error" class="alert alert-danger mt-3">
              {{ error }}
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

const phone = ref('')
const password = ref('')
const loading = ref(false)
const error = ref('')
const router = useRouter()
const authStore = useAuthStore()

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  const result = await authStore.login(phone.value, password.value)
  
  if (result.success) {
    router.push('/chats')
  } else {
    error.value = result.error
  }
  
  loading.value = false
}
</script>