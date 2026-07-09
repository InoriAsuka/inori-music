/**
 * Generic Pagination component — works with the v1 offset/limit pagination API.
 * CatalogPaginationMeta shape: { limit, offset, total, hasMore }
 */
"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface OffsetPagination {
  limit: number;
  offset: number;
  total: number;
  hasMore: boolean;
}

interface PaginationBarProps {
  pagination: OffsetPagination;
  onPageChange: (page: number) => void;
  className?: string;
}

/** Convert offset/limit to a 1-based page number. */
export function pageFromOffset(offset: number, limit: number): number {
  return Math.floor(offset / limit) + 1;
}

/** Convert 1-based page number to offset. */
export function offsetFromPage(page: number, limit: number): number {
  return (page - 1) * limit;
}

export function PaginationBar({ pagination, onPageChange, className }: PaginationBarProps) {
  const { limit, offset, total } = pagination;
  const page = pageFromOffset(offset, limit);
  const totalPages = Math.ceil(total / limit);

  if (totalPages <= 1) return null;

  const pages = buildPageList(page, totalPages);

  return (
    <div className={cn("flex items-center justify-between gap-2 text-sm", className)}>
      <span className="text-[var(--color-muted-foreground)]">
        {total} item{total !== 1 ? "s" : ""}
      </span>
      <div className="flex items-center gap-1">
        <PageBtn onClick={() => onPageChange(page - 1)} disabled={page <= 1} aria-label="Previous">
          <ChevronLeft size={14} />
        </PageBtn>
        {pages.map((p, i) =>
          p === "…" ? (
            <span key={`ellipsis-${i}`} className="px-2 text-[var(--color-muted-foreground)]">
              …
            </span>
          ) : (
            <PageBtn key={p} onClick={() => onPageChange(p as number)} active={p === page} aria-label={`Page ${p}`}>
              {p}
            </PageBtn>
          )
        )}
        <PageBtn onClick={() => onPageChange(page + 1)} disabled={page >= totalPages} aria-label="Next">
          <ChevronRight size={14} />
        </PageBtn>
      </div>
    </div>
  );
}

function PageBtn({
  children,
  onClick,
  disabled,
  active,
  "aria-label": ariaLabel,
}: {
  children: React.ReactNode;
  onClick: () => void;
  disabled?: boolean;
  active?: boolean;
  "aria-label"?: string;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
      className={cn(
        "flex h-8 w-8 items-center justify-center rounded-md text-sm transition-colors",
        active
          ? "bg-[var(--color-primary)] text-[var(--color-primary-foreground)]"
          : "hover:bg-[var(--color-muted)] text-[var(--color-foreground)]",
        disabled && "opacity-40 pointer-events-none"
      )}
    >
      {children}
    </button>
  );
}

function buildPageList(current: number, total: number): (number | "…")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const pages: (number | "…")[] = [1];
  if (current > 3) pages.push("…");
  const lo = Math.max(2, current - 1);
  const hi = Math.min(total - 1, current + 1);
  for (let p = lo; p <= hi; p++) pages.push(p);
  if (current < total - 2) pages.push("…");
  pages.push(total);
  return pages;
}
