<template>
  <div :class="['mb-3 d-flex', { 'justify-content-end': isOwnMessage }]">
    <div v-if="message.is_deleted" class="text-muted small">
      Сообщение удалено
    </div>
    
    <div v-else-if="message.type && message.type.startsWith('system_')" 
         class="text-center my-2 w-100">
      <div class="badge bg-secondary text-white py-2 px-3 fw-normal opacity-75 d-inline-block">
        <i v-if="message.type === 'system_user_added'" class="bi bi-person-plus me-1"></i>
        <i v-else-if="message.type === 'system_user_removed'" class="bi bi-person-dash me-1"></i>
        <i v-else-if="message.type === 'system_chat_created'" class="bi bi-chat-left-dots me-1"></i>
        <i v-else class="bi bi-info-circle me-1"></i>
        {{ message.content }}
      </div>
      <div class="text-muted small mt-1">
        {{ formatTime(message.created_at) }}
      </div>
    </div>
    
    <div v-else 
         class="p-3 rounded shadow-sm position-relative" 
         :class="isOwnMessage ? 'bg-primary text-white' : 'bg-white'"
         style="max-width: 70%;"
         @mouseenter="showActions = true"
         @mouseleave="showActions = false">
      
      <div v-if="showActions && (canEdit || canDelete)" 
           class="position-absolute bg-dark bg-opacity-75 rounded p-1"
           style="top: -10px; right: -10px; z-index: 100;">
        <div class="btn-group btn-group-sm" role="group">
          <button v-if="canEdit" 
                  class="btn btn-light border-0" 
                  @click.stop="startEdit"
                  title="Редактировать"
                  style="padding: 2px 6px;">
            <i class="bi bi-pencil"></i>
          </button>
          <button v-if="canDelete" 
                  class="btn btn-light border-0" 
                  @click.stop="confirmDelete"
                  title="Удалить"
                  style="padding: 2px 6px;">
            <i class="bi bi-trash text-danger"></i>
          </button>
        </div>
      </div>
      
      <div v-if="editing" class="mb-2">
        <textarea v-model="editContent" 
                  class="form-control" 
                  rows="2"
                  @keydown.enter.exact.prevent="saveEdit"
                  @keydown.esc="cancelEdit"
                  ref="editTextareaRef"
                  style="resize: vertical; min-height: 60px;"></textarea>
        <div class="d-flex gap-1 mt-2">
          <button class="btn btn-sm btn-success" @click.stop="saveEdit">
            <i class="bi bi-check"></i> Сохранить
          </button>
          <button class="btn btn-sm btn-secondary" @click.stop="cancelEdit">
            <i class="bi bi-x"></i> Отмена
          </button>
        </div>
      </div>
      
      <div v-else>
        <div v-if="!isOwnMessage && showSender" 
             class="small mb-1"
             :class="isOwnMessage ? 'text-white-50' : 'text-muted'">
          {{ message.sender?.username || message.sender?.phone }}
        </div>
        
        <div class="message-content" style="white-space: pre-wrap;">
          {{ message.content }}
          <span v-if="message.is_edited" 
                class="small ms-1"
                :class="isOwnMessage ? 'text-white-50' : 'text-muted'">
            <i class="bi bi-pencil"></i> изменено
          </span>
        </div>
        
        <div v-if="message.files && message.files.length > 0" class="mt-2">
          <MessageFile 
            v-for="file in message.files" 
            :key="file.id" 
            :file="file" 
          />
        </div>
        
        <div class="small mt-1 d-flex align-items-center gap-1" 
             :class="isOwnMessage ? 'text-white-50' : 'text-muted'">
          <span>{{ formatTime(message.created_at) }}</span>
          <div v-if="isOwnMessage" class="ms-1 d-flex align-items-center">
            <i v-if="message.readers && message.readers.length > 0 && allReaders.length > 0" 
               class="bi bi-check2-all"
               :class="allReaders.length > 0 ? 'text-info' : 'text-muted'"
               :title="allReaders.length > 0 ? 'Прочитано' : 'Доставлено'"></i>
            <i v-else class="bi bi-check2 text-muted" title="Отправлено"></i>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import MessageFile from './MessageFile.vue'

const props = defineProps({
  message: {
    type: Object,
    required: true
  },
  userId: {
    type: Number,
    required: true
  },
  showSender: {
    type: Boolean,
    default: true
  },
  isChatAdmin: {
    type: Boolean,
    default: false
  },
  chatType: {
    type: String,
    default: 'private'
  }
})

const emit = defineEmits(['edit', 'delete'])

const showActions = ref(false)
const editing = ref(false)
const editContent = ref('')
const editTextareaRef = ref(null)

const isOwnMessage = computed(() => props.message.sender_id === props.userId)

const canEdit = computed(() => {
  return isOwnMessage.value && 
         !props.message.is_deleted &&
         !editing.value
})

const canDelete = computed(() => {
  if (props.message.is_deleted) return false
  if (isOwnMessage.value) return true
  if (props.chatType === 'private') return true
  if (props.isChatAdmin) return true
  return false
})

const allReaders = computed(() => {
  if (!props.message.readers || !Array.isArray(props.message.readers)) return []
  return props.message.readers
})

const formatTime = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleTimeString([], { 
    hour: '2-digit', 
    minute: '2-digit' 
  })
}

const startEdit = () => {
  if (!canEdit.value) return
  
  editContent.value = props.message.content
  editing.value = true
  
  nextTick(() => {
    if (editTextareaRef.value) {
      editTextareaRef.value.focus()
      editTextareaRef.value.select()
    }
  })
}

const saveEdit = () => {
  const trimmedContent = editContent.value.trim()
  if (!trimmedContent || trimmedContent === props.message.content) {
    cancelEdit()
    return
  }
  
  emit('edit', {
    messageId: props.message.id,
    content: trimmedContent
  })
  editing.value = false
}

const cancelEdit = () => {
  editing.value = false
  editContent.value = ''
}

const confirmDelete = () => {
  if (window.confirm('Удалить сообщение? Это действие нельзя отменить.')) {
    emit('delete', props.message.id)
  }
}
</script>