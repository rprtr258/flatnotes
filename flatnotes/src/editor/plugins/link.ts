import { Facet, type Extension, type EditorState } from "@codemirror/state";
import {
  Decoration,
  EditorView,
  WidgetType,
  type DecorationSet,
} from "@codemirror/view";
import { syntaxTree } from "@codemirror/language";
import type { SyntaxNode, SyntaxNodeRef } from "@lezer/common";
import { eventBus } from "../../eventBus";
import { decoratorStateField, isCursorInRange } from "../util";

// Lezer child node names (verified in node_modules/@lezer/markdown):
//   Link, Image  -> container nodes
//   LinkMark     -> the "[", "]", "(", ")", "!" delimiter marks
//   URL          -> the destination URL
//
// Two image shapes are handled:
//   ![alt](url)      standard markdown image (URL child present)
//   ![[filename]]    wikilink embed (inner Link child, no URL)
//
// Images always render. When the cursor is inside the node the raw markdown is
// left in place for editing and the image is additionally shown as a block
// widget on its own line above it.

// The title of the note currently being edited (slash-separated, e.g.
// "sub/note"). Used to resolve relative image paths against the note's
// directory via the authenticated /attachments route.
export const noteTitleFacet = Facet.define<string, string>({
  combine: (values) => (values.length ? values[values.length - 1] : ""),
});

// Builds a URL the backend can serve for an image path, resolving it
// relative to the current note's file path using strict relative-URL
// resolution: the note's filename segment is dropped (so a bare `img.png`
// points at the note's sibling directory), and `.` / `..` segments pop the
// directory stack — `../img.png` reaches the note's parent directory.
// Absolute URLs (schemes, site-absolute paths) are returned unchanged.
function resolveImageSrc(src: string, noteTitle: string): string {
  if (/^[a-z][a-z0-9+.-]*:/i.test(src)) return src; // http(s):, data:, etc.
  if (src.startsWith("/")) return src; // site-absolute path
  // Base is the note's file path; resolution drops the filename segment.
  // (The `.md` extension is immaterial — the last segment is dropped either
  // way — so noteTitle, which is the path sans extension, is the base.)
  const segments = noteTitle.includes("/")
    ? noteTitle.slice(0, noteTitle.lastIndexOf("/")).split("/")
    : [];
  for (const seg of src.split("/")) {
    if (seg === "" || seg === ".") continue;
    if (seg === "..") {
      segments.pop();
      continue;
    }
    segments.push(seg);
  }
  const full = segments.join("/");
  return "/attachments/" + full.split("/").map(encodeURIComponent).join("/");
}

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

// Renders an inline <img> for an Image node.
class ImageWidget extends WidgetType {
  constructor(readonly src: string, readonly alt: string) {
    super();
  }

  eq(other: ImageWidget): boolean {
    return this.src === other.src && this.alt === other.alt;
  }

  toDOM(): HTMLElement {
    const img = document.createElement("img");
    img.src = this.src;
    img.alt = this.alt;
    img.className = "sb-image";
    img.loading = "lazy";
    return img;
  }
}

// Resolves an Image node into {src, alt}. Returns null for unrecognised
// shapes (autolinks, malformed), which the caller leaves raw.
function resolveImage(
  state: EditorState,
  ref: SyntaxNodeRef,
): { src: string; alt: string } | null {
  let url: string | null = null;
  let linkChild: SyntaxNodeRef | null = null;
  let openEnd: number | null = null; // end of the leading "![" mark
  let altClose: number | null = null; // start of the "]" closing the alt text
  for (let child = ref.node.firstChild; child; child = child.nextSibling) {
    if (child.name === "URL") {
      url = state.sliceDoc(child.from, child.to);
    } else if (child.name === "Link") {
      linkChild = child;
    } else if (child.name === "LinkMark") {
      const text = state.sliceDoc(child.from, child.to);
      if (text.startsWith("![")) {
        openEnd = child.to;
      } else if (text === "]" && openEnd !== null && altClose === null) {
        altClose = child.from;
      }
    }
  }
  if (url !== null) {
    // ![alt](url)
    const alt =
      openEnd !== null && altClose !== null
        ? state.sliceDoc(openEnd, altClose)
        : "";
    return { src: url, alt };
  }
  if (linkChild !== null) {
    // ![[filename]] — the inner Link's text is the filename
    const inner = linkTextRange(state, linkChild);
    if (!inner) return null;
    const filename = state.sliceDoc(inner.textFrom, inner.textTo);
    return { src: filename, alt: filename };
  }
  return null;
}

// Builds the SPA note URL for a note path, matching Note.href and
// TreeExplorer.noteHref so wikilinks route through the app's /note handler.
function noteHref(notePath: string): string {
  return "/note/" + notePath.split("/").map(encodeURIComponent).join("/");
}

// Resolves a Link node to a navigation href. A standard `[text](url)` link
// yields its URL; a URL-less link (e.g. a `[[note path]]` wikilink, which the
// parser exposes as a bracketed Link with no URL child) yields the note URL
// built from the link text. Returns null for unrecognised shapes.
function linkHref(state: EditorState, node: SyntaxNode): string | null {
  for (let child = node.firstChild; child; child = child.nextSibling) {
    if (child.name === "URL") return state.sliceDoc(child.from, child.to);
  }
  const range = linkTextRange(state, node);
  if (!range) return null;
  return noteHref(state.sliceDoc(range.textFrom, range.textTo));
}

