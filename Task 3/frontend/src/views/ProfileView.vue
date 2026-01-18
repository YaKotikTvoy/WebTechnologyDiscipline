<template>
  <main class="container py-4">
    <h1 class="mb-4">Профиль</h1>
    
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary"></div>
    </div>
    
    <div v-else-if="!user" class="alert alert-warning">
      Пожалуйста, войдите в систему
    </div>
    
    <div v-else class="row g-4">
      <div class="col-md-4">
        <div class="card">
          <div class="card-body text-center">
            <div class="mb-3">
              <i class="bi bi-person-circle display-1 text-primary"></i>
            </div>
            
            <h3>{{ user.username }}</h3>
            <p class="text-muted">{{ user.email }}</p>
            
            <span :class="['badge', roleBadgeClass]">
              {{ roleText }}
            </span>
            
            <div class="mt-4">
              <p class="text-muted">
                <i class="bi bi-calendar me-2"></i>
                Зарегистрирован: {{ formatDate(user.created_at) }}
              </p>
            </div>
            
            <button @click="logout" class="btn btn-outline-danger w-100 mt-3">
              Выйти
            </button>
          </div>
        </div>
      </div>
      
      <div class="col-md-8">
        <div class="card mb-4">
          <div class="card-body">
            <h5 class="mb-3">Быстрые действия</h5>
            <div class="d-flex flex-wrap gap-3">
              <router-link to="/cart" class="btn btn-outline-primary">
                Корзина
              </router-link>
              
              <!-- Показывать только если пользователь продавец -->
              <router-link v-if="isSeller" to="/seller" class="btn btn-outline-success">
                Панель продавца
              </router-link>
              
              <!-- Показывать только если пользователь администратор -->
              <router-link v-if="isAdmin" to="/admin" class="btn btn-outline-danger">
                Админ-панель
              </router-link>
            </div>
          </div>
        </div>
        
        <div class="card">
          <div class="card-body">
            <h5 class="mb-3">История заказов</h5>
            <div class="text-center py-4 text-muted">
              <i class="bi bi-receipt display-3 mb-3"></i>
              <p>Заказов пока нет</p>
              <router-link to="/products" class="btn btn-primary">
                К покупкам
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { auth, authState, apiRequest } from '@/utils/auth'

export default {
  name: 'ProfileView',
  setup() {
    const router = useRouter()
    const loading = ref(false)

    // Используем реактивное состояние
    const user = computed(() => authState.user)
    const isSeller = computed(() => auth.isSeller())
    const isAdmin = computed(() => auth.isAdmin())

    const roleText = computed(() => {
      const roles = {
        'customer': 'Покупатель',
        'seller': 'Продавец',
        'admin': 'Администратор'
      }
      return roles[user.value?.role] || user.value?.role
    })

    const roleBadgeClass = computed(() => {
      switch (user.value?.role) {
        case 'admin': return 'bg-danger'
        case 'seller': return 'bg-success'
        default: return 'bg-secondary'
      }
    })

    const fetchProfile = async () => {
      if (!auth.isAuthenticated()) return
      
      loading.value = true
      try {
        const data = await apiRequest('/api/profile')
        if (data.success) {
          // Обновляем и localStorage и реактивное состояние
          localStorage.setItem('user', JSON.stringify(data.data))
          auth.login(auth.getToken(), data.data) // Это обновит реактивное состояние
        }
      } catch (error) {
        console.error('Ошибка загрузки профиля:', error)
      } finally {
        loading.value = false
      }
    }

    const logout = () => {
      auth.logout()
      router.push('/login')
    }

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    }

    onMounted(() => {
      fetchProfile()
    })

    return {
      loading,
      user,
      isSeller,
      isAdmin,
      roleText,
      roleBadgeClass,
      logout,
      formatDate
    }
  }
}
</script>