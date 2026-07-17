/**
 * Cross-device player-state API wrappers (v5.4.0).
 *
 *   GET /api/v1/me/player-state  → 200 RemotePlayerState | 404 (never reported)
 *   PUT /api/v1/me/player-state  → 200 RemotePlayerState (server-assigned updatedAt)
 *
 * These endpoints post-date the generated `types/api.gen.ts` (the OpenAPI
 * contract is bumped in a separate server task), so — unlike the openapi-fetch
 * wrappers in this folder — these use raw `fetch` against the documented Go
 * contract (internal/playerstate). The response mirrors `playerstate.PlayerState`
 * exactly; `updatedAt` is the server clock used for last-write-wins conflict
 * resolution. GET returns null on 404 (the user has never reported a state).
 */
import { apiBaseUrl } from "@/lib/api/client";

/** One user's cross-device playback snapshot (server contract: playerstate.PlayerState). */
export interface RemotePlayerState {
  /** Ordered track IDs in the queue. */
  queue: string[];
  /** Index into `queue` of the active track. */
  currentIndex: number;
  /** Playback offset into the current track, in seconds. */
  positionSeconds: number;
  /** Repeat mode. */
  repeat: "off" | "one" | "all";
  /** Shuffle enabled. */
  shuffle: boolean;
  /** Volume in [0, 1]. */
  volume: number;
  /** Playback speed multiplier. */
  speed: number;
  /** Transport state at report time. */
  status: string;
  /** Server-assigned write time (RFC 3339). Absent on locally-built states. */
  updatedAt: string;
}

/** Client-supplied fields; the server assigns `updatedAt`. */
export type PlayerStateUpload = Omit<RemotePlayerState, "updatedAt">;

function authHeaders(token: string, withBody: boolean): HeadersInit {
  const headers: Record<string, string> = { Authorization: `Bearer ${token}` };
  if (withBody) headers["Content-Type"] = "application/json";
  return headers;
}

/**
 * Fetch the remote player state, or null when the user has never reported one
 * (server answers 404 → playerstate.ErrNotFound). Throws on other non-2xx.
 */
export async function getRemotePlayerState(token: string): Promise<RemotePlayerState | null> {
  const res = await fetch(`${apiBaseUrl()}/api/v1/me/player-state`, {
    method: "GET",
    headers: authHeaders(token, false),
  });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`player-state GET failed: ${res.status}`);
  return (await res.json()) as RemotePlayerState;
}

/**
 * Upsert the remote player state (single row, last-write-wins). Returns the
 * stored state including the server-assigned `updatedAt`.
 */
export async function putRemotePlayerState(token: string, state: PlayerStateUpload): Promise<RemotePlayerState> {
  const res = await fetch(`${apiBaseUrl()}/api/v1/me/player-state`, {
    method: "PUT",
    headers: authHeaders(token, true),
    body: JSON.stringify(state),
  });
  if (!res.ok) throw new Error(`player-state PUT failed: ${res.status}`);
  return (await res.json()) as RemotePlayerState;
}
