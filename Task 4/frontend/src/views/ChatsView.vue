<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-4">
        <div class="card">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center mb-3">
              <h5 class="mb-0">Чаты</h5>
              <button @click="showCreateModal = true" class="btn btn-sm btn-primary">
                Создать чат
              </button>
            </div>
            <div class="list-group">
              <router-link 
                v-for="chat in chatStore.chats" 
                :key="chat.id"
                :to="`/chat/${chat.id}`"
                class="list-group-item list-group-item-action"
                :class="{ 'active': currentChatId === chat.id }"
              >
                <div class="d-flex justify-content-between">
                  <strong>{{ chat.name }}</strong>
                  <span v-if="chat.unread_count" class="badge bg-danger">
                    {{ chat.unread_count }}
                  </span>
                </div>
                <small class="text-muted">
                  Участников: {{ chat.member_count }}
                </small>
              </router-link>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-8">
        <router-view/>
      </div>
    </div>

    <div v-if="showCreateModal" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Создать чат</h5>
            <button @click="showCreateModal = false" class="btn-close"></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="createChat">
              <div class="mb-3">
                <label class="form-label">Название чата</label>
                <input v-model="newChat.name" class="form-control" required>
              </div>
              <div class="mb-3">
                <label class="form-label">Описание</label>
                <textarea v-model="newChat.description" class="form-control"></textarea>
              </div>
              <div class="mb-3 form-check">
                <input v-model="newChat.is_public" type="checkbox" class="form-check-input">
                <label class="form-check-label">Публичный чат</label>
              </div>
              <div class="mb-3 form-check">
                <input v-model="newChat.only_admin_invite" type="checkbox" class="form-check-input">
                <label class="form-check-label">Только админы могут приглашать</label>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-primary" :disabled="creating">
                  Создать
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '../stores/chat'

const chatStore = useChatStore()
const route = useRoute()
const showCreateModal = ref(false)
const creating = ref(false)

const newChat = ref({
  name: '',
  description: '',
  is_public: false,
  only_admin_invite: false
})

const currentChatId = computed(() => route.params.id)

onMounted(() => {
  chatStore.fetchChats()
  setInterval(() => {
    chatStore.fetchChats()
  }, 30000)
})

const createChat = async () => {
  creating.value = true
  const result = await chatStore.createChat(newChat.value)
  
  if (result.success) {
    showCreateModal.value = false
    newChat.value = { name: '', description: '', is_public: false, only_admin_invite: false }
    chatStore.fetchChats()
  }
  
  creating.value = false
}
</script>