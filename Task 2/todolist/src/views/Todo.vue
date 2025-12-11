<template>
    <div class="container mt-5">
        <h1 class="mb-4">Редактор заметок</h1>

        <!-- Отображение существующих заметок -->
        <div v-for="(note, i) in notes" :key="i" class="card mb-4 shadow-sm">
            <div class="card-body">
                <div class="d-flex justify-content-between align-items-start mb-3">
                    <div v-if="!editingNote[i]">
                        <h5 class="card-title mb-0">{{ note.title }}</h5>
                        <p class="text-muted small mb-0 mt-1">Задач: {{ note.todos.length }}</p>
                    </div>
                    <div v-else class="flex-grow-1">
                        <input type="text" v-model="editingNoteTitle[i]" class="form-control" placeholder="Название заметки"/>
                    </div>
                    
                    <div class="btn-group ms-3">
                        <button v-if="!editingNote[i]" 
                                @click="startEditingNote(i)" 
                                class="btn btn-outline-primary btn-sm">
                            ✏️
                        </button>
                        <template v-else>
                            <button @click="saveNote(i)" class="btn btn-success btn-sm">✔️</button>
                            <button @click="cancelEditingNote(i)" class="btn btn-secondary btn-sm">✕</button>
                        </template>
                    </div>
                </div>

                <!-- Список задач -->
                <ul class="list-group list-group-flush mb-3">
                    <li v-for="(todo, j) in note.todos" :key="j" 
                        class="list-group-item d-flex align-items-center py-2">
                        <div v-if="!editingTodo[i]?.[j]" class="d-flex justify-content-between align-items-center w-100">
                            <div class="d-flex align-items-center">
                                <input type="checkbox" 
                                       class="form-check-input me-3" 
                                       @change="toggleTodo(i, j)" 
                                       :checked="todo.completed"/>
                                <span :class="{ 
                                    'text-decoration-line-through text-muted': todo.completed,
                                    'fw-bold': !todo.completed
                                }">
                                    {{ todo.text }}
                                </span>
                            </div>
                            <div class="btn-group">
                                <button @click="startEditingTodo(i, j)" 
                                        class="btn btn-outline-secondary btn-sm">
                                    ✏️
                                </button>
                                <button @click="removeTodo(i, j)" 
                                        class="btn btn-outline-danger btn-sm">
                                    🗑️
                                </button>
                            </div>
                        </div>
                        <div v-else class="d-flex w-100 align-items-center">
                            <input type="text" 
                                   v-model="editingTodoText[i][j]" 
                                   class="form-control form-control-sm me-2" 
                                   placeholder="Текст задачи"
                                   @keyup.enter="saveTodo(i, j)"
                                   ref="todoInput"/>
                            <div class="btn-group">
                                <button @click="saveTodo(i, j)" class="btn btn-success btn-sm">✔️</button>
                                <button @click="cancelEditingTodo(i, j)" class="btn btn-secondary btn-sm">✕</button>
                            </div>
                        </div>
                    </li>
                    
                    <!-- Сообщение если нет задач -->
                    <li v-if="note.todos.length === 0" class="list-group-item text-muted text-center py-3">
                        Нет задач. Добавьте первую задачу ниже.
                    </li>
                </ul>

                <!-- Форма для добавления новой задачи в заметку -->
                <form @submit.prevent="addTodo(i)" class="d-flex mt-2">
                    <input type="text" 
                           v-model="newTodoTexts[i]" 
                           class="form-control form-control-sm me-2" 
                           placeholder="Новая задача..."
                           ref="newTodoInput"/>
                    <button type="submit" class="btn btn-primary btn-sm">+ Добавить</button>
                </form>

                <!-- Кнопка удаления заметки -->
                <div class="mt-3 pt-3 border-top">
                    <button @click="removeNote(i)" class="btn btn-danger btn-sm">
                        🗑️ Удалить заметку
                    </button>
                </div>
            </div>
        </div>

        <!-- Форма для добавления новой заметки -->
        <div class="card">
            <div class="card-body">
                <h5 class="card-title mb-3">Добавить новую заметку</h5>
                <form @submit.prevent="addNote" class="d-flex">
                    <input type="text" 
                           v-model="newNoteTitle" 
                           class="form-control me-2" 
                           placeholder="Введите название заметки..."
                           required/>
                    <button type="submit" class="btn btn-success">+ Создать заметку</button>
                </form>
            </div>
        </div>

        <!-- Сообщение если нет заметок -->
        <div v-if="notes.length === 0" class="text-center mt-5">
            <div class="alert alert-info">
                <h5>Заметок пока нет</h5>
                <p class="mb-0">Создайте первую заметку, используя форму выше.</p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import Swal from 'sweetalert2';

// Начинаем с пустого массива заметок
const notes = ref([]);

const newNoteTitle = ref('');
const newTodoTexts = ref([]);
const editingNote = ref({});
const editingNoteTitle = ref({}); // Для хранения редактируемого заголовка заметки
const editingTodo = ref({});
const editingTodoText = ref({}); // Для хранения редактируемого текста задачи

// Инициализируем массив для новых задач при монтировании
onMounted(() => {
    // Начинаем с пустого массива
    newTodoTexts.value = [];
});

function addNote() {
    const title = newNoteTitle.value.trim();
    
    if (!title) {
        Swal.fire({
            title: 'Ошибка!',
            text: 'Введите название заметки',
            icon: 'error',
            timer: 1500
        });
        return;
    }
    
    // Добавляем новую заметку
    notes.value.push({ 
        title: title, 
        todos: []
    });
    
    // Добавляем пустую строку для новой заметки в массиве newTodoTexts
    newTodoTexts.value.push('');
    
    // Сбрасываем поле ввода
    newNoteTitle.value = '';
}

