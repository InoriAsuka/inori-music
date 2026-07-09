/**
 * Favorites page — /library/favorites
 * GET /api/v1/me/favorites/tracks → FavoritesPage
 * FavoritesPage may include full tracks when catalog service is available.
 */
"use client";

import { useEffect, useState } from "react";
import { Heart, Trash2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { TrackRowSkeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

interface FavoriteTrack {
  id: string;
  title: string;
  durationMs: number;
}

const PAGE_SIZE = 50;

export default function FavoritesPage() {
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);
  const [tracks, setTracks] = useState<FavoriteTrack[]>([]);
  const [trackIds, setTrackIds] = useState<string[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  async function load() {
    if (!token) return;
    setLoading(true);
    const { data } = await authedApi(token).GET("/api/v1/me/favorites/tracks", {
      params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
    });
    if (data) {
      setTrackIds(data.trackIds ?? []);
      setTracks(
        (data.tracks ?? []).map((t) => ({
          id: t.id,
          title: t.title,
          durationMs: t.durationMs ?? 0,
        }))
      );
      if (data.pagination) {
        const p = data.pagination;
        if (p.limit != null && p.offset != null && p.total != null && p.hasMore != null) {
          setPagination({ limit: p.limit, offset: p.offset, total: p.total, hasMore: p.hasMore });
        }
      }
    }
    setLoading(false);
  }

  useEffect(() => {
    load();
  }, [token, page]); // eslint-disable-line react-hooks/exhaustive-deps

  async function removeFavorite(trackId: string) {
    if (!token) return;
    await authedApi(token).DELETE("/api/v1/me/favorites/tracks/{trackId}", {
      params: { path: { trackId } },
    });
    await load();
  }

  function playFrom(idx: number) {
    const q = tracks.map((t) => ({
      id: t.id,
      title: t.title,
      artistName: "",
      albumTitle: "",
      durationSeconds: Math.round(t.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, idx);
  }

  const hasCatalogTracks = tracks.length > 0;
  const empty = !loading && trackIds.length === 0 && tracks.length === 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Heart size={22} className="text-[var(--color-primary)]" />
        <h1 className="text-2xl font-bold">Favorites</h1>
      </div>

      {empty && <EmptyState title="No favorites yet" description="Favorite tracks will appear here." />}

      {!empty && (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {loading
            ? Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                  <TrackRowSkeleton />
                </div>
              ))
            : hasCatalogTracks
              ? tracks.map((t, idx) => (
                  <div
                    key={t.id}
                    className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
                  >
                    <span className="w-6 text-right text-sm text-[var(--color-muted-foreground)]">
                      {offsetFromPage(page, PAGE_SIZE) + idx + 1}
                    </span>
                    <button type="button" onClick={() => playFrom(idx)} className="min-w-0 flex-1 text-left">
                      <p className="truncate text-sm font-medium">{t.title}</p>
                    </button>
                    <span className="text-xs text-[var(--color-muted-foreground)]">
                      {formatDuration(t.durationMs / 1000)}
                    </span>
                    <button
                      type="button"
                      onClick={() => removeFavorite(t.id)}
                      className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]"
                      title="Remove favorite"
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                ))
              : trackIds.map((id, idx) => (
                  <div
                    key={id}
                    className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0"
                  >
                    <span className="w-6 text-right text-sm text-[var(--color-muted-foreground)]">
                      {offsetFromPage(page, PAGE_SIZE) + idx + 1}
                    </span>
                    <code className="min-w-0 flex-1 truncate text-xs">{id}</code>
                    <button
                      type="button"
                      onClick={() => removeFavorite(id)}
                      className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]"
                      title="Remove favorite"
                    >
                      <Trash2 size={15} />
                    </button>
                  </div>
                ))}
        </div>
      )}

      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
