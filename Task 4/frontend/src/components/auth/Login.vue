<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Вход</h4>
          </div>
          <div class="card-body">
            <form @submit.prevent="handleLogin">
              <div class="mb-3">
                <label for="phone" class="form-label">Телефон</label>
                <input
                  v-model="form.phone"
                  type="tel"
                  class="form-control"
                  id="phone"
                  required
                  placeholder="7XXXXXXXXXX (11 цифр)"
                  pattern="7\d{10}"
                  title="Формат: 7XXXXXXXXXX (11 цифр, начинается с 7)"
                />
                <div class="form-text">Формат: 7XXXXXXXXXX</div>
                </div>
              <div class="mb-3">
                <label for="password" class="form-label">Пароль</label>
                <input
                  v-model="form.password"
                  type="password"
                  class="form-control"
                  id="password"
                  required
                  placeholder="Введите пароль"
                />
              </div>
              <div v-if="error" class="alert alert-danger">
                {{ error }}
              </div>
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? 'Вход...' : 'Войти' }}
              </button>
              <router-link to="/register" class="btn btn-link">
                Регистрация
              </router-link>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { normalizePhone } from '@/utils/phoneUtils'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  phone: '',
  password: ''
})
const loading = ref(false)
const error = ref('')

const handleLogin = async () => {
  loading.value = true
  error.value = ''

  const normalizedPhone = normalizePhone(form.value.phone)
  
  const result = await authStore.login(normalizedPhone, form.value.password)
  
  if (result.success) {
    router.push('/')
  } else {
    error.value = result.error
  }
  
  loading.value = false
}
</script>