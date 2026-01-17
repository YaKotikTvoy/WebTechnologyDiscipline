import { createStore } from 'vuex'
import createPersistedState from 'vuex-persistedstate'
import api from '@/services/api' // Используем наш настроенный axios

export default createStore({

  plugins: [createPersistedState({
    storage: window.localStorage,
    reducer: (state) => ({
      user: state.user, // Сохраняем только пользователя
      cart: state.cart,
      cartCount: state.cartCount,
      cartTotal: state.cartTotal
    })
  })],
  
  mutations: {
    SET_USER(state, user) {
      state.user = user
    },
    
    CLEAR_USER(state) {
      state.user = null
    },
    
    SET_CART(state, cartData) {
      state.cart = cartData.items || []
      state.cartCount = cartData.count || 0
      state.cartTotal = cartData.total || 0
    },
    
    CLEAR_CART(state) {
      state.cart = []
      state.cartCount = 0
      state.cartTotal = 0
    }
  },
  
  actions: {
    async login({ commit }, credentials) {
      try {
        const response = await api.post('/api/login', credentials)
        commit('SET_USER', response.data.user)
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка авторизации' 
        }
      }
    },
    
    async register({ commit }, userData) {
      try {
        const response = await api.post('/api/register', userData)
        commit('SET_USER', response.data.user)
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка регистрации' 
        }
      }
    },
    
    async logout({ commit }) {
      try {
        await api.post('/api/logout')
      } catch (error) {
        console.error('Ошибка выхода:', error)
      }
      commit('CLEAR_USER')
      commit('CLEAR_CART')
    },
    
    async fetchProfile({ commit }) {
      console.log('🔄 Vuex: fetchProfile action вызван')
      try {
        const response = await api.get('/api/profile')
        console.log('✅ Vuex: Профиль загружен:', response.data)
        commit('SET_USER', response.data)
        return { success: true, data: response.data }
      } catch (error) {
        console.log('❌ Vuex: Ошибка загрузки профиля:', error.message)
        
        // Попробуем проверить через /api/check-auth
        try {
          const checkResponse = await api.get('/api/check-auth')
          console.log('🔍 Vuex: Проверка через check-auth:', checkResponse.data)
          
          if (checkResponse.data.authenticated) {
            commit('SET_USER', checkResponse.data.user)
            return { success: true, data: checkResponse.data.user }
          }
        } catch (checkError) {
          console.log('❌ Vuex: Ошибка check-auth:', checkError.message)
        }
        
        commit('CLEAR_USER')
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка загрузки профиля' 
        }
      }
    },
    
    async fetchCart({ commit }) {
      try {
        const response = await api.get('/api/cart')
        commit('SET_CART', response.data)
        return { success: true, data: response.data }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка загрузки корзины' 
        }
      }
    },
    
    async addToCart({ commit, dispatch }, { productId, quantity = 1 }) {
      try {
        console.log('🛒 Добавляем товар в корзину:', productId)
        const response = await api.post('/api/cart/add', {
          product_id: productId,
          quantity: quantity
        })
        
        console.log('✅ Ответ от сервера:', response.data)
        
        // Обновляем корзину после добавления
        await dispatch('fetchCart')
        
        return { success: true, data: response.data }
      } catch (error) {
        console.error('❌ Ошибка добавления в корзину:', error.response?.data)
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка добавления в корзину' 
        }
      }
    },
    
    async updateCartItem({ commit, dispatch }, { itemId, quantity }) {
      try {
        await api.put(`/api/cart/update/${itemId}`, { quantity })
        await dispatch('fetchCart')
        return { success: true }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка обновления корзины' 
        }
      }
    },
    
    async removeFromCart({ commit, dispatch }, itemId) {
      try {
        await api.delete(`/api/cart/remove/${itemId}`)
        await dispatch('fetchCart')
        return { success: true }
      } catch (error) {
        return { 
          success: false, 
          error: error.response?.data?.error || 'Ошибка удаления из корзины' 
        }
      }
    }
  },
  
  getters: {
    isAuthenticated: state => !!state.user,
    isAdmin: state => state.user?.role === 'admin',
    isSeller: state => state.user?.role === 'seller' || state.user?.role === 'admin',
    getUser: state => state.user,
    getCart: state => state.cart,
    getCartCount: state => state.cartCount,
    getCartTotal: state => state.cartTotal
  }
})