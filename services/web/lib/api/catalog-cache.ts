/**
 * Lightweight in-memory caches for catalog lookups.
 *
 * The v1 CatalogTrack only carries artistId/albumId, not denormalized names.
 * These helpers maintain a per-session Map so we can resolve names on demand
 * without hammering the API on every render.
 *
 * All functions are fire-and-forget safe — they never throw, returning ""
 * on any error so callers can unconditionally use the result.
 */

import { authedApi } from "@/lib/api/client";

type Fetcher = ReturnType<typeof authedApi>;

// --- Artist cache --------------------------------------------------------

const artistCache = new Map<string, string>(); // id → name

export async function resolveArtistName(api: Fetcher, artistId: string): Promise<string> {
  if (!artistId) return "";
  const cached = artistCache.get(artistId);
  if (cached !== undefined) return cached;
  try {
    const { data } = await api.GET("/api/v1/catalog/artists/{id}", {
      params: { path: { id: artistId } },
    });
    const name = data?.name ?? "";
    artistCache.set(artistId, name);
    return name;
  } catch {
    return "";
  }
}

export async function resolveArtistNames(
  api: Fetcher,
  artistIds: string[]
): Promise<Map<string, string>> {
  const result = new Map<string, string>();
  await Promise.all(
    [...new Set(artistIds)].map(async (id) => {
      result.set(id, await resolveArtistName(api, id));
    })
  );
  return result;
}

// --- Album cache ---------------------------------------------------------

const albumCache = new Map<string, string>(); // id → title

export async function resolveAlbumTitle(api: Fetcher, albumId: string): Promise<string> {
  if (!albumId) return "";
  const cached = albumCache.get(albumId);
  if (cached !== undefined) return cached;
  try {
    const { data } = await api.GET("/api/v1/catalog/albums/{id}", {
      params: { path: { id: albumId } },
    });
    const title = data?.title ?? "";
    albumCache.set(albumId, title);
    return title;
  } catch {
    return "";
  }
}
