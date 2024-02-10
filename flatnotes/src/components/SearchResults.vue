<script>
import * as constants from "../constants";
import * as helpers from "../helpers";

import EventBus from "../eventBus";
import { SearchResult } from "../classes";
import api from "../api";

export default {
  watch: {
    searchTerm() {
      this.init();
    },

    sortBy() {
      helpers.setSearchParam(constants.params.sortBy, this.sortBy);
    },

    showHighlights() {
      helpers.setSearchParam(
        constants.params.showHighlights,
        this.showHighlights
      );
    },
  },

  methods: {
    getSearchResults() {
      this.searchFailed = false;
      this.searchResultsIncludeHighlights = false;
      api("/api/search", {
        params: {
          term: this.searchTerm,
        },
      })
        .then((response) => {
          this.searchResults = [];
          if (response.length == 0) {
            this.searchFailedIcon = "search";
            this.searchFailedMessage = "No Results";
            this.searchFailed = true;
          } else {
            response.forEach((responseItem) => {
              const searchResult = SearchResult(responseItem);
              this.searchResults.push(searchResult);
              this.searchResultsIncludeHighlights = !this.searchResultsIncludeHighlights && searchResult.includesHighlights;
            });
          }
        })
        .catch((error) => {
          if (!error.handled) {
            this.searchFailed = true;
            EventBus.$emit("unhandledServerError", error);
          }
        });
    },

    openNote(href, event) {
      EventBus.$emit("navigate", href, event);
    },

    init() {
      this.sortBy = helpers.getSearchParamInt(constants.params.sortBy, 0);
      this.showHighlights = helpers.getSearchParamBool(
        constants.params.showHighlights,
        true
      );
      this.getSearchResults();
    },
  },

  created() {
    this.init();
  },
};
</script>

<style lang="scss" scoped>
@import "../colours";

.sort-select {
  padding-inline: 6px;
}

.group-name {
  padding-left: 8px;
  font-weight: bold;
  font-size: 32px;
  color: var(--colour-text-very-muted);
  margin-bottom: 8px;
}

.result p {
  margin: 0;
}

.result-title {
  color: var(--colour-text);
}

.last-modified {
  color: var(--colour-text-muted);
  font-size: 12px;
}

.result-contents {
  color: var(--colour-text-muted);
}
</style>

<style lang="scss">
@import "../colours";

.match {
  font-weight: bold;
  color: var(--colour-brand);
}

.tag {
  color: white;
  font-size: 14px;
  background-color: var(--colour-brand);
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
