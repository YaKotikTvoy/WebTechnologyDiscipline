<template>
  <div class="container mt-5">
    <div class="row justify-content-center">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Register</h4>
          </div>
          <div class="card-body">
            <form @submit.prevent="handleRegister">
              <div class="mb-3">
                <label for="phone" class="form-label">Phone</label>
                <input
                  v-model="form.phone"
                  type="text"
                  class="form-control"
                  id="phone"
                  required
                />
              </div>
              <div class="mb-3">
                <label for="password" class="form-label">Password</label>
                <input
                  v-model="form.password"
                  type="password"
                  class="form-control"
                  id="password"
                  required
                />
              </div>
              <div v-if="error" class="alert alert-danger">
                {{ error }}
              </div>
              <div v-if="success" class="alert alert-info">
                Check console for verification code
              </div>
              <button type="submit" class="btn btn-primary" :disabled="loading">
                {{ loading ? 'Registering...' : 'Register' }}
              </button>
              <router-link to="/login" class="btn btn-link">
                Login
              </router-link>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  phone: '',
  password: ''
})
const loading = ref(false)
const error = ref('')
const success = ref(false)

const handleRegister = async () => {
  loading.value = true
  error.value = ''
  success.value = false

  const result = await authStore.register(form.value.phone, form.value.password)
  
  if (result.success) {
    success.value = true
    console.log('Verification code for', form.value.phone, ': 123456')
    setTimeout(() => router.push('/'), 2000)
  } else {
    error.value = result.error
  }
  
  loading.value = false
}
</script>