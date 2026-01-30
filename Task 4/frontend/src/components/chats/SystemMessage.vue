<template>
  <div>
    <div>
      <i v-if="type === 'user_added'" class="bi bi-person-plus"></i>
      <i v-else-if="type === 'user_removed'" class="bi bi-person-dash"></i>
      <i v-else-if="type === 'chat_created'" class="bi bi-chat-left-dots"></i>
      <i v-else class="bi bi-info-circle"></i>
      {{ content }}
    </div>
    <div>{{ formatTime(timestamp) }}</div>
  </div>
</template>

<script setup>
const props = defineProps({
  content: {
    type: String,
    required: true
  },
  type: {
    type: String,
    default: 'info'
  },
  timestamp: {
    type: String,
    required: true
  }
})

const formatTime = (dateString) => {
  return new Date(dateString).toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}
</script>