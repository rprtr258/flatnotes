import type { Extension, EditorState } from "@codemirror/state";
import { Decoration, type DecorationSet, WidgetType } from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import { decoratorStateField, isCursorInRange } from "../util";

// Renders a whole GFM `Table` block as an HTML table widget when the cursor
// is outside the block. Line 1 is the header, line 2 the delimiter (skipped),
// remaining lines are body rows. Cells split on `|`, trimmed, with leading
// and trailing empty pipes dropped.
class TableWidget extends WidgetType {
  constructor(readonly source: string) {
    super();
  }

  eq(other: TableWidget): boolean {
    return this.source === other.source;
  }

  toDOM(): HTMLElement {
    const lines = this.source.split("\n");
    const parseRow = (line: string): string[] => {
      const cells = line.split("|").map((c) => c.trim());
      if (cells.length > 0 && cells[0] === "") cells.shift();
      if (cells.length > 0 && cells[cells.length - 1] === "") cells.pop();
      return cells;
    };

    const wrap = document.createElement("div");
    wrap.className = "sb-table";
    const table = document.createElement("table");

    const thead = document.createElement("thead");
    const headerRow = document.createElement("tr");
    if (lines.length > 0) {
      for (const cell of parseRow(lines[0])) {
        const th = document.createElement("th");
        th.textContent = cell;
        headerRow.appendChild(th);
      }
    }
    thead.appendChild(headerRow);
    table.appendChild(thead);

    const tbody = document.createElement("tbody");
    for (const line of lines.slice(2)) {
      const tr = document.createElement("tr");
      for (const cell of parseRow(line)) {
        const td = document.createElement("td");
        td.textContent = cell;
        tr.appendChild(td);
      }
      tbody.appendChild(tr);
    }
    table.appendChild(tbody);
    wrap.appendChild(table);
    return wrap;
  }
}

export function tablePlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; to: number; value: Decoration }[] = [];
    syntaxTree(state).iterate({
      enter(node) {
        if (node.name !== "Table") return;
        if (isCursorInRange(state, node.from, node.to)) return;
        decorations.push({
          from: node.from,
          to: node.to,
          value: Decoration.replace({
            widget: new TableWidget(state.sliceDoc(node.from, node.to)),
            block: true,
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
