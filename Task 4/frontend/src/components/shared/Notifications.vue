<template>
  <div class="notifications">
    <div
      v-for="notification in notifications"
      :key="notification.id"
      class="notification alert"
      :class="{
        'alert-info': notification.type === 'friend_request',
        'alert-warning': notification.type === 'chat_invite'
      }"
      role="alert"
    >
      <div class="d-flex justify-content-between align-items-center">
        <div>
          <span v-if="notification.type === 'friend_request'">
            Новый запрос в друзья от {{ notification.data.sender.username || notification.data.sender.phone }}
          </span>
          <span v-if="notification.type === 'chat_invite'">
            Вас добавили в чат
          </span>
        </div>
        <button
          @click="markAsRead(notification.id)"
          class="btn-close"
          :disabled="notification.read"
        ></button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useWebSocketStore } from '@/stores/ws'

const wsStore = useWebSocketStore()

const notifications = computed(() => wsStore.notifications)

const markAsRead = (notificationId) => {
  wsStore.markAsRead(notificationId)
}
</script>

<style scoped>
.notifications {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 1000;
  max-width: 300px;
}

.notification {
  margin-bottom: 10px;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>