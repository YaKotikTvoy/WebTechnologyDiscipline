<template>
  <div v-if="show" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Профиль</h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="text-center mb-4">
            <div class="rounded-circle bg-primary text-white d-inline-flex align-items-center justify-content-center mb-3" 
                 style="width: 80px; height: 80px; font-size: 2rem;">
              {{ getUserInitial() }}
            </div>
            <h4>{{ authStore.user?.username || 'Без имени' }}</h4>
            <div class="text-muted">{{ authStore.user?.phone }}</div>
          </div>

          <form @submit.prevent="save">
            <div class="mb-3">
              <label class="form-label">Имя пользователя</label>
              <input v-model="username" type="text" class="form-control" placeholder="Введите имя">
            </div>
            
            <div v-if="error" class="alert alert-danger">{{ error }}</div>
            <div v-if="success" class="alert alert-success">{{ success }}</div>
            
            <div class="d-flex justify-content-end gap-2">
              <button type="button" class="btn btn-secondary" @click="$emit('close')">Закрыть</button>
              <button type="submit" class="btn btn-primary" :disabled="saving">Сохранить</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const emit = defineEmits(['close'])
const authStore = useAuthStore()

const props = defineProps({
  show: Boolean
})

const username = ref('')
const saving = ref(false)
const error = ref('')
const success = ref('')

onMounted(() => {
  username.value = authStore.user?.username || ''
})

const getUserInitial = () => {
  if (!authStore.user) return '?'
  if (authStore.user.username) return authStore.user.username.charAt(0).toUpperCase()
  return authStore.user.phone ? authStore.user.phone.slice(-1) : '?'
}

const save = async () => {
  if (!username.value.trim()) {
    error.value = 'Введите имя пользователя'
    return
  }
  
  saving.value = true
  error.value = ''
  success.value = ''
  
  try {
    const result = await authStore.updateProfile(username.value.trim())
    
    if (result.success) {
      success.value = 'Профиль обновлен'
      setTimeout(() => {
        emit('close')
      }, 1000)
    } else {
      error.value = result.error || 'Ошибка обновления профиля'
    }
  } catch {
    error.value = 'Ошибка обновления профиля'
  } finally {
    saving.value = false
  }
}
</script>