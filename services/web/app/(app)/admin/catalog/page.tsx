/**
 * Admin catalog — /admin/catalog
 * Basic CRUD management over artists, albums, and tracks.
 */
"use client";

import { useEffect, useState } from "react";
import { Trash2, Save } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { PaginationBar, type OffsetPagination, offsetFromPage } from "@/components/ui/PaginationBar";
import { Skeleton } from "@/components/ui/Skeleton";
import { EmptyState } from "@/components/ui/EmptyState";

const PAGE_SIZE = 30;
type Tab = "artists" | "albums" | "tracks";

type Row = { id: string; title: string; subtitle?: string; raw: unknown };

export default function AdminCatalogPage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [tab, setTab] = useState<Tab>("artists");
  const [rows, setRows] = useState<Row[]>([]);
  const [pagination, setPagination] = useState<OffsetPagination | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState<{ id: string; title: string } | null>(null);

  async function load() {
    if (!admin) { setLoading(false); return; }
    setLoading(true);
    const query = { limit: PAGE_SIZE, offset: offsetFromPage(page, PAGE_SIZE) };
    if (tab === "artists") {
      const { data } = await admin.GET("/api/v1/admin/catalog/artists", { params: { query } });
      if (data) { setRows((data.artists ?? []).map((a) => ({ id: a.id, title: a.name, raw: a }))); if (data.pagination) setPagination(data.pagination); }
    } else if (tab === "albums") {
      const { data } = await admin.GET("/api/v1/admin/catalog/albums", { params: { query } });
      if (data) { setRows((data.albums ?? []).map((a) => ({ id: a.id, title: a.title, subtitle: a.releaseYear ? String(a.releaseYear) : undefined, raw: a }))); if (data.pagination) setPagination(data.pagination); }
    } else {
      const { data } = await admin.GET("/api/v1/admin/catalog/tracks", { params: { query } });
      if (data) { setRows((data.tracks ?? []).map((t) => ({ id: t.id, title: t.title, subtitle: `${Math.round((t.durationMs ?? 0) / 1000)}s`, raw: t }))); if (data.pagination) setPagination(data.pagination); }
    }
    setLoading(false);
  }

  useEffect(() => { setPage(1); }, [tab]);
  useEffect(() => { load(); }, [admin, tab, page]); // eslint-disable-line react-hooks/exhaustive-deps

  async function deleteRow(id: string) {
    if (!admin || !window.confirm(`Delete ${tab.slice(0, -1)}?`)) return;
    if (tab === "artists") await admin.DELETE("/api/v1/admin/catalog/artists/{id}", { params: { path: { id } } });
    else if (tab === "albums") await admin.DELETE("/api/v1/admin/catalog/albums/{id}", { params: { path: { id } } });
    else await admin.DELETE("/api/v1/admin/catalog/tracks/{id}", { params: { path: { id } } });
    await load();
  }

  async function saveTitle() {
    if (!admin || !editing) return;
    if (tab === "artists") {
      await admin.PATCH("/api/v1/admin/catalog/artists/{id}", { params: { path: { id: editing.id } }, body: { name: editing.title } });
    } else if (tab === "albums") {
      await admin.PATCH("/api/v1/admin/catalog/albums/{id}", { params: { path: { id: editing.id } }, body: { title: editing.title } });
    } else {
      await admin.PATCH("/api/v1/admin/catalog/tracks/{id}", { params: { path: { id: editing.id } }, body: { title: editing.title } });
    }
    setEditing(null);
    await load();
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Catalog</h1>
      <AdminTokenPanel />
      {!hasAccess && <EmptyState title="Admin access required" description="Sign in as an admin or paste a bootstrap token." />}
      {hasAccess && (
        <>
          <div className="flex gap-2">
            {(["artists", "albums", "tracks"] as Tab[]).map((t) => (
              <button key={t} onClick={() => setTab(t)} className={tab === t ? "rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-sm font-semibold text-[var(--color-primary-foreground)]" : "rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)]"}>
                {t}
              </button>
            ))}
          </div>

          <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
            {loading ? Array.from({ length: 8 }).map((_, i) => <div key={i} className="border-b border-[var(--color-border)] px-4 py-3 last:border-0"><Skeleton className="h-6 w-full" /></div>) : rows.map((r) => (
              <div key={r.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 last:border-0">
                <div className="min-w-0 flex-1">
                  {editing?.id === r.id ? (
                    <input value={editing.title} onChange={(e) => setEditing({ id: r.id, title: e.target.value })} className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-2 py-1 text-sm outline-none focus:border-[var(--color-primary)]" />
                  ) : (
                    <button onClick={() => setEditing({ id: r.id, title: r.title })} className="block w-full truncate text-left text-sm font-medium hover:underline">{r.title}</button>
                  )}
                  <p className="truncate text-xs text-[var(--color-muted-foreground)]">{r.id}{r.subtitle ? ` · ${r.subtitle}` : ""}</p>
                </div>
                {editing?.id === r.id && <button onClick={saveTitle} className="rounded p-1.5 text-[var(--color-primary)]" title="Save"><Save size={15} /></button>}
                <button onClick={() => deleteRow(r.id)} className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]" title="Delete"><Trash2 size={15} /></button>
              </div>
            ))}
          </div>
          {pagination && <PaginationBar pagination={pagination} onPageChange={setPage} />}
        </>
      )}
    </div>
  );
}
