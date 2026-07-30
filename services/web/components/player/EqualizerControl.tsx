"use client";

import { useEqStore } from "@/store/eq";
import { SlidersHorizontal } from "lucide-react";

export function EqualizerControl({ onClick }: { onClick: () => void }) {
  const { enabled } = useEqStore();
  return (
    <button
      type="button"
      onClick={onClick}
      title="Equalizer"
      className={`transition-colors ${enabled ? "text-[var(--color-primary)]" : "text-[var(--color-text-muted)] hover:text-[var(--color-text)]"}`}
    >
      <SlidersHorizontal size={16} />
    </button>
  );
}
