"use client";

/**
 * Modal — minimal accessible dialog primitive.
 *
 * The repo had no dialog component (confirmations used window.confirm), so this
 * fills the gap for the playlist create/rename/add flows. Provides:
 * - role="dialog" + aria-modal + aria-labelledby wiring
 * - Escape-to-close and backdrop-click-to-close
 * - initial focus into the panel and body scroll lock while open
 *
 * Deliberately small — animation matches QueueDrawer's motion/backdrop feel.
 * Destructive confirmations that don't need a form still use window.confirm.
 */
import { useEffect, useId, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export function Modal({ open, onClose, title, children }: ModalProps) {
  const panelRef = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const [mounted, setMounted] = useState(false);
  const previousFocusRef = useRef<HTMLElement | null>(null);

  useEffect(() => setMounted(true), []);

  useEffect(() => {
    if (!open) return;
    previousFocusRef.current = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        onClose();
        return;
      }
      if (e.key !== "Tab" || !panelRef.current) return;
      const focusable = Array.from(
        panelRef.current.querySelectorAll<HTMLElement>(
          'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
        )
      );
      if (focusable.length === 0) {
        e.preventDefault();
        panelRef.current.focus();
        return;
      }
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    }
    document.addEventListener("keydown", onKey);
    document.body.style.overflow = "hidden";
    if (panelRef.current && !panelRef.current.open) panelRef.current.showModal();
    panelRef.current?.focus();
    return () => {
      document.removeEventListener("keydown", onKey);
      document.body.style.overflow = "";
      if (panelRef.current?.open) panelRef.current.close();
      previousFocusRef.current?.focus();
    };
  }, [open, onClose]);

  if (!mounted) return null;

  return createPortal(
    <AnimatePresence>
      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <motion.div
            className="absolute inset-0 bg-black/60"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onClose}
          />
          <motion.dialog
            ref={panelRef}
            aria-labelledby={titleId}
            tabIndex={-1}
            onCancel={(event) => {
              event.preventDefault();
              onClose();
            }}
            initial={{ opacity: 0, scale: 0.96, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.96, y: 8 }}
            transition={{ duration: 0.15 }}
            className="relative z-10 w-full max-w-md rounded-xl border border-[var(--color-border)] bg-[var(--color-card)] shadow-2xl outline-none"
          >
            <div className="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3">
              <h2 id={titleId} className="text-sm font-semibold">
                {title}
              </h2>
              <button
                type="button"
                onClick={onClose}
                aria-label="Close"
                className="rounded p-1.5 text-[var(--color-muted-foreground)] hover:bg-[var(--color-muted)] hover:text-[var(--color-foreground)]"
              >
                <X size={16} />
              </button>
            </div>
            <div className="p-4">{children}</div>
          </motion.dialog>
        </div>
      )}
    </AnimatePresence>,
    document.body
  );
}
