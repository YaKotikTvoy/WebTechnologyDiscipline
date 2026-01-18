<template>
  <nav class="navbar navbar-expand-lg navbar-light bg-light">
    <div class="container">
      <router-link class="navbar-brand" to="/">
        CatPC
      </router-link>
      
      <div class="navbar-nav">
        <router-link class="nav-link" to="/">Главная</router-link>
        <router-link class="nav-link" to="/products">Техника</router-link>
        <router-link class="nav-link" to="/cart">Корзина</router-link>
      </div>
      
      <div class="navbar-nav ms-auto">
        <div v-if="!isAuthenticated">
          <router-link class="btn btn-outline-primary me-2" to="/login">Вход</router-link>
          <router-link class="btn btn-primary" to="/register">Регистрация</router-link>
        </div>
        
        <div v-else class="dropdown">
          <button class="btn btn-outline-secondary dropdown-toggle" type="button" data-bs-toggle="dropdown">
            {{ user.username }}
          </button>
          <ul class="dropdown-menu">
            <li><router-link class="dropdown-item" to="/profile">Профиль</router-link></li>
            <li v-if="isSeller"><router-link class="dropdown-item" to="/seller">Продавец</router-link></li>
            <li v-if="isAdmin"><router-link class="dropdown-item" to="/admin">Админ</router-link></li>
            <li><hr class="dropdown-divider"></li>
            <li><button class="dropdown-item text-danger" @click="logout">Выйти</button></li>
          </ul>
        </div>
      </div>
    </div>
  </nav>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'HeaderComponent',
  setup() {
    const router = useRouter()
    const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

    const updateAuth = () => {
      user.value = JSON.parse(localStorage.getItem('user') || 'null')
    }

    onMounted(() => {
      window.addEventListener('auth-change', updateAuth)
    })

    onUnmounted(() => {
      window.removeEventListener('auth-change', updateAuth)
    })

    const isAuthenticated = computed(() => !!user.value)
    const isSeller = computed(() => user.value?.role === 'seller' || user.value?.role === 'admin')
    const isAdmin = computed(() => user.value?.role === 'admin')

    const logout = () => {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.dispatchEvent(new CustomEvent('auth-change'))
      router.push('/login')
    }

    return {
      user,
      isAuthenticated,
      isSeller,
      isAdmin,
      logout
    }
  }
}
</script>