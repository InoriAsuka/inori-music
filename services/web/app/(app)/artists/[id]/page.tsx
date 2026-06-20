/**
 * Artist detail page — /artists/[id]
 * CatalogTrack: { id, title, durationMs (not durationSeconds!) }
 */
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Artwork } from "@/components/ui/Artwork";
import { Skeleton, TrackRowSkeleton } from "@/components/ui/Skeleton";
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

export default function ArtistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [name, setName] = useState("");
  const [albums, setAlbums] = useState<{ id: string; title: string; year?: number }[]>([]);
  const [tracks, setTracks] = useState<{ id: string; title: string; durationMs: number }[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token || !id) return;
    const client = authedApi(token);
    Promise.all([
      client.GET("/api/v1/catalog/artists/{id}", { params: { path: { id } } }),
      client.GET("/api/v1/catalog/artists/{id}/albums", { params: { path: { id } } }),
      client.GET("/api/v1/catalog/artists/{id}/tracks", { params: { path: { id } } }),
    ]).then(([artistRes, albumsRes, tracksRes]) => {
      if (artistRes.data) setName(artistRes.data.name);
      if (albumsRes.data?.albums) {
        setAlbums(albumsRes.data.albums.map((a) => ({
          id: a.id,
          title: a.title,
          year: a.releaseYear ?? undefined,
        })));
      }
      if (tracksRes.data?.tracks) {
        setTracks(tracksRes.data.tracks.map((t) => ({
          id: t.id,
          title: t.title,
          durationMs: t.durationMs ?? 0,
        })));
      }
      setLoading(false);
    });
  }, [token, id]);

  function playFromIndex(idx: number) {
    const q = tracks.map((t) => ({
      id: t.id,
      title: t.title,
      artistName: name,
      albumTitle: "",
      durationSeconds: Math.round(t.durationMs / 1000),
      playbackUrl: "",
    }));
    playQueue(q, idx);
  }

  return (
    <div className="space-y-8">
      <Link href="/artists" className="flex items-center gap-1 text-sm text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]">
        <ArrowLeft size={14} /> Artists
      </Link>

      <div className="flex items-end gap-6">
        <div className="flex h-32 w-32 items-center justify-center rounded-xl bg-[var(--color-muted)] text-5xl font-bold text-[var(--color-muted-foreground)]">
          {loading ? "" : name.charAt(0).toUpperCase()}
        </div>
        <div>
          {loading ? <Skeleton className="h-8 w-48 mb-2" /> : <h1 className="text-3xl font-bold">{name}</h1>}
          <p className="text-sm text-[var(--color-muted-foreground)]">
            {albums.length} albums · {tracks.length} tracks
          </p>
        </div>
      </div>

      {albums.length > 0 && (
        <section>
          <h2 className="mb-3 text-lg font-semibold">Albums</h2>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
            {albums.map((album) => (
              <Link key={album.id} href={`/albums/${album.id}`} className="group flex flex-col gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] p-3 hover:bg-[var(--color-muted)] transition-colors">
                <Artwork alt={album.title} size="lg" className="w-full h-auto aspect-square" />
                <p className="truncate font-medium text-sm">{album.title}</p>
                {album.year && <p className="text-xs text-[var(--color-muted-foreground)]">{album.year}</p>}
              </Link>
            ))}
          </div>
        </section>
      )}

      <section>
        <h2 className="mb-3 text-lg font-semibold">All Tracks</h2>
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {loading
            ? Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                  <TrackRowSkeleton />
                </div>
              ))
            : tracks.map((t, idx) => (
                <div
                  key={t.id}
                  onClick={() => playFromIndex(idx)}
                  className="flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
                >
                  <span className="w-5 text-center text-sm text-[var(--color-muted-foreground)]">{idx + 1}</span>
                  <div className="flex-1 min-w-0">
                    <p className="truncate text-sm font-medium">{t.title}</p>
                  </div>
                  <span className="text-xs text-[var(--color-muted-foreground)]">
                    {formatDuration(t.durationMs / 1000)}
                  </span>
                </div>
              ))}
        </div>
      </section>
    </div>
  );
}
