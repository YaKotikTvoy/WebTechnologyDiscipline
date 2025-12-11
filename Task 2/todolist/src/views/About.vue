<template>
    <div class="container mt-5">
        <h1 class="mb-4">О себе</h1>

        <p>Привет! Меня зовут Артём, и я увлечённый разработчик, стремящийся создать универсальный редактор для проектирования и инженерных расчётов.</p>

        <h2 class="mt-4">Мой проект</h2>
        <p>На сегодняшний день существует множество средств для проектирования в различных предметных областях, таких как схемотехника, бизнес-процессы, ПО, CAD и CAE. Каждое из этих средств имеет свои нотации и ограничения. Моя цель — создать редактор, который сможет поддерживать множество нотаций и позволит пользователям проектировать и моделировать в различных предметных областях.</p>

        <p>Идея заключается в том, чтобы создать редактор, который может читать заранее запрограммированные нотации, а также предоставлять функционал для логики работы и отладки диаграмм. Это позволит пользователям не только проектировать, но и тестировать свои схемы, как будто они реализованы в действительности.</p>

        <p>Если вам интересна эта идея, я буду рад обсудить её с вами и услышать ваши мысли!</p>

        <h2 class="mt-4">Контактная информация</h2>
        <ul class="text-center">
            <li>
                <a href="https://github.com/YaKotikTvoy" target="_blank">
                    <img :src="githubAvatar" alt="GitHub Avatar" class="avatar" />
                    GitHub
                </a>
            </li>
            <li>
                <a href="https://gitverse.ru/YaKotikTvoy" target="_blank">
                    <img :src="gitverseAvatar" alt="GitVerse Avatar" class="avatar" />
                    GitVerse
                </a>
            </li>
        </ul>
    </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const githubAvatar = ref('');
const gitverseAvatar = ref('');

onMounted(async () => {
    try {
        const response = await fetch('https://api.github.com/users/YaKotikTvoy');
        const data = await response.json();
        githubAvatar.value = data.avatar_url;
    } catch (error) {
        console.error('Ошибка при получении аватара GitHub:', error);
    }

    try {
        gitverseAvatar.value = 'https://gitverse.ru/sc/avatars/fff76101ce7f26804f11f133482b1704e46ac03153284cc86b5e440fcbb75f69';
    } catch (error) {
        console.error('Ошибка при получении аватара GitVerse:', error);
    }
});
</script>

<style scoped>
.avatar {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    margin-right: 10px;
}

ul {
    list-style-type: none;
    padding: 0;
}

ul li {
    display: inline-block;
    margin: 10px;
}

ul li a {
    display: flex;
    flex-direction: column;
    align-items: center;
}
</style>