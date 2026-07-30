"use client";

import { useEqStore } from "@/store/eq";
import { EQ_BANDS, type EqBandId } from "@/lib/audio/eqPresets";
import { X } from "lucide-react";

export function EqualizerPanel({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { enabled, preset, gains, toggle, setPreset, setBand, reset } = useEqStore();

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-[var(--color-scrim)]" onClick={onClose}>
      <div
        className="w-full max-w-lg rounded-2xl border border-[var(--color-border)] bg-[var(--color-surface)] p-6 space-y-5"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between">
          <h2 className="text-base font-semibold text-[var(--color-text)]">Equalizer</h2>
          <button type="button" onClick={onClose} className="rounded p-1.5 text-[var(--color-text-muted)] hover:text-[var(--color-text)]">
            <X size={16} />
          </button>
        </div>

        <div className="flex items-center justify-between">
          <label className="flex items-center gap-2 text-sm text-[var(--color-text)]">
            <input type="checkbox" checked={enabled} onChange={toggle} className="h-4 w-4 rounded border-[var(--color-border)]" />
            Enable EQ
          </label>
          <select
            value={preset}
            onChange={(e) => setPreset(e.target.value as typeof preset)}
            className="rounded-md border border-[var(--color-border)] bg-[var(--color-void)] px-2 py-1.5 text-xs text-[var(--color-text)] outline-none focus:border-[var(--color-primary)]"
          >
            <option value="flat">Flat</option>
            <option value="bass-boost">Bass Boost</option>
            <option value="treble-boost">Treble Boost</option>
            <option value="vocal">Vocal</option>
            <option value="electronic">Electronic</option>
          </select>
        </div>

        <div className="grid grid-cols-10 gap-2">
          {EQ_BANDS.map((band) => (
            <div key={band.id} className="flex flex-col items-center gap-1">
              <input
                type="range"
                min={band.minDb}
                max={band.maxDb}
                step={band.stepDb}
                value={gains[band.id]}
                onChange={(e) => setBand(band.id, Number.parseFloat(e.target.value))}
                disabled={!enabled}
                className="h-32 w-2 appearance-none rounded-full bg-[var(--color-border)] accent-[var(--color-primary)] disabled:opacity-40"
                style={{ writingMode: "vertical-lr", direction: "rtl" }}
              />
              <span className="text-[10px] text-[var(--color-text-muted)]">{band.label}</span>
              <span className="text-[10px] font-medium text-[var(--color-text)]">{gains[band.id] > 0 ? "+" : ""}{gains[band.id].toFixed(1)}</span>
            </div>
          ))}
        </div>

        <div className="flex justify-end">
          <button
            type="button"
            onClick={reset}
            className="rounded-md border border-[var(--color-border)] px-3 py-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]"
          >
            Reset
          </button>
        </div>
      </div>
    </div>
  );
}
