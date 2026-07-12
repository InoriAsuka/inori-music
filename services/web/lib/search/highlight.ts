/**
 * Parses a backend highlight snippet like "hello <mark>w</mark>orld" into
 * plain text/marked segments, for safe rendering without dangerouslySetInnerHTML.
 * Returns null when raw is null/blank/has no <mark> tags, signalling the
 * caller to render plain text instead.
 */
export interface HighlightSegment {
  text: string;
  marked: boolean;
}

const MARK_OPEN = "<mark>";
const MARK_CLOSE = "</mark>";

export function parseHighlight(raw: string | null | undefined): HighlightSegment[] | null {
  if (!raw) return null;
  if (!raw.includes(MARK_OPEN)) return null;

  const out: HighlightSegment[] = [];
  let rest = raw;
  while (rest.length > 0) {
    const o = rest.indexOf(MARK_OPEN);
    if (o < 0) {
      out.push({ text: rest, marked: false });
      break;
    }
    if (o > 0) {
      out.push({ text: rest.slice(0, o), marked: false });
    }
    const c = rest.indexOf(MARK_CLOSE, o);
    if (c < 0) {
      // Mismatched tag: emit the rest as plain text.
      out.push({ text: rest.slice(o), marked: false });
      break;
    }
    const inner = rest.slice(o + MARK_OPEN.length, c);
    if (inner.length > 0) {
      out.push({ text: inner, marked: true });
    }
    rest = rest.slice(c + MARK_CLOSE.length);
  }
  return out;
}
