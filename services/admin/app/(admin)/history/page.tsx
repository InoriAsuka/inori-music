"use client";

import { useEffect, useState } from "react";
import { Trash2, ChevronLeft, ChevronRight } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";
import { HistoryTimelineChart } from "@/components/admin/HistoryTimelineChart";

type Granularity = "day" | "week" | "month";

const PAGE = 50;

interface TimelineBucket { label: string; count: number; }

export default function HistoryPage() {
  const client = useAdminClient();
  const [events, setEvents] = useState<{ id: string; userId: string; trackId: string; playedAt: string }[]>([]);
  const [stats, setStats] = useState<{ totalEvents: number; uniqueUsers: number; uniqueTracks: number } | null>(null);
  const [topTracks, setTopTracks] = useState<{ trackId: string; playCount: number }[]>([]);
  const [topUsers, setTopUsers] = useState<{ userId: string; playCount: number }[]>([]);
  const [buckets, setBuckets] = useState<TimelineBucket[]>([]);
  const [granularity, setGranularity] = useState<Granularity>("day");
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);

  async function load() {
    if (!client) return;
    setLoading(true);
    const [histRes, statsRes, tracksRes, usersRes, tlRes] = await Promise.all([
      client.GET("/api/v1/admin/history", { params: { query: { limit: PAGE, offset, order: "desc" } } }),
      client.GET("/api/v1/admin/history/stats"),
      client.GET("/api/v1/admin/history/top-tracks", { params: { query: { limit: 5 } } }),
      client.GET("/api/v1/admin/history/top-users", { params: { query: { limit: 5 } } }),
      client.GET("/api/v1/admin/history/timeline", { params: { query: { granularity } } }),
    ]);

    if (histRes.data) {
      setEvents((histRes.data.events ?? []).map((e) => ({ id: e.id, userId: e.userId, trackId: e.trackId, playedAt: e.playedAt })));
      setTotal((histRes.data.pagination as { total?: number } | undefined)?.total ?? 0);
    }
    if (statsRes.data) setStats(statsRes.data);
    if (tracksRes.data?.tracks) setTopTracks(tracksRes.data.tracks.map((t) => ({ trackId: t.trackId, playCount: t.playCount })));
    if (usersRes.data) setTopUsers(((usersRes.data as { users?: { userId: string; playCount: number }[] }).users ?? []));
    if (tlRes.data) {
      const tl = (tlRes.data as { buckets?: { period: string; count: number }[] }).buckets ?? [];
      setBuckets(tl.map((b) => ({ label: b.period, count: b.count })));
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [client, offset, granularity]); // eslint-disable-line react-hooks/exhaustive-deps

  async function del(id: string) {
    if (!client) return;
    await client.DELETE("/api/v1/admin/history/{eventId}", { params: { path: { eventId: id } } });
    await load();
  }

  async function clearWindow() {
    const until = window.prompt("Delete all events until (ISO timestamp):");
    if (!until || !client) return;
    await client.DELETE("/api/v1/admin/history", { params: { query: { until } } });
    await load();
  }

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">HISTORY</h1>
        <button onClick={clearWindow} className="rounded-md border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:border-[var(--color-danger)] hover:text-[var(--color-danger)] transition-colors">
          Clear by date
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-3">
        {[["Play events", stats?.totalEvents], ["Unique users", stats?.uniqueUsers], ["Unique tracks", stats?.uniqueTracks]].map(([label, val]) => (
          <div key={label as string} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
            <p className="text-xs text-[var(--color-text-muted)]">{label}</p>
            <p className="mt-2 font-mono text-2xl font-bold text-[var(--color-text)]">{(val as number | undefined)?.toLocaleString() ?? "—"}</p>
          </div>
        ))}
      </div>

      {/* Timeline chart */}
      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
        <HistoryTimelineChart buckets={buckets} granularity={granularity} onGranularityChange={setGranularity} />
      </div>

      {/* Top charts */}
      <div className="grid gap-4 sm:grid-cols-2">
        {[{ title: "Top Tracks", items: topTracks.map((t) => ({ key: t.trackId, count: t.playCount })) },
          { title: "Top Users", items: topUsers.map((u) => ({ key: u.userId, count: u.playCount })) }].map(({ title, items }) => (
          <div key={title} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
            <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">{title}</p>
            <div className="space-y-1.5">
              {items.length === 0 ? <p className="text-xs text-[var(--color-text-muted)]">No data</p> : items.map((item, i) => (
                <div key={item.key} className="flex items-center gap-2">
                  <span className="w-5 text-right font-mono text-xs text-[var(--color-text-muted)]">{i + 1}</span>
                  <code className="flex-1 truncate text-xs text-[var(--color-text)]">{item.key}</code>
                  <span className="font-mono text-xs text-[var(--color-secondary)]">{item.count}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Events list */}
      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] overflow-hidden">
        {loading ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : events.map((e) => (
          <div key={e.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-surface-raised)]">
            <div className="flex-1 grid grid-cols-2 gap-x-4">
              <p className="truncate font-mono text-xs"><span className="text-[var(--color-text-muted)]">user </span><span className="text-[var(--color-text)]">{e.userId}</span></p>
              <p className="truncate font-mono text-xs"><span className="text-[var(--color-text-muted)]">track </span><span className="text-[var(--color-text)]">{e.trackId}</span></p>
              <p className="col-span-2 text-xs text-[var(--color-text-muted)]">{new Date(e.playedAt).toLocaleString()}</p>
            </div>
            <button onClick={() => del(e.id)} className="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"><Trash2 size={13} /></button>
          </div>
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-[var(--color-text-muted)]">
          <span>{total} events</span>
          <div className="flex items-center gap-2">
            <button onClick={() => setOffset(Math.max(0, offset - PAGE))} disabled={page <= 1} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronLeft size={16} /></button>
            <span>{page} / {totalPages}</span>
            <button onClick={() => setOffset(offset + PAGE)} disabled={page >= totalPages} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronRight size={16} /></button>
          </div>
        </div>
      )}
    </div>
  );
}
