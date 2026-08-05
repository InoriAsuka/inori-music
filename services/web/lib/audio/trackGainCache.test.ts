import { describe, expect, it, vi, beforeEach } from "vitest";
import { resolveReplayGainDb, __resetReplayGainCache } from "./trackGainCache";
import type { authedApi } from "@/lib/api/client";

type Fetcher = ReturnType<typeof authedApi>;

/**
 * The cache is module-level, so tests must reset it between runs. The
 * export is intentionally ugly (`__resetReplayGainCache`) to discourage
 * importing it outside tests; callers that need isolation should call
 * it at the start of each test.
 */

beforeEach(() => {
  __resetReplayGainCache();
});

function makeApi(gain: number | null) {
  const api = {
    GET: vi.fn(),
  };
  api.GET.mockResolvedValue({
    data: { replayGainDb: gain },
    error: null,
  });
  return api as unknown as Fetcher;
}

describe("resolveReplayGainDb", () => {
  it("returns the cached value on a repeated call for the same track id", async () => {
    const api = makeApi(3);
    const first = await resolveReplayGainDb(api, "t1");
    const second = await resolveReplayGainDb(api, "t1");
    expect(first).toBe(3);
    expect(second).toBe(3);
    expect(api.GET).toHaveBeenCalledTimes(1);
  });

  it("fetches once per distinct track id", async () => {
    const api = makeApi(2);
    const a = await resolveReplayGainDb(api, "t1");
    const b = await resolveReplayGainDb(api, "t2");
    expect(a).toBe(2);
    expect(b).toBe(2);
    expect(api.GET).toHaveBeenCalledTimes(2);
  });

  it("caches null when the server returns no replayGainDb", async () => {
    const api = makeApi(null);
    const result = await resolveReplayGainDb(api, "t1");
    expect(result).toBeNull();
    expect(api.GET).toHaveBeenCalledTimes(1);
  });

  it("does not cache transient fetch failures", async () => {
    const api = { GET: vi.fn() };
    api.GET.mockRejectedValue(new Error("network"));

    const first = await resolveReplayGainDb(api as any, "t1");
    expect(first).toBeNull();

    api.GET.mockResolvedValue({ data: { replayGainDb: 1 }, error: null });
    const second = await resolveReplayGainDb(api as any, "t1");
    expect(second).toBe(1);
    expect(api.GET).toHaveBeenCalledTimes(2);
  });

  it("returns null for an empty track id without fetching", async () => {
    const api = makeApi(0);
    const result = await resolveReplayGainDb(api, "");
    expect(result).toBeNull();
    expect(api.GET).not.toHaveBeenCalled();
  });
});
