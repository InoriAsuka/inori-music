"use client";

/**
 * CopyFromCatalogDialog — copy a catalog playlist into the viewer's personal library.
 *
 * The button is rendered in the catalog playlist detail header. The user
 * confirms the target name (defaults to the source name) and the operation
 * copies all tracks in bulk via POST /api/v1/catalog/playlists/{id}/copy,
 * which persists the `sourceCatalogId` on the user-side record so the link
 * is reversible and future sync is possible.
 *
 * Error paths are explicit — the dialog shows a retry button rather than
 * silently failing, because the caller (catalog detail page) can then
 * close the dialog and continue browsing.
 */
import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { CheckCircle2, X } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import { copyCatalogPlaylist, type UserPlaylist } from "@/lib/api/me-playlists";
import { cn } from "@/lib/utils";

export interface CopyFromCatalogDialogProps {
  open: boolean;
  onClose: () => void;
  catalogId: string;
  catalogName: string;
}

export function CopyFromCatalogDialog({
  open,
  onClose,
  catalogId,
  catalogName,
}: CopyFromCatalogDialogProps) {
  const token =
    typeof window !== "undefined"
      ? localStorage.getItem("token")
      : null;

  const [targetName, setTargetName] = useState(catalogName);
  const [copying, setCopying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<UserPlaylist | null>(null);
  const nameRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (open) {
      setTargetName(catalogName);
      setError(null);
      setResult(null);
      setCopying(false);
      requestAnimationFrame(() => nameRef.current?.focus());
    }
  }, [open, catalogName]);

  async function handleCopy(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !targetName.trim() || copying) return;
    setCopying(true);
    setError(null);
    try {
      const playlist = await copyCatalogPlaylist(token, catalogId);
      setResult(playlist);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Couldn't copy the playlist. Please try again."
      );
    } finally {
      setCopying(false);
    }
  }

  return createPortal(
    <Modal open={open} onClose={onClose} title="Copy to my library">
      <form onSubmit={handleCopy} className="space-y-4">
        <div>
          <label
            htmlFor="copy-target-name"
            className="mb-1 block text-xs font-medium text-[var(--color-text-secondary)]"
          >
            Playlist name
          </label>
          <input
            ref={nameRef}
            id="copy-target-name"
            type="text"
            value={targetName}
            onChange={(e) => setTargetName(e.target.value)}
            className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
            maxLength={120}
          />
        </div>

        {error && (
          <p className="text-xs text-[var(--color-danger)]">{error}</p>
        )}

        {result && (
          <div className="flex items-center gap-2 text-xs text-[var(--color-success)]">
            <CheckCircle2 size={14} />
            <span>
              Copied {result.trackIds?.length ?? 0} tracks to &quot;{result.name}&quot;
            </span>
          </div>
        )}

        <div className="flex justify-end gap-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-3 py-1.5 text-sm text-[var(--color-text-muted)] hover:bg-[var(--color-muted)] hover:text-[var(--color-text)]"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={!targetName.trim() || copying}
            className="rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-sm font-medium text-[var(--color-primary-foreground)] disabled:opacity-50"
          >
            {copying ? "Copying…" : "Copy"}
          </button>
        </div>
      </form>
    </Modal>,
    document.body
  );
}
