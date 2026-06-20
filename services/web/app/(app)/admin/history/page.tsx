/**
 * Admin history — /admin/history
 * Global playback history with stats and delete controls.
 * HistoryStats: { totalEvents, uniqueUsers, uniqueTracks }
 * TopTracksResult.tracks: TrackPlayCount[] { trackId, playCount }
 * TopUsersResult.users: UserPlayCount[] { userId, playCount }
 */
"use client";

import { useEffect, useState } from "react";
import { Trash2, Activity } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";

interface PlayEvent { id: string; userId: string; trackId: string; playedAt: string; createdAt: string; }
interface Stats { totalEvents: number; uniqueUsers: number; uniqueTracks: number; }

const PAGE_SIZE = 50;

export default function AdminHistoryPage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [events, setEvents] = useState<PlayEvent[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [stats, setStats] = useState<Stats | null>(null);
  const [topTracks, setTopTracks] = useState<{ trackId: string; playCount: number }[]>([]);
  const [topUsers, setTopUsers] = useState<{ userId: string; playCount: number }[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  async function load() {
    if (!admin) { setLoading(false); return; }
    setLoading(true);
    const [histRes, statsRes, tracksRes, usersRes] = await Promise.all([
      admin.GET("/api/v1/admin/history", {
        params: { query: { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE), order: "desc" } },
      }),
      admin.GET("/api/v1/admin/history/stats"),
      admin.GET("/api/v1/admin/history/top-tracks", { params: { query: { limit: 5 } } }),
      admin.GET("/api/v1/admin/history/top-users", { params: { query: { limit: 5 } } }),
    ]);

    if (histRes.data) {
      setEvents((histRes.data.events ?? []).map((e) => ({ id: e.id, userId: e.userId, trackId: e.trackId, playedAt: e.playedAt, createdAt: e.createdAt })));
      if (histRes.data.pagination) setPagination(histRes.data.pagination);
    }
    if (statsRes.data) setStats(statsRes.data);
    if (tracksRes.data?.tracks) setTopTracks(tracksRes.data.tracks.map((t) => ({ trackId: t.trackId, playCount: t.playCount })));
    if (usersRes.data) {
      const users = (usersRes.data as unknown as { users?: { userId: string; playCount: number }[] }).users ?? [];
      setTopUsers(users);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [admin, page]); // eslint-disable-line react-hooks/exhaustive-deps

  async function deleteEvent(id: string) {
    if (!admin) return;
    await admin.DELETE("/api/v1/admin/history/{eventId}", { params: { path: { eventId: id } } });
    await load();
  }

  async function clearWindow() {
    if (!admin) return;
    const until = window.prompt("Delete all events until (ISO timestamp, e.g. 2024-01-01T00:00:00Z):");
    if (!until) return;
    await admin.DELETE("/api/v1/admin/history", { params: { query: { until } } });
    await load();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Activity size={22} className="text-[var(--color-primary)]" />
          <h1 className="text-2xl font-bold">History</h1>
        </div>
        <button onClick={clearWindow} className="rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)]">
          Clear by date
        </button>
      </div>

      <AdminTokenPanel />
      {!hasAccess && <EmptyState title="Admin access required" description="Sign in as an admin or paste a bootstrap token." />}

      {hasAccess && (
        <>
          <div className="grid gap-4 sm:grid-cols-3">
            {(["totalEvents", "uniqueUsers", "uniqueTracks"] as const).map((k) => (
              <div key={k} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
                <p className="text-sm text-[var(--color-muted-foreground)]">{k === "totalEvents" ? "Play events" : k === "uniqueUsers" ? "Unique users" : "Unique tracks"}</p>
                {loading ? <Skeleton className="mt-2 h-7 w-16" /> : <p className="mt-2 text-2xl font-bold">{stats?.[k] ?? 0}</p>}
              </div>
            ))}
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
              <p className="mb-2 font-semibold">Top Tracks</p>
              {topTracks.length === 0 ? <p className="text-sm text-[var(--color-muted-foreground)]">No data</p> : topTracks.map((t, i) => (
                <div key={t.trackId} className="flex items-center gap-2 py-1 text-sm">
                  <span className="w-5 text-right text-[var(--color-muted-foreground)]">{i + 1}</span>
                  <code className="flex-1 truncate text-xs">{t.trackId}</code>
                  <span className="text-[var(--color-muted-foreground)]">{t.playCount}</span>
                </div>
              ))}
            </div>
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
              <p className="mb-2 font-semibold">Top Users</p>
              {topUsers.length === 0 ? <p className="text-sm text-[var(--color-muted-foreground)]">No data</p> : topUsers.map((u, i) => (
                <div key={u.userId} className="flex items-center gap-2 py-1 text-sm">
                  <span className="w-5 text-right text-[var(--color-muted-foreground)]">{i + 1}</span>
                  <code className="flex-1 truncate text-xs">{u.userId}</code>
                  <span className="text-[var(--color-muted-foreground)]">{u.playCount}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
            {loading ? Array.from({ length: 8 }).map((_, i) => <div key={i} className="border-b border-[var(--color-border)] px-4 py-3 last:border-0"><Skeleton className="h-5 w-full" /></div>) : events.length === 0 ? (
              <div className="p-8 text-center text-sm text-[var(--color-muted-foreground)]">No events</div>
            ) : events.map((e) => (
              <div key={e.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors">
                <div className="min-w-0 flex-1 grid grid-cols-[1fr_1fr] gap-x-4">
                  <p className="truncate text-xs"><span className="text-[var(--color-muted-foreground)]">user</span> <code>{e.userId}</code></p>
                  <p className="truncate text-xs"><span className="text-[var(--color-muted-foreground)]">track</span> <code>{e.trackId}</code></p>
                  <p className="col-span-2 text-xs text-[var(--color-muted-foreground)]">{new Date(e.playedAt).toLocaleString()}</p>
                </div>
                <button onClick={() => deleteEvent(e.id)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]" title="Delete event">
                  <Trash2 size={15} />
                </button>
              </div>
            ))}
          </div>

          {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
        </>
      )}
    </div>
  );
}
