import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useSleepTimerStore, isAfterTrackArmed, SLEEP_TIMER_PRESET_MINUTES } from "./sleepTimer";
import { usePlayerStore } from "./player";
import { useAuthStore } from "./auth";

/**
 * Unit tests for the sleep timer store state machine.
 *
 * Uses Vitest fake timers so fixed-duration countdowns can be advanced
 * deterministically. Mirrors the mobile `SleepTimerNotifier` test coverage:
 * fixed-duration expiry, after-track mode, cancellation, and repeated-set
 * replacement.
 *
 * The player's `pause()` is replaced with a spy via `setState` (not
 * `vi.spyOn(getState(), ...)`): zustand's `set` spreads state into fresh
 * objects that carry a spied method forward, so `restoreAllMocks` cannot
 * clean the live store. We install and restore the real action explicitly.
 */

const realPause = usePlayerStore.getState().pause;
let pauseMock: ReturnType<typeof vi.fn<() => void>>;

function resetStores() {
  useSleepTimerStore.getState().cancel();
  pauseMock = vi.fn<() => void>();
  usePlayerStore.setState({ status: "playing", pause: pauseMock });
}

beforeEach(() => {
  vi.useFakeTimers();
  resetStores();
});

afterEach(() => {
  // Ensure no ticker survives a test, then restore the real player action.
  useSleepTimerStore.getState().cancel();
  vi.clearAllTimers();
  vi.useRealTimers();
  usePlayerStore.setState({ pause: realPause });
});

describe("initial state", () => {
  it("is idle with no timer armed", () => {
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("off");
    expect(s.active).toBe(false);
    expect(s.presetMinutes).toBeNull();
    expect(s.endsAtMs).toBeNull();
    expect(s.remainingMs).toBeNull();
  });

  it("exposes the plan's fixed presets", () => {
    expect(SLEEP_TIMER_PRESET_MINUTES).toEqual([15, 30, 45, 60]);
  });
});

describe("startFixed", () => {
  it("arms a countdown with the requested duration", () => {
    useSleepTimerStore.getState().startFixed(30);
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("fixed");
    expect(s.active).toBe(true);
    expect(s.presetMinutes).toBe(30);
    expect(s.remainingMs).toBe(30 * 60_000);
    expect(s.endsAtMs).not.toBeNull();
  });

  it("ticks down remaining time as the clock advances", () => {
    useSleepTimerStore.getState().startFixed(1);
    vi.advanceTimersByTime(15_000);
    // ~45s left after 15s of a 60s timer.
    expect(useSleepTimerStore.getState().remainingMs).toBeLessThanOrEqual(45_000);
    expect(useSleepTimerStore.getState().remainingMs).toBeGreaterThan(40_000);
    expect(useSleepTimerStore.getState().active).toBe(true);
    expect(useSleepTimerStore.getState().presetMinutes).toBe(1);
  });

  it("pauses the player and clears itself on expiry", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(1);
    vi.advanceTimersByTime(60_000);
    expect(pause).toHaveBeenCalledTimes(1);
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("off");
    expect(s.active).toBe(false);
    expect(s.presetMinutes).toBeNull();
    expect(s.remainingMs).toBeNull();
    expect(s.endsAtMs).toBeNull();
  });

  it("does not keep pausing after expiry (ticker is stopped)", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(1);
    vi.advanceTimersByTime(120_000);
    expect(pause).toHaveBeenCalledTimes(1);
  });
});

describe("startAfterTrack", () => {
  it("arms after-current-track mode without a countdown", () => {
    useSleepTimerStore.getState().startAfterTrack();
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("after-track");
    expect(s.active).toBe(true);
    expect(s.presetMinutes).toBeNull();
    expect(s.remainingMs).toBeNull();
    expect(s.endsAtMs).toBeNull();
    expect(isAfterTrackArmed(s)).toBe(true);
  });

  it("does not pause on a timer tick (only on track end)", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startAfterTrack();
    vi.advanceTimersByTime(600_000);
    expect(pause).not.toHaveBeenCalled();
    expect(useSleepTimerStore.getState().active).toBe(true);
  });
});

