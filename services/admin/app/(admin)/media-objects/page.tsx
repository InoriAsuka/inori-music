"use client";

import { useEffect, useState } from "react";
import { Trash2, ChevronLeft, ChevronRight } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";
import { formatDateTime } from "@/lib/utils";

const PAGE = 50;

export default function MediaObjectsPage() {
  const client = useAdminClient();
  const [objects, setObjects] = useState<{ id: string; objectKey: string; mimeType: string; sizeBytes: number; lifecycleState: string; backendId: string }[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);

  async function load() {
    if (!client) return;
    setLoading(true);
    const { data } = await client.GET("/api/v1/admin/media/objects", { params: { query: { limit: PAGE, offset } } });
    if (data) {
      setObjects((data.objects ?? []).map((o) => ({ id: o.id, objectKey: o.objectKey, mimeType: o.mimeType, sizeBytes: o.sizeBytes, lifecycleState: o.lifecycleState, backendId: o.backendId })));
      setTotal((data as { pagination?: { total?: number } }).pagination?.total ?? 0);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [client, offset]); // eslint-disable-line react-hooks/exhaustive-deps

  const STATE_COLOR: Record<string, string> = {
    active: "text-[var(--color-success)] border-[var(--color-success)]",
    archived: "text-[var(--color-warning)] border-[var(--color-warning)]",
    deleted: "text-[var(--color-danger)] border-[var(--color-danger)]",
  };

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">MEDIA OBJECTS</h1>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] overflow-hidden">
        <div className="grid grid-cols-[1fr_120px_90px_auto] border-b border-[var(--color-border)] px-4 py-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          <span>Object key</span><span>MIME</span><span>State</span><span className="w-16"></span>
        </div>
        {loading ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : objects.length === 0 ? (
          <div className="py-8 text-center text-sm text-[var(--color-text-muted)]">No media objects.</div>
        ) : objects.map((o) => (
          <div key={o.id} className="grid grid-cols-[1fr_120px_90px_auto] items-center gap-0 border-b border-[var(--color-border)] px-4 py-3 last:border-0 hover:bg-[var(--color-surface-raised)]">
            <div>
              <p className="font-mono text-xs text-[var(--color-text)]">{o.objectKey}</p>
              <p className="font-mono text-xs text-[var(--color-text-muted)]">{o.id.slice(0, 12)}… · {o.backendId.slice(0, 8)}…</p>
            </div>
            <span className="font-mono text-xs text-[var(--color-text-muted)]">{o.mimeType}</span>
            <span className={`rounded border px-1.5 py-0.5 text-[10px] font-semibold w-fit ${STATE_COLOR[o.lifecycleState] ?? "text-[var(--color-text-muted)] border-[var(--color-border)]"}`}>
              {o.lifecycleState}
            </span>
            <div className="w-16 flex justify-end">
              <span className="font-mono text-xs text-[var(--color-text-muted)]">
                {(o.sizeBytes / 1024 / 1024).toFixed(1)} MB
              </span>
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
    </div>
  );
}
