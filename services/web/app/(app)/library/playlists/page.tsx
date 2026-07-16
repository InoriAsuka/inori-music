/**
 * User playlists list — /library/playlists
 *
 * Consumes /api/v1/me/playlists: list, create (POST), rename (PATCH),
 * delete (DELETE with confirmation). Catalog playlists (server-managed) live
 * under /playlists and are unaffected — this is the personal-library scope.
 */
"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { ListMusic, Plus, Pencil, Trash2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { EmptyState } from "@/components/ui/EmptyState";
import { Skeleton } from "@/components/ui/Skeleton";
import { Modal } from "@/components/ui/Modal";
import {
  listUserPlaylists,
  createUserPlaylist,
  renameUserPlaylist,
  deleteUserPlaylist,
  type UserPlaylist,
} from "@/lib/api/me-playlists";

export default function UserPlaylistsPage() {
  const token = useAuthStore((s) => s.token);
  const [playlists, setPlaylists] = useState<UserPlaylist[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  // Create / rename dialog state
  const [createOpen, setCreateOpen] = useState(false);
  const [renaming, setRenaming] = useState<UserPlaylist | null>(null);
  const [nameInput, setNameInput] = useState("");
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError(false);
    try {
      setPlaylists(await listUserPlaylists(token));
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  function openCreate() {
    setNameInput("");
    setCreateOpen(true);
  }
  function openRename(pl: UserPlaylist) {
    setNameInput(pl.name);
    setRenaming(pl);
  }
  function closeDialogs() {
    setCreateOpen(false);
    setRenaming(null);
    setNameInput("");
  }

  async function submitCreate(e: React.FormEvent) {
    e.preventDefault();
    const name = nameInput.trim();
    if (!name || !token || saving) return;
    setSaving(true);
    try {
      const created = await createUserPlaylist(token, { name });
      setPlaylists((prev) => [created, ...prev]);
      closeDialogs();
    } catch {
      setError(true);
    } finally {
      setSaving(false);
    }
  }

  async function submitRename(e: React.FormEvent) {
    e.preventDefault();
    const name = nameInput.trim();
    if (!name || !token || !renaming || saving) return;
    setSaving(true);
    try {
      const updated = await renameUserPlaylist(token, renaming.id, name);
      setPlaylists((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
      closeDialogs();
    } catch {
      setError(true);
    } finally {
      setSaving(false);
    }
  }

  async function remove(pl: UserPlaylist) {
    if (!token) return;
    if (!window.confirm(`Delete playlist "${pl.name}"? This cannot be undone.`)) return;
    // Optimistic removal with rollback on failure.
    const prev = playlists;
    setPlaylists((cur) => cur.filter((p) => p.id !== pl.id));
    try {
      await deleteUserPlaylist(token, pl.id);
    } catch {
      setPlaylists(prev);
      setError(true);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <ListMusic size={22} className="text-[var(--color-primary)]" />
          <h1 className="text-2xl font-bold">My Playlists</h1>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="flex items-center gap-2 rounded-full bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-foreground)] hover:opacity-90 transition-opacity"
        >
          <Plus size={15} /> New
        </button>
      </div>

      {error && !loading && (
        <div className="flex items-center justify-between rounded-lg border border-[var(--color-destructive)]/40 bg-[var(--color-destructive)]/10 px-4 py-3 text-sm text-[var(--color-destructive)]">
          <span>Couldn’t load your playlists.</span>
          <button type="button" onClick={load} className="font-medium underline">
            Retry
          </button>
        </div>
      )}

      {loading ? (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 last:border-0">
              <Skeleton className="h-8 w-8 rounded" />
              <Skeleton className="h-4 w-40" />
            </div>
          ))}
        </div>
      ) : playlists.length === 0 && !error ? (
        <EmptyState
          title="No playlists yet"
          description="Create a playlist to start collecting your favorite tracks."
        />
      ) : (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
          {playlists.map((pl) => (
            <div
              key={pl.id}
              className="group flex items-center gap-3 border-b border-[var(--color-border)] px-4 py-3 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
            >
              <Link href={`/library/playlists/${pl.id}`} className="flex min-w-0 flex-1 items-center gap-3">
                <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-[var(--color-muted)] text-[var(--color-muted-foreground)]">
                  <ListMusic size={16} />
                </span>
                <span className="min-w-0">
                  <span className="block truncate text-sm font-medium">{pl.name}</span>
                  <span className="block text-xs text-[var(--color-muted-foreground)]">
                    {pl.trackIds.length} {pl.trackIds.length === 1 ? "track" : "tracks"}
                  </span>
                </span>
              </Link>
              <button
                type="button"
                onClick={() => openRename(pl)}
                aria-label={`Rename ${pl.name}`}
                className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
              >
                <Pencil size={15} />
              </button>
              <button
                type="button"
                onClick={() => remove(pl)}
                aria-label={`Delete ${pl.name}`}
                className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:text-[var(--color-destructive)]"
              >
                <Trash2 size={15} />
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Create dialog */}
      <Modal open={createOpen} onClose={closeDialogs} title="New playlist">
        <PlaylistNameForm
          value={nameInput}
          onChange={setNameInput}
          onSubmit={submitCreate}
          submitLabel="Create"
          saving={saving}
        />
      </Modal>

      {/* Rename dialog */}
      <Modal open={renaming != null} onClose={closeDialogs} title="Rename playlist">
        <PlaylistNameForm
          value={nameInput}
          onChange={setNameInput}
          onSubmit={submitRename}
          submitLabel="Save"
          saving={saving}
        />
      </Modal>
    </div>
  );
}

function PlaylistNameForm({
  value,
  onChange,
  onSubmit,
  submitLabel,
  saving,
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit: (e: React.FormEvent) => void;
  submitLabel: string;
  saving: boolean;
}) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Playlist name"
        aria-label="Playlist name"
        className="w-full rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
      />
      <div className="flex justify-end">
        <button
          type="submit"
          disabled={!value.trim() || saving}
          className="rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-[var(--color-primary-foreground)] disabled:opacity-50"
        >
          {submitLabel}
        </button>
      </div>
    </form>
  );
}
