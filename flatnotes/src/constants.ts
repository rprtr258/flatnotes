export const basePaths = {
  home:   "/",
  login:  "/login",
  note:   "/note",
  search: "/search",
  new:    "/new",
} as const;

export const params = {
  searchTerm:     "term",
  redirect:       "redirect",
  showHighlights: "showHighlights",
  sortBy:         "sortBy",
} as const;

export const alphabet = [
  "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M",
  "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
];

export const searchSortOptions = {
  score:        0,
  title:        1,
  lastModified: 2,
} as const;

export const authTypes = {
  none:     "none",
  readOnly: "read_only",
  password: "password",
  totp:     "totp",
} as const;

export type AuthType = (typeof authTypes)[keyof typeof authTypes];
export type SearchSortOption = (typeof searchSortOptions)[keyof typeof searchSortOptions];
