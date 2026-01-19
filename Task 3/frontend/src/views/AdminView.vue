<template>
  <main class="container py-4">
    <h1 class="mb-4">Админ-панель CatPC</h1>
    
    <div v-if="!isAdmin" class="alert alert-danger">
      <i class="bi bi-shield-exclamation me-2"></i>
      Доступ запрещен. Только администратор CatPC имеет доступ к этой панели.
    </div>
    
    <div v-else>
      <!-- Уведомление для главного администратора -->
      <div v-if="isMainAdmin" class="alert alert-info mb-4">
        <i class="bi bi-shield-check me-2"></i>
        Вы - главный администратор CatPC. Только вы можете назначать других администраторов.
      </div>
      
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'users' }]" 
                  @click="activeTab = 'users'; fetchUsers()">
            Пользователи
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'pending' }]" 
                  @click="activeTab = 'pending'; fetchPendingProducts()">
            Товары на проверке
            <span v-if="pendingProducts.length > 0" class="badge bg-danger ms-1">
              {{ pendingProducts.length }}
            </span>
          </button>
        </li>
      </ul>
      
      <!-- Вкладка пользователей -->
      <div v-if="activeTab === 'users'">
        <div v-if="loading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Загрузка...</span>
          </div>
        </div>
        
        <div v-else class="table-responsive">
          <table class="table table-hover">
            <thead>
              <tr>
                <th>ID</th>
                <th>Имя</th>
                <th>Email</th>
                <th>Роль</th>
                <th>Статус</th>
                <th>Защита</th>
                <th>Дата регистрации</th>
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id" 
                  :class="{ 
                    'table-warning': user.username === 'CatPC',
                    'table-secondary': !user.is_active 
                  }">
                <td>{{ user.id }}</td>
                <td>
                  <strong>{{ user.username }}</strong>
                  <span v-if="user.username === 'CatPC'" class="badge bg-info ms-2">Главный</span>
                  <span v-if="user.id === currentUser?.id" class="badge bg-primary ms-2">Вы</span>
                </td>
                <td>{{ user.email }}</td>
                <td>
                  <select class="form-select form-select-sm" 
                          v-model="user.role" 
                          @change="updateUserRole(user.id, user.role, user.username, user.is_protected)"
                          :disabled="user.is_protected || (user.id === currentUser?.id && !isMainAdmin)">
                    <option value="customer">Покупатель</option>
                    <option value="seller">Продавец</option>
                    <option value="admin" :disabled="!isMainAdmin">Администратор</option>
                  </select>
                </td>
                <td>
                  <span class="badge" :class="user.is_active ? 'bg-success' : 'bg-danger'">
                    {{ user.is_active ? 'Активен' : 'Заблокирован' }}
                  </span>
                </td>
                <td>
                  <span v-if="user.is_protected" class="badge bg-warning">
                    <i class="bi bi-shield-lock"></i> Защищен
                  </span>
                  <span v-else class="text-muted">—</span>
                </td>
                <td>{{ formatDate(user.created_at) }}</td>
                <td>
                  <div class="btn-group btn-group-sm">
                    <button @click="toggleUserActive(user.id, user.is_active, user.username, user.is_protected)" 
                            class="btn btn-outline-warning"
                            :disabled="user.is_protected || user.id === currentUser?.id"
                            :title="user.is_protected ? 'Защищенный пользователь' : user.id === currentUser?.id ? 'Нельзя заблокировать себя' : ''">
                      <i class="bi" :class="user.is_active ? 'bi-lock' : 'bi-unlock'"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          
          <div v-if="users.length === 0" class="text-center py-5 text-muted">
            <i class="bi bi-people display-4 mb-3"></i>
            <h4>Нет пользователей</h4>
          </div>
          
          <div class="alert alert-info mt-3">
            <i class="bi bi-info-circle me-2"></i>
            <strong>Права администраторов:</strong><br>
            • <strong>Главный администратор (CatPC)</strong> - может назначать администраторов, менять все роли<br>
            • <strong>Обычные администраторы</strong> - могут менять роли только на "покупатель" или "продавец"<br>
            • <strong>Защищенные пользователи</strong> - нельзя изменить роль или заблокировать
          </div>
        </div>
      </div>
      
      <!-- Вкладка товаров на проверке -->
      <div v-if="activeTab === 'pending'">
        <div v-if="pendingLoading" class="text-center py-5">
          <div class="spinner-border text-primary" role="status">
            <span class="visually-hidden">Загрузка...</span>
          </div>
        </div>
        
        <div v-else-if="!pendingProducts || pendingProducts.length === 0" class="text-center py-5">
          <i class="bi bi-check-circle display-1 text-success mb-3"></i>
          <h4 class="text-success">Нет товаров на проверке</h4>
          <p class="text-muted">Все товары одобрены</p>
        </div>
        
        <div v-else>
          <div class="alert alert-info mb-4">
            <i class="bi bi-info-circle me-2"></i>
            Товары, добавленные продавцами, ожидают вашего одобрения. 
            Нажмите на товар для просмотра деталей.
          </div>
          
          <div class="row">
            <div v-for="product in pendingProducts" :key="product.id" class="col-md-6 col-lg-4 mb-4">
              <div class="card h-100 shadow-sm border-warning" 
                   @click="showProductPreview(product)"
                   style="cursor: pointer; transition: transform 0.2s;"
                   @mouseover="$event.currentTarget.style.transform = 'translateY(-5px)'"
                   @mouseout="$event.currentTarget.style.transform = 'translateY(0)'">
                
                <div class="card-img-top text-center bg-light p-3" style="height: 200px;">
                  <img :src="getImageUrl(product.image)" 
                       class="img-fluid h-100" 
                       style="object-fit: contain;"
                       :alt="product.name"
                       @error="handleImageError"
                       v-if="product.image">
                  <div v-else class="h-100 d-flex align-items-center justify-content-center">
                    <i class="bi bi-image display-4 text-muted"></i>
                  </div>
                </div>
                
                <div class="card-body">
                  <div class="d-flex justify-content-between align-items-start mb-2">
                    <h5 class="card-title mb-0 text-truncate">{{ product.name }}</h5>
                    <span class="badge bg-warning">Ожидает</span>
                  </div>
                  
                  <p class="card-text small text-muted mb-3" style="height: 60px; overflow: hidden;">
                    {{ product.description }}
                  </p>
                  
                  <div class="mb-3">
                    <small class="text-muted">
                      <i class="bi bi-person me-1"></i>
                      Продавец: <strong>{{ product.username || 'Неизвестно' }}</strong>
                    </small>
                  </div>
                  
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                    <span class="badge" :class="product.stock > 0 ? 'bg-success' : 'bg-danger'">
                      {{ product.stock }} шт.
                    </span>
                  </div>
                  
                  <div class="btn-group w-100">
                    <button @click.stop="approveProduct(product.id)" 
                            class="btn btn-success btn-sm">
                      <i class="bi bi-check-circle me-1"></i>Одобрить
                    </button>
                    <button @click.stop="forceDeleteProduct(product.id)" 
                            class="btn btn-danger btn-sm">
                      <i class="bi bi-trash me-1"></i>Удалить
                    </button>
                    <button @click.stop="showProductPreview(product)" 
                            class="btn btn-info btn-sm">
                      <i class="bi bi-eye me-1"></i>Просмотр
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Модальное окно предварительного просмотра товара -->
      <div v-if="showPreviewModal" class="modal show d-block" 
           style="background: rgba(0,0,0,0.7); position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 1060;">
        <div class="modal-dialog modal-lg modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header bg-primary text-white">
              <h5 class="modal-title">
                <i class="bi bi-eye me-2"></i>Предпросмотр товара
              </h5>
              <button type="button" class="btn-close btn-close-white" @click="closePreview"></button>
            </div>
            
            <div class="modal-body">
              <div v-if="previewProduct" class="row g-4">
                <div class="col-md-6">
                  <div class="card">
                    <div class="card-body text-center">
                      <img :src="getImageUrl(previewProduct.image)" 
                           class="img-fluid rounded" 
                           style="max-height: 300px; object-fit: contain;"
                           :alt="previewProduct.name"
                           @error="handleImageError"
                           v-if="previewProduct.image">
                      <div v-else class="py-5 text-muted">
                        <i class="bi bi-image display-1"></i>
                        <p class="mt-3">Изображение отсутствует</p>
                      </div>
                    </div>
                  </div>
                </div>
                
                <div class="col-md-6">
                  <div class="card">
                    <div class="card-body">
                      <h4 class="mb-3">{{ previewProduct.name }}</h4>
                      
                      <div class="mb-3">
                        <h6>Описание:</h6>
                        <div class="bg-light p-3 rounded">
                          <pre style="white-space: pre-wrap; font-family: inherit; margin: 0;">{{ previewProduct.description }}</pre>
                        </div>
                      </div>
                      
                      <div class="row mb-3">
                        <div class="col-6">
                          <div class="card bg-light">
                            <div class="card-body p-2">
                              <small class="text-muted">Цена</small>
                              <div class="h5 text-primary mb-0">{{ formatPrice(previewProduct.price) }} ₽</div>
                            </div>
                          </div>
                        </div>
                        <div class="col-6">
                          <div class="card bg-light">
                            <div class="card-body p-2">
                              <small class="text-muted">Наличие</small>
                              <div :class="['mb-0', previewProduct.stock > 0 ? 'text-success' : 'text-danger']">
                                {{ previewProduct.stock > 0 ? `${previewProduct.stock} шт.` : 'Нет в наличии' }}
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div class="mb-3">
                        <small class="text-muted">
                          <i class="bi bi-person me-1"></i>
                          Продавец: <strong>{{ previewProduct.username || 'Неизвестно' }}</strong>
                        </small>
                      </div>
                      
                      <div class="alert alert-warning">
                        <i class="bi bi-info-circle me-2"></i>
                        <small>Так товар будет выглядеть на прилавке после одобрения</small>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
            
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closePreview">
                Закрыть
              </button>
              <button type="button" class="btn btn-success" 
                      @click="approveProduct(previewProduct.id)" 
                      v-if="previewProduct">
                <i class="bi bi-check-circle me-1"></i>Одобрить товар
              </button>
              <button type="button" class="btn btn-danger" 
                      @click="forceDeleteProduct(previewProduct.id)" 
                      v-if="previewProduct">
                <i class="bi bi-trash me-1"></i>Удалить товар
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
  name: 'AdminView',
  setup() {
    const router = useRouter()
    const activeTab = ref('users')
    const users = ref([])
    const pendingProducts = ref([])
    const loading = ref(false)
    const pendingLoading = ref(false)
    const showPreviewModal = ref(false)
    const previewProduct = ref(null)

    const currentUser = computed(() => authState.user)
    const isAdmin = computed(() => auth.isAdmin())
    const isMainAdmin = computed(() => currentUser.value?.username === 'CatPC')

    // Функция для получения URL изображения
    const getImageUrl = (imageName) => {
      if (!imageName) return ''
      // Если это уже полный URL или путь
      if (imageName.startsWith('http') || imageName.startsWith('/img/')) {
        return imageName
      }
      // Иначе это имя файла - берем из backend
      return `http://localhost:1323/img/${imageName}`
    }

    // Обработка ошибок загрузки изображений
    const handleImageError = (event) => {
      event.target.src = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100" viewBox="0 0 100 100"><rect width="100" height="100" fill="%23f8f9fa"/><text x="50" y="50" font-family="Arial" font-size="14" fill="%236c757d" text-anchor="middle" dy=".3em">No image</text></svg>'
    }

    // Показ предварительного просмотра товара
    const showProductPreview = (product) => {
      previewProduct.value = product
      showPreviewModal.value = true
    }

    // Закрытие предварительного просмотра
    const closePreview = () => {
      showPreviewModal.value = false
      previewProduct.value = null
    }

    // Загрузка пользователей
    const fetchUsers = async () => {
      if (!isAdmin.value) return
      
      loading.value = true
      console.log('🟡 Загрузка пользователей...')
      
      try {
        const data = await apiRequest('/api/admin/users')
        console.log('📥 Получены данные пользователей:', data)
        
        if (data.success) {
          console.log(`✅ Успешно. Пользователей: ${data.data?.length || 0}`)
          users.value = data.data || []
        } else {
          console.error('❌ Ошибка от сервера:', data.error)
        }
      } catch (error) {
        console.error('❌ Ошибка загрузки пользователей:', error)
      } finally {
        loading.value = false
      }
    }

    // Загрузка товаров на проверке
    const fetchPendingProducts = async () => {
      if (!isAdmin.value) return
      
      pendingLoading.value = true
      try {
        const data = await apiRequest('/api/admin/pending-products')
        if (data.success) {
          pendingProducts.value = data.data || []
        }
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
      } finally {
        pendingLoading.value = false
      }
    }

    // Обновление роли пользователя
    const updateUserRole = async (userId, newRole, username, isProtected) => {
      // Проверка на защищенного пользователя
      if (isProtected) {
        alert('Нельзя изменить роль защищенного пользователя')
        await fetchUsers()
        return
      }
      
      // Проверка на назначение администратора
      if (newRole === 'admin' && !isMainAdmin.value) {
        alert('Только главный администратор CatPC может назначать администраторов')
        await fetchUsers()
        return
      }
      
      // Проверка на изменение самого себя
      if (userId === currentUser.value?.id) {
        if (!confirm('Вы уверены что хотите изменить свою роль?')) {
          await fetchUsers()
          return
        }
      }
      
      try {
        const data = await apiRequest(`/api/admin/users/${userId}/role`, {
          method: 'PUT',
          body: JSON.stringify({ role: newRole })
        })
        
        if (data.success) {
          // Если сервер вернул новый токен (при изменении своей роли)
          if (data.data?.new_token) {
            auth.login(data.data.new_token, data.data.user)
            alert('Роль обновлена. Новый токен сохранен.')
          } 
          // Если изменили свою роль, но токена нет в ответе
          else if (userId === currentUser.value?.id) {
            alert('Ваша роль изменена. Пожалуйста, войдите заново.')
            auth.logout()
            router.push('/login')
            return
          } 
          // Если изменили роль другому пользователю
          else {
            alert('Роль пользователя обновлена')
          }
          
          await fetchUsers()
        } else {
          alert(data.error || 'Ошибка')
          await fetchUsers()
        }
      } catch (error) {
        alert('Ошибка обновления роли')
        await fetchUsers()
      }
    }

    // Блокировка/разблокировка пользователя
    const toggleUserActive = async (userId, isActive, username, isProtected) => {
      // Проверка на защищенного пользователя
      if (isProtected) {
        alert('Нельзя заблокировать защищенного пользователя')
        return
      }
      
      // Проверка на самого себя
      if (userId === currentUser.value?.id) {
        alert('Нельзя заблокировать себя')
        return
      }
      
      if (!confirm(`${isActive ? 'Заблокировать' : 'Разблокировать'} пользователя "${username}"?`)) return
      
      try {
        const data = await apiRequest(`/api/admin/users/${userId}/active`, {
          method: 'PUT'
        })
        if (data.success) {
          alert(`Пользователь ${isActive ? 'заблокирован' : 'разблокирован'}`)
          await fetchUsers()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка изменения статуса')
      }
    }

    // Одобрение товара
    const approveProduct = async (productId) => {
      if (!confirm('Одобрить этот товар?')) return
      
      try {
        const data = await apiRequest(`/api/admin/products/${productId}/approve`, {
          method: 'PUT'
        })
        if (data.success) {
          alert('Товар одобрен')
          closePreview()
          await fetchPendingProducts()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка одобрения товара')
      }
    }

    // Принудительное удаление товара
    const forceDeleteProduct = async (productId) => {
      if (!confirm('Принудительно удалить этот товар? Это действие нельзя отменить.')) return
      
      try {
        const data = await apiRequest(`/api/admin/products/${productId}/force`, {
          method: 'DELETE'
        })
        if (data.success) {
          alert('Товар удален')
          closePreview()
          await fetchPendingProducts()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка удаления товара')
      }
    }

    // Форматирование цены
    const formatPrice = (price) => {
      return new Intl.NumberFormat('ru-RU').format(price)
    }

    // Форматирование даты
    const formatDate = (dateString) => {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    }

    // Проверка авторизации
    const checkAuth = () => {
      if (!isAdmin.value) {
        alert('Только администраторы имеют доступ к этой панели')
        router.push('/')
        return false
      }
      return true
    }

    onMounted(() => {
      if (checkAuth()) {
        fetchUsers()
      }
    })

    return {
      activeTab,
      users,
      pendingProducts,
      loading,
      pendingLoading,
      showPreviewModal,
      previewProduct,
      currentUser,
      isAdmin,
      isMainAdmin,
      fetchUsers,
      fetchPendingProducts,
      updateUserRole,
      toggleUserActive,
      showProductPreview,
      closePreview,
      approveProduct,
      forceDeleteProduct,
      getImageUrl,
      handleImageError,
      formatPrice,
      formatDate
    }
  }
}
</script>

<style scoped>
.card {
  transition: all 0.3s ease;
}

.card:hover {
  box-shadow: 0 5px 15px rgba(0,0,0,0.1);
}

.badge {
  font-size: 0.8em;
}

.modal {
  backdrop-filter: blur(5px);
}

.img-fluid {
  max-width: 100%;
  height: auto;
}

.table-warning {
  background-color: rgba(255, 193, 7, 0.1);
}

.table-secondary {
  background-color: rgba(108, 117, 125, 0.1);
}

.btn-group-sm > .btn {
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
}

.form-select-sm {
  padding: 0.25rem 2.25rem 0.25rem 0.5rem;
  font-size: 0.875rem;
}
</style>