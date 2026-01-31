<template>
  <div class="h-100 d-flex flex-column align-items-center justify-content-center bg-light">
    <div class="text-center">
      <i class="bi bi-chat-dots display-1 text-muted mb-3"></i>
      <h4 class="text-muted">Выберите чат</h4>
      <p class="text-muted">Выберите чат из списка слева или создайте новый</p>
      <button class="btn btn-primary" @click="openAddContactModal">
        <i class="bi bi-person-plus me-1"></i> Добавить контакт
      </button>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useChatsStore } from '@/stores/chats'

const chatsStore = useChatsStore()

const openAddContactModal = () => {
  const event = new CustomEvent('open-add-contact')
  window.dispatchEvent(event)
}

onMounted(() => {
  chatsStore.setActiveChat(null)
})
</script>