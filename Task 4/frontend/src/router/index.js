import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/components/auth/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/components/auth/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/components/auth/Profile.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('@/components/chats/ChatList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/chats/:id',
    name: 'Chat',
    component: () => import('@/components/chats/ChatWindow.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/friends',
    name: 'Friends',
    component: () => import('@/components/contacts/FriendsList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/friends/add',
    name: 'AddFriend',
    component: () => import('@/components/contacts/AddFriend.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/friends/requests',
    name: 'FriendRequests',
    component: () => import('@/components/contacts/FriendRequests.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (!requiresAuth && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router