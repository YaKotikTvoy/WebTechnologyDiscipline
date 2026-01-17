<template>
  <main class="container">
    <h1>Полный список техники, имеющийся на складе</h1>
    
    <!-- Таблица -->
    <table class="table">
      <thead>
        <tr>
          <th scope="col">Артикул</th>
          <th scope="col">Наименование</th>
          <th scope="col">Количество</th>
          <th scope="col">Цена</th>
          <th scope="col">Действия</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="product in products" :key="product.id">
          <th scope="row">{{ product.id }}</th>
          <td>{{ product.name }}</td>
          <td>{{ product.stock }} шт.</td>
          <td>{{ formatPrice(product.price) }} р.</td>
          <td>
            <button @click="addToCartHandler(product)" class="btn btn-sm btn-warning me-2" 
                    :disabled="product.stock === 0">
              <i class="bi bi-cart-plus"></i>
            </button>
            <router-link :to="'/product/' + product.id" class="btn btn-sm btn-info">
              <i class="bi bi-info-circle"></i>
            </router-link>
          </td>
        </tr>
      </tbody>
    </table>

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
import api from '@/services/api'
import { mapActions, mapGetters } from 'vuex'

export default {
  name: 'ProductsTableView',
  data() {
    return {
      products: [],
      currentPage: 1,
      totalPages: 1,
      total: 0
    }
  },
  computed: {
    ...mapGetters(['isAuthenticated'])
  },
  methods: {
    ...mapActions(['addToCart']),
    
    async fetchProducts(page = 1) {
      try {
        const response = await api.get(`/api/products?page=${page}&limit=10`)
        this.products = response.data.products
        this.currentPage = response.data.page
        this.totalPages = response.data.totalPages
        this.total = response.data.total
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        alert('Не удалось загрузить товары')
      }
    },
    
    changePage(page) {
      if (page >= 1 && page <= this.totalPages) {
        this.fetchProducts(page)
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
    }
  },
  mounted() {
    this.fetchProducts(this.currentPage)
  }
}
</script>