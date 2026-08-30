<script setup lang="ts">
import { computed } from "vue";
import * as constants from "../constants";
import { eventBus } from "../eventBus";
import Logo from "./Logo.vue";
import Icon from "../ui/Icon.vue";
import { vTooltip } from "../ui/tooltip";

const props = withDefaults(
  defineProps<{
    showLogo?: boolean;
    authType?: string | null;
    darkTheme?: boolean;
  }>(),
  {
    showLogo: true,
    authType: null,
    darkTheme: false,
  },
);

defineEmits<{
  logout: [];
  toggleTheme: [];
  search: [];
}>();

const azHref = computed(() => {
  const sp = new URLSearchParams();
  sp.set(constants.params.searchTerm, "*");
  sp.set(constants.params.sortBy, String(constants.searchSortOptions.title));
  sp.set(constants.params.showHighlights, String(false));
  return `${constants.basePaths.search}?${sp.toString()}`;
});

const showLogOutButton = computed(
  () =>
    props.authType != null &&
    ![constants.authTypes.none, constants.authTypes.readOnly].includes(props.authType as never),
);

const showNewButton = computed(
  () => props.authType != null && props.authType !== constants.authTypes.readOnly,
);

function navigate(href: string, event?: Event): void {
  eventBus.emit("navigate", { href, event });
}
</script>

<template>
  <div class="d-flex justify-content-between align-items-center">
    <!-- Logo -->
    <a
      :href="constants.basePaths.home"
      @click.prevent="navigate(constants.basePaths.home, $event)"
    >
      <Logo :class="{ invisible: !showLogo }" responsive />
    </a>

    <!-- Buttons -->
    <div class="d-flex">
      <!-- Log Out -->
      <button
        v-if="showLogOutButton"
        type="button"
        class="bttn"
        @click="$emit('logout')"
      >
        <Icon name="box-arrow-right" /> Log Out
      </button>

      <!-- New Note -->
      <a
        v-if="showNewButton"
        :href="constants.basePaths.new"
        class="bttn"
        @click.prevent="navigate(constants.basePaths.new, $event)"
        v-tooltip="'Create a New Note'"
      >
        <Icon name="plus-circle" /> New
      </a>

      <!-- Theme Toggle -->
      <button
        type="button"
        id="theme-button"
        class="bttn"
        @click="$emit('toggleTheme')"
        v-tooltip="'Toggle Theme'"
      >
        <Icon :name="darkTheme ? 'sun' : 'moon'" />
      </button>

      <!-- A-Z -->
      <a
        :href="azHref"
        class="bttn"
        @click.prevent="navigate(azHref, $event)"
        v-tooltip="'Show All Notes'"
        >A-Z</a
      >

      <!-- Search -->
      <button
        type="button"
        id="search-button"
        class="bttn"
        @click="$emit('search')"
        v-tooltip="'Search (Keyboard Shortcut: /)'"
      >
        <Icon name="search" />
      </button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
// Use visibility hidden instead of v-show to maintain consistent height
.invisible {
  visibility: hidden;
}
</style>
