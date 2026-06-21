"use client";

import { formatDateTime } from "@/lib/utils";

interface TimelineEvent {
  label: string;
  timestamp: string;
  detail?: string;
  variant?: "default" | "primary" | "success" | "danger";
}

const VARIANT_DOT: Record<string, string> = {
  default: "bg-[var(--color-border-glow)]",
  primary: "bg-[var(--color-primary)]",
  success: "bg-[var(--color-success)]",
  danger:  "bg-[var(--color-danger)]",
};

export function MediaObjectTimeline({ events }: { events: TimelineEvent[] }) {
  if (events.length === 0) {
    return <p className="text-sm text-[var(--color-text-muted)]">No timeline events.</p>;
  }

  return (
    <ol className="relative ml-3 border-l border-[var(--color-border)]">
      {events.map((ev, i) => (
        <li key={i} className="mb-6 ml-5">
          <span className={`absolute -left-1.5 flex h-3 w-3 items-center justify-center rounded-full ${VARIANT_DOT[ev.variant ?? "default"]}`} />
          <p className="text-xs text-[var(--color-text-muted)]">{formatDateTime(ev.timestamp)}</p>
          <h3 className="text-sm font-semibold text-[var(--color-text)]">{ev.label}</h3>
          {ev.detail && <p className="font-mono text-xs text-[var(--color-text-muted)]">{ev.detail}</p>}
        </li>
      ))}
    </ol>
  );
}