// Ensures a bare URL is openable: leaves scheme-bearing URLs (http,
// mailto, data, ...) and protocol-relative `//host` URLs alone, and
// prefixes `https://` for schemeless hosts (`www.example.com`).
function normalizeUrl(url: string): string {
  if (/^[a-z][a-z0-9+.-]*:/i.test(url)) return url;
  if (url.startsWith("//")) return "https:" + url;
  return "https://" + url;
}

// Navigates to a link href. Note routes go through the SPA navigate event
// (so ctrl/cmd-click still opens a new tab via the app's handler); everything
// else — external URLs, static files, attachments — opens in a new tab.
function openLink(href: string, event?: MouseEvent): void {
  if (href.startsWith("/note/")) {
    eventBus.emit("navigate", { href, event });
  } else {
    window.open(href, "_blank", "noopener,noreferrer");
  }
}

// Resolves a clicked node to a navigation href. Handles bracketed Links
// (`[text](url)` / `[[note]]`) and bare/autolink URLs (a `URL` node, whether
// standalone or inside an `Autolink` — the URL text is the address).
function resolveClickHref(state: EditorState, node: SyntaxNode): string | null {
  if (node.name === "Link") return linkHref(state, node);
  if (node.name === "URL") return normalizeUrl(state.sliceDoc(node.from, node.to));
  return null;
}

// Hides the surrounding markup of inline links, renders images inline, and
// navigates (instead of placing the cursor) when a rendered link is clicked.
export function linkPlugin(): Extension {
  const decorationsField = decoratorStateField((state: EditorState): DecorationSet => {
    const decorations: { from: number; to: number; value: Decoration }[] = [];
    const noteTitle = state.facet(noteTitleFacet);
    syntaxTree(state).iterate({
      enter(node) {
        if (node.name === "Image") {
          const img = resolveImage(state, node);
          if (!img) return;
          const src = resolveImageSrc(img.src, noteTitle);
          if (isCursorInRange(state, node.from, node.to)) {
            // Cursor inside: keep the raw markdown editable and show the
            // image as a block widget on its own line above it.
            decorations.push({
              from: node.from,
              to: node.from,
              value: Decoration.widget({
                widget: new ImageWidget(src, img.alt),
                block: true,
              }),
            });
          } else {
            decorations.push({
              from: node.from,
              to: node.to,
              value: Decoration.replace({
                widget: new ImageWidget(src, img.alt),
                block: false,
              }),
            });
          }
          return;
        }
        if (node.name === "URL") {
          // Bare/autolink URLs (standalone `https://x.com` or the inner
          // URL of `<https://x.com>`). URLs that belong to a Link/Image are
          // handled (or hidden) by their parent branch.
          const parent = node.node.parent;
          if (parent && (parent.name === "Link" || parent.name === "Image")) {
            return;
          }
          if (isCursorInRange(state, node.from, node.to)) return;
          decorations.push({
            from: node.from,
            to: node.to,
            value: Decoration.mark({ class: "sb-link" }),
          });
          return;
        }
        if (node.name !== "Link") return;
        if (isCursorInRange(state, node.from, node.to)) return;
        const range = linkTextRange(state, node);
        if (!range) return; // autolink or unrecognised shape: leave raw
        // Hide the leading markup ("[") and trailing markup ("](url)").
        decorations.push({
          from: node.from,
          to: range.textFrom,
          value: Decoration.replace({}),
        });
        decorations.push({
          from: range.textFrom,
          to: range.textTo,
          value: Decoration.mark({ class: "sb-link" }),
        });
        decorations.push({
          from: range.textTo,
          to: node.to,
          value: Decoration.replace({}),
        });
      },
    });
    return Decoration.set(
      decorations.map((d) => d.value.range(d.from, d.to)),
      true,
    );
  });
  // Clicking a rendered link navigates to its address instead of moving the
  // cursor. Handled on mousedown (CodeMirror places the cursor on mousedown,
  // so returning true here is what actually suppresses that).
  const clickHandler = EditorView.domEventHandlers({
    mousedown(event: MouseEvent, view: EditorView): boolean {
      if (event.button !== 0) return false;
      const domTarget = event.target;
      if (!(domTarget instanceof Element) || !domTarget.closest(".sb-link")) {
        return false;
      }
      const pos = view.posAtCoords({ x: event.clientX, y: event.clientY });
      if (pos == null) return false;
      let target: SyntaxNode | null = null;
      for (
        let n: SyntaxNode | null = syntaxTree(view.state).resolve(pos, 1);
        n;
        n = n.parent
      ) {
        if (n.name === "Link" || n.name === "URL") {
          target = n;
          break;
        }
      }
      if (!target) return false;
      const href = resolveClickHref(view.state, target);
      if (!href) return false;
      openLink(href, event);
      return true;
    },
  });
  return [decorationsField, clickHandler];
}
