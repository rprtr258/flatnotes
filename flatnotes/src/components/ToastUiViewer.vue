<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import Viewer from "@toast-ui/editor/viewer";
import type { EditorPlugin, CustomHTMLRenderer } from "@toast-ui/editor";

const props = defineProps<{
  initialValue?: string;
  options?: {
    customHTMLRenderer?: CustomHTMLRenderer;
    plugins?: EditorPlugin[];
    extendedAutolinks?: boolean;
    [key: string]: unknown;
  };
}>();

const el = ref<HTMLElement | null>(null);
let viewerInstance: Viewer | null = null;

function create(): void {
  if (!el.value) return;
  viewerInstance = new Viewer({
    el: el.value,
    initialValue: props.initialValue,
    customHTMLRenderer: props.options?.customHTMLRenderer,
    plugins: props.options?.plugins,
    extendedAutolinks: props.options?.extendedAutolinks,
  });
}

onMounted(create);

watch(
  () => props.initialValue,
  (value) => {
    if (viewerInstance && value != null) {
      viewerInstance.setMarkdown(value);
    }
  },
);

onBeforeUnmount(() => {
  viewerInstance?.destroy();
  viewerInstance = null;
});
</script>

<template>
  <div ref="el"></div>
</template>
