<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-8">
        <div class="card">
          <div class="card-header">
            <h4>Friend Requests</h4>
          </div>
          <div class="card-body">
            <div v-if="requests.length === 0" class="text-center py-3">
              No friend requests
            </div>
            <div v-else class="list-group">
              <div
                v-for="request in requests"
                :key="request.id"
                class="list-group-item"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ request.sender.phone }}</strong>
                    <div class="text-muted small">
                      {{ request.sender.username }}
                    </div>
                    <small class="text-muted">
                      {{ formatDate(request.created_at) }}
                    </small>
                  </div>
                  <div v-if="request.status === 'pending'">
                    <button
                      @click="respondToRequest(request.id, 'accepted')"
                      class="btn btn-sm btn-success me-2"
                      :disabled="responding"
                    >
                      Accept
                    </button>
                    <button
                      @click="respondToRequest(request.id, 'rejected')"
                      class="btn btn-sm btn-danger"
                      :disabled="responding"
                    >
                      Reject
                    </button>
                  </div>
                  <div v-else>
                    <span :class="{
                      'badge bg-success': request.status === 'accepted',
                      'badge bg-danger': request.status === 'rejected'
                    }">
                      {{ request.status }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { useFriendsStore } from '@/stores/friends'

const friendsStore = useFriendsStore()

const requests = ref([])
const responding = ref(false)

onMounted(async () => {
  await friendsStore.fetchFriendRequests()
  requests.value = friendsStore.friendRequests
})

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString()
}

const respondToRequest = async (requestId, status) => {
  responding.value = true
  await friendsStore.respondToRequest(requestId, status)
  requests.value = friendsStore.friendRequests
  responding.value = false
}
</script>