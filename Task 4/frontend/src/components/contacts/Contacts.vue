<template>
  <div class="container mt-4">
    <div class="row">
      <div class="col-md-8 mx-auto">
        <div class="card">
          <div class="card-header">
            <div class="d-flex align-items-center">
              <button class="btn btn-sm btn-outline-secondary me-2" @click="$router.push('/')">
                <i class="bi bi-arrow-left"></i>
              </button>
              <h5 class="mb-0">Контакты</h5>
              <div class="ms-auto">
                <button class="btn btn-sm btn-primary" @click="addContact">
                  <i class="bi bi-person-plus"></i> Добавить
                </button>
              </div>
            </div>
          </div>
          <div class="card-body">
            <div class="input-group mb-3">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input type="text" 
                     v-model="searchQuery" 
                     class="form-control" 
                     placeholder="Поиск контактов">
            </div>

            <div v-if="friends.length === 0" class="text-center py-4">
              <i class="bi bi-people display-1 text-muted mb-3"></i>
              <p>Контактов пока нет</p>
            </div>
            
            <div v-else class="list-group">
              <div v-for="friend in filteredFriends" :key="friend.id" 
                   class="list-group-item list-group-item-action">
                <div class="d-flex align-items-center">
                  <div class="rounded-circle bg-secondary text-white d-flex align-items-center justify-content-center me-3" 
                       style="width: 50px; height: 50px;">
                    {{ getFriendInitial(friend.friend) }}
                  </div>
                  <div class="flex-grow-1">
                    <div class="fw-bold">{{ friend.friend.username || friend.friend.phone }}</div>
                    <div class="text-muted small">{{ friend.friend.phone }}</div>
                  </div>
                  <button class="btn btn-sm btn-outline-primary" 
                          @click="startChat(friend.friend.phone)">
                    <i class="bi bi-chat"></i>
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useFriendsStore } from '@/stores/friends'
import { useChatsStore } from '@/stores/chats'

const router = useRouter()
const friendsStore = useFriendsStore()
const chatsStore = useChatsStore()

const friends = ref([])
const searchQuery = ref('')

onMounted(async () => {
  await loadFriends()
})

const loadFriends = async () => {
  await friendsStore.fetchFriends()
  friends.value = friendsStore.friends
}

const filteredFriends = computed(() => {
  if (!searchQuery.value) return friends.value
  const query = searchQuery.value.toLowerCase()
  return friends.value.filter(friend => {
    const username = friend.friend.username?.toLowerCase() || ''
    const phone = friend.friend.phone?.toLowerCase() || ''
    return username.includes(query) || phone.includes(query)
  })
})

const startChat = async (phone) => {
  const result = await chatsStore.createPrivateChat(phone)
  if (result.success && result.chat) {
    router.push(`/chats/${result.chat.id}`)
  }
}

const addContact = () => {
  const phone = prompt('Введите номер телефона:')
  if (phone) {
    friendsStore.sendFriendRequest(phone).then(() => {
      loadFriends()
    })
  }
}

const getFriendInitial = (friend) => {
  if (!friend) return '?'
  if (friend.username) return friend.username.charAt(0).toUpperCase()
  return friend.phone ? friend.phone.slice(-1) : '?'
}
</script>