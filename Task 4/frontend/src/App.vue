<template>
  <div id="app">
    <nav v-if="isAuthenticated" class="navbar navbar-dark bg-dark">
      <div class="container">
        <a class="navbar-brand" href="/">WebChat</a>
        <div>
          <router-link to="/chats" class="btn btn-outline-light me-2">Чаты</router-link>
          <router-link to="/profile" class="btn btn-outline-light me-2">Профиль</router-link>
          <button @click="logout" class="btn btn-danger">Выйти</button>
        </div>
      </div>
    </nav>
    <router-view/>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from './stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const isAuthenticated = computed(() => authStore.token)

const logout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style>
#app {
  min-height: 100vh;
}
</style>