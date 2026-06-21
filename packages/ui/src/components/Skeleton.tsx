import { cn } from "./utils";

export function Skeleton({ className }: { className?: string }) {
  return <div className={cn("animate-pulse rounded-md bg-[var(--color-surface-raised)]", className)} />;
}
