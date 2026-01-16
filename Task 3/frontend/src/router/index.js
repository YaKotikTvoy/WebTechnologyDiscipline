import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

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
    path: '/employee/:id',
    name: 'employee',
    component: () => import('../components/EmployeeProfile.vue'),
    props: true
  },
  {
    path: '/product/:id',
    name: 'product-detail',
    component: () => import('../views/ProductDetailView.vue'),
    props: true
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router