import { describe, expect, it } from "vitest";
import {
  canFinalizeReservedSlot,
  isPreloadedUrlStale,
  isSlotReusable,
  isStandbyReadyForOccurrence,
  occurrencesMatch,
  PRELOAD_PROGRESS_RATIO_THRESHOLD,
  PRELOAD_REMAINING_SECONDS_THRESHOLD,
  PRESIGNED_URL_SAFETY_MARGIN_MS,
  PRESIGNED_URL_TTL_MS,
  shouldRestartRepeatOne,
  shouldStartNaturalCrossfade,
  shouldTriggerPreload,
} from "./gaplessEngine";

describe("shouldTriggerPreload", () => {
  it("does not trigger before 50% progress with >30s remaining", () => {
    // 200s track, 60s in: 30% progress, 140s remaining — neither condition met.
    expect(shouldTriggerPreload(60, 200)).toBe(false);
  });

  it("triggers once progress exceeds 50%", () => {
    // 200s track, 101s in: 50.5% progress.
    expect(shouldTriggerPreload(101, 200)).toBe(true);
  });

  it("does not trigger at exactly 50% progress (strict >)", () => {
    expect(shouldTriggerPreload(100, 200)).toBe(false);
  });

  it("triggers when remaining time drops below 30s even if progress is under 50%", () => {
    // 1000s track, 975s in: 2.5% progress, but only 25s remaining.
    expect(shouldTriggerPreload(975, 1000)).toBe(true);
  });

  it("does not trigger at exactly 30s remaining (strict <)", () => {
    // 60s track, 30s in: 50% progress boundary AND 30s remaining boundary — neither strictly met.
    expect(shouldTriggerPreload(30, 60)).toBe(false);
  });

  it("returns false for zero duration (metadata not loaded yet)", () => {
    expect(shouldTriggerPreload(0, 0)).toBe(false);
  });

  it("returns false for negative/invalid duration", () => {
    expect(shouldTriggerPreload(10, -5)).toBe(false);
    expect(shouldTriggerPreload(10, Number.NaN)).toBe(false);
  });

  it("returns false for negative currentTime", () => {
    expect(shouldTriggerPreload(-1, 200)).toBe(false);
  });

  it("uses the documented threshold constants", () => {
    expect(PRELOAD_PROGRESS_RATIO_THRESHOLD).toBe(0.5);
    expect(PRELOAD_REMAINING_SECONDS_THRESHOLD).toBe(30);
  });
});

describe("isPreloadedUrlStale", () => {
  const NOW = 1_000_000_000_000; // arbitrary fixed epoch ms

  it("is not stale immediately after resolving", () => {
    expect(isPreloadedUrlStale(NOW, NOW)).toBe(false);
  });

  it("is not stale a few minutes after resolving", () => {
    expect(isPreloadedUrlStale(NOW, NOW + 5 * 60 * 1000)).toBe(false);
  });

  it("is stale once within the safety margin of the 15-minute TTL", () => {
    const resolvedAt = NOW;
    const justInsideMargin = NOW + PRESIGNED_URL_TTL_MS - PRESIGNED_URL_SAFETY_MARGIN_MS;
    expect(isPreloadedUrlStale(resolvedAt, justInsideMargin)).toBe(true);
  });

  it("is not stale just before the safety margin kicks in", () => {
    const resolvedAt = NOW;
    const justBeforeMargin = NOW + PRESIGNED_URL_TTL_MS - PRESIGNED_URL_SAFETY_MARGIN_MS - 1;
    expect(isPreloadedUrlStale(resolvedAt, justBeforeMargin)).toBe(false);
  });

  it("is stale for an invalid/zero resolvedAt", () => {
    expect(isPreloadedUrlStale(0, NOW)).toBe(true);
    expect(isPreloadedUrlStale(-1, NOW)).toBe(true);
    expect(isPreloadedUrlStale(Number.NaN, NOW)).toBe(true);
  });

  it("is stale long after resolving (e.g. long paused track)", () => {
    expect(isPreloadedUrlStale(NOW, NOW + 20 * 60 * 1000)).toBe(true);
  });
});

describe("queue occurrence readiness", () => {
  const NOW = 1_000_000_000_000;
  const firstOccurrence = { queueIndex: 0, trackId: "duplicate" };
  const secondOccurrence = { queueIndex: 1, trackId: "duplicate" };

  it("does not equate duplicate track IDs at different queue positions", () => {
    expect(occurrencesMatch(firstOccurrence, secondOccurrence)).toBe(false);
  });

  it("requires the intended occurrence and a completed media load", () => {
    expect(
      isStandbyReadyForOccurrence(
        { occurrence: firstOccurrence, ready: false, resolvedAtMs: NOW },
        firstOccurrence,
        NOW
      )
    ).toBe(false);
    expect(
      isStandbyReadyForOccurrence(
        { occurrence: firstOccurrence, ready: true, resolvedAtMs: NOW },
        secondOccurrence,
        NOW
      )
    ).toBe(false);
  });

  it("accepts only a fresh, ready slot for the exact intended occurrence", () => {
    expect(
      isStandbyReadyForOccurrence(
        { occurrence: secondOccurrence, ready: true, resolvedAtMs: NOW },
        secondOccurrence,
        NOW + 1000
      )
    ).toBe(true);
  });
});

describe("natural-end crossfade timing", () => {
  it("starts at the configured lead time when standby is ready", () => {
    expect(shouldStartNaturalCrossfade(98, 100, 2, true)).toBe(true);
    expect(shouldStartNaturalCrossfade(98.5, 100, 2, true)).toBe(true);
  });

  it("does not start early or after the track already ended", () => {
    expect(shouldStartNaturalCrossfade(97.9, 100, 2, true)).toBe(false);
    expect(shouldStartNaturalCrossfade(100, 100, 2, true)).toBe(false);
  });

  it("keeps ended as fallback when standby is not ready", () => {
    expect(shouldStartNaturalCrossfade(99, 100, 2, false)).toBe(false);
  });
});

describe("repeat-one transition decision", () => {
  it("restarts only for repeat-one after ended", () => {
    expect(shouldRestartRepeatOne("one", "ended")).toBe(true);
    expect(shouldRestartRepeatOne("one", "natural-crossfade")).toBe(false);
    expect(shouldRestartRepeatOne("off", "ended")).toBe(false);
    expect(shouldRestartRepeatOne("all", "ended")).toBe(false);
  });
});

describe("fade-out slot reuse reservation", () => {
  it("blocks only the reserved fading generation", () => {
    expect(isSlotReusable(7, 7)).toBe(false);
    expect(isSlotReusable(7, 8)).toBe(true);
    expect(isSlotReusable(null, 7)).toBe(true);
  });
});

describe("reserved source finalization guard", () => {
  it("allows finalization only for the exact inactive reserved generation", () => {
    expect(canFinalizeReservedSlot(0, 1, 7, 7, 7)).toBe(true);
  });

  it("makes delayed cleanup harmless after immediate rejection cleanup clears the reservation", () => {
    expect(canFinalizeReservedSlot(0, 1, 7, 7, null)).toBe(false);
  });

  it("blocks finalization after rapid navigation or slot reuse", () => {
    expect(canFinalizeReservedSlot(0, 0, 7, 7, 7)).toBe(false);
    expect(canFinalizeReservedSlot(0, 1, 7, 8, 7)).toBe(false);
  });
});
