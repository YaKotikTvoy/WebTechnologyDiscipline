<template>
  <main class="container">
    <div class="row justify-content-center">
      <div class="col-md-6 col-lg-5">
        <div class="card shadow">
          <div class="card-body p-4">
            <h2 class="text-center mb-4">Регистрация</h2>
            
            <div v-if="error" class="alert alert-danger alert-dismissible fade show" role="alert">
              {{ error }}
              <button type="button" class="btn-close" @click="error = ''"></button>
            </div>
            
            <div v-if="success" class="alert alert-success alert-dismissible fade show" role="alert">
              {{ success }}
              <button type="button" class="btn-close" @click="success = ''"></button>
            </div>
            
            <form @submit.prevent="handleRegister">
              <div class="mb-3">
                <label for="username" class="form-label">Имя пользователя</label>
                <input type="text" class="form-control" id="username" v-model="form.username" required>
                <div class="form-text">От 3 до 50 символов</div>
              </div>
              
              <div class="mb-3">
                <label for="email" class="form-label">Email</label>
                <input type="email" class="form-control" id="email" v-model="form.email" required>
              </div>
              
              <div class="mb-3">
                <label for="password" class="form-label">Пароль</label>
                <input type="password" class="form-control" id="password" v-model="form.password" required>
                <div class="form-text">Минимум 6 символов</div>
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
              <p class="mb-0">
                Уже есть аккаунт? 
                <router-link to="/login" class="text-decoration-none">Войти</router-link>
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
  name: 'RegisterView',
  data() {
    return {
      form: {
        username: '',
        email: '',
        password: '',
        confirmPassword: ''
      },
      loading: false,
      error: '',
      success: ''
    }
  },
  methods: {
    ...mapActions(['register']),
    
    async handleRegister() {
      // Валидация
      if (this.form.password !== this.form.confirmPassword) {
        this.error = 'Пароли не совпадают'
        return
      }
      
      if (this.form.password.length < 6) {
        this.error = 'Пароль должен быть не менее 6 символов'
        return
      }
      
      if (this.form.username.length < 3 || this.form.username.length > 50) {
        this.error = 'Имя пользователя должно быть от 3 до 50 символов'
        return
      }
      
      this.loading = true
      this.error = ''
      this.success = ''
      
      const result = await this.register({
        username: this.form.username,
        email: this.form.email,
        password: this.form.password
      })
      
      if (result.success) {
        this.success = 'Регистрация успешна! Вы будете перенаправлены на главную страницу.'
        setTimeout(() => {
          this.$router.push('/')
        }, 2000)
      } else {
        this.error = result.error
      }
      
      this.loading = false
    }
  }
}
</script>