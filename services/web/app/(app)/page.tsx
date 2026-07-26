/**
 * Home page — stats cards + recently added items.
 * CatalogStats: { artists, albums, tracks, playlists }
 * RecentCatalogResult: { items: RecentCatalogItem[] }
 * RecentCatalogItem: { kind, artist?, album?, track?, playlist?, addedAt }
 */
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Music2, Users, Disc3, ListMusic } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { Skeleton, TrackRowSkeleton } from "@/components/ui/Skeleton";
import { Artwork } from "@/components/ui/Artwork";
import { usePlayerStore } from "@/store/player";
import { formatDuration, cn } from "@/lib/utils";

interface StatsData {
  tracks: number;
  artists: number;
  albums: number;
  playlists: number;
}

interface RecentTrack {
  id: string;
  title: string;
  durationMs: number;
}

export default function HomePage() {
  const token = useAuthStore((s) => s.token);
  const [stats, setStats] = useState<StatsData | null>(null);
  const [recent, setRecent] = useState<RecentTrack[]>([]);
  const [loading, setLoading] = useState(true);
  const playQueue = usePlayerStore((s) => s.playQueue);

  useEffect(() => {
    if (!token) return;
    const client = authedApi(token);

    async function load() {
      const [statsRes, recentRes] = await Promise.all([
        client.GET("/api/v1/catalog/stats"),
        client.GET("/api/v1/catalog/recently-added", { params: { query: { limit: 10 } } }),
      ]);

      if (statsRes.data) {
        setStats({
          tracks: statsRes.data.tracks,
          artists: statsRes.data.artists,
          albums: statsRes.data.albums,
          playlists: statsRes.data.playlists,
        });
      }

      // RecentCatalogResult has items with polymorphic kind.
      // Extract tracks from the items array.
      if (recentRes.data?.items) {
        const trackItems = recentRes.data.items
          .filter((item) => item.kind === "track" && item.track)
          .map((item) => item.track!)
          .map((t) => ({
            id: t.id,
            title: t.title,
            durationMs: t.durationMs ?? 0,
          }));
        setRecent(trackItems);
      }

      setLoading(false);
    }

    load();
  }, [token]);

  function playFrom(idx: number) {
    const q = recent.map((t) => ({
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
    <div className="space-y-10">
      <h1 className="font-display text-3xl font-bold tracking-tight text-[var(--color-text)]">Home</h1>

      {/* Bento stats grid — Tracks is the hero tile at 2x2, the rest are 1x1. */}
      <div className="grid grid-cols-2 gap-3 sm:gap-4 md:grid-cols-4 md:grid-rows-2">
        <StatCard
          label="Tracks"
          value={stats?.tracks}
          icon={<Music2 size={22} />}
          href="/tracks"
          loading={loading}
          hero
        />
        <StatCard label="Artists" value={stats?.artists} icon={<Users size={18} />} href="/artists" loading={loading} />
        <StatCard label="Albums" value={stats?.albums} icon={<Disc3 size={18} />} href="/albums" loading={loading} />
        <StatCard
          label="Playlists"
          value={stats?.playlists}
          icon={<ListMusic size={18} />}
          href="/playlists"
          loading={loading}
          wide
        />
      </div>

      {/* Recently added */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-display text-xl font-bold text-[var(--color-text)]">Recently Added</h2>
          <Link
            href="/tracks"
            className="rounded-md text-sm font-medium text-[var(--color-primary)] hover:text-[var(--color-primary-hover)] transition-colors"
          >
            View all
          </Link>
        </div>

        <div className="overflow-hidden rounded-2xl border border-[var(--color-border)] bg-[var(--color-surface)] card-soft">
          {loading
            ? Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                  <TrackRowSkeleton />
                </div>
              ))
            : recent.map((track, idx) => (
                <button
                  key={track.id}
                  type="button"
                  onClick={() => playFrom(idx)}
                  className="group flex w-full items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 text-left last:border-0 hover:bg-[var(--color-surface-raised)] transition-colors"
                >
                  <span className="w-5 shrink-0 text-center text-sm tabular text-[var(--color-text-muted)] group-hover:text-[var(--color-primary)] transition-colors">
                    {idx + 1}
                  </span>
                  <Artwork
                    alt={track.title}
                    size="sm"
                    className="transition-transform duration-200 group-hover:-translate-y-0.5"
                  />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-[var(--color-text)]">{track.title}</p>
                  </div>
                  <span className="shrink-0 text-xs tabular text-[var(--color-text-muted)]">
                    {formatDuration(track.durationMs / 1000)}
                  </span>
                </button>
              ))}
        </div>
      </section>
    </div>
  );
}

function StatCard({
  label,
  value,
  icon,
  href,
  loading,
  hero,
  wide,
}: {
  label: string;
  value?: number;
  icon: React.ReactNode;
  href: string;
  loading: boolean;
  /** Renders as the 2x2 focal tile with a sakura wash and oversized watermark. */
  hero?: boolean;
  /** Fills the remaining two columns on the second row so the grid has no gap. */
  wide?: boolean;
}) {
  return (
    <Link
      href={href}
      className={cn(
        "group relative flex flex-col justify-between overflow-hidden rounded-2xl border border-[var(--color-border)] p-4 transition-transform duration-200 hover:scale-[1.02] card-soft hover:card-float",
        hero
          ? "col-span-2 row-span-2 bg-gradient-to-br from-[var(--color-primary-dim)] via-[var(--color-surface)] to-[var(--color-secondary-dim)] p-5"
          : "bg-[var(--color-surface)] hover:bg-[var(--color-surface-raised)]",
        wide && "md:col-span-2"
      )}
    >
      {hero && (
        <Music2
          size={168}
          strokeWidth={1}
          aria-hidden="true"
          className="pointer-events-none absolute -right-8 -bottom-8 text-[var(--color-primary)] opacity-[0.14]"
        />
      )}

      <span className={cn("relative", hero ? "text-[var(--color-primary)]" : "text-[var(--color-text-secondary)]")}>
        {icon}
      </span>

      <div className="relative mt-6 space-y-0.5">
        {loading ? (
          <Skeleton className={hero ? "h-12 w-28" : "h-7 w-16"} />
        ) : (
          <span
            className={cn(
              "block font-display font-bold tabular text-[var(--color-text)]",
              hero ? "text-5xl sm:text-6xl" : "text-2xl"
            )}
          >
            {value?.toLocaleString() ?? "—"}
          </span>
        )}
        <span
          className={cn(
            "block text-[var(--color-text-secondary)]",
            hero ? "text-base font-medium" : "text-sm"
          )}
        >
          {label}
        </span>
      </div>
    </Link>
  );
}
