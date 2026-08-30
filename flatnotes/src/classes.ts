import { basePaths } from "./constants";
import type { SearchResultModel } from "./types";

export class Note {
  title: string;
  lastModified: number | null;
  content?: string;

  // lastModified is null only for the unsaved "new note" placeholder.
  constructor(title: string = "", lastModified: number | null = null, content?: string) {
    this.title = title;
    this.lastModified = lastModified;
    this.content = content;
  }

  get href(): string {
    return `${basePaths.note}/${this.title.split("/").map(encodeURIComponent).join("/")}`;
  }

  get lastModifiedAsDate(): Date {
    return new Date((this.lastModified ?? 0) * 1000);
  }

  get lastModifiedAsString(): string {
    return this.lastModifiedAsDate.toLocaleString();
  }
}

export class SearchResult extends Note {
  score: number;
  titleHighlights: string | null;
  contentHighlights: string | null;
  tagMatches: string[] | null;

  constructor(searchResult: SearchResultModel) {
    super(searchResult.title, searchResult.lastModified);
    this.score = searchResult.score;
    this.titleHighlights = searchResult.titleHighlights;
    this.contentHighlights = searchResult.contentHighlights;
    this.tagMatches = searchResult.tagMatches;
  }

  get titleHighlightsOrTitle(): string {
    return this.titleHighlights ? this.titleHighlights : this.title;
  }

  get includesHighlights(): boolean {
    return Boolean(
      this.titleHighlights ||
        this.contentHighlights ||
        (this.tagMatches != null && this.tagMatches.length > 0),
    );
  }
}
