/**
 * PlayerPopover — a small anchored menu used by the speed and sleep-timer
 * controls in the player bar.
 *
 * Accessibility / dismissal (project conventions, no new dependency):
 *   • Trigger is a real <button> with aria-haspopup / aria-expanded.
 *   • Panel has role="menu" with wrapping ArrowUp / ArrowDown and Home / End
 *     navigation; Escape closes it and restores focus to the trigger.
 *   • On open, focus moves to the requested item (or the first menu item), and
 *     outside dismissal restores the trigger only if focus would be lost.
 *
 * The panel opens upward by default since the player bar sits at the bottom of
 * the viewport; fullscreen callers can opt into downward placement.
 */
"use client";

import { AnimatePresence, motion } from "motion/react";
import { useEffect, useId, useRef, useState } from "react";
import { cn } from "@/lib/utils";

const MENU_ITEM_SELECTOR = '[role="menuitem"], [role="menuitemradio"]';

export type PlayerPopoverAlign = "left" | "center" | "right";
export type PlayerPopoverPlacement = "above" | "below";

function getMenuItems(panel: HTMLElement | null): HTMLElement[] {
  if (!panel) return [];
  return Array.from(panel.querySelectorAll<HTMLElement>(MENU_ITEM_SELECTOR)).filter(
    (item) => item.getAttribute("aria-disabled") !== "true"
  );
}

function focusMenuItem(items: HTMLElement[], index: number) {
  items.forEach((item, itemIndex) => {
    item.tabIndex = itemIndex === index ? 0 : -1;
  });
  items[index]?.focus();
}

export function PlayerPopover({
  label,
  title,
  active,
  badge,
  children,
  align = "center",
  placement = "above",
}: {
  /** Accessible label for the trigger button. */
  label: string;
  /** Trigger icon / content. */
  title: React.ReactNode;
  /** Whether the underlying feature is active (highlights the trigger). */
  active?: boolean;
  /** Optional small badge rendered over the trigger (e.g. "1.5×" or countdown). */
  badge?: React.ReactNode;
  /** Menu contents (rendered inside role="menu"). */
  children: React.ReactNode;
  align?: PlayerPopoverAlign;
  placement?: PlayerPopoverPlacement;
}) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useEffect(() => {
    if (!open) return;

    function onPointerDown(e: PointerEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        const focusWasInMenu = panelRef.current?.contains(document.activeElement) ?? false;
        setOpen(false);
        if (focusWasInMenu) {
          requestAnimationFrame(() => {
            if (document.activeElement === document.body) triggerRef.current?.focus();
          });
        }
      }
    }
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        e.preventDefault();
        e.stopPropagation();
        setOpen(false);
        triggerRef.current?.focus();
        return;
      }
      if (!panelRef.current || !["ArrowDown", "ArrowUp", "Home", "End"].includes(e.key)) return;
      const items = getMenuItems(panelRef.current);
      if (items.length === 0) return;
      e.preventDefault();
      const currentIndex = items.indexOf(document.activeElement as HTMLElement);
      if (e.key === "Home") focusMenuItem(items, 0);
      else if (e.key === "End") focusMenuItem(items, items.length - 1);
      else if (e.key === "ArrowDown") focusMenuItem(items, (currentIndex + 1 + items.length) % items.length);
      else focusMenuItem(items, (currentIndex - 1 + items.length) % items.length);
    }
    document.addEventListener("pointerdown", onPointerDown);
    document.addEventListener("keydown", onKeyDown);
    // Move focus into the menu so it's immediately keyboard-navigable.
    const items = getMenuItems(panelRef.current);
    const autoFocusItem = panelRef.current?.querySelector<HTMLElement>("[data-autofocus]");
    const initialIndex = autoFocusItem ? items.indexOf(autoFocusItem) : 0;
    if (items.length > 0) focusMenuItem(items, Math.max(0, initialIndex));
    return () => {
      document.removeEventListener("pointerdown", onPointerDown);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [open]);

  return (
    <div ref={wrapperRef} className="relative flex items-center">
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-label={label}
        title={label}
        onClick={() => setOpen((v) => !v)}
        className={cn(
          "relative flex h-10 w-10 items-center justify-center rounded transition-colors",
          active || open
            ? "text-[var(--color-primary)]"
            : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
        )}
      >
        {title}
        {badge != null && (
          <span className="absolute -right-1 -top-1 min-w-[1.1rem] rounded-full bg-[var(--color-primary)] px-1 text-center text-[9px] font-bold leading-4 text-[var(--color-primary-fg)]">
            {badge}
          </span>
        )}
      </button>

      <AnimatePresence>
        {open && (
          <motion.div
            ref={panelRef}
            id={menuId}
            role="menu"
            aria-label={label}
            initial={{ opacity: 0, y: placement === "above" ? 6 : -6, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: placement === "above" ? 6 : -6, scale: 0.98 }}
            transition={{ duration: 0.14 }}
            className={cn(
              "absolute z-50 max-h-[min(22rem,calc(100dvh-5rem))] min-w-[10rem] overflow-y-auto rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-1 shadow-2xl",
              placement === "above" ? "bottom-full mb-2" : "top-full mt-2",
              align === "left"
                ? "left-0"
                : align === "right"
                  ? "right-0"
                  : "left-1/2 -translate-x-1/2"
            )}
            onClick={(e) => e.stopPropagation()}
          >
            {children}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

/**
 * A single selectable row inside a PlayerPopover menu. Renders as
 * role="menuitemradio" so assistive tech announces the current selection.
 */
export function PlayerPopoverItem({
  children,
  onSelect,
  selected,
  autoFocus,
}: {
  children: React.ReactNode;
  onSelect: () => void;
  selected?: boolean;
  autoFocus?: boolean;
}) {
  return (
    <button
      type="button"
      role="menuitemradio"
      aria-checked={selected}
      data-autofocus={autoFocus ? "" : undefined}
      onClick={onSelect}
      className={cn(
        "flex w-full items-center justify-between gap-4 rounded-md px-3 py-1.5 text-left text-sm transition-colors",
        selected
          ? "bg-[var(--color-primary-dim)] text-[var(--color-primary)]"
          : "text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-raised)] hover:text-[var(--color-text)]"
      )}
    >
      {children}
    </button>
  );
}
