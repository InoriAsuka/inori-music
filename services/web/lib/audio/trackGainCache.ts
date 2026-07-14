/**
 * Per-track ReplayGain lookup cache — mirrors the mobile client's
 * PlayerNotifier._trackCache pattern (services/mobile/lib/src/player/player_notifier.dart
 * _resolveTrack): cache CatalogTrack fetches by id so repeated plays of the
 * same track (skip back, playlist loop, gapless re-preload) don't refetch.
 *
 * TrackPlaybackDescriptor does not carry replayGainDb, so this is a second,
 * independent fetch against GET /api/v1/catalog/tracks/{id}.
 */
import type { authedApi } from "@/lib/api/client";

type Fetcher = ReturnType<typeof authedApi>;

const gainCache = new Map<string, number | null>(); // trackId -> replayGainDb (null = analyzed-but-absent or fetch failed)

/**
 * Resolves and caches `replayGainDb` for a track. Returns null on fetch
 * failure or when the server has no value — callers should treat null the
 * same as "no ReplayGain data" (computeReplayGain(null) => unity gain).
 */
export async function resolveReplayGainDb(api: Fetcher, trackId: string): Promise<number | null> {
  if (!trackId) return null;
  const cached = gainCache.get(trackId);
  if (cached !== undefined) return cached;
  try {
    const { data } = await api.GET("/api/v1/catalog/tracks/{id}", {
      params: { path: { id: trackId } },
    });
    const db = data?.replayGainDb ?? null;
    gainCache.set(trackId, db);
    return db;
  } catch {
    // Don't cache failures — a transient network error shouldn't permanently
    // deny gain to a track that would resolve fine on retry.
    return null;
  }
}

/** Test-only reset so specs don't leak cache state across cases. */
export function __resetReplayGainCache(): void {
  gainCache.clear();
}
