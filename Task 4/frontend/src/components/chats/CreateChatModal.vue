<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal">
      <div class="modal-header">
        <h5>Новый приватный чат</h5>
        <button @click="$emit('close')" class="btn-close"></button>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <label class="form-label">Номер телефона</label>
          <input v-model="phoneInput" type="text" placeholder="7XXXXXXXXXX">
        </div>
        
        <button @click="createPrivateChat" class="btn-create" :disabled="creating">
          {{ creating ? 'Создание...' : 'Создать чат' }}
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

const phoneInput = ref('')
const creating = ref(false)
const error = ref('')

const createPrivateChat = async () => {
  creating.value = true
  error.value = ''
  
  try {
    const result = await chatsStore.createPrivateChat(phoneInput.value.trim())
    
    if (result.success) {
      emit('created', result.chat)
      emit('close')
      
      if (result.chat?.id) {
        router.push(`/chats/${result.chat.id}`)
      }
    } else {
      error.value = result.error
    }
  } catch (err) {
    error.value = err.message || 'Ошибка создания чата'
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

.form-group {
  margin-bottom: 15px;
}

.form-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 5px;
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