<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-8">
        <div class="card">
          <div class="card-header">
            <h4>Запросы в друзья</h4>
          </div>
          <div class="card-body">
            <div v-if="requests.length === 0" class="text-center py-3">
              Нет запросов в друзья
            </div>
            <div v-else class="list-group">
              <div
                v-for="request in requests"
                :key="request.id"
                class="list-group-item"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ request.sender.username || formatPhone(request.sender.phone) }}</strong>
                    <div class="text-muted small">
                      {{ formatDate(request.created_at) }}
                    </div>
                  </div>
                  <div v-if="request.status === 'pending'">
                    <button
                      @click="respondToRequest(request.id, 'accepted')"
                      class="btn btn-sm btn-success me-2"
                      :disabled="responding"
                    >
                      Принять
                    </button>
                    <button
                      @click="respondToRequest(request.id, 'rejected')"
                      class="btn btn-sm btn-danger"
                      :disabled="responding"
                    >
                      Отклонить
                    </button>
                  </div>
                  <div v-else>
                    <span :class="{
                      'badge bg-success': request.status === 'accepted',
                      'badge bg-danger': request.status === 'rejected'
                    }">
                      {{ request.status === 'accepted' ? 'Принято' : 'Отклонено' }}
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
import { formatPhone } from '@/utils/phoneUtils'

const friendsStore = useFriendsStore()

const requests = ref([])
const responding = ref(false)

onMounted(async () => {
  await friendsStore.fetchFriendRequests()
  requests.value = friendsStore.friendRequests
})

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const respondToRequest = async (requestId, status) => {
  responding.value = true
  await friendsStore.respondToRequest(requestId, status)
  requests.value = friendsStore.friendRequests
  responding.value = false
}
</script>