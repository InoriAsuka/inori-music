/**
 * History page — /library/history
 *
 * Two tabs:
 *   Stats  — 30-day timeline bar chart + top-10 tracks榜单
 *   Events — paginated play event list with checkbox multi-select + batch delete
 */
"use client";

import { useEffect, useState, useCallback } from "react";
import { History, Trash2, BarChart2, List, CheckSquare, Square } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { BarChart } from "@/components/ui/BarChart";
import { cn } from "@/lib/utils";

// ─── Types ────────────────────────────────────────────────────────────────────

interface PlayEvent {
  id: string;
  trackId: string;
  playedAt: string;
}

interface HistoryStats {
  totalEvents: number;
  uniqueTracks: number;
}

interface TopTrack {
  trackId: string;
  playCount: number;
}

interface TimelineBucket {
  bucketStart: string;
  eventCount: number;
}

const PAGE_SIZE = 50;
type Tab = "stats" | "events";

// ─── Helpers ──────────────────────────────────────────────────────────────────

function isoDay(date: Date): string {
  return date.toISOString().slice(0, 10);
}

/** since/until covering the last 30 calendar days (UTC). */
function last30Days(): { since: string; until: string } {
  const until = new Date();
  until.setUTCHours(23, 59, 59, 999);
  const since = new Date(until);
  since.setUTCDate(since.getUTCDate() - 29);
  since.setUTCHours(0, 0, 0, 0);
  return { since: since.toISOString(), until: until.toISOString() };
}

function dayLabel(iso: string): string {
  const d = new Date(iso);
  return `${d.getMonth() + 1}/${d.getDate()}`;
}

// ─── Page ────────────────────────────────────────────────────────────────────

