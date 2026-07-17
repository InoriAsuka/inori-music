/**
 * Leading-edge throttle with trailing coalesce (v5.4.0 cross-device reporting).
 *
 * Mirrors the coalescing behaviour of the player store's `createThrottledStorage`
 * but generalised for arbitrary payloads: the FIRST `schedule` after an idle
 * window emits immediately (leading edge), further calls within `intervalMs`
 * are coalesced to a single trailing emit carrying the LATEST value. `flush`
 * forces the pending value out now (used for urgent transitions like a song
 * change); `cancel` drops any pending emit and timer (used on logout).
 *
 * Uses `Date.now()` + `setTimeout`, so tests drive it with
 * `vi.useFakeTimers()` + `vi.setSystemTime()` (see throttle.test.ts).
 */
export interface Throttle<T> {
  /** Queue `value`; emits now if the window has elapsed, else schedules a trailing emit. */
  schedule(value: T): void;
  /** Emit the pending value immediately (no-op if nothing pending). */
  flush(): void;
  /** Drop the pending value and cancel the timer without emitting. */
  cancel(): void;
}

export function createThrottle<T>(intervalMs: number, emit: (value: T) => void): Throttle<T> {
  let lastEmitAt = 0;
  let timer: ReturnType<typeof setTimeout> | null = null;
  let pending: { value: T } | null = null;

  function fire() {
    timer = null;
    if (!pending) return;
    const { value } = pending;
    pending = null;
    lastEmitAt = Date.now();
    emit(value);
  }

  function clearTimer() {
    if (timer) {
      clearTimeout(timer);
      timer = null;
    }
  }

  return {
    schedule(value) {
      pending = { value };
      const elapsed = Date.now() - lastEmitAt;
      if (elapsed >= intervalMs) {
        clearTimer();
        fire();
        return;
      }
      if (!timer) timer = setTimeout(fire, intervalMs - elapsed);
    },
    flush() {
      clearTimer();
      fire();
    },
    cancel() {
      clearTimer();
      pending = null;
    },
  };
}
