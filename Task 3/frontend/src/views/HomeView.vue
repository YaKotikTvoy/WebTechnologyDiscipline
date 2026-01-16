<template>
  <main class="container">
    <h1>CatPC - Приветствует вас😸</h1>
    <p>Мы занимаемся тем, что продаем компьютерную технику - рабочие станции, ноутбуки, мобильные телефоны, планшеты, неттопы, принтеры, МФУ и другое =)</p>
    <p>Также можем подобрать для вас вариант техники, соответствующим вашим требованиям.</p>

    <!-- Карусель -->
    <div id="carouselExampleIndicators" class="carousel slide mb-5" data-bs-ride="carousel">
      <div class="carousel-indicators">
        <button type="button" data-bs-target="#carouselExampleIndicators" data-bs-slide-to="0" class="active"></button>
        <button type="button" data-bs-target="#carouselExampleIndicators" data-bs-slide-to="1"></button>
        <button type="button" data-bs-target="#carouselExampleIndicators" data-bs-slide-to="2"></button>
      </div>
      <div class="carousel-inner">
        <div class="carousel-item active">
          <img src="/img/Computer-set-2-_01___responsive_1920_1080.png" class="d-block w-100" alt="Computer set">
        </div>
        <div class="carousel-item">
          <img src="/img/Laptop.png" class="d-block w-100" alt="Laptop">
        </div>
        <div class="carousel-item">
          <img src="/img/Refurbished_Mobile_Phones_2.png" class="d-block w-100" alt="Phones">
        </div>
      </div>
      <button class="carousel-control-prev" type="button" data-bs-target="#carouselExampleIndicators" data-bs-slide="prev">
        <span class="carousel-control-prev-icon"></span>
      </button>
      <button class="carousel-control-next" type="button" data-bs-target="#carouselExampleIndicators" data-bs-slide="next">
        <span class="carousel-control-next-icon"></span>
      </button>
    </div>

    <!-- Отзывы -->
    <section class="py-5 bg-light">
      <div class="container">
        <h2 class="text-center mb-5">Отзывы наших клиентов</h2>
        <div class="row">
          <!-- Отзыв 1 -->
          <div class="col-md-4 mb-4">
            <div class="card h-100">
              <div class="card-body">
                <div class="d-flex align-items-center mb-3">
                  <div class="bg-primary rounded-circle d-flex align-items-center justify-content-center" 
                       style="width: 50px; height: 50px;">
                    <span class="text-white">ИИ</span>
                  </div>
                  <div class="ms-3">
                    <h5 class="card-title mb-0">Иван Иванов</h5>
                    <small class="text-muted">15.03.2024</small>
                  </div>
                </div>
                <p class="card-text">"Отличный сервис! Быстро и качественно выполнили работу. Обязательно буду рекомендовать друзьям."</p>
                <div class="text-warning">
                  ★★★★★
                </div>
              </div>
            </div>
          </div>

          <!-- Отзыв 2 -->
          <div class="col-md-4 mb-4">
            <div class="card h-100">
              <div class="card-body">
                <div class="d-flex align-items-center mb-3">
                  <div class="bg-success rounded-circle d-flex align-items-center justify-content-center" 
                       style="width: 50px; height: 50px;">
                    <span class="text-white">ПС</span>
                  </div>
                  <div class="ms-3">
                    <h5 class="card-title mb-0">Петр Сидоров</h5>
                    <small class="text-muted">10.03.2024</small>
                  </div>
                </div>
                <p class="card-text">"Очень доволен результатом. Профессиональный подход и внимательное отношение к деталям."</p>
                <div class="text-warning">
                  ★★★★☆
                </div>
              </div>
            </div>
          </div>
          
          <!-- Отзыв 3 -->
          <div class="col-md-4 mb-4">
            <div class="card h-100">
              <div class="card-body">
                <div class="d-flex align-items-center mb-3">
                  <div class="bg-info rounded-circle d-flex align-items-center justify-content-center" 
                       style="width: 50px; height: 50px;">
                    <span class="text-white">МК</span>
                  </div>
                  <div class="ms-3">
                    <h5 class="card-title mb-0">Мария Ковалева</h5>
                    <small class="text-muted">05.03.2024</small>
                  </div>
                </div>
                <p class="card-text">"Прекрасная работа! Все сделали в срок, даже раньше обещанного. Качество на высоте!"</p>
                <div class="text-warning">
                  ★★★★★
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<script>
export default {
  name: 'HomeView'
}
</script>