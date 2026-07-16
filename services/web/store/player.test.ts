import { beforeEach, describe, expect, it } from "vitest";
import {
  usePlayerStore,
  PLAYER_STORAGE_KEY,
  DEFAULT_PLAYBACK_SPEED,
  MIN_PLAYBACK_SPEED,
  MAX_PLAYBACK_SPEED,
  clampPlaybackSpeed,
  type QueueTrack,
} from "./player";

/**
 * Unit tests for the player store's queue state logic.
 *
 * These test pure state transitions (enqueue/skip/reorder/remove) via the
 * zustand store directly — no DOM/HTMLAudioElement involved (that lives in
 * useAudio, exercised by Playwright instead).
 */

function track(id: string): QueueTrack {
  return {
    id,
    title: `Track ${id}`,
    artistName: "Artist",
    albumTitle: "Album",
    durationSeconds: 180,
    playbackUrl: `https://example.test/${id}.mp3`,
  };
}

// Reset the store to its initial state before every test so tests don't
// leak state through the shared zustand singleton.
function resetStore() {
  usePlayerStore.setState({
    queue: [],
    currentIndex: -1,
    status: "idle",
    positionSeconds: 0,
    volume: 1,
    speed: DEFAULT_PLAYBACK_SPEED,
    shuffle: false,
    repeat: "off",
    restoredPending: false,
  });
}

beforeEach(() => {
  resetStore();
});

describe("playQueue", () => {
  it("replaces the queue and sets currentIndex/status", () => {
    const tracks = [track("a"), track("b"), track("c")];
    usePlayerStore.getState().playQueue(tracks, 1);
    const s = usePlayerStore.getState();
    expect(s.queue).toEqual(tracks);
    expect(s.currentIndex).toBe(1);
    expect(s.status).toBe("loading");
    expect(s.positionSeconds).toBe(0);
  });

  it("defaults startIndex to 0", () => {
    usePlayerStore.getState().playQueue([track("a")]);
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });
});

describe("enqueue", () => {
  it("appends to the end of the queue", () => {
    usePlayerStore.getState().playQueue([track("a")]);
    usePlayerStore.getState().enqueue(track("b"));
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["a", "b"]);
  });
});

describe("enqueueNext", () => {
  it("inserts immediately after the currently playing track", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().enqueueNext(track("x"));
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["a", "x", "b"]);
  });

  it("appends when nothing is currently loaded (currentIndex = -1)", () => {
    resetStore();
    usePlayerStore.getState().enqueueNext(track("x"));
    // currentIndex -1 + 1 = 0, so it should splice at index 0 (front)
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["x"]);
  });
});

describe("clearQueue", () => {
  it("empties the queue and resets index/status/position", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().setPosition(42);
    usePlayerStore.getState().clearQueue();
    const s = usePlayerStore.getState();
    expect(s.queue).toEqual([]);
    expect(s.currentIndex).toBe(-1);
    expect(s.status).toBe("idle");
    expect(s.positionSeconds).toBe(0);
  });
});

describe("removeFromQueue", () => {
  it("removes the track at the given index", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().removeFromQueue(1);
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["a", "c"]);
  });

  it("decrements currentIndex when removing a track before it", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 2);
    usePlayerStore.getState().removeFromQueue(0);
    expect(usePlayerStore.getState().currentIndex).toBe(1);
  });

  it("leaves currentIndex unchanged when removing a track after it", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().removeFromQueue(2);
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });
});

describe("reorderQueue", () => {
  it("moves a track from one index to another", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().reorderQueue(0, 2);
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["b", "c", "a"]);
  });

  it("tracks the currently playing item when it is the one moved", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().reorderQueue(0, 2);
    expect(usePlayerStore.getState().currentIndex).toBe(2);
  });

  it("shifts currentIndex left when a track moves from before it to after it", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 1);
    usePlayerStore.getState().reorderQueue(0, 2);
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });

  it("shifts currentIndex right when a track moves from after it to before it", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 1);
    usePlayerStore.getState().reorderQueue(2, 0);
    expect(usePlayerStore.getState().currentIndex).toBe(2);
  });

  it("is a no-op for equal indices or out-of-range indices", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().reorderQueue(0, 0);
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["a", "b"]);
    usePlayerStore.getState().reorderQueue(0, 5);
    expect(usePlayerStore.getState().queue.map((t) => t.id)).toEqual(["a", "b"]);
  });
});

