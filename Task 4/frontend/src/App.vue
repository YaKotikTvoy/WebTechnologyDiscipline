<template>
  <div id="app">
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
      <div class="container-fluid">
        <router-link class="navbar-brand" to="/">WebChat</router-link>
        <button
          class="navbar-toggler"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#navbarNav"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul v-if="authStore.isAuthenticated" class="navbar-nav me-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/">Чаты</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/friends">Друзья</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/notifications">
                Уведомления
                <span v-if="wsStore.unreadNotifications > 0" class="badge bg-danger ms-1">
                  {{ wsStore.unreadNotifications }}
                </span>
              </router-link>
            </li>
          </ul>
          <ul v-if="authStore.isAuthenticated" class="navbar-nav ms-auto">
            <li class="nav-item dropdown">
              <a
                class="nav-link dropdown-toggle"
                href="#"
                role="button"
                data-bs-toggle="dropdown"
              >
                {{ authStore.user?.username || authStore.user?.phone }}
              </a>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <router-link to="/profile" class="dropdown-item">
                    Профиль
                  </router-link>
                </li>
                <li><hr class="dropdown-divider"></li>
                <li>
                  <button @click="logout" class="dropdown-item">
                    Выйти
                  </button>
                </li>
                <li>
                  <button @click="logoutAll" class="dropdown-item">
                    Выйти со всех устройств
                  </button>
                </li>
              </ul>
            </li>
          </ul>
          <ul v-else class="navbar-nav ms-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/login">Вход</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/register">Регистрация</router-link>
            </li>
          </ul>
        </div>
      </div>
    </nav>    
    <router-view />
  </div>
</template>

<script setup>
import { onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { useRouter } from 'vue-router'

const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()

onMounted(() => {
  if (authStore.isAuthenticated) {
    authStore.fetchUser()
    wsStore.connect()
  }
})

watch(() => authStore.isAuthenticated, (authenticated) => {
  if (authenticated) {
    wsStore.connect()
  } else {
    wsStore.disconnect()
    router.push('/login')
  }
})

const logout = () => {
  authStore.logout()
  wsStore.disconnect()
}

const logoutAll = () => {
  authStore.logoutAll()
  wsStore.disconnect()
}
</script>