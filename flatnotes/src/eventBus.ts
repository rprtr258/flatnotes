import mitt from "mitt";

export type NavigatePayload = { href: string; event?: Event };

export type AppEvents = {
  navigate: NavigatePayload;
  "unhandled-server-error": { error: unknown };
  "update-note-title": { title: string };
  "highlight-search-input": void;
};

export const eventBus = mitt<AppEvents>();
