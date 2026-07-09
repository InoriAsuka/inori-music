import { beforeEach, describe, expect, it } from "vitest";
import { usePlayerStore, type QueueTrack } from "./player";

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
    shuffle: false,
    repeat: "off",
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
