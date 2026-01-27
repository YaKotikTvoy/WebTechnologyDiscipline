<template>
  <div class="container-fluid mt-3">
    <div class="row">
      <div class="col-md-4">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <h5 class="mb-0">Chats</h5>
            <button
              @click="showCreateChat = true"
              class="btn btn-sm btn-primary"
            >
              New Chat
            </button>
          </div>
          <div class="card-body p-0">
            <div v-if="chats.length === 0" class="text-center py-3">
              No chats yet
            </div>
            <div v-else class="list-group list-group-flush">
              <router-link
                v-for="chat in chats"
                :key="chat.id"
                :to="`/chats/${chat.id}`"
                class="list-group-item list-group-item-action"
                :class="{ active: chat.id === currentChatId }"
              >
                <div class="d-flex w-100 justify-content-between">
                  <h6 class="mb-1">
                    {{ chat.name || getChatName(chat) }}
                  </h6>
                  <small class="text-muted">
                    {{ formatLastMessageTime(chat) }}
                  </small>
                </div>
                <p class="mb-1 text-truncate">
                  {{ getLastMessage(chat) }}
                </p>
                <small v-if="chat.type === 'group'" class="text-muted">
                  Group • {{ chat.members.length }} members
                </small>
                <small v-else class="text-muted">
                  Private chat
                </small>
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <div class="col-md-8">
        <div class="card">
          <div class="card-body text-center py-5">
            <h5>Select a chat to start messaging</h5>
            <p class="text-muted">
              Choose a conversation from the list or create a new one
            </p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateChat" class="modal show d-block" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Create New Chat</h5>
            <button
              type="button"
              class="btn-close"
              @click="showCreateChat = false"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Chat Type</label>
              <select v-model="newChat.type" class="form-select">
                <option value="private">Private Chat</option>
                <option value="group">Group Chat</option>
              </select>
            </div>
            <div v-if="newChat.type === 'group'" class="mb-3">
              <label class="form-label">Group Name</label>
              <input
                v-model="newChat.name"
                type="text"
                class="form-control"
                placeholder="Enter group name"
              />
            </div>
            <div class="mb-3">
              <label class="form-label">
                {{ newChat.type === 'private' ? 'Friend Phone' : 'Friend Phones' }}
              </label>
              <input
                v-model="newChat.phoneInput"
                type="text"
                class="form-control"
                :placeholder="newChat.type === 'private' ? 'Enter friend phone' : 'Enter phone numbers separated by commas'"
                @keyup.enter="addPhone"
              />
              <small class="text-muted">
                {{ newChat.type === 'private' ? 'Enter phone number of your friend' : 'Enter phone numbers of friends to add to group' }}
              </small>
            </div>
            <div v-if="newChat.memberPhones.length > 0" class="mb-3">
              <div
                v-for="(phone, index) in newChat.memberPhones"
                :key="index"
                class="badge bg-primary me-2 mb-2"
              >
                {{ phone }}
                <button
                  type="button"
                  class="btn-close btn-close-white ms-1"
                  @click="removePhone(index)"
                ></button>
              </div>
            </div>
            <div v-if="error" class="alert alert-danger">
              {{ error }}
            </div>
          </div>
          <div class="modal-footer">
            <button
              type="button"
              class="btn btn-secondary"
              @click="showCreateChat = false"
            >
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="createChat"
              :disabled="creating"
            >
              {{ creating ? 'Creating...' : 'Create Chat' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const chatsStore = useChatsStore()
const authStore = useAuthStore()

const chats = ref([])
const showCreateChat = ref(false)
const creating = ref(false)
const error = ref('')

const currentChatId = computed(() => {
  return route.params.id ? parseInt(route.params.id) : null
})

const newChat = ref({
  type: 'private',
  name: '',
  phoneInput: '',
  memberPhones: []
})

onMounted(async () => {
  await chatsStore.fetchChats()
  chats.value = chatsStore.chats
})

const getChatName = (chat) => {
  if (chat.type === 'private') {
    const otherMember = chat.members.find(m => m.id !== authStore.user?.id)
    return otherMember ? otherMember.phone : 'Private Chat'
  }
  return chat.name || 'Group Chat'
}

const getLastMessage = (chat) => {
  return 'Start conversation...'
}

const formatLastMessageTime = (chat) => {
  return ''
}

const addPhone = () => {
  const phone = newChat.value.phoneInput.trim()
  if (phone && !newChat.value.memberPhones.includes(phone)) {
    newChat.value.memberPhones.push(phone)
    newChat.value.phoneInput = ''
  }
}

const removePhone = (index) => {
  newChat.value.memberPhones.splice(index, 1)
}

const createChat = async () => {
  creating.value = true
  error.value = ''

  if (newChat.value.type === 'private' && newChat.value.memberPhones.length !== 1) {
    error.value = 'Private chat requires exactly one friend'
    creating.value = false
    return
  }

  if (newChat.value.type === 'group' && newChat.value.memberPhones.length === 0) {
    error.value = 'Group chat requires at least one member'
    creating.value = false
    return
  }

  let result
  if (newChat.value.type === 'private') {
    result = await chatsStore.createPrivateChat(newChat.value.memberPhones[0])
  } else {
    result = await chatsStore.createGroupChat(newChat.value.name, newChat.value.memberPhones)
  }

  if (result.success) {
    showCreateChat.value = false
    newChat.value = {
      type: 'private',
      name: '',
      phoneInput: '',
      memberPhones: []
    }
    await chatsStore.fetchChats()
    chats.value = chatsStore.chats
    router.push(`/chats/${result.chat.id}`)
  } else {
    error.value = result.error
  }

  creating.value = false
}
</script>

<style scoped>
.list-group-item.active {
  background-color: #007bff;
  border-color: #007bff;
}
</style>