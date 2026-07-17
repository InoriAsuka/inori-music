/**
 * Cross-device search-history API wrappers (v5.4.0).
 *
 *   GET    /api/v1/me/search-history  → 200 { queries: string[] } (≤20, newest first)
 *   PUT    /api/v1/me/search-history  → 200 (replace-all; server dedups + caps at 20)
 *   DELETE /api/v1/me/search-history  → 204 (clear)
 *
 * Like player-state, these endpoints post-date the generated `types/api.gen.ts`
 * so they use raw `fetch` against the documented Go contract
 * (internal/searchhistory). The server owns dedup/cap/timestamp semantics; the
 * client sends its merged newest-first list and reads back the canonical order.
 */
import { apiBaseUrl } from "@/lib/api/client";

interface SearchHistoryResponse {
  queries?: string[];
}

function authHeaders(token: string, withBody: boolean): HeadersInit {
  const headers: Record<string, string> = { Authorization: `Bearer ${token}` };
  if (withBody) headers["Content-Type"] = "application/json";
  return headers;
}

/** Fetch the user's remote search history (newest first, ≤20). Throws on non-2xx. */
export async function getRemoteSearchHistory(token: string): Promise<string[]> {
  const res = await fetch(`${apiBaseUrl()}/api/v1/me/search-history`, {
    method: "GET",
    headers: authHeaders(token, false),
  });
  if (!res.ok) throw new Error(`search-history GET failed: ${res.status}`);
  const body = (await res.json()) as SearchHistoryResponse;
  return body.queries ?? [];
}

/**
 * Replace the user's remote search history with `queries` (newest first). The
 * server re-dedups, drops blanks, and caps at 20 — the client sends its already
 * merged list. Resolves on any 2xx (the endpoint returns no meaningful body).
 */
export async function putRemoteSearchHistory(token: string, queries: string[]): Promise<void> {
  const res = await fetch(`${apiBaseUrl()}/api/v1/me/search-history`, {
    method: "PUT",
    headers: authHeaders(token, true),
    body: JSON.stringify({ queries }),
  });
  if (!res.ok) throw new Error(`search-history PUT failed: ${res.status}`);
}

/** Clear the user's remote search history. Resolves on any 2xx (204 expected). */
export async function deleteRemoteSearchHistory(token: string): Promise<void> {
  const res = await fetch(`${apiBaseUrl()}/api/v1/me/search-history`, {
    method: "DELETE",
    headers: authHeaders(token, false),
  });
  if (!res.ok) throw new Error(`search-history DELETE failed: ${res.status}`);
}
