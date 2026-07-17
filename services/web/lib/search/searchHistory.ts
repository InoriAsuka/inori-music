"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { deleteRemoteSearchHistory, getRemoteSearchHistory, putRemoteSearchHistory } from "@/lib/api/search-history";

const KEY = "inori.search.history";
const MAX_ENTRIES = 20;
/** Quiet period before a mutated history is pushed to the server. */
const REMOTE_SYNC_DEBOUNCE_MS = 1500;

/** Adds `query` to the front of `entries`, deduped and capped at MAX_ENTRIES. Trims whitespace; ignores blank queries. */
export function addEntry(entries: string[], query: string): string[] {
  const trimmed = query.trim();
  if (!trimmed) return entries;
  return [trimmed, ...entries.filter((q) => q !== trimmed)].slice(0, MAX_ENTRIES);
}

/** Removes `query` from `entries`. */
export function removeEntry(entries: string[], query: string): string[] {
  return entries.filter((q) => q !== query);
}

/**
 * Merge two newest-first history lists (v5.4.0 cross-device sync), deduped and
 * capped at MAX_ENTRIES. `local` wins ties (kept ahead of the same query from
 * `remote`) so a query the user just issued on this device stays at the front,
 * but every remote-only query is preserved in remote order behind the local
 * ones. Whitespace is trimmed and blanks dropped, matching `addEntry`.
 *
 * Interleaving by true timestamp isn't possible — the local list carries no
 * per-entry times — so this is a stable "local-first, then remote" union, which
 * keeps the just-typed query prominent while still surfacing history from other
 * devices. The result is the list PUT back to the server as the new canonical
 * order.
 */
export function mergeHistories(local: string[], remote: string[]): string[] {
  const seen = new Set<string>();
  const merged: string[] = [];
  for (const q of [...local, ...remote]) {
    const trimmed = q.trim();
    if (!trimmed || seen.has(trimmed)) continue;
    seen.add(trimmed);
    merged.push(trimmed);
    if (merged.length >= MAX_ENTRIES) break;
  }
  return merged;
}

function read(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((v): v is string => typeof v === "string") : [];
  } catch {
    return [];
  }
}

function write(entries: string[]) {
  if (typeof window === "undefined") return;
  localStorage.setItem(KEY, JSON.stringify(entries));
}

/**
 * Persists recent search queries locally, most-recent-first, deduped,
 * capped at MAX_ENTRIES. Mirrors the Flutter SearchHistoryNotifier pattern.
 *
 * When a viewer `token` is supplied (v5.4.0 cross-device sync), the hook also:
 *  - on mount/login: GETs the remote history, merges it with local (dedup,
 *    newest-first, cap 20), writes the union locally, and PUTs it back so both
 *    sides converge;
 *  - on add/remove: debounces a PUT of the new local list to the server;
 *  - on clear: DELETEs the remote history.
 *
 * All remote calls are best-effort — a network failure never blocks the local
 * experience (offline stays local; the next mutation re-syncs). Passing no
 * token keeps the pre-v5.4.0 local-only behaviour (used before login).
 */
export function useSearchHistory(token?: string | null) {
  const [history, setHistory] = useState<string[]>([]);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // Latest token in a ref so debounced/settled callbacks never fire against a
  // stale (logged-out) token.
  const tokenRef = useRef<string | null>(token ?? null);
  tokenRef.current = token ?? null;

  // Restore local, then merge with remote once a token is available.
  useEffect(() => {
    const local = read();
    setHistory(local);
    if (!token) return;
    let cancelled = false;
    getRemoteSearchHistory(token)
      .then((remote) => {
        if (cancelled) return;
        const merged = mergeHistories(local, remote);
        // Only rewrite/push when the merge actually changed the local list, to
        // avoid a redundant PUT when both sides already agree.
        write(merged);
        setHistory(merged);
        const changedRemote = merged.length !== remote.length || merged.some((q, i) => q !== remote[i]);
        if (changedRemote && tokenRef.current) void putRemoteSearchHistory(tokenRef.current, merged).catch(() => {});
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [token]);

  /** Debounced replace-all push of the current local list to the server. */
  const scheduleRemotePut = useCallback((entries: string[]) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      const t = tokenRef.current;
      if (t) void putRemoteSearchHistory(t, entries).catch(() => {});
    }, REMOTE_SYNC_DEBOUNCE_MS);
  }, []);

  // Flush any pending PUT timer on unmount.
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  const add = useCallback(
    (query: string) => {
      setHistory((prev) => {
        const next = addEntry(prev, query);
        if (next === prev) return prev;
        write(next);
        scheduleRemotePut(next);
        return next;
      });
    },
    [scheduleRemotePut]
  );

  const remove = useCallback(
    (query: string) => {
      setHistory((prev) => {
        const next = removeEntry(prev, query);
        write(next);
        scheduleRemotePut(next);
        return next;
      });
    },
    [scheduleRemotePut]
  );

  const clear = useCallback(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    setHistory([]);
    if (typeof window !== "undefined") localStorage.removeItem(KEY);
    const t = tokenRef.current;
    if (t) void deleteRemoteSearchHistory(t).catch(() => {});
  }, []);

  return { history, add, remove, clear };
}
