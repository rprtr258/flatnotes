<script>
import * as constants from "../constants";

import EventBus from "../eventBus";
import Mousetrap from "mousetrap";
import { Note } from "../classes";
import api from "../api";
import codeSyntaxHighlight from "@toast-ui/editor-plugin-code-syntax-highlight/dist/toastui-editor-plugin-code-syntax-highlight-all.js";

const customHTMLRenderer = {
  heading(node, { entering, getChildrenText }) {
    const tagName = `h${node.level}`;
    if (entering) {
      return {
        type: "openTag",
        tagName,
        attributes: {
          id: getChildrenText(node)
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

export default {
  props: {
    titleToLoad: { type: String, default: null },
    authType: { type: String, default: null },
  },

  data() {
    return {
      editMode: false,
      draftSaveTimeout: null,
      currentNote: null,
      titleInput: null,
      initialContent: null,
      noteLoadFailed: false,
      noteLoadFailedIcon: null,
      noteLoadFailedMessage: "Failed to load Note",
      viewerOptions: {
        customHTMLRenderer: customHTMLRenderer,
        plugins: [codeSyntaxHighlight],
        extendedAutolinks: true,
      },
      editorOptions: {
        customHTMLRenderer: customHTMLRenderer,
        plugins: [codeSyntaxHighlight],
      },
    };
  },

  watch: {
    titleToLoad() {
      if (this.titleToLoad !== this.currentNote?.title) {
        this.init();
      }
    },
  },

  methods: {
    loadNote(title) {
      this.noteLoadFailed = false;
      api(`/api/notes/${encodeURIComponent(title)}`)
        .then((response) => {
          this.currentNote = Note(
            response.title,
            response.lastModified,
            response.content
          );
          // EventBus.$emit("updateDocumentTitle", this.currentNote.title);
        })
        .catch((error) => {
          if (error.handled) {
            return;
          } else if (
            typeof error.response !== "undefined" &&
            error.response.status == 404
          ) {
            this.noteLoadFailedIcon = "file-earmark-x";
            this.noteLoadFailedMessage = "Note not found";
            this.noteLoadFailed = true;
          } else {
            EventBus.$emit("unhandledServerError", error);
            this.noteLoadFailed = true;
          }
        });
    },

    getContentForEditor() {
      let draftContent = localStorage.getItem(this.currentNote.title);
      if (draftContent) {
        if (confirm("Do you want to resume the saved draft?")) {
          return draftContent;
        } else {
          localStorage.removeItem(this.currentNote.title);
        }
      }
      return this.currentNote.content;
    },

    setBeforeUnloadConfirmation(enable = true) {
      if (enable) {
        window.onbeforeunload = () => true;
      } else {
        window.onbeforeunload = null;
      }
    },

    setEditMode(editMode = true) {
      // To Edit Mode
      if (editMode === true) {
        this.setBeforeUnloadConfirmation(true);
        this.titleInput = this.currentNote.title;
        let draftContent = localStorage.getItem(this.currentNote.title);

        if (draftContent) {
          this.$bvModal
            .msgBoxConfirm(
              "There is an unsaved draft of this note stored in this browser. Do you want to resume the draft version or delete it?",
              {
                centered: true,
                title: "Resume Draft?",
                okTitle: "Resume Draft",
                cancelTitle: "Delete Draft",
                cancelVariant: "danger",
              }
            )
            .then((response) => {
              if (response == true) {
                this.initialContent = draftContent;
              } else {
                this.initialContent = this.currentNote.content;
                localStorage.removeItem(this.currentNote.title);
              }
              this.editMode = true;
            });
        } else {
          this.initialContent = this.currentNote.content;
          this.editMode = true;
        }
      }
      // To View Mode
      else {
        this.titleInput = null;
        this.initialContent = null;
        this.setBeforeUnloadConfirmation(false);
        this.editMode = false;
      }
    },

    getEditorContent() {
      if (typeof this.$refs.toastUiEditor != "undefined") {
        return this.$refs.toastUiEditor.invoke("getMarkdown");
      } else {
        return null;
      }
    },

    saveDefaultEditorMode() {
      let isWysiwygMode = this.$refs.toastUiEditor.invoke("isWysiwygMode");
      localStorage.setItem(
        "defaultEditorMode",
        isWysiwygMode ? "wysiwyg" : "markdown"
      );
    },

    loadDefaultEditorMode() {
      let defaultWysiwygMode = localStorage.getItem("defaultEditorMode");
      if (defaultWysiwygMode) {
        return defaultWysiwygMode;
      } else {
        return "markdown";
      }
    },

    clearDraftSaveTimeout() {
      if (this.draftSaveTimeout != null) {
        clearTimeout(this.draftSaveTimeout);
      }
    },

    startDraftSaveTimeout() {
      this.clearDraftSaveTimeout();
      this.draftSaveTimeout = setTimeout(this.saveDraft, 1000);
    },

    saveDraft() {
      let content = this.getEditorContent();
      if (content) {
        localStorage.setItem(this.currentNote.title, content);
      }
    },

    existingTitleToast() {
      this.$bvToast.toast(
        "A note with this title already exists. Please try again with a new title.",
        {
          title: "Duplicate ✘",
          variant: "danger",
          noCloseButton: true,
          toaster: "b-toaster-bottom-right",
        }
      );
    },

    saveNote() {
      let newContent = this.getEditorContent();

      this.saveDefaultEditorMode();

      // Title Validation
      if (typeof this.titleInput == "string") {
        this.titleInput = this.titleInput.trim();
      }
      if (!this.titleInput) {
        this.$bvToast.toast("Cannot save note without a title ✘", {
          variant: "danger",
          noCloseButton: true,
          toaster: "b-toaster-bottom-right",
        });
        return;
      }

      const reservedCharacters = /[<>:"/\\|?*]/;
      if (reservedCharacters.test(this.titleInput)) {
        this.$bvToast.toast(
          'Due to filename restrictions, the following characters are not allowed in a note title: <>:"/\\|?*',
          {
            variant: "danger",
            noCloseButton: true,
            toaster: "b-toaster-bottom-right",
          }
        );
        return;
      }

      if (this.currentNote.lastModified == null) { // New Note
        api(`/api/notes`, {
          body: {
            title: this.titleInput,
            content: newContent,
          },
        })
          .then(this.saveNoteResponseHandler)
          .catch((error) => {
            if (error.handled) {
              return;
            } else if (
              typeof error.response !== "undefined" &&
              error.response.status == 409
            ) {
              this.existingTitleToast();
            } else {
              EventBus.$emit("unhandledServerError", error);
            }
          });
      } else if (newContent != this.currentNote.content || this.titleInput != this.currentNote.title) { // Modified Note
        api(`/api/notes/${encodeURIComponent(this.currentNote.title)}`, {
          method: "PATCH",
          body: {
            newTitle: this.titleInput,
            newContent: newContent,
          },
        })
          .then(this.saveNoteResponseHandler)
          .catch((error) => {
            if (error.handled) {
              return;
            } else if (
              typeof error.response !== "undefined" &&
              error.response.status == 409
            ) {
              this.existingTitleToast();
            } else {
              EventBus.$emit("unhandledServerError", error);
            }
          });
      } else { // No Change
        localStorage.removeItem(this.currentNote.title);
        this.setEditMode(false);
        this.noteSavedToast();
      }
    },

    saveNoteResponseHandler(response) {
      localStorage.removeItem(this.currentNote.title);
      this.currentNote = new Note(
        response.title,
        response.lastModified,
        response.content
      );
      EventBus.$emit("updateNoteTitle", this.currentNote.title);
      history.replaceState(null, "", this.currentNote.href);
      this.setEditMode(false);
      this.noteSavedToast();
    },

    noteSavedToast() {
      this.$bvToast.toast("Note saved ✓", {
        variant: "success",
        noCloseButton: true,
        toaster: "b-toaster-bottom-right",
      });
    },

    cancelNote() {
      localStorage.removeItem(this.currentNote.title);
      if (this.currentNote.lastModified == null) {
        // Cancelling a new note
        EventBus.$emit("navigate", constants.basePaths.home);
      } else {
        this.setEditMode(false);
      }
    },

    confirmCancelNote() {
      let newContent = this.getEditorContent();
      if (
        newContent != this.currentNote.content ||
        this.titleInput != this.currentNote.title
      ) {
        this.$bvModal
          .msgBoxConfirm(
            `Are you sure you want to close the note '${this.currentNote.title}' without saving?`,
            {
              centered: true,
              title: "Confirm Closure",
              okTitle: "Yes, Close",
              okVariant: "warning",
            }
          )
          .then((response) => {
            if (response == true) {
              this.cancelNote();
            }
          });
      } else {
        this.cancelNote();
      }
    },

    deleteNote() {
      this.$bvModal
        .msgBoxConfirm(
          `Are you sure you want to delete the note '${this.currentNote.title}'?`,
          {
            centered: true,
            title: "Confirm Deletion",
            okTitle: "Delete",
            okVariant: "danger",
          }
        )
        .then((response) => {
          if (!response) {
            return;
          }

          api(`/api/notes/${encodeURIComponent(this.currentNote.title)}`, {
            method: "DELETE",
          })
            .then(() => {
              this.$emit("note-deleted");
              EventBus.$emit("navigate", constants.basePaths.home);
            })
            .catch((error) => {
              if (!error.handled) {
                EventBus.$emit("unhandledServerError", error);
              }
            });
        });
    },

    init() {
      this.currentNote = null;
      if (this.titleToLoad) {
        this.loadNote(this.titleToLoad);
        this.setEditMode(false);
      } else {
        this.currentNote = new Note();
        this.setEditMode(true);
      }
    },
  },

  created() {
    // 'e' to edit
    Mousetrap.bind("e", () => {
      if (this.editMode == false && this.canModify) {
        this.setEditMode(true);
      }
    });

    // 'ctrl+s' to save
    // Mousetrap.bind("ctrl+s", () => {
    //   if (this.editMode == true) {
    //     this.saveNote();
    //     return false;
    //   }
    // });

    this.init();
  },
};
</script>

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
