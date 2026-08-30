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
import {
  syntaxHighlighting,
  HighlightStyle,
} from "@codemirror/language";
import { tags } from "@lezer/highlight";
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

// Theme-aware syntax highlighting. Colours are CSS variables, so the same
// style keeps contrast on both light and dark themes (the built-in
// defaultHighlightStyle is light-only and goes near-invisible on dark).
const highlightStyle = HighlightStyle.define([
  { tag: tags.link, color: "var(--colour-brand)", textDecoration: "underline" },
  { tag: tags.heading, fontWeight: "bold", textDecoration: "underline" },
  { tag: tags.strong, fontWeight: "bold" },
  { tag: tags.emphasis, fontStyle: "italic" },
  { tag: tags.strikethrough, textDecoration: "line-through" },
  { tag: tags.keyword, color: "var(--colour-brand)" },
  { tag: [tags.atom, tags.bool, tags.url, tags.contentSeparator, tags.labelName], color: "var(--colour-brand)" },
  { tag: [tags.literal, tags.inserted, tags.number], color: "var(--colour-text)" },
  { tag: [tags.string, tags.deleted], color: "var(--colour-brand)" },
  { tag: [tags.regexp, tags.escape, tags.special(tags.string)], color: "var(--colour-brand)" },
  { tag: tags.variableName, color: "var(--colour-text)" },
  { tag: tags.local(tags.variableName), color: "var(--colour-text)" },
  { tag: tags.definition(tags.variableName), color: "var(--colour-text)" },
  { tag: tags.special(tags.variableName), color: "var(--colour-text)" },
  { tag: tags.propertyName, color: "var(--colour-text)" },
  { tag: tags.definition(tags.propertyName), color: "var(--colour-text)" },
  { tag: [tags.typeName, tags.namespace, tags.className, tags.macroName], color: "var(--colour-text)" },
  { tag: tags.comment, color: "var(--colour-text-muted)" },
  { tag: tags.meta, color: "var(--colour-text)" },
  { tag: tags.invalid, color: "#f00" },
]);

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
        syntaxHighlighting(highlightStyle),
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
