"use client";

import { useEffect, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { X, Play, Pause, SkipBack, SkipForward } from "lucide-react";
import { useCurrentTrack, useIsPlaying, usePlayerStore } from "@/store/player";
import { Artwork } from "@/components/ui/Artwork";
import { SpeedControl } from "./SpeedControl";
import { SleepTimerControl } from "./SleepTimerControl";
import { formatDuration } from "@/lib/utils";

/**
 * Samples the artwork down to a single pixel to get its average colour, used to
 * tint the fullscreen backdrop. Returns null when there is no artwork or the
 * image is cross-origin without CORS headers (which taints the canvas).
 */
function useArtworkTint(src?: string | null): string | null {
  const [tint, setTint] = useState<string | null>(null);

  useEffect(() => {
    if (!src) {
      setTint(null);
      return;
    }

    let cancelled = false;
    const img = new Image();
    img.crossOrigin = "anonymous";
    img.src = src;

    img.onload = () => {
      if (cancelled) return;
      try {
        const canvas = document.createElement("canvas");
        canvas.width = 1;
        canvas.height = 1;
        const ctx = canvas.getContext("2d");
        if (!ctx) return;
        ctx.drawImage(img, 0, 0, 1, 1);
        const [r, g, b] = ctx.getImageData(0, 0, 1, 1).data;
        setTint(`rgb(${r} ${g} ${b})`);
      } catch {
        // Tainted canvas — fall back to the theme gradient.
        setTint(null);
      }
    };
    img.onerror = () => {
      if (!cancelled) setTint(null);
    };

    return () => {
      cancelled = true;
    };
  }, [src]);

  return tint;
}

export function FullscreenPlayer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const track = useCurrentTrack();
  const playing = useIsPlaying();
  const { play, pause, skipToNext, skipToPrevious, positionSeconds } = usePlayerStore();
  const duration = track?.durationSeconds ?? 0;
  const progress = duration > 0 ? positionSeconds / duration : 0;
  const tint = useArtworkTint(track?.artworkUrl);

  return (
    <AnimatePresence>
      {open && track && (
        <motion.div
          className="fixed inset-0 z-50 flex flex-col overflow-hidden bg-[var(--color-void)] p-4 sm:hidden"
          initial={{ y: "100%" }}
          animate={{ y: 0 }}
          exit={{ y: "100%" }}
          transition={{ duration: 0.22 }}
        >
          {/* Ambient wash — driven by the artwork when we could sample it. */}
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0"
            style={{
              background: tint
                ? `radial-gradient(60rem 40rem at 50% 18%, color-mix(in srgb, ${tint} 30%, transparent), transparent 70%)`
                : "radial-gradient(60rem 40rem at 50% 18%, color-mix(in srgb, var(--color-primary-soft) 38%, transparent), transparent 70%)",
            }}
          />

          <div className="relative z-20 flex items-center justify-between">
            {/* Keep utility menus at the top and open them downward so short
                mobile viewports never clip them against the bottom edge. */}
            <div className="flex items-center gap-2">
              <SpeedControl placement="below" align="left" />
              <SleepTimerControl placement="below" align="left" />
            </div>
            <button
              type="button"
              onClick={onClose}
              aria-label="Close fullscreen player"
              className="flex h-11 w-11 items-center justify-center rounded-full border border-[var(--color-border-strong)] text-[var(--color-text-secondary)]"
            >
              <X size={18} />
            </button>
          </div>
          <div className="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-evenly gap-3 py-2">
            <Artwork
              alt={track.title}
              src={track.artworkUrl}
              size="lg"
              className={`h-[min(18rem,38dvh)] w-[min(18rem,38dvh)] rounded-full glow-primary ${
                playing ? "vinyl-spin" : ""
              }`}
            />
            <div className="w-full text-center">
              <h2 className="truncate font-display text-xl font-bold text-[var(--color-text)]">{track.title}</h2>
              <p className="mt-1 truncate text-sm text-[var(--color-text-secondary)]">
                {track.artistName || track.albumTitle}
              </p>
            </div>
            <div className="w-full max-w-sm space-y-2">
              <div className="h-1.5 rounded-full bg-[var(--color-overlay)]">
                <div
                  className="h-full rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-soft)]"
                  style={{ width: `${progress * 100}%` }}
                />
              </div>
              <div className="flex justify-between font-mono text-xs tabular text-[var(--color-text-muted)]">
                <span>{formatDuration(positionSeconds)}</span>
                <span>{formatDuration(duration)}</span>
              </div>
            </div>
            <div className="flex items-center gap-8">
              <button
                type="button"
                onClick={skipToPrevious}
                aria-label="Previous track"
                className="text-[var(--color-text-secondary)] transition-transform active:scale-95"
              >
                <SkipBack size={28} fill="currentColor" />
              </button>
              <button
                type="button"
                onClick={playing ? pause : play}
                aria-label={playing ? "Pause" : "Play"}
                className="flex h-16 w-16 items-center justify-center rounded-full bg-[var(--color-primary)] text-[var(--color-primary-ink)] transition-transform active:scale-95 glow-primary"
              >
                {playing ? (
                  <Pause size={28} fill="currentColor" />
                ) : (
                  <Play size={28} fill="currentColor" className="ml-1" />
                )}
              </button>
              <button
                type="button"
                onClick={skipToNext}
                aria-label="Next track"
                className="text-[var(--color-text-secondary)] transition-transform active:scale-95"
              >
                <SkipForward size={28} fill="currentColor" />
              </button>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
