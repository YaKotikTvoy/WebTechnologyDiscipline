<template>
  <main class="container py-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card">
          <div class="card-body p-4">
            <h2 class="text-center mb-4">Регистрация</h2>
            
            <div v-if="error" class="alert alert-danger">
              {{ error }}
            </div>
            
            <div v-if="success" class="alert alert-success">
              {{ success }}
            </div>
            
            <form @submit.prevent="handleRegister">
              <div class="mb-3">
                <label for="username" class="form-label">Имя пользователя</label>
                <input type="text" class="form-control" id="username" v-model="form.username" required>
              </div>
              
              <div class="mb-3">
                <label for="email" class="form-label">Email</label>
                <input type="email" class="form-control" id="email" v-model="form.email" required>
              </div>
              
              <div class="mb-3">
                <label for="password" class="form-label">Пароль</label>
                <input type="password" class="form-control" id="password" v-model="form.password" required>
              </div>
              
              <div class="mb-3">
                <label for="confirmPassword" class="form-label">Подтверждение пароля</label>
                <input type="password" class="form-control" id="confirmPassword" v-model="form.confirmPassword" required>
              </div>
              
              <button type="submit" class="btn btn-primary w-100" :disabled="loading">
                <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
                Зарегистрироваться
              </button>
            </form>
            
            <div class="text-center mt-3">
              <router-link to="/login" class="text-decoration-none">
                Уже есть аккаунт? Войти
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from '@/utils/auth'

export default {
  name: 'RegisterView',
  setup() {
    const router = useRouter()
    const form = ref({
      username: '',
      email: '',
      password: '',
      confirmPassword: ''
    })
    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const handleRegister = async () => {
      if (form.value.password !== form.value.confirmPassword) {
        error.value = 'Пароли не совпадают'
        return
      }
      
      if (form.value.password.length < 6) {
        error.value = 'Пароль должен быть не менее 6 символов'
        return
      }
      
      loading.value = true
      error.value = ''
      success.value = ''
      
      try {
        const response = await fetch('http://localhost:1323/api/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            username: form.value.username,
            email: form.value.email,
            password: form.value.password
          })
        })
        
        const data = await response.json()
        
        if (data.success) {
          success.value = 'Регистрация успешна! Перенаправляем...'
          auth.login(data.data.token, data.data.user)
          
          setTimeout(() => {
            router.push('/')
          }, 1500)
        } else {
          error.value = data.error
        }
      } catch (err) {
        error.value = 'Ошибка регистрации'
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      loading,
      error,
      success,
      handleRegister
    }
  }
}
</script>