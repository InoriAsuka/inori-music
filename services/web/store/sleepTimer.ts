/**
 * Sleep timer store — non-persisted, session-scoped playback auto-stop.
 *
 * Two mutually-exclusive modes, mirroring the Flutter `SleepTimerNotifier`:
 *   • "fixed"       — a 15/30/45/60-minute countdown; on expiry the player is
 *                     paused and the timer clears itself.
 *   • "after-track" — pause once the currently playing track ends (wired into
 *                     useAudio's `ended` path, which must NOT advance/repeat/
 *                     crossfade past that track — see useAudio integration).
 *
 * Deliberately NOT persisted: a sleep timer is a short-lived, in-session
 * intent. A page refresh clears it (see the plan's explicit non-goal).
 *
 * The interval handle for the fixed countdown lives at module scope, never in
 * the serialisable store state, so state stays a plain snapshot and cleanup is
 * deterministic (every state transition routes through `stopTicker`).
 */
"use client";

import { create } from "zustand";
import { usePlayerStore } from "@/store/player";

/** Fixed-duration presets offered by the UI, in minutes. */
export const SLEEP_TIMER_PRESET_MINUTES = [15, 30, 45, 60] as const;

export type SleepTimerMode = "off" | "fixed" | "after-track";

/** How often the fixed-countdown ticker updates remaining time (ms). */
const TICK_INTERVAL_MS = 1000;

interface SleepTimerState {
  /** Current mode; "off" means no timer armed. */
  mode: SleepTimerMode;
  /** True while a timer (either mode) is armed. */
  active: boolean;
  /** Original duration selected for a fixed timer; null for after-track/off. */
  presetMinutes: number | null;
  /** Epoch ms at which a fixed timer fires; null for after-track/off. */
  endsAtMs: number | null;
  /**
   * Remaining time for a fixed countdown in ms (clamped ≥ 0), refreshed each
   * tick for the UI. Null for after-track/off. UI derives seconds from this.
   */
  remainingMs: number | null;

  /** Arm a fixed-duration countdown (minutes). Replaces any existing timer. */
  startFixed: (minutes: number) => void;
  /** Arm after-current-track mode. Replaces any existing timer. */
  startAfterTrack: () => void;
  /** Cancel any armed timer and reset to "off". */
  cancel: () => void;
  /**
   * Recompute remaining time for a fixed countdown; on expiry, pause the
   * player and clear the timer. Driven by the internal ticker and safe to
   * call directly (used by tests / manual ticks). No-op unless a fixed timer
   * is armed.
   */
  tick: () => void;
  /**
   * Called from useAudio when the active track ends. If after-track mode is
   * armed, pause the player, clear the timer, and return true so the caller
   * skips advancing/repeating/crossfading. Returns false otherwise.
   */
  handleTrackEnded: () => boolean;
}

const IDLE_STATE = {
  mode: "off" as SleepTimerMode,
  active: false,
  presetMinutes: null,
  endsAtMs: null,
  remainingMs: null,
};

/**
 * Module-scoped interval handle for the fixed countdown. Kept out of store
 * state so the snapshot stays serialisable and cleanup is deterministic.
 */
let tickerHandle: ReturnType<typeof setInterval> | null = null;

function stopTicker() {
  if (tickerHandle !== null) {
    clearInterval(tickerHandle);
    tickerHandle = null;
  }
}

export const useSleepTimerStore = create<SleepTimerState>((set, get) => ({
  ...IDLE_STATE,

  startFixed(minutes) {
    stopTicker();
    const durationMs = Math.max(0, minutes) * 60_000;
    const endsAtMs = Date.now() + durationMs;
    set({ mode: "fixed", active: true, presetMinutes: minutes, endsAtMs, remainingMs: durationMs });
    tickerHandle = setInterval(() => get().tick(), TICK_INTERVAL_MS);
  },

  startAfterTrack() {
    stopTicker();
    set({ mode: "after-track", active: true, presetMinutes: null, endsAtMs: null, remainingMs: null });
  },

  cancel() {
    stopTicker();
    set({ ...IDLE_STATE });
  },

  tick() {
    const { mode, endsAtMs } = get();
    if (mode !== "fixed" || endsAtMs === null) return;
    const remaining = endsAtMs - Date.now();
    if (remaining <= 0) {
      stopTicker();
      set({ ...IDLE_STATE });
      usePlayerStore.getState().pause();
      return;
    }
    set({ remainingMs: remaining });
  },

  handleTrackEnded() {
    const { mode, active } = get();
    if (mode !== "after-track" || !active) return false;
    stopTicker();
    set({ ...IDLE_STATE });
    usePlayerStore.getState().pause();
    return true;
  },
}));

/** Whether after-current-track auto-stop is currently armed. */
export function isAfterTrackArmed(state: SleepTimerState): boolean {
  return state.mode === "after-track" && state.active;
}
