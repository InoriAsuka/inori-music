/**
 * Pure helper functions for the dual-element gapless engine — extracted from
 * useAudio.ts so preload-trigger and URL-expiry logic can be unit tested
 * without mounting React / constructing real HTMLAudioElements.
 */

/** Preload the next track once current progress exceeds 50% or remaining time drops below 30s. */
export const PRELOAD_REMAINING_SECONDS_THRESHOLD = 30;
export const PRELOAD_PROGRESS_RATIO_THRESHOLD = 0.5;

/**
 * Returns true when playback of the current track has progressed far enough
 * that the next track should start preloading into the standby element.
 * Mirrors the plan's "进度 > 50% 或剩余 < 30s" condition. Guards against
 * unknown/zero duration (metadata not loaded yet) by requiring a positive
 * duration before either condition can fire.
 */
export function shouldTriggerPreload(currentTimeSeconds: number, durationSeconds: number): boolean {
  if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) return false;
  if (!Number.isFinite(currentTimeSeconds) || currentTimeSeconds < 0) return false;
  const remaining = durationSeconds - currentTimeSeconds;
  const progressRatio = currentTimeSeconds / durationSeconds;
  return progressRatio > PRELOAD_PROGRESS_RATIO_THRESHOLD || remaining < PRELOAD_REMAINING_SECONDS_THRESHOLD;
}

/** Presigned URLs are valid for 15 minutes (see TrackPlaybackDescriptor docs); treat anything preloaded this long ago as stale before swap-in. */
export const PRESIGNED_URL_TTL_MS = 15 * 60 * 1000;
/** Safety margin subtracted from the TTL so a swap doesn't race a URL expiring mid-transition. */
export const PRESIGNED_URL_SAFETY_MARGIN_MS = 30 * 1000;

/**
 * Returns true when a preloaded descriptor fetched at `resolvedAtMs` is too
 * old to trust for an immediate swap (e.g. the user paused on the current
 * track for a long time before it ended) and should be re-fetched first.
 */
export function isPreloadedUrlStale(resolvedAtMs: number, nowMs: number): boolean {
  if (!Number.isFinite(resolvedAtMs) || resolvedAtMs <= 0) return true;
  return nowMs - resolvedAtMs >= PRESIGNED_URL_TTL_MS - PRESIGNED_URL_SAFETY_MARGIN_MS;
}

/** A queue occurrence distinguishes duplicate track IDs at different queue positions. */
export interface QueueOccurrence {
  queueIndex: number;
  trackId: string;
}

/** Minimal slot state needed to decide whether a preloaded element is safe to activate. */
export interface PreloadedSlotState {
  occurrence: QueueOccurrence | null;
  ready: boolean;
  resolvedAtMs: number;
}

export function occurrencesMatch(a: QueueOccurrence | null, b: QueueOccurrence | null): boolean {
  return a !== null && b !== null && a.queueIndex === b.queueIndex && a.trackId === b.trackId;
}

/**
 * A standby element is usable only when the exact intended queue occurrence
 * finished loading and its signed URL is still fresh. A matching track ID by
 * itself is insufficient because queues may contain the same track twice.
 */
export function isStandbyReadyForOccurrence(
  slot: PreloadedSlotState,
  intended: QueueOccurrence,
  nowMs: number
): boolean {
  return slot.ready && occurrencesMatch(slot.occurrence, intended) && !isPreloadedUrlStale(slot.resolvedAtMs, nowMs);
}

/**
 * Starts a real overlap crossfade at the configured lead time, but only when
 * the intended standby occurrence is ready. `ended` remains the fallback if
 * the slot is not ready in time.
 */
export function shouldStartNaturalCrossfade(
  currentTimeSeconds: number,
  durationSeconds: number,
  crossfadeSeconds: number,
  standbyReady: boolean
): boolean {
  if (!standbyReady || crossfadeSeconds <= 0) return false;
  if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) return false;
  if (!Number.isFinite(currentTimeSeconds) || currentTimeSeconds < 0) return false;
  const remaining = durationSeconds - currentTimeSeconds;
  return remaining > 0 && remaining <= crossfadeSeconds;
}

/** Repeat-one restarts only after a real ended event, never during lead-time crossfade checks. */
export function shouldRestartRepeatOne(repeat: "off" | "one" | "all", reason: "natural-crossfade" | "ended"): boolean {
  return repeat === "one" && reason === "ended";
}

/**
 * A fade source may be finalized only while it is inactive, still contains
 * the captured load generation, and that exact generation remains reserved.
 * Clearing the reservation makes later cleanup attempts idempotent no-ops.
 */
export function canFinalizeReservedSlot(
  sourceSlotIndex: number,
  activeSlotIndex: number,
  expectedLoadId: number,
  currentLoadId: number,
  reservedLoadId: number | null
): boolean {
  return sourceSlotIndex !== activeSlotIndex && expectedLoadId === currentLoadId && reservedLoadId === expectedLoadId;
}

/**
 * A fade-out reserves only the captured load generation. The slot becomes
 * reusable once cleanup clears the reservation, or if a later legitimate
 * load generation has already replaced it.
 */
