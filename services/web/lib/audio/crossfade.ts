/**
 * Crossfade — persisted on/off toggle for the gapless engine's optional
 * linear-ramp crossfade between tracks (behind the same settings toggle
 * mentioned by the plan: "设置项共享同一开关文案").
 *
 * Unlike the Flutter client's 0-8s continuous slider, the web engine uses a
 * fixed short ramp (CROSSFADE_SECONDS) — simple boolean on/off keeps the
 * dual-element swap logic easy to reason about while still delivering the
 * audible smoothing effect.
 */
"use client";

const STORAGE_KEY = "inori.audio.crossfadeEnabled";

/** Fixed crossfade ramp duration in seconds, applied when the toggle is on. */
export const CROSSFADE_SECONDS = 2;

export function isCrossfadeEnabled(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return localStorage.getItem(STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

export function setCrossfadeEnabled(enabled: boolean): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, String(enabled));
  } catch {
    // Storage unavailable — non-fatal.
  }
}
