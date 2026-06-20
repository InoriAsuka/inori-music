/**
 * Admin catalog import — /admin/import
 * Single-track import and batch JSON import.
 * CatalogImportRequest: { albumId?, artistId?, discNumber?, durationMs?, mediaObjectId, sortTitle?, title, trackNumber? }
 * CatalogBatchImportRequest: { items: CatalogImportRequest[] }
 * CatalogBatchImportResult: { total, imported, failed, items }
 */
"use client";

import { FormEvent, useState } from "react";
import { Upload, Loader2 } from "lucide-react";
import { AdminTokenPanel } from "@/components/admin/AdminTokenPanel";
import { useAdminApi, useHasAdminAccess } from "@/hooks/useAdminApi";
import { EmptyState } from "@/components/ui/EmptyState";

export default function AdminImportPage() {
  const admin = useAdminApi();
  const hasAccess = useHasAdminAccess();
  const [tab, setTab] = useState<"single" | "batch">("single");

  // Single form
  const [single, setSingle] = useState({ title: "", mediaObjectId: "", artistId: "", albumId: "", trackNumber: "" });
  const [singleLoading, setSingleLoading] = useState(false);
  const [singleResult, setSingleResult] = useState<string | null>(null);
  const [singleError, setSingleError] = useState<string | null>(null);

  // Batch form
  const [batchJson, setBatchJson] = useState("");
  const [batchLoading, setBatchLoading] = useState(false);
  const [batchResult, setBatchResult] = useState<{ total: number; imported: number; failed: number } | null>(null);
  const [batchError, setBatchError] = useState<string | null>(null);

  async function handleSingle(e: FormEvent) {
    e.preventDefault();
    if (!admin) return;
    setSingleLoading(true); setSingleError(null); setSingleResult(null);
    const body = {
      title: single.title,
      mediaObjectId: single.mediaObjectId,
      ...(single.artistId && { artistId: single.artistId }),
      ...(single.albumId && { albumId: single.albumId }),
      ...(single.trackNumber && { trackNumber: parseInt(single.trackNumber) }),
    };
    const { data, error } = await admin.POST("/api/v1/admin/catalog/import", { body });
    if (error) setSingleError("Import failed.");
    else if (data) setSingleResult(`Imported: ${data.id} — "${data.title}"`);
    setSingleLoading(false);
  }

  async function handleBatch(e: FormEvent) {
    e.preventDefault();
    if (!admin) return;
    setBatchLoading(true); setBatchError(null); setBatchResult(null);
    let items;
    try {
      const parsed = JSON.parse(batchJson);
      items = Array.isArray(parsed) ? parsed : parsed.items;
    } catch { setBatchError("Invalid JSON."); setBatchLoading(false); return; }
    const { data, error } = await admin.POST("/api/v1/admin/catalog/batch-import", { body: { items } });
    if (error) setBatchError("Batch import failed.");
    else if (data) setBatchResult({ total: data.total, imported: data.imported, failed: data.failed });
    setBatchLoading(false);
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Import</h1>
      <AdminTokenPanel />
      {!hasAccess && <EmptyState title="Admin access required" description="Sign in as an admin or paste a bootstrap token." />}
      {hasAccess && (
        <>
          <div className="flex gap-2">
            {(["single", "batch"] as const).map((t) => (
              <button key={t} onClick={() => setTab(t)} className={tab === t ? "rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-sm font-semibold text-[var(--color-primary-foreground)]" : "rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm hover:bg-[var(--color-muted)]"}>
                {t === "single" ? "Single track" : "Batch JSON"}
              </button>
            ))}
          </div>

          {tab === "single" && (
            <form onSubmit={handleSingle} className="max-w-lg space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-6">
              {[
                { id: "title", label: "Title *", placeholder: "Track title", required: true },
                { id: "mediaObjectId", label: "Media Object ID *", placeholder: "UUID", required: true },
                { id: "artistId", label: "Artist ID", placeholder: "(optional UUID)", required: false },
                { id: "albumId", label: "Album ID", placeholder: "(optional UUID)", required: false },
                { id: "trackNumber", label: "Track #", placeholder: "(optional number)", required: false },
              ].map(({ id, label, placeholder, required }) => (
                <div key={id} className="space-y-1.5">
                  <label className="text-sm font-medium">{label}</label>
                  <input value={single[id as keyof typeof single]} onChange={(e) => setSingle((s) => ({ ...s, [id]: e.target.value }))} required={required} placeholder={placeholder} className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]" />
                </div>
              ))}
              {singleError && <p className="text-sm text-[var(--color-destructive)]">{singleError}</p>}
              {singleResult && <p className="text-sm text-green-600">{singleResult}</p>}
              <button type="submit" disabled={singleLoading} className="flex items-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-60">
                {singleLoading ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />} Import track
              </button>
            </form>
          )}

          {tab === "batch" && (
            <form onSubmit={handleBatch} className="max-w-2xl space-y-4 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-6">
              <div className="space-y-1.5">
                <label className="text-sm font-medium">JSON items array (or <code>{"{\"items\":[...]}"}</code>)</label>
                <textarea
                  value={batchJson}
                  onChange={(e) => setBatchJson(e.target.value)}
                  rows={10}
                  required
                  className="w-full font-mono rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 text-xs outline-none focus:border-[var(--color-primary)]"
                  placeholder={JSON.stringify([{ title: "Track 1", mediaObjectId: "uuid" }], null, 2)}
                />
              </div>
              {batchError && <p className="text-sm text-[var(--color-destructive)]">{batchError}</p>}
              {batchResult && (
                <div className="rounded-md bg-[var(--color-muted)] px-4 py-3 text-sm">
                  Total: {batchResult.total} · Imported: <span className="text-green-600 font-semibold">{batchResult.imported}</span> · Failed: <span className="text-[var(--color-destructive)] font-semibold">{batchResult.failed}</span>
                </div>
              )}
              <button type="submit" disabled={batchLoading} className="flex items-center gap-2 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 disabled:opacity-60">
                {batchLoading ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />} Batch import
              </button>
            </form>
          )}
        </>
      )}
    </div>
  );
}
