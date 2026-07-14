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
export function shouldRestartRepeatOne(
  repeat: "off" | "one" | "all",
  reason: "natural-crossfade" | "ended"
): boolean {
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
  return (
    sourceSlotIndex !== activeSlotIndex &&
    expectedLoadId === currentLoadId &&
    reservedLoadId === expectedLoadId
  );
}

/**
 * A fade-out reserves only the captured load generation. The slot becomes
 * reusable once cleanup clears the reservation, or if a later legitimate
 * load generation has already replaced it.
 */
export function isSlotReusable(reservedLoadId: number | null, currentLoadId: number): boolean {
  return reservedLoadId === null || reservedLoadId !== currentLoadId;
}
