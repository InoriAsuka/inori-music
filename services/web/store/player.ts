/**
 * Player store — global playback queue and state.
 *
 * Manages the current track, queue, and playback status.
 * The actual HTMLAudioElement lives in useAudio (hook), not here —
 * the store only holds serialisable state.
 *
 * Persistence: queue/currentIndex/positionSeconds/volume/shuffle/repeat are
 * persisted to localStorage via zustand's `persist` middleware so a page
 * refresh can restore the last playback state (queue rebuilt, position
 * seeked to) WITHOUT autoplay — see useAudio's restore effect. Position
 * writes are throttled (see `setPosition`) so the 250ms UI tick doesn't
 * hammer localStorage on every frame. Cleared on logout via `clearPersisted`
 * (called from store/auth.ts clearSession).
 */
"use client";

import { create } from "zustand";
import { persist, type PersistStorage, type StorageValue } from "zustand/middleware";

export interface QueueTrack {
  id: string;
  title: string;
  artistName: string;
  albumTitle: string;
  /** Duration in seconds */
  durationSeconds: number;
  /** Signed playback URL (obtained from /api/v1/catalog/tracks/{id}/playback) */
  playbackUrl: string;
  /** Optional artwork URL */
  artworkUrl?: string;
}

type PlaybackStatus = "idle" | "loading" | "playing" | "paused" | "error";

export const PLAYER_STORAGE_KEY = "inori-player";
/** Minimum interval between persisted position writes (ms). */
const POSITION_PERSIST_THROTTLE_MS = 5000;

/** Playback speed bounds (inclusive). UI presets stay within this range; the
 *  store clamps any arbitrary input so boundary values are always safe. */
export const MIN_PLAYBACK_SPEED = 0.5;
export const MAX_PLAYBACK_SPEED = 2.0;
export const DEFAULT_PLAYBACK_SPEED = 1;

/** Clamp an arbitrary speed into [MIN_PLAYBACK_SPEED, MAX_PLAYBACK_SPEED],
 *  falling back to the default for non-finite input (NaN/±Infinity). */
export function clampPlaybackSpeed(speed: number): number {
  if (!Number.isFinite(speed)) return DEFAULT_PLAYBACK_SPEED;
  return Math.max(MIN_PLAYBACK_SPEED, Math.min(MAX_PLAYBACK_SPEED, speed));
}

interface PlayerState {
  queue: QueueTrack[];
  /** Index into queue of the currently active track (−1 = nothing loaded). */
  currentIndex: number;
  status: PlaybackStatus;
  /** Current playback position in seconds (updated by useAudio hook). */
  positionSeconds: number;
  /** Volume 0–1 */
  volume: number;
  /** Playback speed, clamped to [MIN_PLAYBACK_SPEED, MAX_PLAYBACK_SPEED]. */
  speed: number;
  /** Shuffle on/off */
  shuffle: boolean;
  /** Repeat: off | one | all */
  repeat: "off" | "one" | "all";
  /**
   * True once a persisted queue has been restored on load but not yet
   * resumed by a user gesture. useAudio checks this to seek position
   * without autoplaying. Cleared by `acknowledgeRestore()`.
   */
  restoredPending: boolean;

  // ── Queue management ──────────────────────────────────────────────────
  /** Replace queue and start playing from index. */
  playQueue: (tracks: QueueTrack[], startIndex?: number) => void;
  /** Append a track to the end of the queue. */
  enqueue: (track: QueueTrack) => void;
  /** Enqueue a track next (after the currently playing one). */
  enqueueNext: (track: QueueTrack) => void;
  /** Clear the queue. */
  clearQueue: () => void;
  /** Remove a single track from the queue by index. */
  removeFromQueue: (index: number) => void;
  /** Reorder queue by moving one item. Keeps current track selected. */
  reorderQueue: (fromIndex: number, toIndex: number) => void;

  // ── Playback control ──────────────────────────────────────────────────
  play: () => void;
  pause: () => void;
  skipToNext: () => void;
  skipToPrevious: () => void;
  skipToIndex: (index: number) => void;

  // ── State setters (used by useAudio hook) ─────────────────────────────
  setStatus: (status: PlaybackStatus) => void;
  setPosition: (seconds: number) => void;
  setVolume: (volume: number) => void;
  /** Set playback speed; clamped to [MIN_PLAYBACK_SPEED, MAX_PLAYBACK_SPEED]. */
  setSpeed: (speed: number) => void;
  toggleShuffle: () => void;
  cycleRepeat: () => void;

