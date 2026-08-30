import type { Extension, EditorState } from "@codemirror/state";
import { Decoration, type DecorationSet } from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import { decoratorStateField, invisibleDecoration, isCursorInRange } from "../util";

const MARK_CLASS: Record<string, string> = {
  Emphasis: "sb-em",
  StrongEmphasis: "sb-strong",
  Strikethrough: "sb-strike",
  InlineCode: "sb-code",
};

// The Lezer delimiter mark node name wrapping each emphasis variant.
const MARK_NODE: Record<string, string> = {
  Emphasis: "EmphasisMark",
  StrongEmphasis: "EmphasisMark",
  Strikethrough: "StrikethroughMark",
  InlineCode: "CodeMark",
};

// Hides the leading/trailing delimiter marks of emphasis/inline-code nodes
// and marks the inner content. Honours the cursor-adjacency rule: when the
// cursor is inside the node the raw markdown is left untouched.
export function emphasisPlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; value: Decoration }[] = [];
    syntaxTree(state).iterate({
      enter(node) {
        const cls = MARK_CLASS[node.name];
        if (!cls) return;
        if (isCursorInRange(state, node.from, node.to)) return;
        const markName = MARK_NODE[node.name];
        const first = node.node.firstChild;
        const last = node.node.lastChild;
        if (!first || !last || first.name !== markName || last.name !== markName) return;
        if (first.to >= last.from) return; // empty content
        decorations.push({ from: node.from, value: invisibleDecoration });
        decorations.push({ from: first.to, value: Decoration.mark({ class: cls }) });
        decorations.push({ from: last.from, value: invisibleDecoration });
      },
    });
    return Decoration.set(
      decorations.map((d) => d.value.range(d.from)),
      true,
    );
  });
}
