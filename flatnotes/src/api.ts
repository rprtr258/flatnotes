import { basePaths, params } from "./constants";
import { eventBus } from "./eventBus";
import { getToken } from "./tokenStorage";
import type { ApiOptions } from "./types";

export class ApiError extends Error {
  status: number;
  body: unknown;
  handled = false;

  constructor(status: number, body: unknown) {
    super(`API error ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

function redirectUrl(): string {
  return `${basePaths.login}?${params.redirect}=${encodeURIComponent(
    window.location.pathname + window.location.search,
  )}`;
}

export default async function api<T = unknown>(path: string, options: ApiOptions = {}): Promise<T> {
  const { body, params: queryParams, method } = options;

  let url = path;
  if (queryParams) {
    const sp = new URLSearchParams();
    for (const [key, value] of Object.entries(queryParams)) {
      if (value != null) sp.set(key, String(value));
    }
    url += "?" + sp.toString();
  }

  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (path !== "/api/token") {
    headers["Authorization"] = `Bearer ${getToken()}`;
  }

  const response = await fetch(url, {
    method: method || (body ? "POST" : "GET"),
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (response.status === 401 && path !== "/api/token") {
    eventBus.emit("navigate", { href: redirectUrl() });
    const err = new ApiError(401, null);
    err.handled = true;
    throw err;
  }

  const text = await response.text();
  let parsed: unknown = null;
  if (text) {
    try {
      parsed = JSON.parse(text);
    } catch {
      parsed = text;
    }
  }

  if (!response.ok) {
    throw new ApiError(response.status, parsed);
  }

  return parsed as T;
}
