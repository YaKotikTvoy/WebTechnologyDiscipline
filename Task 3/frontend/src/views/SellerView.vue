<template>
  <main class="container py-4">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h1 class="mb-0">Панель продавца</h1>
      <button @click="showAddModal = true" class="btn btn-success">
        Добавить товар
      </button>
    </div>
    
    <div v-if="!isSeller" class="alert alert-warning">
      У вас нет прав продавца
    </div>
    
    <div v-else>
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'my' }]" @click="activeTab = 'my'">
            Мои товары
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'pending' }]" @click="activeTab = 'pending'">
            Ожидают одобрения
          </button>
        </li>
      </ul>
      
      <div v-if="activeTab === 'my'">
        <div v-if="products.length === 0" class="text-center py-5 text-muted">
          <h4>У вас пока нет товаров</h4>
          <button @click="showAddModal = true" class="btn btn-primary mt-3">
            Добавить первый товар
          </button>
        </div>
        
        <div v-else class="row g-4">
          <div v-for="product in products" :key="product.id" class="col-md-6 col-lg-4">
            <div class="card h-100">
              <div class="card-body">
                <div class="d-flex justify-content-between align-items-start mb-3">
                  <h5 class="mb-0">{{ product.name }}</h5>
                  <span :class="['badge', product.is_approved ? 'bg-success' : 'bg-warning']">
                    {{ product.is_approved ? 'Одобрен' : 'На проверке' }}
                  </span>
                </div>
                
                <p class="small text-muted mb-3">{{ truncateDescription(product.description, 80) }}</p>
                
                <div class="d-flex justify-content-between align-items-center mb-3">
                  <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                  <span :class="['badge', product.stock > 0 ? 'bg-success' : 'bg-danger']">
                    {{ product.stock }} шт.
                  </span>
                </div>
                
                <div class="d-flex gap-2">
                  <button @click="editProduct(product)" class="btn btn-outline-primary btn-sm">
                    Редактировать
                  </button>
                  <button @click="deleteProduct(product.id)" class="btn btn-outline-danger btn-sm">
                    Удалить
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div v-if="activeTab === 'pending'">
        <div v-if="pendingProducts.length === 0" class="text-center py-5 text-success">
          <h4>Все товары одобрены</h4>
        </div>
        
        <div v-else>
          <div class="alert alert-info mb-4">
            Эти товары ожидают проверки администратором
          </div>
          
          <div class="list-group">
            <div v-for="product in pendingProducts" :key="product.id" class="list-group-item">
              <div class="d-flex justify-content-between align-items-center">
                <div>
                  <h6 class="mb-1">{{ product.name }}</h6>
                  <small class="text-muted">{{ truncateDescription(product.description, 100) }}</small>
                </div>
                <div>
                  <button @click="editProduct(product)" class="btn btn-outline-primary btn-sm me-2">
                    Редактировать
                  </button>
                  <button @click="deleteProduct(product.id)" class="btn btn-outline-danger btn-sm">
                    Удалить
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div v-if="showAddModal || showEditModal" class="modal show d-block" style="background: rgba(0,0,0,0.5)">
        <div class="modal-dialog">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ editingProduct ? 'Редактировать товар' : 'Добавить товар' }}</h5>
              <button type="button" class="btn-close" @click="closeModal"></button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="saveProduct">
                <div class="mb-3">
                  <label class="form-label">Название</label>
                  <input type="text" class="form-control" v-model="productForm.name" required>
                </div>
                
                <div class="mb-3">
                  <label class="form-label">Описание</label>
                  <textarea class="form-control" v-model="productForm.description" rows="3" required></textarea>
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
                  <input type="text" class="form-control" v-model="productForm.image" placeholder="Имя файла" required>
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
import { ref, computed, onMounted } from 'vue'
import { useStore } from 'vuex'

export default {
  name: 'SellerView',
  setup() {
    const store = useStore()
    const activeTab = ref('my')
    const products = ref([])
    const pendingProducts = ref([])
    const loading = ref(false)
    const showAddModal = ref(false)
    const showEditModal = ref(false)
    const editingProduct = ref(null)
    const saving = ref(false)
    const productForm = ref({
      name: '',
      description: '',
      price: 0,
      stock: 0,
      image: ''
    })

    const user = computed(() => store.getters.getUser)
    const isSeller = computed(() => store.getters.isSeller)

    const fetchMyProducts = async () => {
      loading.value = true
      const token = localStorage.getItem('token')
      
      try {
        const response = await fetch('http://localhost:1323/api/seller/my-products', {
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
        if (data.success) {
          const allProducts = data.data || []
          products.value = allProducts.filter(p => p.is_approved)
          pendingProducts.value = allProducts.filter(p => !p.is_approved)
        }
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        alert('Не удалось загрузить товары')
      } finally {
        loading.value = false
      }
    }

    const editProduct = (product) => {
      editingProduct.value = product
      productForm.value = {
        name: product.name,
        description: product.description,
        price: product.price,
        stock: product.stock,
        image: product.image
      }
      showEditModal.value = true
    }

    const saveProduct = async () => {
      saving.value = true
      const token = localStorage.getItem('token')
      
      try {
        if (editingProduct.value) {
          const response = await fetch(`http://localhost:1323/api/seller/products/${editingProduct.value.id}`, {
            method: 'PUT',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(productForm.value)
          })
          const data = await response.json()
          if (data.success) {
            alert('Товар обновлен')
            closeModal()
            await fetchMyProducts()
          } else {
            alert(data.error || 'Ошибка')
          }
        } else {
          const response = await fetch('http://localhost:1323/api/seller/products', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(productForm.value)
          })
          const data = await response.json()
          if (data.success) {
            alert('Товар добавлен')
            closeModal()
            await fetchMyProducts()
          } else {
            alert(data.error || 'Ошибка')
          }
        }
      } catch (error) {
        alert('Ошибка сохранения')
      } finally {
        saving.value = false
      }
    }

    const deleteProduct = async (productId) => {
      if (!confirm('Удалить этот товар?')) return
      
      const token = localStorage.getItem('token')
      
      try {
        const response = await fetch(`http://localhost:1323/api/seller/products/${productId}`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
        if (data.success) {
          alert('Товар удален')
          await fetchMyProducts()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка удаления')
      }
    }

    const closeModal = () => {
      showAddModal.value = false
      showEditModal.value = false
      editingProduct.value = null
      productForm.value = {
        name: '',
        description: '',
        price: 0,
        stock: 0,
        image: ''
      }
    }

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    const truncateDescription = (text, length) => {
      if (!text) return ''
      return text.length > length ? text.substring(0, length) + '...' : text
    }

    onMounted(async () => {
      if (isSeller.value) {
        await fetchMyProducts()
      }
    })

    return {
      activeTab,
      products,
      pendingProducts,
      loading,
      showAddModal,
      showEditModal,
      editingProduct,
      saving,
      productForm,
      user,
      isSeller,
      editProduct,
      saveProduct,
      deleteProduct,
      closeModal,
      formatPrice,
      truncateDescription
    }
  }
}
</script>