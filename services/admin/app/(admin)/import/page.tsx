"use client";

import { FormEvent, useState } from "react";
import { Upload, Loader2 } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";

export default function ImportPage() {
  const client = useAdminClient();
  const [tab, setTab] = useState<"single" | "batch">("single");
  const [single, setSingle] = useState({ title: "", mediaObjectId: "", artistId: "", albumId: "", trackNumber: "" });
  const [batchJson, setBatchJson] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function submitSingle(e: FormEvent) {
    e.preventDefault(); if (!client) return;
    setLoading(true); setError(null); setResult(null);
    const body = {
      title: single.title, mediaObjectId: single.mediaObjectId,
      ...(single.artistId && { artistId: single.artistId }),
      ...(single.albumId && { albumId: single.albumId }),
      ...(single.trackNumber && { trackNumber: parseInt(single.trackNumber) }),
    };
    const { data, error: e2 } = await client.POST("/api/v1/admin/catalog/import", { body });
    if (e2) setError("Import failed.");
    else if (data) setResult(`✓ Imported: ${data.id} — "${data.title}"`);
    setLoading(false);
  }

  async function submitBatch(e: FormEvent) {
    e.preventDefault(); if (!client) return;
    setLoading(true); setError(null); setResult(null);
    let items;
    try { const p = JSON.parse(batchJson); items = Array.isArray(p) ? p : p.items; }
    catch { setError("Invalid JSON."); setLoading(false); return; }
    const { data, error: e2 } = await client.POST("/api/v1/admin/catalog/batch-import", { body: { items } });
    if (e2) setError("Batch import failed.");
    else if (data) setResult(`✓ ${data.imported}/${data.total} imported, ${data.failed} failed`);
    setLoading(false);
  }

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">IMPORT</h1>
      <div className="flex gap-1 rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-1 w-fit">
        {(["single", "batch"] as const).map((t) => (
          <button key={t} onClick={() => { setTab(t); setError(null); setResult(null); }}
            className={tab === t ? "rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-xs font-semibold text-[var(--color-primary-fg)]"
              : "rounded-md px-3 py-1.5 text-xs text-[var(--color-text-secondary)] hover:text-[var(--color-text)]"}
          >{t === "single" ? "Single track" : "Batch JSON"}</button>
        ))}
      </div>

      {tab === "single" && (
        <form onSubmit={submitSingle} className="max-w-lg space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6">
          {[
            { key: "title", label: "Title *", req: true }, { key: "mediaObjectId", label: "Media Object ID *", req: true },
            { key: "artistId", label: "Artist ID" }, { key: "albumId", label: "Album ID" }, { key: "trackNumber", label: "Track #" },
          ].map(({ key, label, req }) => (
            <div key={key} className="space-y-1">
              <label className="text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">{label}</label>
              <input type="text" required={req} value={single[key as keyof typeof single]} onChange={(e) => setSingle((s) => ({ ...s, [key]: e.target.value }))}
                className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] placeholder:text-[var(--color-text-muted)]"
              />
            </div>
          ))}
          {error && <p className="text-sm text-[var(--color-danger)]">{error}</p>}
          {result && <p className="text-sm text-[var(--color-success)]">{result}</p>}
          <button type="submit" disabled={loading} className="flex items-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-60">
            {loading ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />} Import
          </button>
        </form>
      )}

      {tab === "batch" && (
        <form onSubmit={submitBatch} className="max-w-2xl space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6">
          <div className="space-y-1">
            <label className="text-xs font-medium uppercase tracking-wider text-[var(--color-text-muted)]">JSON array or {`{"items": [...]}`}</label>
            <textarea value={batchJson} onChange={(e) => setBatchJson(e.target.value)} rows={10} required
              className="w-full font-mono rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
              placeholder={JSON.stringify([{ title: "Track 1", mediaObjectId: "uuid" }], null, 2)}
            />
          </div>
          {error && <p className="text-sm text-[var(--color-danger)]">{error}</p>}
          {result && <p className="rounded-md border border-[var(--color-success)] bg-[var(--color-success)] bg-opacity-10 px-4 py-3 text-sm text-[var(--color-success)]">{result}</p>}
          <button type="submit" disabled={loading} className="flex items-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-60">
            {loading ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />} Batch import
          </button>
        </form>
      )}
    </div>
  );
}
