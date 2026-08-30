export interface TokenResponse {
  access_token: string;
  token_type: string;
}

export interface NoteResponse {
  title: string;
  lastModified: number;
}

export interface NoteContentResponseModel extends NoteResponse {
  content: string;
}

export interface SearchResultModel {
  score: number;
  title: string;
  lastModified: number;
  titleHighlights: string | null;
  contentHighlights: string | null;
  tagMatches: string[] | null;
}

export interface ConfigResponse {
  authType: string;
}

export interface ApiOptions {
  method?: string;
  body?: unknown;
  params?: Record<string, string | number | boolean | null | undefined>;
}
