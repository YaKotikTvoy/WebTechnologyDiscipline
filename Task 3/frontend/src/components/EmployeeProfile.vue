<template>
  <main class="container py-4">
    <div class="row g-4">
      <div class="col-md-4">
        <div class="card">
          <div class="card-body text-center">
            <img :src="getEmployeeImage" class="rounded-circle mb-3" :alt="employee.name" style="width: 150px; height: 150px; object-fit: cover;">
            <h3>{{ employee.name }}</h3>
            <p class="text-muted">{{ employee.position }}</p>
            
            <div class="mt-4">
              <h5>Контакты</h5>
              <p>
                <a :href="'mailto:' + employee.email + '?body=' + encodeURIComponent(employee.emailBody)">
                  {{ employee.email }}
                </a>
              </p>
              <p>{{ employee.phone }}</p>
            </div>
            
            <router-link to="/about" class="btn btn-outline-primary mt-3">
              Назад к команде
            </router-link>
          </div>
        </div>
      </div>
      
      <div class="col-md-8">
        <div class="card">
          <div class="card-body">
            <h2 class="mb-3">О сотруднике</h2>
            <p class="lead">{{ employee.bio }}</p>
            
            <h4 class="mt-4">Обязанности</h4>
            <ul>
              <li v-for="(duty, index) in employee.duties" :key="index">{{ duty }}</li>
            </ul>
            
            <h4 class="mt-4">Достижения</h4>
            <ul>
              <li v-for="(achievement, index) in employee.achievements" :key="'ach' + index">{{ achievement }}</li>
            </ul>
            
            <h4 class="mt-4">Навыки</h4>
            <div class="mb-3">
              <span v-for="(skill, index) in employee.skills" :key="'skill' + index" class="badge bg-success me-1 mb-1">
                {{ skill }}
              </span>
            </div>
            
            <div v-if="employee.contactNote">
              <h4>Связаться по вопросам</h4>
              <p>{{ employee.contactNote }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
export default {
  name: 'EmployeeProfile',
  data() {
    return {
      employees: {
        1: {
          id: 1,
          name: 'Учредитель CatPC',
          position: 'Генеральный директор',
          bio: 'Отвечает за стратегическое продвижение компании',
          email: 'CatPC@inbox.ru',
          emailBody: 'CatPC, здравствуйте, можно устроиться на работу.',
          phone: '+7 (920) 928-09-67',
          image: 'p1.jpeg',
          duties: [
            'Общее руководство компанией',
            'Стратегическое планирование',
            'Управление финансами',
            'Разработка стратегии',
            'Переговоры с партнерами'
          ],
          achievements: [
            'Открыл предприятие CatPC в 2020 году',
            'Привлек первых клиентов',
            'Собрал штат',
            'Заключил партнерские соглашения'
          ],
          skills: ['Стратегия', 'Аналитика']
        },
        2: {
          id: 2,
          name: 'Главный администратор',
          position: 'Отдел выдачи техники',
          bio: 'Отвечает за выдачу и прием техники',
          email: 'CatPC@inbox.ru',
          emailBody: 'CatPC, когда я смогу забрать устройство?',
          phone: '+7 (965) 846-79-48',
          image: 'p2.jpeg',
          duties: [
            'Выдача техники клиентам',
            'Прием техники обратно',
            'Контроль качества',
            'Ведение документации',
            'Консультирование'
          ],
          achievements: [
            'Обработал более 1000 заказов',
            'Внедрил систему проверки',
            'Снизил возвраты на 30%',
            'Получил положительные отзывы'
          ],
          skills: ['Работа с клиентами', 'Знание техники', 'Документооборот'],
          contactNote: 'Выдача техники, прием возвратов, консультация'
        },
        3: {
          id: 3,
          name: 'Маркетолог',
          position: 'Отдел маркетинга и продаж',
          bio: 'Отвечает за продвижение компании',
          email: 'CatPC.Sales@inbox.ru',
          emailBody: 'CatPC, хочу заказать рабочую станцию',
          phone: '+7 (923) 857-46-32',
          image: 'p3.jpeg',
          duties: [
            'Ведение медиа-пространства',
            'Поиск клиентов',
            'Маркетинговые кампании',
            'Анализ рынка',
            'Социальные сети'
          ],
          achievements: [
            'Увеличил поток клиентов на 40%',
            'Запустил рекламные кампании',
            'Привлек корпоративных клиентов',
            'Разработал стратегию'
          ],
          skills: ['Digital-маркетинг', 'SMM', 'Аналитика', 'Продажи'],
          contactNote: 'Заказ техники, консультация, маркетинг'
        }
      }
    }
  },
  computed: {
    employeeId() {
      return parseInt(this.$route.params.id) || 1
    },
    employee() {
      return this.employees[this.employeeId] || this.employees[1]
    },
    getEmployeeImage() {
      return `/img/${this.employee.image}`
    }
  }
}
</script>