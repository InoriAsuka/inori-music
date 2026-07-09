/**
 * Track detail page — /tracks/[id]
 *
 * Displays: title, artist, album, duration, genre, track/disc number,
 * isFavorite toggle, play-count stats (via /me/history/tracks/{id}/stats),
 * and a Play button.
 */
"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Play, Heart, Music2, Disc3, Clock, Hash, Tag, BarChart2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { usePlayerStore } from "@/store/player";
import { Artwork } from "@/components/ui/Artwork";
import { Skeleton } from "@/components/ui/Skeleton";
import { formatDuration, cn } from "@/lib/utils";
import { resolveArtistName, resolveAlbumTitle } from "@/lib/api/catalog-cache";

interface TrackDetail {
  id: string;
  title: string;
  artistId: string;
  artistName: string;
  albumId: string;
  albumTitle: string;
  durationMs: number;
  genre?: string;
  trackNumber?: number;
  discNumber?: number;
  isFavorite: boolean;
}

interface TrackStats {
  totalPlays: number;
  firstPlayedAt?: string;
  lastPlayedAt?: string;
}

export default function TrackDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);

  const [track, setTrack] = useState<TrackDetail | null>(null);
  const [stats, setStats] = useState<TrackStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [fav, setFav] = useState(false);
  const [favLoading, setFavLoading] = useState(false);

  useEffect(() => {
    if (!token || !id) return;
    const client = authedApi(token);
    Promise.all([
      client.GET("/api/v1/catalog/tracks/{id}", { params: { path: { id } } }),
      client.GET("/api/v1/me/history/tracks/{trackId}/stats", {
        params: { path: { trackId: id } },
      }),
    ]).then(async ([trackRes, statsRes]) => {
      if (trackRes.data) {
        const t = trackRes.data;
        const [artistName, albumTitle] = await Promise.all([
          resolveArtistName(client, t.artistId ?? ""),
          resolveAlbumTitle(client, t.albumId ?? ""),
        ]);
        const detail: TrackDetail = {
          id: t.id,
          title: t.title,
          artistId: t.artistId ?? "",
          artistName,
          albumId: t.albumId ?? "",
          albumTitle,
          durationMs: t.durationMs ?? 0,
          genre: t.genre ?? undefined,
          trackNumber: t.trackNumber ?? undefined,
          discNumber: t.discNumber ?? undefined,
          isFavorite: t.isFavorite ?? false,
        };
        setTrack(detail);
        setFav(detail.isFavorite);
      }
      if (statsRes.data) {
        setStats({
          totalPlays: statsRes.data.totalPlays,
          firstPlayedAt: statsRes.data.firstPlayedAt ?? undefined,
          lastPlayedAt: statsRes.data.lastPlayedAt ?? undefined,
        });
      }
      setLoading(false);
    });
  }, [token, id]);

  async function toggleFavorite() {
    if (!token || !track || favLoading) return;
    setFavLoading(true);
    const client = authedApi(token);
    try {
      if (fav) {
        await client.DELETE("/api/v1/me/favorites/tracks/{trackId}", {
          params: { path: { trackId: track.id } },
        });
        setFav(false);
      } else {
        await client.POST("/api/v1/me/favorites/tracks/{trackId}", {
          params: { path: { trackId: track.id } },
        });
        setFav(true);
      }
    } finally {
      setFavLoading(false);
    }
  }

  function playTrack() {
    if (!track) return;
    playQueue(
      [
        {
          id: track.id,
          title: track.title,
          artistName: track.artistName,
          albumTitle: track.albumTitle,
          durationSeconds: Math.round(track.durationMs / 1000),
          playbackUrl: "",
        },
      ],
      0
    );
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-6 w-24" />
        <div className="flex gap-6">
          <Skeleton className="h-40 w-40 shrink-0 rounded-xl" />
          <div className="flex-1 space-y-3 pt-2">
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-4 w-1/4" />
          </div>
        </div>
      </div>
    );
  }

  if (!track) {
    return <div className="py-20 text-center text-[var(--color-text-muted)]">Track not found.</div>;
  }

  return (
    <div className="space-y-8">
      {/* Back */}
      <button
        type="button"
        onClick={() => router.back()}
        className="flex items-center gap-1.5 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
      >
        <ArrowLeft size={15} />
        Back
      </button>

      {/* Hero */}
      <div className="flex flex-col gap-6 sm:flex-row sm:items-end">
        <Artwork alt={track.title} src={undefined} size="lg" className="h-40 w-40 shrink-0 rounded-xl" />
        <div className="space-y-2">
          <p className="text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">Track</p>
          <h1 className="text-3xl font-bold leading-tight">{track.title}</h1>
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-[var(--color-text-secondary)]">
            {track.artistName && (
              <Link
                href={`/artists/${track.artistId}`}
                className="font-medium text-[var(--color-text)] hover:text-[var(--color-primary)] transition-colors"
              >
                {track.artistName}
              </Link>
            )}
            {track.albumTitle && (
              <>
                <span className="text-[var(--color-border)]">·</span>
                <Link href={`/albums/${track.albumId}`} className="hover:text-[var(--color-primary)] transition-colors">
                  {track.albumTitle}
                </Link>
              </>
            )}
          </div>

          {/* Actions */}
          <div className="flex items-center gap-3 pt-2">
            <button
              type="button"
              onClick={playTrack}
              className="flex h-10 items-center gap-2 rounded-full bg-[var(--color-primary)] px-5 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 transition-opacity glow-primary"
            >
              <Play size={15} fill="currentColor" />
              Play
            </button>
            <button
              type="button"
              onClick={toggleFavorite}
              disabled={favLoading}
              title={fav ? "Remove from favorites" : "Add to favorites"}
              className={cn(
                "flex h-10 w-10 items-center justify-center rounded-full border transition-colors",
                fav
                  ? "border-[var(--color-primary)] text-[var(--color-primary)]"
                  : "border-[var(--color-border)] text-[var(--color-text-muted)] hover:border-[var(--color-primary)] hover:text-[var(--color-primary)]"
              )}
            >
              <Heart size={16} fill={fav ? "currentColor" : "none"} />
            </button>
          </div>
        </div>
      </div>

      {/* Detail grid */}
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <MetaRow icon={<Clock size={14} />} label="Duration">
          {formatDuration(track.durationMs / 1000)}
        </MetaRow>
        {track.genre && (
          <MetaRow icon={<Tag size={14} />} label="Genre">
            {track.genre}
          </MetaRow>
        )}
        {track.trackNumber != null && (
          <MetaRow icon={<Hash size={14} />} label="Track №">
            {track.discNumber != null && track.discNumber > 1
              ? `${track.discNumber}-${track.trackNumber}`
              : String(track.trackNumber)}
          </MetaRow>
        )}
        <MetaRow icon={<Music2 size={14} />} label="Artist">
          <Link href={`/artists/${track.artistId}`} className="hover:text-[var(--color-primary)] transition-colors">
            {track.artistName || "—"}
          </Link>
        </MetaRow>
        <MetaRow icon={<Disc3 size={14} />} label="Album">
          <Link href={`/albums/${track.albumId}`} className="hover:text-[var(--color-primary)] transition-colors">
            {track.albumTitle || "—"}
          </Link>
        </MetaRow>
        <MetaRow icon={<BarChart2 size={14} />} label="Plays">
          {stats?.totalPlays != null ? (
            <span>
              {stats.totalPlays.toLocaleString()}
              {stats.lastPlayedAt && (
                <span className="ml-2 text-xs text-[var(--color-text-muted)]">
                  · last {new Date(stats.lastPlayedAt).toLocaleDateString()}
                </span>
              )}
            </span>
          ) : (
            "—"
          )}
        </MetaRow>
      </div>
    </div>
  );
}

function MetaRow({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] px-4 py-3">
      <span className="mt-0.5 shrink-0 text-[var(--color-text-muted)]">{icon}</span>
      <div className="min-w-0">
        <p className="text-xs text-[var(--color-text-muted)]">{label}</p>
        <p className="mt-0.5 truncate text-sm font-medium">{children}</p>
      </div>
    </div>
  );
}
