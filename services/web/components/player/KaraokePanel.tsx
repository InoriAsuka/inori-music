"use client";

/**
 * KaraokePanel — fullscreen karaoke view with per-word progressive fill.
 *
 * Uses a native <dialog> via showModal() for top-layer semantics, Escape
 * handling and focus restoration (same approach as components/ui/Modal).
 * Word fill is a background-clip gradient driven by wordProgress(), so a
 * partially-sung word fills left-to-right rather than flipping colour.
 *
 * Lines without inline <mm:ss.xx> timings fall back to whole-line highlight.
 */

import { useEffect, useMemo, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import { useCurrentTrack } from "@/store/player";
import { useAuthStore } from "@/store/auth";
import { fetchLyrics } from "@/lib/lyrics/fetchLyrics";
import type { LyricLine } from "@/lib/lyrics/lyricLine";
import { activeLineIndex, wordProgress } from "@/lib/karaoke/progress";
import { useSmoothPosition } from "@/lib/karaoke/useSmoothPosition";
import { cn } from "@/lib/utils";

export function KaraokePanel({ open, onClose }: { open: boolean; onClose: () => void }) {
  const track = useCurrentTrack();
  const token = useAuthStore((s) => s.token);
  const positionMs = useSmoothPosition(open);

  const [lines, setLines] = useState<LyricLine[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [mounted, setMounted] = useState(false);

  const dialogRef = useRef<HTMLDialogElement>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const lineRefs = useRef<Map<number, HTMLDivElement>>(new Map());

  const trackId = track?.id ?? null;

  useEffect(() => setMounted(true), []);

  useEffect(() => {
    if (!trackId || !open) return;
    let cancelled = false;
    setLoading(true);
    fetchLyrics(trackId, { token: token ?? undefined }).then((result) => {
      if (cancelled) return;
      setLines(result);
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [trackId, token, open]);

  useEffect(() => {
    if (!open) return;
    previousFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    document.body.style.overflow = "hidden";
    if (dialogRef.current && !dialogRef.current.open) dialogRef.current.showModal();
    dialogRef.current?.focus();
    return () => {
      document.body.style.overflow = "";
      if (dialogRef.current?.open) dialogRef.current.close();
      previousFocusRef.current?.focus();
    };
  }, [open]);

  const activeIndex = useMemo(() => (lines ? activeLineIndex(lines, positionMs) : -1), [lines, positionMs]);

  useEffect(() => {
    if (!open || activeIndex < 0) return;
    lineRefs.current.get(activeIndex)?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [activeIndex, open]);

  if (!mounted) return null;

  return createPortal(
    <AnimatePresence>
      {open && (
        <motion.dialog
          ref={dialogRef}
          aria-label="Karaoke"
          tabIndex={-1}
          onCancel={(event) => {
            event.preventDefault();
            onClose();
          }}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.2 }}
          className="fixed inset-0 z-50 m-0 h-full max-h-full w-full max-w-full bg-[var(--color-void)] p-0 outline-none backdrop:bg-transparent"
        >
          <div className="flex h-full flex-col">
            <div className="flex shrink-0 items-center justify-between px-6 py-4">
              <div className="min-w-0">
                <p className="truncate font-display text-sm font-bold tracking-widest text-[var(--color-primary)]">
                  KARAOKE
                </p>
                {track && (
                  <p className="truncate text-xs text-[var(--color-text-muted)]">
                    {track.title}
                    {track.artistName ? ` — ${track.artistName}` : ""}
                  </p>
                )}
              </div>
              <button
                type="button"
                onClick={onClose}
                aria-label="Close karaoke"
                className="rounded p-2 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
              >
                <X size={20} />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto px-6 py-[35vh]">
              {!track ? (
                <p className="text-center text-base text-[var(--color-text-muted)]">No track playing.</p>
              ) : loading ? (
                <p className="text-center text-base text-[var(--color-text-muted)]">Loading lyrics…</p>
              ) : !lines || lines.length === 0 ? (
                <p className="text-center text-base text-[var(--color-text-muted)]">No lyrics available.</p>
              ) : (
                <div className="mx-auto max-w-3xl space-y-8">
                  {lines.map((line, i) => (
                    <div
                      key={`${line.timestampMs}-${i}`}
                      ref={(el) => {
                        if (el) lineRefs.current.set(i, el);
                        else lineRefs.current.delete(i);
                      }}
                      className={cn(
                        "text-center font-display text-3xl font-bold leading-snug transition-all duration-300 sm:text-4xl",
                        i === activeIndex ? "scale-100 opacity-100" : "scale-95 opacity-40"
                      )}
                    >
                      <KaraokeLine
                        line={line}
                        active={i === activeIndex}
                        positionMs={positionMs}
                        nextLineMs={lines[i + 1]?.timestampMs ?? null}
                      />
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </motion.dialog>
      )}
    </AnimatePresence>,
    document.body
  );
}

function KaraokeLine({
  line,
  active,
  positionMs,
  nextLineMs,
}: {
  line: LyricLine;
  active: boolean;
  positionMs: number;
  nextLineMs: number | null;
}) {
  const words = line.words;
  if (!words || words.length === 0) {
    return <p className={active ? "text-[var(--color-primary)]" : "text-[var(--color-text-muted)]"}>{line.text}</p>;
  }
  return (
    <p className="text-[var(--color-text-muted)]">
      {words.map((word, i) => (
        <KaraokeWord
          key={`${word.offsetMs}-${i}`}
          text={word.text}
          fill={active ? wordProgress(words, i, positionMs, nextLineMs) : 0}
        />
      ))}
    </p>
  );
}

/** Renders `text` with the leading `fill` fraction painted in the primary colour. */
function KaraokeWord({ text, fill }: { text: string; fill: number }) {
  if (fill <= 0) return <span>{text}</span>;
  if (fill >= 1) return <span className="text-[var(--color-primary)]">{text}</span>;
  const stop = `${fill * 100}%`;
  return (
    <span
      className="bg-clip-text text-transparent"
      style={{
        backgroundImage: `linear-gradient(to right, var(--color-primary) ${stop}, var(--color-text-muted) ${stop})`,
      }}
    >
      {text}
    </span>
  );
}
