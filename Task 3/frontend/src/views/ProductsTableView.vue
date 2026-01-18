<template>
  <main class="container py-4">
    <h1 class="mb-4">Полный список техники</h1>
    
    <div class="table-responsive">
      <table class="table table-hover">
        <thead>
          <tr>
            <th>Артикул</th>
            <th>Наименование</th>
            <th>Количество</th>
            <th>Цена</th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in products" :key="product.id">
            <td>{{ product.id }}</td>
            <td>{{ product.name }}</td>
            <td>{{ product.stock }} шт.</td>
            <td>{{ formatPrice(product.price) }} ₽</td>
            <td>
              <div class="d-flex gap-1">
                <button @click="addToCart(product)" class="btn btn-warning btn-sm" :disabled="product.stock === 0">
                  В корзину
                </button>
                <router-link :to="'/product/' + product.id" class="btn btn-outline-primary btn-sm">
                  Подробно
                </router-link>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
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
  </main>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { auth } from '@/utils/auth'
import { apiRequest } from '@/utils/auth'

export default {
  name: 'ProductsTableView',
  setup() {
    const router = useRouter()
    const products = ref([])
    const currentPage = ref(1)
    const totalPages = ref(1)
    const loading = ref(false)

    const fetchProducts = async (page = 1) => {
      loading.value = true
      try {
        const data = await apiRequest(`/api/products?page=${page}&limit=10`)
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

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

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
      formatPrice
    }
  }
}
</script>