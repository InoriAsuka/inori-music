import { cn } from "./utils";

interface NeonCardProps {
  children?: any;
  className?: string;
  glow?: "primary" | "secondary" | "sakura" | "none";
  [key: string]: any;
}

const GLOW_MAP = {
  primary: "shadow-[0_0_8px_0_color-mix(in_srgb,var(--color-primary)_25%,transparent)]",
  secondary: "shadow-[0_0_8px_0_color-mix(in_srgb,var(--color-secondary)_25%,transparent)]",
  sakura: "shadow-[0_0_8px_0_color-mix(in_srgb,var(--color-sakura)_25%,transparent)]",
  none: "",
};

export function NeonCard({ children, className, glow = "none", ...props }: NeonCardProps) {
  return (
    <div
      className={cn(
        "rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] transition-colors",
        GLOW_MAP[glow],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
