<script setup lang="ts">
import { ref, watch, nextTick } from "vue";
import Modal from "../ui/Modal.vue";
import Icon from "../ui/Icon.vue";
import api from "../api";
import { SearchResult } from "../classes";
import { eventBus } from "../eventBus";
import type { SearchResultModel } from "../types";

const props = defineProps<{ modelValue: boolean }>();

const emit = defineEmits<{
  "update:modelValue": [value: boolean];
}>();

const query = ref("");
const input = ref<HTMLInputElement | null>(null);
const results = ref<SearchResult[]>([]);
const selectedIndex = ref(0);
const loading = ref(false);
const maxResults = 50;

let searchToken = 0;
let debounceTimer: number | undefined;

function fetchResults(): void {
  const token = ++searchToken;
  loading.value = true;
  const term = query.value.trim() === "" ? "*" : query.value;
  api<SearchResultModel[]>("/api/search", { params: { term } })
    .then((response) => {
      if (token !== searchToken) return;
      results.value = response.slice(0, maxResults).map((r) => new SearchResult(r));
      selectedIndex.value = 0;
      loading.value = false;
    })
    .catch(() => {
      if (token !== searchToken) return;
      results.value = [];
      loading.value = false;
    });
}

function onInput(): void {
  if (debounceTimer)
    window.clearTimeout(debounceTimer);
  debounceTimer = window.setTimeout(fetchResults, 150);
}

function openNote(index: number): void {
  const note = results.value[index];
  if (!note)
    return;
  eventBus.emit("navigate", { href: note.href });
  emit("update:modelValue", false);
}

function onKeydown(event: KeyboardEvent): void {
  if (results.value.length === 0)
    return;
  if (event.key === "ArrowDown") {
    event.preventDefault();
    selectedIndex.value = (selectedIndex.value + 1) % results.value.length;
  } else if (event.key === "ArrowUp") {
    event.preventDefault();
    selectedIndex.value = (selectedIndex.value - 1 + results.value.length) % results.value.length;
  } else if (event.key === "Enter") {
    event.preventDefault();
    openNote(selectedIndex.value);
  }
}

watch(
  () => props.modelValue,
  async (open) => {
    if (open) {
      query.value = "";
      results.value = [];
      selectedIndex.value = 0;
      await nextTick();
      input.value?.focus();
      fetchResults();
    } else if (debounceTimer) {
      window.clearTimeout(debounceTimer);
    }
  },
);
</script>

<template>
  <Modal
    :model-value="props.modelValue"
    :hide-header="true"
    :hide-footer="true"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="open-modal">
      <div class="open-modal-input-row">
        <Icon name="search" />
        <input
          ref="input"
          type="text"
          class="open-modal-input"
          placeholder="Open a note…"
          v-model="query"
          @input="onInput"
          @keydown="onKeydown"
        />
      </div>
      <ul v-if="results.length > 0" class="open-modal-list">
        <li
          v-for="(result, index) in results"
          :key="result.title"
          :class="{ selected: index === selectedIndex }"
          @click="openNote(index)"
          @mouseenter="selectedIndex = index"
        >
          {{ result.title }}
        </li>
      </ul>
      <div v-else-if="!loading" class="open-modal-empty">No notes found</div>
    </div>
  </Modal>
</template>

<style lang="scss" scoped>
@import "../colours";

.open-modal {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.open-modal-input-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--colour-border);
  padding-bottom: 0.5rem;

  :deep(svg), :deep(.bi) {
    color: var(--colour-text-muted);
    flex-shrink: 0;
  }
}

.open-modal-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  color: var(--colour-text);

  &::placeholder {
    color: var(--colour-text-very-muted);
  }
}

.open-modal-list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 300px;
  overflow-y: auto;

  li {
    padding: 0.4rem 0.5rem;
    border-radius: 0.25rem;
    cursor: pointer;
    color: var(--colour-text);

    &.selected {
      background-color: var(--colour-background-highlight);
    }
  }
}

.open-modal-empty {
  color: var(--colour-text-muted);
  padding: 0.25rem 0.5rem;
}
</style>
