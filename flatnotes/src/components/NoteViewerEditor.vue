<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import * as constants from "../constants";
import { eventBus } from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import MarkdownEditor from "../editor/MarkdownEditor.vue";
import Icon from "../ui/Icon.vue";
import { Note } from "../classes";
import api, { ApiError } from "../api";
import { toast } from "../composables/useToast";
import { msgBoxConfirm } from "../composables/useModal";
import type { NoteContentResponseModel } from "../types";

const props = withDefaults(
  defineProps<{
    titleToLoad?: string | null;
    authType?: string | null;
  }>(),
  {
    titleToLoad: null,
    authType: null,
  },
);

const emit = defineEmits<{ "note-deleted": [] }>();

let saveTimeout: number | null = null;
const currentNote = ref<Note | null>(null);
const titleInput = ref<string | null>(null);
const initialContent = ref<string | null>(null);
const noteLoadFailed = ref(false);
const noteLoadFailedIcon = ref<string | null>(null);
const noteLoadFailedMessage = ref("Failed to load Note");

const markdownEditor = ref<{ getContent: () => string } | null>(null);

const canModify = computed(
  () => props.authType != null && props.authType !== constants.authTypes.readOnly,
);

function loadNote(title: string): void {
  noteLoadFailed.value = false;
  api<NoteContentResponseModel>(`/api/notes/${encodeURIComponent(title)}`)
    .then((response) => {
      currentNote.value = new Note(response.title, response.lastModified, response.content);
      enterEditMode();
    })
    .catch((error) => {
      const err = error as ApiError;
      if (err.handled) {
        return;
      } else if (err.status === 404) {
        noteLoadFailedIcon.value = "file-earmark-x";
        noteLoadFailedMessage.value = "Note not found";
        noteLoadFailed.value = true;
      } else {
        eventBus.emit("unhandled-server-error", { error });
        noteLoadFailed.value = true;
      }
    });
}

function setBeforeUnloadConfirmation(enable = true): void {
  if (enable) {
    window.onbeforeunload = () => true;
  } else {
    window.onbeforeunload = null;
  }
}

function enterEditMode(): void {
  if (!currentNote.value) return;
  titleInput.value = currentNote.value.title;
  initialContent.value = currentNote.value.content ?? "";
}

function getEditorContent(): string | null {
  if (markdownEditor.value) {
    return markdownEditor.value.getContent();
  }
  return null;
}

function clearSaveTimeout(): void {
  if (saveTimeout != null) {
    clearTimeout(saveTimeout);
  }
}

function scheduleSave(): void {
  clearSaveTimeout();
  setBeforeUnloadConfirmation(true);
  saveTimeout = window.setTimeout(saveNote, 1000);
}


function existingTitleToast(): void {
  toast("A note with this title already exists. Please try again with a new title.", { variant: "danger" });
}

function saveNoteResponseHandler(response: NoteContentResponseModel): void {
  localStorage.removeItem(currentNote.value!.title);
  currentNote.value = new Note(response.title, response.lastModified, response.content);
  titleInput.value = response.title;
  initialContent.value = response.content ?? "";
  setBeforeUnloadConfirmation(false);
  eventBus.emit("update-note-title", { title: currentNote.value.title });
  history.replaceState(null, "", currentNote.value.href);
  noteSavedToast();
}

function noteSavedToast(): void {
  toast("Note saved ✓", { variant: "success" });
}