describe("skipToNext", () => {
  it("advances currentIndex and sets status to loading", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().skipToNext();
    const s = usePlayerStore.getState();
    expect(s.currentIndex).toBe(1);
    expect(s.status).toBe("loading");
    expect(s.positionSeconds).toBe(0);
  });

  it("does nothing when the queue is empty", () => {
    resetStore();
    usePlayerStore.getState().skipToNext();
    expect(usePlayerStore.getState().currentIndex).toBe(-1);
  });

  it("restarts the current track when repeat is 'one'", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().cycleRepeat(); // off -> all
    usePlayerStore.getState().cycleRepeat(); // all -> one
    usePlayerStore.getState().skipToNext();
    expect(usePlayerStore.getState().currentIndex).toBe(0);
    expect(usePlayerStore.getState().status).toBe("loading");
  });

  it("wraps to index 0 when repeat is 'all' and at the end of the queue", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().cycleRepeat(); // off -> all
    usePlayerStore.getState().skipToNext();
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });

  it("sets status to idle at the end of the queue with no repeat", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().skipToNext();
    expect(usePlayerStore.getState().status).toBe("idle");
    // currentIndex is left as-is when playback naturally stops.
    expect(usePlayerStore.getState().currentIndex).toBe(1);
  });

  it("picks a pseudo-random index when shuffle is on", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().toggleShuffle();
    usePlayerStore.getState().skipToNext();
    const idx = usePlayerStore.getState().currentIndex;
    expect(idx).toBeGreaterThanOrEqual(0);
    expect(idx).toBeLessThan(3);
  });
});

describe("skipToPrevious", () => {
  it("moves to the previous index and resets position", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().skipToPrevious();
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });

  it("restarts the current track instead of skipping back when more than 3s in", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().setPosition(10);
    usePlayerStore.getState().skipToPrevious();
    expect(usePlayerStore.getState().currentIndex).toBe(1);
    expect(usePlayerStore.getState().positionSeconds).toBe(0);
  });

  it("clamps to index 0 at the start of the queue", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().skipToPrevious();
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });

  it("does nothing when the queue is empty", () => {
    resetStore();
    usePlayerStore.getState().skipToPrevious();
    expect(usePlayerStore.getState().currentIndex).toBe(-1);
  });
});

describe("skipToIndex", () => {
  it("jumps directly to a valid index", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b"), track("c")], 0);
    usePlayerStore.getState().skipToIndex(2);
    expect(usePlayerStore.getState().currentIndex).toBe(2);
    expect(usePlayerStore.getState().status).toBe("loading");
  });

  it("ignores out-of-range indices", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 0);
    usePlayerStore.getState().skipToIndex(5);
    expect(usePlayerStore.getState().currentIndex).toBe(0);
    usePlayerStore.getState().skipToIndex(-1);
    expect(usePlayerStore.getState().currentIndex).toBe(0);
  });
});

describe("play/pause", () => {
  it("sets status to playing/paused", () => {
    usePlayerStore.getState().play();
    expect(usePlayerStore.getState().status).toBe("playing");
    usePlayerStore.getState().pause();
    expect(usePlayerStore.getState().status).toBe("paused");
  });
});

describe("setVolume", () => {
  it("clamps volume to [0, 1]", () => {
    usePlayerStore.getState().setVolume(1.5);
    expect(usePlayerStore.getState().volume).toBe(1);
    usePlayerStore.getState().setVolume(-0.5);
    expect(usePlayerStore.getState().volume).toBe(0);
    usePlayerStore.getState().setVolume(0.42);
    expect(usePlayerStore.getState().volume).toBe(0.42);
  });
});

describe("setSpeed", () => {
  it("defaults to 1×", () => {
    expect(usePlayerStore.getState().speed).toBe(DEFAULT_PLAYBACK_SPEED);
    expect(DEFAULT_PLAYBACK_SPEED).toBe(1);
  });

  it("sets an in-range speed unchanged", () => {
    usePlayerStore.getState().setSpeed(1.5);
    expect(usePlayerStore.getState().speed).toBe(1.5);
    usePlayerStore.getState().setSpeed(0.75);
    expect(usePlayerStore.getState().speed).toBe(0.75);
  });

  it("clamps below the minimum to MIN_PLAYBACK_SPEED", () => {
    usePlayerStore.getState().setSpeed(0.1);
    expect(usePlayerStore.getState().speed).toBe(MIN_PLAYBACK_SPEED);
    expect(MIN_PLAYBACK_SPEED).toBe(0.5);
  });

  it("clamps above the maximum to MAX_PLAYBACK_SPEED", () => {
    usePlayerStore.getState().setSpeed(5);
    expect(usePlayerStore.getState().speed).toBe(MAX_PLAYBACK_SPEED);
    expect(MAX_PLAYBACK_SPEED).toBe(2);
  });

  it("accepts the exact boundary values", () => {
    usePlayerStore.getState().setSpeed(MIN_PLAYBACK_SPEED);
    expect(usePlayerStore.getState().speed).toBe(0.5);
    usePlayerStore.getState().setSpeed(MAX_PLAYBACK_SPEED);
    expect(usePlayerStore.getState().speed).toBe(2);
  });

  it("falls back to the default for non-finite input", () => {
    usePlayerStore.getState().setSpeed(Number.NaN);
    expect(usePlayerStore.getState().speed).toBe(DEFAULT_PLAYBACK_SPEED);
    usePlayerStore.getState().setSpeed(Number.POSITIVE_INFINITY);
    expect(usePlayerStore.getState().speed).toBe(DEFAULT_PLAYBACK_SPEED);
  });
});

