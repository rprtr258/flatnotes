<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import * as constants from "../constants";
import * as helpers from "../helpers";
import { eventBus } from "../eventBus";
import LoadingIndicator from "./LoadingIndicator.vue";
import SearchInput from "./SearchInput.vue";
import Icon from "../ui/Icon.vue";
import { vTooltip } from "../ui/tooltip";
import { SearchResult } from "../classes";
import api from "../api";
import type { SearchResultModel } from "../types";

const props = defineProps<{ searchTerm: string | null }>();

const searchFailed = ref(false);
const searchFailedMessage = ref("Failed to load Search Results");
const searchFailedIcon = ref<string | null>(null);
const searchResults = ref<SearchResult[] | null>(null);
const searchResultsIncludeHighlights = ref(false);
const sortBy = ref(0);
const showHighlights = ref(true);

// The original referenced an undefined `sortOptions` (the Order select rendered
// empty). Populated here with the three sort options so the control works as
// clearly intended.
const sortOptions = [
  constants.searchSortOptions.score,
  constants.searchSortOptions.title,
  constants.searchSortOptions.lastModified,
];

interface ResultGroup {
  name: string;
  searchResults: SearchResult[];
}

const sortByIsGrouped = computed(() => sortBy.value === constants.searchSortOptions.title);

const resultsGrouped = computed<ResultGroup[]>(() => {
  if (!searchResults.value) return [];
  switch (sortBy.value) {
    case constants.searchSortOptions.title:
      return resultsByTitle();
    case constants.searchSortOptions.lastModified:
      return resultsByLastModified();
    default:
      return resultsByScore();
  }
});

watch(
  () => props.searchTerm,
  () => init(),
);

watch(sortBy, (value) => {
  helpers.setSearchParam(constants.params.sortBy, String(value));
});

watch(showHighlights, (value) => {
  helpers.setSearchParam(constants.params.showHighlights, String(value));
});

function getSearchResults(): void {
  searchFailed.value = false;
  searchResultsIncludeHighlights.value = false;
  api<SearchResultModel[]>("/api/search", {
    params: {
      term: props.searchTerm ?? "",
    },
  })
    .then((response) => {
      searchResults.value = [];
      if (response.length === 0) {
        searchFailedIcon.value = "search";
        searchFailedMessage.value = "No Results";
        searchFailed.value = true;
      } else {
        response.forEach((responseItem) => {
          const searchResult = new SearchResult(responseItem);
          searchResults.value!.push(searchResult);
          if (searchResultsIncludeHighlights.value === false && searchResult.includesHighlights) {
            searchResultsIncludeHighlights.value = true;
          }
        });
      }
    })
    .catch((error) => {
      if (!(error as { handled?: boolean }).handled) {
        searchFailed.value = true;
        eventBus.emit("unhandled-server-error", { error });
      }
    });
}

function resultsByScore(): ResultGroup[] {
  return [
    {
      name: "_",
      searchResults: [...searchResults.value!].sort((i, j) => j.score - i.score),
    },
  ];
}

function resultsByLastModified(): ResultGroup[] {
  return [
    {
      name: "_",
      searchResults: searchResults.value!.sort(
        (a, b) => (b.lastModified ?? 0) - (a.lastModified ?? 0),
      ),
    },
  ];
}

function resultsByTitle(): ResultGroup[] {
  const notesGroupedDict: Record<string, SearchResult[]> = {};
  const specialCharGroupTitle = "#";
  [specialCharGroupTitle, ...constants.alphabet].forEach((group) => {
    notesGroupedDict[group] = [];
  });

  searchResults.value!.forEach((searchResult) => {
    const firstCharUpper = searchResult.title[0].toUpperCase();
    if (constants.alphabet.includes(firstCharUpper)) {
      notesGroupedDict[firstCharUpper].push(searchResult);
    } else {
      notesGroupedDict[specialCharGroupTitle].push(searchResult);
    }
  });

  const notesGroupedArray: ResultGroup[] = [];
  Object.entries(notesGroupedDict).forEach(([name, results]) => {
    if (results.length) {
      notesGroupedArray.push({
        name,
        searchResults: results.sort((a, b) => a.title.localeCompare(b.title)),
      });
    }
  });

  notesGroupedArray.sort((a, b) => a.name.localeCompare(b.name));
  return notesGroupedArray;
}

function openNote(href: string, event?: Event): void {
  eventBus.emit("navigate", { href, event });
}

const sortOptionStrings: Record<number, string> = {
  0: "Score",
  1: "Title",
  2: "Last Modified",
};

function sortOptionToString(sortOption: number): string {
  return sortOptionStrings[sortOption];
}

function init(): void {
  sortBy.value = helpers.getSearchParamInt(constants.params.sortBy, 0) ?? 0;
  showHighlights.value = helpers.getSearchParamBool(constants.params.showHighlights, true) ?? true;
  getSearchResults();
}

onMounted(init);
</script>

<template>
  <div class="mb-4">
    <!-- Input -->
    <SearchInput :initial-value="searchTerm ?? undefined" class="mb-1" />

    <!-- Searching -->
    <div
      v-if="searchResults == null || searchResults.length === 0"
      class="h-100 d-flex flex-column justify-content-center"
    >
      <LoadingIndicator
        :failed="searchFailed"
        :failed-icon="searchFailedIcon ?? undefined"
        :failed-message="searchFailedMessage"
      />
    </div>
    <div v-else>
      <!-- Controls -->
      <div class="mb-3">
        <select v-model="sortBy" class="bttn sort-select">
          <option v-for="option in sortOptions" :key="option" :value="option" class="p-0">
            Order: {{ sortOptionToString(option) }}
          </option>
        </select>

        <button
          v-if="searchResultsIncludeHighlights"
          type="button"
          class="bttn"
          @click="showHighlights = !showHighlights"
        >
          <Icon :name="showHighlights ? 'eye-slash' : 'eye'" />
          {{ showHighlights ? "Hide" : "Show" }} Highlights
        </button>
      </div>

      <!-- Results -->
      <div
        v-for="group in resultsGrouped"
        :key="group.name"
        :class="{ 'mb-5': sortByIsGrouped }"
      >
        <p v-if="sortByIsGrouped" class="group-name">{{ group.name }}</p>
        <div
          v-for="result in group.searchResults"
          :key="result.title"
          class="bttn result"
          :class="{ 'mb-3': searchResultsIncludeHighlights && showHighlights }"
        >
          <p>{{result.score.toFixed(5)}}</p>
          <a :href="result.href" @click.prevent="openNote(result.href, $event)">
            <div class="d-flex justify-content-between">
              <p
                class="result-title"
                v-html="
                  showHighlights ? result.titleHighlightsOrTitle : result.title
                "
              />
              <span
                class="last-modified d-none d-md-block"
                v-tooltip="'Last Modified'"
              >
                {{ result.lastModifiedAsString }}
              </span>
            </div>
            <p
              v-show="showHighlights"
              class="result-contents"
              v-html="result.contentHighlights"
            />
            <div v-show="showHighlights">
              <span v-for="tag in result.tagMatches" :key="tag" class="tag me-2"
                >#{{ tag }}</span
              >
            </div>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

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
