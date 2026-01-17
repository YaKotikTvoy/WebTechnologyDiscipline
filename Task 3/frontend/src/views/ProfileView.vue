<template>
  <main class="container">
    <h1 class="mb-4">Профиль пользователя</h1>
    
    <div v-if="loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>
    
    <div v-else-if="!user" class="alert alert-warning">
      Пожалуйста, войдите в систему
    </div>
    
    <div v-else class="row">
      <!-- Информация о пользователе -->
      <div class="col-md-4 mb-4">
        <div class="card shadow">
          <div class="card-body text-center">
            <div class="mb-3">
              <i class="bi bi-person-circle display-1 text-primary"></i>
            </div>
            
            <h3>{{ user.username }}</h3>
            <p class="text-muted">{{ user.email }}</p>
            
            <span class="badge fs-6" :class="{
              'bg-secondary': user.role === 'customer',
              'bg-primary': user.role === 'seller',
              'bg-danger': user.role === 'admin'
            }">
              {{ roleText }}
            </span>
            
            <div class="mt-4">
              <p class="text-muted mb-1">
                <i class="bi bi-calendar me-2"></i>
                Зарегистрирован: {{ formatDate(user.created_at) }}
              </p>
            </div>
            
            <button @click="logout" class="btn btn-outline-danger mt-3 w-100">
              <i class="bi bi-box-arrow-right me-2"></i>Выйти
            </button>
          </div>
        </div>
      </div>
      
      <!-- Статистика и действия -->
      <div class="col-md-8">
        <!-- Быстрые действия -->
        <div class="card shadow mb-4">
          <div class="card-body">
            <h5 class="card-title">Быстрые действия</h5>
            <div class="row">
              <div class="col-md-4 mb-3">
                <router-link to="/cart" class="btn btn-outline-primary w-100">
                  <i class="bi bi-cart me-2"></i>Корзина
                  <span v-if="cartCount > 0" class="badge bg-danger ms-2">{{ cartCount }}</span>
                </router-link>
              </div>
              
              <div v-if="isSeller" class="col-md-4 mb-3">
                <router-link to="/seller" class="btn btn-outline-success w-100">
                  <i class="bi bi-shop me-2"></i>Панель продавца
                </router-link>
              </div>
              
              <div v-if="isAdmin" class="col-md-4 mb-3">
                <router-link to="/admin" class="btn btn-outline-danger w-100">
                  <i class="bi bi-gear me-2"></i>Админ-панель
                </router-link>
              </div>
            </div>
          </div>
        </div>
        
        <!-- История заказов (заглушка) -->
        <div class="card shadow">
          <div class="card-body">
            <h5 class="card-title">История заказов</h5>
            <div class="text-center py-4">
              <i class="bi bi-receipt display-3 text-muted mb-3"></i>
              <p class="text-muted">Заказов пока нет</p>
              <router-link to="/products" class="btn btn-primary">
                <i class="bi bi-arrow-right me-2"></i>К покупкам
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {
  name: 'ProfileView',
  data() {
    return {
      loading: false
    }
  },
  computed: {
    ...mapGetters([
      'getUser',
      'isSeller',
      'isAdmin',
      'getCartCount'
    ]),
    user() {
      return this.getUser
    },
    cartCount() {
      return this.getCartCount
    },
    roleText() {
      const roles = {
        'customer': 'Покупатель',
        'seller': 'Продавец',
        'admin': 'Администратор'
      }
      return roles[this.user?.role] || this.user?.role
    }
  },
  methods: {
    ...mapActions(['logout', 'fetchProfile']),
    
    async handleLogout() {
      await this.logout()
      this.$router.push('/')
    },
    
    formatDate(dateString) {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    }
  },
  async mounted() {
    this.loading = true
    await this.fetchProfile()
    this.loading = false
  }
}
</script>