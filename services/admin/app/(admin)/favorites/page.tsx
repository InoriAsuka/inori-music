"use client";

/**
 * Admin favorites management — /admin/favorites
 *
 * Allows admins to view and manage per-user favorite tracks.
 * Enter a userId to list their favorites, then remove individual tracks
 * or clear all favorites for that user.
 *
 * Endpoints consumed:
 *   GET    /api/v1/admin/favorites/users/{userId}/tracks
 *   DELETE /api/v1/admin/favorites/users/{userId}/tracks
 *   DELETE /api/v1/admin/favorites/users/{userId}/tracks/{trackId}
 */

import { useState } from "react";
import { Heart, Trash2, Search, X } from "lucide-react";
import { useAdminClient } from "@/hooks/useAdminClient";

interface FavoriteTrack {
  id: string;
  title: string;
}

export default function FavoritesPage() {
  const client = useAdminClient();
  const [userId, setUserId] = useState("");
  const [inputValue, setInputValue] = useState("");
  const [tracks, setTracks] = useState<FavoriteTrack[]>([]);
  const [loading, setLoading] = useState(false);
  const [clearing, setClearing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function lookup() {
    const uid = inputValue.trim();
    if (!uid || !client) return;
    setLoading(true);
    setError(null);
    setTracks([]);
    setUserId(uid);

    const { data, error: apiErr } = await client.GET("/api/v1/admin/favorites/users/{userId}/tracks", {
      params: { path: { userId: uid } },
    });

    if (apiErr || !data) {
      setError("Failed to load favorites — check the user ID.");
    } else {
      const items = (data as { tracks?: { id: string; title?: string }[] }).tracks ?? [];
      setTracks(items.map((t) => ({ id: t.id, title: t.title ?? t.id })));
    }
    setLoading(false);
  }

  async function removeTrack(trackId: string) {
    if (!client || !userId) return;
    await client.DELETE("/api/v1/admin/favorites/users/{userId}/tracks/{trackId}", {
      params: { path: { userId, trackId } },
    });
    setTracks((prev) => prev.filter((t) => t.id !== trackId));
  }

  async function clearAll() {
    if (!client || !userId || !window.confirm(`Clear all favorites for user ${userId}?`)) return;
    setClearing(true);
    await client.DELETE("/api/v1/admin/favorites/users/{userId}/tracks", {
      params: { path: { userId } },
    });
    setTracks([]);
    setClearing(false);
  }

  return (
    <div className="space-y-6">
      <h1 className="font-display text-xl font-bold tracking-wider text-[var(--color-primary)]">FAVORITES</h1>

      {/* Lookup form */}
      <div className="flex gap-3 rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-4">
        <input
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && lookup()}
          placeholder="User ID…"
          className="flex-1 rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-3 py-2 text-sm text-[var(--color-text)] outline-none focus:border-[var(--color-primary)] placeholder:text-[var(--color-text-muted)]"
        />
        <button
          type="button"
          onClick={lookup}
          disabled={loading || !inputValue.trim()}
          className="flex items-center gap-1.5 rounded-md bg-[var(--color-primary)] px-4 py-2 text-sm font-semibold text-[var(--color-primary-fg)] hover:opacity-90 disabled:opacity-40"
        >
          <Search size={14} />
          Look up
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="flex items-center gap-2 rounded-xl border border-[var(--color-danger)]/40 bg-[var(--color-danger)]/5 px-4 py-3 text-sm text-[var(--color-danger)]">
          <X size={14} /> {error}
        </div>
      )}

      {/* Results */}
      {userId && !loading && !error && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Heart size={16} className="text-[var(--color-primary)]" />
              <p className="text-sm font-medium text-[var(--color-text)]">
                {tracks.length} favorite{tracks.length !== 1 ? "s" : ""} for{" "}
                <code className="text-[var(--color-text-muted)]">{userId}</code>
              </p>
            </div>
            {tracks.length > 0 && (
              <button
                type="button"
                onClick={clearAll}
                disabled={clearing}
                className="flex items-center gap-1.5 rounded-md border border-[var(--color-danger)]/40 px-3 py-1.5 text-xs text-[var(--color-danger)] hover:bg-[var(--color-danger)]/10 disabled:opacity-50 transition-colors"
              >
                <Trash2 size={12} />
                {clearing ? "Clearing…" : "Clear all"}
              </button>
            )}
          </div>

          {tracks.length === 0 ? (
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
              No favorites for this user.
            </div>
          ) : (
            <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] divide-y divide-[var(--color-border)]">
              {tracks.map((t) => (
                <div
                  key={t.id}
                  className="flex items-center gap-3 px-4 py-3 hover:bg-[var(--color-surface-raised)] transition-colors"
                >
                  <Heart size={14} className="shrink-0 text-[var(--color-primary)]" fill="currentColor" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-[var(--color-text)]">{t.title}</p>
                    <p className="font-mono text-xs text-[var(--color-text-muted)]">{t.id}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeTrack(t.id)}
                    className="shrink-0 rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-danger)] transition-colors"
                    title="Remove favorite"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {loading && (
        <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]">
          Loading…
        </div>
      )}
    </div>
  );
}
