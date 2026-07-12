/**
 * Search page — /search
 * CatalogSearchResult: { query, items: SearchResultItem[] }
 * SearchResultItem: { kind, artist?, album?, track? }
 * kind ∈ "artist" | "album" | "track" (no "playlist" per spec)
 * API params: { q, types, limit } (no "playlist" type in spec)
 */
"use client";

import { useState, useEffect, useRef, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Search, Loader2, History, X, Trash2 } from "lucide-react";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { usePlayerStore } from "@/store/player";
import { Artwork } from "@/components/ui/Artwork";
import { Highlighted } from "@/components/ui/Highlighted";
import { useSearchHistory } from "@/lib/search/searchHistory";
import { formatDuration, cn } from "@/lib/utils";

interface Artist {
  id: string;
  name: string;
  highlight?: string | null;
}
interface Album {
  id: string;
  title: string;
  highlight?: string | null;
}
interface Track {
  id: string;
  title: string;
  durationMs: number;
  highlight?: string | null;
}

interface SearchResults {
  artists: Artist[];
  albums: Album[];
  tracks: Track[];
}

function SearchPageInner() {
  const router = useRouter();
  const params = useSearchParams();
  const token = useAuthStore((s) => s.token);
  const playQueue = usePlayerStore((s) => s.playQueue);
  const { history, add: addHistory, remove: removeHistory, clear: clearHistory } = useSearchHistory();

  const initialQ = params.get("q") ?? "";
  const [query, setQuery] = useState(initialQ);
  const [results, setResults] = useState<SearchResults | null>(null);
  const [loading, setLoading] = useState(false);
  const [focused, setFocused] = useState(false);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);

  const visibleHistory = history.slice(0, 10);
  const showHistory = focused && !query.trim() && visibleHistory.length > 0;

  useEffect(() => {
    setHistoryIndex(-1);
  }, [showHistory]);

  function selectHistoryQuery(q: string) {
    setQuery(q);
    setFocused(false);
    inputRef.current?.blur();
  }

  function runSearch(q: string) {
    if (!token || !q.trim()) {
      setResults(null);
      return;
    }
    setLoading(true);
    authedApi(token)
      .GET("/api/v1/catalog/search", {
        params: { query: { q, types: "artist,album,track", limit: 8 } },
      })
      .then(({ data }) => {
        if (data?.items) {
          const artists: Artist[] = [];
          const albums: Album[] = [];
          const tracks: Track[] = [];
          for (const item of data.items) {
            if (item.kind === "artist" && item.artist) {
              artists.push({ id: item.artist.id, name: item.artist.name, highlight: item.highlight });
            } else if (item.kind === "album" && item.album) {
              albums.push({ id: item.album.id, title: item.album.title, highlight: item.highlight });
            } else if (item.kind === "track" && item.track) {
              tracks.push({
                id: item.track.id,
                title: item.track.title,
                durationMs: item.track.durationMs ?? 0,
                highlight: item.highlight,
              });
            }
          }
          setResults({ artists, albums, tracks });
        }
      })
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      if (query.trim()) {
        router.replace(`/search?q=${encodeURIComponent(query.trim())}`, { scroll: false });
        runSearch(query);
        addHistory(query);
      } else {
        router.replace("/search", { scroll: false });
        setResults(null);
      }
    }, 300);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query]);

  const empty = results && !results.artists.length && !results.albums.length && !results.tracks.length;

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Search input */}
      <div className="relative">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-muted-foreground)]" />
        {loading && (
          <Loader2
            size={16}
            className="absolute right-3 top-1/2 -translate-y-1/2 animate-spin text-[var(--color-muted-foreground)]"
          />
        )}
        <input
          ref={inputRef}
          type="search"
          placeholder="Search tracks, artists, albums…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setFocused(true)}
          onBlur={() => {
            // Defer so a click on a history item registers before we hide it.
            setTimeout(() => setFocused(false), 150);
          }}
          onKeyDown={(e) => {
            if (!showHistory) return;
            if (e.key === "ArrowDown") {
              e.preventDefault();
              setHistoryIndex((i) => Math.min(visibleHistory.length - 1, i + 1));
            } else if (e.key === "ArrowUp") {
              e.preventDefault();
              setHistoryIndex((i) => Math.max(0, i - 1));
            } else if (e.key === "Enter" && historyIndex >= 0) {
              e.preventDefault();
              selectHistoryQuery(visibleHistory[historyIndex]);
            } else if (e.key === "Escape") {
              setFocused(false);
            }
          }}
          className="w-full rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] py-3 pl-9 pr-4 text-sm outline-none focus:border-[var(--color-primary)] focus:ring-2 focus:ring-[var(--color-primary)] focus:ring-opacity-20 transition-colors"
        />

        {showHistory && (
          <div className="absolute left-0 right-0 top-full z-10 mt-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] shadow-lg">
            <div className="flex items-center justify-between px-4 pb-2 pt-3">
              <span className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-[var(--color-muted-foreground)]">
                <History size={12} />
                Recent searches
              </span>
              <button
                type="button"
                onClick={clearHistory}
                title="Clear history"
                className="text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
              >
                <Trash2 size={14} />
              </button>
            </div>
            <ul>
              {visibleHistory.map((q, idx) => (
                <li key={q}>
                  <div
                    className={cn(
                      "flex cursor-pointer items-center gap-2 px-4 py-2 text-sm transition-colors",
                      idx === historyIndex ? "bg-[var(--color-muted)]" : "hover:bg-[var(--color-muted)]"
                    )}
                    onMouseEnter={() => setHistoryIndex(idx)}
                    onClick={() => selectHistoryQuery(q)}
                  >
                    <History size={14} className="shrink-0 text-[var(--color-muted-foreground)]" />
                    <span className="min-w-0 flex-1 truncate">{q}</span>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        removeHistory(q);
                      }}
                      className="shrink-0 text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]"
                    >
                      <X size={13} />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>

      {!query.trim() && !showHistory && (
        <p className="text-center text-sm text-[var(--color-muted-foreground)]">Type to search your music library.</p>
      )}
      {empty && (
        <p className="text-center text-sm text-[var(--color-muted-foreground)]">No results for &ldquo;{query}&rdquo;</p>
      )}

      {results && (
        <div className="space-y-8">
          {/* Tracks */}
          {results.tracks.length > 0 && (
            <section>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-muted-foreground)]">
                Tracks
              </h2>
              <div className="rounded-xl border border-[var(--color-border)] bg-[var(--color-card)]">
                {results.tracks.map((t, idx) => (
                  <div
                    key={t.id}
                    onClick={() => {
                      const q = results.tracks.map((tr) => ({
                        id: tr.id,
                        title: tr.title,
                        artistName: "",
                        albumTitle: "",
                        durationSeconds: Math.round(tr.durationMs / 1000),
                        playbackUrl: "",
                      }));
                      playQueue(q, idx);
                    }}
                    className="flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors"
                  >
                    <Artwork alt={t.title} size="sm" />
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium">
                        <Highlighted raw={t.highlight} plain={t.title} />
                      </p>
                    </div>
                    <span className="text-xs text-[var(--color-muted-foreground)]">
                      {formatDuration(t.durationMs / 1000)}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          )}

          {/* Artists */}
          {results.artists.length > 0 && (
            <section>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-muted-foreground)]">
                Artists
              </h2>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {results.artists.map((a) => (
                  <Link
                    key={a.id}
                    href={`/artists/${a.id}`}
                    className="flex flex-col items-center gap-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-4 text-center hover:bg-[var(--color-muted)] transition-colors"
                  >
                    <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[var(--color-muted)] text-xl font-bold">
                      {a.name.charAt(0).toUpperCase()}
                    </div>
                    <p className="truncate text-sm font-medium">
                      <Highlighted raw={a.highlight} plain={a.name} />
                    </p>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {/* Albums */}
          {results.albums.length > 0 && (
            <section>
              <h2 className="mb-3 text-xs font-semibold uppercase tracking-wider text-[var(--color-muted-foreground)]">
                Albums
              </h2>
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                {results.albums.map((a) => (
                  <Link
                    key={a.id}
                    href={`/albums/${a.id}`}
                    className="flex flex-col gap-2 rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] p-3 hover:bg-[var(--color-muted)] transition-colors"
                  >
                    <Artwork alt={a.title} size="lg" className="w-full h-auto aspect-square" />
                    <p className="truncate text-sm font-medium">
                      <Highlighted raw={a.highlight} plain={a.title} />
                    </p>
                  </Link>
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense>
      <SearchPageInner />
    </Suspense>
  );
}
