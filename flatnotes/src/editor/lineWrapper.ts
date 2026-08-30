import type { Extension, EditorState } from "@codemirror/state";
import { Decoration, type DecorationSet } from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import type { SyntaxNodeRef } from "@lezer/common";
import { decoratorStateField } from "./util";

// Maps lezer markdown block node names to line CSS classes.
const CLASS_MAP: Record<string, string> = {
  ATXHeading1: "sb-line-h1",
  ATXHeading2: "sb-line-h2",
  ATXHeading3: "sb-line-h3",
  ATXHeading4: "sb-line-h4",
  ATXHeading5: "sb-line-h5",
  ATXHeading6: "sb-line-h6",
  SetextHeading1: "sb-line-h1",
  SetextHeading2: "sb-line-h2",
  ListItem: "sb-line-li",
  Blockquote: "sb-line-blockquote",
  FencedCode: "sb-line-code",
  HorizontalRule: "sb-line-hr",
};

// Returns a line decoration for a block node, or undefined when the node has
// no associated line class. Intended for use inside syntaxTree(state).iterate.
export function lineWrapper(
  classes: Record<string, string>,
): (node: SyntaxNodeRef) => void | { from: number; to: number; decoration: Decoration } {
  return (node) => {
    const cls = classes[node.name];
    if (!cls) return;
    return {
      from: node.from,
      to: node.from,
      decoration: Decoration.line({ class: cls }),
    };
  };
}

// Plugin that wraps every recognised block line with its CSS class.
export function lineWrapperPlugin(): Extension {
  return decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; value: Decoration }[] = [];
    const wrap = lineWrapper(CLASS_MAP);
    syntaxTree(state).iterate({
      enter(node) {
        const res = wrap(node);
        if (res) decorations.push({ from: res.from, value: res.decoration });
      },
    });
    return Decoration.set(
      decorations.map((d) => d.value.range(d.from)),
      true,
    );
  });
}
