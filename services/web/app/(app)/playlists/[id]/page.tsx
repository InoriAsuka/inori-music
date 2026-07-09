/**
 * Playlist detail page — /playlists/[id]
 * PlaylistTracksResult: { tracks: CatalogTrack[] }
 * CatalogTrack has no position field — position is inferred from array index.
 */
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Skeleton, TrackRowSkeleton } from "@/components/ui/Skeleton";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

export default function PlaylistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [name, setName] = useState("");
  const [tracks, setTracks] = useState<
    {
      id: string;
      title: string;
      durationMs: number;
    }[]
  >([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token || !id) return;
    const client = authedApi(token);
    Promise.all([
      client.GET("/api/v1/catalog/playlists/{id}", { params: { path: { id } } }),
      client.GET("/api/v1/catalog/playlists/{id}/tracks", { params: { path: { id } } }),
    ]).then(([plRes, tracksRes]) => {
      if (plRes.data) setName(plRes.data.name);
      if (tracksRes.data?.tracks) {
        setTracks(
          tracksRes.data.tracks.map((t) => ({
            id: t.id,
            title: t.title,
            durationMs: t.durationMs ?? 0,
          }))
        );
      }
      setLoading(false);
    });
  }, [token, id]);

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
    <div className="space-y-8">
      <Link
        href="/playlists"
        className="flex items-center gap-1 text-sm text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
      >
        <ArrowLeft size={14} /> Playlists
      </Link>

      <div className="flex items-end gap-4">
        <div>
          {loading ? <Skeleton className="h-8 w-48 mb-2" /> : <h1 className="text-3xl font-bold">{name}</h1>}
          <p className="text-sm text-[var(--color-muted-foreground)]">{tracks.length} tracks</p>
          <button
            type="button"
            onClick={() => playFrom(0)}
            disabled={tracks.length === 0}
            className="mt-3 flex items-center gap-2 rounded-full bg-[var(--color-primary)] px-6 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-50 transition-opacity"
          >
            <Play size={14} fill="currentColor" /> Play
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
        {loading
          ? Array.from({ length: 6 }).map((_, i) => (
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
                <span className="w-6 text-right text-sm text-[var(--color-muted-foreground)]">{idx + 1}</span>
                <div className="flex-1 min-w-0">
                  <p className="truncate text-sm font-medium">{t.title}</p>
                </div>
                <span className="text-xs text-[var(--color-muted-foreground)]">
                  {formatDuration(t.durationMs / 1000)}
                </span>
              </div>
            ))}
      </div>
    </div>
  );
}
