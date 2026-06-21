import { cn } from "./utils";

type HealthStatus = "healthy" | "unhealthy" | "unknown" | "disabled";

const CONFIG: Record<HealthStatus, { color: string; label: string }> = {
  healthy:   { color: "bg-[var(--color-success)] text-[var(--color-success)]", label: "Healthy" },
  unhealthy: { color: "bg-[var(--color-danger)] text-[var(--color-danger)]",   label: "Unhealthy" },
  unknown:   { color: "bg-[var(--color-text-muted)] text-[var(--color-text-muted)]", label: "Unknown" },
  disabled:  { color: "bg-[var(--color-text-muted)] text-[var(--color-text-muted)]", label: "Disabled" },
};

export function StorageHealthBadge({ status }: { status?: string }) {
  const s = (status as HealthStatus) ?? "unknown";
  const cfg = CONFIG[s] ?? CONFIG.unknown;
  return (
    <span className={cn("inline-flex items-center gap-1.5 text-xs font-medium", cfg.color)}>
      <span className={cn("inline-block h-2 w-2 rounded-full pulse-dot", cfg.color.split(" ")[0])} />
      {cfg.label}
    </span>
  );
}
