<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Добавить друга</h4>
          </div>
          <div class="card-body">
            <form @submit.prevent="searchUser">
              <div class="input-group mb-3">
                <input
                  v-model="searchPhone"
                  type="text"
                  class="form-control"
                  placeholder="Введите номер телефона"
                  required
                />
                <button class="btn btn-primary" type="submit" :disabled="searching">
                  {{ searching ? 'Поиск...' : 'Найти' }}
                </button>
              </div>
            </form>

            <div v-if="searchResults.length > 0" class="mt-3">
              <div class="card">
                <div class="card-body">
                  <div
                    v-for="user in searchResults"
                    :key="user.id"
                    class="d-flex justify-content-between align-items-center"
                  >
                    <div>
                      <strong>{{ formatPhone(user.phone) }}</strong>
                      <div class="text-muted small">
                        {{ user.username || 'Без имени' }}
                      </div>
                    </div>
                    <button
                      @click="sendFriendRequest(user.phone)"
                      class="btn btn-sm btn-primary"
                      :disabled="sending"
                    >
                      Добавить в друзья
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="searchError" class="alert alert-danger mt-3">
              {{ searchError }}
            </div>
            <div v-if="successMessage" class="alert alert-success mt-3">
              {{ successMessage }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useFriendsStore } from '@/stores/friends'
import { normalizePhone, formatPhone } from '@/utils/phoneUtils'

const friendsStore = useFriendsStore()

const searchPhone = ref('')
const searchResults = ref([])
const searching = ref(false)
const sending = ref(false)
const searchError = ref('')
const successMessage = ref('')

const searchUser = async () => {
  searching.value = true
  searchError.value = ''
  successMessage.value = ''

  const normalizedPhone = normalizePhone(searchPhone.value)
  const result = await friendsStore.searchUser(normalizedPhone)
  
  if (result.success) {
    searchResults.value = friendsStore.searchResults
  } else {
    searchError.value = result.error || 'Пользователь не найден'
    searchResults.value = []
  }
  
  searching.value = false
}

const sendFriendRequest = async (phone) => {
  sending.value = true
  searchError.value = ''
  successMessage.value = ''

  const result = await friendsStore.sendFriendRequest(phone)
  
  if (result.success) {
    successMessage.value = 'Запрос в друзья отправлен'
    searchPhone.value = ''
    searchResults.value = []
  } else {
    searchError.value = result.error || 'Не удалось отправить запрос'
  }
  
  sending.value = false
}
</script>