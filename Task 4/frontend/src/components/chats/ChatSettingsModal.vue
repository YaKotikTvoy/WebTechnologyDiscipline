<template>
  <div v-if="show" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5)">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Настройки чата</h5>
          <button type="button" class="btn-close" @click="$emit('close')"></button>
        </div>
        <div class="modal-body">
          <div class="mb-4">
            <div class="d-flex align-items-center mb-3">
              <div class="rounded-circle d-flex align-items-center justify-content-center me-3" 
                   :class="chatColor"
                   style="width: 50px; height: 50px; font-size: 1.2rem;">
                {{ chatInitial }}
              </div>
                <div>
                    <h5 class="mb-0">{{ chatName }}</h5>
                    <div v-if="chatType !== 'private' && memberCount > 0" class="text-muted small">
                        {{ memberCount }} участников
                    </div>
                    <div class="text-muted small">{{ chatType === 'private' ? 'Приватный чат' : 'Групповой чат' }}</div>
                </div>
            </div>
          </div>

          <div v-if="chatType === 'private'" class="border-top pt-3">
            <h6 class="mb-3">Управление чатом</h6>
            
            <button class="btn btn-outline-danger w-100 mb-2" 
                    @click="confirmDeletePrivateChat">
              <i class="bi bi-trash me-2"></i> Удалить чат
            </button>
          </div>

          <div v-else class="border-top pt-3">
            <h6 class="mb-3">Управление группой</h6>
            
            <div v-if="isCreator" class="mb-3">
              <button class="btn btn-outline-danger w-100 mb-2" 
                      @click="confirmDeleteGroupChat">
                <i class="bi bi-trash me-2"></i> Удалить группу для себя
              </button>
              
              <button v-if="isCreator" 
                      class="btn btn-danger w-100 mb-2"
                      @click="confirmDeleteGroupForAll">
                <i class="bi bi-trash-fill me-2"></i> Удалить группу для всех
              </button>
            </div>
            
            <div v-else>
              <button class="btn btn-outline-danger w-100" 
                      @click="confirmLeaveGroup">
                <i class="bi bi-box-arrow-right me-2"></i> Покинуть группу
              </button>
            </div>
          </div>

          <div v-if="chatType === 'group'" class="border-top pt-3 mt-3">
            <h6 class="mb-3">Участники ({{ memberCount }})</h6>
            
            <div class="members-list" style="max-height: 200px; overflow-y: auto;">
              <div v-for="member in members" :key="member.id" 
                   class="d-flex align-items-center mb-2 p-2 border rounded">
                <div class="rounded-circle bg-primary text-white d-flex align-items-center justify-content-center me-3" 
                     style="width: 40px; height: 40px;">
                  {{ getMemberInitial(member) }}
                </div>
                <div class="flex-grow-1">
                  <div class="fw-bold">{{ member.username || member.phone }}</div>
                  <div class="small text-muted">{{ member.phone }}</div>
                </div>
                <div v-if="member.id === creatorId" class="badge bg-warning">Создатель</div>
                <div v-else-if="isCreator" class="small">
                  <button class="btn btn-sm btn-outline-danger" 
                          @click="confirmRemoveMember(member)"
                          title="Удалить из группы">
                    <i class="bi bi-person-dash"></i>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div v-if="error" class="alert alert-danger mx-3">
          {{ error }}
        </div>
        
        <div v-if="success" class="alert alert-success mx-3">
          {{ success }}
        </div>
        
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="$emit('close')">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useChatsStore } from '@/stores/chats'

const router = useRouter()
const emit = defineEmits(['close', 'deleted', 'left', 'member-removed'])

const props = defineProps({
  show: Boolean,
  chatId: Number,
  chatName: String,
  chatType: String,
  members: {
    type: Array,
    default: () => []
  },
  creatorId: Number,
  currentUserId: Number,
  chatColor: {
    type: String,
    default: 'bg-primary'
  }
})

const loading = ref(false)
const error = ref('')
const success = ref('')

const chatInitial = computed(() => {
  return props.chatName ? props.chatName.charAt(0).toUpperCase() : 'Ч'
})

const memberCount = computed(() => {
  return props.members.length
})

const isCreator = computed(() => {
  return props.currentUserId === props.creatorId
})

const getMemberInitial = (member) => {
  if (!member) return '?'
  if (member.username) return member.username.charAt(0).toUpperCase()
  return member.phone ? member.phone.slice(-1) : '?'
}

const confirmDeletePrivateChat = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.delete(`/chats/${props.chatId}?forAll=true`)
    
    if (response.data.success) {
      success.value = 'Чат удален'
      setTimeout(() => {
        emit('deleted', props.chatId)
        emit('close')
      }, 500)
    }
  } catch (err) {
    error.value = err.response?.data || 'Ошибка удаления чата'
  } finally {
    loading.value = false
  }
}


const confirmDeleteGroupChat = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.delete(`/chats/${props.chatId}`)
    
    if (response.data.success) {
      success.value = 'Вы покинули группу'
      setTimeout(() => {
        emit('left', props.chatId)
        emit('close')
      }, 1000)
    }
  } catch (err) {
    error.value = err.response?.data || 'Ошибка выхода из группы'
  } finally {
    loading.value = false
  }
}

const confirmDeleteGroupForAll = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.delete(`/chats/${props.chatId}?forAll=true`)
    
    if (response.data.success) {
      success.value = 'Группа удалена для всех участников'
      setTimeout(() => {
        emit('deleted', props.chatId)
        emit('close')
      }, 1000)
    }
  } catch (err) {
    error.value = err.response?.data || 'Ошибка удаления группы'
  } finally {
    loading.value = false
  }
}

const confirmLeaveGroup = async () => {
  loading.value = true
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.delete(`/chats/${props.chatId}`)
    
    if (response.data.success) {
      success.value = 'Вы покинули группу'
      setTimeout(() => {
        emit('left', props.chatId)
        emit('close')
      }, 1000)
    }
  } catch (err) {
    error.value = err.response?.data || 'Ошибка выхода из группы'
  } finally {
    loading.value = false
  }
}

const confirmRemoveMember = async (member) => {
  loading.value = true
  error.value = ''
  success.value = ''
  
  try {
    const { api } = await import('@/services/api')
    const response = await api.delete(`/chats/${props.chatId}/members/${member.id}`)
    
    if (response.data.success) {
      success.value = 'Участник удален'
      setTimeout(() => {
        emit('member-removed', member.id)
      }, 1000)
    }
  } catch (err) {
    error.value = err.response?.data || 'Ошибка удаления участника'
  } finally {
    loading.value = false
  }
}
</script>