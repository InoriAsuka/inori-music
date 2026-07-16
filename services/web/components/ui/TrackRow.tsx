/**
 * TrackRow — reusable track list item with:
 * - play on click / click index
 * - isFavorite heart button (toggle)
 * - duration display
 * - overflow (⋯) menu with "Add to playlist" (rendered only when authenticated)
 *
 * artistName is optional; if omitted the row still renders correctly.
 *
 * API stability: every new capability is opt-out via defaulted props, so the
 * existing callers (`tracks/page.tsx`, `albums/[id]/page.tsx`) stay
 * source-compatible. The overflow menu self-suppresses when no auth token is
 * available, so it is never forced onto a token-less context.
 */
"use client";

import { Heart, MoreHorizontal, ListPlus } from "lucide-react";
import Link from "next/link";
import { cn, formatDuration } from "@/lib/utils";
import { useAuthStore } from "@/store/auth";
import { authedApi } from "@/lib/api/client";
import { useEffect, useRef, useState } from "react";
import { AddToPlaylistDialog } from "@/components/ui/AddToPlaylistDialog";

const MENU_ITEM_SELECTOR = '[role="menuitem"], [role="menuitemradio"]';

function getMenuItems(panel: HTMLElement | null): HTMLElement[] {
  if (!panel) return [];
  return Array.from(panel.querySelectorAll<HTMLElement>(MENU_ITEM_SELECTOR)).filter(
    (item) => item.getAttribute("aria-disabled") !== "true"
  );
}

function focusMenuItem(items: HTMLElement[], index: number) {
  items.forEach((item, itemIndex) => {
    item.tabIndex = itemIndex === index ? 0 : -1;
  });
  items[index]?.focus();
}

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
  showFavorite?: boolean;
  className?: string;
  leading?: React.ReactNode;
  trailing?: React.ReactNode;
  /**
   * Whether to show the overflow menu with "Add to playlist". Defaults to true,
   * but the menu still only renders when an auth token exists — so token-less
   * usages are never forced to show a non-functional menu.
   */
  showMenu?: boolean;
}

export function TrackRow({
  track,
  index,
  onPlay,
  onFavoriteChange,
  showIndex = true,
  showFavorite = true,
  className,
  leading,
  trailing,
  showMenu = true,
}: TrackRowProps) {
  const token = useAuthStore((s) => s.token);
  const [fav, setFav] = useState(track.isFavorite ?? false);
  const [favLoading, setFavLoading] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [addOpen, setAddOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const menuPanelRef = useRef<HTMLDivElement>(null);
  const menuTriggerRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!menuOpen) return;

    const items = getMenuItems(menuPanelRef.current);
    if (items.length > 0) {
      items.forEach((item, index) => {
        item.tabIndex = index === 0 ? 0 : -1;
      });
      items[0].focus();
    }

    function onPointerDown(e: PointerEvent) {
      if (!menuRef.current || menuRef.current.contains(e.target as Node)) return;

      const focusWasInMenu = menuPanelRef.current?.contains(document.activeElement) ?? false;
      setMenuOpen(false);
      if (focusWasInMenu) {
        requestAnimationFrame(() => {
          if (document.activeElement === document.body) menuTriggerRef.current?.focus();
        });
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        e.stopPropagation();
        setMenuOpen(false);
        menuTriggerRef.current?.focus();
        return;
      }
      if (!menuPanelRef.current || !["ArrowDown", "ArrowUp", "Home", "End"].includes(e.key)) return;
      const items = getMenuItems(menuPanelRef.current);
      if (items.length === 0) return;
      e.preventDefault();
      const currentIndex = items.indexOf(document.activeElement as HTMLElement);
      if (e.key === "Home") focusMenuItem(items, 0);
      else if (e.key === "End") focusMenuItem(items, items.length - 1);
      else if (e.key === "ArrowDown") focusMenuItem(items, (currentIndex + 1 + items.length) % items.length);
      else focusMenuItem(items, (currentIndex - 1 + items.length) % items.length);
    }
    document.addEventListener("pointerdown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [menuOpen]);

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

  const canShowMenu = showMenu && !!token;

  return (
    <div
      onClick={onPlay}
      className={cn(
        "group flex cursor-pointer items-center gap-3 border-b border-[var(--color-border)] px-4 py-2.5 last:border-0 hover:bg-[var(--color-muted)] transition-colors",
        className
      )}
    >
      {leading && (
        <span className="contents" onClick={(e) => e.stopPropagation()}>
          {leading}
        </span>
      )}
      {showIndex && index != null && (
        <span className="w-6 shrink-0 text-right text-sm text-[var(--color-muted-foreground)]">{index}</span>
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

      {showFavorite && (
        <button
          type="button"
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
      )}

      {canShowMenu && (
        <div ref={menuRef} className="relative shrink-0" onClick={(e) => e.stopPropagation()}>
          <button
            ref={menuTriggerRef}
            type="button"
            onClick={() => setMenuOpen((o) => !o)}
            aria-haspopup="menu"
            aria-expanded={menuOpen}
            aria-label="Track options"
            className="rounded p-1.5 text-transparent transition-colors group-hover:text-[var(--color-muted-foreground)] hover:!text-[var(--color-foreground)] aria-expanded:text-[var(--color-foreground)]"
          >
            <MoreHorizontal size={15} />
          </button>
          {menuOpen && (
            <div
              ref={menuPanelRef}
              role="menu"
              className="absolute right-0 top-full z-20 mt-1 w-44 overflow-hidden rounded-lg border border-[var(--color-border)] bg-[var(--color-card)] py-1 shadow-xl"
            >
              <button
                type="button"
                role="menuitem"
                onClick={() => {
                  setMenuOpen(false);
                  setAddOpen(true);
                }}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-[var(--color-muted)]"
              >
                <ListPlus size={15} className="text-[var(--color-muted-foreground)]" />
                Add to playlist
              </button>
            </div>
          )}
        </div>
      )}

      <span className="w-12 shrink-0 text-right text-xs text-[var(--color-muted-foreground)]">
        {formatDuration(track.durationMs / 1000)}
      </span>

      {trailing && (
        <span className="contents" onClick={(e) => e.stopPropagation()}>
          {trailing}
        </span>
      )}

      {token && (
        <AddToPlaylistDialog
          open={addOpen}
          onClose={() => setAddOpen(false)}
          token={token}
          trackId={track.id}
          trackTitle={track.title}
        />
      )}
    </div>
  );
}
