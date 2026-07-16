/**
 * Pure helpers for the playback-speed and sleep-timer player controls.
 *
 * Kept free of React/DOM so they can be unit-tested under the repo's
 * node-environment Vitest setup (see vitest.config.ts). The stores own the
 * state machines; these just describe the fixed preset menus and format the
 * derived UI strings.
 */

/**
 * Playback-speed presets, exactly matching the mobile/Flutter tiers and the
 * store's [MIN_PLAYBACK_SPEED, MAX_PLAYBACK_SPEED] bounds.
 */
export const SPEED_PRESETS = [0.5, 0.75, 1, 1.25, 1.5, 2] as const;

/** The neutral rate; UI hides its speed badge only at this value. */
export const DEFAULT_SPEED = 1;

/** Whether a speed is the neutral 1× (badge/indicator suppressed). */
export function isDefaultSpeed(speed: number): boolean {
  return speed === DEFAULT_SPEED;
}

/**
 * Render a speed as a compact label, e.g. 1 → "1×", 1.5 → "1.5×",
 * 0.75 → "0.75×". Trailing zeros are dropped by Number's own formatting.
 */
export function formatSpeedLabel(speed: number): string {
  return `${speed}×`;
}

/**
 * Format a fixed-timer remaining duration (ms) as m:ss for the countdown
 * badge. Rounds UP to the next whole second so the badge never shows 0:00
 * while time still remains. Non-positive / null input renders "0:00".
 */
export function formatSleepCountdown(remainingMs: number | null | undefined): string {
  if (remainingMs == null || remainingMs <= 0) return "0:00";
  const totalSeconds = Math.ceil(remainingMs / 1000);
  const m = Math.floor(totalSeconds / 60);
  const s = totalSeconds % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}
