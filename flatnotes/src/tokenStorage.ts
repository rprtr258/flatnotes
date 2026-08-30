const tokenStorageKey = "token";

function getCookieString(token?: string): string {
  return `${tokenStorageKey}=${token ?? ""}; path=/attachments; SameSite=Strict`;
}

export function setToken(token: string, persist = false): void {
  document.cookie = getCookieString(token);
  sessionStorage.setItem(tokenStorageKey, token);
  if (persist === true) {
    localStorage.setItem(tokenStorageKey, token);
  }
}

export function getToken(): string | null {
  return sessionStorage.getItem(tokenStorageKey);
}

export function loadToken(): void {
  const token = localStorage.getItem(tokenStorageKey);
  if (token != null) {
    setToken(token, false);
  }
}

export function clearToken(): void {
  sessionStorage.removeItem(tokenStorageKey);
  localStorage.removeItem(tokenStorageKey);
  document.cookie = getCookieString() + "; expires=Thu, 01 Jan 1970 00:00:00 GMT";
}
