<template>
  <a :href="fileUrl" target="_blank" class="text-decoration-none">
    
    <div v-if="isImage" class="file-preview mb-2">
      <img :src="fileUrl" 
           :alt="file.filename"
           class="img-thumbnail"
           style="max-width: 200px; max-height: 200px;">
      <div class="small text-muted mt-1">{{ file.filename }}</div>
    </div>
    
    <div v-else class="d-flex align-items-center p-2 bg-white rounded border">
      <i class="bi bi-file-earmark me-2 fs-4"></i>
      <div>
        <div class="fw-bold">{{ file.filename }}</div>
        <div class="small text-muted">{{ formatFileSize(file.filesize) }}</div>
      </div>
    </div>
  </a>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  file: {
    type: Object,
    required: true
  }
})

const isImage = computed(() => {
  const imageTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp']
  return imageTypes.includes(props.file.mime_type) || 
         props.file.filename.match(/\.(jpg|jpeg|png|gif|webp)$/i)
})

const fileUrl = computed(() => {
  return `http://localhost:8080/uploads/${props.file.filepath}`
})

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>