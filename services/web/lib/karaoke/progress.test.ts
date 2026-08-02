import { describe, expect, it } from "vitest";
import { activeLineIndex, wordProgress } from "@/lib/karaoke/progress";
import type { LyricLine, LyricWord } from "@/lib/lyrics/lyricLine";

/**
 * These cases mirror `services/mobile/test/karaoke_progress_test.dart`
 * one-for-one so both clients highlight identically.
 */

const line = (timestampMs: number, text: string, words?: LyricWord[]): LyricLine => ({
  timestampMs,
  text,
  words,
});
const word = (offsetMs: number, text: string): LyricWord => ({ offsetMs, text });

describe("activeLineIndex", () => {
  const lines = [line(0, "a"), line(1000, "b"), line(2000, "c")];

  it("returns -1 before the first line starts", () => {
    expect(activeLineIndex([line(500, "a")], 0)).toBe(-1);
  });

  it("returns the last reached line", () => {
    expect(activeLineIndex(lines, 1500)).toBe(1);
  });

  it("is inclusive of the exact timestamp", () => {
    expect(activeLineIndex(lines, 1000)).toBe(1);
  });

  it("stays on the final line past the end", () => {
    expect(activeLineIndex(lines, 99_000)).toBe(2);
  });

  it("handles an empty list", () => {
    expect(activeLineIndex([], 1000)).toBe(-1);
  });
});

describe("wordProgress", () => {
  const words = [word(1000, "one"), word(2000, "two")];

  it("is 0 before the word starts", () => {
    expect(wordProgress(words, 0, 500, null)).toBe(0);
  });

  it("is 0 exactly at the word start", () => {
    expect(wordProgress(words, 0, 1000, null)).toBe(0);
  });

  it("interpolates linearly to the next word", () => {
    expect(wordProgress(words, 0, 1500, null)).toBeCloseTo(0.5, 9);
  });

  it("is 1 once the next word has started", () => {
    expect(wordProgress(words, 0, 2500, null)).toBe(1);
  });

  it("last word uses the next line start as its end", () => {
    expect(wordProgress(words, 1, 2500, 3000)).toBeCloseTo(0.5, 9);
  });

  it("last word falls back to a fixed tail without a next line", () => {
    // Tail is 800ms, so +400ms is half way.
    expect(wordProgress(words, 1, 2400, null)).toBeCloseTo(0.5, 9);
  });

  it("never exceeds 1 past the tail", () => {
    expect(wordProgress(words, 1, 60_000, null)).toBe(1);
  });

  it("returns 1 for a zero-length span rather than dividing by zero", () => {
    const same = [word(1000, "a"), word(1000, "b")];
    expect(wordProgress(same, 0, 1200, null)).toBe(1);
  });

  it("out-of-range index is 0", () => {
    expect(wordProgress(words, 9, 5000, null)).toBe(0);
    expect(wordProgress(words, -1, 5000, null)).toBe(0);
  });
});
