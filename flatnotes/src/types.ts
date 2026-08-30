export type TokenResponse = {
  access_token: string,
  token_type: string,
};

export type NoteResponse = {
  title: string,
  lastModified: number,
};

export type NoteContentResponseModel = NoteResponse & {
  content: string,
};

export type SearchResultModel = {
  score: number,
  title: string,
  lastModified: number,
  titleHighlights: string | null,
  contentHighlights: string | null,
  tagMatches: string[] | null,
};

export type ConfigResponse = {
  authType: string,
};

export type TodoItem = {
  text: string,
  done: boolean,
};

export type NoteTodos = {
  title: string,
  lastModified: number,
  todos: TodoItem[],
};

export type ApiOptions = {
  method?: string,
  body?: unknown,
  params?: Record<string, string | number | boolean | null | undefined>,
};
