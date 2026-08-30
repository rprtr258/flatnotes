import type { Extension } from "@codemirror/state";
import { lineWrapperPlugin } from "./lineWrapper";
import { linkPlugin } from "./plugins/link";
import { emphasisPlugin } from "./plugins/emphasis";
import { taskPlugin } from "./plugins/task";
import { tablePlugin } from "./plugins/table";
import { frontmatterPlugin } from "./plugins/frontmatter";

// All live-preview ("clean") editor plugins, in application order.
export function cleanPlugins(): Extension[] {
  return [
    lineWrapperPlugin(),
    linkPlugin(),
    emphasisPlugin(),
    taskPlugin(),
    tablePlugin(),
    frontmatterPlugin(),
  ];
}
