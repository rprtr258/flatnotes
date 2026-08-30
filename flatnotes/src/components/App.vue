<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from "vue";
import Mousetrap from "mousetrap";

import NavBar from "./NavBar.vue";
import Login from "./Login.vue";
import Logo from "./Logo.vue";
import SearchInput from "./SearchInput.vue";
import OpenModal from "./OpenModal.vue";
import SearchResults from "./SearchResults.vue";
import RecentlyModified from "./RecentlyModified.vue";
import NoteViewerEditor from "./NoteViewerEditor.vue";
import Todos from "./Todos.vue";
import TreeExplorer from "./TreeExplorer.vue";

import Modal from "../ui/Modal.vue";
import ToastHost from "../ui/ToastHost.vue";
import ConfirmModal from "../ui/ConfirmModal.vue";

import * as constants from "../constants.js";
import * as helpers from "../helpers.js";
import { clearToken } from "../tokenStorage.js";
import { eventBus } from "../eventBus.js";
import api from "../api.js";
import type { ConfigResponse } from "../types.js";
import { toast } from "../composables/useToast.js";

const views = {
  login: 0,
  home: 1,
  note: 2,
  search: 3,
  todos: 4,
} as const;

const authType = ref<string | null>(null);
const currentView = ref<number>(views.home);
const noteTitle = ref<string | null>(null);
const searchTerm = ref<string | null>(null);
const darkTheme = ref(false);
const searchModalOpen = ref(false);
const openModalOpen = ref(false);

watch(darkTheme, (value) => {
  if (value) {
    document.body.classList.add("dark-theme");
  } else {
    document.body.classList.remove("dark-theme");
  }
});

function updateDocumentTitle(suffix?: string): void {
  window.document.title = (suffix ? `${suffix} - ` : "") + "flatnotes";
}

function focusSearchInput(): void {
  const input = document.getElementById("search-input") as HTMLInputElement | null;
  if (input) {
    input.focus();
    input.select();
  }
}

function route(): void {
  const path = window.location.pathname.split("/");
  const basePath = `/${path[1]}`;
  searchModalOpen.value = false;

  if (basePath === constants.basePaths.home) {
    updateDocumentTitle();
    currentView.value = views.home;
    nextTick(() => focusSearchInput());
  } else if (basePath === constants.basePaths.search) {
    updateDocumentTitle("Search");
    searchTerm.value = helpers.getSearchParam(constants.params.searchTerm);
    currentView.value = views.search;
  } else if (basePath === constants.basePaths.todos) {
    updateDocumentTitle("Todos");
    currentView.value = views.todos;
  } else if (basePath === constants.basePaths.new) {
    updateDocumentTitle("New Note");
    currentView.value = views.note;
  } else if (basePath === constants.basePaths.note) {
    noteTitle.value = decodeURIComponent(path.slice(2).join("/"));
    updateDocumentTitle(noteTitle.value);
    currentView.value = views.note;
  } else if (basePath === constants.basePaths.login) {
    updateDocumentTitle("Log In");
    currentView.value = views.login;
  }
}

function navigate(href: string, event?: Event): void {
  if (event != undefined && (event as MouseEvent).ctrlKey === true) {
    window.open(href);
  } else {
    history.pushState(null, "", href);
    noteTitle.value = null;
    searchTerm.value = null;
    route();
  }
}

function loadConfig(): void {
  api<ConfigResponse>("/api/config")
    .then((response) => {
      authType.value = response.authType;
    })
    .catch((error) => {
      if (!(error as { handled?: boolean }).handled) {
        unhandledServerErrorToast(error);
      }
    });
}

function logout(): void {
  clearToken();
  navigate(constants.basePaths.login);
}

function noteDeletedToast(): void {
  toast("Note deleted ✓", { variant: "success" });
}

function openSearch(): void {
  if (([views.home, views.search] as number[]).includes(currentView.value)) {
    focusSearchInput();
    eventBus.emit("highlight-search-input");
  } else if (currentView.value !== views.login) {
    searchModalOpen.value = true;
  }
}

function unhandledServerErrorToast(error: unknown): void {
  console.log(error);
  toast("Unknown error communicating with the server. Please try again.", { variant: "danger" });
}

function toggleTheme(): void {
  darkTheme.value = !darkTheme.value;
  localStorage.setItem("darkTheme", String(darkTheme.value));
}

