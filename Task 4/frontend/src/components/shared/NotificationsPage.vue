<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-8">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h4>Уведомления</h4>
            <button
              @click="clearAll"
              class="btn btn-sm btn-outline-secondary"
              :disabled="notifications.length === 0"
            >
              Очистить все
            </button>
          </div>
          <div class="card-body">
            <div v-if="notifications.length === 0" class="text-center py-3">
              Нет уведомлений
            </div>
            <div v-else class="list-group">
              <div
                v-for="notification in sortedNotifications"
                :key="notification.id"
                class="list-group-item"
                :class="{ 'bg-light': notification.read }"
              >
                <div class="d-flex justify-content-between align-items-start">
                  <div class="flex-grow-1">
                    <div class="d-flex align-items-center mb-1">
                      <div class="me-2">
                        <i 
                          class="bi"
                          :class="{
                            'bi-person-plus-fill text-primary': notification.type === 'friend_request',
                            'bi-chat-left-fill text-success': notification.type === 'chat_invite',
                            'bi-chat-right-fill text-info': notification.type === 'chat_message'
                          }"
                        ></i>
                      </div>
                      <div>
                        <strong>{{ getNotificationTitle(notification) }}</strong>
                        <div class="text-muted small">
                          {{ formatTime(notification.createdAt) }}
                        </div>
                      </div>
                    </div>
                    
                    <div class="mt-2">
                      <div v-if="notification.type === 'friend_request'">
                        <p class="mb-1">
                          {{ notification.data.sender.username || notification.data.sender.phone }} хочет добавить вас в друзья
                        </p>
                        <div class="d-flex gap-2 mt-2">
                          <button
                            @click="acceptFriendRequest(notification.data.id)"
                            class="btn btn-sm btn-success"
                            :disabled="processing"
                          >
                            Принять
                          </button>
                          <button
                            @click="rejectFriendRequest(notification.data.id)"
                            class="btn btn-sm btn-danger"
                            :disabled="processing"
                          >
                            Отклонить
                          </button>
                        </div>
                      </div>
                      
                      <div v-if="notification.type === 'chat_invite'">
                        <p class="mb-1">
                          Вас пригласили в чат "{{ notification.data.name || 'Без названия' }}"
                        </p>
                        <div class="d-flex gap-2 mt-2">
                          <button
                            @click="joinChat(notification.data.id)"
                            class="btn btn-sm btn-primary"
                            :disabled="processing"
                          >
                            Присоединиться
                          </button>
                          <button
                            @click="declineChatInvite(notification.data.id)"
                            class="btn btn-sm btn-outline-secondary"
                            :disabled="processing"
                          >
                            Отклонить
                          </button>
                        </div>
                      </div>
                      
                      <div v-if="notification.type === 'chat_message'">
                        <p class="mb-1">
                          Новое сообщение в чате "{{ notification.data.chatName }}"
                        </p>
                        <div class="bg-white p-2 rounded border">
                          <small class="text-muted">
                            {{ notification.data.senderName }}:
                          </small>
                          <div class="mt-1">{{ notification.data.content }}</div>
                        </div>
                        <button
                          @click="goToChat(notification.data.chatId)"
                          class="btn btn-sm btn-outline-primary mt-2"
                        >
                          Перейти к чату
                        </button>
                      </div>
                    </div>
                  </div>
                  <button
                    @click="markAsRead(notification.id)"
                    class="btn btn-close"
                  ></button>
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
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocketStore } from '@/stores/ws'
import { useFriendsStore } from '@/stores/friends'
import { useChatsStore } from '@/stores/chats'
import { api } from '@/services/api'

const router = useRouter()
const wsStore = useWebSocketStore()
const friendsStore = useFriendsStore()
const chatsStore = useChatsStore()

const processing = ref(false)
const notifications = computed(() => wsStore.notifications)
const sortedNotifications = computed(() => {
  return [...notifications.value].sort((a, b) => 
    new Date(b.createdAt) - new Date(a.createdAt)
  )
})

const getNotificationTitle = (notification) => {
  switch (notification.type) {
    case 'friend_request':
      return 'Запрос в друзья'
    case 'chat_invite':
      return 'Приглашение в чат'
    case 'chat_message':
      return 'Новое сообщение'
    default:
      return 'Уведомление'
  }
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

const acceptFriendRequest = async (requestId) => {
  processing.value = true
  await friendsStore.respondToRequest(requestId, 'accepted')
  wsStore.markNotificationAsReadByData('friend_request', requestId)
  processing.value = false
}

const rejectFriendRequest = async (requestId) => {
  processing.value = true
  await friendsStore.respondToRequest(requestId, 'rejected')
  wsStore.markNotificationAsReadByData('friend_request', requestId)
  processing.value = false
}

const joinChat = async (chatId) => {
  processing.value = true
  try {
    await api.post(`/chats/${chatId}/join`)
    wsStore.markNotificationAsReadByData('chat_invite', chatId)
    await chatsStore.fetchChats()
    router.push(`/chats/${chatId}`)
  } catch (error) {
    console.error('Failed to join chat:', error)
  }
  processing.value = false
}

const declineChatInvite = async (chatId) => {
  wsStore.markNotificationAsReadByData('chat_invite', chatId)
}

const goToChat = (chatId) => {
  wsStore.markNotificationAsReadByData('chat_message', chatId)
  router.push(`/chats/${chatId}`)
}

const markAsRead = (notificationId) => {
  wsStore.markAsRead(notificationId)
}

const clearAll = () => {
  wsStore.clearNotifications()
}
</script>