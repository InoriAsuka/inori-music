import { cn } from "./utils";

type LifecycleState = "active" | "archived" | "deleted" | string;

const STATE_STYLES: Record<string, string> = {
  active:   "border-[var(--color-success)] bg-[color-mix(in_srgb,var(--color-success)_12%,transparent)] text-[var(--color-success)]",
  archived: "border-[var(--color-warning)] bg-[color-mix(in_srgb,var(--color-warning)_12%,transparent)] text-[var(--color-warning)]",
  deleted:  "border-[var(--color-danger)] bg-[var(--color-danger-dim)] text-[var(--color-danger)]",
};

export function LifecyclePill({ state }: { state: LifecycleState }) {
  return (
    <span className={cn("inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-semibold", STATE_STYLES[state] ?? "border-[var(--color-border)] text-[var(--color-text-muted)]")}>
      {state}
    </span>
  );
}
