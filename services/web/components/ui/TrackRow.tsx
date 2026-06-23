/**
 * TrackRow — reusable track list item with:
 * - play on click / click index
 * - isFavorite heart button (toggle)
 * - duration display
 *
 * artistName is optional; if omitted the row still renders correctly.
 */
"use client";

import { Heart } from "lucide-react";
import Link from "next/link";
import { cn, formatDuration } from "@/lib/utils";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { useState } from "react";

export interface TrackRowData {
  id: string;
  title: string;
  artistName?: string;
  albumTitle?: string;
  durationMs: number;
  isFavorite?: boolean;
}

interface TrackRowProps {
  track: TrackRowData;
  index?: number;
  onPlay: () => void;
  onFavoriteChange?: (trackId: string, nowFavorite: boolean) => void;
  showIndex?: boolean;
  className?: string;
}

export function TrackRow({
  track,
  index,
  onPlay,
  onFavoriteChange,
  showIndex = true,
  className,
}: TrackRowProps) {
  const token = useAuthStore((s) => s.token);
  const [fav, setFav] = useState(track.isFavorite ?? false);
  const [favLoading, setFavLoading] = useState(false);

  async function toggleFavorite(e: React.MouseEvent) {
    e.stopPropagation();
    if (!token || favLoading) return;
    setFavLoading(true);
    const client = authedApi(token);
    try {
      if (fav) {
        await client.DELETE("/api/v1/me/favorites/tracks/{trackId}", {
          params: { path: { trackId: track.id } },
        });
        setFav(false);
        onFavoriteChange?.(track.id, false);
      } else {
        await client.POST("/api/v1/me/favorites/tracks/{trackId}", {
          params: { path: { trackId: track.id } },
        });
        setFav(true);
        onFavoriteChange?.(track.id, true);
      }
    } finally {
      setFavLoading(false);
    }
  }

  return (
    <div
      onClick={onPlay}
      className={cn(
        "group flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors",
        className
      )}
    >
      {showIndex && index != null && (
        <span className="w-6 shrink-0 text-right text-sm text-[var(--color-muted-foreground)]">
          {index}
        </span>
      )}

      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">
          <Link
            href={`/tracks/${track.id}`}
            onClick={(e) => e.stopPropagation()}
            className="hover:text-[var(--color-primary)] transition-colors"
          >
            {track.title}
          </Link>
        </p>
        {(track.artistName || track.albumTitle) && (
          <p className="truncate text-xs text-[var(--color-muted-foreground)]">
            {[track.artistName, track.albumTitle].filter(Boolean).join(" · ")}
          </p>
        )}
      </div>

      <button
        onClick={toggleFavorite}
        disabled={favLoading}
        className={cn(
          "shrink-0 rounded p-1.5 transition-colors",
          fav
            ? "text-[var(--color-primary)]"
            : "text-transparent group-hover:text-[var(--color-muted-foreground)] hover:!text-[var(--color-primary)]"
        )}
        title={fav ? "Remove from favorites" : "Add to favorites"}
      >
        <Heart size={14} fill={fav ? "currentColor" : "none"} />
      </button>

      <span className="w-12 shrink-0 text-right text-xs text-[var(--color-muted-foreground)]">
        {formatDuration(track.durationMs / 1000)}
      </span>
    </div>
  );
}
