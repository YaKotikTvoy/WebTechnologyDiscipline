<template>
  <main class="container">
    <h1 class="mb-4">Админ-панель CatPC</h1>
    
    <div v-if="!isAdmin" class="alert alert-danger">
      <i class="bi bi-shield-exclamation me-2"></i>
      Доступ запрещен. Только администратор CatPC имеет доступ к этой панели.
    </div>
    
    <div v-else>
      <!-- Меню -->
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'users' }]" 
                  @click="activeTab = 'users'; fetchUsers()">
            Пользователи
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'pending' }]" 
                  @click="activeTab = 'pending'; fetchPendingProducts()">
            Товары на проверке
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'all' }]" 
                  @click="activeTab = 'all'; fetchAllProducts()">
            Все товары
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'stats' }]" 
                  @click="activeTab = 'stats'; fetchStats()">
            Статистика
          </button>
        </li>
      </ul>
      
      <!-- Контент вкладок -->
      
      <!-- Пользователи -->
      <div v-if="activeTab === 'users'">
        <div v-if="loading" class="text-center my-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Загрузка...</span>
          </div>
        </div>
        
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>ID</th>
                <th>Имя</th>
                <th>Email</th>
                <th>Роль</th>
                <th>Статус</th>
                <th>Дата регистрации</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id" 
                  :class="{ 'table-secondary': !user.is_active }">
                <td>{{ user.id }}</td>
                <td>
                  <strong>{{ user.username }}</strong>
                  <span v-if="user.id === getUser?.id" class="badge bg-info ms-2">Вы</span>
                </td>
                <td>{{ user.email }}</td>
                <td>
                  <select class="form-select form-select-sm" v-model="user.role" 
                          @change="updateUserRole(user.id, user.role)" 
                          :disabled="user.id === getUser?.id">
                    <option value="customer">Покупатель</option>
                    <option value="seller">Продавец</option>
                    <option value="admin">Администратор</option>
                  </select>
                </td>
                <td>
                  <span class="badge" :class="user.is_active ? 'bg-success' : 'bg-danger'">
                    {{ user.is_active ? 'Активен' : 'Заблокирован' }}
                  </span>
                </td>
                <td>{{ formatDate(user.created_at) }}</td>
                <td>
                  <div class="btn-group btn-group-sm">
                    <button @click="toggleUserActive(user.id, user.is_active)" 
                            class="btn btn-outline-warning"
                            :disabled="user.id === getUser?.id">
                      <i class="bi" :class="user.is_active ? 'bi-lock' : 'bi-unlock'"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <!-- Товары на проверке -->
      <div v-if="activeTab === 'pending'">
        <div v-if="pendingLoading" class="text-center my-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Загрузка...</span>
          </div>
        </div>
        
        <div v-else-if="!pendingProducts || pendingProducts.length === 0" class="text-center py-5">
          <i class="bi bi-check-circle display-1 text-success mb-3"></i>
          <h4 class="text-success">Нет товаров на проверке</h4>
          <p class="text-muted">Все товары одобрены</p>
        </div>
        
        <div v-else>
          <div class="alert alert-info mb-4">
            <i class="bi bi-info-circle me-2"></i>
            Товары, добавленные продавцами, ожидают вашего одобрения
          </div>
          
          <div class="row">
            <div v-for="product in pendingProducts" :key="product.id" class="col-md-6 mb-4">
              <div class="card h-100 shadow-sm">
                <div class="card-body">
                  <div class="d-flex justify-content-between align-items-start mb-3">
                    <h5 class="card-title mb-0">{{ product.name }}</h5>
                    <span class="badge bg-warning">Ожидает</span>
                  </div>
                  
                  <p class="card-text small mb-3">{{ product.description }}</p>
                  
                  <div class="mb-3">
                    <strong>Продавец:</strong> {{ product.username || 'Неизвестно' }}
                  </div>
                  
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                    <span class="badge" :class="product.stock > 0 ? 'bg-success' : 'bg-danger'">
                      {{ product.stock }} шт.
                    </span>
                  </div>
                  
                  <div class="btn-group w-100">
                    <button @click="approveProduct(product.id)" class="btn btn-success">
                      <i class="bi bi-check-circle me-2"></i>Одобрить
                    </button>
                    <button @click="forceDeleteProduct(product.id)" class="btn btn-danger">
                      <i class="bi bi-trash me-2"></i>Удалить
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Все товары -->
      <div v-if="activeTab === 'all'">
        <div class="mb-3">
          <input type="text" class="form-control" placeholder="Поиск товаров..." 
                 v-model="searchQuery" @input="debounceSearch">
        </div>
        
        <div v-if="allProductsLoading" class="text-center my-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Загрузка...</span>
          </div>
        </div>
        
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>ID</th>
                <th>Название</th>
                <th>Продавец</th>
                <th>Цена</th>
                <th>Остаток</th>
                <th>Статус</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="product in allProducts" :key="product.id">
                <td>{{ product.id }}</td>
                <td>
                  <strong>{{ product.name }}</strong>
                  <div class="small text-muted">{{ truncateDescription(product.description, 50) }}</div>
                </td>
                <td>
                  <span v-if="product.user_id === getUser?.id" class="badge bg-info">CatPC</span>
                  <span v-else>{{ product.username || 'Неизвестно' }}</span>
                </td>
                <td>{{ formatPrice(product.price) }} ₽</td>
                <td>
                  <span class="badge" :class="product.stock > 0 ? 'bg-success' : 'bg-danger'">
                    {{ product.stock }} шт.
                  </span>
                </td>
                <td>
                  <span class="badge" :class="product.is_approved ? 'bg-success' : 'bg-warning'">
                    {{ product.is_approved ? 'Одобрен' : 'На проверке' }}
                  </span>
                </td>
                <td>
                  <div class="btn-group btn-group-sm">
                    <router-link :to="'/product/' + product.id" class="btn btn-outline-info">
                      <i class="bi bi-eye"></i>
                    </router-link>
                    <button @click="forceDeleteProduct(product.id)" class="btn btn-outline-danger">
                      <i class="bi bi-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <!-- Статистика -->
      <div v-if="activeTab === 'stats'">
        <div class="row">
          <div class="col-md-3 mb-4">
            <div class="card bg-primary text-white">
              <div class="card-body">
                <h5 class="card-title">Всего пользователей</h5>
                <p class="card-text display-4">{{ stats.totalUsers || 0 }}</p>
              </div>
            </div>
          </div>
          
          <div class="col-md-3 mb-4">
            <div class="card bg-success text-white">
              <div class="card-body">
                <h5 class="card-title">Продавцы</h5>
                <p class="card-text display-4">{{ stats.sellers || 0 }}</p>
              </div>
            </div>
          </div>
          
          <div class="col-md-3 mb-4">
            <div class="card bg-info text-white">
              <div class="card-body">
                <h5 class="card-title">Всего товаров</h5>
                <p class="card-text display-4">{{ stats.totalProducts || 0 }}</p>
              </div>
            </div>
          </div>
          
          <div class="col-md-3 mb-4">
            <div class="card bg-warning text-white">
              <div class="card-body">
                <h5 class="card-title">На проверке</h5>
                <p class="card-text display-4">{{ stats.pendingProducts || 0 }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import api from '@/services/api'
import { mapGetters } from 'vuex'
import { debounce } from 'lodash'

export default {
  name: 'AdminView',
  data() {
    return {
      activeTab: 'users',
      users: [],
      pendingProducts: [],
      allProducts: [],
      loading: false,
      pendingLoading: false,
      allProductsLoading: false,
      searchQuery: '',
      stats: {},
      debouncedSearch: null
    }
  },
  computed: {
    ...mapGetters(['isAdmin', 'getUser'])
  },
  created() {
    this.debouncedSearch = debounce(this.fetchAllProducts, 500)
  },
  methods: {
    async fetchUsers() {
      this.loading = true
      try {
        const response = await api.get('/api/admin/users')
        this.users = response.data || [] // защита от null
      } catch (error) {
        console.error('Ошибка загрузки пользователей:', error)
        this.users = [] // устанавливаем пустой массив при ошибке
        alert('Не удалось загрузить пользователей')
      } finally {
        this.loading = false
      }
    },
    
    async fetchPendingProducts() {
      this.pendingLoading = true
      try {
        const response = await api.get('/api/admin/pending-products')
        this.pendingProducts = response.data || [] // защита от null
        
        console.log('Товары на проверке:', this.pendingProducts)
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        this.pendingProducts = [] // устанавливаем пустой массив при ошибке
        alert('Не удалось загрузить товары на проверке')
      } finally {
        this.pendingLoading = false
      }
    },
    
    async fetchAllProducts() {
      this.allProductsLoading = true
      try {
        const params = {
          search: this.searchQuery || undefined
        }
        
        Object.keys(params).forEach(key => params[key] === undefined && delete params[key])
        
        const response = await api.get('/api/products', { params })
        this.allProducts = response.data.products || [] // защита от null
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        this.allProducts = [] // устанавливаем пустой массив при ошибке
        alert('Не удалось загрузить товары')
      } finally {
        this.allProductsLoading = false
      }
    },
    
    async fetchStats() {
      try {
        // Заглушка для статистики
        this.stats = {
          totalUsers: this.users.length,
          sellers: this.users.filter(u => u.role === 'seller').length,
          totalProducts: this.allProducts.length,
          pendingProducts: this.pendingProducts.length
        }
      } catch (error) {
        console.error('Ошибка загрузки статистики:', error)
      }
    },
    
    async updateUserRole(userId, newRole) {
      try {
        await api.put(`/api/admin/users/${userId}/role`, { role: newRole })
        alert('Роль пользователя обновлена')
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка обновления роли')
        // Откатываем изменения
        await this.fetchUsers()
      }
    },
    
    async toggleUserActive(userId, isActive) {
      if (userId === this.getUser?.id) {
        alert('Нельзя заблокировать себя')
        return
      }
      
      if (!confirm(`${isActive ? 'Заблокировать' : 'Разблокировать'} этого пользователя?`)) return
      
      try {
        await api.put(`/api/admin/users/${userId}/active`)
        alert(`Пользователь ${isActive ? 'заблокирован' : 'разблокирован'}`)
        await this.fetchUsers()
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка изменения статуса')
      }
    },
    
    async approveProduct(productId) {
      if (!confirm('Одобрить этот товар?')) return
      
      try {
        await api.put(`/api/admin/products/${productId}/approve`)
        alert('Товар одобрен')
        await this.fetchPendingProducts()
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка одобрения товара')
      }
    },
    
    async forceDeleteProduct(productId) {
      if (!confirm('Принудительно удалить этот товар? Это действие нельзя отменить.')) return
      
      try {
        await api.delete(`/api/admin/products/${productId}/force`)
        alert('Товар удален')
        await this.fetchPendingProducts()
        await this.fetchAllProducts()
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка удаления товара')
      }
    },
    
    formatPrice(price) {
      return new Intl.NumberFormat('ru-RU').format(price)
    },
    
    formatDate(dateString) {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    },
    
    truncateDescription(text, length) {
      if (!text) return ''
      if (text.length > length) {
        return text.substring(0, length) + '...'
      }
      return text
    }
  },
  async mounted() {
    if (this.isAdmin) {
      await this.fetchUsers()
      await this.fetchStats()
    }
  },
  watch: {
    isAdmin: {
      immediate: true,
      handler(newVal) {
        if (!newVal && this.$route.name === 'admin') {
          this.$router.push('/')
        }
      }
    },
    
    activeTab(newTab) {
      if (newTab === 'stats') {
        this.fetchStats()
      }
    }
  }
}
</script>