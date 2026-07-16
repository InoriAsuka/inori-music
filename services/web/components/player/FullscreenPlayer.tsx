"use client";

import { AnimatePresence, motion } from "motion/react";
import { X, Play, Pause, SkipBack, SkipForward } from "lucide-react";
import { useCurrentTrack, useIsPlaying, usePlayerStore } from "@/store/player";
import { Artwork } from "@/components/ui/Artwork";
import { SpeedControl } from "./SpeedControl";
import { SleepTimerControl } from "./SleepTimerControl";
import { formatDuration } from "@/lib/utils";

export function FullscreenPlayer({ open, onClose }: { open: boolean; onClose: () => void }) {
  const track = useCurrentTrack();
  const playing = useIsPlaying();
  const { play, pause, skipToNext, skipToPrevious, positionSeconds } = usePlayerStore();
  const duration = track?.durationSeconds ?? 0;
  const progress = duration > 0 ? positionSeconds / duration : 0;

  return (
    <AnimatePresence>
      {open && track && (
        <motion.div
          className="fixed inset-0 z-50 flex flex-col overflow-hidden bg-[var(--color-void)] p-4 sm:hidden scanlines"
          initial={{ y: "100%" }}
          animate={{ y: 0 }}
          exit={{ y: "100%" }}
          transition={{ duration: 0.22 }}
        >
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
              className="flex h-10 w-10 items-center justify-center rounded-full border border-[var(--color-border)] text-[var(--color-text-secondary)]"
            >
              <X size={18} />
            </button>
          </div>
          <div className="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-evenly gap-3 py-2">
            <Artwork
              alt={track.title}
              src={track.artworkUrl}
              size="lg"
              className="h-[min(18rem,38dvh)] w-[min(18rem,38dvh)] glow-primary"
            />
            <div className="w-full text-center">
              <h2 className="truncate text-xl font-semibold text-[var(--color-text)]">{track.title}</h2>
              <p className="mt-1 truncate text-sm text-[var(--color-text-secondary)]">
                {track.artistName || track.albumTitle}
              </p>
            </div>
            <div className="w-full max-w-sm space-y-2">
              <div className="h-1 rounded-full bg-[var(--color-surface-raised)]">
                <div
                  className="h-full rounded-full bg-[var(--color-primary)]"
                  style={{ width: `${progress * 100}%` }}
                />
              </div>
              <div className="flex justify-between font-mono text-xs text-[var(--color-text-muted)]">
                <span>{formatDuration(positionSeconds)}</span>
                <span>{formatDuration(duration)}</span>
              </div>
            </div>
            <div className="flex items-center gap-8">
              <button type="button" onClick={skipToPrevious} className="text-[var(--color-text-secondary)]">
                <SkipBack size={28} fill="currentColor" />
              </button>
              <button
                type="button"
                onClick={playing ? pause : play}
                className="flex h-16 w-16 items-center justify-center rounded-full bg-[var(--color-primary)] text-[var(--color-primary-fg)] glow-primary"
              >
                {playing ? (
                  <Pause size={28} fill="currentColor" />
                ) : (
                  <Play size={28} fill="currentColor" className="ml-1" />
                )}
              </button>
              <button type="button" onClick={skipToNext} className="text-[var(--color-text-secondary)]">
                <SkipForward size={28} fill="currentColor" />
              </button>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
