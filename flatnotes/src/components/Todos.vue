<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import * as constants from "../constants";
import { eventBus } from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import Icon from "../ui/Icon.vue";
import { vTooltip } from "../ui/tooltip";
import api from "../api";
import type { NoteTodos } from "../types";

const loading = ref(true);
const failed = ref(false);
const noteTodos = ref<NoteTodos[]>([]);
const hideDone = ref(true);

const totalTodos = computed(() =>
  noteTodos.value
    .flatMap(note => note.todos)
    .filter(note => !note.done || !hideDone.value)
    .length,
);

const filteredNoteTodos = computed(() =>
  hideDone.value
    ? noteTodos.value
        .map(note => ({
          ...note,
          todos: note.todos.filter(todo => !todo.done),
        }))
        .filter((note) => note.todos.length > 0)
    : noteTodos.value,
);

function load(): void {
  loading.value = true;
  failed.value = false;
  api<NoteTodos[]>("/api/todos")
    .then((response) => {
      noteTodos.value = response;
    })
    .catch((error) => {
      if (!(error as { handled?: boolean }).handled) {
        failed.value = true;
        eventBus.emit("unhandled-server-error", { error });
      }
    })
    .finally(() => {
      loading.value = false;
    });
}

function noteHref(title: string): string {
  return `${constants.basePaths.note}/${encodeURIComponent(title)}`;
}

function openNote(href: string, event?: Event): void {
  eventBus.emit("navigate", { href, event });
}

function lastModifiedString(ts: number): string {
  return new Date(ts * 1000).toLocaleString();
}

onMounted(load);
</script>

<template>
  <div class="mb-4">
    <h2 class="page-title mb-4">Todos</h2>

    <!-- Loading / failed -->
    <div v-if="loading" class="d-flex justify-content-center">
      <LoadingIndicator />
    </div>
    <div v-else-if="failed" class="d-flex justify-content-center">
      <LoadingIndicator :failed="true" failed-message="Failed to load Todos" />
    </div>

    <!-- Empty -->
    <div v-else-if="totalTodos === 0" class="d-flex justify-content-center">
      <LoadingIndicator :show-loader="false" :failed="true" failed-icon="check2-square" failed-message="No Todos Found" />
    </div>

    <div v-else>
      <!-- Controls -->
      <div class="d-flex align-items-center mb-4">
        <span class="text-muted me-3">{{ totalTodos }} todo{{ totalTodos === 1 ? "" : "s" }}</span>
        <button type="button" class="bttn" @click="hideDone = !hideDone">
          <Icon :name="hideDone ? 'eye' : 'eye-slash'" />
          {{ hideDone ? "Show" : "Hide" }} Done
        </button>
      </div>

      <!-- Groups -->
      <div v-for="note in filteredNoteTodos" :key="note.title" class="mb-4 bttn result">
        <div class="d-flex justify-content-between">
          <div>
            <a
              :href="noteHref(note.title)"
              class="note-title"
              @click.prevent="openNote(noteHref(note.title), $event)"
            >
              {{ note.title }}
            </a>
            <ul class="todo-list">
              <li
                v-for="(todo, i) in note.todos"
                :key="i"
                class="todo-item"
                :class="{ done: todo.done }"
              >
                <Icon :name="todo.done ? 'check-square' : 'square'" class="todo-check" />
                <span class="todo-text">{{ todo.text }}</span>
              </li>
            </ul>
          </div>

          <span class="last-modified ms-2" v-tooltip="'Last Modified'">
            {{ lastModifiedString(note.lastModified) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.page-title {
  color: var(--colour-text);
}

.note-title {
  color: var(--colour-text);
  font-size: 18px;
  font-weight: bold;
}

.last-modified {
  color: var(--colour-text-muted);
  font-size: 12px;
}

.todo-list {
  list-style: none;
  padding: 0;
  margin: 12px 0 0 0;
}

.todo-item {
  display: flex;
  align-items: center;
  padding: 4px 0;
  color: var(--colour-text);
}

.todo-check {
  color: var(--colour-brand);
  margin-right: 8px;
  font-size: 16px;
}

.todo-text {
  flex: 1;
}

.todo-item.done .todo-text {
  color: var(--colour-text-muted);
  text-decoration: line-through;
}
</style>
