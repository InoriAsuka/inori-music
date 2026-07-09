/**
 * Tracks list page — /tracks
 * Resolves artistName from artistId via catalog cache.
 */
"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { TrackRowSkeleton } from "@/components/ui/Skeleton";
import { TrackRow, type TrackRowData } from "@/components/ui/TrackRow";
import { usePlayerStore } from "@/store/player";
import { resolveArtistNames } from "@/lib/api/catalog-cache";

const PAGE_SIZE = 50;

export default function TracksPage() {
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);
  const [tracks, setTracks] = useState<TrackRowData[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    const client = authedApi(token);
    client
      .GET("/api/v1/catalog/tracks", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
      })
      .then(async ({ data }) => {
        if (!data) return;
        const raw = data.tracks ?? [];
        const names = await resolveArtistNames(
          client,
          raw.map((t) => t.artistId)
        );
        setTracks(
          raw.map((t) => ({
            id: t.id,
            title: t.title,
            artistName: names.get(t.artistId) ?? "",
            durationMs: t.durationMs ?? 0,
            isFavorite: t.isFavorite,
          }))
        );
        if (data.pagination) setPagination(data.pagination);
      })
      .finally(() => setLoading(false));
  }, [token, page]);

  function playFrom(idx: number) {
    const q = tracks.map((t) => ({
      id: t.id,
      title: t.title,
      artistName: t.artistName ?? "",
      albumTitle: t.albumTitle ?? "",
      durationSeconds: Math.round(t.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, idx);
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Tracks</h1>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
        <div className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2 text-xs font-medium text-[var(--color-muted-foreground)]">
          <span className="w-6 text-right">#</span>
          <span className="flex-1">Title</span>
          <span className="w-6" />
          <span className="w-12 text-right">Time</span>
        </div>

        {loading
          ? Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                <TrackRowSkeleton />
              </div>
            ))
          : tracks.map((t, idx) => (
              <TrackRow
                key={t.id}
                track={t}
                index={offsetFromPage(page, PAGE_SIZE) + idx + 1}
                onPlay={() => playFrom(idx)}
              />
            ))}
      </div>

      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
