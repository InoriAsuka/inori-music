import { describe, expect, it } from "vitest";
import { computeReplayGain } from "./replayGain";

/**
 * ReplayGain calculation — mirrors the Flutter client's formula
 * (10^(db/20), clamped to [0.1, 2.0]) so cross-platform loudness matches.
 */
describe("computeReplayGain", () => {
  it("returns unity gain (1.0) for null", () => {
    expect(computeReplayGain(null)).toBe(1.0);
  });

  it("returns unity gain (1.0) for undefined", () => {
    expect(computeReplayGain(undefined)).toBe(1.0);
  });

  it("returns unity gain (1.0) for NaN", () => {
    expect(computeReplayGain(Number.NaN)).toBe(1.0);
  });

  it("returns unity gain (1.0) for 0 dB", () => {
    expect(computeReplayGain(0)).toBeCloseTo(1.0, 10);
  });

  it("computes 10^(db/20) for a typical negative dB value", () => {
    // -6 dB -> 10^(-6/20) ≈ 0.50119
    expect(computeReplayGain(-6)).toBeCloseTo(0.501187, 5);
  });

  it("computes 10^(db/20) for a typical positive dB value", () => {
    // +3 dB -> 10^(3/20) ≈ 1.41254
    expect(computeReplayGain(3)).toBeCloseTo(1.412538, 5);
  });

  it("clamps extreme negative dB to the 0.1 floor", () => {
    // -40 dB would compute to 0.01, far below the floor.
    expect(computeReplayGain(-40)).toBe(0.1);
  });

  it("clamps extreme positive dB to the 2.0 ceiling", () => {
    // +40 dB would compute to 100, far above the ceiling.
    expect(computeReplayGain(40)).toBe(2.0);
  });

  it("clamps exactly at the boundary values", () => {
    // 20*log10(0.1) = -20 dB, 20*log10(2.0) ≈ 6.0206 dB
    expect(computeReplayGain(-20)).toBeCloseTo(0.1, 10);
    expect(computeReplayGain(20 * Math.log10(2))).toBeCloseTo(2.0, 5);
  });
});
