/**
 * Player store — global playback queue and state.
 *
 * Manages the current track, queue, and playback status.
 * The actual HTMLAudioElement lives in useAudio (hook), not here —
 * the store only holds serialisable state.
 */
"use client";

import { create } from "zustand";

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

interface PlayerState {
  queue: QueueTrack[];
  /** Index into queue of the currently active track (−1 = nothing loaded). */
  currentIndex: number;
  status: PlaybackStatus;
  /** Current playback position in seconds (updated by useAudio hook). */
  positionSeconds: number;
  /** Volume 0–1 */
  volume: number;
  /** Shuffle on/off */
  shuffle: boolean;
  /** Repeat: off | one | all */
  repeat: "off" | "one" | "all";

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
  toggleShuffle: () => void;
  cycleRepeat: () => void;
}

export const usePlayerStore = create<PlayerState>()((set, get) => ({
  queue: [],
  currentIndex: -1,
  status: "idle",
  positionSeconds: 0,
  volume: 1,
  shuffle: false,
  repeat: "off",

  playQueue(tracks, startIndex = 0) {
    set({ queue: tracks, currentIndex: startIndex, status: "loading", positionSeconds: 0 });
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
    set({ queue: [], currentIndex: -1, status: "idle", positionSeconds: 0 });
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
    set({ status: "playing" });
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
        else { set({ status: "idle" }); return; }
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

  setStatus(status) { set({ status }); },
  setPosition(positionSeconds) { set({ positionSeconds }); },
  setVolume(volume) { set({ volume: Math.max(0, Math.min(1, volume)) }); },
  toggleShuffle() { set((s) => ({ shuffle: !s.shuffle })); },
  cycleRepeat() {
    set((s) => ({
      repeat: s.repeat === "off" ? "all" : s.repeat === "all" ? "one" : "off",
    }));
  },
}));

// Convenience selectors
export const useCurrentTrack = () =>
  usePlayerStore((s) => s.queue[s.currentIndex] ?? null);
export const useIsPlaying = () =>
  usePlayerStore((s) => s.status === "playing");
