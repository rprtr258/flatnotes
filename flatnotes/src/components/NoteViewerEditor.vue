<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import * as constants from "../constants";
import { eventBus } from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import ToastUiEditor from "./ToastUiEditor.vue";
import ToastUiViewer from "./ToastUiViewer.vue";
import Icon from "../ui/Icon.vue";
import { vTooltip } from "../ui/tooltip";
import Mousetrap from "mousetrap";
import { Note } from "../classes";
import api, { ApiError } from "../api";
import { toast } from "../composables/useToast";
import { msgBoxConfirm } from "../composables/useModal";
import type { EditorType, EditorPlugin, CustomHTMLRenderer } from "@toast-ui/editor";
import type { NoteContentResponseModel } from "../types";
import codeSyntaxHighlight from "@toast-ui/editor-plugin-code-syntax-highlight/dist/toastui-editor-plugin-code-syntax-highlight-all.js";

const customHTMLRenderer: CustomHTMLRenderer = {
  heading(node: { level: number }, context: { entering: boolean; getChildrenText: (n: unknown) => string }) {
    const tagName = `h${node.level}`;
    if (context.entering) {
      return {
        type: "openTag",
        tagName,
        attributes: {
          id: context
            .getChildrenText(node)
            .toLowerCase()
            .replace(/[^a-z0-9-\s]*/g, "")
            .trim()
            .replace(/\s/g, "-"),
        },
      };
    }
    return { type: "closeTag", tagName };
  },
};

const viewerOptions = {
  customHTMLRenderer,
  plugins: [codeSyntaxHighlight as unknown as EditorPlugin],
  extendedAutolinks: true,
};

const editorOptions = {
  customHTMLRenderer,
  plugins: [codeSyntaxHighlight as unknown as EditorPlugin],
};

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

const editMode = ref(false);
let draftSaveTimeout: number | null = null;
let titleSaveTimeout: number | null = null;
const currentNote = ref<Note | null>(null);
const titleInput = ref<string | null>(null);
const initialContent = ref<string | null>(null);
const noteLoadFailed = ref(false);
const noteLoadFailedIcon = ref<string | null>(null);
const noteLoadFailedMessage = ref("Failed to load Note");

const toastUiEditor = ref<{ invoke: <T = unknown>(method: string, ...args: unknown[]) => T } | null>(null);
const viewTitleEl = ref<HTMLElement | null>(null);
const viewTitleText = ref("");

watch(
  viewTitleText,
  (text) => {
    if (viewTitleEl.value && viewTitleEl.value.innerText !== text) {
      viewTitleEl.value.innerText = text;
    }
  },
  { flush: "post" },
);

const canModify = computed(
  () => props.authType != null && props.authType !== constants.authTypes.readOnly,
);

