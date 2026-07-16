import { describe, expect, it } from "vitest";
import {
  PRELOAD_PROGRESS_RATIO_THRESHOLD,
  PRELOAD_REMAINING_SECONDS_THRESHOLD,
  PRESIGNED_URL_SAFETY_MARGIN_MS,
  PRESIGNED_URL_TTL_MS,
  type PlaybackCycle,
  canEndedAdvance,
  canFinalizeReservedSlot,
  canSettleActivePlay,
  cyclesMatch,
  hasCurrentEndedEvidence,
  isPreloadedUrlStale,
  isSlotReusable,
  isStandbyReadyForOccurrence,
  occurrencesMatch,
  shouldOpenFreshEndedCycle,
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

describe("active play settlement guard", () => {
  it("accepts the exact active generation while playback is requested", () => {
    expect(canSettleActivePlay("loading", 1, 1, 8, 8, 3, 3)).toBe(true);
    expect(canSettleActivePlay("playing", 1, 1, 8, 8, 3, 3)).toBe(true);
  });

  it("rejects a settlement after pause or fixed timer expiry", () => {
    expect(canSettleActivePlay("paused", 1, 1, 8, 8, 3, 3)).toBe(false);
  });

  it("rejects stale destinations after rapid slot or generation changes", () => {
    expect(canSettleActivePlay("loading", 0, 1, 8, 8, 3, 3)).toBe(false);
    expect(canSettleActivePlay("loading", 1, 1, 8, 9, 3, 3)).toBe(false);
    expect(canSettleActivePlay("loading", 1, 1, 8, 8, 3, 4)).toBe(false);
  });
});

describe("ended-event advance guard", () => {
  // A playback cycle is (loadId, playGen). loadId is the media-load identity;
  // playGen distinguishes ended cycles of the SAME loaded media across replays.
  const cycle = (loadId: number, playGen = 0): PlaybackCycle => ({ loadId, playGen });
  const SOURCE = cycle(42);

  it("advances on a normal ended while still playing", () => {
    // Natural end of the active track: player is still "playing", the source
    // cycle has neither been consumed nor transitioned away.
    expect(canEndedAdvance("playing", SOURCE, null, null)).toBe(true);
  });

  it("ignores a late ended after a fixed sleep-timer shutdown (scenario 1)", () => {
    // tick() expiry paused both slots synchronously; a delayed ended on the
    // still-active source must not advance/prepare/play the next track.
    expect(canEndedAdvance("paused", SOURCE, null, null)).toBe(false);
    // Idle/error shutdowns are equally inert.
    expect(canEndedAdvance("idle", SOURCE, null, null)).toBe(false);
    expect(canEndedAdvance("error", SOURCE, null, null)).toBe(false);
  });

  it("ignores a duplicate after-current-track ended on the consumed cycle (scenario 2)", () => {
    // The first after-track ended marked this cycle consumed and paused the
    // player; a duplicate ended on the same element and cycle must be inert
    // even if a stale "playing" status were still observed.
    expect(canEndedAdvance("playing", SOURCE, SOURCE, null)).toBe(false);
    expect(canEndedAdvance("paused", SOURCE, SOURCE, null)).toBe(false);
  });

  it("still advances a fresh media load whose predecessor was consumed", () => {
    // A different, later media-load cycle is unaffected by a prior consumed one.
    expect(canEndedAdvance("playing", cycle(43), SOURCE, null)).toBe(true);
  });

  it("advances a REPLAYED cycle of the same media after its ended was consumed (after-track replay)", () => {
    // Defect 1: after-track ended consumed cycle (42,0) and paused. The user
    // presses play; the SAME media/loadId replays but its playback cycle was
    // reopened to (42,1). Its legitimate next ended must advance — the consumed
    // record for (42,0) no longer matches.
    const consumed = cycle(42, 0);
    const replayed = cycle(42, 1);
    expect(canEndedAdvance("playing", replayed, consumed, null)).toBe(true);
  });

  it("blocks a second advance of a cycle already transitioned via crossfade", () => {
    expect(canEndedAdvance("playing", SOURCE, null, SOURCE)).toBe(false);
    // A different media load than the last transition is still allowed.
    expect(canEndedAdvance("playing", SOURCE, null, cycle(43))).toBe(true);
  });

  it("advances a REPLAYED cycle of the same media after end-of-queue transition (end-of-queue replay)", () => {
    // Defect 2: end-of-queue set lastTransition to (42,0) and went idle. Play
    // replays the same media/loadId with reopened cycle (42,1); its next ended
    // must advance because the transition record for (42,0) no longer matches.
    const transitioned = cycle(42, 0);
    const replayed = cycle(42, 1);
    expect(canEndedAdvance("playing", replayed, null, transitioned)).toBe(true);
  });

  it("does not advance from a mid-load state (only a live playing source ends)", () => {
    expect(canEndedAdvance("loading", SOURCE, null, null)).toBe(false);
  });
});

describe("playback-cycle identity", () => {
  it("matches only when both media-load and cycle generation are equal", () => {
    expect(cyclesMatch({ loadId: 5, playGen: 0 }, { loadId: 5, playGen: 0 })).toBe(true);
    expect(cyclesMatch({ loadId: 5, playGen: 0 }, { loadId: 5, playGen: 1 })).toBe(false);
    expect(cyclesMatch({ loadId: 5, playGen: 0 }, { loadId: 6, playGen: 0 })).toBe(false);
  });

  it("never matches a null cycle (no record yet)", () => {
    expect(cyclesMatch(null, { loadId: 5, playGen: 0 })).toBe(false);
    expect(cyclesMatch({ loadId: 5, playGen: 0 }, null)).toBe(false);
    expect(cyclesMatch(null, null)).toBe(false);
  });
});

describe("explicit-resume replay boundary", () => {
  const live: PlaybackCycle = { loadId: 42, playGen: 0 };

  it("reopens the cycle when resuming media whose ended was consumed (after-track defect)", () => {
    // The live cycle equals the after-track consumed record → a fresh cycle
    // must be minted so the replayed media's next ended can advance.
    expect(shouldOpenFreshEndedCycle(live, live, null)).toBe(true);
  });

  it("reopens the cycle when resuming media transitioned off at end-of-queue (end-of-queue defect)", () => {
    expect(shouldOpenFreshEndedCycle(live, null, live)).toBe(true);
  });

  it("does NOT reopen on a normal mid-track pause→play", () => {
    // Neither the consumed nor the transition record holds this live cycle:
    // no ended terminated it, so replay must not bump the cycle. This keeps a
    // genuinely-old async ended from being revived into a valid new end.
    expect(shouldOpenFreshEndedCycle(live, null, null)).toBe(false);
    // Records from OTHER cycles must not trigger a reopen either.
    expect(shouldOpenFreshEndedCycle(live, { loadId: 41, playGen: 0 }, { loadId: 40, playGen: 3 })).toBe(false);
  });

  it("does NOT reopen after a fixed-timer expiry resume", () => {
    // A fixed timer pauses via pause(), never consuming/transitioning an ended,
    // so neither record holds the live cycle: playGen stays put and the delayed
    // ended remains inert exactly as before.
    expect(shouldOpenFreshEndedCycle(live, null, null)).toBe(false);
  });

  it("does not reopen a cycle already reopened once (records hold the OLD generation)", () => {
    // After one reopen the live cycle is (42,1) while records still hold (42,0):
    // resuming again (e.g. paused mid-replay) must not bump a second time.
    const reopened: PlaybackCycle = { loadId: 42, playGen: 1 };
    const stale: PlaybackCycle = { loadId: 42, playGen: 0 };
    expect(shouldOpenFreshEndedCycle(reopened, stale, stale)).toBe(false);
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

/**
 * End-to-end ended-cycle simulations. Vitest runs in a `node` environment with
 * no DOM/Audio, so the hook cannot be mounted (Playwright e2e covers that).
 * Instead we mirror the exact ref-update ordering useAudio performs around an
 * `ended` and an explicit resume, driving the SAME pure guards the hook calls.
 * This models the two verified replay defects deterministically.
 */
describe("native ended evidence", () => {
  it("accepts a trusted native event while the element still reports ended", () => {
    expect(hasCurrentEndedEvidence(true, true)).toBe(true);
  });

  it("rejects a trusted event queued before resume or source replacement", () => {
    expect(hasCurrentEndedEvidence(true, false)).toBe(false);
  });

  it("allows synthetic ended events used by the Playwright playback probe", () => {
    expect(hasCurrentEndedEvidence(false, false)).toBe(true);
  });
});

describe("ended-cycle replay simulation (mirrors useAudio ref ordering)", () => {
  /** Minimal mirror of the hook's active-slot + guard-ref state. */
  interface Harness {
    status: "idle" | "loading" | "playing" | "paused" | "error";
    active: PlaybackCycle;
    consumed: PlaybackCycle | null;
    transitioned: PlaybackCycle | null;
  }

  /** The ended handler's decision + the after-track consume side effect. */
  function onEnded(h: Harness, afterTrackArmed: boolean): "advanced" | "consumed" | "inert" {
    if (!canEndedAdvance(h.status, h.active, h.consumed, h.transitioned)) return "inert";
    if (afterTrackArmed) {
      // handleTrackEnded() pauses synchronously; hook records consumed cycle.
      h.consumed = { ...h.active };
      h.status = "paused";
      return "consumed";
    }
    return "advanced";
  }

  /** End-of-queue advance: hook records the transition and goes idle. */
  function endOfQueueEnded(h: Harness): "advanced" | "inert" {
    if (!canEndedAdvance(h.status, h.active, h.consumed, h.transitioned)) return "inert";
    h.transitioned = { ...h.active };
    h.status = "idle"; // skipToNext() past the last track clears playback.
    return "advanced";
  }

  /** The Play/Pause status effect resume boundary + the pending replay. */
  function resume(h: Harness): void {
    if (shouldOpenFreshEndedCycle(h.active, h.consumed, h.transitioned)) {
      h.active = { ...h.active, playGen: h.active.playGen + 1 };
    }
    h.status = "playing";
  }

  it("Defect 1: after-track ended → resume → next ended advances the fresh cycle", () => {
    const h: Harness = { status: "playing", active: { loadId: 42, playGen: 0 }, consumed: null, transitioned: null };

    // First (after-track) ended: consumes cycle (42,0), pauses.
    expect(onEnded(h, true)).toBe("consumed");
    // A duplicate late ended BEFORE resume is inert (same cycle, now paused).
    expect(onEnded(h, false)).toBe("inert");

    // User presses play — same media/loadId replays; boundary reopens to (42,1).
    resume(h);
    expect(h.active).toEqual({ loadId: 42, playGen: 1 });

    // The replayed media's legitimate next ended (no timer this time) advances.
    expect(onEnded(h, false)).toBe("advanced");
  });

  it("Defect 2: end-of-queue ended → resume → next ended advances the fresh cycle", () => {
    const h: Harness = { status: "playing", active: { loadId: 42, playGen: 0 }, consumed: null, transitioned: null };

    // End-of-queue ended: records transition (42,0), goes idle.
    expect(endOfQueueEnded(h)).toBe("advanced");
    // A duplicate ended before resume is inert.
    expect(endOfQueueEnded(h)).toBe("inert");

    // User presses play — same media/loadId replays; boundary reopens to (42,1).
    resume(h);
    expect(h.active).toEqual({ loadId: 42, playGen: 1 });

    // The replayed media's next end-of-queue ended advances again.
    expect(endOfQueueEnded(h)).toBe("advanced");
  });

  it("Invariant: normal mid-track pause→play does NOT reopen or revive a stale ended", () => {
    const h: Harness = { status: "playing", active: { loadId: 42, playGen: 0 }, consumed: null, transitioned: null };

    // User pauses mid-track (no ended fired); a stray async ended is inert.
    h.status = "paused";
    expect(onEnded(h, false)).toBe("inert");

    // Resume: no ended terminated the cycle, so playGen must NOT bump.
    resume(h);
    expect(h.active).toEqual({ loadId: 42, playGen: 0 });

    // A genuine natural end now advances exactly once, as normal.
    expect(onEnded(h, false)).toBe("advanced");
  });

  it("Invariant: fixed-timer expiry resume does NOT reopen the cycle (delayed ended stays inert)", () => {
    const h: Harness = { status: "playing", active: { loadId: 42, playGen: 0 }, consumed: null, transitioned: null };

    // Fixed timer fires: pause() synchronously, WITHOUT consuming an ended.
    h.status = "paused";
    // Late ended on the still-active source is inert (paused).
    expect(onEnded(h, false)).toBe("inert");

    // Resume: neither record holds the cycle → no bump.
    resume(h);
    expect(h.active).toEqual({ loadId: 42, playGen: 0 });
  });
});
