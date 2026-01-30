<template>
  <div class="p-3 border-bottom bg-white">
    <div class="d-flex align-items-center">
      <button v-if="showBackButton" 
              class="btn btn-link text-dark d-md-none me-2" 
              @click="$emit('back')">
        <i class="bi bi-arrow-left"></i>
      </button>
      <div class="d-flex align-items-center flex-grow-1">
        <div class="rounded-circle d-flex align-items-center justify-content-center me-3" 
             :class="chatColor"
             style="width: 50px; height: 50px; font-size: 1.2rem;">
          {{ chatInitial }}
        </div>
        <div>
          <h5 class="mb-0">{{ chatTitle }}</h5>
          <div class="text-muted small">{{ memberCount }} участников</div>
        </div>
      </div>
      
      <slot name="actions">
        <!-- Сюда можно вставить кнопки действий -->
      </slot>
    </div>
  </div>
</template>

<script setup>
defineProps({
  chatTitle: {
    type: String,
    default: 'Чат'
  },
  chatColor: {
    type: String,
    default: 'bg-primary'
  },
  chatInitial: {
    type: String,
    default: 'Ч'
  },
  memberCount: {
    type: Number,
    default: 0
  },
  showBackButton: {
    type: Boolean,
    default: true
  }
})

defineEmits(['back'])
</script>