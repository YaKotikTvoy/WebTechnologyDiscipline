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
                <img :src="getImageUrl(product.image)" class="product-image">
              </div>
              <div class="col-md-8 d-flex flex-column">
                <div class="product-name">{{ product.name }}</div>
                <div class="product-description">{{ truncateDescription(product.description, 150) }}</div>
                <div class="flex-grow-1"></div>
                <div class="d-flex justify-content-end align-items-center mt-3">
                  <div class="product-price me-3">
                    Цена: <strong>{{ formatPrice(product.price) }} р.</strong>
                  </div>
                  <div class="btn-group">
                    <button @click="addToCart(product)" class="btn btn-warning" 
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
    <nav v-if="totalPages > 1">
      <ul class="pagination justify-content-center">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <button class="page-link" @click="changePage(currentPage - 1)">Назад</button>
        </li>
        
        <li v-for="page in totalPages" :key="page" class="page-item" :class="{ active: page === currentPage }">
          <button class="page-link" @click="changePage(page)">{{ page }}</button>
        </li>
        
        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <button class="page-link" @click="changePage(currentPage + 1)">Вперёд</button>
        </li>
      </ul>
    </nav>
  </main>
</template>

<script>
import axios from 'axios'

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
  methods: {
    async fetchProducts(page = 1) {
      this.loading = true
      try {
        const response = await axios.get(`http://localhost:1323/api/products?page=${page}&limit=6`)
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
    addToCart(product) {
      if (product.stock === 0) {
        alert('Товар временно отсутствует')
        return
      }
      alert(`Товар "${product.name}" добавлен в корзину!`)
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