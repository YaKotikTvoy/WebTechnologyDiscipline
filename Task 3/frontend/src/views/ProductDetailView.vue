<template>
  <main class="container">
    <router-link to="/products" class="btn btn-outline-primary mb-4">
      <i class="bi bi-arrow-left"></i> Назад к товарам
    </router-link>

    <div v-if="loading" class="text-center">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>

    <div v-else-if="product" class="card shadow">
      <div class="card-body">
        <div class="row">
          <div class="col-md-4">
            <img :src="getImageUrl(product.image)" class="img-fluid rounded mb-3" :alt="product.name">
          </div>
          <div class="col-md-8">
            <div class="d-flex justify-content-between align-items-start mb-3">
              <h2 class="card-title mb-0">{{ product.name }}</h2>
              <button @click="addToCartHandler(product)" class="btn btn-warning" 
                      :disabled="product.stock === 0">
                <i class="bi bi-cart-plus"></i> В корзину
              </button>
            </div>
            
            <div class="mb-4">
              <span class="badge bg-secondary me-2">Артикул: {{ product.id }}</span>
              <span v-if="product.user_id" class="badge bg-info">
                <i class="bi bi-shop"></i> Продавец: {{ product.username || 'CatPC' }}
              </span>
            </div>

            <div class="mb-4">
              <h4>Характеристики</h4>
              <div class="bg-light p-3 rounded">
                <pre class="mb-0" style="white-space: pre-wrap; font-family: inherit;">{{ product.description }}</pre>
              </div>
            </div>

            <div class="row mb-4">
              <div class="col-md-6">
                <div class="card">
                  <div class="card-body">
                    <h5 class="card-title">Наличие</h5>
                    <p class="card-text">
                      <span v-if="product.stock > 0" class="text-success">
                        <i class="bi bi-check-circle"></i> В наличии: {{ product.stock }} шт.
                      </span>
                      <span v-else class="text-danger">
                        <i class="bi bi-x-circle"></i> Нет в наличии
                      </span>
                    </p>
                  </div>
                </div>
              </div>
              <div class="col-md-6">
                <div class="card">
                  <div class="card-body">
                    <h5 class="card-title">Цена</h5>
                    <p class="card-text display-6 text-primary">
                      {{ formatPrice(product.price) }} ₽
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="alert alert-danger">
      Товар не найден
    </div>
  </main>
</template>

<script>
import api from '@/services/api'
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'ProductDetailView',
  data() {
    return {
      product: null,
      loading: true
    }
  },
  computed: {
    productId() {
      return parseInt(this.$route.params.id)
    },
    ...mapGetters(['isAuthenticated'])
  },
  methods: {
    ...mapActions(['addToCart']),
    
    async fetchProduct() {
      this.loading = true
      try {
        const response = await api.get(`/api/products/${this.productId}`)
        this.product = response.data
      } catch (error) {
        console.error('Ошибка загрузки товара:', error)
        if (error.response?.status === 404) {
          this.$router.push('/products')
        } else {
          alert('Не удалось загрузить информацию о товаре')
        }
      } finally {
        this.loading = false
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
    
    getImageUrl(imageName) {
      return `/img/${imageName}`
    },
    
    formatPrice(price) {
      return new Intl.NumberFormat('ru-RU').format(price)
    }
  },
  mounted() {
    this.fetchProduct()
  }
}
</script>