describe("clampPlaybackSpeed", () => {
  it("clamps arbitrary boundary input into range", () => {
    expect(clampPlaybackSpeed(-100)).toBe(MIN_PLAYBACK_SPEED);
    expect(clampPlaybackSpeed(0)).toBe(MIN_PLAYBACK_SPEED);
    expect(clampPlaybackSpeed(100)).toBe(MAX_PLAYBACK_SPEED);
    expect(clampPlaybackSpeed(1.25)).toBe(1.25);
    expect(clampPlaybackSpeed(Number.NaN)).toBe(DEFAULT_PLAYBACK_SPEED);
  });
});

describe("toggleShuffle / cycleRepeat", () => {
  it("toggles shuffle on and off", () => {
    expect(usePlayerStore.getState().shuffle).toBe(false);
    usePlayerStore.getState().toggleShuffle();
    expect(usePlayerStore.getState().shuffle).toBe(true);
    usePlayerStore.getState().toggleShuffle();
    expect(usePlayerStore.getState().shuffle).toBe(false);
  });

  it("cycles repeat off -> all -> one -> off", () => {
    expect(usePlayerStore.getState().repeat).toBe("off");
    usePlayerStore.getState().cycleRepeat();
    expect(usePlayerStore.getState().repeat).toBe("all");
    usePlayerStore.getState().cycleRepeat();
    expect(usePlayerStore.getState().repeat).toBe("one");
    usePlayerStore.getState().cycleRepeat();
    expect(usePlayerStore.getState().repeat).toBe("off");
  });
});

describe("useCurrentTrack / useIsPlaying selectors", () => {
  it("useCurrentTrack reflects the queue via getState (selector logic exercised directly)", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    const s = usePlayerStore.getState();
    expect(s.queue[s.currentIndex]?.id).toBe("b");
  });

  it("returns null when currentIndex is -1 (nothing loaded)", () => {
    resetStore();
    const s = usePlayerStore.getState();
    expect(s.queue[s.currentIndex] ?? null).toBeNull();
  });
});

describe("persistence: acknowledgeRestore / clearPersisted", () => {
  it("acknowledgeRestore clears restoredPending without touching other state", () => {
    usePlayerStore.setState({ restoredPending: true, positionSeconds: 42 });
    usePlayerStore.getState().acknowledgeRestore();
    const s = usePlayerStore.getState();
    expect(s.restoredPending).toBe(false);
    expect(s.positionSeconds).toBe(42);
  });

  it("clearPersisted resets queue/index/status/position/shuffle/repeat", () => {
    usePlayerStore.getState().playQueue([track("a"), track("b")], 1);
    usePlayerStore.getState().toggleShuffle();
    usePlayerStore.getState().cycleRepeat();
    usePlayerStore.getState().setPosition(77);
    usePlayerStore.getState().clearPersisted();
    const s = usePlayerStore.getState();
    expect(s.queue).toEqual([]);
    expect(s.currentIndex).toBe(-1);
    expect(s.status).toBe("idle");
    expect(s.positionSeconds).toBe(0);
    expect(s.shuffle).toBe(false);
    expect(s.repeat).toBe("off");
    expect(s.restoredPending).toBe(false);
  });

  it("clearPersisted preserves volume (a device preference, not session state)", () => {
    usePlayerStore.getState().setVolume(0.3);
    usePlayerStore.getState().clearPersisted();
    expect(usePlayerStore.getState().volume).toBe(0.3);
  });

  it("clearPersisted preserves speed (a device preference, consistent with volume)", () => {
    usePlayerStore.getState().setSpeed(1.5);
    usePlayerStore.getState().clearPersisted();
    expect(usePlayerStore.getState().speed).toBe(1.5);
  });

  it("playQueue clears any pending restoredPending flag", () => {
    usePlayerStore.setState({ restoredPending: true });
    usePlayerStore.getState().playQueue([track("a")]);
    expect(usePlayerStore.getState().restoredPending).toBe(false);
  });

  it("play() clears restoredPending (user gesture acknowledges the restore)", () => {
    usePlayerStore.setState({ restoredPending: true });
    usePlayerStore.getState().play();
    expect(usePlayerStore.getState().restoredPending).toBe(false);
  });
});