function saveNote(): void {
  const newContent = getEditorContent();

  // Title Validation
  if (typeof titleInput.value === "string") {
    titleInput.value = titleInput.value.trim();
  }
  if (!titleInput.value) {
    // Nothing to save without a title; wait for the user to type one.
    return;
  }

  const reservedCharacters = /[<>:"\\|?*]/;
  if (reservedCharacters.test(titleInput.value)) {
    toast('Due to filename restrictions, the following characters are not allowed in a note title: <>:"\\|?*', {
      variant: "danger",
    });
    return;
  }

  if (currentNote.value!.lastModified == null) {
    // New Note
    api<NoteContentResponseModel>(`/api/notes`, {
      body: {
        title: titleInput.value,
        content: newContent,
      },
    })
      .then(saveNoteResponseHandler)
      .catch((error) => {
        const err = error as ApiError;
        if (err.handled) {
          return;
        } else if (err.status === 409) {
          existingTitleToast();
        } else {
          eventBus.emit("unhandled-server-error", { error });
        }
      });
  } else if (newContent !== currentNote.value!.content || titleInput.value !== currentNote.value!.title) {
    // Modified Note
    api<NoteContentResponseModel>(`/api/notes/${encodeURIComponent(currentNote.value!.title)}`, {
      method: "PATCH",
      body: {
        newTitle: titleInput.value,
        newContent: newContent,
      },
    })
      .then(saveNoteResponseHandler)
      .catch((error) => {
        const err = error as ApiError;
        if (err.handled) {
          return;
        } else if (err.status === 409) {
          existingTitleToast();
        } else {
          eventBus.emit("unhandled-server-error", { error });
        }
      });
  } else {
    // No change
    return;
  }
}

function cancelNote(): void {
  localStorage.removeItem(currentNote.value!.title);
  eventBus.emit("navigate", { href: constants.basePaths.home });
}

function confirmCancelNote(): void {
  const newContent = getEditorContent();
  if (
    newContent !== currentNote.value!.content ||
    titleInput.value !== currentNote.value!.title
  ) {
    msgBoxConfirm(
      `Are you sure you want to close the note '${currentNote.value!.title}' without saving?`,
      {
        centered: true,
        title: "Confirm Closure",
        okVariant: "warning",
      },
    ).then((response) => {
      if (response === true) {
        cancelNote();
      }
    });
  } else {
    cancelNote();
  }
}

function deleteNote(): void {
  msgBoxConfirm(`Are you sure you want to delete the note '${currentNote.value!.title}'?`, {
    centered: true,
    title: "Confirm Deletion",
    okVariant: "danger",
  }).then((response) => {
    if (response === true) {
      api<void>(`/api/notes/${encodeURIComponent(currentNote.value!.title)}`, {
        method: "DELETE",
      })
        .then(() => {
          emit("note-deleted");
          eventBus.emit("navigate", { href: constants.basePaths.home });
        })
        .catch((error) => {
          if (!(error as ApiError).handled) {
            eventBus.emit("unhandled-server-error", { error });
          }
        });
    }
  });
}

function init(): void {
  currentNote.value = null;
  if (props.titleToLoad) {
    loadNote(props.titleToLoad);
  } else {
    currentNote.value = new Note();
    enterEditMode();
  }
}

watch(
  () => props.titleToLoad,
  () => {
    if (props.titleToLoad !== currentNote.value?.title) {
      init();
    }
  },
);

onMounted(() => {
  init();
});

onBeforeUnmount(() => {
  clearSaveTimeout();
  setBeforeUnloadConfirmation(false);
});
</script>

<template>
  <!-- Note -->
  <div class="pb-4">
    <!-- Loading -->
    <div v-if="currentNote == null" class="h-100 d-flex flex-column justify-content-center">
      <LoadingIndicator
        :failed="noteLoadFailed"
        :failed-icon="noteLoadFailedIcon ?? undefined"
        :failed-message="noteLoadFailedMessage"
      />
    </div>

    <!-- Loaded -->
    <div v-else class="d-flex flex-column h-100">
      <!-- Buttons -->
      <div class="d-flex justify-content-between flex-wrap align-items-end mb-3">
        <!-- Title -->
        <input
          type="text"
          class="h2 title-input flex-grow-1"
          v-model="titleInput"
          :readonly="!canModify"
          placeholder="Title"
          @input="scheduleSave"
        />

        <!-- Buttons -->
        <div class="d-flex">
          <!-- Delete -->
          <button
            v-if="canModify"
            type="button"
            class="bttn"
            @click="deleteNote"
          >
            <Icon name="trash" /> Delete
          </button>

          <!-- Cancel -->
          <button v-if="canModify" type="button" class="bttn" @click="confirmCancelNote">
            <Icon name="arrow-return-left" /> Cancel
          </button>
        </div>
      </div>

      <!-- Editor -->
      <div class="note flex-grow-1">
        <MarkdownEditor
          :initial-value="initialContent ?? ''"
          :read-only="!canModify"
          ref="markdownEditor"
          @change="scheduleSave"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.title-input {
  border: none;

  // Override user agent styling
  background-color: transparent;
  color: var(--colour-text);
  padding: 0;

  &:focus {
    outline: none;
  }
}
</style>