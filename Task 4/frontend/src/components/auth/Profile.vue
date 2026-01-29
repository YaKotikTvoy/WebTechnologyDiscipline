<template>
  <div class="container mt-4">
    <div class="row justify-content-center">
      <div class="col-md-8">
        <div class="card">
          <div class="card-header">
            <div class="d-flex align-items-center">
              <button class="btn btn-sm btn-outline-secondary me-2" @click="$router.push('/')">
                <i class="bi bi-arrow-left"></i>
              </button>
              <h5 class="mb-0">Профиль</h5>
            </div>
          </div>
          <div class="card-body">
            <div class="text-center mb-4">
              <div class="rounded-circle bg-primary text-white d-inline-flex align-items-center justify-content-center mb-3" 
                   style="width: 100px; height: 100px; font-size: 2.5rem;">
                {{ getUserInitial() }}
              </div>
              <h3>{{ authStore.user?.username || 'Без имени' }}</h3>
              <div class="text-muted">{{ authStore.user?.phone }}</div>
            </div>

            <form @submit.prevent="updateProfile">
              <div class="mb-3">
                <label class="form-label">Имя пользователя</label>
                <input v-model="username" type="text" class="form-control" placeholder="Введите имя">
              </div>
              <button type="submit" class="btn btn-primary w-100">Сохранить</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')

onMounted(() => {
  username.value = authStore.user?.username || ''
})

const getUserInitial = () => {
  if (!authStore.user) return '?'
  if (authStore.user.username) return authStore.user.username.charAt(0).toUpperCase()
  return authStore.user.phone ? authStore.user.phone.slice(-1) : '?'
}

const updateProfile = async () => {
  if (username.value.trim()) {
    await authStore.updateProfile(username.value.trim())
    alert('Профиль обновлен')
  }
}
</script>