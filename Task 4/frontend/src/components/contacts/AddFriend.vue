<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-6">
        <div class="card">
          <div class="card-header">
            <h4>Add Friend</h4>
          </div>
          <div class="card-body">
            <form @submit.prevent="searchUser">
              <div class="input-group mb-3">
                <input
                  v-model="searchPhone"
                  type="text"
                  class="form-control"
                  placeholder="Enter phone number"
                  required
                />
                <button class="btn btn-primary" type="submit" :disabled="searching">
                  {{ searching ? 'Searching...' : 'Search' }}
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
                      <strong>{{ user.phone }}</strong>
                      <div class="text-muted small">
                        {{ user.username }}
                      </div>
                    </div>
                    <button
                      @click="sendFriendRequest(user.phone)"
                      class="btn btn-sm btn-primary"
                      :disabled="sending"
                    >
                      Add Friend
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

  const result = await friendsStore.searchUser(searchPhone.value)
  
  if (result.success) {
    searchResults.value = friendsStore.searchResults
  } else {
    searchError.value = result.error
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
    successMessage.value = 'Friend request sent successfully'
    searchPhone.value = ''
    searchResults.value = []
  } else {
    searchError.value = result.error
  }
  
  sending.value = false
}
</script>