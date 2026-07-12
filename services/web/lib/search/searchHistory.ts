"use client";

import { useCallback, useEffect, useState } from "react";

const KEY = "inori.search.history";
const MAX_ENTRIES = 20;

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
 */
export function useSearchHistory() {
  const [history, setHistory] = useState<string[]>([]);

  useEffect(() => {
    setHistory(read());
  }, []);

  const add = useCallback((query: string) => {
    setHistory((prev) => {
      const next = addEntry(prev, query);
      write(next);
      return next;
    });
  }, []);

  const remove = useCallback((query: string) => {
    setHistory((prev) => {
      const next = removeEntry(prev, query);
      write(next);
      return next;
    });
  }, []);

  const clear = useCallback(() => {
    setHistory([]);
    if (typeof window !== "undefined") localStorage.removeItem(KEY);
  }, []);

  return { history, add, remove, clear };
}
