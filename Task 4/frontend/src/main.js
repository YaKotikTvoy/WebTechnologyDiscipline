
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap'
import 'vue3-emoji-picker/css'
import 'bootstrap-icons/font/bootstrap-icons.css'

import { useChatsStore } from '@/stores/chats'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)

const chatsStore = useChatsStore()
chatsStore.initializeFromStorage()

app.mount('#app')