/**
 * User-playlist API helpers — thin typed wrappers over the generated
 * `/api/v1/me/playlists*` operations. Centralised so every caller uses the
 * exact contracts (and the correct verbs) rather than open-coding paths:
 *
 *   GET    /me/playlists                    → list
 *   POST   /me/playlists                    → create        { name, description? }
 *   GET    /me/playlists/{id}               → get
 *   PATCH  /me/playlists/{id}               → rename        { name?, description? }
 *   DELETE /me/playlists/{id}               → delete
 *   GET    /me/playlists/{id}/tracks        → tracks (full CatalogTrack[])
 *   POST   /me/playlists/{id}/tracks        → append        { trackId }  (empty 2xx)
 *   PUT    /me/playlists/{id}/tracks        → replace order { trackIds }  ← reorder / remove-occurrence
 *   DELETE /me/playlists/{id}/tracks/{tid}  → remove FIRST occurrence      (empty 2xx)
 *
 * Note the reorder path is PUT-replace, not a dedicated PATCH-order endpoint —
 * there is no such endpoint in the contract.
 *
 * Body-shape note: the append/remove endpoints answer with an empty 2xx even
 * though the OpenAPI contract types a `UserPlaylist` body. Their wrappers key
 * success off the absence of an error and return void — a `!data` guard would
 * misread openapi-fetch's `data: undefined` (empty body) as a failure.
 */
import { authedApi } from "@/lib/api/client";
import type { components } from "@/types/api.gen";

export type UserPlaylist = components["schemas"]["UserPlaylist"];
export type CatalogTrack = components["schemas"]["CatalogTrack"];

export class UserPlaylistNotFoundError extends Error {
  constructor() {
    super("User playlist not found");
    this.name = "UserPlaylistNotFoundError";
  }
}

export async function listUserPlaylists(token: string): Promise<UserPlaylist[]> {
  const { data, error } = await authedApi(token).GET("/api/v1/me/playlists");
  if (error) throw error;
  return data?.playlists ?? [];
}

export async function getUserPlaylist(token: string, id: string): Promise<UserPlaylist> {
  const { data, error, response } = await authedApi(token).GET("/api/v1/me/playlists/{id}", {
    params: { path: { id } },
  });
  if (response.status === 404) throw new UserPlaylistNotFoundError();
  if (error || !data) throw error ?? new Error("get playlist failed");
  return data;
}

export async function getUserPlaylistTracks(token: string, id: string): Promise<CatalogTrack[]> {
  const { data, error } = await authedApi(token).GET("/api/v1/me/playlists/{id}/tracks", {
    params: { path: { id } },
  });
  if (error) throw error;
  return data?.tracks ?? [];
}

export async function createUserPlaylist(
  token: string,
  body: components["schemas"]["CreateUserPlaylistRequest"]
): Promise<UserPlaylist> {
  const { data, error } = await authedApi(token).POST("/api/v1/me/playlists", { body });
  if (error || !data) throw error ?? new Error("create failed");
  return data;
}

export async function renameUserPlaylist(token: string, id: string, name: string): Promise<UserPlaylist> {
  const { data, error } = await authedApi(token).PATCH("/api/v1/me/playlists/{id}", {
    params: { path: { id } },
    body: { name },
  });
  if (error || !data) throw error ?? new Error("rename failed");
  return data;
}

export async function deleteUserPlaylist(token: string, id: string): Promise<void> {
  const { error } = await authedApi(token).DELETE("/api/v1/me/playlists/{id}", {
    params: { path: { id } },
  });
  if (error) throw error;
}

/**
 * Append a track (duplicates allowed). The OpenAPI contract types a
 * `UserPlaylist` body, but the server answers with an empty 200 — so we key
 * success off the absence of an error, not the presence of a body, and return
 * void. (openapi-fetch yields `data: undefined` for an empty/`Content-Length: 0`
 * 2xx, which a `!data` guard would wrongly treat as a failure.)
 */
export async function appendUserPlaylistTrack(token: string, id: string, trackId: string): Promise<void> {
  const { error } = await authedApi(token).POST("/api/v1/me/playlists/{id}/tracks", {
    params: { path: { id } },
    body: { trackId },
  });
  if (error) throw error;
}

/**
 * Replace the full ordered id list — used for drag-reorder AND for removing a
 * specific (non-first) occurrence, both of which must preserve duplicates.
 */
export async function replaceUserPlaylistTracks(
  token: string,
  id: string,
  trackIds: string[]
): Promise<UserPlaylist> {
  const { data, error } = await authedApi(token).PUT("/api/v1/me/playlists/{id}/tracks", {
    params: { path: { id } },
    body: { trackIds },
  });
  if (error || !data) throw error ?? new Error("reorder failed");
  return data;
}

/**
 * Remove the FIRST occurrence of `trackId` (backend semantics). Like the
 * append endpoint, the server returns an empty 2xx (204) even though the
 * OpenAPI contract types a `UserPlaylist` body — so success is the absence of
 * an error, and a `!data` guard would spuriously throw on every success and
 * trigger an optimistic rollback.
 */
export async function removeFirstUserPlaylistTrack(token: string, id: string, trackId: string): Promise<void> {
  const { error } = await authedApi(token).DELETE("/api/v1/me/playlists/{id}/tracks/{trackId}", {
    params: { path: { id, trackId } },
  });
  if (error) throw error;
}