function updateNoteTitle(title: string): void {
  noteTitle.value = title;
  updateDocumentTitle(title);
}

eventBus.on("navigate", (payload) => navigate(payload.href, payload.event));
eventBus.on("unhandled-server-error", (payload) => unhandledServerErrorToast(payload.error));
eventBus.on("update-note-title", (payload) => updateNoteTitle(payload.title));

Mousetrap.bind("ctrl+shift+f", () => {
  openSearch();
  return false;
});

Mousetrap.bind("ctrl+k", () => {
  if (currentView.value !== views.login) {
    openModalOpen.value = true;
  }
  return false;
});

loadConfig();

const storedDarkTheme = localStorage.getItem("darkTheme");
if (storedDarkTheme != null) {
  darkTheme.value = storedDarkTheme === "true";
} else if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
  darkTheme.value = true;
}

route();

function onPopState(): void {
  route();
}

onMounted(() => {
  window.addEventListener("popstate", onPopState);
});

onUnmounted(() => {
  window.removeEventListener("popstate", onPopState);
  eventBus.off("navigate");
  eventBus.off("unhandled-server-error");
  eventBus.off("update-note-title");
  Mousetrap.unbind("ctrl+shift+f");
  Mousetrap.unbind("ctrl+k");
});
</script>

<template>
  <div class="d-flex flex-column h-100">
    <!-- Search Modal -->
    <Modal v-model="searchModalOpen" :hide-header="true" :hide-footer="true">
      <SearchInput />
    </Modal>

    <!-- Open Modal (Ctrl+K) -->
    <OpenModal v-model="openModalOpen" />

    <!-- Nav Bar -->
    <NavBar
      v-if="currentView !== views.login"
      class="w-100 mb-5"
      :show-logo="currentView !== views.home"
      :auth-type="authType"
      :dark-theme="darkTheme"
      @logout="logout()"
      @toggle-theme="toggleTheme()"
      @search="openSearch()"
    ></NavBar>

    <!-- Login -->
    <Login v-if="currentView === views.login" class="flex-grow-1" :auth-type="authType"></Login>

    <!-- Main layout: sidebar + content -->
    <div v-else class="d-flex flex-row flex-grow-1 main-layout">
      <TreeExplorer
        class="sidebar d-none d-md-flex"
        :active-title="noteTitle"
      />
      <div class="main-content d-flex flex-column flex-grow-1">
        <!-- Home -->
        <div
          v-if="currentView === views.home"
          class="home-view align-self-center d-flex flex-column justify-content-center align-items-center flex-grow-1 w-100"
        >
          <Logo class="mb-3" />
          <SearchInput :initial-value="searchTerm ?? undefined" class="search-input mb-4" />
          <div v-if="authType != null && authType !== constants.authTypes.readOnly">
            <RecentlyModified class="recently-modified" :max-notes="5" />
          </div>
        </div>

        <!-- Search Results -->
        <div
          v-if="currentView === views.search"
          class="flex-grow-1 search-results-view d-flex flex-column"
        >
          <SearchResults :search-term="searchTerm" class="flex-grow-1" />
        </div>

        <!-- Todos -->
        <div
          v-if="currentView === views.todos"
          class="flex-grow-1 todos-view d-flex flex-column"
        >
          <Todos class="flex-grow-1" />
        </div>

        <!-- Note -->
        <NoteViewerEditor
          v-if="currentView === views.note"
          class="flex-grow-1"
          :title-to-load="noteTitle"
          :auth-type="authType"
          @note-deleted="noteDeletedToast"
        ></NoteViewerEditor>
      </div>
    </div>

    <!-- Global overlays -->
    <ToastHost />
    <ConfirmModal />
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.main-layout {
  min-height: 0;
}

.sidebar {
  width: 240px;
  min-width: 240px;
  border-right: 1px solid var(--colour-border);
  overflow: hidden;
}

.main-content {
  min-height: 0;
}

.home-view {
  max-width: 500px;
}

.search-results-view {
  max-width: 700px;
}

.todos-view {
  max-width: 700px;
}

.search-input {
  box-shadow: 0 0 20px var(--colour-shadow);
}

.recently-modified {
  // Prevent UI from moving during load
  min-height: 190px;
}
</style>
