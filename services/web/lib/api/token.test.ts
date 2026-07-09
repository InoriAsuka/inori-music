import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

/**
 * Unit tests for token.ts — localStorage-backed auth token persistence.
 *
 * token.ts guards every localStorage access behind `typeof window !== "undefined"`
 * so it's SSR-safe. We simulate both environments: a minimal in-memory
 * localStorage stand-in for the "browser" case, and a real absence of
 * `window` for the "server" case — without pulling in jsdom just for this.
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

describe("token.ts — browser environment", () => {
  beforeEach(() => {
    installFakeLocalStorage();
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("storeToken persists the token under the app-specific key", async () => {
    const { storeToken, getStoredToken } = await import("./token");
    storeToken("abc123");
    expect(getStoredToken()).toBe("abc123");
  });

  it("getStoredToken returns null when nothing is stored", async () => {
    const { getStoredToken } = await import("./token");
    expect(getStoredToken()).toBeNull();
  });

  it("removeToken clears both the token and user keys", async () => {
    const { storeToken, removeToken, getStoredToken } = await import("./token");
    storeToken("abc123");
    localStorage.setItem("inori_auth_user", "some-user-json");
    removeToken();
    expect(getStoredToken()).toBeNull();
    expect(localStorage.getItem("inori_auth_user")).toBeNull();
  });

  it("overwriting storeToken replaces the previous value", async () => {
    const { storeToken, getStoredToken } = await import("./token");
    storeToken("first");
    storeToken("second");
    expect(getStoredToken()).toBe("second");
  });
});

describe("token.ts — server (no window) environment", () => {
  beforeEach(() => {
    vi.stubGlobal("window", undefined);
    vi.resetModules();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("getStoredToken returns null without throwing when window is undefined", async () => {
    const { getStoredToken } = await import("./token");
    expect(getStoredToken()).toBeNull();
  });

  it("storeToken and removeToken are no-ops without throwing when window is undefined", async () => {
    const { storeToken, removeToken } = await import("./token");
    expect(() => storeToken("abc123")).not.toThrow();
    expect(() => removeToken()).not.toThrow();
  });
});
