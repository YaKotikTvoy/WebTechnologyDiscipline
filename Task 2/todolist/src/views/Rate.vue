<template>
  <q-page class="q-pa-lg flex flex-center">
    <div class="text-center">
      <h2 class="q-mb-lg" style="color: #2c3e50;">Ваша оценка</h2>
      

      <div class="q-mb-md" style="font-size: 0;">
        <button
          v-for="star in 5"
          :key="star"
          @mouseenter="hoverRating = star"
          @mouseleave="hoverRating = 0"
          @click="setRating(star)"
          class="star-btn"
          :style="{
            fontSize: '3rem',
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: star <= currentRating || star <= hoverRating ? '#FFD700' : '#CCCCCC',
            transition: 'transform 0.2s, color 0.2s',
            margin: '0 5px',
            padding: '0'
          }"
        >
          ★
        </button>
      </div>

      <div v-if="currentRating > 0" class="q-mt-md">
        <div style="font-size: 2.5rem; font-weight: bold; color: #2c3e50;">
          {{ currentRating }}/5
        </div>
        <div style="font-size: 1.2rem; color: #7f8c8d; margin-top: 8px;">
          {{ getRatingText(currentRating) }}
        </div>
      </div>
      <div v-else class="q-mt-md" style="color: #95a5a6;">
        Выберите оценку
      </div>

      <button
        v-if="currentRating > 0"
        @click="resetRating"
        style="
          margin-top: 30px;
          padding: 8px 20px;
          background: #ecf0f1;
          border: none;
          border-radius: 20px;
          color: #7f8c8d;
          cursor: pointer;
          font-size: 1rem;
        "
      >
        Сбросить оценку
      </button>
    </div>
  </q-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const currentRating = ref(0);
const hoverRating = ref(0);

function setRating(rating: number) {
  currentRating.value = rating;
}

function resetRating() {
  currentRating.value = 0;
}

function getRatingText(rating: number): string {
  const ratingTexts: Record<number, string> = {
    1: 'Плохо',
    2: 'Неудовлетворительно',
    3: 'Удовлетворительно',
    4: 'Хорошо',
    5: 'Отлично'
  };
  return ratingTexts[rating] || '';
}
</script>

<style>
.star-btn:hover {
  transform: scale(1.2);
}
</style>