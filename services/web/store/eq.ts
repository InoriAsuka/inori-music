"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";
import { DEFAULT_EQ_GAINS, EQ_BANDS, type EqBandId, EQ_PRESETS, type EqPresetName } from "@/lib/audio/eqPresets";

export interface EqState {
  enabled: boolean;
  /** A preset key, or "custom" once any band is hand-adjusted (matches Flutter). */
  preset: EqPresetName | "custom";
  gains: Record<EqBandId, number>;
  toggle: () => void;
  setPreset: (preset: EqPresetName) => void;
  setBand: (id: EqBandId, db: number) => void;
  reset: () => void;
}

export const useEqStore = create<EqState>()(
  persist(
    (set) => ({
      enabled: false,
      preset: "flat",
      gains: { ...DEFAULT_EQ_GAINS },
      toggle: () => set((s) => ({ enabled: !s.enabled })),
      setPreset: (preset) =>
        set({
          preset,
          gains: { ...EQ_PRESETS[preset] },
        }),
      setBand: (id, db) =>
        set((s) => ({
          gains: { ...s.gains, [id]: db },
          preset: "custom",
        })),
      reset: () =>
        set({
          enabled: false,
          preset: "flat",
          gains: { ...DEFAULT_EQ_GAINS },
        }),
    }),
    {
      name: "inori-eq",
      partialize: (s) => ({ enabled: s.enabled, preset: s.preset, gains: s.gains }),
    }
  )
);
