import type { LyricLine } from "./lyricLine";
import { parseLrc } from "./lrcParser";
import { parseSrt } from "./srtParser";

interface LyricsResponse {
  format: "lrc" | "srt";
  content: string;
  translation?: string;
}

/**
 * Fetches and parses lyrics for trackId via the given API base URL.
 * Returns null when the track has no lyrics (404) or on any other error.
 */
export async function fetchLyrics(
  trackId: string,
  opts: { baseUrl?: string; token?: string } = {}
): Promise<LyricLine[] | null> {
  if (!trackId) return null;
  try {
    const res = await fetch(`${opts.baseUrl ?? ""}/api/v1/catalog/tracks/${trackId}/lyrics`, {
      headers: opts.token ? { Authorization: `Bearer ${opts.token}` } : undefined,
    });
    if (res.status === 404) return null;
    if (!res.ok) return null;
    const data = (await res.json()) as LyricsResponse;
    if (!data.content) return null;

    let lines: LyricLine[] | null;
    if (data.format === "lrc") {
      lines = parseLrc(data.content);
    } else if (data.format === "srt") {
      lines = parseSrt(data.content);
    } else {
      return null;
    }

    if (data.translation) {
      lines = mergeTranslation(lines, data.translation);
    }
    return lines;
  } catch {
    return null;
  }
}

/**
 * Aligns translation text to lines by splitting on newlines and assigning
 * one translation line per lyric line, in order. When the line counts
 * don't match, the translation is dropped (to avoid misaligned pairing).
 */
function mergeTranslation(lines: LyricLine[], translation: string): LyricLine[] {
  const translationLines = translation
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
  if (translationLines.length !== lines.length) {
    return lines;
  }
  return lines.map((line, i) => ({ ...line, translation: translationLines[i] }));
}
