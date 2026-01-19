import { reactive, readonly } from 'vue'

// Реактивное состояние
const state = reactive({
  user: null,
  token: null
})

// Инициализируем из localStorage
const initFromStorage = () => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  
  state.token = token
  state.user = userStr ? JSON.parse(userStr) : null
}

// Инициализируем при загрузке
initFromStorage()

// Слушаем события изменения авторизации
window.addEventListener('storage', initFromStorage)
window.addEventListener('auth-change', initFromStorage)

export const auth = {
  // Сохранить данные пользователя
  login: (token, user) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    state.token = token
    state.user = user
    window.dispatchEvent(new CustomEvent('auth-change'))
  },
  
  // Выйти
  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    state.token = null
    state.user = null
    window.dispatchEvent(new CustomEvent('auth-change'))
  },
  
  // Обновить пользователя
  updateUser: (user) => {
    localStorage.setItem('user', JSON.stringify(user))
    state.user = user
    window.dispatchEvent(new CustomEvent('auth-change'))
  },
  
  // Получить токен
  getToken: () => state.token,
  
  // Получить пользователя
  getUser: () => state.user,
  
  // Проверить авторизацию
  isAuthenticated: () => !!state.token,
  
  // Проверить роль
  isSeller: () => {
    const user = state.user
    return user?.role === 'seller' || user?.role === 'admin'
  },
  
  // Проверить админа
  isAdmin: () => {
    const user = state.user
    return user?.role === 'admin'
  }
}

// Экспортируем реактивное состояние
export const authState = readonly(state)

// API функция с авторизацией
export const apiRequest = async (url, options = {}) => {
  const token = auth.getToken()
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers
  }
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  
  try {
    const response = await fetch(`http://localhost:1323${url}`, {
      ...options,
      headers
    })
    
    const data = await response.json()
    
    if (!response.ok) {
      throw new Error(data.error || `HTTP error! status: ${response.status}`)
    }
    
    return data
  } catch (error) {
    console.error('API Request error:', error)
    throw error
  }
}

// Добавьте эту функцию в auth.js
export const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('ru-RU')
}