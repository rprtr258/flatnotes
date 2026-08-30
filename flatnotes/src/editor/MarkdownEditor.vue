<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import { EditorState } from "@codemirror/state";
import {
  EditorView,
  keymap,
  highlightSpecialChars,
  drawSelection,
  highlightActiveLine,
} from "@codemirror/view";
import { history, defaultKeymap, historyKeymap } from "@codemirror/commands";
import { searchKeymap } from "@codemirror/search";
import { syntaxHighlighting, defaultHighlightStyle } from "@codemirror/language";
import { markdown, markdownLanguage } from "@codemirror/lang-markdown";
import { languages } from "@codemirror/language-data";
import { GFM } from "@lezer/markdown";
import { cleanPlugins } from "./clean";
import "./editor.scss";

const props = defineProps<{
  initialValue?: string;
  readOnly?: boolean;
}>();

const emit = defineEmits<{ change: [] }>();

const hostEl = ref<HTMLElement | null>(null);
let view: EditorView | null = null;

onMounted(() => {
  if (!hostEl.value) return;
  view = new EditorView({
    state: EditorState.create({
      doc: props.initialValue ?? "",
      extensions: [
        EditorState.readOnly.of(!!props.readOnly),
        EditorView.editable.of(!props.readOnly),
        EditorView.lineWrapping,
        highlightSpecialChars(),
        drawSelection(),
        highlightActiveLine(),
        history(),
        keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
        markdown({
          base: markdownLanguage,
          codeLanguages: languages,
          addKeymap: true,
          extensions: [GFM],
        }),
        syntaxHighlighting(defaultHighlightStyle),
        ...cleanPlugins(),
        EditorView.updateListener.of((tr) => {
          if (tr.docChanged) emit("change");
        }),
      ],
    }),
    parent: hostEl.value,
  });
});

onBeforeUnmount(() => {
  view?.destroy();
  view = null;
});

watch(
  () => props.initialValue,
  (incoming) => {
    if (!view || incoming == null) return;
    const current = view.state.doc.toString();
    if (incoming === current) return;
    view.dispatch({
      changes: { from: 0, to: current.length, insert: incoming },
    });
  },
);

function getContent(): string {
  return view?.state.doc.toString() ?? "";
}

defineExpose({ getContent });
</script>

<template>
  <div ref="hostEl" class="cm-host"></div>
</template>
