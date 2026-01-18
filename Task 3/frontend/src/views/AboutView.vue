<template>
  <main class="container py-4">
    <h1 class="mb-4">О компании CatPC</h1>
    <p class="lead mb-4">Продаем технику физическим и юридическим лицам с 2020 года.</p>
    
    <h2 class="my-5">Наша команда</h2>
    <div class="row g-4">
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <router-link :to="'/employee/1'">
              <img src="/img/p1.jpeg" class="rounded-circle mb-3" alt="Директор" style="width: 150px; height: 150px; object-fit: cover;">
            </router-link>
            <h5 class="mb-1">Учредитель CatPC</h5>
            <p class="text-muted">Генеральный директор</p>
            <a :href="'mailto:CatPC@inbox.ru?body=' + encodeURIComponent('CatPC, здравствуйте, можно устроиться на работу.')" class="d-block">
              CatPC@inbox.ru
            </a>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <router-link :to="'/employee/2'">
              <img src="/img/p2.jpeg" class="rounded-circle mb-3" alt="Администратор" style="width: 150px; height: 150px; object-fit: cover;">
            </router-link>
            <h5 class="mb-1">Главный администратор</h5>
            <p class="text-muted">Отдел выдачи техники</p>
            <a :href="'mailto:CatPC@inbox.ru?body=' + encodeURIComponent('CatPC, когда я смогу забрать устройство?')" class="d-block">
              CatPC@inbox.ru
            </a>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body text-center">
            <router-link :to="'/employee/3'">
              <img src="/img/p3.jpeg" class="rounded-circle mb-3" alt="Маркетолог" style="width: 150px; height: 150px; object-fit: cover;">
            </router-link>
            <h5 class="mb-1">Маркетолог и менеджер</h5>
            <p class="text-muted">Отдел маркетинга и продаж</p>
            <a :href="'mailto:CatPC.Sales@inbox.ru?body=' + encodeURIComponent('CatPC, хочу заказать рабочую станцию для моих нужд по работе.')" class="d-block">
              CatPC.Sales@inbox.ru
            </a>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
export default {
  name: 'AboutView'
}
</script>