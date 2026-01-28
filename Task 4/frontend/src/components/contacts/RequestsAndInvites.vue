<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-12">
        <h3>Запросы и приглашения</h3>
        
        <div v-if="infoNotifications.length > 0" class="card mb-4">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h5>Уведомления</h5>
            <button @click="clearInfoNotifications" class="btn btn-sm btn-outline-secondary">
              Очистить все
            </button>
          </div>
          <div class="card-body">
            <div class="list-group">
              <div
                v-for="notification in infoNotifications"
                :key="notification.id"
                class="list-group-item"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <div class="mb-1">{{ notification.data.message }}</div>
                    <div class="text-muted small">
                      {{ formatTime(notification.createdAt) }}
                    </div>
                  </div>
                  <button
                    @click="removeInfoNotification(notification.id)"
                    class="btn btn-close"
                  ></button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card mb-4">
          <div class="card-header">
            <h5>Запросы в друзья</h5>
          </div>
          <div class="card-body">
            <div v-if="friendRequests.length === 0" class="text-center py-3">
              Нет запросов в друзья
            </div>
            <div v-else class="list-group">
              <div
                v-for="request in friendRequests"
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
                      @click="respondToFriendRequest(request.id, 'accepted')"
                      class="btn btn-sm btn-success me-2"
                      :disabled="responding"
                    >
                      Принять
                    </button>
                    <button
                      @click="respondToFriendRequest(request.id, 'rejected')"
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

        <div class="card mb-4">
          <div class="card-header">
            <h5>Приглашения в чаты</h5>
          </div>
          <div class="card-body">
            <div v-if="chatInvites.length === 0" class="text-center py-3">
              Нет приглашений в чаты
            </div>
            <div v-else class="list-group">
              <div
                v-for="invite in chatInvites"
                :key="invite.id"
                class="list-group-item"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ invite.chat?.name || 'Без названия' }}</strong>
                    <div class="text-muted small">
                      {{ formatDate(invite.created_at) }}
                    </div>
                    <div class="small">
                      Пригласил: {{ invite.inviter?.username || invite.inviter?.phone }}
                    </div>
                  </div>
                  <div v-if="invite.status === 'pending'">
                    <button
                      @click="respondToChatInvite(invite.id, 'accepted')"
                      class="btn btn-sm btn-success me-2"
                      :disabled="responding"
                    >
                      Принять
                    </button>
                    <button
                      @click="respondToChatInvite(invite.id, 'rejected')"
                      class="btn btn-sm btn-danger"
                      :disabled="responding"
                    >
                      Отклонить
                    </button>
                  </div>
                  <div v-else>
                    <span :class="{
                      'badge bg-success': invite.status === 'accepted',
                      'badge bg-danger': invite.status === 'rejected'
                    }">
                      {{ invite.status === 'accepted' ? 'Принято' : 'Отклонено' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card" v-if="isAdminInAnyChat">
          <div class="card-header">
            <h5>Заявки на вступление в ваши чаты</h5>
          </div>
          <div class="card-body">
            <div v-if="chatJoinRequests.length === 0" class="text-center py-3">
              Нет заявок на вступление
            </div>
            <div v-else class="list-group">
              <div
                v-for="request in chatJoinRequests"
                :key="request.id"
                class="list-group-item"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ request.user?.username || request.user?.phone }}</strong>
                    <div class="text-muted small">
                      Хочет вступить в: {{ request.chat?.name || 'Чат' }}
                    </div>
                    <div class="text-muted small">
                      {{ formatDate(request.created_at) }}
                    </div>
                  </div>
                  <div v-if="request.status === 'pending'">
                    <button
                      @click="respondToChatJoinRequest(request.id, 'accepted')"
                      class="btn btn-sm btn-success me-2"
                      :disabled="responding"
                    >
                      Принять
                    </button>
                    <button
                      @click="respondToChatJoinRequest(request.id, 'rejected')"
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
import { onMounted, ref, computed, watch } from 'vue'
import { useFriendsStore } from '@/stores/friends'
import { useChatsStore } from '@/stores/chats'
import { useWebSocketStore } from '@/stores/ws'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/services/api'
import { formatPhone } from '@/utils/phoneUtils'

const friendsStore = useFriendsStore()
const chatsStore = useChatsStore()
const wsStore = useWebSocketStore()
const authStore = useAuthStore()

const friendRequests = ref([])
const chatInvites = ref([])
const chatJoinRequests = ref([])
const responding = ref(false)

const infoNotifications = computed(() => {
  return wsStore.notifications.filter(n => n.type === 'info')
})

const isAdminInAnyChat = computed(() => {
  return chatJoinRequests.value.length > 0
})

onMounted(async () => {
  await loadFriendRequests()
  await loadChatInvites()
  await loadChatJoinRequests()
})

watch(() => wsStore.notifications, () => {
  loadFriendRequests()
  loadChatInvites()
  loadChatJoinRequests()
}, { deep: true })

const loadFriendRequests = async () => {
  await friendsStore.fetchFriendRequests()
  friendRequests.value = friendsStore.friendRequests
}

const loadChatInvites = async () => {
  try {
    const response = await api.get('/chats/invites')
    chatInvites.value = response.data
  } catch (error) {
    console.error('Ошибка загрузки приглашений:', error)
  }
}

const loadChatJoinRequests = async () => {
  try {
    await chatsStore.fetchChats()
    
    const allRequests = []
    for (const chat of chatsStore.chats) {
      const members = chat.members || []
      const member = members.find(m => m.id === authStore.user?.id)
      
      if (member?.is_admin || chat.created_by === authStore.user?.id) {
        try {
          const response = await api.get(`/chats/${chat.id}/join-requests`)
          const requests = response.data.map(req => ({
            ...req,
            chat: chat
          }))
          allRequests.push(...requests)
        } catch (error) {
          console.error(`Ошибка загрузки заявок для чата ${chat.id}:`, error)
        }
      }
    }
    
    chatJoinRequests.value = allRequests
  } catch (error) {
    console.error('Ошибка загрузки заявок на вступление:', error)
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatTime = (dateString) => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now - date
  
  if (diff < 60000) return 'только что'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} мин назад`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} ч назад`
  
  return date.toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const respondToFriendRequest = async (requestId, status) => {
  responding.value = true
  await friendsStore.respondToRequest(requestId, status)
  await loadFriendRequests()
  wsStore.markNotificationAsReadByData('friend_request', requestId)
  responding.value = false
}

const respondToChatInvite = async (inviteId, status) => {
  responding.value = true
  try {
    await api.put(`/chats/invites/${inviteId}`, { status })
    await loadChatInvites()
    wsStore.markNotificationAsReadByData('chat_invite', inviteId)
    if (status === 'accepted') {
      await chatsStore.fetchChats()
    }
  } catch (error) {
    console.error('Ошибка обработки приглашения:', error)
  }
  responding.value = false
}

const respondToChatJoinRequest = async (requestId, status) => {
  responding.value = true
  try {
    await api.put(`/chat-join-requests/${requestId}`, { status })
    await loadChatJoinRequests()
    wsStore.markNotificationAsReadByData('chat_join_request', requestId)
    if (status === 'accepted') {
      await chatsStore.fetchChats()
    }
  } catch (error) {
    console.error('Ошибка обработки заявки:', error)
  }
  responding.value = false
}

const removeInfoNotification = (notificationId) => {
  wsStore.markAsRead(notificationId)
}

const clearInfoNotifications = () => {
  infoNotifications.value.forEach(notification => {
    wsStore.markAsRead(notification.id)
  })
}
</script>