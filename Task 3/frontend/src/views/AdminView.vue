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
            Товары, добавленные продавцами, ожидают вашего одобрения
          </div>
          
          <div class="row">
            <div v-for="product in pendingProducts" :key="product.id" class="col-md-6 mb-4">
              <div class="card h-100 shadow-sm">
                <div class="card-body">
                  <div class="d-flex justify-content-between align-items-start mb-3">
                    <h5 class="card-title mb-0">{{ product.name }}</h5>
                    <span class="badge bg-warning">Ожидает</span>
                  </div>
                  
                  <p class="card-text small mb-3">{{ product.description }}</p>
                  
                  <div class="mb-3">
                    <strong>Продавец:</strong> {{ product.username || 'Неизвестно' }}
                  </div>
                  
                  <div class="d-flex justify-content-between align-items-center mb-3">
                    <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                    <span class="badge" :class="product.stock > 0 ? 'bg-success' : 'bg-danger'">
                      {{ product.stock }} шт.
                    </span>
                  </div>
                  
                  <div class="btn-group w-100">
                    <button @click="approveProduct(product.id)" class="btn btn-success">
                      <i class="bi bi-check-circle me-2"></i>Одобрить
                    </button>
                    <button @click="forceDeleteProduct(product.id)" class="btn btn-danger">
                      <i class="bi bi-trash me-2"></i>Удалить
                    </button>
                  </div>
                </div>
              </div>
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
import { auth, authState } from '@/utils/auth'
import { apiRequest } from '@/utils/auth'
import { formatDate } from '@/utils/auth'

export default {
  name: 'AdminView',
  setup() {
    const router = useRouter()
    const activeTab = ref('users')
    const users = ref([])
    const pendingProducts = ref([])
    const loading = ref(false)
    const pendingLoading = ref(false)

    const currentUser = computed(() => authState.user)
    const isAdmin = computed(() => auth.isAdmin())
    const isMainAdmin = computed(() => currentUser.value?.username === 'CatPC')

    const fetchUsers = async () => {
      if (!isAdmin.value) return
      
      loading.value = true
      try {
        const data = await apiRequest('/api/admin/users')
        if (data.success) {
          users.value = data.data || []
        }
      } catch (error) {
        console.error('Ошибка загрузки пользователей:', error)
      } finally {
        loading.value = false
      }
    }

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
      if (userId === currentUser.value?.id && newRole !== 'admin' && currentUser.value?.role === 'admin') {
        if (!confirm('Вы уверены что хотите снять с себя права администратора?')) {
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
          alert('Роль пользователя обновлена')
          
          // Если изменили свою роль - обновляем данные
          if (currentUser.value && currentUser.value.id === userId) {
            const profileData = await apiRequest('/api/profile')
            if (profileData.success) {
              auth.updateUser(profileData.data)
              alert('Ваша роль изменена. Обновите страницу.')
            }
          }
        } else {
          alert(data.error || 'Ошибка')
          await fetchUsers()
        }
      } catch (error) {
        alert('Ошибка обновления роли')
        await fetchUsers()
      }
    }

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

    const approveProduct = async (productId) => {
      if (!confirm('Одобрить этот товар?')) return
      
      try {
        const data = await apiRequest(`/api/admin/products/${productId}/approve`, {
          method: 'PUT'
        })
        if (data.success) {
          alert('Товар одобрен')
          await fetchPendingProducts()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка одобрения товара')
      }
    }

    const forceDeleteProduct = async (productId) => {
      if (!confirm('Принудительно удалить этот товар? Это действие нельзя отменить.')) return
      
      try {
        const data = await apiRequest(`/api/admin/products/${productId}/force`, {
          method: 'DELETE'
        })
        if (data.success) {
          alert('Товар удален')
          await fetchPendingProducts()
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка удаления товара')
      }
    }

    const formatPrice = (price) => {
      return new Intl.NumberFormat('ru-RU').format(price)
    }

    const formatDate = (dateString) => {
      if (!dateString) return '-'
      return new Date(dateString).toLocaleDateString('ru-RU')
    }

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
      currentUser,
      isAdmin,
      isMainAdmin,
      fetchUsers,
      fetchPendingProducts,
      updateUserRole,
      toggleUserActive,
      approveProduct,
      forceDeleteProduct,
      formatPrice,
      formatDate
    }
  }
}
</script>