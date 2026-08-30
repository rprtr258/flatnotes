import type { Extension, EditorState } from "@codemirror/state";
import { Decoration, type DecorationSet, WidgetType } from "@codemirror/view";
import { decoratorStateField, isCursorInRange } from "../util";

// @lezer/markdown has no frontmatter grammar node, so the block is detected
// manually: a leading block at offset 0 delimited by "---" lines.

// A muted, monospace block widget rendering the raw frontmatter source.
class FrontmatterWidget extends WidgetType {
  constructor(readonly source: string) {
    super();
  }

  eq(other: FrontmatterWidget): boolean {
    return this.source === other.source;
  }

  toDOM(): HTMLElement {
    const div = document.createElement("div");
    div.className = "sb-frontmatter";
    div.textContent = this.source;
    return div;
  }
}

// Renders a leading YAML frontmatter block as a muted widget when the cursor
// is outside it. Honours the cursor-adjacency rule: raw markdown is shown
// when the cursor is inside the block.
export function frontmatterPlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const doc = state.doc.toString();
    const match = /^---\r?\n[\s\S]*?\r?\n---(?:\r?\n|$)/.exec(doc);
    if (!match) return Decoration.none;
    const from = 0;
    const to = match[0].length;
    if (isCursorInRange(state, from, to)) return Decoration.none;
    return Decoration.set([
      Decoration.replace({
        widget: new FrontmatterWidget(match[0]),
        block: true,
      }).range(from, to),
    ]);
  });
}
