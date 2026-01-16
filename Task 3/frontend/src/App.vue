<template>
  <div id="app">
    <HeaderComponent />
    <router-view />
    <footer class="bg-light text-center text-lg-start mt-5">
      <div class="container p-4">
        <div class="row">
          <div class="col-lg-6 col-md-12 mb-4 mb-md-0">
            <h5 class="text-uppercase">CatPC</h5>
            <p>Продажа компьютерной техники с 2020 года</p>
          </div>
          <div class="col-lg-6 col-md-12 mb-4 mb-md-0">
            <h5 class="text-uppercase">Контакты</h5>
            <p><i class="bi-telephone"></i> +7 (900) 123-45-67</p>
            <p><i class="bi-envelope"></i> info@catpc.ru</p>
          </div>
        </div>
      </div>
      <div class="text-center p-3 bg-dark text-white">
        © 2024 CatPC. Все права защищены.
      </div>
    </footer>
  </div>
</template>

<script>
import HeaderComponent from './components/HeaderComponent.vue'

export default {
  name: 'App',
  components: {
    HeaderComponent
  }
}
</script>

<style>
#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.router-view {
  flex: 1;
}

footer {
  margin-top: auto;
}

.nav-link.active {
  font-weight: bold;
  color: #0d6efd !important;
}
</style>