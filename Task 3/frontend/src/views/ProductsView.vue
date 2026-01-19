<template>
  <main class="container py-4">
    <h1 class="mb-4">Каталог техники</h1>
    
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary"></div>
    </div>
    
    <div v-else>
      <div class="row g-4">
        <div v-for="product in products" :key="product.id" class="col-md-6">
          <div class="card h-100">
            <div class="row g-0">
              <div class="col-md-4">
                <img :src="getImageUrl(product.image)" class="img-fluid rounded-start" :alt="product.name" style="height: 200px; object-fit: cover;">
              </div>
              <div class="col-md-8">
                <div class="card-body d-flex flex-column h-100">
                  <h5 class="card-title">{{ product.name }}</h5>
                  <p class="card-text small text-muted">{{ truncateDescription(product.description, 100) }}</p>
                  
                  <div class="mt-auto">
                    <div class="d-flex justify-content-between align-items-center mb-3">
                      <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                      <span :class="['badge', product.stock > 0 ? 'bg-success' : 'bg-danger']">
                        {{ product.stock > 0 ? 'В наличии' : 'Нет в наличии' }}
                      </span>
                    </div>
                    
                    <div class="d-flex gap-2">
                      <button @click="addToCart(product)" class="btn btn-warning btn-sm" :disabled="product.stock === 0">
                        В корзину
                      </button>
                      <router-link :to="'/product/' + product.id" class="btn btn-outline-primary btn-sm">
                        Подробнее
                      </router-link>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <nav v-if="totalPages > 1" class="mt-4">
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
    </div>
  </main>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from '@/utils/auth'
import { apiRequest } from '@/utils/auth'

export default {
  name: 'ProductsView',
  setup() {
    const router = useRouter()
    const products = ref([])
    const currentPage = ref(1)
    const totalPages = ref(1)
    const loading = ref(false)

    const fetchProducts = async (page = 1) => {
      loading.value = true
      try {
        const data = await apiRequest(`/api/products?page=${page}&limit=6`)
        if (data.success) {
          products.value = data.data.products
          currentPage.value = data.data.page
          totalPages.value = data.data.totalPages
        }
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
      } finally {
        loading.value = false
      }
    }

    const changePage = (page) => {
      if (page >= 1 && page <= totalPages.value && page !== currentPage.value) {
        fetchProducts(page)
      }
    }

    const addToCart = async (product) => {
      if (product.stock === 0) {
        alert('Товар отсутствует')
        return
      }
      
      if (!auth.isAuthenticated()) {
        if (confirm('Для добавления в корзину нужно войти. Перейти на страницу входа?')) {
          router.push('/login')
        }
        return
      }
      
      try {
        const data = await apiRequest('/api/cart/add', {
          method: 'POST',
          body: JSON.stringify({
            product_id: product.id,
            quantity: 1
          })
        })
        
        if (data.success) {
          alert(`Товар "${product.name}" добавлен в корзину!`)
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка соединения')
      }
    }

    const getImageUrl = (imageName) => {
      if (!imageName) return ''
      
      if (imageName.startsWith('http') || imageName.startsWith('/img/')) {
        return imageName
      }
      
      return `http://localhost:1323/img/${imageName}`
    }

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    const truncateDescription = (text, length) => {
      if (!text) return ''
      return text.length > length ? text.substring(0, length) + '...' : text
    }

    onMounted(() => {
      fetchProducts(currentPage.value)
    })

    return {
      products,
      currentPage,
      totalPages,
      loading,
      changePage,
      addToCart,
      getImageUrl,
      formatPrice,
      truncateDescription
    }
  }
}
</script>