describe("handleTrackEnded", () => {
  it("pauses, clears, and returns true when after-track is armed", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startAfterTrack();
    const consumed = useSleepTimerStore.getState().handleTrackEnded();
    expect(consumed).toBe(true);
    expect(pause).toHaveBeenCalledTimes(1);
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("off");
    expect(s.active).toBe(false);
  });

  it("returns false and does nothing when no timer is armed", () => {
    const pause = pauseMock;
    const consumed = useSleepTimerStore.getState().handleTrackEnded();
    expect(consumed).toBe(false);
    expect(pause).not.toHaveBeenCalled();
  });

  it("returns false in fixed mode (fixed does not stop on track end)", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(30);
    const consumed = useSleepTimerStore.getState().handleTrackEnded();
    expect(consumed).toBe(false);
    expect(pause).not.toHaveBeenCalled();
    // Fixed countdown remains armed.
    expect(useSleepTimerStore.getState().mode).toBe("fixed");
  });
});

describe("cancel", () => {
  it("resets state and stops the countdown", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(1);
    useSleepTimerStore.getState().cancel();
    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("off");
    expect(s.active).toBe(false);
    expect(s.remainingMs).toBeNull();
    // Confirm the ticker does not fire after cancel.
    vi.advanceTimersByTime(120_000);
    expect(pause).not.toHaveBeenCalled();
  });

  it("clears an armed after-track mode", () => {
    useSleepTimerStore.getState().startAfterTrack();
    useSleepTimerStore.getState().cancel();
    expect(useSleepTimerStore.getState().mode).toBe("off");
    expect(useSleepTimerStore.getState().handleTrackEnded()).toBe(false);
  });
});

describe("repeated setting replacement", () => {
  it("replacing a fixed timer resets remaining and only fires once", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(100); // long
    useSleepTimerStore.getState().startFixed(1); // replace with 60s
    expect(useSleepTimerStore.getState().remainingMs).toBe(60_000);
    vi.advanceTimersByTime(60_000);
    expect(pause).toHaveBeenCalledTimes(1);
    expect(useSleepTimerStore.getState().active).toBe(false);
  });

  it("switching from fixed to after-track stops the countdown", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(1);
    useSleepTimerStore.getState().startAfterTrack();
    expect(useSleepTimerStore.getState().mode).toBe("after-track");
    // The old fixed ticker must not fire.
    vi.advanceTimersByTime(120_000);
    expect(pause).not.toHaveBeenCalled();
  });

  it("switching from after-track to fixed drops after-track semantics", () => {
    useSleepTimerStore.getState().startAfterTrack();
    useSleepTimerStore.getState().startFixed(30);
    expect(useSleepTimerStore.getState().mode).toBe("fixed");
    expect(useSleepTimerStore.getState().handleTrackEnded()).toBe(false);
  });
});

describe("session boundary (logout)", () => {
  it("clearSession cancels a fixed timer so it never crosses the session boundary", () => {
    const pause = pauseMock;
    useSleepTimerStore.getState().startFixed(30);
    expect(useSleepTimerStore.getState().active).toBe(true);

    useAuthStore.getState().clearSession();

    const s = useSleepTimerStore.getState();
    expect(s.mode).toBe("off");
    expect(s.active).toBe(false);
    expect(s.remainingMs).toBeNull();
    // The module-scoped ticker must be stopped — no pause after logout.
    vi.advanceTimersByTime(60_000);
    expect(pause).not.toHaveBeenCalled();
  });

  it("clearSession disarms after-track mode", () => {
    useSleepTimerStore.getState().startAfterTrack();
    useAuthStore.getState().clearSession();
    expect(useSleepTimerStore.getState().mode).toBe("off");
    expect(useSleepTimerStore.getState().handleTrackEnded()).toBe(false);
  });
});
