/**
 * Tracks list page — /tracks
 * v1 uses offset/limit, CatalogTrack has durationMs not durationSeconds.
 */
"use client";

import { useEffect, useState } from "react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { TrackRowSkeleton } from "@/components/ui/Skeleton";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

interface Track {
  id: string;
  title: string;
  durationMs: number;
}

const PAGE_SIZE = 50;

export default function TracksPage() {
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) return;
    setLoading(true);
    authedApi(token)
      .GET("/api/v1/catalog/tracks", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) } },
      })
      .then(({ data }) => {
        if (data) {
          setTracks((data.tracks ?? []).map((t) => ({
            id: t.id,
            title: t.title,
            durationMs: t.durationMs ?? 0,
          })));
          if (data.pagination) setPagination(data.pagination);
        }
      })
      .finally(() => setLoading(false));
  }, [token, page]);

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

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Tracks</h1>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
        <div className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2 text-xs font-medium text-[var(--color-muted-foreground)]">
          <span className="w-6 text-right">#</span>
          <span className="flex-1">Title</span>
          <span className="w-12 text-right">Time</span>
        </div>

        {loading
          ? Array.from({ length: 10 }).map((_, i) => (
              <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                <TrackRowSkeleton />
              </div>
            ))
          : tracks.map((t, idx) => (
              <div
                key={t.id}
                onClick={() => playFrom(idx)}
                className="flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
              >
                <span className="w-6 text-right text-sm text-[var(--color-muted-foreground)]">
                  {offsetFromPage(page, PAGE_SIZE) + idx + 1}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="truncate text-sm font-medium">{t.title}</p>
                </div>
                <span className="w-12 text-right text-xs text-[var(--color-muted-foreground)]">
                  {formatDuration(t.durationMs / 1000)}
                </span>
              </div>
            ))}
      </div>

      {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
    </div>
  );
}
