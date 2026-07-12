import { parseHighlight } from "@/lib/search/highlight";

/**
 * Renders a backend highlight snippet (e.g. "hello <mark>w</mark>orld") as
 * safe React spans — no dangerouslySetInnerHTML. Falls back to plain text
 * when raw is absent or has no <mark> tags.
 */
export function Highlighted({ raw, plain, className }: { raw?: string | null; plain: string; className?: string }) {
  const segments = parseHighlight(raw);
  if (!segments) {
    return <span className={className}>{plain}</span>;
  }
  return (
    <span className={className}>
      {segments.map((seg, i) =>
        seg.marked ? (
          <mark key={i} className="bg-transparent font-semibold text-[var(--color-primary)]">
            {seg.text}
          </mark>
        ) : (
          <span key={i}>{seg.text}</span>
        )
      )}
    </span>
  );
}
