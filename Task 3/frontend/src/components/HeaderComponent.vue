<template>
  <header>
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
      <div class="container-fluid">
        <router-link class="navbar-brand" to="/">
          CatPC<img src="/img/771298.png" width="30" height="30">
        </router-link>
        
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
          <span class="navbar-toggler-icon"></span>
        </button>
        
        <div class="collapse navbar-collapse" id="navbarNav">
          <!-- Основное меню -->
          <ul class="navbar-nav me-auto">
            <li class="nav-item">
              <router-link class="nav-link" to="/" :class="{ active: $route.name === 'home' }">
                <i class="bi-house"></i> Главная
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/products" :class="{ active: $route.name === 'products' }">
                <i class="bi-pc-display"></i> Техника
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/about" :class="{ active: $route.name === 'about' }">
                <i class="bi-people"></i> О нас
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/products-table" :class="{ active: $route.name === 'products-table' }">
                <i class="bi-list-ul"></i> Полный список
              </router-link>
            </li>
            
            <!-- Меню для продавцов -->
            <li v-if="isSeller" class="nav-item">
              <router-link class="nav-link" to="/seller" :class="{ active: $route.name === 'seller' }">
                <i class="bi-shop"></i> Панель продавца
              </router-link>
            </li>
            
            <!-- Меню для администраторов -->
            <li v-if="isAdmin" class="nav-item">
              <router-link class="nav-link" to="/admin" :class="{ active: $route.name === 'admin' }">
                <i class="bi-gear"></i> Админ-панель
              </router-link>
            </li>
          </ul>
          
          <!-- Правая часть: корзина и профиль -->
          <ul class="navbar-nav">
            <!-- Корзина -->
            <li class="nav-item">
              <router-link class="nav-link position-relative" to="/cart" :class="{ active: $route.name === 'cart' }">
                <i class="bi-cart"></i> Корзина
                <span v-if="cartCount > 0" class="position-absolute top-0 start-100 translate-middle badge rounded-pill bg-danger">
                  {{ cartCount }}
                </span>
              </router-link>
            </li>
            
            <!-- Профиль / Авторизация -->
            <li v-if="!isAuthenticated" class="nav-item dropdown">
              <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                <i class="bi-person"></i> Войти
              </a>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <router-link class="dropdown-item" to="/login">
                    <i class="bi-box-arrow-in-right me-2"></i>Вход
                  </router-link>
                </li>
                <li>
                  <router-link class="dropdown-item" to="/register">
                    <i class="bi-person-plus me-2"></i>Регистрация
                  </router-link>
                </li>
              </ul>
            </li>
            
            <!-- Авторизованный пользователь -->
            <li v-else class="nav-item dropdown">
              <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                <i class="bi-person-circle me-1"></i>{{ user.username }}
                <span class="badge ms-1" :class="{
                  'bg-secondary': user.role === 'customer',
                  'bg-primary': user.role === 'seller',
                  'bg-danger': user.role === 'admin'
                }">
                  {{ userRoleText }}
                </span>
              </a>
              <ul class="dropdown-menu dropdown-menu-end">
                <li>
                  <router-link class="dropdown-item" to="/profile">
                    <i class="bi-person me-2"></i>Профиль
                  </router-link>
                </li>
                <li><hr class="dropdown-divider"></li>
                <li>
                  <a class="dropdown-item text-danger" href="#" @click.prevent="logout">
                    <i class="bi-box-arrow-right me-2"></i>Выйти
                  </a>
                </li>
              </ul>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  </header>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {
  name: 'HeaderComponent',
  computed: {
    ...mapGetters([
      'isAuthenticated',
      'isAdmin',
      'isSeller',
      'getUser',
      'getCartCount'
    ]),
    user() {
      return this.getUser || {}
    },
    cartCount() {
      return this.getCartCount
    },
    userRoleText() {
      const roles = {
        'customer': 'Покупатель',
        'seller': 'Продавец',
        'admin': 'Администратор'
      }
      return roles[this.user.role] || this.user.role
    }
  },
  methods: {
    ...mapActions(['logout']),
    async handleLogout() {
      await this.logout()
      this.$router.push('/')
    }
  }
}
</script>