function addTodo(noteIndex) {
    const todoText = newTodoTexts.value[noteIndex]?.trim();
    
    if (!todoText) {
        Swal.fire({
            title: 'Ошибка!',
            text: 'Введите текст задачи',
            icon: 'error',
            timer: 1500
        });
        return;
    }
    
    // Добавляем задачу
    notes.value[noteIndex].todos.push({ 
        text: todoText, 
        completed: false 
    });
    
    // Очищаем поле ввода
    newTodoTexts.value[noteIndex] = '';
    
    // Фокусируемся на поле ввода после добавления
    nextTick(() => {
        const inputs = document.querySelectorAll('input[type="text"]');
        if (inputs[noteIndex]) {
            inputs[noteIndex].focus();
        }
    });
}

function removeNote(index) {
    Swal.fire({
        title: 'Вы уверены?',
        text: 'Заметка и все её задачи будут удалены!',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonText: 'Да, удалить!',
        cancelButtonText: 'Отмена',
        confirmButtonColor: '#d33',
        cancelButtonColor: '#3085d6',
    }).then((result) => {
        if (result.isConfirmed) {
            const noteTitle = notes.value[index].title;
            notes.value.splice(index, 1);
            newTodoTexts.value.splice(index, 1); // Удаляем соответствующий input
            
            Swal.fire({
                title: 'Удалено!',
                text: `Заметка "${noteTitle}" удалена`,
                icon: 'success',
                timer: 1500,
                showConfirmButton: false
            });
        }
    });
}

function removeTodo(noteIndex, todoIndex) {
    Swal.fire({
        title: 'Удалить задачу?',
        text: 'Задача будет удалена',
        icon: 'question',
        showCancelButton: true,
        confirmButtonText: 'Удалить',
        cancelButtonText: 'Отмена',
        confirmButtonColor: '#d33',
        cancelButtonColor: '#3085d6',
    }).then((result) => {
        if (result.isConfirmed) {
            notes.value[noteIndex].todos.splice(todoIndex, 1);
        }
    });
}

function toggleTodo(noteIndex, todoIndex) {
    notes.value[noteIndex].todos[todoIndex].completed = !notes.value[noteIndex].todos[todoIndex].completed;
}

function startEditingNote(index) {
    editingNote.value[index] = true;
    editingNoteTitle.value[index] = notes.value[index].title; // Сохраняем оригинальный текст
}

function saveNote(index) {
    const title = editingNoteTitle.value[index]?.trim();
    
    if (!title) {
        Swal.fire({
            title: 'Ошибка!',
            text: 'Название заметки не может быть пустым',
            icon: 'error',
            timer: 1500
        });
        return;
    }
    
    notes.value[index].title = title;
    editingNote.value[index] = false;
    delete editingNoteTitle.value[index]; // Очищаем временное значение
}

function cancelEditingNote(index) {
    editingNote.value[index] = false;
    delete editingNoteTitle.value[index]; // Очищаем временное значение
}

function startEditingTodo(noteIndex, todoIndex) {
    if (!editingTodo.value[noteIndex]) {
        editingTodo.value[noteIndex] = {};
    }
    if (!editingTodoText.value[noteIndex]) {
        editingTodoText.value[noteIndex] = {};
    }
    
    editingTodo.value[noteIndex][todoIndex] = true;
    editingTodoText.value[noteIndex][todoIndex] = notes.value[noteIndex].todos[todoIndex].text; // Сохраняем оригинальный текст
    
    // Фокусируемся на поле ввода после начала редактирования
    nextTick(() => {
        const input = document.querySelector(`input[ref="todoInput"]`);
        if (input) {
            input.focus();
            input.select();
        }
    });
}

function saveTodo(noteIndex, todoIndex) {
    const text = editingTodoText.value[noteIndex]?.[todoIndex]?.trim();
    
    if (!text) {
        Swal.fire({
            title: 'Ошибка!',
            text: 'Текст задачи не может быть пустым',
            icon: 'error',
            timer: 1500
        });
        return;
    }
    
    notes.value[noteIndex].todos[todoIndex].text = text;
    editingTodo.value[noteIndex][todoIndex] = false;
    
    // Очищаем временные данные
    if (editingTodoText.value[noteIndex]) {
        delete editingTodoText.value[noteIndex][todoIndex];
    }
}

function cancelEditingTodo(noteIndex, todoIndex) {
    editingTodo.value[noteIndex][todoIndex] = false;
    
    // Очищаем временные данные
    if (editingTodoText.value[noteIndex]) {
        delete editingTodoText.value[noteIndex][todoIndex];
    }
}
</script>

<style scoped>
.card {
    transition: all 0.3s ease;
    border: 1px solid #e0e0e0;
}

.card:hover {
    box-shadow: 0 5px 15px rgba(0,0,0,0.1);
    transform: translateY(-2px);
}

.list-group-item {
    transition: background-color 0.2s;
    border-left: none;
    border-right: none;
}

.list-group-item:hover {
    background-color: #f8f9fa;
}

.list-group-item:first-child {
    border-top: none;
}

.list-group-item:last-child {
    border-bottom: none;
}

.btn-sm {
    padding: 0.25rem 0.5rem;
    font-size: 0.875rem;
}

.form-control-sm {
    padding: 0.25rem 0.5rem;
    font-size: 0.875rem;
}

.text-decoration-line-through {
    color: #6c757d;
    opacity: 0.7;
}

.card-title {
    color: #2c3e50;
}

.alert {
    border-radius: 10px;
    border: none;
    box-shadow: 0 3px 10px rgba(0,0,0,0.08);
}
</style>