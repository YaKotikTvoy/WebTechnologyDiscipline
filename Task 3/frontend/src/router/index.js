import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import { auth } from '@/utils/auth'

const routes = [
  {
    path: '/',
    name: 'home',
    component: HomeView
  },
  {
    path: '/about',
    name: 'about',
    component: () => import('../views/AboutView.vue')
  },
  {
    path: '/products',
    name: 'products',
    component: () => import('../views/ProductsView.vue')
  },
  {
    path: '/products-table',
    name: 'products-table',
    component: () => import('../views/ProductsTableView.vue')
  },
  {
    path: '/product/:id',
    name: 'product-detail',
    component: () => import('../views/ProductDetailView.vue'),
    props: true
  },
  {
    path: '/employee/:id',
    name: 'employee',
    component: () => import('../components/EmployeeProfile.vue'),
    props: true
  },
  
  // Аутентификация
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
    meta: { guestOnly: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guestOnly: true }
  },
  
  // Авторизованные пользователи
  {
    path: '/profile',
    name: 'profile',
    component: () => import('../views/ProfileView.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/cart',
    name: 'cart',
    component: () => import('../views/CartView.vue'),
    meta: { requiresAuth: true }
  },
  
  // Продавцы
  {
    path: '/seller',
    name: 'seller',
    component: () => import('../views/SellerView.vue'),
    meta: { requiresAuth: true, requiresSeller: true }
  },
  
  // Администраторы
  {
    path: '/admin',
    name: 'admin',
    component: () => import('../views/AdminView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Навигационные хуки
router.beforeEach((to, from, next) => {
  const isAuthenticated = auth.isAuthenticated()
  const isSeller = auth.isSeller()
  const isAdmin = auth.isAdmin()
  
  // Требуется авторизация
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next('/login')
  }
  
  // Только для гостей
  if (to.meta.guestOnly && isAuthenticated) {
    return next('/')
  }
  
  // Только для продавцов
  if (to.meta.requiresSeller && !isSeller) {
    alert('Только продавцы имеют доступ к этой странице')
    return next('/')
  }
  
  // Только для администраторов
  if (to.meta.requiresAdmin && !isAdmin) {
    alert('Только администраторы имеют доступ к этой странице')
    return next('/')
  }
  
  next()
})

export default router