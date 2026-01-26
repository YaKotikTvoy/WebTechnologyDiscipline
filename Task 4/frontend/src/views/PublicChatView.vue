<template>
  <div class="container mt-3">
    <div class="card">
      <div class="card-header">
        <h5 class="mb-0">Публичные чаты</h5>
      </div>
      <div class="card-body">
        <div v-if="loading" class="text-center">
          <div class="spinner-border"></div>
        </div>
        <div v-else>
          <div class="row">
            <div v-for="chat in publicChats" :key="chat.id" class="col-md-4 mb-3">
              <div class="card h-100">
                <div class="card-body">
                  <h6 class="card-title">{{ chat.name }}</h6>
                  <p v-if="chat.description" class="card-text small text-muted">{{ chat.description }}</p>
                  <div class="mt-2">
                    <small>Участников: {{ chat.member_count }}</small>
                  </div>
                </div>
                <div class="card-footer">
                  <button @click="joinPublicChat(chat.id)" class="btn btn-sm btn-primary w-100">
                    Присоединиться
                  </button>
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
import { ref, onMounted } from 'vue'
import { useChatStore } from '../stores/chat'

const chatStore = useChatStore()
const publicChats = ref([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  publicChats.value = await chatStore.fetchPublicChats()
  loading.value = false
})

const joinPublicChat = (chatId) => {
  alert(`Вы присоединились к чату ${chatId}`)
}
</script>