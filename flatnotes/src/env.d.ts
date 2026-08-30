/// <reference types="vite/client" />

declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<Record<string, unknown>, Record<string, unknown>, unknown>;
  export default component;
}

// @toast-ui/editor 3.2.2 ships typings that reference an internal `@t/` path
// alias not resolvable under its package.json "exports" map, so vue-tsc cannot
// resolve them. Ambient declarations with the minimal surface used here:
declare module "@toast-ui/editor" {
  export type EditorType = "markdown" | "wysiwyg";
  export type PreviewStyle = "tab" | "vertical";
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  export type EditorPlugin = (context: any, options?: any) => unknown;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  export type CustomHTMLRenderer = Record<string, (node: any, context: any) => any>;

  export type EditorOptions = {
    el: HTMLElement;
    initialValue?: string;
    initialEditType?: EditorType;
    previewStyle?: PreviewStyle;
    height?: string;
    events?: { change?: (editorType: EditorType) => void };
    plugins?: EditorPlugin[];
    customHTMLRenderer?: CustomHTMLRenderer;
    [key: string]: unknown;
  };

  export default class Editor {
    constructor(options: EditorOptions);
    getMarkdown(): string;
    isWysiwygMode(): boolean;
    destroy(): void;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    [method: string]: any;
  }

  export type ViewerOptions = {
    el: HTMLElement;
    initialValue?: string;
    plugins?: EditorPlugin[];
    customHTMLRenderer?: CustomHTMLRenderer;
    extendedAutolinks?: boolean;
    [key: string]: unknown;
  };

  export class Viewer {
    constructor(options: ViewerOptions);
    setMarkdown(markdown: string): void;
    destroy(): void;
  }
}

declare module "@toast-ui/editor/viewer" {
  import { Viewer } from "@toast-ui/editor";
  export default Viewer;
}

declare module "@toast-ui/editor-plugin-code-syntax-highlight/dist/toastui-editor-plugin-code-syntax-highlight-all.js" {
  import type { EditorPlugin } from "@toast-ui/editor";
  const plugin: EditorPlugin;
  export default plugin;
}
