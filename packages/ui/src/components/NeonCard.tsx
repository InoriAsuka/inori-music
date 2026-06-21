import { cn } from "./utils";

interface NeonCardProps extends React.HTMLAttributes<HTMLDivElement> {
  glow?: "primary" | "secondary" | "sakura" | "none";
}

const GLOW_MAP = {
  primary: "shadow-[0_0_16px_3px_color-mix(in_srgb,#9b5cff_25%,transparent)]",
  secondary: "shadow-[0_0_16px_3px_color-mix(in_srgb,#0fd4c0_25%,transparent)]",
  sakura: "shadow-[0_0_16px_3px_color-mix(in_srgb,#ff5fa0_25%,transparent)]",
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