export default function HistoryPage() {
  const token = useAuthStore((s) => s.token);
  const [tab, setTab] = useState<Tab>("stats");

  // Stats tab
  const [stats, setStats] = useState<HistoryStats | null>(null);
  const [topTracks, setTopTracks] = useState<TopTrack[]>([]);
  const [timeline, setTimeline] = useState<TimelineBucket[]>([]);
  const [statsLoading, setStatsLoading] = useState(true);

  // Events tab
  const [events, setEvents] = useState<PlayEvent[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [eventsLoading, setEventsLoading] = useState(true);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [batchDeleting, setBatchDeleting] = useState(false);

  const loadStats = useCallback(async () => {
    if (!token) return;
    setStatsLoading(true);
    const client = authedApi(token);
    const { since, until } = last30Days();
    const [statsRes, topRes, timelineRes] = await Promise.all([
      client.GET("/api/v1/me/history/stats"),
      client.GET("/api/v1/me/history/top-tracks", {
        params: { query: { limit: 10 } },
      }),
      client.GET("/api/v1/me/history/timeline", {
        params: { query: { since, until, granularity: "day" } },
      }),
    ]);
    if (statsRes.data) setStats(statsRes.data);
    if (topRes.data?.tracks) setTopTracks(topRes.data.tracks);
    if (timelineRes.data?.buckets) setTimeline(timelineRes.data.buckets);
    setStatsLoading(false);
  }, [token]);

  const loadEvents = useCallback(async () => {
    if (!token) return;
    setEventsLoading(true);
    setSelected(new Set()); // clear selection on reload
    const res = await authedApi(token).GET("/api/v1/me/history", {
      params: {
        query: {
          limit: PAGE_SIZE,
          offset: offsetFromPage(page, PAGE_SIZE),
          order: "desc",
        },
      },
    });
    if (res.data) {
      setEvents(
        (res.data.events ?? []).map((e) => ({
          id: e.id,
          trackId: e.trackId,
          playedAt: e.playedAt,
        }))
      );
      if (res.data.pagination) setPagination(res.data.pagination);
    }
    setEventsLoading(false);
  }, [token, page]);

  useEffect(() => {
    loadStats();
  }, [loadStats]);
  useEffect(() => {
    loadEvents();
  }, [loadEvents]);

  async function deleteEvent(eventId: string) {
    if (!token) return;
    await authedApi(token).DELETE("/api/v1/me/history/{eventId}", {
      params: { path: { eventId } },
    });
    await Promise.all([loadEvents(), loadStats()]);
  }

  async function batchDeleteSelected() {
    if (!token || selected.size === 0 || batchDeleting) return;
    setBatchDeleting(true);
    const ids = [...selected];
    // API accepts max 100 per call — chunk if needed
    for (let i = 0; i < ids.length; i += 100) {
      await authedApi(token).POST("/api/v1/me/history/batch-delete", {
        body: { ids: ids.slice(i, i + 100) },
      });
    }
    setBatchDeleting(false);
    await Promise.all([loadEvents(), loadStats()]);
  }

  function toggleSelect(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleSelectAll() {
    if (selected.size === events.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(events.map((e) => e.id)));
    }
  }

  async function clearHistory() {
    if (!token) return;
    if (!window.confirm("Clear all playback history? This cannot be undone.")) return;
    await authedApi(token).DELETE("/api/v1/me/history");
    setPage(1);
    await Promise.all([loadEvents(), loadStats()]);
  }

  // Fill missing days so every day in the 30-day window has a bar
  const chartData = (() => {
    const map = new Map(timeline.map((b) => [isoDay(new Date(b.bucketStart)), b.eventCount]));
    const result: { label: string; value: number }[] = [];
    const now = new Date();
    for (let i = 29; i >= 0; i--) {
      const d = new Date(now);
      d.setUTCDate(now.getUTCDate() - i);
      const key = isoDay(d);
      result.push({ label: dayLabel(key), value: map.get(key) ?? 0 });
    }
    return result;
  })();

  const isEmpty = !statsLoading && !eventsLoading && (stats?.totalEvents ?? 0) === 0 && events.length === 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <History size={22} className="text-[var(--color-primary)]" />
          <h1 className="text-2xl font-bold">History</h1>
        </div>
        <button
          type="button"
          onClick={clearHistory}
          disabled={(stats?.totalEvents ?? 0) === 0}
          className="rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)] disabled:opacity-40 transition-colors"
        >
          Clear all
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] p-1 w-fit">
        {(["stats", "events"] as Tab[]).map((t) => (
          <button
            type="button"
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              "flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
              tab === t
                ? "bg-[var(--color-primary)] text-[var(--color-primary-fg)]"
                : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
            )}
          >
            {t === "stats" ? <BarChart2 size={14} /> : <List size={14} />}
            {t === "stats" ? "Stats" : "Events"}
          </button>
        ))}
      </div>

      {isEmpty && <EmptyState title="No playback history" description="Tracks you play will appear here." />}

      {/* ── Stats tab ────────────────────────────────────────────── */}
      {tab === "stats" && !isEmpty && (
        <div className="space-y-6">
          <div className="grid gap-4 sm:grid-cols-2">
            <StatCard label="Total plays" value={stats?.totalEvents} loading={statsLoading} />
            <StatCard label="Unique tracks" value={stats?.uniqueTracks} loading={statsLoading} />
          </div>

          <section>
            <h2 className="mb-3 text-base font-semibold">Plays — last 30 days</h2>
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
              {statsLoading ? <Skeleton className="h-32 w-full" /> : <BarChart data={chartData} height={128} />}
            </div>
          </section>

          <section>
            <h2 className="mb-3 text-base font-semibold">Top tracks (all time)</h2>
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] divide-y divide-[var(--color-border)]">
              {statsLoading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="px-4 py-3">
                    <Skeleton className="h-4 w-full" />
                  </div>
                ))
              ) : topTracks.length === 0 ? (
                <p className="px-4 py-6 text-center text-sm text-[var(--color-text-muted)]">No data yet</p>
              ) : (
                topTracks.map((t, i) => (
                  <div key={t.trackId} className="flex items-center gap-3 px-4 py-2.5">
                    <span className="w-5 shrink-0 text-center text-sm font-bold text-[var(--color-text-muted)]">
                      {i + 1}
                    </span>
                    <span className="flex-1 truncate font-mono text-xs text-[var(--color-text-secondary)]">
                      {t.trackId}
                    </span>
                    <span className="shrink-0 rounded-full bg-[var(--color-primary)]/15 px-2 py-0.5 text-xs font-medium text-[var(--color-primary)]">
                      {t.playCount} plays
                    </span>
                  </div>
                ))
              )}
            </div>
          </section>
        </div>
      )}

      {/* ── Events tab ───────────────────────────────────────────── */}
      {tab === "events" && !isEmpty && (
        <div className="space-y-4">
          {/* Batch toolbar */}
          <div className="flex items-center gap-3">
            <button
              type="button"
              onClick={toggleSelectAll}
              className="flex items-center gap-1.5 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
            >
              {selected.size === events.length && events.length > 0 ? (
                <CheckSquare size={15} className="text-[var(--color-primary)]" />
              ) : (
                <Square size={15} />
              )}
              {selected.size === events.length && events.length > 0 ? "Deselect all" : "Select all"}
            </button>
            {selected.size > 0 && (
              <button
                type="button"
                onClick={batchDeleteSelected}
                disabled={batchDeleting}
                className="flex items-center gap-1.5 rounded-md bg-[var(--color-danger)]/10 px-3 py-1.5 text-sm font-medium text-[var(--color-danger)] hover:bg-[var(--color-danger)]/20 disabled:opacity-50 transition-colors"
              >
                <Trash2 size={13} />
                {batchDeleting ? "Deleting…" : `Delete ${selected.size}`}
              </button>
            )}
          </div>

          <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] divide-y divide-[var(--color-border)]">
            {eventsLoading
              ? Array.from({ length: 10 }).map((_, i) => (
                  <div key={i} className="px-4 py-3">
                    <Skeleton className="h-5 w-full" />
                  </div>
                ))
              : events.map((e) => (
                  <div
                    key={e.id}
                    className={cn(
                      "flex items-center gap-3 px-4 py-2.5 hover:bg-[var(--color-muted)] transition-colors",
                      selected.has(e.id) && "bg-[var(--color-primary)]/5"
                    )}
                  >
                    <button
                      type="button"
                      onClick={() => toggleSelect(e.id)}
                      className="shrink-0 text-[var(--color-text-muted)] hover:text-[var(--color-primary)] transition-colors"
                    >
                      {selected.has(e.id) ? (
                        <CheckSquare size={15} className="text-[var(--color-primary)]" />
                      ) : (
                        <Square size={15} />
                      )}
                    </button>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-mono text-xs text-[var(--color-text-secondary)]">{e.trackId}</p>
                      <p className="text-xs text-[var(--color-text-muted)]">{new Date(e.playedAt).toLocaleString()}</p>
                    </div>
                    <button
                      type="button"
                      onClick={() => deleteEvent(e.id)}
                      className="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)] transition-colors"
                      title="Delete event"
                    >
                      <Trash2 size={14} />
                    </button>
                  </div>
                ))}
          </div>
          {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
        </div>
      )}
    </div>
  );
}

function StatCard({
  label,
  value,
  loading,
}: {
  label: string;
  value?: number;
  loading: boolean;
}) {
  return (
    <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
      <p className="text-sm text-[var(--color-text-muted)]">{label}</p>
      {loading ? (
        <Skeleton className="mt-2 h-8 w-20" />
      ) : (
        <p className="mt-2 text-3xl font-bold">{value?.toLocaleString() ?? "0"}</p>
      )}
    </div>
  );
}
