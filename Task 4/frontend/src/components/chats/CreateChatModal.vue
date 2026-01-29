<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h5>Новый чат</h5>
        <button @click="$emit('close')" class="btn-close"></button>
      </div>
      <div class="modal-body">
        <div class="modal-tabs">
          <button @click="chatType = 'private'" :class="{ active: chatType === 'private' }">
            Приватный чат
          </button>
          <button @click="chatType = 'group'" :class="{ active: chatType === 'group' }">
            Групповой чат
          </button>
        </div>
        
        <div v-if="chatType === 'group'" class="form-group">
          <input v-model="groupName" type="text" placeholder="Название группы">
        </div>
        
        <div class="form-group">
          <input v-model="phoneInput" type="text" placeholder="Номер телефона">
          <button @click="addPhone" class="btn-add">+</button>
        </div>
        
        <div v-if="phones.length > 0" class="phones-list">
          <div v-for="(phone, idx) in phones" :key="idx" class="phone-item">
            {{ phone }}
            <button @click="removePhone(idx)" class="btn-remove">×</button>
          </div>
        </div>
        
        <button @click="createChat" class="btn-create" :disabled="creating">
          {{ creating ? 'Создание...' : 'Создать' }}
        </button>
        
        <div v-if="error" class="error">{{ error }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useChatsStore } from '@/stores/chats'
import { useRouter } from 'vue-router'

const emit = defineEmits(['close', 'created'])

const chatsStore = useChatsStore()
const router = useRouter()

const chatType = ref('private')
const groupName = ref('')
const phoneInput = ref('')
const phones = ref([])
const creating = ref(false)
const error = ref('')

const addPhone = () => {
  const phone = phoneInput.value.trim()
  if (phone && !phones.value.includes(phone)) {
    phones.value.push(phone)
    phoneInput.value = ''
  }
}

const removePhone = (idx) => {
  phones.value.splice(idx, 1)
}

const createChat = async () => {
  creating.value = true
  error.value = ''
  
  try {
    let result
    if (chatType.value === 'private') {
      if (phones.value.length !== 1) {
        throw new Error('Приватный чат требует одного участника')
      }
      result = await chatsStore.createPrivateChat(phones.value[0])
    } else {
      if (!groupName.value.trim()) {
        throw new Error('Введите название группы')
      }
      result = await chatsStore.createGroupChat(groupName.value, phones.value, true)
    }
    
    if (result.success) {
      emit('created', result.chat)
      emit('close')
    } else {
      error.value = result.error
    }
  } catch (err) {
    error.value = err.message
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 10px;
  width: 400px;
  max-width: 90%;
}

.modal-header {
  padding: 15px;
  border-bottom: 1px solid #ddd;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-body {
  padding: 15px;
}

.modal-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.modal-tabs button {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
}

.modal-tabs button.active {
  background: #0088cc;
  color: white;
  border-color: #0088cc;
}

.form-group {
  display: flex;
  gap: 10px;
  margin-bottom: 15px;
}

.form-group input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 5px;
}

.btn-add {
  padding: 8px 15px;
  background: #6c757d;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.phones-list {
  margin-bottom: 15px;
}

.phone-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 10px;
  background: #f0f0f0;
  border-radius: 5px;
  margin-bottom: 5px;
}

.btn-remove {
  background: none;
  border: none;
  color: #dc3545;
  cursor: pointer;
  font-size: 18px;
}

.btn-create {
  width: 100%;
  padding: 10px;
  background: #0088cc;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.btn-create:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.error {
  color: #dc3545;
  margin-top: 10px;
  font-size: 14px;
}
</style>