function loadNote(title: string): void {
  noteLoadFailed.value = false;
  api<NoteContentResponseModel>(`/api/notes/${encodeURIComponent(title)}`)
    .then((response) => {
      currentNote.value = new Note(response.title, response.lastModified, response.content);
      viewTitleText.value = response.title;
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

function setEditMode(edit = true): void {
  // To Edit Mode
  if (edit === true) {
    setBeforeUnloadConfirmation(true);
    titleInput.value = currentNote.value!.title;
    const draftContent = localStorage.getItem(currentNote.value!.title);

    if (draftContent) {
      msgBoxConfirm(
        "There is an unsaved draft of this note stored in this browser. Do you want to resume the draft version or delete it?",
        {
          centered: true,
          title: "Resume Draft?",
          okVariant: "primary",
        },
      ).then((response) => {
        if (response === true) {
          initialContent.value = draftContent;
        } else {
          initialContent.value = currentNote.value!.content ?? "";
          localStorage.removeItem(currentNote.value!.title);
        }
        editMode.value = true;
      });
    } else {
      initialContent.value = currentNote.value!.content ?? "";
      editMode.value = true;
    }
  }
  // To View Mode
  else {
    titleInput.value = null;
    initialContent.value = null;
    setBeforeUnloadConfirmation(false);
    editMode.value = false;
  }
}

function getEditorContent(): string | null {
  if (toastUiEditor.value) {
    return toastUiEditor.value.invoke<string>("getMarkdown");
  }
  return null;
}

function saveDefaultEditorMode(): void {
  const isWysiwygMode = toastUiEditor.value!.invoke<boolean>("isWysiwygMode");
  localStorage.setItem("defaultEditorMode", isWysiwygMode ? "wysiwyg" : "markdown");
}

function loadDefaultEditorMode(): EditorType {
  const defaultEditorMode = localStorage.getItem("defaultEditorMode");
  return (defaultEditorMode as EditorType) ?? "markdown";
}

function clearDraftSaveTimeout(): void {
  if (draftSaveTimeout != null) {
    clearTimeout(draftSaveTimeout);
  }
}

function startDraftSaveTimeout(): void {
  clearDraftSaveTimeout();
  draftSaveTimeout = window.setTimeout(saveDraft, 1000);
}

function saveDraft(): void {
  const content = getEditorContent();
  if (content) {
    localStorage.setItem(currentNote.value!.title, content);
  }
}

const reservedTitleCharacters = /[<>:"\\|?*]/;

function existingTitleToast(): void {
  toast("A note with this title already exists. Please try again with a new title.", { variant: "danger" });
}

function saveNoteResponseHandler(response: NoteContentResponseModel): void {
  localStorage.removeItem(currentNote.value!.title);
  currentNote.value = new Note(response.title, response.lastModified, response.content);
  viewTitleText.value = response.title;
  eventBus.emit("update-note-title", { title: currentNote.value.title });
  history.replaceState(null, "", currentNote.value.href);
  setEditMode(false);
  noteSavedToast();
}

function noteSavedToast(): void {
  toast("Note saved ✓", { variant: "success" });
}

function clearTitleSaveTimeout(): void {
  if (titleSaveTimeout != null) {
    clearTimeout(titleSaveTimeout);
    titleSaveTimeout = null;
  }
}

function saveTitle(showToast: boolean): void {
  clearTitleSaveTimeout();
  if (!canModify.value || !currentNote.value || !viewTitleEl.value) {
    return;
  }
  const newTitle = viewTitleEl.value.innerText.trim();
  const oldTitle = currentNote.value.title;
  if (newTitle === oldTitle) {
    if (showToast) {
      viewTitleEl.value.innerText = oldTitle;
    }
    return;
  }
  if (!newTitle) {
    if (showToast) {
      viewTitleEl.value.innerText = oldTitle;
      toast("Cannot save note without a title ✘", { variant: "danger" });
    }
    return;
  }
  if (reservedTitleCharacters.test(newTitle)) {
    if (showToast) {
      viewTitleEl.value.innerText = oldTitle;
      toast('Due to filename restrictions, the following characters are not allowed in a note title: <>:"\\\\|?*', {
        variant: "danger",
      });
    }
    return;
  }
  api<NoteContentResponseModel>(`/api/notes/${encodeURIComponent(oldTitle)}`, {
    method: "PATCH",
    body: { newTitle },
  })
    .then((response) => {
      if (currentNote.value?.title !== oldTitle) {
        return;
      }
      currentNote.value = new Note(response.title, response.lastModified, response.content);
      eventBus.emit("update-note-title", { title: currentNote.value.title });
      history.replaceState(null, "", currentNote.value.href);
      if (showToast) {
        viewTitleText.value = response.title;
        toast("Note title updated ✓", { variant: "success" });
      }
    })
    .catch((error) => {
      if (currentNote.value?.title !== oldTitle) {
        return;
      }
      viewTitleEl.value!.innerText = currentNote.value!.title;
      const err = error as ApiError;
      if (err.handled) {
        return;
      } else if (err.status === 409) {
        existingTitleToast();
      } else {
        eventBus.emit("unhandled-server-error", { error });
      }
    });
}

function scheduleTitleSave(): void {
  clearTitleSaveTimeout();
  titleSaveTimeout = window.setTimeout(() => saveTitle(false), 600);
}

function flushTitleSave(): void {
  clearTitleSaveTimeout();
  saveTitle(true);
}

function saveNote(): void {
  const newContent = getEditorContent();

  saveDefaultEditorMode();

  // Title Validation
  if (typeof titleInput.value === "string") {
    titleInput.value = titleInput.value.trim();
  }
  if (!titleInput.value) {
    toast("Cannot save note without a title ✘", { variant: "danger" });
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
    // No Change
    localStorage.removeItem(currentNote.value!.title);
    setEditMode(false);
    noteSavedToast();
  }
}

function cancelNote(): void {
  localStorage.removeItem(currentNote.value!.title);
  if (currentNote.value!.lastModified == null) {
    // Cancelling a new note
    eventBus.emit("navigate", { href: constants.basePaths.home });
  } else {
    setEditMode(false);
  }
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
    setEditMode(false);
  } else {
    currentNote.value = new Note();
    viewTitleText.value = "";
    setEditMode(true);
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
  // 'e' to edit
  Mousetrap.bind("e", () => {
    if (editMode.value === false && canModify.value) {
      setEditMode(true);
    }
  });
  init();
});

onBeforeUnmount(() => {
  Mousetrap.unbind("e");
  clearDraftSaveTimeout();
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
        <h2
          v-if="editMode === false"
          ref="viewTitleEl"
          class="title"
          :contenteditable="canModify"
          :title="viewTitleText"
          @input="scheduleTitleSave"
          @blur="flushTitleSave"
          @keydown.enter.prevent.exact="($event.target as HTMLElement).blur()"
        >
        </h2>
        <input
          v-else
          type="text"
          class="h2 title-input flex-grow-1"
          v-model="titleInput"
          placeholder="Title"
        />

        <!-- Buttons -->
        <div class="d-flex">
          <!-- Edit -->
          <button
            v-if="canModify && editMode === false && noteLoadFailed === false"
            type="button"
            class="bttn"
            @click="setEditMode(true)"
            v-tooltip="'Keyboard Shortcut: e'"
          >
            <Icon name="pencil-square" /> Edit
          </button>

          <!-- Delete -->
          <button
            v-if="canModify && editMode === false && noteLoadFailed === false"
            type="button"
            class="bttn"
            @click="deleteNote"
          >
            <Icon name="trash" /> Delete
          </button>

          <!-- Cancel -->
          <button v-if="editMode === true" type="button" class="bttn" @click="confirmCancelNote">
            <Icon name="arrow-return-left" /> Cancel
          </button>

          <!-- Save -->
          <button v-if="editMode === true" type="button" class="bttn" @click="saveNote">
            <Icon name="check-square" /> Save
          </button>
        </div>
      </div>

      <!-- Viewer -->
      <div v-if="editMode === false" class="note note-viewer">
        <ToastUiViewer :initial-value="currentNote.content" :options="viewerOptions" />
      </div>

      <!-- Editor -->
      <div v-else class="note flex-grow-1">
        <ToastUiEditor
          :initial-value="initialContent ?? undefined"
          :initial-edit-type="loadDefaultEditorMode()"
          preview-style="tab"
          ref="toastUiEditor"
          :options="editorOptions"
          height="100%"
          @change="startDraftSaveTimeout"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.title {
  min-width: 300px;
  height: 1.5em;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow-x: hidden;
  color: var(--colour-text);
  margin: 0;

  &[contenteditable="true"] {
    cursor: text;
    &:hover, &:focus {
      outline: none;
    }
  }
}

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

<style lang="scss">
// Toast UI Markdown Editor
@import "@toast-ui/editor/dist/toastui-editor.css";
@import "@toast-ui/editor/dist/toastui-editor-viewer.css";
@import "prismjs/themes/prism.css";
@import "@toast-ui/editor-plugin-code-syntax-highlight/dist/toastui-editor-plugin-code-syntax-highlight.css";

@import "../colours";
@import "../toastui-editor-theme.scss";

.ProseMirror {
  font-family: "Inter", sans-serif;
}

@mixin note-padding {
  padding: min(2vw, 30px) min(3vw, 40px);
}

.toastui-editor-contents {
  font-family: "Inter", sans-serif;
  h1, h2, h3, h4, h5, h6 {
    border-bottom: none;
  }
  @include note-padding;
}

.toastui-editor-defaultUI .ProseMirror {
  @include note-padding;
}

// Override the default font-family for code blocks as some of the fallbacks are not monospace
.toastui-editor-contents code,
.toastui-editor-contents pre,
.toastui-editor-md-code,
.toastui-editor-md-code-block {
  font-family: Consolas, "Lucida Console", Monaco, "Andale Mono", monospace;
}

// Disable checkboxes in view mode. See https://github.com/nhn/tui.editor/issues/1087.
.note-viewer li.task-list-item {
  pointer-events: none;
  a {
    pointer-events: auto;
  }
}
</style>
