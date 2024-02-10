<script>
import * as constants from "../constants";

import EventBus from "../eventBus";

export default {
  props: { initialValue: { type: String } },

  data() {
    return {
      searchTermInput: null,
      includeHighlightClass: false,
    };
  },

  watch: {
    initialValue() {
      this.init();
    },
  },

  methods: {
    search() {
      if (this.searchTermInput) {
        this.searchTermInput = this.searchTermInput.trim();
      }
      if (this.searchTermInput) {
        EventBus.$emit(
          "navigate",
          `${constants.basePaths.search}?${
            constants.params.searchTerm
          }=${encodeURIComponent(this.searchTermInput)}`
        );
      } else {
        this.$bvToast.toast("Please enter a search term ✘", {
          variant: "danger",
          noCloseButton: true,
          toaster: "b-toaster-bottom-right",
        });
      }
    },

    highlightSearchInput() {
      let parent = this;
      this.includeHighlightClass = true;
      setTimeout(() => {
        parent.includeHighlightClass = false;
      }, 1500);
    },

    init() {
      this.searchTermInput = this.initialValue;
    },
  },

  mounted() {
    this.$refs.input.focus();
    this.$refs.input.select();
  },

  created() {
    EventBus.$on("highlight-search-input", this.highlightSearchInput);
    this.init();
  },
};
</script>

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

  svg {
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
