<template>
  <main class="container">
    <div class="row">
      <!-- Левая колонка с профилем -->
      <div class="col-md-4">
        <div class="card shadow">
          <div class="card-body text-center">
            <img :src="getEmployeeImage" class="rounded-circle mb-3" :alt="employee.name" style="width: 200px; height: 200px; object-fit: cover;">
            <h3>{{ employee.name }}</h3>
            <p class="text-muted">{{ employee.position }}</p>
            
            <div class="mt-4">
              <h5>Контакты</h5>
              <p>
                <i class="bi-envelope me-2"></i>
                <a :href="'mailto:' + employee.email + '?body=' + encodeURIComponent(employee.emailBody)">
                  {{ employee.email }}
                </a>
              </p>
              <p><i class="bi-telephone me-2"></i>{{ employee.phone }}</p>
            </div>
            
            <router-link to="/about" class="btn btn-outline-primary mt-3">
              <i class="bi-arrow-left me-2"></i>Назад к команде
            </router-link>
          </div>
        </div>
      </div>
      
      <!-- Правая колонка с информацией -->
      <div class="col-md-8">
        <div class="card shadow">
          <div class="card-body">
            <h2 class="card-title">О сотруднике</h2>
            <p class="lead">{{ employee.bio }}</p>
            
            <h4 class="mt-4">Обязанности</h4>
            <ul>
              <li v-for="(duty, index) in employee.duties" :key="index">{{ duty }}</li>
            </ul>
            
            <h4 class="mt-4">Достижения</h4>
            <ul>
              <li v-for="(achievement, index) in employee.achievements" :key="'ach' + index">
                {{ achievement }}
              </li>
            </ul>
            
            <h4 class="mt-4">Навыки</h4>
            <div class="mb-2">
              <span v-for="(skill, index) in employee.skills" :key="'skill' + index" 
                    class="badge bg-success me-1 mb-1">
                {{ skill }}
              </span>
            </div>
            
            <div v-if="employee.contactNote" class="mt-4">
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
            'Стратегическое планирование и развитие бизнеса',
            'Управление финансовыми потоками',
            'Разработка продуктовой стратегии',
            'Ведение переговоров с ключевыми партнерами'
          ],
          achievements: [
            'Открыл предприятие CatPC в 2020 году',
            'Привлек самых первых клиентов',
            'Собрал весь штат',
            'Заключил партнерские соглашения с ведущими поставщиками техники'
          ],
          skills: ['Стратегия предприятия', 'Аналитика']
        },
        2: {
          id: 2,
          name: 'Главный администратор полки',
          position: 'Отдел выдачи техники',
          bio: 'Отвечает за выдачу и прием техники, контроль качества обслуживания клиентов',
          email: 'CatPC@inbox.ru',
          emailBody: 'CatPC, когда я смогу забрать устройство?',
          phone: '+7 (965) 846-79-48',
          image: 'p2.jpeg',
          duties: [
            'Выдача техники клиентам',
            'Прием техники обратно при необходимости',
            'Контроль качества устройств перед выдачей',
            'Ведение документации по выданной технике',
            'Консультирование клиентов по техническим вопросам'
          ],
          achievements: [
            'Обработал более 1000 заказов без нареканий',
            'Внедрил систему проверки техники перед выдачей',
            'Снизил количество возвратов на 30%',
            'Получил множество положительных отзывов от клиентов'
          ],
          skills: ['Работа с клиентами', 'Знание техники', 'Документооборот', 'Контроль качества'],
          contactNote: 'Выдача техники, прием возвратов, консультация по устройствам'
        },
        3: {
          id: 3,
          name: 'Маркетолог и менеджер продаж',
          position: 'Отдел маркетинга и продаж',
          bio: 'Отвечает за продвижение компании и привлечение новых клиентов',
          email: 'CatPC.Sales@inbox.ru',
          emailBody: 'CatPC, хочу заказать рабочую станцию для моих нужд по работе.',
          phone: '+7 (923) 857-46-32',
          image: 'p3.jpeg',
          duties: [
            'Разработка и ведение медиа-пространства компании',
            'Поиск и привлечение новых клиентов',
            'Проведение маркетинговых кампаний',
            'Анализ рынка и конкурентов',
            'Ведение социальных сетей компании',
            'Консультирование клиентов по подбору техники'
          ],
          achievements: [
            'Увеличил поток клиентов на 40% за последний год',
            'Запустил успешные рекламные кампании в социальных сетях',
            'Привлек 5 корпоративных клиентов',
            'Разработал новую стратегию позиционирования компании'
          ],
          skills: ['Digital-маркетинг', 'SMM', 'Аналитика', 'Продажи', 'Копирайтинг'],
          contactNote: 'Заказ рабочей станции, консультация по подбору техники, сотрудничество, маркетинг'
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