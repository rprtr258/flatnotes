<script setup lang="ts">
import { ref, onMounted } from "vue";
import { eventBus } from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import { SearchResult } from "../classes";
import api from "../api";
import type { SearchResultModel } from "../types";

const props = defineProps<{ maxNotes?: number }>();

const notes = ref<SearchResult[] | null>(null);
const tags = ref<string[] | null>(null);
const loadingFailed = ref(false);
const loadingFailedMessage = ref("Failed to load notes");
const loadingFailedIcon = ref<string | null>(null);

function getNotes(): void {
  loadingFailed.value = false;
  api<SearchResultModel[]>("/api/search", {
    params: {
      term: "*",
      sort: "lastModified",
      order: "desc",
      limit: props.maxNotes,
    },
  })
    .then((response) => {
      notes.value = [];
      if (response.length) {
        response.forEach((searchResult) => {
          notes.value!.push(new SearchResult(searchResult));
        });
      } else {
        loadingFailedMessage.value = "Click the 'New' button at the top of the page to add your first note";
        loadingFailedIcon.value = "pencil";
        loadingFailed.value = true;
      }
    })
    .catch((error) => {
      loadingFailed.value = true;
      if (!(error as { handled?: boolean }).handled) {
        eventBus.emit("unhandled-server-error", { error });
      }
    });
}

function getTags(): void {
  loadingFailed.value = false;
  api<string[]>("/api/tags")
    .then((response) => {
      tags.value = [];
      if (response.length) {
        response.forEach((tag) => {
          tags.value!.push(tag);
        });
      } else {
        loadingFailedMessage.value = "No tags";
        loadingFailedIcon.value = "pencil";
        loadingFailed.value = true;
      }
    })
    .catch((error) => {
      loadingFailed.value = true;
      if (!(error as { handled?: boolean }).handled) {
        eventBus.emit("unhandled-server-error", { error });
      }
    });
}

function openNote(href: string, event?: Event): void {
  eventBus.emit("navigate", { href, event });
}

function openTag(tag: string, event?: Event): void {
  eventBus.emit("navigate", { href: "/search?term=" + encodeURIComponent("#" + tag), event });
}

onMounted(() => {
  getNotes();
  getTags();
});
</script>

<template>
  <div class="justify-content-top">
    <!-- Loading -->
    <div v-if="notes == null || notes.length == 0" class="h-100 d-flex flex-column justify-content-center">
      <LoadingIndicator
        :failed="loadingFailed"
        :failed-message="loadingFailedMessage"
        :failed-icon="loadingFailedIcon ?? undefined"
        :show-loader="false"
      />
    </div>
    <!-- Notes Loaded -->
    <div v-else
      class="d-flex flex-row align-items-start"
    >
      <div
        class="d-flex flex-column align-items-center"
      >
        <p class="mini-header mb-1">RECENTLY MODIFIED</p>
        <a
          v-for="note in notes"
          :key="note.title"
          class="bttn"
          :href="note.href"
          @click.prevent="openNote(note.href, $event)"
          >{{ note.title }}</a
        >
      </div>
      <div class="d-flex flex-column align-items-center">
        <p class="mini-header mb-1">TAGS</p>
        <a
          v-for="tag in tags"
          :key="tag"
          class="bttn"
          :href="tag"
          @click.prevent="openTag(tag, $event)"
          >#{{ tag }}</a
        >
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.mini-header {
  font-size: 12px;
  font-weight: bold;
  color: var(--colour-text-very-muted);
}
</style>
