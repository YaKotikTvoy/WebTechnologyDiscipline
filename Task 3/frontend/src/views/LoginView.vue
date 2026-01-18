<template>
  <main class="container py-5">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card">
          <div class="card-body p-4">
            <h2 class="text-center mb-4">Вход</h2>
            
            <div v-if="error" class="alert alert-danger">
              {{ error }}
            </div>
            
            <form @submit.prevent="handleLogin">
              <div class="mb-3">
                <label for="username" class="form-label">Логин или Email</label>
                <input type="text" class="form-control" id="username" v-model="form.username" required>
              </div>
              
              <div class="mb-3">
                <label for="password" class="form-label">Пароль</label>
                <input type="password" class="form-control" id="password" v-model="form.password" required>
              </div>
              
              <button type="submit" class="btn btn-primary w-100" :disabled="loading">
                <span v-if="loading" class="spinner-border spinner-border-sm me-2"></span>
                Войти
              </button>
            </form>
            
            <div class="text-center mt-3">
              <router-link to="/register" class="text-decoration-none">
                Нет аккаунта? Зарегистрироваться
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
import { useStore } from 'vuex'

export default {
  name: 'LoginView',
  setup() {
    const router = useRouter()
    const store = useStore()
    const form = ref({
      username: '',
      password: ''
    })
    const loading = ref(false)
    const error = ref('')

    const handleLogin = async () => {
      loading.value = true
      error.value = ''
      
      try {
        const result = await store.dispatch('login', form.value)
        
        if (result.success) {
          const token = result.data.data.token
          localStorage.setItem('token', token)
          window.dispatchEvent(new CustomEvent('auth-change'))
          router.push('/')
        } else {
          error.value = result.error
        }
      } catch (err) {
        error.value = 'Ошибка соединения'
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      loading,
      error,
      handleLogin
    }
  }
}
</script>