describe("persistence: serialize/restore round-trip (merge logic)", () => {
  // These exercise the zustand `persist` config's `partialize`/`merge`
  // functions directly, mirroring how the middleware itself invokes them
  // during store creation, without needing a real localStorage/browser
  // environment or reaching into zustand internals.

  async function loadPersistConfig() {
    // Re-import fresh so we can introspect the store's persist options via
    // the public `usePlayerStore.persist` API zustand's middleware attaches.
    const mod = await import("./player");
    return mod.usePlayerStore;
  }

  it("PLAYER_STORAGE_KEY matches the persist middleware's configured name", async () => {
    const store = await loadPersistConfig();
    expect(store.persist.getOptions().name).toBe(PLAYER_STORAGE_KEY);
  });

  it("partialize omits transient fields (status, restoredPending) from the persisted snapshot", async () => {
    const store = await loadPersistConfig();
    const tracks: QueueTrack[] = [track("a"), track("b")];
    store.getState().playQueue(tracks, 1);
    store.getState().setPosition(55);
    store.getState().setVolume(0.6);
    store.getState().setSpeed(1.25);
    store.getState().toggleShuffle();
    store.getState().cycleRepeat();

    const partialize = store.persist.getOptions().partialize;
    expect(partialize).toBeDefined();
    if (!partialize) throw new Error("partialize must be defined");
    const snapshot = partialize(store.getState());

    expect(snapshot).toMatchObject({
      queue: tracks,
      currentIndex: 1,
      positionSeconds: 55,
      volume: 0.6,
      speed: 1.25,
      shuffle: true,
      repeat: "all",
    });
    expect(snapshot).not.toHaveProperty("status");
    expect(snapshot).not.toHaveProperty("restoredPending");
  });

  it("merge restores persisted fields, forces status to idle, and sets restoredPending when a track was mid-queue", async () => {
    const store = await loadPersistConfig();
    const merge = store.persist.getOptions().merge;
    expect(merge).toBeDefined();
    if (!merge) throw new Error("merge must be defined");

    const persisted = {
      queue: [track("a"), track("b")],
      currentIndex: 1,
      positionSeconds: 88,
      volume: 0.4,
      speed: 1.75,
      shuffle: true,
      repeat: "one" as const,
    };
    const merged = merge(persisted, store.getState()) as ReturnType<typeof store.getState>;

    expect(merged.queue).toEqual(persisted.queue);
    expect(merged.currentIndex).toBe(1);
    expect(merged.positionSeconds).toBe(88);
    expect(merged.volume).toBe(0.4);
    expect(merged.speed).toBe(1.75);
    expect(merged.shuffle).toBe(true);
    expect(merged.repeat).toBe("one");
    // Never resume playback automatically after a restore.
    expect(merged.status).toBe("idle");
    expect(merged.restoredPending).toBe(true);
  });

  it("merge does not set restoredPending when there is no persisted queue (fresh install)", async () => {
    const store = await loadPersistConfig();
    const merge = store.persist.getOptions().merge;
    if (!merge) throw new Error("merge must be defined");
    const merged = merge(undefined, store.getState()) as ReturnType<typeof store.getState>;
    expect(merged.restoredPending).toBe(false);
    expect(merged.status).toBe("idle");
  });

  it("merge clamps a corrupt persisted speed into range", async () => {
    const store = await loadPersistConfig();
    const merge = store.persist.getOptions().merge;
    if (!merge) throw new Error("merge must be defined");
    const persisted = {
      queue: [track("a")],
      currentIndex: 0,
      positionSeconds: 0,
      volume: 1,
      speed: 99,
      shuffle: false,
      repeat: "off" as const,
    };
    const merged = merge(persisted, store.getState()) as ReturnType<typeof store.getState>;
    expect(merged.speed).toBe(MAX_PLAYBACK_SPEED);
  });

  it("merge does not set restoredPending when persisted currentIndex is -1 (queue cleared before refresh)", async () => {
    const store = await loadPersistConfig();
    const merge = store.persist.getOptions().merge;
    if (!merge) throw new Error("merge must be defined");
    const persisted = {
      queue: [],
      currentIndex: -1,
      positionSeconds: 0,
      volume: 1,
      shuffle: false,
      repeat: "off" as const,
    };
    const merged = merge(persisted, store.getState()) as ReturnType<typeof store.getState>;
    expect(merged.restoredPending).toBe(false);
  });
});
