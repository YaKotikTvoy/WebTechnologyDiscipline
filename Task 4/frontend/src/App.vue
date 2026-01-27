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
              <router-link class="nav-link" to="/">
                Чаты
                <span v-if="unreadChatsCount > 0" class="badge bg-danger ms-1">
                  {{ unreadChatsCount }}
                </span>
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/friends">Друзья</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/requests">
                Запросы и приглашения
                <span v-if="wsStore.pendingRequests > 0" class="badge bg-danger ms-1">
                  {{ wsStore.pendingRequests }}
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
import { onMounted, watch, computed, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import { useChatsStore } from '@/stores/chats'
import { useRouter } from 'vue-router'

const router = useRouter()
const authStore = useAuthStore()
const wsStore = useWebSocketStore()
const chatsStore = useChatsStore()

const unreadChatsCount = ref(0)

const calculateUnreadChats = () => {
  unreadChatsCount.value = chatsStore.chats.reduce((total, chat) => {
    return total + (chat.unreadCount || 0)
  }, 0)
}

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

watch(() => chatsStore.chats, () => {
  calculateUnreadChats()
}, { deep: true })

const logout = () => {
  authStore.logout()
  wsStore.disconnect()
}
</script>