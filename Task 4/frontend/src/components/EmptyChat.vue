<template>
  <div class="h-100 d-flex flex-column align-items-center justify-content-center bg-light">
    <div class="text-center">
      <i class="bi bi-chat-dots display-1 text-muted mb-3"></i>
      <h4 class="text-muted">Выберите чат</h4>
      <p class="text-muted">Выберите чат из списка слева или создайте новый</p>
      <button class="btn btn-primary mb-2" @click="openCreateChatModal">
        <i class="bi bi-plus-circle me-1"></i> Новый чат
      </button>
      <button class="btn btn-outline-primary" @click="openAddContactModal">
        <i class="bi bi-person-plus me-1"></i> Добавить контакт
      </button>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useChatsStore } from '@/stores/chats'

const chatsStore = useChatsStore()

const openCreateChatModal = () => {
  const event = new CustomEvent('open-create-chat')
  window.dispatchEvent(event)
}

const openAddContactModal = () => {
  const event = new CustomEvent('open-add-contact')
  window.dispatchEvent(event)
}

onMounted(() => {
  chatsStore.setActiveChat(null)
})
</script>