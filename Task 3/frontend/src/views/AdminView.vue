<template>
  <main class="container py-4">
    <h1 class="mb-4">Админ-панель</h1>
    
    <div v-if="!isAdmin" class="alert alert-danger">
      Доступ запрещен
    </div>
    
    <div v-else>
      <ul class="nav nav-tabs mb-4">
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'users' }]" @click="activeTab = 'users'; fetchUsers()">
            Пользователи
          </button>
        </li>
        <li class="nav-item">
          <button :class="['nav-link', { active: activeTab === 'pending' }]" @click="activeTab = 'pending'; fetchPendingProducts()">
            Товары на проверке
          </button>
        </li>
      </ul>
      
      <div v-if="activeTab === 'users'">
        <div v-if="loading" class="text-center py-5">
          <div class="spinner-border text-primary"></div>
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
                <th>Действия</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>{{ user.id }}</td>
                <td>{{ user.username }}</td>
                <td>{{ user.email }}</td>
                <td>
                  <select class="form-select form-select-sm" v-model="user.role" @change="updateUserRole(user.id, user.role)">
                    <option value="customer">Покупатель</option>
                    <option value="seller">Продавец</option>
                    <option value="admin">Администратор</option>
                  </select>
                </td>
                <td>
                  <span :class="['badge', user.is_active ? 'bg-success' : 'bg-danger']">
                    {{ user.is_active ? 'Активен' : 'Заблокирован' }}
                  </span>
                </td>
                <td>
                  <button @click="toggleUserActive(user.id, user.is_active)" class="btn btn-outline-warning btn-sm">
                    {{ user.is_active ? 'Заблокировать' : 'Разблокировать' }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      
      <div v-if="activeTab === 'pending'">
        <div v-if="pendingLoading" class="text-center py-5">
          <div class="spinner-border text-primary"></div>
        </div>
        
        <div v-else-if="!pendingProducts || pendingProducts.length === 0" class="text-center py-5 text-success">
          <h4>Нет товаров на проверке</h4>
        </div>
        
        <div v-else class="row g-4">
          <div v-for="product in pendingProducts" :key="product.id" class="col-md-6">
            <div class="card h-100">
              <div class="card-body">
                <h5 class="card-title">{{ product.name }}</h5>
                <p class="small text-muted mb-3">{{ product.description }}</p>
                
                <div class="mb-3">
                  <strong>Продавец:</strong> {{ product.username || 'Неизвестно' }}
                </div>
                
                <div class="d-flex justify-content-between align-items-center mb-3">
                  <span class="h5 text-primary">{{ formatPrice(product.price) }} ₽</span>
                  <span :class="['badge', product.stock > 0 ? 'bg-success' : 'bg-danger']">
                    {{ product.stock }} шт.
                  </span>
                </div>
                
                <div class="d-flex gap-2">
                  <button @click="approveProduct(product.id)" class="btn btn-success flex-grow-1">
                    Одобрить
                  </button>
                  <button @click="forceDeleteProduct(product.id)" class="btn btn-danger">
                    Удалить
                  </button>
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
import { useStore } from 'vuex'

export default {
  name: 'AdminView',
  setup() {
    const router = useRouter()
    const store = useStore()
    const activeTab = ref('users')
    const users = ref([])
    const pendingProducts = ref([])
    const loading = ref(false)
    const pendingLoading = ref(false)

    const currentUser = computed(() => store.getters.getUser)
    const isAdmin = computed(() => store.getters.isAdmin)

    const getToken = () => localStorage.getItem('token')

    const fetchUsers = async () => {
      loading.value = true
      const token = getToken()
      
      try {
        const response = await fetch('http://localhost:1323/api/admin/users', {
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
        if (data.success) {
          users.value = data.data || []
        }
      } catch (error) {
        console.error('Ошибка загрузки пользователей:', error)
        alert('Не удалось загрузить пользователей')
      } finally {
        loading.value = false
      }
    }

    const fetchPendingProducts = async () => {
      pendingLoading.value = true
      const token = getToken()
      
      try {
        const response = await fetch('http://localhost:1323/api/admin/pending-products', {
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
        if (data.success) {
          pendingProducts.value = data.data || []
        }
      } catch (error) {
        console.error('Ошибка загрузки товаров:', error)
        alert('Не удалось загрузить товары')
      } finally {
        pendingLoading.value = false
      }
    }

    const updateUserRole = async (userId, newRole) => {
      const token = getToken()
      
      try {
        const response = await fetch(`http://localhost:1323/api/admin/users/${userId}/role`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify({ role: newRole })
        })
        const data = await response.json()
        if (data.success) {
          alert('Роль обновлена')
        } else {
          alert(data.error || 'Ошибка')
        }
      } catch (error) {
        alert('Ошибка обновления роли')
      }
    }

    const toggleUserActive = async (userId, isActive) => {
      if (!confirm(`${isActive ? 'Заблокировать' : 'Разблокировать'} этого пользователя?`)) return
      
      const token = getToken()
      
      try {
        const response = await fetch(`http://localhost:1323/api/admin/users/${userId}/active`, {
          method: 'PUT',
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
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
      
      const token = getToken()
      
      try {
        const response = await fetch(`http://localhost:1323/api/admin/products/${productId}/approve`, {
          method: 'PUT',
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
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
      if (!confirm('Удалить этот товар?')) return
      
      const token = getToken()
      
      try {
        const response = await fetch(`http://localhost:1323/api/admin/products/${productId}/force`, {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${token}` }
        })
        const data = await response.json()
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

    const formatPrice = (price) => new Intl.NumberFormat('ru-RU').format(price)

    const checkAuth = () => {
      const token = getToken()
      const user = currentUser.value
      
      if (!token || !user) {
        alert('Для доступа к админ-панели необходимо войти в систему')
        router.push('/login')
        return false
      }
      
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
      fetchUsers,
      fetchPendingProducts,
      updateUserRole,
      toggleUserActive,
      approveProduct,
      forceDeleteProduct,
      formatPrice
    }
  }
}
</script>