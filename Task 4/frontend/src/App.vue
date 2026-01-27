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
              <router-link class="nav-link" to="/">Chats</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/friends">Friends</router-link>
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
                {{ authStore.user?.phone }}
              </a>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <button @click="logout" class="dropdown-item">
                    Logout
                  </button>
                </li>
                <li>
                  <button @click="logoutAll" class="dropdown-item">
                    Logout from all devices
                  </button>
                </li>
              </ul>
            </li>
          </ul>
          <ul v-else class="navbar-nav ms-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/login">Login</router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/register">Register</router-link>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <Notifications />
    
    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useWebSocketStore } from '@/stores/ws'
import Notifications from '@/components/shared/Notifications.vue'

const authStore = useAuthStore()
const wsStore = useWebSocketStore()

onMounted(() => {
  if (authStore.isAuthenticated) {
    authStore.fetchUser()
    wsStore.connect()
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