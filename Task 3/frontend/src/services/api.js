import axios from 'axios'

// Создаем экземпляр axios с настройками
const api = axios.create({
  baseURL: 'http://localhost:1323',
  withCredentials: true, // ← ВАЖНО: отправляем cookies
  headers: {
    'Content-Type': 'application/json',
  }
})

// Перехватчик для ошибок
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      // Если не авторизован, перенаправляем на логин
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default api