export function isSlotReusable(reservedLoadId: number | null, currentLoadId: number): boolean {
  return reservedLoadId === null || reservedLoadId !== currentLoadId;
}

/**
 * Identity of one playback cycle of a single media element.
 *
 * `loadId` is the media-LOAD identity: a new src/occurrence is loaded into the
 * element. `playGen` is the playback-CYCLE identity: pressing play again on the
 * *same* already-loaded media (resume/replay from the current position, without
 * a reload) opens a new cycle on the same `loadId`. The two are distinct because
 * `loadId` alone cannot tell "this element replayed" apart from "the same stale
 * event fired again": both carry the same `loadId`, but only a replay bumps
 * `playGen`. Guard refs record the whole cycle so a stale record from a
 * terminated cycle stops matching once the user explicitly replays that media.
 */
export interface PlaybackCycle {
  loadId: number;
  playGen: number;
}

/** Two playback cycles are the same only when both media-load and cycle match. */
export function cyclesMatch(a: PlaybackCycle | null, b: PlaybackCycle | null): boolean {
  return a !== null && b !== null && a.loadId === b.loadId && a.playGen === b.playGen;
}

/**
 * Decides whether an `ended` event fired on a given source cycle may
 * advance/repeat playback. Late or duplicate `ended` events must be inert
 * after a sleep-timer shutdown or a prior transition off the same source.
 *
 * Guards, in order:
 *   • Playback intent — only a still-playing player advances. A fixed sleep
 *     timer expiry pauses the player synchronously, so a delayed `ended` on
 *     the still-active source sees `paused`/`idle`/`error` and is rejected.
 *     (`loading` is excluded: a natural end always fires while `playing`, and
 *     an ended event should never drive a slot that is mid-load.)
 *   • Consumed cycle — the first after-current-track `ended` marks its source
 *     cycle consumed; a duplicate `ended` on the same element and cycle is
 *     rejected even though the timer has already cleared. An explicit replay of
 *     the same media opens a NEW cycle (bumped `playGen`), so the consumed
 *     record from the prior cycle no longer matches and the fresh end advances.
 *   • Prior transition — the source cycle already advanced (crossfade or an
 *     earlier ended) must not advance a second time. Replaying that same media
 *     (end-of-queue resume) likewise opens a new cycle, clearing the match.
 */
export function canEndedAdvance(
  status: "idle" | "loading" | "playing" | "paused" | "error",
  sourceCycle: PlaybackCycle,
  consumedEndedCycle: PlaybackCycle | null,
  lastTransitionSourceCycle: PlaybackCycle | null
): boolean {
  return (
    status === "playing" &&
    !cyclesMatch(consumedEndedCycle, sourceCycle) &&
    !cyclesMatch(lastTransitionSourceCycle, sourceCycle)
  );
}

/**
 * Native `ended` events are valid only while the element still reports ended.
 * A native event queued before resume or source replacement may dispatch later,
 * after the element has started a new cycle; rejecting that stale event prevents
 * it from being attributed to the mutable current slot generation. Synthetic
 * events remain allowed for the Playwright probe that drives track completion.
 */
export function hasCurrentEndedEvidence(eventIsTrusted: boolean, mediaEnded: boolean): boolean {
  return !eventIsTrusted || mediaEnded;
}

/**
 * The explicit-resume replay boundary. When the player transitions into
 * `playing` on the ACTIVE slot without any reload (the status effect resumes
 * the same already-loaded media), the slot's `ended` cycle must be reopened —
 * but only if its current cycle was already terminated by a prior `ended`
 * (consumed by after-track shutdown, or transitioned off at end-of-queue).
 *
 * Returns true exactly when a fresh cycle must be minted:
 *   • the current cycle equals the consumed record (after-track defect), OR
 *   • the current cycle equals the last-transition record (end-of-queue defect).
 *
 * Returns false for a normal mid-track pause→play (neither record matches the
 * live cycle) and after a fixed-timer expiry (which terminates via `pause`, not
 * via a consumed/transitioned `ended`, so neither record holds this cycle):
 * those paths must NOT reopen the cycle, preserving stale-event inertness.
 */
export function shouldOpenFreshEndedCycle(
  currentCycle: PlaybackCycle,
  consumedEndedCycle: PlaybackCycle | null,
  lastTransitionSourceCycle: PlaybackCycle | null
): boolean {
  return cyclesMatch(consumedEndedCycle, currentCycle) || cyclesMatch(lastTransitionSourceCycle, currentCycle);
}

/**
 * An async play settlement may affect player state only while playback is
 * still requested and the exact destination generation remains active.
 * `loading` represents autoplay/skip intent until play() resolves.
 */
export function canSettleActivePlay(
  status: "idle" | "loading" | "playing" | "paused" | "error",
  destinationSlotIndex: number,
  activeSlotIndex: number,
  expectedLoadId: number,
  currentLoadId: number,
  expectedPlayRequestId: number,
  currentPlayRequestId: number
): boolean {
  const playbackRequested = status === "loading" || status === "playing";
  return (
    playbackRequested &&
    destinationSlotIndex === activeSlotIndex &&
    expectedLoadId === currentLoadId &&
    expectedPlayRequestId === currentPlayRequestId
  );
}
