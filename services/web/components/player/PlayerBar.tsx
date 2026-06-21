/**
 * PlayerBar — fixed bottom bar with playback controls.
 * Mounts useAudio to drive the HTMLAudioElement.
 * Shows an error state when playback fails.
 */
"use client";

import {
  Play, Pause, SkipBack, SkipForward,
  Volume2, VolumeX, Repeat, Repeat1, Shuffle, AlertCircle,
} from "lucide-react";
import { usePlayerStore, useCurrentTrack, useIsPlaying } from "@/store/player";
import { useAudio } from "@/hooks/useAudio";
import { Artwork } from "@/components/ui/Artwork";
import { formatDuration } from "@/lib/utils";
import { cn } from "@/lib/utils";

export function PlayerBar() {
  const currentTrack = useCurrentTrack();
  const isPlaying = useIsPlaying();
  const {
    status, positionSeconds, volume, shuffle, repeat,
    play, pause, skipToNext, skipToPrevious,
    setVolume, toggleShuffle, cycleRepeat,
    queue, currentIndex,
  } = usePlayerStore();

  const { seek } = useAudio();

  const duration = currentTrack?.durationSeconds ?? 0;
  const progress = duration > 0 ? positionSeconds / duration : 0;
  const isError = status === "error";

  if (!currentTrack) {
    return (
      <div className="flex h-20 shrink-0 items-center justify-center border-t border-[var(--color-border)] bg-[var(--color-card)] px-4 text-sm text-[var(--color-muted-foreground)]">
        No track playing
      </div>
    );
  }

  return (
    <div className="flex h-20 shrink-0 items-center gap-4 border-t border-[var(--color-border)] bg-[var(--color-card)] px-4">
      {/* Track info */}
      <div className="flex w-56 min-w-0 items-center gap-3">
        <Artwork alt={currentTrack.title} src={currentTrack.artworkUrl} size="sm" />
        <div className="min-w-0">
          <p className="truncate text-sm font-medium">{currentTrack.title}</p>
          <p className="truncate text-xs text-[var(--color-muted-foreground)]">
            {currentTrack.artistName || currentTrack.albumTitle || ""}
          </p>
        </div>
      </div>

      {/* Controls */}
      <div className="flex flex-1 flex-col items-center gap-1.5">
        {isError ? (
          <div className="flex items-center gap-2 text-sm text-[var(--color-destructive)]">
            <AlertCircle size={14} />
            Playback failed — check storage backend
            <button
              onClick={skipToNext}
              className="ml-2 rounded-md border border-[var(--color-border)] px-2 py-0.5 text-xs hover:bg-[var(--color-muted)]"
            >
              Skip
            </button>
          </div>
        ) : (
          <>
            {/* Buttons */}
            <div className="flex items-center gap-4">
              <ControlBtn onClick={toggleShuffle} active={shuffle} title="Shuffle">
                <Shuffle size={16} />
              </ControlBtn>

              <ControlBtn onClick={skipToPrevious} title="Previous">
                <SkipBack size={20} fill="currentColor" />
              </ControlBtn>

              <button
                onClick={isPlaying ? pause : play}
                className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--color-foreground)] text-[var(--color-background)] hover:opacity-80 transition-opacity"
                title={isPlaying ? "Pause" : "Play"}
              >
                {isPlaying
                  ? <Pause size={18} fill="currentColor" />
                  : <Play size={18} fill="currentColor" className="ml-0.5" />}
              </button>

              <ControlBtn
                onClick={skipToNext}
                title="Next"
                disabled={queue.length === 0 || currentIndex >= queue.length - 1}
              >
                <SkipForward size={20} fill="currentColor" />
              </ControlBtn>

              <ControlBtn onClick={cycleRepeat} active={repeat !== "off"} title="Repeat">
                {repeat === "one" ? <Repeat1 size={16} /> : <Repeat size={16} />}
              </ControlBtn>
            </div>

            {/* Progress bar */}
            <div className="flex w-full max-w-lg items-center gap-2">
              <span className="w-10 text-right text-xs text-[var(--color-muted-foreground)]">
                {formatDuration(positionSeconds)}
              </span>
              <div
                className="relative h-1 flex-1 cursor-pointer rounded-full bg-[var(--color-muted)]"
                onClick={(e) => {
                  const rect = e.currentTarget.getBoundingClientRect();
                  seek(((e.clientX - rect.left) / rect.width) * duration);
                }}
              >
                <div
                  className="absolute inset-y-0 left-0 rounded-full bg-[var(--color-foreground)]"
                  style={{ width: `${progress * 100}%` }}
                />
              </div>
              <span className="w-10 text-xs text-[var(--color-muted-foreground)]">
                {formatDuration(duration)}
              </span>
            </div>
          </>
        )}
      </div>

      {/* Volume */}
      <div className="hidden w-32 items-center gap-2 lg:flex">
        <button
          onClick={() => setVolume(volume > 0 ? 0 : 0.7)}
          className="text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)] transition-colors"
          title={volume > 0 ? "Mute" : "Unmute"}
        >
          {volume === 0 ? <VolumeX size={16} /> : <Volume2 size={16} />}
        </button>
        <input
          type="range" min={0} max={1} step={0.01} value={volume}
          onChange={(e) => setVolume(parseFloat(e.target.value))}
          className="h-1 w-full cursor-pointer accent-[var(--color-foreground)]"
          title="Volume"
        />
      </div>
    </div>
  );
}

function ControlBtn({
  children, onClick, active, disabled, title,
}: {
  children: React.ReactNode;
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title?: string;
}) {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={cn(
        "flex items-center justify-center rounded p-1.5 transition-colors",
        active
          ? "text-[var(--color-primary)]"
          : "text-[var(--color-muted-foreground)] hover:text-[var(--color-foreground)]",
        disabled && "opacity-30 pointer-events-none"
      )}
    >
      {children}
    </button>
  );
}
