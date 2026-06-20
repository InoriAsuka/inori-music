/**
 * Album detail page — /albums/[id]
 */
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Artwork } from "@/components/ui/Artwork";
import { Skeleton, TrackRowSkeleton } from "@/components/ui/Skeleton";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

export default function AlbumDetailPage() {
  const { id } = useParams<{ id: string }>();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [album, setAlbum] = useState<{ title: string; year?: number } | null>(null);
  const [tracks, setTracks] = useState<{
    id: string;
    title: string;
    trackNumber?: number;
    discNumber?: number;
    durationMs: number;
  }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token || !id) return;
    const client = authedApi(token);
    Promise.all([
      client.GET("/api/v1/catalog/albums/{id}", { params: { path: { id } } }),
      client.GET("/api/v1/catalog/albums/{id}/tracks", { params: { path: { id } } }),
    ]).then(([albumRes, tracksRes]) => {
      if (albumRes.data) {
        setAlbum({ title: albumRes.data.title, year: albumRes.data.releaseYear ?? undefined });
      }
      if (tracksRes.data?.tracks) {
        setTracks(tracksRes.data.tracks.map((t) => ({
          id: t.id,
          title: t.title,
          trackNumber: t.trackNumber ?? undefined,
          discNumber: t.discNumber ?? undefined,
          durationMs: t.durationMs ?? 0,
        })));
      }
      setLoading(false);
    });
  }, [token, id]);

  const totalMs = tracks.reduce((sum, t) => sum + t.durationMs, 0);

  function playAll(startIdx = 0) {
    const q = tracks.map((t) => ({
      id: t.id,
      title: t.title,
      artistName: "",
      albumTitle: album?.title ?? "",
      durationSeconds: Math.round(t.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, startIdx);
  }

  return (
    <div className="space-y-8">
      <Link href="/albums" className="flex items-center gap-1 text-sm text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]">
        <ArrowLeft size={14} /> Albums
      </Link>

      <div className="flex flex-col gap-6 sm:flex-row sm:items-end">
        <Artwork alt={album?.title ?? ""} size="lg" className="h-48 w-48 shrink-0" />
        <div className="space-y-2">
          {loading
            ? <Skeleton className="h-8 w-64 mb-2" />
            : <h1 className="text-3xl font-bold">{album?.title}</h1>}
          {album?.year && (
            <p className="text-[var(--color-muted-foreground)]">{album.year}</p>
          )}
          <p className="text-sm text-[var(--color-muted-foreground)]">
            {tracks.length} tracks · {formatDuration(totalMs / 1000)}
          </p>
          <button
            onClick={() => playAll(0)}
            disabled={tracks.length === 0}
            className="mt-2 flex items-center gap-2 rounded-full bg-[var(--color-primary)] px-6 py-2.5 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-50 transition-opacity"
          >
            <Play size={14} fill="currentColor" /> Play
          </button>
        </div>
      </div>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
        {loading
          ? Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                <TrackRowSkeleton />
              </div>
            ))
          : tracks.map((t, idx) => (
              <div
                key={t.id}
                onClick={() => playAll(idx)}
                className="flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
              >
                <span className="w-6 text-right text-sm text-[var(--color-muted-foreground)]">
                  {t.trackNumber ?? idx + 1}
                </span>
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
