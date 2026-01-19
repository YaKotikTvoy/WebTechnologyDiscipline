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
      
      <!-- Модальное окно добавления/редактирования товара -->
      <div v-if="showAddModal || showEditModal" class="modal show d-block" style="background: rgba(0,0,0,0.5); position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 1050;">
        <div class="modal-dialog modal-lg">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">{{ editingProduct ? 'Редактировать товар' : 'Добавить товар' }}</h5>
              <button type="button" class="btn-close" @click="closeModal"></button>
            </div>
            <div class="modal-body">
              <form id="productForm" @submit.prevent="saveProduct">
                <div class="mb-3">
                  <label class="form-label">Название <span class="text-danger">*</span></label>
                  <input type="text" class="form-control" v-model="productForm.name" required>
                </div>
                
                <div class="mb-3">
                  <label class="form-label">Описание <span class="text-danger">*</span></label>
                  <textarea class="form-control" v-model="productForm.description" rows="3" required></textarea>
                </div>
                
                <div class="row">
                  <div class="col-md-6 mb-3">
                    <label class="form-label">Цена (₽) <span class="text-danger">*</span></label>
                    <input type="number" class="form-control" v-model="productForm.price" min="0" step="0.01" required>
                  </div>
                  
                  <div class="col-md-6 mb-3">
                    <label class="form-label">Количество <span class="text-danger">*</span></label>
                    <input type="number" class="form-control" v-model="productForm.stock" min="0" required>
                  </div>
                </div>
                
                <div class="mb-3">
                  <label class="form-label">Изображение <span class="text-danger">*</span></label>
                  
                  <!-- Превью нового изображения -->
                  <div v-if="imagePreview" class="mb-2">
                    <div class="d-flex align-items-center">
                      <img :src="imagePreview" class="img-thumbnail" style="max-height: 150px; max-width: 150px; object-fit: contain;" alt="Превью">
                      <button type="button" @click="removeImagePreview" class="btn btn-sm btn-danger ms-2">
                        <i class="bi bi-x"></i> Удалить
                      </button>
                    </div>
                    <small class="text-muted">Новое изображение (будет загружено)</small>
                  </div>
                  
                  <!-- Текущее изображение при редактировании -->
                  <div v-else-if="editingProduct && productForm.image" class="mb-2">
                    <img :src="getImageUrl(productForm.image)" class="img-thumbnail" style="max-height: 150px; max-width: 150px; object-fit: contain;" alt="Текущее">
                    <small class="text-muted d-block mt-1">Текущее изображение</small>
                  </div>
                  
                  <!-- Поле выбора файла -->
                  <div class="mt-2">
                    <label for="imageUpload" class="btn btn-outline-secondary btn-sm">
                      <i class="bi bi-cloud-upload me-1"></i> Выбрать файл
                    </label>
                    <input type="file" 
                           class="form-control d-none" 
                           @change="handleImageUpload" 
                           accept="image/jpeg,image/png,image/gif,image/webp" 
                           ref="fileInput"
                           id="imageUpload">
                    <div class="form-text">
                      Поддерживаемые форматы: JPG, PNG, GIF, WEBP (макс. 10MB)
                    </div>
                  </div>
                  
                  <!-- Информация о выбранном файле -->
                  <div v-if="selectedFile" class="mt-2">
                    <div class="alert alert-info py-2">
                      <i class="bi bi-info-circle me-2"></i>
                      Выбран файл: <strong>{{ selectedFile.name }}</strong> 
                      ({{ formatFileSize(selectedFile.size) }})
                    </div>
                  </div>
                  
                  <!-- Скрытое поле для совместимости -->
                  <input type="hidden" v-model="productForm.image">
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
import { useRouter } from 'vue-router'
import { auth, authState, apiRequest } from '@/utils/auth'

