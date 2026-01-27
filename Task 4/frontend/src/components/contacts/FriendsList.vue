<template>
  <div class="container mt-3">
    <div class="row">
      <div class="col-md-12">
        <div class="d-flex justify-content-between align-items-center mb-3">
          <h3>Друзья</h3>
          <div>
            <router-link to="/friends/add" class="btn btn-primary me-2">
              Добавить друга
            </router-link>
            <router-link to="/friends/requests" class="btn btn-outline-primary">
              Запросы в друзья
            </router-link>
          </div>
        </div>

        <div class="card">
          <div class="card-body">
            <div v-if="friends.length === 0" class="text-center py-3">
              Друзей пока нет
            </div>
            <div v-else class="list-group">
              <div
                v-for="friend in friends"
                :key="friend.id"
                class="list-group-item list-group-item-action"
              >
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{{ friend.friend.username || formatPhone(friend.friend.phone) }}</strong>
                  </div>
                  <div>
                    <button
                      @click="startChat(friend.friend.phone)"
                      class="btn btn-sm btn-outline-primary me-2"
                      :disabled="loading"
                    >
                      Написать
                    </button>
                    <button
                      @click="removeFriend(friend.friend.id)"
                      class="btn btn-sm btn-outline-danger"
                    >
                      Удалить
                    </button>
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
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useFriendsStore } from '@/stores/friends'
import { useChatsStore } from '@/stores/chats'
import { formatPhone } from '@/utils/phoneUtils'

const router = useRouter()
const friendsStore = useFriendsStore()
const chatsStore = useChatsStore()

const friends = ref([])
const loading = ref(false)

onMounted(async () => {
  await friendsStore.fetchFriends()
  friends.value = friendsStore.friends
})

const startChat = async (phone) => {
  loading.value = true
  const result = await chatsStore.createPrivateChat(phone)
  if (result.success) {
    router.push(`/chats/${result.chat.id}`)
  }
  loading.value = false
}

const removeFriend = async (friendId) => {
  if (confirm('Удалить из друзей?')) {
    await friendsStore.removeFriend(friendId)
    friends.value = friendsStore.friends
  }
}
</script>