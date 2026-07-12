"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { X, Languages } from "lucide-react";
import { useCurrentTrack, usePlayerStore } from "@/store/player";
import { useAuthStore } from "@/store/auth";
import { fetchLyrics } from "@/lib/lyrics/fetchLyrics";
import { useBilingualToggle } from "@/lib/lyrics/bilingualToggle";
import type { LyricLine } from "@/lib/lyrics/lyricLine";
import { cn } from "@/lib/utils";

export function LyricsPanel({ open, onClose }: { open: boolean; onClose: () => void }) {
  const track = useCurrentTrack();
  const positionSeconds = usePlayerStore((s) => s.positionSeconds);
  const token = useAuthStore((s) => s.token);
  const [bilingual, setBilingual] = useBilingualToggle();

  const [lines, setLines] = useState<LyricLine[] | null>(null);
  const [loading, setLoading] = useState(false);

  const trackId = track?.id ?? null;

  useEffect(() => {
    if (!trackId) {
      setLines(null);
      return;
    }
    let cancelled = false;
    setLoading(true);
    fetchLyrics(trackId, { token: token ?? undefined }).then((result) => {
      if (!cancelled) {
        setLines(result);
        setLoading(false);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [trackId, token]);

  const positionMs = positionSeconds * 1000;

  const activeIndex = useMemo(() => {
    if (!lines || lines.length === 0) return -1;
    let idx = -1;
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].timestampMs <= positionMs) idx = i;
      else break;
    }
    return idx;
  }, [lines, positionMs]);

  const containerRef = useRef<HTMLDivElement | null>(null);
  const lineRefs = useRef<Map<number, HTMLDivElement>>(new Map());

  useEffect(() => {
    if (!open || activeIndex < 0) return;
    const el = lineRefs.current.get(activeIndex);
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [activeIndex, open]);

  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            className="fixed inset-0 z-40 bg-black/50"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.aside
            className="fixed right-0 top-0 z-50 flex h-full w-full max-w-md flex-col border-l border-[var(--color-border)] bg-[var(--color-surface)] shadow-2xl"
            initial={{ x: "100%" }}
            animate={{ x: 0 }}
            exit={{ x: "100%" }}
            transition={{ duration: 0.18 }}
          >
            <div className="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3">
              <h2 className="font-display text-sm font-bold tracking-widest text-[var(--color-primary)]">LYRICS</h2>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setBilingual(!bilingual)}
                  title="Toggle translation"
                  className={cn(
                    "rounded p-1.5 transition-colors",
                    bilingual
                      ? "text-[var(--color-primary)]"
                      : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
                  )}
                >
                  <Languages size={16} />
                </button>
                <button
                  type="button"
                  onClick={onClose}
                  className="rounded p-1.5 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
                >
                  <X size={16} />
                </button>
              </div>
            </div>

            <div ref={containerRef} className="flex-1 overflow-y-auto px-6 py-8">
              {!track ? (
                <p className="py-12 text-center text-sm text-[var(--color-text-muted)]">No track playing.</p>
              ) : loading ? (
                <p className="py-12 text-center text-sm text-[var(--color-text-muted)]">Loading lyrics…</p>
              ) : !lines || lines.length === 0 ? (
                <p className="py-12 text-center text-sm text-[var(--color-text-muted)]">No lyrics available.</p>
              ) : (
                <div className="space-y-5">
                  {lines.map((line, i) => (
                    <div
                      key={`${line.timestampMs}-${i}`}
                      ref={(el) => {
                        if (el) lineRefs.current.set(i, el);
                        else lineRefs.current.delete(i);
                      }}
                      className={cn(
                        "text-lg font-medium leading-relaxed transition-colors duration-300",
                        i === activeIndex
                          ? "text-[var(--color-primary)]"
                          : "text-[var(--color-text-muted)]"
                      )}
                    >
                      <LyricLineText line={line} active={i === activeIndex} positionMs={positionMs} />
                      {bilingual && line.translation && (
                        <p
                          className={cn(
                            "mt-1 text-sm font-normal",
                            i === activeIndex ? "text-[var(--color-text-secondary)]" : "text-[var(--color-text-muted)]"
                          )}
                        >
                          {line.translation}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </motion.aside>
        </>
      )}
    </AnimatePresence>
  );
}

function LyricLineText({ line, active, positionMs }: { line: LyricLine; active: boolean; positionMs: number }) {
  if (!line.words || line.words.length === 0) {
    return <p>{line.text}</p>;
  }
  return (
    <p>
      {line.words.map((word, i) => {
        const sung = active && positionMs >= word.offsetMs;
        return (
          <span
            key={`${word.offsetMs}-${i}`}
            className={sung ? "text-[var(--color-primary)]" : undefined}
          >
            {word.text}
          </span>
        );
      })}
    </p>
  );
}
