"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, ShieldCheck, X } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";

const PAGE = 50;

const STATE_COLOR: Record<string, string> = {
  active: "text-[var(--color-success)] border-[var(--color-success)]",
  archived: "text-[var(--color-warning)] border-[var(--color-warning)]",
  deleted: "text-[var(--color-danger)] border-[var(--color-danger)]",
  staged: "text-[var(--color-text-muted)] border-[var(--color-border)]",
};

type MO = {
  id: string; objectKey: string; mimeType: string; sizeBytes: number;
  lifecycleState: string; backendId: string; assetKind?: string;
  contentHash?: string;
};

type TimelineEvent = { at: string; type: string; lifecycleState?: string; previousLifecycleState?: string; status?: string; message?: string; };

interface DetailPanel {
  mo: MO;
  timeline: TimelineEvent[];
  timelineLoading: boolean;
  verifying: boolean;
  verifyResult: string | null;
  lifecycleLoading: boolean;
}

export default function MediaObjectsPage() {
  const client = useAdminClient();
  const [objects, setObjects] = useState<MO[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);

  // Stats + duplicates
  const [stats, setStats] = useState<{ totalObjects: number; totalSizeBytes: number; byLifecycleState: Record<string, number> } | null>(null);
  const [dupeCount, setDupeCount] = useState<number | null>(null);
  const [statsLoading, setStatsLoading] = useState(false);

  // Detail drawer
  const [detail, setDetail] = useState<DetailPanel | null>(null);

  async function load() {
    if (!client) return;
    setLoading(true);
    const { data } = await client.GET("/api/v1/admin/media/objects", { params: { query: { limit: PAGE, offset } } });
    if (data) {
      setObjects((data.objects ?? []).map((o) => ({
        id: o.id, objectKey: o.objectKey, mimeType: o.mimeType, sizeBytes: o.sizeBytes,
        lifecycleState: o.lifecycleState, backendId: o.backendId,
        assetKind: (o as { assetKind?: string }).assetKind,
        contentHash: (o as { contentHash?: string }).contentHash,
      })));
      setTotal((data as { pagination?: { total?: number } }).pagination?.total ?? 0);
    }
    setLoading(false);
  }

  async function loadStats() {
    if (!client) return;
    setStatsLoading(true);
    const [statsRes, dupeRes] = await Promise.all([
      client.GET("/api/v1/admin/media/objects/stats"),
      client.GET("/api/v1/admin/media/objects/duplicates"),
    ]);
    if (statsRes.data) {
      setStats({
        totalObjects: (statsRes.data as { totalObjects: number }).totalObjects,
        totalSizeBytes: (statsRes.data as { totalSizeBytes: number }).totalSizeBytes,
        byLifecycleState: (statsRes.data as { byLifecycleState: Record<string, number> }).byLifecycleState ?? {},
      });
    }
    if (dupeRes.data) {
      setDupeCount((dupeRes.data as { totalGroups?: number }).totalGroups ?? 0);
    }
    setStatsLoading(false);
  }

  useEffect(() => { load(); }, [client, offset]); // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => { loadStats(); }, [client]); // eslint-disable-line react-hooks/exhaustive-deps

  async function openDetail(mo: MO) {
    if (!client) return;
    const panel: DetailPanel = { mo, timeline: [], timelineLoading: true, verifying: false, verifyResult: null, lifecycleLoading: false };
    setDetail(panel);
    const { data } = await client.GET("/api/v1/admin/media/objects/{id}/timeline", { params: { path: { id: mo.id } } });
    const events = ((data as { events?: TimelineEvent[] } | null)?.events ?? []);
    setDetail((d) => d ? { ...d, timeline: events, timelineLoading: false } : null);
  }

  async function verify(moId: string) {
    if (!client) return;
    setDetail((d) => d ? { ...d, verifying: true, verifyResult: null } : null);
    const { data } = await client.POST("/api/v1/admin/media/objects/{id}/verify", { params: { path: { id: moId } } });
    const status = (data as { status?: string } | null)?.status ?? "unknown";
    setDetail((d) => d ? { ...d, verifying: false, verifyResult: status } : null);
  }

  async function setLifecycle(moId: string, state: string) {
    if (!client) return;
    setDetail((d) => d ? { ...d, lifecycleLoading: true } : null);
    const { data } = await client.POST("/api/v1/admin/media/objects/{id}/lifecycle", {
      params: { path: { id: moId } },
      body: { lifecycleState: state as "staged" | "active" | "archived" | "deleted" },
    });
    if (data) {
      const updated = data as MO;
      setDetail((d) => d ? { ...d, lifecycleLoading: false, mo: { ...d.mo, lifecycleState: updated.lifecycleState } } : null);
      await load();
    } else {
      setDetail((d) => d ? { ...d, lifecycleLoading: false } : null);
    }
  }

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">MEDIA OBJECTS</h1>

      {/* Stats bar */}
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {[
          ["Total", statsLoading ? "…" : (stats?.totalObjects ?? "—").toLocaleString()],
          ["Size", statsLoading ? "…" : stats ? `${(stats.totalSizeBytes / 1024 / 1024 / 1024).toFixed(2)} GB` : "—"],
          ["Active", statsLoading ? "…" : stats?.byLifecycleState?.active ?? "—"],
          ["Dup. groups", statsLoading ? "…" : dupeCount ?? "—"],
        ].map(([label, val]) => (
          <div key={label} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-3">
            <p className="text-xs text-[var(--color-text-muted)]">{label}</p>
            <p className="mt-1 font-mono text-lg font-bold text-[var(--color-text)]">{String(val)}</p>
          </div>
        ))}
      </div>

      {/* List */}
      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] overflow-hidden">
        <div className="grid grid-cols-[1fr_120px_90px_auto] border-b border-[var(--color-border)] px-4 py-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          <span>Object key</span><span>MIME</span><span>State</span><span className="w-20"></span>
        </div>
        {loading ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : objects.length === 0 ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">No media objects.</div>
        ) : objects.map((o) => (
          <div key={o.id} className="grid grid-cols-[1fr_120px_90px_auto] items-center gap-0 border-b border-[var(--color-border)] px-4 py-3 last:border-0 hover:bg-[var(--color-surface-raised)] transition-colors cursor-pointer"
            onClick={() => openDetail(o)}>
            <div>
              <p className="font-mono text-xs text-[var(--color-text)]">{o.objectKey}</p>
              <p className="font-mono text-xs text-[var(--color-text-muted)]">{o.id.slice(0, 12)}… · {o.backendId.slice(0, 8)}…</p>
            </div>
            <span className="font-mono text-xs text-[var(--color-text-muted)]">{o.mimeType}</span>
            <span className={`rounded border px-1.5 py-0.5 text-[10px] font-semibold w-fit ${STATE_COLOR[o.lifecycleState] ?? "text-[var(--color-text-muted)] border-[var(--color-border)]"}`}>
              {o.lifecycleState}
            </span>
            <div className="w-20 flex justify-end">
              <span className="font-mono text-xs text-[var(--color-text-muted)]">{(o.sizeBytes / 1024 / 1024).toFixed(1)} MB</span>
            </div>
          </div>
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-[var(--color-text-muted)]">
          <span>{total} objects</span>
          <div className="flex items-center gap-2">
            <button onClick={() => setOffset(Math.max(0, offset - PAGE))} disabled={page <= 1} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronLeft size={16} /></button>
            <span>{page} / {totalPages}</span>
            <button onClick={() => setOffset(offset + PAGE)} disabled={page >= totalPages} className="rounded p-1 hover:bg-[var(--color-surface-raised)] disabled:opacity-30"><ChevronRight size={16} /></button>
          </div>
        </div>
      )}

      {/* Detail drawer */}
      {detail && (
        <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60" onClick={() => setDetail(null)}>
          <div className="w-full max-w-xl max-h-[80vh] overflow-y-auto rounded-t-2xl sm:rounded-2xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6 space-y-5"
            onClick={(e) => e.stopPropagation()}>

            {/* Header */}
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="font-mono text-xs text-[var(--color-text-muted)] break-all">{detail.mo.id}</p>
                <p className="mt-1 font-mono text-sm text-[var(--color-text)] break-all">{detail.mo.objectKey}</p>
              </div>
              <button onClick={() => setDetail(null)} className="shrink-0 rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"><X size={16} /></button>
            </div>

            {/* Meta */}
            <div className="grid grid-cols-2 gap-2 text-xs">
              {[
                ["MIME", detail.mo.mimeType],
                ["Size", `${(detail.mo.sizeBytes / 1024 / 1024).toFixed(2)} MB`],
                ["Backend", detail.mo.backendId],
                ["Kind", detail.mo.assetKind ?? "—"],
                ["Hash", detail.mo.contentHash?.slice(0, 16) ?? "—"],
              ].map(([k, v]) => (
                <div key={k} className="rounded-lg border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2">
                  <p className="text-[var(--color-text-muted)]">{k}</p>
                  <p className="font-mono text-[var(--color-text)] truncate">{v}</p>
                </div>
              ))}
              {/* Lifecycle state with inline change */}
              <div className="rounded-lg border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2">
                <p className="text-[var(--color-text-muted)] mb-1">Lifecycle</p>
                <select value={detail.mo.lifecycleState} disabled={detail.lifecycleLoading}
                  onChange={(e) => setLifecycle(detail.mo.id, e.target.value)}
                  className="rounded border border-[var(--color-border)] bg-[var(--color-surface)] px-2 py-0.5 text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
                >
                  {["staged", "active", "archived", "deleted"].map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </div>
            </div>

            {/* Verify */}
            <div className="flex items-center gap-3">
              <button onClick={() => verify(detail.mo.id)} disabled={detail.verifying}
                className="flex items-center gap-1.5 rounded-md border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:border-[var(--color-primary)] hover:text-[var(--color-primary)] disabled:opacity-50 transition-colors">
                <ShieldCheck size={13} />
                {detail.verifying ? "Verifying…" : "Verify integrity"}
              </button>
              {detail.verifyResult && (
                <span className={`text-xs font-semibold ${detail.verifyResult === "ok" ? "text-[var(--color-success)]" : "text-[var(--color-danger)]"}`}>
                  {detail.verifyResult}
                </span>
              )}
            </div>

            {/* Timeline */}
            <div>
              <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">Timeline</p>
              {detail.timelineLoading ? (
                <p className="text-xs text-[var(--color-text-muted)]">Loading…</p>
              ) : detail.timeline.length === 0 ? (
                <p className="text-xs text-[var(--color-text-muted)]">No events</p>
              ) : (
                <div className="space-y-1.5 max-h-48 overflow-y-auto">
                  {detail.timeline.map((e, i) => (
                    <div key={i} className="flex items-start gap-2 rounded-lg border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2">
                      <div className="flex-1 min-w-0">
                        <p className="text-xs font-medium text-[var(--color-text)]">{e.type}</p>
                        {e.lifecycleState && (
                          <p className="text-xs text-[var(--color-text-muted)]">
                            {e.previousLifecycleState && <>{e.previousLifecycleState} → </>}{e.lifecycleState}
                          </p>
                        )}
                        {e.message && <p className="text-xs text-[var(--color-text-muted)] truncate">{e.message}</p>}
                      </div>
                      <span className="shrink-0 font-mono text-[10px] text-[var(--color-text-muted)]">
                        {new Date(e.at).toLocaleString()}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
