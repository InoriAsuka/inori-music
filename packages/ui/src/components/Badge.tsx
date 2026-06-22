import { cn } from "./utils";

type BadgeVariant = "default" | "primary" | "success" | "warning" | "danger" | "info";

const VARIANTS: Record<BadgeVariant, string> = {
  default: "border-[var(--color-border)] text-[var(--color-text-muted)]",
  primary: "border-[var(--color-primary)] bg-[var(--color-primary-dim)] text-[var(--color-primary)]",
  success: "border-[var(--color-success)] bg-[color-mix(in_srgb,var(--color-success)_10%,transparent)] text-[var(--color-success)]",
  warning: "border-[var(--color-warning)] bg-[color-mix(in_srgb,var(--color-warning)_10%,transparent)] text-[var(--color-warning)]",
  danger: "border-[var(--color-danger)] bg-[var(--color-danger-dim)] text-[var(--color-danger)]",
  info: "border-[var(--color-info)] bg-[color-mix(in_srgb,var(--color-info)_10%,transparent)] text-[var(--color-info)]",
};

export function Badge({
  children,
  variant = "default",
  className,
}: {
  children?: any;
  variant?: BadgeVariant;
  className?: string;
}) {
  return (
    <span className={cn("inline-flex items-center rounded-full border px-2 py-0.5 text-[10px] font-semibold", VARIANTS[variant], className)}>
      {children}
    </span>
  );
}