  // ── Persistence lifecycle ──────────────────────────────────────────────
  /** Marks the restored queue as acknowledged (called once useAudio has wired up position). */
  acknowledgeRestore: () => void;
  /** Wipes persisted playback state (called on logout). */
  clearPersisted: () => void;
  /**
   * Replace the queue/index/position/modes from a remote (cross-device) player
   * state without autoplaying — status becomes "paused" and restoredPending is
   * set so useAudio seeks to the restored position and waits for a user gesture
   * (same no-autoplay contract as a persisted-refresh restore). Used by the
   * cross-device "Resume from another device?" prompt.
   */
  hydrateRemoteState: (state: {
    queue: QueueTrack[];
    currentIndex: number;
    positionSeconds: number;
    repeat: "off" | "one" | "all";
    shuffle: boolean;
    volume: number;
    speed: number;
  }) => void;
}

/**
 * Throttled localStorage writer for `persist`. zustand's persist middleware
 * writes to storage on every `set()` by default — for a 250ms position tick
 * that means 4 writes/sec. We wrap the storage adapter so position-only
 * changes are coalesced to at most one write per POSITION_PERSIST_THROTTLE_MS,
 * while queue/volume/shuffle/repeat changes (rare, user-initiated) still
 * flush immediately.
 */
function createThrottledStorage(): PersistStorage<PersistedPlayerState> {
  let lastWriteAt = 0;
  let pendingTimer: ReturnType<typeof setTimeout> | null = null;
  let pendingValue: StorageValue<PersistedPlayerState> | null = null;

  function flush(name: string) {
    if (!pendingValue) return;
    lastWriteAt = Date.now();
    localStorage.setItem(name, JSON.stringify(pendingValue));
    pendingValue = null;
  }

  return {
    getItem(name) {
      if (typeof window === "undefined") return null;
      const raw = localStorage.getItem(name);
      if (!raw) return null;
      try {
        return JSON.parse(raw) as StorageValue<PersistedPlayerState>;
      } catch {
        return null;
      }
    },
    setItem(name, value) {
      if (typeof window === "undefined") return;
      pendingValue = value;
      const elapsed = Date.now() - lastWriteAt;
      if (elapsed >= POSITION_PERSIST_THROTTLE_MS) {
        flush(name);
        return;
      }
      if (pendingTimer) clearTimeout(pendingTimer);
      pendingTimer = setTimeout(() => flush(name), POSITION_PERSIST_THROTTLE_MS - elapsed);
    },
    removeItem(name) {
      if (typeof window === "undefined") return;
      if (pendingTimer) {
        clearTimeout(pendingTimer);
        pendingTimer = null;
      }
      pendingValue = null;
      localStorage.removeItem(name);
    },
  };
}

interface PersistedPlayerState {
  queue: QueueTrack[];
  currentIndex: number;
  positionSeconds: number;
  volume: number;
  speed: number;
  shuffle: boolean;
  repeat: "off" | "one" | "all";
}

