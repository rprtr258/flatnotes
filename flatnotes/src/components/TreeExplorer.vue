<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, type ComponentPublicInstance } from "vue";
import { eventBus } from "../eventBus";
import Icon from "../ui/Icon.vue";
import LoadingIndicator from "./LoadingIndicator.vue";
import api from "../api";
import type { SearchResultModel } from "../types";

const props = defineProps<{ activeTitle: string | null }>();

type TreeNode = {
  name: string;
  path: string; // full note title (or directory prefix)
  isDir: boolean;
  children: TreeNode[];
};

type FlatRow = {
  name: string;
  path: string;
  isDir: boolean;
  depth: number;
  expanded: boolean;
};

const notes = ref<SearchResultModel[] | null>(null);
const loadingFailed = ref(false);
const expanded = ref<Set<string>>(new Set());
const activeRowEl = ref<HTMLElement | null>(null);

function noteHref(title: string): string {
  return "/note/" + title.split("/").map(encodeURIComponent).join("/");
}

function buildTree(titles: string[]): TreeNode[] {
  const root: TreeNode = { name: "", path: "", isDir: true, children: [] };
  for (const title of titles) {
    const parts = title.split("/");
    let node = root;
    let prefix = "";
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      prefix = prefix ? prefix + "/" + part : part;
      const isLeaf = i === parts.length - 1;
      let child = node.children.find((c) => c.name === part && c.isDir === !isLeaf);
      if (!child) {
        child = { name: part, path: prefix, isDir: !isLeaf, children: [] };
        node.children.push(child);
      }
      node = child;
    }
  }
  const sortNodes = (nodes: TreeNode[]): void => {
    nodes.sort((a, b) => {
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      return a.name.localeCompare(b.name);
    });
    nodes.forEach((n) => sortNodes(n.children));
  };
  sortNodes(root.children);
  return root.children;
}

// Flatten the tree into visible rows. A node is visible only when every
// ancestor directory is expanded.
const rows = computed<FlatRow[]>(() => {
  if (!notes.value) return [];
  const tree = buildTree(notes.value.map((n) => n.title));
  const out: FlatRow[] = [];
  const walk = (nodes: TreeNode[], depth: number, ancestorsExpanded: boolean): void => {
    for (const node of nodes) {
      if (!ancestorsExpanded) continue;
      const isOpen = expanded.value.has(node.path);
      out.push({
        name: node.name,
        path: node.path,
        isDir: node.isDir,
        depth,
        expanded: isOpen,
      });
      if (node.isDir) {
        walk(node.children, depth + 1, ancestorsExpanded && isOpen);
      }
    }
  };
  walk(tree, 0, true);
  return out;
});

function toggle(path: string): void {
  if (expanded.value.has(path)) {
    expanded.value.delete(path);
  } else {
    expanded.value.add(path);
  }
}

function expandAncestors(title: string): void {
  const parts = title.split("/");
  let prefix = "";
  for (let i = 0; i < parts.length - 1; i++) {
    prefix = prefix ? prefix + "/" + parts[i] : parts[i];
    expanded.value.add(prefix);
  }
}

function setActiveRow(el: Element | ComponentPublicInstance | null): void {
  activeRowEl.value = (el as HTMLElement) ?? null;
}

watch(
  () => [props.activeTitle, notes.value] as const,
  ([title]) => {
    if (!title || !notes.value) return;
    expandAncestors(title);
    void nextTick(() => activeRowEl.value?.scrollIntoView({ block: "nearest" }));
  },
  { immediate: true }
);

function navigate(href: string, event?: Event): void {
  eventBus.emit("navigate", { href, event });
}

function getNotes(): void {
  loadingFailed.value = false;
  api<SearchResultModel[]>("/api/search", {
    params: { term: "*", sort: "title", order: "asc" },
  })
    .then((response) => {
      notes.value = response;
    })
    .catch((error) => {
      loadingFailed.value = true;
      if (!(error as { handled?: boolean }).handled) {
        eventBus.emit("unhandled-server-error", { error });
      }
    });
}

onMounted(getNotes);
</script>

<template>
  <div class="tree-explorer d-flex flex-column h-100">
    <div class="tree-body flex-grow-1 overflow-auto">
      <LoadingIndicator
        v-if="notes == null"
        :failed="loadingFailed"
        :show-loader="false"
        failed-message="Failed to load notes"
      />
      <p v-else-if="rows.length === 0" class="empty px-2">No notes</p>
      <template v-else>
        <component
          :is="row.isDir ? 'div' : 'a'"
          v-for="row in rows"
          :key="row.path"
          class="tree-row"
          :class="[row.isDir ? 'dir' : 'file', { active: !row.isDir && props.activeTitle === row.path }]"
          :style="{ paddingLeft: 8 + row.depth * 14 + 'px' }"
          :href="row.isDir ? undefined : noteHref(row.path)"
          :ref="!row.isDir && props.activeTitle === row.path ? setActiveRow : undefined"
          @click.prevent="row.isDir ? toggle(row.path) : navigate(noteHref(row.path), $event)"
        >
          <Icon v-if="row.isDir" :name="row.expanded ? 'caret-down-fill' : 'caret-right-fill'" />
          <Icon :name="row.isDir ? 'folder' : 'file-text'" />
          <span class="label">{{ row.name }}</span>
        </component>
      </template>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../colours";

.tree-explorer {
  min-height: 0;
}

.tree-body {
  min-height: 0;
}

.empty {
  color: var(--colour-text-muted);
  font-size: 0.9rem;
}

.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  cursor: pointer;
  border-radius: 4px;
  white-space: nowrap;
  font-size: 0.9rem;
  color: var(--colour-text);
  text-decoration: none;

  .label {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  &:hover {
    background-color: var(--colour-background-tint);
  }

  &.active {
    background-color: var(--colour-background-highlight);
    font-weight: 600;
  }
}
</style>
