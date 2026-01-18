<template>
  <main class="container py-4">
    <button @click="$router.push('/products')" class="btn btn-outline-primary mb-4">
      ← Назад к товарам
    </button>

    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary"></div>
    </div>

    <div v-else-if="product" class="card">
      <div class="card-body">
        <div class="row g-4">
          <div class="col-md-5">
            <img :src="getImageUrl(product.image)" class="img-fluid rounded" :alt="product.name">
          </div>
          
          <div class="col-md-7">
            <div class="d-flex justify-content-between align-items-start mb-3">
              <h2 class="mb-0">{{ product.name }}</h2>
              <span :class="['badge', product.is_approved ? 'bg-success' : 'bg-warning']">
                {{ product.is_approved ? 'Одобрен' : 'На проверке' }}
              </span>
            </div>
            
            <div class="mb-4">
              <div class="mb-2">
                <span class="text-muted">Артикул:</span> {{ product.id }}
              </div>
              <div v-if="product.username" class="mb-2">
                <span class="text-muted">Продавец:</span> {{ product.username }}
              </div>
            </div>

            <div class="mb-4">
              <h5>Описание</h5>
              <div class="bg-light p-3 rounded">
                <pre class="mb-0" style="white-space: pre-wrap; font-family: inherit;">{{ product.description }}</pre>
              </div>
            </div>

            <div class="row mb-4">
              <div class="col-md-6">
                <div class="card">
                  <div class="card-body">
                    <h6>Наличие</h6>
                    <div :class="product.stock > 0 ? 'text-success' : 'text-danger'">
                      {{ product.stock > 0 ? `В наличии: ${product.stock} шт.` : 'Нет в наличии' }}
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-md-6">
                <div class="card">
                  <div class="card-body">
                    <h6>Цена</h6>
                    <div class="h4 text-primary mb-0">{{ formatPrice(product.price) }} ₽</div>
                  </div>
                </div>
              </div>
            </div>

            <button @click="addToCart(product)" class="btn btn-warning btn-lg" :disabled="product.stock === 0">
              В корзину
            </button>
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
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { auth } from '@/utils/auth'
import { apiRequest } from '@/utils/auth'

export default {
  name: 'ProductDetail',
  setup() {
    const route = useRoute()
    const router = useRouter()
    const product = ref(null)
    const loading = ref(true)

    const fetchProduct = async () => {
      loading.value = true
      try {
        const data = await apiRequest(`/api/products/${route.params.id}`)
        if (data.success) {
          product.value = data.data
        } else {
          router.push('/products')
        }
      } catch (error) {
        console.error('Ошибка загрузки товара:', error)
        router.push('/products')
      } finally {
        loading.value = false
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
          alert(`Товар "${product.name}" добавлен в корзину`)
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка добавления в корзину')
      }
    }

    const getImageUrl = (imageName) => `/img/${imageName}`

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    onMounted(() => {
      fetchProduct()
    })

    return {
      product,
      loading,
      addToCart,
      getImageUrl,
      formatPrice
    }
  }
}
</script>