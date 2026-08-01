import type { LyricLine, LyricWord } from "@/lib/lyrics/lyricLine";

/** Index of the last line whose timestamp has been reached, or -1 before the first. */
export function activeLineIndex(lines: LyricLine[], positionMs: number): number {
  let idx = -1;
  for (let i = 0; i < lines.length; i++) {
    if (lines[i].timestampMs <= positionMs) idx = i;
    else break;
  }
  return idx;
}

/**
 * Fill ratio (0..1) for one word, used to render the sung portion.
 *
 * A word's duration runs until the next word starts, or — for the last word of
 * a line — until the next line starts. Without a following line, fall back to a
 * fixed tail so the final word still animates instead of snapping.
 */
export function wordProgress(
  words: LyricWord[],
  index: number,
  positionMs: number,
  lineEndMs: number | null
): number {
  const word = words[index];
  if (!word) return 0;
  if (positionMs <= word.offsetMs) return 0;

  const next = words[index + 1];
  const endMs = next ? next.offsetMs : (lineEndMs ?? word.offsetMs + FINAL_WORD_TAIL_MS);
  const span = endMs - word.offsetMs;
  if (span <= 0) return 1;
  return Math.min(1, (positionMs - word.offsetMs) / span);
}

const FINAL_WORD_TAIL_MS = 800;
