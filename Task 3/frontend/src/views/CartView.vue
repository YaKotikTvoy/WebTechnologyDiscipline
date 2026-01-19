<template>
  <main class="container py-4">
    <h1 class="mb-4">Корзина</h1>
    
    <div v-if="loading" class="text-center py-5">
      <div class="spinner-border text-primary"></div>
    </div>
    
    <div v-else-if="cartItems.length === 0" class="text-center py-5">
      <h4 class="text-muted mb-3">Корзина пуста</h4>
      <router-link to="/products" class="btn btn-primary">
        К каталогу
      </router-link>
    </div>
    
    <div v-else class="row g-4">
      <div class="col-lg-8">
        <div class="card">
          <div class="card-body">
            <div v-for="item in cartItems" :key="item.id" class="cart-item border-bottom py-3">
              <div class="row align-items-center">
                <div class="col-3">
                  <img :src="getImageUrl(item.image)" class="img-fluid rounded" :alt="item.name" style="height: 80px; object-fit: contain;">
                </div>
                
                <div class="col-6">
                  <h5 class="mb-1">{{ item.name }}</h5>
                  <div class="text-primary mb-2">{{ formatPrice(item.price) }} ₽</div>
                </div>
                
                <div class="col-3">
                  <div class="d-flex align-items-center justify-content-end">
                    <button @click="updateQuantity(item.id, item.quantity - 1)" class="btn btn-outline-secondary btn-sm" :disabled="item.quantity <= 1">
                      -
                    </button>
                    <span class="mx-2">{{ item.quantity }}</span>
                    <button @click="updateQuantity(item.id, item.quantity + 1)" class="btn btn-outline-secondary btn-sm">
                      +
                    </button>
                    <button @click="removeItem(item.id)" class="btn btn-outline-danger btn-sm ms-2">
                      ×
                    </button>
                  </div>
                  <div class="text-end mt-2">
                    <strong>{{ formatPrice(item.price * item.quantity) }} ₽</strong>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="col-lg-4">
        <div class="card">
          <div class="card-body">
            <h5 class="mb-4">Итого</h5>
            
            <div class="d-flex justify-content-between mb-2">
              <span>Товары ({{ cartCount }} шт.)</span>
              <span>{{ formatPrice(cartTotal) }} ₽</span>
            </div>
            
            <div class="d-flex justify-content-between mb-2">
              <span>Доставка</span>
              <span class="text-success">Бесплатно</span>
            </div>
            
            <hr>
            
            <div class="d-flex justify-content-between mb-4">
              <span class="h5">Общая сумма</span>
              <span class="h4 text-primary">{{ formatPrice(cartTotal) }} ₽</span>
            </div>
            
            <button @click="checkout" class="btn btn-warning w-100 mb-3">
              Оформить заказ
            </button>
            
            <button @click="clearCart" class="btn btn-outline-danger w-100">
              Очистить корзину
            </button>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { auth } from '@/utils/auth'
import { apiRequest } from '@/utils/auth'

export default {
  name: 'CartView',
  setup() {
    const cartItems = ref([])
    const loading = ref(false)

    const cartCount = computed(() => {
      return cartItems.value.reduce((sum, item) => sum + item.quantity, 0)
    })

    const cartTotal = computed(() => {
      return cartItems.value.reduce((sum, item) => sum + (item.price * item.quantity), 0)
    })

    const fetchCart = async () => {
      if (!auth.isAuthenticated()) return
      
      loading.value = true
      try {
        const data = await apiRequest('/api/cart')
        if (data.success) {
          cartItems.value = data.data.items || []
        }
      } catch (error) {
        console.error('Ошибка загрузки корзины:', error)
      } finally {
        loading.value = false
      }
    }

    const updateQuantity = async (itemId, newQuantity) => {
      if (newQuantity < 1) return
      
      loading.value = true
      try {
        const data = await apiRequest(`/api/cart/update/${itemId}`, {
          method: 'PUT',
          body: JSON.stringify({ quantity: newQuantity })
        })
        if (data.success) {
          await fetchCart()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка обновления корзины')
      } finally {
        loading.value = false
      }
    }

    const removeItem = async (itemId) => {
      if (!confirm('Удалить товар из корзины?')) return
      
      loading.value = true
      try {
        const data = await apiRequest(`/api/cart/remove/${itemId}`, {
          method: 'DELETE'
        })
        if (data.success) {
          await fetchCart()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка удаления товара')
      } finally {
        loading.value = false
      }
    }

    const clearCart = async () => {
      if (!confirm('Очистить всю корзину?')) return
      
      loading.value = true
      try {
        for (const item of cartItems.value) {
          await apiRequest(`/api/cart/remove/${item.id}`, {
            method: 'DELETE'
          })
        }
        await fetchCart()
      } catch (error) {
        alert('Ошибка очистки корзины')
      } finally {
        loading.value = false
      }
    }

    const checkout = () => {
      alert('Функция оформления заказа будет реализована позже')
    }

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    const getImageUrl = (imageName) => {
      if (!imageName) return ''
      
      if (imageName.startsWith('http') || imageName.startsWith('/img/')) {
        return imageName
      }
      
      return `http://localhost:1323/img/${imageName}`
    }

    onMounted(() => {
      fetchCart()
    })

    return {
      cartItems,
      loading,
      cartCount,
      cartTotal,
      updateQuantity,
      removeItem,
      clearCart,
      checkout,
      formatPrice,
      getImageUrl
    }
  }
}
</script>