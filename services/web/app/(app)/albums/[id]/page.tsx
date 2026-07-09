/**
 * Album detail page — /albums/[id]
 * Uses shared TrackRow with isFavorite + artistName via cache.
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
import { TrackRow, type TrackRowData } from "@/components/ui/TrackRow";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";
import { resolveArtistName } from "@/lib/api/catalog-cache";

export default function AlbumDetailPage() {
  const { id } = useParams<{ id: string }>();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [album, setAlbum] = useState<{ title: string; artistName?: string; year?: number; artistId: string } | null>(
    null
  );
  const [tracks, setTracks] = useState<TrackRowData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token || !id) return;
    const client = authedApi(token);
    Promise.all([
      client.GET("/api/v1/catalog/albums/{id}", { params: { path: { id } } }),
      client.GET("/api/v1/catalog/albums/{id}/tracks", { params: { path: { id } } }),
    ]).then(async ([albumRes, tracksRes]) => {
      if (albumRes.data) {
        const artistName = await resolveArtistName(client, albumRes.data.artistId);
        setAlbum({
          title: albumRes.data.title,
          artistId: albumRes.data.artistId,
          artistName,
          year: albumRes.data.releaseYear ?? undefined,
        });
      }
      if (tracksRes.data?.tracks) {
        const raw = tracksRes.data.tracks;
        // All tracks share the album's artist — already resolved above.
        const artistName = album?.artistName ?? "";
        setTracks(
          raw.map((t) => ({
            id: t.id,
            title: t.title,
            artistName,
            durationMs: t.durationMs ?? 0,
            isFavorite: t.isFavorite,
          }))
        );
      }
      setLoading(false);
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, id]);

  // Re-resolve tracks when album artistName becomes available
  useEffect(() => {
    if (!album?.artistName || tracks.length === 0) return;
    setTracks((prev) => prev.map((t) => ({ ...t, artistName: album.artistName })));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [album?.artistName]);

  const totalMs = tracks.reduce((sum, t) => sum + t.durationMs, 0);

  function playAll(startIdx = 0) {
    const q = tracks.map((t) => ({
      id: t.id,
      title: t.title,
      artistName: t.artistName ?? "",
      albumTitle: album?.title ?? "",
      durationSeconds: Math.round(t.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, startIdx);
  }

  return (
    <div className="space-y-8">
      <Link
        href="/albums"
        className="flex items-center gap-1 text-sm text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
      >
        <ArrowLeft size={14} /> Albums
      </Link>

      <div className="flex flex-col gap-6 sm:flex-row sm:items-end">
        <Artwork alt={album?.title ?? ""} size="lg" className="h-48 w-48 shrink-0" />
        <div className="space-y-2">
          {loading ? <Skeleton className="h-8 w-64 mb-2" /> : <h1 className="text-3xl font-bold">{album?.title}</h1>}
          <p className="text-[var(--color-muted-foreground)]">
            {album?.artistName && (
              <Link href={`/artists/${album.artistId}`} className="hover:underline">
                {album.artistName}
              </Link>
            )}
            {album?.year && <span className="ml-1 text-[var(--color-muted-foreground)]">· {album.year}</span>}
          </p>
          <p className="text-sm text-[var(--color-muted-foreground)]">
            {tracks.length} tracks · {formatDuration(totalMs / 1000)}
          </p>
          <button
            type="button"
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
          : tracks.map((t, idx) => <TrackRow key={t.id} track={t} index={idx + 1} onPlay={() => playAll(idx)} />)}
      </div>
    </div>
  );
}
