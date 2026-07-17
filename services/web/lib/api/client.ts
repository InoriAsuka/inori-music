/**
 * openapi-fetch client wired to the v1 API.
 *
 * In development, Next.js rewrites /api/v1/* → NEXT_PUBLIC_API_BASE_URL,
 * so we use a relative base URL so the rewrite fires.
 *
 * In production the web container sits behind a reverse proxy that routes
 * /api/v1/* to the API container directly.
 */
import createClient from "openapi-fetch";
import type { paths } from "@/types/api.gen";

// Server-side (RSC / API routes): hit the backend directly.
// Client-side: relative URL so Next.js rewrite fires.
const baseUrl = typeof window === "undefined" ? (process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080") : "";

/**
 * The base URL used for every v1 API call — empty (relative) in the browser so
 * the Next.js rewrite fires, absolute on the server. Exposed for the raw-`fetch`
 * wrappers (player-state / search-history) whose endpoints post-date the
 * generated openapi-fetch `paths` type; they must use the same base as `api`.
 */
export function apiBaseUrl(): string {
  return baseUrl;
}

export const api = createClient<paths>({ baseUrl });

/**
 * Create an authenticated client for viewer (user) requests.
 * Pass the Bearer token obtained from the auth store / cookie.
 */
export function authedApi(token: string) {
  return createClient<paths>({
    baseUrl,
    headers: { Authorization: `Bearer ${token}` },
  });
}

/**
 * Create an authenticated client for admin requests.
 * Uses the static admin token (stored server-side).
 */
export function adminApi(adminToken: string) {
  return createClient<paths>({
    baseUrl,
    headers: { "X-Admin-Token": adminToken },
  });
}