export const usePlayerStore = create<PlayerState>()(
  persist(
    (set, get) => ({
      queue: [],
      currentIndex: -1,
      status: "idle",
      positionSeconds: 0,
      volume: 1,
      speed: DEFAULT_PLAYBACK_SPEED,
      shuffle: false,
      repeat: "off",
      restoredPending: false,

      playQueue(tracks, startIndex = 0) {
        set({ queue: tracks, currentIndex: startIndex, status: "loading", positionSeconds: 0, restoredPending: false });
      },

      enqueue(track) {
        set((s) => ({ queue: [...s.queue, track] }));
      },

      enqueueNext(track) {
        set((s) => {
          const next = s.currentIndex + 1;
          const q = [...s.queue];
          q.splice(next, 0, track);
          return { queue: q };
        });
      },

      clearQueue() {
        set({ queue: [], currentIndex: -1, status: "idle", positionSeconds: 0, restoredPending: false });
      },

      removeFromQueue(index) {
        set((s) => {
          const q = [...s.queue];
          q.splice(index, 1);
          const ci = index < s.currentIndex ? s.currentIndex - 1 : s.currentIndex;
          return { queue: q, currentIndex: ci };
        });
      },

      reorderQueue(fromIndex, toIndex) {
        set((s) => {
          if (fromIndex === toIndex) return s;
          if (fromIndex < 0 || fromIndex >= s.queue.length || toIndex < 0 || toIndex >= s.queue.length) return s;
          const q = [...s.queue];
          const [moved] = q.splice(fromIndex, 1);
          q.splice(toIndex, 0, moved);

          let currentIndex = s.currentIndex;
          if (s.currentIndex === fromIndex) currentIndex = toIndex;
          else if (fromIndex < s.currentIndex && toIndex >= s.currentIndex) currentIndex = s.currentIndex - 1;
          else if (fromIndex > s.currentIndex && toIndex <= s.currentIndex) currentIndex = s.currentIndex + 1;

          return { queue: q, currentIndex };
        });
      },

      play() {
        set({ status: "playing", restoredPending: false });
      },

      pause() {
        set({ status: "paused" });
      },

      skipToNext() {
        const { queue, currentIndex, repeat, shuffle } = get();
        if (queue.length === 0) return;
        if (repeat === "one") {
          set({ positionSeconds: 0, status: "loading" });
          return;
        }
        let next: number;
        if (shuffle) {
          next = Math.floor(Math.random() * queue.length);
        } else {
          next = currentIndex + 1;
          if (next >= queue.length) {
            if (repeat === "all") next = 0;
            else {
              set({ status: "idle" });
              return;
            }
          }
        }
        set({ currentIndex: next, positionSeconds: 0, status: "loading" });
      },

      skipToPrevious() {
        const { queue, currentIndex, positionSeconds } = get();
        if (queue.length === 0) return;
        // If more than 3s in, restart current track instead.
        if (positionSeconds > 3) {
          set({ positionSeconds: 0, status: "loading" });
          return;
        }
        const prev = Math.max(0, currentIndex - 1);
        set({ currentIndex: prev, positionSeconds: 0, status: "loading" });
      },

      skipToIndex(index) {
        const { queue } = get();
        if (index < 0 || index >= queue.length) return;
        set({ currentIndex: index, positionSeconds: 0, status: "loading" });
      },

      setStatus(status) {
        set({ status });
      },
      setPosition(positionSeconds) {
        set({ positionSeconds });
      },
      setVolume(volume) {
        set({ volume: Math.max(0, Math.min(1, volume)) });
      },
      setSpeed(speed) {
        set({ speed: clampPlaybackSpeed(speed) });
      },
      toggleShuffle() {
        set((s) => ({ shuffle: !s.shuffle }));
      },
      cycleRepeat() {
        set((s) => ({
          repeat: s.repeat === "off" ? "all" : s.repeat === "all" ? "one" : "off",
        }));
      },

      acknowledgeRestore() {
        set({ restoredPending: false });
      },

      clearPersisted() {
        set({
          queue: [],
          currentIndex: -1,
          status: "idle",
          positionSeconds: 0,
          shuffle: false,
          repeat: "off",
          restoredPending: false,
        });
      },

      hydrateRemoteState(state) {
        set({
          queue: state.queue,
          currentIndex: state.currentIndex,
          positionSeconds: state.positionSeconds,
          repeat: state.repeat,
          shuffle: state.shuffle,
          volume: state.volume,
          speed: clampPlaybackSpeed(state.speed),
          status: "paused",
          restoredPending: true,
        });
      },
    }),
    {
      name: PLAYER_STORAGE_KEY,
      storage: createThrottledStorage(),
      // Only persist serialisable playback state; status is transient
      // (restored queues start "idle" until the user presses play — no
      // autoplay after refresh) and restoredPending is derived on rehydrate.
      partialize: (s) => ({
        queue: s.queue,
        currentIndex: s.currentIndex,
        positionSeconds: s.positionSeconds,
        volume: s.volume,
        speed: s.speed,
        shuffle: s.shuffle,
        repeat: s.repeat,
      }),
      merge: (persisted, current) => {
        const p = persisted as Partial<PersistedPlayerState> | undefined;
        const hasRestoredTrack = !!p?.queue?.length && (p?.currentIndex ?? -1) >= 0;
        return {
          ...current,
          ...p,
          // Guard against corrupt/out-of-range persisted speed.
          speed: clampPlaybackSpeed(p?.speed ?? current.speed),
          status: "idle",
          restoredPending: hasRestoredTrack,
        };
      },
    }
  )
);

// Convenience selectors
export const useCurrentTrack = () => usePlayerStore((s) => s.queue[s.currentIndex] ?? null);
export const useIsPlaying = () => usePlayerStore((s) => s.status === "playing");
