import type { Extension, EditorState } from "@codemirror/state";
import { Decoration, type DecorationSet } from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import type { SyntaxNodeRef } from "@lezer/common";
import { decoratorStateField, invisibleDecoration, isCursorInRange } from "../util";

// Lezer child node names (verified in node_modules/@lezer/markdown):
//   Link, Image  -> container nodes
//   LinkMark     -> the "[", "]", "(", ")", "!" delimiter marks
//   URL          -> the destination URL

// For a Link/Image node, locate the opening "[" and closing "]" LinkMark
// children and return the text range between them. Returns null when the
// node is not a bracketed link/image (e.g. an autolink).
function linkTextRange(
  state: EditorState,
  ref: SyntaxNodeRef,
): { textFrom: number; textTo: number } | null {
  let open: number | null = null;
  let close: number | null = null;
  for (let child = ref.node.firstChild; child; child = child.nextSibling) {
    if (child.name !== "LinkMark") continue;
    const text = state.sliceDoc(child.from, child.to);
    if (text === "[" && open === null) {
      open = child.to;
    } else if (text === "]" && open !== null && close === null) {
      close = child.from;
    }
  }
  if (open === null || close === null || open >= close) return null;
  return { textFrom: open, textTo: close };
}

// Hides the surrounding markup of inline links/images and marks their text.
// Honours the cursor-adjacency rule: raw markdown is shown when the cursor is
// inside the node.
export function linkPlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; to: number; value: Decoration }[] = [];
    syntaxTree(state).iterate({
      enter(node) {
        if (node.name !== "Link" && node.name !== "Image") return;
        if (isCursorInRange(state, node.from, node.to)) return;
        const range = linkTextRange(state, node);
        if (!range) return; // autolink or unrecognised shape: leave raw
        const cls = node.name === "Image" ? "sb-image" : "sb-link";
        // Hide the leading markup ("![", "[") and trailing markup ("](url)").
        decorations.push({
          from: node.from,
          to: range.textFrom,
          value: invisibleDecoration,
        });
        decorations.push({
          from: range.textFrom,
          to: range.textTo,
          value: Decoration.mark({ class: cls }),
        });
        decorations.push({
          from: range.textTo,
          to: node.to,
          value: invisibleDecoration,
        });
      },
    });
    return Decoration.set(
      decorations.map((d) => d.value.range(d.from, d.to)),
      true,
    );
  });
}
