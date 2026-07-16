"use client";

/**
 * AddToPlaylistDialog — reusable "add this track to one of my playlists" flow.
 *
 * Reached from TrackRow's overflow menu (see TrackRow). Lists the viewer's
 * personal playlists, supports inline create, then appends the given track via
 * POST /me/playlists/{id}/tracks. Self-contained: fetches its own list on open
 * so any track list can surface it without prop-drilling playlist state.
 *
 * The trigger only renders where an auth token is available; the dialog itself
 * assumes a token (passed in) and never guesses backend fields.
 */
import { useCallback, useEffect, useState } from "react";
import { Plus, Check, ListMusic } from "lucide-react";
import { Modal } from "@/components/ui/Modal";
import {
  listUserPlaylists,
  createUserPlaylist,
  appendUserPlaylistTrack,
  type UserPlaylist,
} from "@/lib/api/me-playlists";

interface AddToPlaylistDialogProps {
  open: boolean;
  onClose: () => void;
  token: string;
  trackId: string;
  trackTitle: string;
}

export function AddToPlaylistDialog({ open, onClose, token, trackId, trackTitle }: AddToPlaylistDialogProps) {
  const [playlists, setPlaylists] = useState<UserPlaylist[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [addedIds, setAddedIds] = useState<Set<string>>(() => new Set());
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");

  const loadPlaylists = useCallback(async () => {
    setLoading(true);
    setLoadError(false);
    try {
      setPlaylists(await listUserPlaylists(token));
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    if (!open) return;
    // Reset transient UI each time the dialog is opened for a track.
    setAddedIds(new Set());
    setCreating(false);
    setNewName("");
    setActionError(null);
    void loadPlaylists();
  }, [open, loadPlaylists]);

  async function addTo(playlist: UserPlaylist) {
    if (busyId || addedIds.has(playlist.id)) return;
    setBusyId(playlist.id);
    setActionError(null);
    try {
      await appendUserPlaylistTrack(token, playlist.id, trackId);
      setAddedIds((prev) => new Set(prev).add(playlist.id));
    } catch {
      setActionError(`Couldn’t add the track to "${playlist.name}". Please try again.`);
    } finally {
      setBusyId(null);
    }
  }

  async function createAndAdd(e: React.FormEvent) {
    e.preventDefault();
    const name = newName.trim();
    if (!name || busyId) return;
    setBusyId("__new__");
    setActionError(null);
    let created: UserPlaylist;
    try {
      created = await createUserPlaylist(token, { name });
    } catch {
      setActionError("Couldn’t create the playlist. Please try again.");
      setBusyId(null);
      return;
    }

    setPlaylists((prev) => [created, ...prev.filter((playlist) => playlist.id !== created.id)]);
    setCreating(false);
    setNewName("");
    setBusyId(created.id);
    try {
      await appendUserPlaylistTrack(token, created.id, trackId);
      setAddedIds((prev) => new Set(prev).add(created.id));
    } catch {
      setActionError(`Playlist "${created.name}" was created, but the track wasn’t added. Try adding it below.`);
    } finally {
      setBusyId(null);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title="Add to playlist">
      <p className="mb-3 truncate text-xs text-[var(--color-muted-foreground)]">{trackTitle}</p>

      {actionError && (
        <p className="mb-3 rounded-md bg-[var(--color-destructive)]/10 px-3 py-2 text-xs text-[var(--color-destructive)]">
          {actionError}
        </p>
      )}

      {loading ? (
        <p className="py-6 text-center text-sm text-[var(--color-muted-foreground)]">Loading playlists…</p>
      ) : loadError ? (
        <div className="py-6 text-center text-sm">
          <p className="text-[var(--color-destructive)]">Couldn’t load your playlists.</p>
          <button type="button" onClick={() => void loadPlaylists()} className="mt-2 font-medium text-[var(--color-primary)] underline">
            Retry
          </button>
        </div>
      ) : (
        <>
          <ul className="max-h-64 space-y-1 overflow-y-auto">
            {playlists.length === 0 && !creating && (
              <li className="py-4 text-center text-sm text-[var(--color-muted-foreground)]">
                No playlists yet — create one below.
              </li>
            )}
            {playlists.map((pl) => {
              const added = addedIds.has(pl.id);
              return (
                <li key={pl.id}>
                  <button
                    type="button"
                    onClick={() => addTo(pl)}
                    disabled={busyId != null || added}
                    className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm hover:bg-[var(--color-muted)] disabled:opacity-60"
                  >
                    <ListMusic size={15} className="shrink-0 text-[var(--color-muted-foreground)]" />
                    <span className="min-w-0 flex-1 truncate">{pl.name}</span>
                    {added ? (
                      <Check size={15} className="shrink-0 text-[var(--color-primary)]" />
                    ) : (
                      busyId === pl.id && (
                        <span className="shrink-0 text-xs text-[var(--color-muted-foreground)]">…</span>
                      )
                    )}
                  </button>
                </li>
              );
            })}
          </ul>

          <div className="mt-3 border-t border-[var(--color-border)] pt-3">
            {creating ? (
              <form onSubmit={createAndAdd} className="flex gap-2">
                <input
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  placeholder="New playlist name"
                  aria-label="New playlist name"
                  className="min-w-0 flex-1 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-1.5 text-sm outline-none focus:border-[var(--color-primary)]"
                />
                <button
                  type="submit"
                  disabled={!newName.trim() || busyId != null}
                  className="rounded-md bg-[var(--color-primary)] px-3 py-1.5 text-sm font-medium text-[var(--color-primary-foreground)] disabled:opacity-50"
                >
                  Create
                </button>
              </form>
            ) : (
              <button
                type="button"
                onClick={() => setCreating(true)}
                className="flex items-center gap-2 rounded-md px-3 py-2 text-sm text-[var(--color-primary)] hover:bg-[var(--color-muted)]"
              >
                <Plus size={15} /> New playlist
              </button>
            )}
          </div>
        </>
      )}
    </Modal>
  );
}
