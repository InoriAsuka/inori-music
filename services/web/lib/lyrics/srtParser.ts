import type { LyricLine } from "./lyricLine";

/** Parses SRT-format subtitles into a list of LyricLines (start-time only). */
const TIME_REGEX = /(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*/;

export function parseSrt(content: string): LyricLine[] {
  const lines: LyricLine[] = [];
  // SRT blocks are separated by blank lines
  const blocks = content.split(/\r?\n\r?\n/);
  for (const block of blocks) {
    const blockLines = block.trim().split(/\r?\n/);
    if (blockLines.length < 2) continue;
    // First non-empty line is the sequence number — skip it.
    // Second line is the time range.
    let timeLine: string | undefined;
    let textStart = 1;
    for (let i = 0; i < blockLines.length; i++) {
      if (TIME_REGEX.test(blockLines[i])) {
        timeLine = blockLines[i];
        textStart = i + 1;
        break;
      }
    }
    if (!timeLine) continue;
    const m = TIME_REGEX.exec(timeLine)!;
    const h = Number.parseInt(m[1], 10);
    const min = Number.parseInt(m[2], 10);
    const sec = Number.parseInt(m[3], 10);
    const ms = Number.parseInt(m[4], 10);
    const text = blockLines.slice(textStart).join(" ").trim();
    if (text.length === 0) continue;
    lines.push({
      timestampMs: h * 3_600_000 + min * 60_000 + sec * 1000 + ms,
      text,
    });
  }
  lines.sort((a, b) => a.timestampMs - b.timestampMs);
  return lines;
}
