<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from "vue";
import * as constants from "../constants";
import { eventBus } from "../eventBus";
import Icon from "../ui/Icon.vue";
import { toast } from "../composables/useToast";

const props = defineProps<{ initialValue?: string }>();

const searchTermInput = ref<string | null>(null);
const includeHighlightClass = ref(false);
const input = ref<HTMLInputElement | null>(null);

function search(): void {
  if (searchTermInput.value) {
    searchTermInput.value = searchTermInput.value.trim();
  }
  if (searchTermInput.value) {
    eventBus.emit("navigate", {
      href: `${constants.basePaths.search}?${constants.params.searchTerm}=${encodeURIComponent(searchTermInput.value)}`,
    });
  } else {
    toast("Please enter a search term ✘", { variant: "danger" });
  }
}

let highlightTimer: number | undefined;
function highlightSearchInput(): void {
  includeHighlightClass.value = true;
  highlightTimer = window.setTimeout(() => {
    includeHighlightClass.value = false;
  }, 1500);
}

function init(): void {
  searchTermInput.value = props.initialValue ?? null;
}

watch(
  () => props.initialValue,
  () => init(),
);

onMounted(() => {
  input.value?.focus();
  input.value?.select();
  eventBus.on("highlight-search-input", highlightSearchInput);
  init();
});

onUnmounted(() => {
  eventBus.off("highlight-search-input", highlightSearchInput);
  if (highlightTimer) window.clearTimeout(highlightTimer);
});
</script>

<template>
  <form @submit.prevent="search" class="w-100">
    <div class="input-group w-100">
      <input
        id="search-input"
        ref="input"
        type="text"
        inputmode="search"
        class="form-control"
        :class="{ highlight: includeHighlightClass }"
        placeholder="Search"
        v-model="searchTermInput"
      />
      <button class="btn" type="submit">
        <Icon name="search" />
      </button>
    </div>
  </form>
</template>

<style lang="scss" scoped>
@import "../colours";

@keyframes highlight {
  from {
    background-color: var(--colour-background-highlight);
  }

  to {
    background-color: var(--colour-background-elevated);
  }
}

.highlight {
  animation-name: highlight;
  animation-duration: 1.5s;
}

.btn {
  border: 1px solid var(--colour-border);

  svg, .bi {
    color: var(--colour-text-muted);
  }
}

#search-input {
  background-color: var(--colour-background-elevated);
  border-color: var(--colour-border);
  color: var(--colour-text);

  &:focus {
    background-color: var(--colour-background-elevated);
    color: var(--colour-text);
  }

  &::placeholder {
    color: var(--colour-text-muted);
  }
}
</style>
