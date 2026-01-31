<template>
  <div class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)" v-if="show">
    <div class="modal-dialog modal-dialog-centered modal-lg">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Поиск публичных групп</h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="mb-3">
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-search"></i></span>
              <input v-model="searchQuery" 
                     type="text" 
                     class="form-control" 
                     placeholder="Поиск по названию группы..."
                     @keyup.enter="searchGroups">
              <button class="btn btn-primary" @click="searchGroups">
                <i class="bi bi-search"></i> Найти
              </button>
            </div>
          </div>

          <div v-if="loading" class="text-center py-4">
            <div class="spinner-border spinner-border-sm" role="status">
              <span class="visually-hidden">Загрузка...</span>
            </div>
          </div>

          <div v-else-if="searchResults.length === 0 && searchQuery" class="text-center py-4">
            <i class="bi bi-search display-6 text-muted mb-3"></i>
            <p class="text-muted">Группы не найдены</p>
          </div>

          <div v-else-if="searchResults.length === 0 && !searchQuery" class="text-center py-4">
            <i class="bi bi-chat-square-text display-6 text-muted mb-3"></i>
            <p class="text-muted">Введите название группы для поиска</p>
          </div>

          <div v-else class="groups-list" style="max-height: 400px; overflow-y: auto;">
            <div v-for="group in searchResults" :key="group.id" 
                 class="group-item p-3 border-bottom">
              <div class="d-flex align-items-center justify-content-between">
                <div class="d-flex align-items-center">
                  <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                       style="width: 50px; height: 50px; font-size: 1.2rem;">
                    {{ getGroupInitial(group) }}
                  </div>
                  <div>
                    <h6 class="mb-0">{{ group.name }}</h6>
                    <div class="small text-muted">
                      {{ group.members?.length || 0 }} участников
                      <span v-if="group.created_by" class="ms-2">
                        Создатель: {{ getCreatorName(group) }}
                      </span>
                    </div>
                  </div>
                </div>
                
                <div>
                  <button v-if="isMember(group)" 
                          class="btn btn-outline-secondary" disabled>
                    <i class="bi bi-check-circle me-1"></i> Вы в группе
                  </button>
                  <button v-else 
                          class="btn btn-primary"
                          @click="joinGroup(group)"
                          :disabled="joiningGroupId === group.id">
                    <i class="bi bi-plus-circle me-1"></i> 
                    {{ joiningGroupId === group.id ? 'Вступление...' : 'Вступить' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="error" class="alert alert-danger mt-3">{{ error }}</div>
          <div v-if="success" class="alert alert-success mt-3">{{ success }}</div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="$emit('close')">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useChatsStore } from '@/stores/chats'

const router = useRouter()
const authStore = useAuthStore()
const chatsStore = useChatsStore()

const emit = defineEmits(['close', 'joined'])

const props = defineProps({
  show: Boolean
})

const searchQuery = ref('')
const searchResults = ref([])
const loading = ref(false)
const joiningGroupId = ref(null)
const error = ref('')
const success = ref('')

const searchGroups = async () => {
  if (!searchQuery.value.trim()) return
  
  loading.value = true
  error.value = ''
  searchResults.value = []
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.get(`/chats/search?search=${encodeURIComponent(searchQuery.value.trim())}`)
    
    searchResults.value = response.data || []
  } catch (err) {
    error.value = 'Ошибка поиска групп'
  } finally {
    loading.value = false
  }
}

const getGroupInitial = (group) => {
  if (!group.name) return 'Г'
  return group.name.charAt(0).toUpperCase()
}

const getCreatorName = (group) => {
  if (!group.created_by) return 'Неизвестно'
  
  const creator = group.members?.find(m => m.id === group.created_by)
  if (creator) {
    return creator.username || creator.phone
  }
  return 'Создатель'
}

const isMember = (group) => {
  const userId = authStore.user?.id
  if (!userId) return false
  
  return group.members?.some(member => member.id === userId) || false
}

const joinGroup = async (group) => {
  joiningGroupId.value = group.id
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.post(`/chats/${group.id}/join-public`)
    
    if (response.data.success) {
      success.value = `Вы вступили в группу "${group.name}"`
      
      const groupIndex = searchResults.value.findIndex(g => g.id === group.id)
      if (groupIndex !== -1) {
        searchResults.value[groupIndex].members = [
          ...(searchResults.value[groupIndex].members || []),
          {
            id: authStore.user?.id,
            username: authStore.user?.username,
            phone: authStore.user?.phone
          }
        ]
      }
      
      setTimeout(() => {
        emit('joined', group.id)
        emit('close')
      }, 1500)
    } else {
      error.value = response.data.error || 'Ошибка вступления в группу'
    }
  } catch (err) {
    error.value = err.response?.data?.error || 'Ошибка вступления в группу'
  } finally {
    joiningGroupId.value = null
  }
}

onMounted(() => {
  if (props.show) {
    searchQuery.value = ''
    searchResults.value = []
    error.value = ''
    success.value = ''
  }
})
</script>

<style scoped>
.group-item:hover {
  background-color: #f8f9fa;
}

.groups-list {
  border: 1px solid #dee2e6;
  border-radius: 0.375rem;
}
</style>