/**
 * PlayerBar — Sakura Dusk 版
 * 集成：频谱可视化 · 队列抽屉 · 全屏播放器 · 错误态
 */
"use client";

import { useState, useRef } from "react";
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
import { SpeedControl } from "./SpeedControl";
import { SleepTimerControl } from "./SleepTimerControl";
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
  const isError = status === "error";

  if (!currentTrack) {
    return (
      <div className="flex h-20 shrink-0 items-center justify-center border-t border-[var(--color-border)] bg-[var(--color-void)] px-4 text-sm text-[var(--color-text-muted)]">
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
      <div className="flex h-[72px] shrink-0 items-center gap-1 border-t border-[var(--color-border)] bg-[var(--color-void)] px-2 sm:gap-2 sm:px-4">
        {/* Track info — tap to open fullscreen on mobile */}
        <button
          type="button"
          onClick={() => setFsOpen(true)}
          className="flex min-w-0 flex-1 items-center gap-2 text-left sm:w-56 sm:flex-none sm:shrink-0 sm:gap-3"
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
        <div className="hidden flex-1 flex-col items-center gap-1 min-[480px]:flex">
          {isError ? (
            <div className="flex items-center gap-2 text-sm text-[var(--color-danger)]">
              <AlertCircle size={14} />
              Playback failed
              <button
                type="button"
                onClick={skipToNext}
                className="rounded-md border border-[var(--color-danger)] px-2 py-0.5 text-xs transition-colors hover:bg-[var(--color-danger)] hover:text-[var(--color-primary-ink)]"
              >
                Skip
              </button>
            </div>
          ) : (
            <>
              <div className="flex items-center gap-1 sm:gap-2">
                <span className="hidden md:inline-flex">
                  <ControlBtn onClick={toggleShuffle} active={shuffle} title="Shuffle">
                    <Shuffle size={15} />
                  </ControlBtn>
                </span>

                <ControlBtn onClick={skipToPrevious} title="Previous">
                  <SkipBack size={20} fill="currentColor" />
                </ControlBtn>

                <button
                  type="button"
                  onClick={isPlaying ? pause : play}
                  className="flex h-11 w-11 items-center justify-center rounded-full bg-[var(--color-primary)] text-[var(--color-primary-ink)] transition-transform duration-150 hover:bg-[var(--color-primary-hover)] active:scale-95 glow-primary"
                  title={isPlaying ? "Pause" : "Play"}
                >
                  {isPlaying ? (
                    <Pause size={18} fill="currentColor" />
                  ) : (
                    <Play size={18} fill="currentColor" className="ml-0.5" />
                  )}
                </button>

                <ControlBtn
                  onClick={skipToNext}
                  title="Next"
                  disabled={queue.length === 0 || currentIndex >= queue.length - 1}
                >
                  <SkipForward size={20} fill="currentColor" />
                </ControlBtn>

                <span className="hidden md:inline-flex">
                  <ControlBtn onClick={cycleRepeat} active={repeat !== "off"} title="Repeat">
                    {repeat === "one" ? <Repeat1 size={15} /> : <Repeat size={15} />}
                  </ControlBtn>
                </span>
              </div>

              {/* Progress */}
              <div className="hidden w-full max-w-lg items-center gap-2 sm:flex">
                <span className="w-10 text-right font-mono text-xs tabular text-[var(--color-text-muted)]">
                  {formatDuration(positionSeconds)}
                </span>
                <ProgressBar positionSeconds={positionSeconds} duration={duration} onSeek={seek} />
                <span className="w-10 font-mono text-xs tabular text-[var(--color-text-muted)]">
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

        <div className="hidden sm:block">
          <SpeedControl />
        </div>

        <div className="hidden sm:block">
          <SleepTimerControl />
        </div>

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

function ProgressBar({
  positionSeconds,
  duration,
  onSeek,
}: {
  positionSeconds: number;
  duration: number;
  onSeek: (seconds: number) => void;
}) {
  const trackRef = useRef<HTMLDivElement | null>(null);
  // While dragging, render the finger position instead of the audio element's
  // clock so the handle tracks the pointer without waiting for timeupdate.
  const [dragSeconds, setDragSeconds] = useState<number | null>(null);

  const shown = dragSeconds ?? positionSeconds;
  const progress = duration > 0 ? Math.min(1, Math.max(0, shown / duration)) : 0;

  function secondsAt(clientX: number) {
    const rect = trackRef.current?.getBoundingClientRect();
    if (!rect || rect.width === 0) return 0;
    const ratio = (clientX - rect.left) / rect.width;
    return Math.min(duration, Math.max(0, ratio * duration));
  }

  return (
    <div
      ref={trackRef}
      role="slider"
      tabIndex={0}
      aria-label="Seek"
      aria-valuenow={Math.round(shown)}
      aria-valuemin={0}
      aria-valuemax={Math.round(duration)}
      className="group relative flex h-4 flex-1 cursor-pointer items-center touch-none"
      onPointerDown={(e) => {
        e.currentTarget.setPointerCapture(e.pointerId);
        setDragSeconds(secondsAt(e.clientX));
      }}
      onPointerMove={(e) => {
        if (dragSeconds === null) return;
        setDragSeconds(secondsAt(e.clientX));
      }}
      onPointerUp={(e) => {
        const target = dragSeconds ?? secondsAt(e.clientX);
        setDragSeconds(null);
        onSeek(target);
      }}
      onPointerCancel={() => setDragSeconds(null)}
      onKeyDown={(e) => {
        if (e.key === "ArrowRight") onSeek(Math.min(duration, positionSeconds + 5));
        if (e.key === "ArrowLeft") onSeek(Math.max(0, positionSeconds - 5));
      }}
    >
      <div className="relative h-1.5 w-full rounded-full bg-[var(--color-overlay)] transition-[height] duration-150 group-hover:h-2">
        <div
          className="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-[var(--color-primary)] to-[var(--color-primary-soft)]"
          style={{ width: `${progress * 100}%` }}
        />
        <span
          aria-hidden="true"
          className="absolute top-1/2 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[var(--color-primary)] opacity-0 shadow transition-opacity group-hover:opacity-100"
          style={{ left: `${progress * 100}%`, opacity: dragSeconds !== null ? 1 : undefined }}
        />
      </div>
    </div>
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
        "flex h-11 w-11 shrink-0 items-center justify-center rounded-lg transition-colors",
        active
          ? "text-[var(--color-primary)]"
          : "text-[var(--color-text-muted)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]",
        disabled && "opacity-30 pointer-events-none"
      )}
    >
      {children}
    </button>
  );
}
