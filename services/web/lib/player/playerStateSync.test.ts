import { describe, expect, it } from "vitest";
import { decideResume, reportingSignature, formatResumePosition } from "./playerStateSync";
import type { LocalPlayerSnapshot, RemotePlayerState } from "./player-state";

function t(id: string) {
  return { id };
}

function snap(opts: Partial<LocalPlayerSnapshot> = {}): LocalPlayerSnapshot {
  return {
    queue: [],
    currentIndex: 0,
    positionSeconds: 0,
    repeat: "off",
    shuffle: false,
    volume: 1,
    speed: 1,
    status: "paused",
    ...opts,
  };
}

function remote(opts: Partial<RemotePlayerState> = {}): RemotePlayerState {
  return {
    queue: [],
    currentIndex: 0,
    positionSeconds: 0,
    repeat: "off",
    shuffle: false,
    volume: 1,
    speed: 1,
    status: "paused",
    updatedAt: new Date().toISOString(),
    ...opts,
  };
}

describe("reportingSignature", () => {
  it("differs when track IDs change", () => {
    const a = snap({ queue: [t("a"), t("b")], currentIndex: 0 });
    const b = snap({ queue: [t("a"), t("c")], currentIndex: 0 });
    expect(reportingSignature(a)).not.toBe(reportingSignature(b));
  });

  it("is the same when only position differs", () => {
    const a = snap({ queue: [t("x")], positionSeconds: 10 });
    const b = snap({ queue: [t("x")], positionSeconds: 999 });
    expect(reportingSignature(a)).toBe(reportingSignature(b));
  });

  it("is stable across identical snapshots", () => {
    const s = snap({ queue: [t("a"), t("b")], repeat: "all" });
    expect(reportingSignature(s)).toBe(reportingSignature(s));
  });
});

describe("decideResume", () => {
  it("returns NO_RESUME when remote is null", () => {
    expect(decideResume(null, snap({ queue: [t("a")] }))).toEqual({
      prompt: false,
      remoteTrackId: null,
      positionSeconds: 0,
    });
  });

  it("returns NO_RESUME when remote queue is empty", () => {
    expect(decideResume(remote(), snap({ queue: [t("a")] }))).toEqual({
      prompt: false,
      remoteTrackId: null,
      positionSeconds: 0,
    });
  });

  it("prompts when the remote track differs from local", () => {
    const local = snap({ queue: [t("a"), t("b")], currentIndex: 0, positionSeconds: 0 });
    // Remote: playing track "c" at index 1
    const r = remote({ queue: ["a", "c"], currentIndex: 1, positionSeconds: 30 });
    const decision = decideResume(r, local);
    expect(decision.prompt).toBe(true);
    expect(decision.remoteTrackId).toBe("c");
    expect(decision.positionSeconds).toBe(30);
  });

  it("does NOT prompt for identical track and position (same device)", () => {
    const local = snap({ queue: [t("a")], currentIndex: 0, positionSeconds: 42 });
    const r = remote({ queue: ["a"], currentIndex: 0, positionSeconds: 42 });
    expect(decideResume(r, local).prompt).toBe(false);
  });

  it("does NOT prompt for a tiny position drift within epsilon", () => {
    const local = snap({ queue: [t("a")], positionSeconds: 50 });
    const r = remote({ queue: ["a"], positionSeconds: 52 }); // within 5s default epsilon
    expect(decideResume(r, local).prompt).toBe(false);
  });

  it("prompts when position drift exceeds epsilon", () => {
    const local = snap({ queue: [t("a")], positionSeconds: 50 });
    const r = remote({ queue: ["a"], positionSeconds: 56 }); // 6s > 5s epsilon
    const decision = decideResume(r, local);
    expect(decision.prompt).toBe(true);
    expect(decision.positionSeconds).toBe(56);
  });

  it("uses the custom epsilon when provided", () => {
    const local = snap({ queue: [t("a")], positionSeconds: 50 });
    const r = remote({ queue: ["a"], positionSeconds: 52 }); // 2s < 3s epsilon
    expect(decideResume(r, local, 3).prompt).toBe(false);
    expect(decideResume(r, local, 1).prompt).toBe(true);
  });
});

describe("formatResumePosition", () => {
  it("formats 0 seconds as 0:00", () => {
    expect(formatResumePosition(0)).toBe("0:00");
  });

  it("formats 65 seconds as 1:05", () => {
    expect(formatResumePosition(65)).toBe("1:05");
  });

  it("clamps negative values to 0", () => {
    expect(formatResumePosition(-10)).toBe("0:00");
  });

  it("formats a larger duration correctly", () => {
    expect(formatResumePosition(3725)).toBe("62:05");
  });
});
