<template>
  <main class="container">
    <h1 class="mb-4">Корзина покупок</h1>
    
    <div v-if="loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Загрузка...</span>
      </div>
    </div>
    
    <div v-else-if="cartItems.length === 0" class="text-center py-5">
      <i class="bi bi-cart-x display-1 text-muted mb-3"></i>
      <h3 class="text-muted">Корзина пуста</h3>
      <p>Добавьте товары из каталога</p>
      <router-link to="/products" class="btn btn-primary">
        <i class="bi bi-arrow-left me-2"></i>К каталогу
      </router-link>
    </div>
    
    <div v-else>
      <div class="row">
        <!-- Список товаров -->
        <div class="col-lg-8">
          <div class="card shadow-sm mb-4">
            <div class="card-body">
              <div v-for="item in cartItems" :key="item.id" class="cart-item mb-3 pb-3 border-bottom">
                <div class="row align-items-center">
                  <div class="col-3 col-md-2">
                    <img :src="getImageUrl(item.image)" class="img-fluid rounded" :alt="item.name">
                  </div>
                  
                  <div class="col-6 col-md-7">
                    <h5 class="card-title mb-1">{{ item.name }}</h5>
                    <p class="text-muted mb-2 small">Арт: {{ item.product_id }}</p>
                    <p class="h5 text-primary mb-0">{{ formatPrice(item.price) }} ₽</p>
                  </div>
                  
                  <div class="col-3 col-md-3">
                    <div class="d-flex align-items-center">
                      <button @click="updateQuantity(item.id, item.quantity - 1)" 
                              class="btn btn-outline-secondary btn-sm" 
                              :disabled="item.quantity <= 1">
                        <i class="bi bi-dash"></i>
                      </button>
                      
                      <span class="mx-2">{{ item.quantity }}</span>
                      
                      <button @click="updateQuantity(item.id, item.quantity + 1)" 
                              class="btn btn-outline-secondary btn-sm">
                        <i class="bi bi-plus"></i>
                      </button>
                      
                      <button @click="removeItem(item.id)" class="btn btn-outline-danger btn-sm ms-2">
                        <i class="bi bi-trash"></i>
                      </button>
                    </div>
                    
                    <div class="mt-2 text-end">
                      <small class="text-muted">
                        {{ formatPrice(item.price * item.quantity) }} ₽
                      </small>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Итого и оформление -->
        <div class="col-lg-4">
          <div class="card shadow-sm">
            <div class="card-body">
              <h5 class="card-title mb-4">Итого</h5>
              
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
              
              <button @click="checkout" class="btn btn-warning btn-lg w-100 mb-3">
                <i class="bi bi-bag-check me-2"></i>Оформить заказ
              </button>
              
              <router-link to="/products" class="btn btn-outline-primary w-100">
                <i class="bi bi-arrow-left me-2"></i>Продолжить покупки
              </router-link>
              
              <button @click="clearCart" class="btn btn-outline-danger w-100 mt-2">
                <i class="bi bi-cart-x me-2"></i>Очистить корзину
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {
  name: 'CartView',
  data() {
    return {
      loading: false
    }
  },
  computed: {
    ...mapGetters([
      'getCart',
      'getCartCount',
      'getCartTotal'
    ]),
    cartItems() {
      return this.getCart
    },
    cartCount() {
      return this.getCartCount
    },
    cartTotal() {
      return this.getCartTotal
    }
  },
  methods: {
    ...mapActions(['fetchCart', 'updateCartItem', 'removeFromCart']),
    
    async updateQuantity(itemId, newQuantity) {
      if (newQuantity < 1) return
      
      this.loading = true
      await this.updateCartItem({ itemId, quantity: newQuantity })
      this.loading = false
    },
    
    async removeItem(itemId) {
      if (!confirm('Удалить товар из корзины?')) return
      
      this.loading = true
      await this.removeFromCart(itemId)
      this.loading = false
    },
    
    async clearCart() {
      if (!confirm('Очистить всю корзину?')) return
      
      this.loading = true
      // Удаляем все товары по одному
      for (const item of this.cartItems) {
        await this.removeFromCart(item.id)
      }
      this.loading = false
    },
    
    checkout() {
      alert('Функция оформления заказа будет реализована позже')
    },
    
    formatPrice(price) {
      return new Intl.NumberFormat('ru-RU').format(price)
    },
    
    getImageUrl(imageName) {
      return `/img/${imageName}`
    }
  },
  async mounted() {
    this.loading = true
    await this.fetchCart()
    this.loading = false
  }
}
</script>

<style scoped>
.cart-item {
  transition: background-color 0.2s;
}

.cart-item:hover {
  background-color: #f8f9fa;
}

.img-fluid {
  max-height: 80px;
  object-fit: contain;
}
</style>