"use client";

import { cn } from "@/lib/utils";

export function EmptyState({
  title,
  description,
  className,
}: {
  title: string;
  description?: string;
  className?: string;
}) {
  return (
    <div className={cn("rounded-xl border border-dashed border-[var(--color-border)] p-8 text-center", className)}>
      <p className="font-medium">{title}</p>
      {description && <p className="mt-1 text-sm text-[var(--color-muted-foreground)]">{description}</p>}
    </div>
  );
}
