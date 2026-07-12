/** A single word-level timing fragment within a LyricLine, used for karaoke-style progressive highlighting. */
export interface LyricWord {
  /** Offset in milliseconds from the start of the track. */
  offsetMs: number;
  text: string;
}

/** A single line of lyrics with its start timestamp. */
export interface LyricLine {
  /** Timestamp in milliseconds from the start of the track. */
  timestampMs: number;
  text: string;
  /** Word-level timing fragments for karaoke highlighting, or undefined when the source line has no inline `<mm:ss.xx>` tags. */
  words?: LyricWord[];
  /** Aligned translation text for this line, or undefined when no translation was uploaded or the translation line count didn't match. */
  translation?: string;
}
