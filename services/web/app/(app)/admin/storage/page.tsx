/**
 * Admin storage — /admin/storage
 * Lists storage backends, shows health/capacity, allows probe, enable/disable, set-default.
 */
"use client";

import { useEffect, useState } from "react";
import { RefreshCw, Zap, Power, Star, Trash2 } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";
import { cn } from "@/lib/utils";

interface Backend {
  id: string;
  displayName: string;
  type: string;
  enabled: boolean;
  isDefault: boolean;
  healthStatus?: string;
  capacity?: { totalBytes: number; usedBytes: number; availableBytes: number };
}

function formatBytes(b: number): string {
  if (b === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(b) / Math.log(1024));
  return `${(b / 1024 ** i).toFixed(1)} ${units[i]}`;
}

export default function AdminStoragePage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [backends, setBackends] = useState<Backend[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [probing, setProbing] = useState<string | null>(null);
  const [probeResults, setProbeResults] = useState<Record<string, string>>({});

  async function load() {
    if (!admin) { setLoading(false); return; }
    const { data } = await admin.GET("/api/v1/admin/storage/backends");
    if (data?.backends) {
      const bs: Backend[] = [];
      for (const b of data.backends) {
        const cap = (b as unknown as { lastCapacity?: { totalBytes: number; usedBytes: number; availableBytes: number } }).lastCapacity;
        bs.push({
          id: b.id,
          displayName: b.displayName,
          type: b.type,
          enabled: b.enabled ?? false,
          isDefault: b.isDefault ?? false,
          healthStatus: (b as unknown as { healthStatus?: string }).healthStatus,
          capacity: cap,
        });
      }
      setBackends(bs);
    }
    setLoading(false);
  }

  useEffect(() => { load(); }, [admin]); // eslint-disable-line react-hooks/exhaustive-deps

  async function probe(id: string) {
    if (!admin) return;
    setProbing(id);
    const { data } = await admin.POST("/api/v1/admin/storage/backends/{id}/probe", { params: { path: { id } } });
    if (data) setProbeResults((r) => ({ ...r, [id]: `${data.status}: ${data.message ?? "ok"}${data.checkedAt ? ` (${new Date(data.checkedAt).toLocaleTimeString()})` : ""}` }));
    setProbing(null);
    await load();
  }

  async function refresh() {
    if (!admin) return;
    setRefreshing(true);
    await admin.POST("/api/v1/admin/storage/backends/refresh");
    await load();
    setRefreshing(false);
  }

  async function toggle(b: Backend) {
    if (!admin) return;
    const path = b.enabled ? "/api/v1/admin/storage/backends/{id}/disable" : "/api/v1/admin/storage/backends/{id}/enable";
    await admin.POST(path, { params: { path: { id: b.id } } });
    await load();
  }

  async function setDefault(id: string) {
    if (!admin) return;
    await admin.POST("/api/v1/admin/storage/backends/{id}/default", { params: { path: { id } } });
    await load();
  }

  async function deleteBackend(id: string) {
    if (!admin || !window.confirm("Delete this backend?")) return;
    await admin.DELETE("/api/v1/admin/storage/backends/{id}", { params: { path: { id } } });
    await load();
  }

  const healthColor = (h?: string) =>
    h === "healthy" ? "text-green-600" : h === "unhealthy" ? "text-[var(--color-destructive)]" : "text-[var(--color-muted-foreground)]";

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Storage</h1>
        <button onClick={refresh} disabled={refreshing} className="flex items-center gap-1.5 rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)] disabled:opacity-50">
          <RefreshCw size={14} className={refreshing ? "animate-spin" : ""} /> Refresh all
        </button>
      </div>
      <AdminTokenPanel />
      {!hasAccess && <EmptyState title="Admin access required" description="Sign in as an admin or paste a bootstrap token." />}

      {hasAccess && (
        <div className="space-y-4">
          {loading ? Array.from({ length: 3 }).map((_, i) => <Skeleton key={i} className="h-28 w-full rounded-xl" />) : backends.length === 0 ? (
            <EmptyState title="No storage backends registered" description="Register a backend via the API or CLI." />
          ) : backends.map((b) => (
            <div key={b.id} className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-semibold">{b.displayName}</span>
                    <span className="rounded-full border border-[var(--color-border)] px-2 py-0.5 text-xs">{b.type}</span>
                    {b.isDefault && <span className="rounded-full bg-[var(--color-primary)] px-2 py-0.5 text-xs font-semibold text-[var(--color-primary-foreground)]">default</span>}
                    {!b.enabled && <span className="text-xs text-[var(--color-muted-foreground)]">disabled</span>}
                  </div>
                  <p className="text-xs text-[var(--color-muted-foreground)]">{b.id}</p>
                  <p className={cn("mt-1 text-xs", healthColor(b.healthStatus))}>
                    health: {b.healthStatus ?? "unknown"}
                  </p>
                  {b.capacity && (
                    <p className="text-xs text-[var(--color-muted-foreground)]">
                      {formatBytes(b.capacity.usedBytes)} / {formatBytes(b.capacity.totalBytes)} used ({formatBytes(b.capacity.availableBytes)} free)
                    </p>
                  )}
                  {probeResults[b.id] && <p className="mt-1 text-xs text-[var(--color-muted-foreground)]">{probeResults[b.id]}</p>}
                </div>

                <div className="flex items-center gap-2">
                  <button onClick={() => probe(b.id)} disabled={probing === b.id} className="flex items-center gap-1 rounded-md border border-[var(--color-border)] px-2 py-1 text-xs hover:bg-[var(--color-muted)] disabled:opacity-50">
                    <Zap size={12} /> Probe
                  </button>
                  {!b.isDefault && <button onClick={() => setDefault(b.id)} className="flex items-center gap-1 rounded-md border border-[var(--color-border)] px-2 py-1 text-xs hover:bg-[var(--color-muted)]">
                    <Star size={12} /> Set default
                  </button>}
                  <button onClick={() => toggle(b)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]" title={b.enabled ? "Disable" : "Enable"}>
                    <Power size={15} />
                  </button>
                  <button onClick={() => deleteBackend(b.id)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]" title="Delete">
                    <Trash2 size={15} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
