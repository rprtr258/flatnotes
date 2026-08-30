<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import Editor from "@toast-ui/editor";
import type { PreviewStyle, EditorType, EditorPlugin, CustomHTMLRenderer } from "@toast-ui/editor";
const props = defineProps<{
  initialValue?: string;
  initialEditType?: EditorType;
  previewStyle?: PreviewStyle;
  height?: string;
  options?: {
    customHTMLRenderer?: CustomHTMLRenderer;
    plugins?: EditorPlugin[];
    [key: string]: unknown;
  };
}>();

const emit = defineEmits<{ change: [editorType: EditorType] }>();

const el = ref<HTMLElement | null>(null);
let editorInstance: Editor | null = null;

onMounted(() => {
  if (!el.value) return;
  editorInstance = new Editor({
    el: el.value,
    initialValue: props.initialValue,
    initialEditType: props.initialEditType ?? "markdown",
    previewStyle: props.previewStyle ?? "tab",
    height: props.height ?? "auto",
    customHTMLRenderer: props.options?.customHTMLRenderer,
    plugins: props.options?.plugins,
    events: {
      change: (editorType: EditorType) => emit("change", editorType),
    },
  });
});

watch(
  () => props.initialValue,
  () => {
    // initialValue is only used at construction; ignored after mount (matches vue-editor behaviour)
  },
);

onBeforeUnmount(() => {
  editorInstance?.destroy();
  editorInstance = null;
});

function invoke<T = unknown>(method: string, ...args: unknown[]): T {
  if (!editorInstance) {
    throw new Error("editor not initialised");
  }
  // Call as a method on the instance so Toast UI's internal `this` is bound
  // (its getMarkdown() etc. call this.isMarkdownMode()).
  const fn = (editorInstance as unknown as Record<string, (...a: unknown[]) => T>)[method];
  return fn.apply(editorInstance, args);
}

defineExpose({ invoke });
</script>

<template>
  <div ref="el"></div>
</template>
