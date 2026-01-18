<template>
  <main class="container py-4">
    <h1 class="mb-4">CatPC - Приветствуем вас 😸</h1>
    <p class="lead mb-4">Мы занимаемся продажей компьютерной техники с 2020 года.</p>
    <p>Рабочие станции, ноутбуки, мобильные телефоны, планшеты, неттопы, принтеры, МФУ и другое.</p>
    <p>Можем подобрать технику под ваши требования.</p>

    <div id="homeCarousel" class="carousel slide my-5">
      <div class="carousel-inner">
        <div class="carousel-item active">
          <img src="/img/Computer-set-2-_01___responsive_1920_1080.png" class="d-block w-100 rounded" alt="Компьютерный набор">
        </div>
        <div class="carousel-item">
          <img src="/img/Laptop.png" class="d-block w-100 rounded" alt="Ноутбук">
        </div>
        <div class="carousel-item">
          <img src="/img/Refurbished_Mobile_Phones_2.png" class="d-block w-100 rounded" alt="Мобильные телефоны">
        </div>
      </div>
      <button class="carousel-control-prev" type="button" data-bs-target="#homeCarousel" data-bs-slide="prev">
        <span class="carousel-control-prev-icon"></span>
      </button>
      <button class="carousel-control-next" type="button" data-bs-target="#homeCarousel" data-bs-slide="next">
        <span class="carousel-control-next-icon"></span>
      </button>
    </div>

    <h2 class="my-5">Отзывы клиентов</h2>
    <div class="row g-4">
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body">
            <div class="d-flex align-items-center mb-3">
              <div class="bg-primary rounded-circle text-white d-flex align-items-center justify-content-center me-3" style="width: 50px; height: 50px;">
                ИИ
              </div>
              <div>
                <h5 class="mb-0">Иван Иванов</h5>
                <small class="text-muted">15.03.2024</small>
              </div>
            </div>
            <p>"Отличный сервис! Быстро и качественно."</p>
            <div class="text-warning">★★★★★</div>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body">
            <div class="d-flex align-items-center mb-3">
              <div class="bg-success rounded-circle text-white d-flex align-items-center justify-content-center me-3" style="width: 50px; height: 50px;">
                ПС
              </div>
              <div>
                <h5 class="mb-0">Петр Сидоров</h5>
                <small class="text-muted">10.03.2024</small>
              </div>
            </div>
            <p>"Профессиональный подход и внимательное отношение."</p>
            <div class="text-warning">★★★★☆</div>
          </div>
        </div>
      </div>
      
      <div class="col-md-4">
        <div class="card h-100">
          <div class="card-body">
            <div class="d-flex align-items-center mb-3">
              <div class="bg-info rounded-circle text-white d-flex align-items-center justify-content-center me-3" style="width: 50px; height: 50px;">
                МК
              </div>
              <div>
                <h5 class="mb-0">Мария Ковалева</h5>
                <small class="text-muted">05.03.2024</small>
              </div>
            </div>
            <p>"Все сделали в срок, качество на высоте!"</p>
            <div class="text-warning">★★★★★</div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
export default {
  name: 'HomeView'
}
</script>