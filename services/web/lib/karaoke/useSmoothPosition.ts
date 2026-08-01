"use client";

import { useEffect, useRef, useState } from "react";
import { usePlayerStore } from "@/store/player";

/**
 * Smooth playback position for karaoke word highlighting.
 *
 * The player store only ticks every 250ms, which makes per-word highlighting
 * visibly steppy. This interpolates between store updates with rAF: each store
 * tick re-anchors the baseline, and frames in between extrapolate from the
 * wall clock scaled by playback speed. Only runs while `active`.
 */
export function useSmoothPosition(active: boolean): number {
  const positionSeconds = usePlayerStore((s) => s.positionSeconds);
  const status = usePlayerStore((s) => s.status);
  const speed = usePlayerStore((s) => s.speed);

  const [smoothMs, setSmoothMs] = useState(positionSeconds * 1000);
  const anchorRef = useRef({ positionMs: positionSeconds * 1000, atMs: 0 });

  // Re-anchor whenever the authoritative position lands.
  useEffect(() => {
    anchorRef.current = { positionMs: positionSeconds * 1000, atMs: performance.now() };
    setSmoothMs(positionSeconds * 1000);
  }, [positionSeconds]);

  useEffect(() => {
    if (!active || status !== "playing") return;
    let frame = 0;
    const tick = () => {
      const { positionMs, atMs } = anchorRef.current;
      setSmoothMs(positionMs + (performance.now() - atMs) * speed);
      frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [active, status, speed]);

  return smoothMs;
}
