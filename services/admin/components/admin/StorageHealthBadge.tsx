/** Re-export from packages/ui for convenience within services/admin. */
import { cn } from "@/lib/utils";

type HealthStatus = "healthy" | "unhealthy" | "unknown" | "disabled" | string;

const CONFIG: Record<string, { textColor: string; dotColor: string; label: string }> = {
  healthy:   { textColor: "text-[var(--color-success)]", dotColor: "bg-[var(--color-success)]", label: "Healthy" },
  unhealthy: { textColor: "text-[var(--color-danger)]",  dotColor: "bg-[var(--color-danger)]",  label: "Unhealthy" },
  disabled:  { textColor: "text-[var(--color-text-muted)]", dotColor: "bg-[var(--color-text-muted)]", label: "Disabled" },
  unknown:   { textColor: "text-[var(--color-text-muted)]", dotColor: "bg-[var(--color-text-muted)]", label: "Unknown" },
};

export function StorageHealthBadge({ status }: { status?: HealthStatus }) {
  const s = status ?? "unknown";
  const cfg = CONFIG[s] ?? CONFIG.unknown;
  return (
    <span className={cn("inline-flex items-center gap-1.5 text-xs font-medium", cfg.textColor)}>
      <span className={cn("inline-block h-2 w-2 rounded-full", cfg.dotColor, s === "healthy" && "pulse-dot")} />
      {cfg.label}
    </span>
  );
}