export default {
  name: 'SellerView',
  setup() {
    const router = useRouter()
    const activeTab = ref('my')
    const products = ref([])
    const pendingProducts = ref([])
    const loading = ref(false)
    const showAddModal = ref(false)
    const showEditModal = ref(false)
    const editingProduct = ref(null)
    const saving = ref(false)
    const imagePreview = ref('')
    const selectedFile = ref(null)
    const fileInput = ref(null)
    
    const productForm = ref({
      name: '',
      description: '',
      price: 0,
      stock: 0,
      image: ''
    })

    const user = computed(() => authState.user)
    const isSeller = computed(() => auth.isSeller())

    // Функция для загрузки товаров продавца
    const fetchMyProducts = async () => {
      if (!isSeller.value) return
      
      loading.value = true
      try {
        const data = await apiRequest('/api/seller/my-products')
        if (data.success) {
          const allProducts = data.data || []
          products.value = allProducts.filter(p => p.is_approved)
          pendingProducts.value = allProducts.filter(p => !p.is_approved)
        }
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
      } finally {
        loading.value = false
      }
    }

    // Обработка выбора файла
    const handleImageUpload = (event) => {
      const file = event.target.files[0]
      if (!file) return

      // Проверка размера файла (10MB)
      if (file.size > 10 * 1024 * 1024) {
        alert('Файл слишком большой (максимум 10MB)')
        event.target.value = ''
        return
      }

      // Проверка типа файла
      const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
      if (!validTypes.includes(file.type)) {
        alert('Поддерживаются только изображения (JPG, PNG, GIF, WEBP)')
        event.target.value = ''
        return
      }

      selectedFile.value = file

      // Создаем превью
      const reader = new FileReader()
      reader.onload = (e) => {
        imagePreview.value = e.target.result
      }
      reader.readAsDataURL(file)
    }

    // Удаление превью
    const removeImagePreview = () => {
      imagePreview.value = ''
      selectedFile.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }

    // Форматирование размера файла
    const formatFileSize = (bytes) => {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    // Сохранение товара
    const saveProduct = async () => {
      if (!isSeller.value) return
      
      saving.value = true
      try {
        const token = localStorage.getItem('token')
        
        // Создаем FormData для отправки файла
        const formData = new FormData()
        formData.append('name', productForm.value.name)
        formData.append('description', productForm.value.description)
        formData.append('price', productForm.value.price.toString())
        formData.append('stock', productForm.value.stock.toString())
        
        // Если выбран новый файл, добавляем его
        if (selectedFile.value) {
          formData.append('image', selectedFile.value)
        } else if (productForm.value.image) {
          // Иначе используем существующее имя файла
          formData.append('image', productForm.value.image)
        }

        // Определяем URL и метод
        let url = 'http://localhost:1323/api/seller/products'
        let method = 'POST'

        if (editingProduct.value) {
          url = `http://localhost:1323/api/seller/products/${editingProduct.value.id}`
          method = 'PUT'
        }

        // Отправляем запрос
        const response = await fetch(url, {
          method: method,
          headers: {
            'Authorization': `Bearer ${token}`
            // НЕ добавляем Content-Type - браузер сам установит с boundary для FormData
          },
          body: formData
        })

        const data = await response.json()
        if (data.success) {
          alert(editingProduct.value ? 'Товар обновлен' : 'Товар добавлен')
          closeModal()
          await fetchMyProducts()
        } else {
          alert(data.error || 'Ошибка сохранения товара')
        }
      } catch (error) {
        console.error('Ошибка сохранения:', error)
        alert('Ошибка сохранения товара')
      } finally {
        saving.value = false
      }
    }

    // Редактирование товара
    const editProduct = (product) => {
      editingProduct.value = product
      productForm.value = {
        name: product.name,
        description: product.description,
        price: product.price,
        stock: product.stock,
        image: product.image
      }
      imagePreview.value = ''
      selectedFile.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
      showEditModal.value = true
    }

    // Удаление товара
    const deleteProduct = async (productId) => {
      if (!confirm('Удалить этот товар?')) return
      
      try {
        const data = await apiRequest(`/api/seller/products/${productId}`, {
          method: 'DELETE'
        })
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

    // Закрытие модального окна
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
      imagePreview.value = ''
      selectedFile.value = null
      if (fileInput.value) {
        fileInput.value.value = ''
      }
    }


    const getImageUrl = (imageName) => {
      if (!imageName) return ''
      if (imageName.startsWith('http') || imageName.startsWith('/img/')) {
        return imageName
      }
      return `/img/${imageName}`
    }

    // Форматирование цены
    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    // Обрезание описания
    const truncateDescription = (text, length) => {
      if (!text) return ''
      return text.length > length ? text.substring(0, length) + '...' : text
    }

    // Загрузка данных при монтировании
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
      imagePreview,
      selectedFile,
      fileInput,
      user,
      isSeller,
      editProduct,
      saveProduct,
      deleteProduct,
      closeModal,
      handleImageUpload,
      removeImagePreview,
      formatFileSize,
      getImageUrl,
      formatPrice,
      truncateDescription,
      fetchMyProducts
    }
  }
}
</script>

<style scoped>
.modal-backdrop {
  opacity: 0.5;
}

.img-thumbnail {
  background-color: #f8f9fa;
  border: 1px solid #dee2e6;
}

/* Анимация для превью */
.img-thumbnail {
  transition: transform 0.2s;
}

.img-thumbnail:hover {
  transform: scale(1.05);
}
</style>