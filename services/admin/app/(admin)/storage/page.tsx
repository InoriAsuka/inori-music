"use client";

import { useEffect, useState } from "react";
import { RefreshCw, Zap, Power, Star, Trash2 } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";
import { formatBytes, cn } from "@/lib/utils";
import { StorageHealthBadge } from "@/components/admin/StorageHealthBadge";

interface Backend {
  id: string; displayName: string; type: string;
  enabled: boolean; isDefault: boolean;
  healthStatus?: string;
  capacity?: { totalBytes: number; usedBytes: number; availableBytes: number };
}

export default function StoragePage() {
  const client = useAdminClient();
  const [backends, setBackends] = useState<Backend[]>([]);
  const [loading, setLoading] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [probeResults, setProbeResults] = useState<Record<string, string>>({});

  async function load() {
    if (!client) return;
    setLoading(true);
    const { data } = await client.GET("/api/v1/admin/storage/backends");
    if (data?.backends) {
      setBackends(data.backends.map((b) => ({
        id: b.id, displayName: b.displayName, type: b.type,
        enabled: b.enabled ?? false, isDefault: b.isDefault ?? false,
        healthStatus: (b as { healthStatus?: string }).healthStatus,
        capacity: (b as { lastCapacity?: { totalBytes: number; usedBytes: number; availableBytes: number } }).lastCapacity,
      })));
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [client]); // eslint-disable-line react-hooks/exhaustive-deps

  async function refresh() {
    if (!client) return;
    setRefreshing(true);
    await client.POST("/api/v1/admin/storage/backends/refresh");
    await load(); setRefreshing(false);
  }

  async function probe(id: string) {
    if (!client) return;
    const { data } = await client.POST("/api/v1/admin/storage/backends/{id}/probe", { params: { path: { id } } });
    if (data) setProbeResults((r) => ({ ...r, [id]: `${data.status}: ${data.message ?? "ok"}${data.checkedAt ? ` @ ${new Date(data.checkedAt).toLocaleTimeString()}` : ""}` }));
    await load();
  }

  async function toggle(b: Backend) {
    if (!client) return;
    const path = b.enabled ? "/api/v1/admin/storage/backends/{id}/disable" : "/api/v1/admin/storage/backends/{id}/enable";
    await client.POST(path, { params: { path: { id: b.id } } });
    await load();
  }

  async function setDefault(id: string) {
    if (!client) return;
    await client.POST("/api/v1/admin/storage/backends/{id}/default", { params: { path: { id } } });
    await load();
  }

  async function del(id: string) {
    if (!client || !window.confirm("Delete backend?")) return;
    await client.DELETE("/api/v1/admin/storage/backends/{id}", { params: { path: { id } } });
    await load();
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">STORAGE</h1>
        <button onClick={refresh} disabled={refreshing}
          className="flex items-center gap-1.5 rounded-md border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:border-[var(--color-border-glow)] hover:text-[var(--color-text)] disabled:opacity-50">
          <RefreshCw size={13} className={refreshing ? "animate-spin" : ""} /> Refresh all
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-[var(--color-text-muted)]">Loading…</div>
      ) : backends.length === 0 ? (
        <div className="rounded-xl border border-dashed border-[var(--color-border)] p-12 text-center text-sm text-[var(--color-text-muted)]">
          No storage backends registered.
        </div>
      ) : (
        <div className="space-y-4">
          {backends.map((b) => {
            const usedPct = b.capacity ? (b.capacity.usedBytes / b.capacity.totalBytes) * 100 : 0;
            return (
              <div key={b.id} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-5">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div className="space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-semibold text-[var(--color-text)]">{b.displayName}</span>
                      <span className="rounded border border-[var(--color-border)] px-1.5 py-0.5 font-mono text-xs text-[var(--color-text-muted)]">{b.type}</span>
                      {b.isDefault && <span className="rounded-full bg-[var(--color-primary-dim)] border border-[var(--color-primary)] px-2 py-0.5 text-[10px] font-semibold text-[var(--color-primary)]">DEFAULT</span>}
                      {!b.enabled && <span className="text-xs text-[var(--color-text-muted)]">disabled</span>}
                    </div>
                    <p className="font-mono text-xs text-[var(--color-text-muted)]">{b.id}</p>
                    <StorageHealthBadge status={b.healthStatus} />
                    {b.capacity && (
                      <div className="w-48 space-y-1">
                        <div className="h-1.5 rounded-full bg-[var(--color-surface-raised)]">
                          <div className="h-full rounded-full bg-[var(--color-secondary)]" style={{ width: `${Math.min(100, usedPct)}%` }} />
                        </div>
                        <p className="text-xs text-[var(--color-text-muted)]">
                          {formatBytes(b.capacity.usedBytes)} / {formatBytes(b.capacity.totalBytes)} · {formatBytes(b.capacity.availableBytes)} free
                        </p>
                      </div>
                    )}
                    {probeResults[b.id] && <p className="text-xs text-[var(--color-info)]">{probeResults[b.id]}</p>}
                  </div>

                  <div className="flex items-center gap-1.5">
                    <ActionBtn onClick={() => probe(b.id)} label="Probe"><Zap size={13} /></ActionBtn>
                    {!b.isDefault && <ActionBtn onClick={() => setDefault(b.id)} label="Set default"><Star size={13} /></ActionBtn>}
                    <ActionBtn onClick={() => toggle(b)} label={b.enabled ? "Disable" : "Enable"}><Power size={13} /></ActionBtn>
                    <ActionBtn onClick={() => del(b.id)} label="Delete" danger><Trash2 size={13} /></ActionBtn>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}

function ActionBtn({ children, onClick, label, danger }: { children: React.ReactNode; onClick: () => void; label: string; danger?: boolean }) {
  return (
    <button onClick={onClick} title={label}
      className={cn("rounded-md border border-[var(--color-border)] px-2 py-1.5 text-xs transition-colors hover:border-[var(--color-border-glow)]",
        danger ? "text-[var(--color-text-muted)] hover:border-[var(--color-danger)] hover:text-[var(--color-danger)]"
               : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]")}
    >{children}</button>
  );
}
