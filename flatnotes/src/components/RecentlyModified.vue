<script>
import EventBus from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import { SearchResult } from "../classes";
import api from "../api";

export default {
  components: {
    LoadingIndicator,
  },

  props: {
    maxNotes: { type: Number },
  },

  data: function () {
    return {
      notes: null,
      tags: null,
      loadingFailed: false,
      loadingFailedMessage: "Failed to load notes",
      loadingFailedIcon: null,
    };
  },

  methods: {
    getNotes: function () {
      let parent = this;
      this.loadingFailed = false;
      api
        .get("/api/search", {
          params: {
            term: "*",
            sort: "lastModified",
            order: "desc",
            limit: this.maxNotes,
          },
        })
        .then(function (response) {
          parent.notes = [];
          if (response.data.length) {
            response.data.forEach(function (searchResult) {
              parent.notes.push(new SearchResult(searchResult));
            });
          } else {
            parent.loadingFailedMessage = "Click the 'New' button at the top of the page to add your first note";
            parent.loadingFailedIcon = "pencil";
            parent.loadingFailed = true;
          }
        })
        .catch(function (error) {
          parent.loadingFailed = true;
          if (!error.handled) {
            EventBus.$emit("unhandledServerError");
          }
        });
    },

    getTags: function () {
      let parent = this;
      this.loadingFailed = false;
      api
        .get("/api/tags")
        .then(function (response) {
          parent.tags = [];
          if (response.data.length) {
            response.data.forEach(function (tag) {
              parent.tags.push(tag);
            });
          } else {
            parent.loadingFailedMessage = "No tags";
            parent.loadingFailedIcon = "pencil";
            parent.loadingFailed = true;
          }
        })
        .catch(function (error) {
          parent.loadingFailed = true;
          if (!error.handled) {
            EventBus.$emit("unhandledServerError");
          }
        });
    },

    openNote: function (href, event) {
      EventBus.$emit("navigate", href, event);
    },

    openTag: function (tag, event) {
      EventBus.$emit("navigate", "/search?term=" + encodeURIComponent("#" + tag), event);
    },
  },

  created: function () {
    this.getNotes();
    this.getTags();
  },
};
</script>

<template>
  <div
    class="justify-content-top"
  >
    <!-- Loading -->
    <div
      v-if="notes == null || notes.length == 0"
      class="h-100 d-flex flex-column justify-content-center"
    >
      <LoadingIndicator
        :failed="loadingFailed"
        :failedMessage="loadingFailedMessage"
        :failedBootstrapIcon="loadingFailedIcon"
        :show-loader="false"
      />
    </div> <!-- Notes Loaded -->
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
          >{{ note.title }}</a>
        </div>
        <div
          class="d-flex flex-column align-items-center"
        >
          <p class="mini-header mb-1">TAGS</p>
          <a
            v-for="tag in tags"
            :key="tag"
            class="bttn"
            :href="tag"
            @click.prevent="openTag(tag, $event)"
          >#{{ tag }}</a>
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
