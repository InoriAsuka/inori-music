import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

/**
 * Unit tests for crossfade.ts — localStorage-backed crossfade toggle.
 * Same SSR-safe shape as token.ts; see lib/api/token.test.ts for the
 * fake-localStorage pattern reused here.
 */

function installFakeLocalStorage() {
  const store = new Map<string, string>();
  const fakeLocalStorage = {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => {
      store.set(key, value);
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => store.clear(),
  };
  vi.stubGlobal("window", {});
  vi.stubGlobal("localStorage", fakeLocalStorage);
  return fakeLocalStorage;
}

describe("crossfade.ts — browser environment", () => {
  let fakeLocalStorage: ReturnType<typeof installFakeLocalStorage>;

  beforeEach(() => {
    fakeLocalStorage = installFakeLocalStorage();
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("defaults to disabled when nothing is stored", async () => {
    const { isCrossfadeEnabled } = await import("./crossfade");
    expect(isCrossfadeEnabled()).toBe(false);
  });

  it("persists true across reads", async () => {
    const { isCrossfadeEnabled, setCrossfadeEnabled } = await import("./crossfade");
    setCrossfadeEnabled(true);
    expect(isCrossfadeEnabled()).toBe(true);
  });

  it("persists false explicitly, overwriting a prior true", async () => {
    const { isCrossfadeEnabled, setCrossfadeEnabled } = await import("./crossfade");
    setCrossfadeEnabled(true);
    setCrossfadeEnabled(false);
    expect(isCrossfadeEnabled()).toBe(false);
  });

  it("treats any stored value other than the literal string 'true' as disabled", async () => {
    fakeLocalStorage.setItem("inori.audio.crossfadeEnabled", "1");
    const { isCrossfadeEnabled } = await import("./crossfade");
    expect(isCrossfadeEnabled()).toBe(false);
  });

  it("stores the flag under the app-specific key", async () => {
    const { setCrossfadeEnabled } = await import("./crossfade");
    setCrossfadeEnabled(true);
    expect(fakeLocalStorage.getItem("inori.audio.crossfadeEnabled")).toBe("true");
  });

  it("does not throw when localStorage access throws (private browsing / quota)", async () => {
    vi.stubGlobal("localStorage", {
      getItem: () => {
        throw new Error("blocked");
      },
      setItem: () => {
        throw new Error("blocked");
      },
    });
    const { isCrossfadeEnabled, setCrossfadeEnabled } = await import("./crossfade");
    expect(isCrossfadeEnabled()).toBe(false);
    expect(() => setCrossfadeEnabled(true)).not.toThrow();
  });
});

describe("crossfade.ts — server (no window) environment", () => {
  beforeEach(() => {
    vi.stubGlobal("window", undefined);
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("isCrossfadeEnabled returns false without throwing when window is undefined", async () => {
    const { isCrossfadeEnabled } = await import("./crossfade");
    expect(isCrossfadeEnabled()).toBe(false);
  });

  it("setCrossfadeEnabled is a no-op without throwing when window is undefined", async () => {
    const { setCrossfadeEnabled } = await import("./crossfade");
    expect(() => setCrossfadeEnabled(true)).not.toThrow();
  });
});

describe("CROSSFADE_SECONDS", () => {
  it("is a short, fixed ramp duration in seconds", async () => {
    const { CROSSFADE_SECONDS } = await import("./crossfade");
    expect(CROSSFADE_SECONDS).toBe(2);
  });
});
