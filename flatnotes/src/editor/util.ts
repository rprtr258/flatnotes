import {
  StateField,
  type Extension,
  type EditorState,
} from "@codemirror/state";
import { EditorView, Decoration, type DecorationSet } from "@codemirror/view";
import type { SyntaxNodeRef } from "@lezer/common";

// A fully invisible replacement decoration. Used to hide raw markdown markup
// that has been rendered into a widget/line decoration.
export const invisibleDecoration = Decoration.replace({});

// A StateField that recomputes its DecorationSet whenever the document or
// selection changes, and provides those decorations to the view.
export function decoratorStateField(
  create: (state: EditorState) => DecorationSet,
): Extension {
  return StateField.define<DecorationSet>({
    create,
    update(value, tr) {
      if (tr.docChanged || tr.selection) return create(tr.state);
      return value;
    },
    provide: (f) => EditorView.decorations.from(f),
  });
}

// Returns true when the editor cursor (any selection range) overlaps the
// given [from, to) range. Plugins use this to keep raw markup visible while
// the user is editing inside a node.
export function isCursorInRange(
  state: EditorState,
  from: number,
  to: number,
): boolean {
  return state.selection.ranges.some((r) => r.from <= to && r.to >= from);
}

// For a block-level node, returns the range of leading source markup that can
// be hidden once the node has been decorated, or null when nothing should be
// hidden. Honours the cursor-adjacency rule by refusing to hide when the
// cursor sits in the leading markup range.
export function hideBlockSource(
  node: SyntaxNodeRef,
  state: EditorState,
): { from: number; to: number } | null {
  const full = node.node;
  if (!full) return null;
  const child = full.firstChild;
  if (!child || child.from <= full.from) return null;
  if (isCursorInRange(state, full.from, child.from)) return null;
  return { from: full.from, to: child.from };
}
