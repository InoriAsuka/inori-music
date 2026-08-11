import type { NextConfig } from "next";
import { PHASE_PRODUCTION_BUILD } from "next/constants";

const DEV_FALLBACK_API_BASE_URL = "http://localhost:8080";

/**
 * next.config.ts — exported as a function of Next.js's build "phase" rather
 * than a static object, so we can tell "building right now" apart from
 * "serving an already-built app".
 *
 * Why phase and not NODE_ENV: `next start` (serving a finished build) *also*
 * defaults NODE_ENV to "production" — see node_modules/next/dist/bin/next's
 * `defaultEnv = commandName === "dev" ? "development" : "production"`. So a
 * `NODE_ENV === "production"` check can't distinguish the two; gating on it
 * would risk crashing an already-running production server (e.g. if a
 * runtime env var were ever briefly unset) instead of only ever failing a
 * build. `phase`, which Next.js itself passes in, can: it is
 * PHASE_PRODUCTION_BUILD only for `next build`; `next start` gets
 * PHASE_PRODUCTION_SERVER instead (see node_modules/next/dist/build/index.js
 * vs. node_modules/next/dist/server/next.js). Gating on PHASE_PRODUCTION_BUILD
 * means a missing env var can only ever fail a *build*, loudly, at the
 * moment the misconfiguration is made.
 *
 * Incident (2026-08-11): the old `?? "http://localhost:8080"` fallback
 * silently baked "http://localhost:8080" into the production build's
 * `.next/routes-manifest.json` when NEXT_PUBLIC_API_BASE_URL wasn't wired
 * through Docker's build-args (the Dockerfile had no matching `ARG`). A
 * standalone `next start` server reads that manifest verbatim and never
 * re-runs `rewrites()`, so no runtime env var could fix it after the fact —
 * see requirement.md v5.35.0 for the full chain. (Same bug class as web's
 * next.config.ts; admin is a separate Next.js app with its own build.)
 */
const nextConfig = (phase: string): NextConfig => {
  const apiBase = process.env.NEXT_PUBLIC_API_BASE_URL;

  if (!apiBase && phase === PHASE_PRODUCTION_BUILD) {
    throw new Error(
      `NEXT_PUBLIC_API_BASE_URL is not set. A production build refuses to silently fall back to ${DEV_FALLBACK_API_BASE_URL} (see the 2026-08-11 incident in requirement.md) — set it explicitly (Docker build-arg, CI env, or .env.production).`
    );
  }

  const resolvedApiBase = apiBase ?? DEV_FALLBACK_API_BASE_URL;

  return {
    output: "standalone",
    basePath: "/admin",
    transpilePackages: ["@inori/ui"],
    async rewrites() {
      return [
        {
          source: "/api/v1/:path*",
          destination: `${resolvedApiBase}/api/v1/:path*`,
        },
      ];
    },
    eslint: { ignoreDuringBuilds: true },
  };
};

export default nextConfig;
