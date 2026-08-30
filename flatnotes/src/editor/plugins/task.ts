import type { Extension, EditorState } from "@codemirror/state";
import {
  Decoration,
  type DecorationSet,
  EditorView,
  WidgetType,
} from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import { decoratorStateField, isCursorInRange } from "../util";

// Replaces a GFM task marker (`[ ]` / `[x]`) with a clickable checkbox widget.
// The marker is the `TaskMarker` lezer node (3 chars at the start of a Task).
// While the cursor is inside the marker the raw text is shown for editing.
class TaskCheckboxWidget extends WidgetType {
  constructor(readonly markerFrom: number, readonly checked: boolean) {
    super();
  }

  eq(other: TaskCheckboxWidget): boolean {
    return this.markerFrom === other.markerFrom && this.checked === other.checked;
  }

  toDOM(view: EditorView): HTMLElement {
    const input = document.createElement("input");
    input.type = "checkbox";
    input.checked = this.checked;
    input.className = "sb-task";
    // Toggle the source marker on click. `this.checked` is the pre-click
    // state, so we insert the opposite marker.
    input.addEventListener("click", (event) => {
      event.preventDefault();
      view.dispatch({
        changes: {
          from: this.markerFrom,
          to: this.markerFrom + 3,
          insert: this.checked ? "[ ]" : "[x]",
        },
      });
    });
    return input;
  }
}

export function taskPlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; to: number; value: Decoration }[] = [];
    syntaxTree(state).iterate({
      enter(node) {
        if (node.name !== "TaskMarker") return;
        if (isCursorInRange(state, node.from, node.to)) return;
        const marker = state.sliceDoc(node.from, node.to);
        const checked = /[xX]/.test(marker);
        decorations.push({
          from: node.from,
          to: node.to,
          value: Decoration.replace({
            widget: new TaskCheckboxWidget(node.from, checked),
            block: false,
          }),
        });
      },
    });
    return Decoration.set(
      decorations.map((d) => d.value.range(d.from, d.to)),
      true,
    );
  });
}
