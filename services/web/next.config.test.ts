import { PHASE_DEVELOPMENT_SERVER, PHASE_PRODUCTION_BUILD } from "next/constants";
import { afterEach, describe, expect, it, vi } from "vitest";
import nextConfig from "./next.config";

/**
 * Guard test for the 2026-08-11 incident: a production build with
 * NEXT_PUBLIC_API_BASE_URL unset used to silently fall back to
 * "http://localhost:8080" and bake that into .next/routes-manifest.json —
 * permanently, since a standalone `next start` never re-reads env vars for
 * rewrites (see next.config.ts's module doc comment for the full chain).
 *
 * This file tests the function that *produces* the manifest's rewrite
 * destination directly (fast, no real build). A slower, end-to-end guard
 * that actually runs `next build` and inspects the real
 * .next/routes-manifest.json artifact lives at
 * scripts/verify-production-build-requires-api-base-url.mjs — see
 * .plan/ for why the check is split across both layers.
 *
 * Uses vi.stubEnv/unstubAllEnvs (not `delete process.env.X` / `process.env.X
 * = undefined`) to unset the var: process.env is a special auto-stringifying
 * object, so a plain assignment of `undefined` becomes the *string*
 * "undefined" instead of actually unsetting the key — vi.stubEnv(name,
 * undefined) is vitest's own env-stubbing utility and internally does a real
 * `delete`, verified empirically before relying on it here.
 */
describe("next.config.ts — NEXT_PUBLIC_API_BASE_URL contract", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("throws during a production build (PHASE_PRODUCTION_BUILD) when the var is unset", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", undefined);
    expect(() => nextConfig(PHASE_PRODUCTION_BUILD)).toThrow(/NEXT_PUBLIC_API_BASE_URL/);
  });

  it("throws during a production build when the var is set but empty", () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "");
    expect(() => nextConfig(PHASE_PRODUCTION_BUILD)).toThrow(/NEXT_PUBLIC_API_BASE_URL/);
  });

  it("does NOT throw during a production build when the var is set, and never falls back to localhost", async () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", "http://api:8080");
    const config = nextConfig(PHASE_PRODUCTION_BUILD);
    const rewrites = await config.rewrites?.();
    const destinations = JSON.stringify(rewrites);
    expect(destinations).toContain("http://api:8080/api/v1/:path*");
    expect(destinations).not.toContain("localhost");
  });

  it("keeps the localhost fallback in the dev server phase when the var is unset", async () => {
    vi.stubEnv("NEXT_PUBLIC_API_BASE_URL", undefined);
    expect(() => nextConfig(PHASE_DEVELOPMENT_SERVER)).not.toThrow();
    const config = nextConfig(PHASE_DEVELOPMENT_SERVER);
    const rewrites = await config.rewrites?.();
    expect(JSON.stringify(rewrites)).toContain("http://localhost:8080/api/v1/:path*");
  });
});
