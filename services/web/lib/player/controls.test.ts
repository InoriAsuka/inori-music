import { describe, expect, it } from "vitest";
import {
  SPEED_PRESETS,
  DEFAULT_SPEED,
  isDefaultSpeed,
  formatSpeedLabel,
  formatSleepCountdown,
} from "./controls";
import {
  MIN_PLAYBACK_SPEED,
  MAX_PLAYBACK_SPEED,
  DEFAULT_PLAYBACK_SPEED,
} from "@/store/player";
import { SLEEP_TIMER_PRESET_MINUTES } from "@/store/sleepTimer";

/**
 * Unit tests for the pure player-control helpers. These carry the fixed
 * preset contracts (speed tiers, sleep durations) and the badge formatting
 * that the popover UI relies on.
 */

describe("SPEED_PRESETS", () => {
  it("is exactly the agreed tier list, mobile-aligned", () => {
    expect([...SPEED_PRESETS]).toEqual([0.5, 0.75, 1, 1.25, 1.5, 2]);
  });

  it("stays within the store's clamp bounds and includes the default", () => {
    for (const preset of SPEED_PRESETS) {
      expect(preset).toBeGreaterThanOrEqual(MIN_PLAYBACK_SPEED);
      expect(preset).toBeLessThanOrEqual(MAX_PLAYBACK_SPEED);
    }
    expect(SPEED_PRESETS).toContain(DEFAULT_PLAYBACK_SPEED);
    expect(DEFAULT_SPEED).toBe(DEFAULT_PLAYBACK_SPEED);
  });
});

describe("isDefaultSpeed", () => {
  it("is true only for 1×", () => {
    expect(isDefaultSpeed(1)).toBe(true);
    expect(isDefaultSpeed(0.5)).toBe(false);
    expect(isDefaultSpeed(1.5)).toBe(false);
    expect(isDefaultSpeed(2)).toBe(false);
  });
});

describe("formatSpeedLabel", () => {
  it("renders compact multiplier labels without trailing zeros", () => {
    expect(formatSpeedLabel(1)).toBe("1×");
    expect(formatSpeedLabel(1.5)).toBe("1.5×");
    expect(formatSpeedLabel(0.75)).toBe("0.75×");
    expect(formatSpeedLabel(2)).toBe("2×");
  });
});

describe("formatSleepCountdown", () => {
  it("renders m:ss, rounding up so a running timer never shows 0:00", () => {
    expect(formatSleepCountdown(60_000)).toBe("1:00");
    expect(formatSleepCountdown(90_000)).toBe("1:30");
    expect(formatSleepCountdown(1)).toBe("0:01");
    expect(formatSleepCountdown(59_400)).toBe("1:00"); // 59.4s → ceil 60
    expect(formatSleepCountdown(15 * 60_000)).toBe("15:00");
  });

  it("zero-pads seconds", () => {
    expect(formatSleepCountdown(9_000)).toBe("0:09");
    expect(formatSleepCountdown(65_000)).toBe("1:05");
  });

  it("renders 0:00 for null / non-positive input", () => {
    expect(formatSleepCountdown(null)).toBe("0:00");
    expect(formatSleepCountdown(undefined)).toBe("0:00");
    expect(formatSleepCountdown(0)).toBe("0:00");
    expect(formatSleepCountdown(-5_000)).toBe("0:00");
  });

  it("formats every fixed preset as a whole-minute countdown", () => {
    for (const minutes of SLEEP_TIMER_PRESET_MINUTES) {
      expect(formatSleepCountdown(minutes * 60_000)).toBe(`${minutes}:00`);
    }
  });
});
