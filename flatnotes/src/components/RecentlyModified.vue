<script>
import EventBus from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import { SearchResult } from "../classes";
import api from "../api";

export default {
  data() {
    return {
      notes: null,
      tags: null,
      loadingFailed: false,
      loadingFailedMessage: "Failed to load notes",
      loadingFailedIcon: null,
    };
  },

  methods: {
    getNotes() {
      let parent = this;
      this.loadingFailed = false;
      api("/api/search", {
          params: {
            term: "*",
            sort: "lastModified",
            order: "desc",
            limit: this.maxNotes,
          },
        })
        .then((response) => {
          parent.notes = [];
          if (response.length) {
            response.forEach((searchResult) => {
              parent.notes.push(SearchResult(searchResult));
            });
          } else {
            parent.loadingFailedMessage = "Click the 'New' button at the top of the page to add your first note";
            parent.loadingFailedIcon = "pencil";
            parent.loadingFailed = true;
          }
        })
        .catch((error) => {
          parent.loadingFailed = true;
          if (!error.handled) {
            EventBus.$emit("unhandledServerError", error);
          }
        });
    },

    getTags() {
      let parent = this;
      this.loadingFailed = false;
      api("/api/tags")
        .then((response) => {
          parent.tags = [];
          if (response.length) {
            response.forEach((tag) => {
              parent.tags.push(tag);
            });
          } else {
            parent.loadingFailedMessage = "No tags";
            parent.loadingFailedIcon = "pencil";
            parent.loadingFailed = true;
          }
        })
        .catch((error) => {
          parent.loadingFailed = true;
          if (!error.handled) {
            EventBus.$emit("unhandledServerError", error);
          }
        });
    },

    openNote(href, event) {
      EventBus.$emit("navigate", href, event);
    },

    openTag(tag, event) {
      EventBus.$emit("navigate", "/search?term=" + encodeURIComponent("#" + tag), event);
    },
  },

  created() {
    this.getNotes();
    this.getTags();
  },
};
</script>

<style lang="scss" scoped>
@import "../colours";

.mini-header {
  font-size: 12px;
  font-weight: bold;
  color: var(--colour-text-very-muted);
}
</style>
