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
import { usePlayerStore } from "@/store/player";
import { formatDuration } from "@/lib/utils";

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

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">Home</h1>

      {/* Stats cards */}
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <StatCard label="Tracks" value={stats?.tracks} icon={<Music2 size={18} />} href="/tracks" loading={loading} />
        <StatCard label="Artists" value={stats?.artists} icon={<Users size={18} />} href="/artists" loading={loading} />
        <StatCard label="Albums" value={stats?.albums} icon={<Disc3 size={18} />} href="/albums" loading={loading} />
        <StatCard label="Playlists" value={stats?.playlists} icon={<ListMusic size={18} />} href="/playlists" loading={loading} />
      </div>

      {/* Recently added */}
      <section>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Recently Added</h2>
          <Link href="/tracks" className="text-sm text-[var(--color-primary)] hover:underline">
            View all
          </Link>
        </div>

        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {loading
            ? Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="border-b border-[var(--color-border)] px-4 last:border-0">
                  <TrackRowSkeleton />
                </div>
              ))
            : recent.map((track, idx) => (
                <div
                  key={track.id}
                  onClick={() => {
                    const q = recent.map((t) => ({
                      id: t.id,
                      title: t.title,
                      artistName: "",
                      albumTitle: "",
                      durationSeconds: Math.round(t.durationMs / 1000),
                      playbackUrl: "",
                    }));
                    playQueue(q, idx);
                  }}
                  className="group flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
                >
                  <span className="w-5 text-center text-sm text-[var(--color-muted-foreground)]">{idx + 1}</span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">{track.title}</p>
                  </div>
                  <span className="text-xs text-[var(--color-muted-foreground)]">
                    {formatDuration(track.durationMs / 1000)}
                  </span>
                </div>
              ))}
        </div>
      </section>
    </div>
  );
}

function StatCard({
  label, value, icon, href, loading,
}: {
  label: string;
  value?: number;
  icon: React.ReactNode;
  href: string;
  loading: boolean;
}) {
  return (
    <Link
      href={href}
      className="flex flex-col gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 hover:bg-[var(--color-muted)] transition-colors"
    >
      <span className="text-[var(--color-muted-foreground)]">{icon}</span>
      {loading
        ? <Skeleton className="h-7 w-16" />
        : <span className="text-2xl font-bold">{value?.toLocaleString() ?? "—"}</span>}
      <span className="text-sm text-[var(--color-muted-foreground)]">{label}</span>
    </Link>
  );
}
