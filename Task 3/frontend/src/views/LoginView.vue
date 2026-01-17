<template>
  <main class="container">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card shadow">
          <div class="card-body p-4">
            <h2 class="text-center mb-4">Вход в систему</h2>
            
            <div v-if="error" class="alert alert-danger alert-dismissible fade show" role="alert">
              {{ error }}
              <button type="button" class="btn-close" @click="error = ''"></button>
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
              <p class="mb-0">
                Нет аккаунта? 
                <router-link to="/register" class="text-decoration-none">Зарегистрироваться</router-link>
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { mapActions } from 'vuex'

export default {
  name: 'LoginView',
  data() {
    return {
      form: {
        username: '',
        password: ''
      },
      loading: false,
      error: ''
    }
  },
  methods: {
    ...mapActions(['login']),
    
    async handleLogin() {
      this.loading = true
      this.error = ''
      
      const result = await this.login(this.form)
      
      if (result.success) {
        this.$router.push('/')
      } else {
        this.error = result.error
      }
      
      this.loading = false
    }
  }
}
</script>