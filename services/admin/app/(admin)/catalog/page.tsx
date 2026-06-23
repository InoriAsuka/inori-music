"use client";

import { useEffect, useState } from "react";
import { Trash2, Save, ChevronLeft, ChevronRight, ChevronDown, ChevronUp, Link2, Plus, X } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";

type Tab = "artists" | "albums" | "tracks" | "playlists";
type Row = { id: string; label: string; sub?: string };
type SubRow = { id: string; label: string; sub?: string };

const PAGE = 30;

export default function CatalogPage() {
  const client = useAdminClient();
  const [tab, setTab] = useState<Tab>("artists");
  const [rows, setRows] = useState<Row[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<{ id: string; label: string } | null>(null);

  // Sub-resource expansion
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [subRows, setSubRows] = useState<SubRow[]>([]);
  const [subLoading, setSubLoading] = useState(false);

  // Relink dialog (tracks tab)
  const [relinkId, setRelinkId] = useState<string | null>(null);
  const [relinkMediaId, setRelinkMediaId] = useState("");

  // Playlist track add
  const [addTrackId, setAddTrackId] = useState("");

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

  useEffect(() => { setOffset(0); setExpandedId(null); }, [tab]);
  useEffect(() => { load(); }, [client, tab, offset]); // eslint-disable-line react-hooks/exhaustive-deps

  async function expand(id: string) {
    if (!client) return;
    if (expandedId === id) { setExpandedId(null); return; }
    setExpandedId(id);
    setSubRows([]);
    setSubLoading(true);
    if (tab === "artists") {
      const [albumsRes, tracksRes] = await Promise.all([
        client.GET("/api/v1/admin/catalog/artists/{id}/albums", { params: { path: { id } } }),
        client.GET("/api/v1/admin/catalog/artists/{id}/tracks", { params: { path: { id } } }),
      ]);
      const albums = (albumsRes.data?.albums ?? []).map((a) => ({ id: a.id, label: a.title, sub: "album" }));
      const tracks = (tracksRes.data?.tracks ?? []).map((t) => ({ id: t.id, label: t.title, sub: "track" }));
      setSubRows([...albums, ...tracks]);
    } else if (tab === "albums") {
      const { data } = await client.GET("/api/v1/admin/catalog/albums/{id}/tracks", { params: { path: { id } } });
      setSubRows((data?.tracks ?? []).map((t) => ({ id: t.id, label: t.title, sub: `${Math.round((t.durationMs ?? 0) / 1000)}s` })));
    } else if (tab === "playlists") {
      const { data } = await client.GET("/api/v1/admin/catalog/playlists/{id}/tracks", { params: { path: { id } } });
      setSubRows(((data as { tracks?: { id: string; title: string; durationMs?: number }[] })?.tracks ?? []).map((t) => ({ id: t.id, label: t.title, sub: `${Math.round((t.durationMs ?? 0) / 1000)}s` })));
    }
    setSubLoading(false);
  }

  async function del(id: string) {
    if (!client || !window.confirm("Delete?")) return;
    if (tab === "artists") await client.DELETE("/api/v1/admin/catalog/artists/{id}", { params: { path: { id } } });
    else if (tab === "albums") await client.DELETE("/api/v1/admin/catalog/albums/{id}", { params: { path: { id } } });
    else if (tab === "tracks") await client.DELETE("/api/v1/admin/catalog/tracks/{id}", { params: { path: { id } } });
    else await client.DELETE("/api/v1/admin/catalog/playlists/{id}", { params: { path: { id } } });
    if (expandedId === id) setExpandedId(null);
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

  async function relink() {
    if (!client || !relinkId || !relinkMediaId.trim()) return;
    await client.POST("/api/v1/admin/catalog/tracks/{id}/relink", {
      params: { path: { id: relinkId } },
      body: { mediaObjectId: relinkMediaId.trim() },
    });
    setRelinkId(null); setRelinkMediaId("");
    await load();
  }

  async function addPlaylistTrack(playlistId: string) {
    if (!client || !addTrackId.trim()) return;
    await client.POST("/api/v1/admin/catalog/playlists/{id}/tracks", {
      params: { path: { id: playlistId } },
      body: { trackId: addTrackId.trim() },
    });
    setAddTrackId("");
    await expand(playlistId);
    await load();
  }

  async function removePlaylistTrack(playlistId: string, trackId: string) {
    if (!client) return;
    await client.DELETE("/api/v1/admin/catalog/playlists/{id}/tracks/{trackId}", {
      params: { path: { id: playlistId, trackId } },
    });
    await expand(playlistId);
    await load();
  }

  const totalPages = Math.ceil(total / PAGE);
  const page = Math.floor(offset / PAGE) + 1;
  const TABS: Tab[] = ["artists", "albums", "tracks", "playlists"];
  const hasExpand = tab === "artists" || tab === "albums" || tab === "playlists";

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
          <span className="flex-1">Name / Title</span><span className="w-24">ID</span><span className="w-24"></span>
        </div>
        {loading ? (
          <div className="px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">Loading…</div>
        ) : rows.map((r) => (
          <div key={r.id}>
            <div className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 hover:bg-[var(--color-surface-raised)] transition-colors">
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
                  <Btn onClick={save} title="Save"><Save size={13} /></Btn>
                )}
                {tab === "tracks" && (
                  <Btn onClick={() => { setRelinkId(r.id); setRelinkMediaId(""); }} title="Relink media object"><Link2 size={13} /></Btn>
                )}
                {hasExpand && (
                  <Btn onClick={() => expand(r.id)} title={expandedId === r.id ? "Collapse" : "Expand"}>
                    {expandedId === r.id ? <ChevronUp size={13} /> : <ChevronDown size={13} />}
                  </Btn>
                )}
                <Btn onClick={() => del(r.id)} title="Delete" danger><Trash2 size={13} /></Btn>
              </div>
            </div>

            {/* Sub-resource panel */}
            {expandedId === r.id && (
              <div className="border-b border-[var(--color-border)] bg-[var(--color-void)] px-6 py-3 space-y-2">
                {subLoading ? (
                  <p className="text-xs text-[var(--color-text-muted)]">Loading…</p>
                ) : subRows.length === 0 ? (
                  <p className="text-xs text-[var(--color-text-muted)]">No items</p>
                ) : (
                  <div className="divide-y divide-[var(--color-border)] rounded-lg border border-[var(--color-border)] overflow-hidden">
                    {subRows.map((s) => (
                      <div key={s.id} className="flex items-center gap-3 px-3 py-2 hover:bg-[var(--color-surface-raised)]">
                        <span className="flex-1 truncate text-xs text-[var(--color-text)]">{s.label}</span>
                        <span className="text-xs text-[var(--color-text-muted)]">{s.sub}</span>
                        {tab === "playlists" && (
                          <button onClick={() => removePlaylistTrack(r.id, s.id)} className="rounded p-1 text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"><X size={12} /></button>
                        )}
                      </div>
                    ))}
                  </div>
                )}
                {/* Playlist: add track */}
                {tab === "playlists" && (
                  <div className="flex items-center gap-2 pt-1">
                    <input
                      value={addTrackId}
                      onChange={(e) => setAddTrackId(e.target.value)}
                      placeholder="Track ID to add…"
                      className="flex-1 rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-2 py-1.5 text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] placeholder:text-[var(--color-text-muted)]"
                    />
                    <button onClick={() => addPlaylistTrack(r.id)} disabled={!addTrackId.trim()}
                      className="flex items-center gap-1 rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-xs font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-40">
                      <Plus size={12} /> Add
                    </button>
                  </div>
                )}
              </div>
            )}
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

      {/* Relink dialog */}
      {relinkId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={() => setRelinkId(null)}>
          <div className="w-full max-w-sm rounded-2xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6 space-y-4" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h2 className="font-semibold text-[var(--color-text)]">Relink media object</h2>
              <button onClick={() => setRelinkId(null)} className="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]"><X size={16} /></button>
            </div>
            <p className="text-xs text-[var(--color-text-muted)]">Track ID: <code className="text-[var(--color-text)]">{relinkId}</code></p>
            <input
              value={relinkMediaId}
              onChange={(e) => setRelinkMediaId(e.target.value)}
              placeholder="New media object ID…"
              className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] placeholder:text-[var(--color-text-muted)]"
            />
            <button onClick={relink} disabled={!relinkMediaId.trim()}
              className="w-full flex items-center justify-center gap-1.5 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-40">
              <Link2 size={14} /> Relink
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function Btn({ children, onClick, title, danger }: { children: React.ReactNode; onClick: () => void; title?: string; danger?: boolean }) {
  return (
    <button onClick={onClick} title={title}
      className={`rounded p-1.5 transition-colors ${danger ? "text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"}`}
    >{children}</button>
  );
}
