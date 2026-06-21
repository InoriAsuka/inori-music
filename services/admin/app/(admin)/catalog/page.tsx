"use client";

import { useEffect, useState } from "react";
import { Trash2, Save, ChevronLeft, ChevronRight } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";

type Tab = "artists" | "albums" | "tracks" | "playlists";
type Row = { id: string; label: string; sub?: string };

const PAGE = 30;

export default function CatalogPage() {
  const client = useAdminClient();
  const [tab, setTab] = useState<Tab>("artists");
  const [rows, setRows] = useState<Row[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<{ id: string; label: string } | null>(null);

  async function load() {
    if (!client) return;
    setLoading(true);
    const q = { limit: PAGE, offset };
    if (tab === "artists") {
      const { data } = await client.GET("/api/v1/admin/catalog/artists", { params: { query: q } });
      if (data) { setRows((data.artists ?? []).map((a) => ({ id: a.id, label: a.name }))); setTotal(data.pagination?.total ?? 0); }
    } else if (tab === "albums") {
      const { data } = await client.GET("/api/v1/admin/catalog/albums", { params: { query: q } });
      if (data) { setRows((data.albums ?? []).map((a) => ({ id: a.id, label: a.title, sub: a.releaseYear ? String(a.releaseYear) : undefined }))); setTotal(data.pagination?.total ?? 0); }
    } else if (tab === "tracks") {
      const { data } = await client.GET("/api/v1/admin/catalog/tracks", { params: { query: q } });
      if (data) { setRows((data.tracks ?? []).map((t) => ({ id: t.id, label: t.title, sub: `${Math.round((t.durationMs ?? 0) / 1000)}s` }))); setTotal(data.pagination?.total ?? 0); }
    } else {
      const { data } = await client.GET("/api/v1/admin/catalog/playlists", { params: { query: q } });
      if (data) { setRows((data.playlists ?? []).map((p) => ({ id: p.id, label: p.name, sub: `${p.trackIds.length} tracks` }))); setTotal(data.pagination?.total ?? 0); }
    }
    setLoading(false);
  }

  useEffect(() => { setOffset(0); }, [tab]);
  useEffect(() => { load(); }, [client, tab, offset]); // eslint-disable-line react-hooks/exhaustive-deps

  async function del(id: string) {
    if (!client || !window.confirm("Delete?")) return;
    if (tab === "artists") await client.DELETE("/api/v1/admin/catalog/artists/{id}", { params: { path: { id } } });
    else if (tab === "albums") await client.DELETE("/api/v1/admin/catalog/albums/{id}", { params: { path: { id } } });
    else if (tab === "tracks") await client.DELETE("/api/v1/admin/catalog/tracks/{id}", { params: { path: { id } } });
    else await client.DELETE("/api/v1/admin/catalog/playlists/{id}", { params: { path: { id } } });
    await load();
  }

  async function save() {
    if (!client || !editing) return;
    if (tab === "artists") await client.PATCH("/api/v1/admin/catalog/artists/{id}", { params: { path: { id: editing.id } }, body: { name: editing.label } });
    else if (tab === "albums") await client.PATCH("/api/v1/admin/catalog/albums/{id}", { params: { path: { id: editing.id } }, body: { title: editing.label } });
    else if (tab === "tracks") await client.PATCH("/api/v1/admin/catalog/tracks/{id}", { params: { path: { id: editing.id } }, body: { title: editing.label } });
    else await client.PATCH("/api/v1/admin/catalog/playlists/{id}", { params: { path: { id: editing.id } }, body: { name: editing.label } });
    setEditing(null); await load();
  }

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;
  const TABS: Tab[] = ["artists", "albums", "tracks", "playlists"];

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">CATALOG</h1>

      <div className="flex gap-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-1 w-fit">
        {TABS.map((t) => (
          <button key={t} onClick={() => setTab(t)}
            className={tab === t
              ? "rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-xs font-semibold text-[var(--color-primary-fg)]"
              : "rounded-md px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)]"}
          >{t}</button>
        ))}
      </div>

      <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] overflow-hidden">
        <div className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-2 text-xs font-semibold uppercase tracking-wider text-[var(--color-text-muted)]">
          <span className="flex-1">Name / Title</span><span className="w-24">ID</span><span className="w-20"></span>
        </div>
        {loading ? (
          <div className="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : rows.map((r) => (
          <div key={r.id} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 last:border-0 hover:bg-[var(--color-surface-raised)] transition-colors">
            <div className="flex-1 min-w-0">
              {editing?.id === r.id ? (
                <input value={editing.label} onChange={(e) => setEditing({ id: r.id, label: e.target.value })}
                  className="w-full rounded-md border border-[var(--color-primary)] bg-[var(--color-void)] px-2 py-1 text-sm text-[var(--color-text)] outline-none"
                />
              ) : (
                <button onClick={() => setEditing({ id: r.id, label: r.label })} className="truncate text-sm font-medium text-[var(--color-text)] hover:text-[var(--color-primary-hover)] text-left w-full">
                  {r.label}
                </button>
              )}
              {r.sub && <p className="text-xs text-[var(--color-text-muted)]">{r.sub}</p>}
            </div>
            <span className="w-24 truncate font-mono text-xs text-[var(--color-text-muted)]">{r.id.slice(0, 8)}…</span>
            <div className="flex items-center gap-1">
              {editing?.id === r.id && (
                <button onClick={save} className="rounded p-1.5 text-[var(--color-primary)] hover:bg-[var(--color-primary-dim)]"><Save size={13} /></button>
              )}
              <button onClick={() => del(r.id)} className="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"><Trash2 size={13} /></button>
            </div>
          </div>
        ))}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm text-[var(--color-text-muted)]">
          <span>{total} items</span>
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
