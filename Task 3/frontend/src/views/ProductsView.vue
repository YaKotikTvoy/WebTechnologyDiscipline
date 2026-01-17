<template>
  <main class="container">
    <h1>Техника</h1>
    
    <div v-if="loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>
    
    <div v-else class="row">
      <div v-for="product in products" :key="product.id" class="col-md-6 mb-4">
        <div class="card shadow rounded my-3">
          <div class="card-body">
            <div class="row">
              <div class="col-md-4">
                <div class="product-image-container">
                  <img :src="getImageUrl(product.image)" class="product-image" :alt="product.name">
                </div>
              </div>
              
              <div class="col-md-8 d-flex flex-column">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-description">{{ truncateDescription(product.description, 150) }}</div>
                
                <div class="flex-grow-1"></div>
                
                <div class="d-flex justify-content-end align-items-center mt-3">
                  <div class="product-price me-3">
                    Цена: <strong>{{ formatPrice(product.price) }} р.</strong>
                    <span v-if="product.stock > 0" class="text-success ms-2">
                      <i class="bi bi-check-circle"></i> В наличии
                    </span>
                    <span v-else class="text-danger ms-2">
                      <i class="bi bi-x-circle"></i> Нет в наличии
                    </span>
                  </div>
                  
                  <div class="btn-group">
                    <button @click="addToCartHandler(product)" class="btn btn-warning" 
                            :disabled="product.stock === 0" title="В корзину">
                      <i class="bi bi-cart"></i>
                      <span class="d-none d-md-inline">В корзину</span>
                    </button>
                    
                    <router-link :to="'/product/' + product.id" class="btn btn-info" title="Подробнее">
                      <i class="bi bi-info-circle"></i>
                      <span class="d-none d-md-inline">Подробнее</span>
                    </router-link>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Пагинация -->
    <nav v-if="totalPages > 1" class="mt-4">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <button class="page-link" @click="changePage(currentPage - 1)">
            <i class="bi bi-chevron-left"></i> Назад
          </button>
        </li>
        
        <li v-for="page in totalPages" :key="page" class="page-item" :class="{ active: page === currentPage }">
          <button class="page-link" @click="changePage(page)">{{ page }}</button>
        </li>
        
        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <button class="page-link" @click="changePage(currentPage + 1)">
            Вперёд <i class="bi bi-chevron-right"></i>
          </button>
        </li>
      </ul>
    </nav>
  </main>
</template>

<script>
import api from '@/services/api'
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'ProductsView',
  data() {
    return {
      products: [],
      currentPage: 1,
      totalPages: 1,
      total: 0,
      loading: false
    }
  },
  computed: {
    ...mapGetters(['isAuthenticated'])
  },
  methods: {
    ...mapActions(['addToCart']),
    
    async fetchProducts(page = 1) {
      this.loading = true
      try {
        const response = await api.get(`/api/products?page=${page}&limit=6`)
        this.products = response.data.products
        this.currentPage = response.data.page
        this.totalPages = response.data.totalPages
        this.total = response.data.total
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        alert('Не удалось загрузить товары')
      } finally {
        this.loading = false
      }
    },
    
    changePage(page) {
      if (page >= 1 && page <= this.totalPages && page !== this.currentPage) {
        this.fetchProducts(page)
        window.scrollTo(0, 0)
      }
    },
    
    async addToCartHandler(product) {

      if (product.stock === 0) {
        alert('Товар временно отсутствует')
        return
      }
      
      if (!this.isAuthenticated) {
        if (confirm('Для добавления товара в корзину нужно войти. Перейти на страницу входа?')) {
          this.$router.push('/login')
        }
        return
      }
      
      const result = await this.addToCart({ productId: product.id, quantity: 1 })
      if (result.success) {
        alert(`Товар "${product.name}" добавлен в корзину!`)
      } else {
        alert(result.error || 'Ошибка добавления в корзину')
      }
    },
    
    formatPrice(price) {
      return new Intl.NumberFormat('ru-RU').format(price)
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
    }
  },
  mounted() {
    this.fetchProducts(this.currentPage)
  }
}
</script>

<style scoped>
.product-image-container {
  width: 100%;
  height: 150px;
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid #dee2e6;
  background-color: #f8f9fa;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  padding: 5px;
  transition: transform 0.3s ease;
}

.product-image:hover {
  transform: scale(1.05);
}

.product-name {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 10px;
  color: #212529;
}

.product-description {
  font-size: 0.95rem;
  color: #6c757d;
  line-height: 1.5;
}

.product-price {
  padding: 8px 12px;
  border-radius: 6px;
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
  font-size: 1.1rem;
}

@media (max-width: 768px) {
  .product-image-container {
    height: 120px;
    margin-bottom: 15px;
  }
  
  .col-md-4, .col-md-8 {
    width: 100%;
    flex: 0 0 100%;
    max-width: 100%;
  }
}
</style>