/**
 * Pure helpers for cross-device player-state sync (v5.4.0).
 *
 * Kept side-effect-free (no store/network access) so the reporting and resume
 * decision logic is unit-testable in isolation (see playerStateSync.test.ts).
 * The React wiring lives in hooks/usePlayerStateSync.ts.
 */
import type { PlayerStateUpload, RemotePlayerState } from "@/lib/api/player-state";
import type { QueueTrack } from "@/store/player";

/** The subset of player-store fields that define a reportable snapshot. */
export interface LocalPlayerSnapshot {
  queue: QueueTrack[];
  currentIndex: number;
  positionSeconds: number;
  repeat: "off" | "one" | "all";
  shuffle: boolean;
  volume: number;
  speed: number;
  status: string;
}

/** Serialise a local snapshot into the server upload shape (queue → track IDs). */
export function toUpload(snapshot: LocalPlayerSnapshot): PlayerStateUpload {
  return {
    queue: snapshot.queue.map((t) => t.id),
    currentIndex: snapshot.currentIndex,
    positionSeconds: snapshot.positionSeconds,
    repeat: snapshot.repeat,
    shuffle: snapshot.shuffle,
    volume: snapshot.volume,
    speed: snapshot.speed,
    status: snapshot.status,
  };
}

/**
 * A "reporting signature" is the identity that, when it changes, warrants a
 * report — everything except the ever-ticking position (which is throttled
 * separately by time, not by change). Two snapshots with the same signature but
 * different positions should NOT trigger an immediate flush, only a throttled
 * position update. Used to distinguish "song/queue/mode changed → flush now"
 * from "position advanced → coalesce".
 */
export function reportingSignature(snapshot: LocalPlayerSnapshot): string {
  return [
    snapshot.queue.map((t) => t.id).join(","),
    snapshot.currentIndex,
    snapshot.repeat,
    snapshot.shuffle,
    snapshot.volume,
    snapshot.speed,
    snapshot.status,
  ].join("|");
}

/** Whether the remote queue/position differs enough from local to bother resuming. */
export interface ResumeDecision {
  /** Show the resume prompt. */
  prompt: boolean;
  /** The remote track id at its currentIndex (for the prompt label lookup), or null. */
  remoteTrackId: string | null;
  /** The remote position, in seconds (for the prompt "mm:ss" label). */
  positionSeconds: number;
}

const NO_RESUME: ResumeDecision = { prompt: false, remoteTrackId: null, positionSeconds: 0 };

/**
 * Decide whether to offer a cross-device resume, given the remote state and the
 * current local snapshot.
 *
 * Prompt when the remote state points at a real track AND it differs from what
 * the local device is already showing — i.e. a different track is queued, or a
 * meaningfully different position on the same track (local empty counts as
 * different). We deliberately do NOT gate on timestamps here: the caller only
 * fetches the remote state on load, and a same-track/same-position remote is
 * simply nothing worth interrupting the user for.
 *
 * `positionEpsilonSeconds` guards against prompting for a sub-threshold drift
 * on the identical track (e.g. the same device's own last report).
 */
export function decideResume(
  remote: RemotePlayerState | null,
  local: LocalPlayerSnapshot,
  positionEpsilonSeconds = 5
): ResumeDecision {
  if (!remote) return NO_RESUME;
  const remoteTrackId = remote.queue[remote.currentIndex] ?? null;
  if (!remoteTrackId) return NO_RESUME;

  const localTrackId = local.queue[local.currentIndex]?.id ?? null;

  // Different track (or local has nothing loaded) → always worth offering.
  if (localTrackId !== remoteTrackId) {
    return { prompt: true, remoteTrackId, positionSeconds: remote.positionSeconds };
  }

  // Same track: only prompt if the remote position is meaningfully ahead of or
  // behind the local one (avoids re-prompting on a device's own stale report).
  if (Math.abs(remote.positionSeconds - local.positionSeconds) > positionEpsilonSeconds) {
    return { prompt: true, remoteTrackId, positionSeconds: remote.positionSeconds };
  }

  return NO_RESUME;
}

/** Format seconds as m:ss for the resume-prompt label. */
export function formatResumePosition(seconds: number): string {
  const total = Math.max(0, Math.floor(seconds));
  const m = Math.floor(total / 60);
  const s = total % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}
