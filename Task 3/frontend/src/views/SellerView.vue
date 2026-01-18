<template>
  <main class="container">

    <div v-if="loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
      <p class="mt-2">Загрузка панели продавца...</p>
    </div>
    <div v-else>
      <div class="d-flex justify-content-between align-items-center mb-4">
        <h1 class="mb-0">Панель продавца</h1>
        <button @click="showAddModal = true" class="btn btn-success">
          <i class="bi bi-plus-circle me-2"></i>Добавить товар
        </button>
      </div>
      
      <div v-if="!isSeller" class="alert alert-warning">
        <i class="bi bi-exclamation-triangle me-2"></i>
        У вас нет прав продавца. Обратитесь к администратору.
      </div>
      
      <div v-else>
        <!-- Меню -->
        <ul class="nav nav-tabs mb-4">
          <li class="nav-item">
            <button :class="['nav-link', { active: activeTab === 'my' }]" 
                    @click="activeTab = 'my'">
              Мои товары
            </button>
          </li>
          <li class="nav-item">
            <button :class="['nav-link', { active: activeTab === 'pending' }]" 
                    @click="activeTab = 'pending'">
              Ожидают одобрения
            </button>
          </li>
          <li class="nav-item">
            <button :class="['nav-link', { active: activeTab === 'sales' }]" 
                    @click="activeTab = 'sales'">
              Статистика продаж
            </button>
          </li>
        </ul>
        
        <!-- Контент вкладок -->
        <div v-if="activeTab === 'my'">
          <div v-if="loading" class="text-center my-5">
            <div class="spinner-border text-primary" role="status">
              <span class="visually-hidden">Загрузка...</span>
            </div>
          </div>
          
          <div v-else-if="products.length === 0" class="text-center py-5">
            <i class="bi bi-inboxes display-1 text-muted mb-3"></i>
            <h4 class="text-muted">У вас пока нет товаров</h4>
            <button @click="showAddModal = true" class="btn btn-primary mt-3">
              <i class="bi bi-plus-circle me-2"></i>Добавить первый товар
            </button>
          </div>
          
          <div v-else class="row">
            <div v-for="product in products" :key="product.id" class="col-md-6 col-lg-4 mb-4">
              <div class="card h-100 shadow-sm">
                <div class="position-relative">
                  <img :src="getImageUrl(product.image)" class="card-img-top" 
                      :alt="product.name" style="height: 200px; object-fit: contain;">
                  <span class="position-absolute top-0 end-0 m-2 badge" 
                        :class="product.is_approved ? 'bg-success' : 'bg-warning'">
                    {{ product.is_approved ? 'Одобрен' : 'На проверке' }}
                  </span>
                </div>
                <div class="card-body d-flex flex-column">
                  <h5 class="card-title">{{ product.name }}</h5>
                  <p class="card-text flex-grow-1 small">{{ truncateDescription(product.description, 80) }}</p>
                  
                  <div class="mt-auto">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <span class="h5 text-primary mb-0">{{ formatPrice(product.price) }} ₽</span>
                      <span class="badge" :class="product.stock > 0 ? 'bg-success' : 'bg-danger'">
                        {{ product.stock }} шт.
                      </span>
                    </div>
                    
                    <div class="btn-group w-100">
                      <button @click="editProduct(product)" class="btn btn-outline-primary btn-sm">
                        <i class="bi bi-pencil"></i>
                      </button>
                      <button @click="deleteProduct(product.id)" class="btn btn-outline-danger btn-sm">
                        <i class="bi bi-trash"></i>
                      </button>
                      <router-link :to="'/product/' + product.id" class="btn btn-outline-info btn-sm">
                        <i class="bi bi-eye"></i>
                      </router-link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Ожидающие одобрения -->
        <div v-if="activeTab === 'pending'">
          <div v-if="pendingProducts.length === 0" class="text-center py-5">
            <i class="bi bi-check-circle display-1 text-success mb-3"></i>
            <h4 class="text-success">Все товары одобрены</h4>
          </div>
          
          <div v-else>
            <div class="alert alert-info">
              <i class="bi bi-info-circle me-2"></i>
              Эти товары ожидают проверки администратором
            </div>
            
            <div class="list-group">
              <div v-for="product in pendingProducts" :key="product.id" class="list-group-item">
                <div class="d-flex justify-content-between align-items-center">
                  <div>
                    <h5 class="mb-1">{{ product.name }}</h5>
                    <p class="mb-1 small text-muted">{{ truncateDescription(product.description, 100) }}</p>
                    <small class="text-muted">Добавлен: {{ formatDate(product.created_at) }}</small>
                  </div>
                  <div>
                    <button @click="editProduct(product)" class="btn btn-outline-primary btn-sm me-2">
                      <i class="bi bi-pencil"></i>
                    </button>
                    <button @click="deleteProduct(product.id)" class="btn btn-outline-danger btn-sm">
                      <i class="bi bi-trash"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Статистика -->
        <div v-if="activeTab === 'sales'">
          <div class="alert alert-info">
            <i class="bi bi-bar-chart me-2"></i>
            Статистика продаж будет доступна после реализации системы заказов
          </div>
        </div>
      </div>
      
      <!-- Модальное окно добавления/редактирования товара -->
      <div v-if="showAddModal || showEditModal" class="modal fade show" style="display: block; background: rgba(0,0,0,0.5)">
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ editingProduct ? 'Редактировать товар' : 'Добавить товар' }}</h5>
              <button type="button" class="btn-close" @click="closeModal"></button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="saveProduct">
                <div class="mb-3">
                  <label class="form-label">Название товара</label>
                  <input type="text" class="form-control" v-model="productForm.name" required>
                </div>
                
                <div class="mb-3">
                  <label class="form-label">Описание</label>
                  <textarea class="form-control" v-model="productForm.description" rows="4" required></textarea>
                </div>
                
                <div class="row">
                  <div class="col-md-6 mb-3">
                    <label class="form-label">Цена (₽)</label>
                    <input type="number" class="form-control" v-model="productForm.price" min="0" step="0.01" required>
                  </div>
                  
                  <div class="col-md-6 mb-3">
                    <label class="form-label">Количество</label>
                    <input type="number" class="form-control" v-model="productForm.stock" min="0" required>
                  </div>
                </div>
                
                <div class="mb-3">
                  <label class="form-label">Изображение</label>
                  <input type="text" class="form-control" v-model="productForm.image" 
                        placeholder="Имя файла (например: product.jpg)" required>
                  <div class="form-text">Укажите имя файла изображения из папки /img/</div>
                </div>
                
                <div class="alert alert-warning" v-if="!isAdmin">
                  <i class="bi bi-info-circle me-2"></i>
                  После сохранения товар будет отправлен на проверку администратору
                </div>
              </form>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">Отмена</button>
              <button type="button" class="btn btn-primary" @click="saveProduct" :disabled="saving">
                <span v-if="saving" class="spinner-border spinner-border-sm me-2"></span>
                {{ editingProduct ? 'Обновить' : 'Добавить' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import api from '@/services/api'
import axios from 'axios'
import { mapGetters } from 'vuex'

export default {
  name: 'SellerView',
  data() {
    return {
      activeTab: 'my',
      products: [],
      pendingProducts: [],
      loading: false,
      showAddModal: false,
      showEditModal: false,
      editingProduct: null,
      saving: false,
      productForm: {
        name: '',
        description: '',
        price: 0,
        stock: 0,
        image: ''
      }
    }
  },
  computed: {
    ...mapGetters(['isSeller', 'isAdmin', 'getUser'])
  },
  methods: {
    async fetchMyProducts() {
      console.log('🛒 fetchMyProducts - начало')
      
      this.loading = true
      try {
        const response = await api.get('/api/seller/my-products')
        console.log('✅ Ответ сервера:', response.data)
        
        // ЗАЩИТА ОТ NULL
        const data = response.data || []
        
        this.products = data.filter(p => p.is_approved)
        this.pendingProducts = data.filter(p => !p.is_approved)
        
        console.log(`✅ Обработано: approved=${this.products.length}, pending=${this.pendingProducts.length}`)
        
      } catch (error) {
        console.error('❌ Ошибка fetchMyProducts:', error)
        
        // Устанавливаем пустые массивы при ошибке
        this.products = []
        this.pendingProducts = []
        
        alert('Не удалось загрузить товары: ' + (error.response?.data?.error || error.message))
      } finally {
        this.loading = false
        console.log('🛒 fetchMyProducts - завершено')
      }
    },
    
    editProduct(product) {
      this.editingProduct = product
      this.productForm = {
        name: product.name,
        description: product.description,
        price: product.price,
        stock: product.stock,
        image: product.image
      }
      this.showEditModal = true
    },
    
    async saveProduct() {
      this.saving = true
      try {
        if (this.editingProduct) {
          await axios.put(`http://localhost:1323/api/seller/products/${this.editingProduct.id}`, this.productForm)
          alert('Товар обновлен и отправлен на повторную проверку')
        } else {
          await axios.post('http://localhost:1323/api/seller/products', this.productForm)
          alert('Товар добавлен и отправлен на проверку')
        }
        this.closeModal()
        await this.fetchMyProducts()
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка сохранения товара')
      } finally {
        this.saving = false
      }
    },
    
    async deleteProduct(productId) {
      if (!confirm('Удалить этот товар?')) return
      
      try {
        await axios.delete(`http://localhost:1323/api/seller/products/${productId}`)
        alert('Товар удален')
        await this.fetchMyProducts()
      } catch (error) {
        alert(error.response?.data?.error || 'Ошибка удаления товара')
      }
    },
    
    closeModal() {
      this.showAddModal = false
      this.showEditModal = false
      this.editingProduct = null
      this.productForm = {
        name: '',
        description: '',
        price: 0,
        stock: 0,
        image: ''
      }
    },
    
    getImageUrl(imageName) {
      return `/img/${imageName}`
    },
    
    truncateDescription(text, length) {
      if (!text) return ''
      if (text.length > length) {
        return text.substring(0, length) + '...'
      }
      return text
    },
    
    formatPrice(price) {
      return new Intl.NumberFormat('ru-RU').format(price)
    },
    
    formatDate(dateString) {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    }
  },
  async mounted() {
    console.log('🔄 SellerView mounted - начало')
    console.log('🔍 Vuex состояние:', {
      isAuthenticated: this.isAuthenticated,
      getUser: this.getUser,
      isSeller: this.isSeller
    })
    
    // 1. Попробуем обновить состояние из сервера
    try {
      console.log('🔄 Проверяем серверную сессию...')
      const result = await this.$store.dispatch('fetchProfile')
      
      if (result.success) {
        console.log('✅ Серверная сессия найдена:', result.data)
      } else {
        console.log('❌ Серверная сессия не найдена')
      }
    } catch (error) {
      console.error('❌ Ошибка проверки сессии:', error)
    }
    
    // 2. После проверки сервера снова проверяем состояние
    console.log('🔍 Обновленное Vuex состояние:', {
      isAuthenticated: this.isAuthenticated,
      getUser: this.getUser,
      isSeller: this.isSeller
    })
    
    // 3. Если все еще не авторизован
    if (!this.isAuthenticated) {
      console.log('❌ Пользователь не авторизован в Vuex')
      
      // Проверим, есть ли сохраненный пользователь в localStorage
      const savedUser = localStorage.getItem('vuex') 
        ? JSON.parse(localStorage.getItem('vuex')).user 
        : null
      
      console.log('📦 Сохраненный пользователь в localStorage:', savedUser)
      
      if (savedUser) {
        console.log('🔄 Восстанавливаем пользователя из localStorage')
        this.$store.commit('SET_USER', savedUser)
      } else {
        console.log('🚫 Перенаправляем на логин')
        alert('Для доступа к панели продавца необходимо войти в систему')
        this.$router.push('/login')
        return
      }
    }
    
    // 4. Проверяем права
    if (!this.isSeller && !this.isAdmin) {
      console.log('❌ У пользователя нет прав продавца/админа')
      alert('Только продавцы и администраторы имеют доступ к этой панели')
      this.$router.push('/')
      return
    }
    
    // 5. Загружаем товары
    await this.fetchMyProducts()
    
    console.log('✅ SellerView mounted - завершено')
  },
  watch: {
    isSeller: {
      immediate: true,
      handler(newVal) {
        if (!newVal && this.$route.name === 'seller') {
          this.$router.push('/')
        }
      }
    }
  }
}
</script>