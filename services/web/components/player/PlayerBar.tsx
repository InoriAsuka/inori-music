/**
 * PlayerBar — Neon Shrine 升级版
 * 集成：频谱可视化 · 队列抽屉 · 全屏播放器 · 错误态
 */
"use client";

import { useState } from "react";
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Volume2,
  VolumeX,
  Repeat,
  Repeat1,
  Shuffle,
  AlertCircle,
  ListMusic,
  ChevronUp,
  Mic2,
} from "lucide-react";
import { usePlayerStore, useCurrentTrack, useIsPlaying } from "@/store/player";
import { useAudio } from "@/hooks/useAudio";
import { Artwork } from "@/components/ui/Artwork";
import { Visualizer } from "./Visualizer";
import { QueueDrawer } from "./QueueDrawer";
import { FullscreenPlayer } from "./FullscreenPlayer";
import { LyricsPanel } from "./LyricsPanel";
import { formatDuration, cn } from "@/lib/utils";

export function PlayerBar() {
  const currentTrack = useCurrentTrack();
  const isPlaying = useIsPlaying();
  const {
    status,
    positionSeconds,
    volume,
    shuffle,
    repeat,
    play,
    pause,
    skipToNext,
    skipToPrevious,
    setVolume,
    toggleShuffle,
    cycleRepeat,
    queue,
    currentIndex,
  } = usePlayerStore();

  const { seek } = useAudio();
  const [queueOpen, setQueueOpen] = useState(false);
  const [fsOpen, setFsOpen] = useState(false);
  const [lyricsOpen, setLyricsOpen] = useState(false);

  const duration = currentTrack?.durationSeconds ?? 0;
  const progress = duration > 0 ? positionSeconds / duration : 0;
  const isError = status === "error";

  if (!currentTrack) {
    return (
      <div className="flex h-20 shrink-0 items-center justify-center border-t border-[var(--color-border)] bg-[var(--color-surface)] px-4 text-sm text-[var(--color-text-muted)]">
        No track playing
      </div>
    );
  }

  return (
    <>
      {/* Visualizer strip */}
      <div className="h-8 w-full bg-[var(--color-void)]">
        <Visualizer />
      </div>

      {/* Main player bar */}
      <div className="flex h-[72px] shrink-0 items-center gap-3 border-t border-[var(--color-border)] bg-[var(--color-surface)] px-3 sm:gap-4 sm:px-4">
        {/* Track info — tap to open fullscreen on mobile */}
        <button
          type="button"
          onClick={() => setFsOpen(true)}
          className="flex min-w-0 w-48 shrink-0 items-center gap-3 text-left sm:w-56"
        >
          <Artwork alt={currentTrack.title} src={currentTrack.artworkUrl} size="sm" />
          <div className="min-w-0">
            <p className="truncate text-sm font-medium text-[var(--color-text)]">{currentTrack.title}</p>
            <p className="truncate text-xs text-[var(--color-text-secondary)]">
              {currentTrack.artistName || currentTrack.albumTitle || ""}
            </p>
          </div>
          <ChevronUp size={14} className="shrink-0 text-[var(--color-text-muted)] sm:hidden" />
        </button>

        {/* Centre controls */}
        <div className="flex flex-1 flex-col items-center gap-1">
          {isError ? (
            <div className="flex items-center gap-2 text-sm text-[var(--color-danger)]">
              <AlertCircle size={14} />
              Playback failed
              <button
                type="button"
                onClick={skipToNext}
                className="rounded-md border border-[var(--color-danger)] px-2 py-0.5 text-xs hover:bg-[var(--color-danger)] hover:text-white"
              >
                Skip
              </button>
            </div>
          ) : (
            <>
              <div className="flex items-center gap-3 sm:gap-4">
                <ControlBtn onClick={toggleShuffle} active={shuffle} title="Shuffle">
                  <Shuffle size={15} />
                </ControlBtn>

                <ControlBtn onClick={skipToPrevious} title="Previous">
                  <SkipBack size={20} fill="currentColor" />
                </ControlBtn>

                <button
                  type="button"
                  onClick={isPlaying ? pause : play}
                  className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--color-primary)] text-[var(--color-primary-fg)] hover:opacity-90 transition-opacity glow-primary"
                  title={isPlaying ? "Pause" : "Play"}
                >
                  {isPlaying ? (
                    <Pause size={17} fill="currentColor" />
                  ) : (
                    <Play size={17} fill="currentColor" className="ml-0.5" />
                  )}
                </button>

                <ControlBtn
                  onClick={skipToNext}
                  title="Next"
                  disabled={queue.length === 0 || currentIndex >= queue.length - 1}
                >
                  <SkipForward size={20} fill="currentColor" />
                </ControlBtn>

                <ControlBtn onClick={cycleRepeat} active={repeat !== "off"} title="Repeat">
                  {repeat === "one" ? <Repeat1 size={15} /> : <Repeat size={15} />}
                </ControlBtn>
              </div>

              {/* Progress */}
              <div className="hidden w-full max-w-lg items-center gap-2 sm:flex">
                <span className="w-10 text-right font-mono text-xs text-[var(--color-text-muted)]">
                  {formatDuration(positionSeconds)}
                </span>
                <div
                  role="slider"
                  tabIndex={0}
                  aria-valuenow={Math.round(positionSeconds)}
                  aria-valuemin={0}
                  aria-valuemax={Math.round(duration)}
                  className="relative h-1 flex-1 cursor-pointer rounded-full bg-[var(--color-border)]"
                  onClick={(e) => {
                    const rect = e.currentTarget.getBoundingClientRect();
                    seek(((e.clientX - rect.left) / rect.width) * duration);
                  }}
                  onKeyDown={(e) => {
                    if (e.key === "ArrowRight") seek(Math.min(duration, positionSeconds + 5));
                    if (e.key === "ArrowLeft") seek(Math.max(0, positionSeconds - 5));
                  }}
                >
                  <div
                    className="absolute inset-y-0 left-0 rounded-full bg-[var(--color-primary)]"
                    style={{ width: `${progress * 100}%` }}
                  />
                </div>
                <span className="w-10 font-mono text-xs text-[var(--color-text-muted)]">
                  {formatDuration(duration)}
                </span>
              </div>
            </>
          )}
        </div>

        {/* Right controls */}
        <div className="hidden items-center gap-2 lg:flex">
          <button
            type="button"
            onClick={() => setVolume(volume > 0 ? 0 : 0.7)}
            className="text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors"
            title={volume > 0 ? "Mute" : "Unmute"}
          >
            {volume === 0 ? <VolumeX size={15} /> : <Volume2 size={15} />}
          </button>
          <input
            type="range"
            min={0}
            max={1}
            step={0.01}
            value={volume}
            onChange={(e) => setVolume(Number.parseFloat(e.target.value))}
            className="h-1 w-24 cursor-pointer accent-[var(--color-primary)]"
            aria-label="Volume"
          />
        </div>

        <ControlBtn onClick={() => setLyricsOpen(true)} title="Lyrics">
          <Mic2 size={16} />
        </ControlBtn>

        <ControlBtn onClick={() => setQueueOpen(true)} title="Queue">
          <ListMusic size={16} />
        </ControlBtn>
      </div>

      <QueueDrawer open={queueOpen} onClose={() => setQueueOpen(false)} />
      <LyricsPanel open={lyricsOpen} onClose={() => setLyricsOpen(false)} />
      <FullscreenPlayer open={fsOpen} onClose={() => setFsOpen(false)} />
    </>
  );
}

function ControlBtn({
  children,
  onClick,
  active,
  disabled,
  title,
}: {
  children: React.ReactNode;
  onClick: () => void;
  active?: boolean;
  disabled?: boolean;
  title?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={title}
      className={cn(
        "flex items-center justify-center rounded p-1.5 transition-colors",
        active ? "text-[var(--color-primary)]" : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]",
        disabled && "opacity-30 pointer-events-none"
      )}
    >
      {children}
    </button>
  );
}
