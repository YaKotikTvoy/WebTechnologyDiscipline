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
          <ul class="navbar-nav">
            <li class="nav-item">
              <router-link class="nav-link" to="/" exact-active-class="active">
                <i class="bi-house"></i> Главная
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/products" active-class="active">
                <i class="bi-pc-display"></i> Техника
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/about" active-class="active">
                <i class="bi-people"></i> О нас
              </router-link>
            </li>
            <li class="nav-item">
              <router-link class="nav-link" to="/products-table" active-class="active">
                <i class="bi-pc-display"></i> Полный список техники
              </router-link>
            </li>
          </ul>
        </div>
      </div>
    </nav>
  </header>
</template>

<script>
export default {
  name: 'HeaderComponent'
}
</script>