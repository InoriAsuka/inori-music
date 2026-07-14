/**
 * ReplayGain — pure gain-calculation helpers + the on/off toggle persistence.
 *
 * Formula mirrors the Flutter client (services/mobile/lib/src/audio/replay_gain_notifier.dart
 * + player_notifier.dart _applyVolumeWithGain): linear gain = 10^(db/20), clamped to
 * [0.1, 2.0] to avoid silence or ear-damaging overshoot from bad analysis data.
 * Applied orthogonally to user volume (audio.volume) via a WebAudio GainNode.
 */
"use client";

const STORAGE_KEY = "inori.audio.replayGainEnabled";
const GAIN_MIN = 0.1;
const GAIN_MAX = 2.0;

/**
 * Computes the linear gain multiplier for a given ReplayGain dB value.
 * Returns 1.0 (unity, no-op) when `db` is null/undefined/NaN — mirrors the
 * Flutter client's fallback when a track hasn't been analyzed yet.
 */
export function computeReplayGain(db: number | null | undefined): number {
  if (db === null || db === undefined || Number.isNaN(db)) return 1.0;
  const linear = 10 ** (db / 20);
  return Math.min(GAIN_MAX, Math.max(GAIN_MIN, linear));
}

/** Reads the persisted ReplayGain toggle. Defaults to `false` (off). */
export function isReplayGainEnabled(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return localStorage.getItem(STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

/** Persists the ReplayGain toggle. */
export function setReplayGainEnabled(enabled: boolean): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, String(enabled));
  } catch {
    // Storage unavailable (private mode / quota) — non-fatal, toggle just won't persist.
  }
}

/** Fired on `window` whenever the toggle changes, so live audio graphs can react. */
export const REPLAY_GAIN_CHANGE_EVENT = "inori:replaygain-change";

export function dispatchReplayGainChange(enabled: boolean): void {
  if (typeof window === "undefined") return;
  window.dispatchEvent(new CustomEvent<boolean>(REPLAY_GAIN_CHANGE_EVENT, { detail: enabled }));
}
