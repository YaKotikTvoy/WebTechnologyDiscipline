<template>
  <div class="p-3 border-top bg-white">
    <form @submit.prevent="handleSubmit" class="d-flex align-items-center">
      
      <button type="button" 
              class="btn btn-outline-secondary me-2" 
              @click="attachFile">
        <i class="bi bi-paperclip fs-5"></i>
      </button>
      
      <div class="flex-grow-1 me-2">
        <textarea v-model="text" 
                  class="form-control" 
                  placeholder="Сообщение..." 
                  rows="1"
                  @keydown="handleKeydown"
                  style="resize: none; min-height: 40px;"
                  ref="textareaRef"></textarea>
      </div>
      
      <button type="button" 
              class="btn btn-outline-secondary me-2" 
              @click="toggleEmojiPicker">
        <i class="bi bi-emoji-smile-fill fs-5"></i>
      </button>
      
      <button type="submit" 
              class="btn btn-primary" 
              :disabled="!canSend">
        <i class="bi bi-send-fill fs-5"></i>
      </button>
    </form>
    
    <div v-if="files.length > 0" class="mt-2">
      <div v-for="(file, index) in files" :key="index" class="badge bg-info me-2 mb-1 d-inline-flex align-items-center">
        {{ file.name }}
        <button type="button" 
                class="btn-close btn-close-white ms-1" 
                style="font-size: 0.7rem;"
                @click="removeFile(index)"></button>
      </div>
    </div>
    
    <div v-if="showEmojiPicker" class="mt-2 p-2 border rounded">
      <div class="d-flex flex-wrap gap-1">
        <button v-for="emoji in emojis" 
                :key="emoji" 
                class="btn btn-sm btn-outline-secondary"
                @click="insertEmoji(emoji)"
                style="font-size: 1.2rem; padding: 2px 6px;">
          {{ emoji }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'

const emit = defineEmits(['send'])

const text = ref('')
const files = ref([])
const showEmojiPicker = ref(false)
const textareaRef = ref(null)

const emojis = ['😊', '😂', '❤️', '👍', '🔥', '🎉', '👏', '🙏', '😎', '🥳', '😢', '🤔', '🎯', '💯']

const canSend = computed(() => {
  return text.value.trim().length > 0 || files.value.length > 0
})

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

const handleSubmit = () => {
  sendMessage()
}

const sendMessage = () => {
  if (!canSend.value) return
  
  emit('send', { 
    text: text.value, 
    files: [...files.value]
  })
  resetForm()
}

const attachFile = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.multiple = true
  input.onchange = (e) => {
    const selectedFiles = Array.from(e.target.files)
    selectedFiles.forEach(file => {
      if (file.size <= 10 * 1024 * 1024) {
        files.value.push(file)
      } else {
        alert(`Файл ${file.name} слишком большой. Максимальный размер: 10MB`)
      }
    })
  }
  input.click()
}

const removeFile = (index) => {
  files.value.splice(index, 1)
}

const toggleEmojiPicker = () => {
  showEmojiPicker.value = !showEmojiPicker.value
}

const insertEmoji = (emoji) => {
  text.value += emoji
  showEmojiPicker.value = false
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.focus()
    }
  })
}

const resetForm = () => {
  text.value = ''
  files.value = []
  showEmojiPicker.value = false
  if (textareaRef.value) {
    textareaRef.value.style.height = 'auto'
  }
}

defineExpose({ resetForm })
</script>