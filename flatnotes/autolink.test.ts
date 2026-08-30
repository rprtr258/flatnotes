import {markdown} from "@codemirror/lang-markdown";
import {EditorState} from "@codemirror/state";
import {syntaxTree} from "@codemirror/language";
import {GFM} from "@lezer/markdown";

const cases = [
  "bare https://example.com here",
  "<https://example.com>",
  "see http://x.com/y yes",
  "www.example.com path",
];
for (const doc of cases) {
  const md = markdown({ addKeymap: true, extensions: [GFM] });
  const s = EditorState.create({ doc, extensions: [md] });
  console.log("===", JSON.stringify(doc));
  syntaxTree(s).iterate({
    enter(n) {
      if (["Link","URL","LinkMark","Autolink"].includes(n.name))
        console.log(n.name, JSON.stringify(s.sliceDoc(n.from, n.to)));
    },
  });
}
