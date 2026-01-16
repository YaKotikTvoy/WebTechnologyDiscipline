<template>
  <main class="container">
    <h1>Компания CatPC</h1>
    <p>Предоставляем самые высококачественные услуги в области продаж техники физическим лицам, а также юридическим, которым необходимо поддерживать КИИ🏬.</p>
    
    <h2>Команда CatPC</h2>
    <div class="row">
      <!-- Сотрудник 1 -->
      <div class="col-md-4">
        <div class="card shadow rounded my-3 employee-card">
          <div class="card-body">
            <div>
              <router-link :to="'/employee/1'">
                <img src="/img/p1.jpeg" class="avatar card-img-top" alt="Генеральный директор">
              </router-link>
            </div>
            <div class="avatar-post">
              <i class="bi-person"></i> Генеральный директор
            </div>
            <div class="avatar-text">Учредитель CatPC</div>
            <div class="avatar-text">
              <a href="mailto:CatPC@inbox.ru?body=CatPC, здравствуйте, можно устроиться на работу.">
                CatPC@inbox.ru
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Сотрудник 2 -->
      <div class="col-md-4">
        <div class="card shadow rounded my-3 employee-card">
          <div class="card-body">
            <div>
              <router-link :to="'/employee/2'">
                <img src="/img/p2.jpeg" class="avatar card-img-top" alt="Администратор">
              </router-link>
            </div>
            <div class="avatar-post">
              <i class="bi-person-check"></i> Главный администратор полки
            </div>
            <div class="avatar-text">Отдаёт устройства на руки, а также их обратно принимает, если что-то не так</div>
            <div class="avatar-text">
              <a href="mailto:CatPC@inbox.ru?body=CatPC, когда я смогу забрать устройство?">
                CatPC@inbox.ru
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Сотрудник 3 -->
      <div class="col-md-4">
        <div class="card shadow rounded my-3 employee-card">
          <div class="card-body">
            <div>
              <router-link :to="'/employee/3'">
                <img src="/img/p3.jpeg" class="avatar card-img-top" alt="Маркетолог">
              </router-link>
            </div>
            <div class="avatar-post">
              <i class="bi-star"></i> Маркетолог и менеджер-продаж
            </div>
            <div class="avatar-text">Занимается ведением медиа пространством магазина, также ищет клиентов.</div>
            <div class="avatar-text">
              <a href="mailto:CatPC.Sales@inbox.ru?body=CatPC, хочу заказать рабочую станцию для моих нужд по работе.">
                CatPC.Sales@inbox.ru
              </a>
            </div>
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