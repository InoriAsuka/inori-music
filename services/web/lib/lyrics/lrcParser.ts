import type { LyricLine, LyricWord } from "./lyricLine";

/**
 * Parses LRC-format lyrics into a list of LyricLines sorted by timestamp.
 *
 * Handles standard [MM:SS.xx] and [MM:SS.xxx] timestamps.
 * Also parses enhanced (word-level karaoke) LRC inline tags of the form
 * `<mm:ss.xx>word`, producing per-word timing fragments in LyricLine.words.
 * Lines without inline tags keep LyricLine.words as undefined.
 * Metadata tags like [ti:], [ar:], [al:] are ignored.
 */
const LINE_TAG_REGEX = /^\[(\d{2}):(\d{2})[.:](\d{2,3})\](.*)$/;
const WORD_TAG_REGEX = /<(\d{2}):(\d{2})[.:](\d{2,3})>/g;

export function parseLrc(content: string): LyricLine[] {
  const lines: LyricLine[] = [];
  for (const raw of content.split("\n")) {
    const line = raw.trim();
    const m = LINE_TAG_REGEX.exec(line);
    if (!m) continue;
    const timestampMs = parseTimestamp(m[1], m[2], m[3]);
    const rawText = (m[4] ?? "").trim();
    if (rawText.length === 0) continue; // skip metadata-only tags

    const words = parseWords(rawText, timestampMs);
    const text = words ? words.map((w) => w.text).join("").trim() : rawText;
    if (text.length === 0) continue;
    lines.push({ timestampMs, text, words });
  }
  lines.sort((a, b) => a.timestampMs - b.timestampMs);
  return lines;
}

/**
 * Splits a line's raw text on inline <mm:ss.xx> tags into word fragments.
 * Returns undefined when the line has no inline tags (plain line-level lyrics).
 * Text preceding the first inline tag, if any, is kept as a fragment
 * timed at the line's own lineTimestampMs.
 */
function parseWords(rawText: string, lineTimestampMs: number): LyricWord[] | undefined {
  const matches = [...rawText.matchAll(WORD_TAG_REGEX)];
  if (matches.length === 0) return undefined;
  const words: LyricWord[] = [];
  const first = matches[0];
  if (first.index! > 0) {
    words.push({ offsetMs: lineTimestampMs, text: rawText.slice(0, first.index) });
  }
  for (let i = 0; i < matches.length; i++) {
    const m = matches[i];
    const offsetMs = parseTimestamp(m[1], m[2], m[3]);
    const start = m.index! + m[0].length;
    const end = i + 1 < matches.length ? matches[i + 1].index! : rawText.length;
    if (end <= start) continue;
    words.push({ offsetMs, text: rawText.slice(start, end) });
  }
  return words.length === 0 ? undefined : words;
}

function parseTimestamp(minStr: string, secStr: string, csStr: string): number {
  const min = Number.parseInt(minStr, 10);
  const sec = Number.parseInt(secStr, 10);
  // normalise to milliseconds: 2 digits = centiseconds, 3 = milliseconds
  const ms = csStr.length === 2 ? Number.parseInt(csStr, 10) * 10 : Number.parseInt(csStr, 10);
  return min * 60_000 + sec * 1000 + ms;
}
