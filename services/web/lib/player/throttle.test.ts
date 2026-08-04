import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { createThrottle } from "./throttle";

beforeEach(() => {
  vi.useFakeTimers();
  // Start at t=1000 so the first call immediately exceeds the 0→interval
  // window and fires on the leading edge, matching the documented contract.
  vi.setSystemTime(1000);
});

afterEach(() => {
  vi.useRealTimers();
});

describe("createThrottle", () => {
  it("emits the first value immediately (leading edge)", () => {
    const emit = vi.fn();
    createThrottle(1000, emit).schedule("a");
    expect(emit).toHaveBeenCalledTimes(1);
    expect(emit).toHaveBeenCalledWith("a");
  });

  it("coalesces a burst into one trailing emit carrying the latest value", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a"); // leading
    t.schedule("b");
    t.schedule("c");
    expect(emit).toHaveBeenCalledTimes(1);
    expect(emit).toHaveBeenLastCalledWith("a");

    vi.advanceTimersByTime(1000);
    expect(emit).toHaveBeenCalledTimes(2);
    expect(emit).toHaveBeenLastCalledWith("c");
  });

  it("emits immediately again once the window has fully elapsed", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a");
    vi.advanceTimersByTime(1000);
    t.schedule("b");

    expect(emit).toHaveBeenCalledTimes(2);
    expect(emit).toHaveBeenLastCalledWith("b");
  });

  it("does not emit a trailing value that was never scheduled", () => {
    const emit = vi.fn();
    createThrottle(1000, emit).schedule("a");

    vi.advanceTimersByTime(5000);
    expect(emit).toHaveBeenCalledTimes(1);
  });

  it("flush emits the pending value immediately", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a"); // leading
    t.schedule("b"); // pending
    t.flush();

    expect(emit).toHaveBeenCalledTimes(2);
    expect(emit).toHaveBeenLastCalledWith("b");
  });

  it("flush with nothing pending is a no-op", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a");
    t.flush();
    t.flush();

    expect(emit).toHaveBeenCalledTimes(1);
  });

  it("cancel drops the pending value without emitting", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a"); // leading
    t.schedule("b"); // pending
    t.cancel();
    vi.advanceTimersByTime(5000);

    expect(emit).toHaveBeenCalledTimes(1);
    expect(emit).toHaveBeenLastCalledWith("a");
  });

  it("keeps working after a cancel", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a");
    t.schedule("b");
    t.cancel();

    vi.advanceTimersByTime(1000);
    t.schedule("c");

    expect(emit).toHaveBeenLastCalledWith("c");
  });

  it("schedule after the window elapses emits immediately (no trailing)", () => {
    const emit = vi.fn();
    const t = createThrottle(1000, emit);

    t.schedule("a"); // leading at t=1000
    vi.advanceTimersByTime(1000); // now t=2000, window elapsed
    t.schedule("b"); // leading again, no pending

    expect(emit).toHaveBeenCalledTimes(2);
    expect(emit).toHaveBeenLastCalledWith("b");
  });
});
