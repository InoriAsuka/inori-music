/**
 * Shared equalizer preset definitions for Web and Flutter.
 *
 * 10-band ISO standard frequencies with ±6 dB range and 0.5 dB steps.
 */

export type EqBandId =
  | "band-31"
  | "band-62"
  | "band-125"
  | "band-250"
  | "band-500"
  | "band-1k"
  | "band-2k"
  | "band-4k"
  | "band-8k"
  | "band-16k";

export interface EqBand {
  id: EqBandId;
  label: string;
  frequency: number;
  minDb: number;
  maxDb: number;
  stepDb: number;
  defaultDb: number;
}

export const EQ_BANDS: readonly EqBand[] = [
  { id: "band-31", label: "31", frequency: 31, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-62", label: "62", frequency: 62, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-125", label: "125", frequency: 125, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-250", label: "250", frequency: 250, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-500", label: "500", frequency: 500, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-1k", label: "1k", frequency: 1000, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-2k", label: "2k", frequency: 2000, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-4k", label: "4k", frequency: 4000, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-8k", label: "8k", frequency: 8000, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
  { id: "band-16k", label: "16k", frequency: 16000, minDb: -6, maxDb: 6, stepDb: 0.5, defaultDb: 0 },
];

export const DEFAULT_EQ_GAINS = EQ_BANDS.reduce<Record<EqBandId, number>>((acc, band) => {
  acc[band.id] = band.defaultDb;
  return acc;
}, {} as Record<EqBandId, number>);

export type EqPresetName = "flat" | "bassBoost" | "vocal" | "electronic";

/**
 * Preset gains, one entry per band in EQ_BANDS order.
 *
 * These values are the authoritative shared definition and must stay
 * byte-for-byte identical to `eqPresets` in
 * `services/mobile/lib/src/audio/eq_settings.dart` — the Flutter client
 * shipped them first, so both clients sound the same on the same preset.
 */
export const EQ_PRESETS: Record<EqPresetName, Record<EqBandId, number>> = {
  flat: { ...DEFAULT_EQ_GAINS },
  bassBoost: {
    "band-31": 6, "band-62": 5, "band-125": 4, "band-250": 2,
    "band-500": 0, "band-1k": 0, "band-2k": 0, "band-4k": 0,
    "band-8k": 0, "band-16k": 0,
  },
  vocal: {
    "band-31": 0, "band-62": 0, "band-125": 0, "band-250": 2,
    "band-500": 4, "band-1k": 4, "band-2k": 3, "band-4k": 2,
    "band-8k": 0, "band-16k": 0,
  },
  electronic: {
    "band-31": 4, "band-62": 3, "band-125": 0, "band-250": -2,
    "band-500": 0, "band-1k": 2, "band-2k": 3, "band-4k": 3,
    "band-8k": 4, "band-16k": 4,
  },
};
