/**
 * History page — /library/history
 * Uses user playback history endpoints.
 */
"use client";

import { useEffect, useState } from "react";
import { History, Trash2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";

interface PlayEvent {
  id: string;
  trackId: string;
  playedAt: string;
  createdAt: string;
}

interface Stats {
  totalEvents: number;
  uniqueTracks: number;
}

interface TopTrack {
  trackId: string;
  playCount: number;
}

const PAGE_SIZE = 50;

export default function HistoryPage() {
  const token = useAuthStore((s) => s.token);
  const [events, setEvents] = useState<PlayEvent[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [topTracks, setTopTracks] = useState<TopTrack[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  async function load() {
    if (!token) return;
    setLoading(true);
    const client = authedApi(token);
    const [historyRes, statsRes, topRes] = await Promise.all([
      client.GET("/api/v1/me/history", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE), order: "desc" } },
      }),
      client.GET("/api/v1/me/history/stats"),
      client.GET("/api/v1/me/history/top-tracks", { params: { query: { limit: 5 } } }),
    ]);

    if (historyRes.data) {
      setEvents((historyRes.data.events ?? []).map((e) => ({
        id: e.id,
        trackId: e.trackId,
        playedAt: e.playedAt,
        createdAt: e.createdAt,
      })));
      if (historyRes.data.pagination) setPagination(historyRes.data.pagination);
    }
    if (statsRes.data) setStats(statsRes.data);
    if (topRes.data?.tracks) {
      setTopTracks(topRes.data.tracks.map((t) => ({ trackId: t.trackId, playCount: t.playCount })));
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [token, page]); // eslint-disable-line react-hooks/exhaustive-deps

  async function deleteEvent(eventId: string) {
    if (!token) return;
    await authedApi(token).DELETE("/api/v1/me/history/{eventId}", { params: { path: { eventId } } });
    await load();
  }

  async function clearHistory() {
    if (!token) return;
    if (!window.confirm("Clear all playback history?")) return;
    await authedApi(token).DELETE("/api/v1/me/history");
    await load();
  }

  const empty = !loading && events.length === 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <History size={22} className="text-[var(--color-primary)]" />
          <h1 className="text-2xl font-bold">History</h1>
        </div>
        <button onClick={clearHistory} disabled={events.length === 0} className="rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)] disabled:opacity-50">
          Clear history
        </button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="Plays" value={stats?.totalEvents} loading={loading} />
        <StatCard label="Unique tracks" value={stats?.uniqueTracks} loading={loading} />
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
          <p className="mb-2 text-sm text-[var(--color-muted-foreground)]">Top tracks</p>
          {loading ? <Skeleton className="h-12 w-full" /> : (
            <div className="space-y-1">
              {topTracks.length === 0 ? <p className="text-sm text-[var(--color-muted-foreground)]">No data</p> : topTracks.map((t) => (
                <p key={t.trackId} className="truncate text-xs"><code>{t.trackId}</code> · {t.playCount} plays</p>
              ))}
            </div>
          )}
        </div>
      </div>

      {empty && <EmptyState title="No playback history" description="Played tracks will appear here." />}

      {!empty && (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {loading ? Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="border-b border-[var(--color-border)] px-4 py-3 last:border-0"><Skeleton className="h-5 w-full" /></div>
          )) : events.map((e) => (
            <div key={e.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors">
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium"><code>{e.trackId}</code></p>
                <p className="text-xs text-[var(--color-muted-foreground)]">Played {new Date(e.playedAt).toLocaleString()}</p>
              </div>
              <button onClick={() => deleteEvent(e.id)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]" title="Delete event">
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

function StatCard({ label, value, loading }: { label: string; value?: number; loading: boolean }) {
  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
      <p className="text-sm text-[var(--color-muted-foreground)]">{label}</p>
      {loading ? <Skeleton className="mt-2 h-7 w-16" /> : <p className="mt-2 text-2xl font-bold">{value ?? 0}</p>}
    </div>
  );
}
