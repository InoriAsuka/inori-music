#!/usr/bin/env node
/**
 * scripts/verify-production-build-requires-api-base-url.mjs
 *
 * Guard for the 2026-08-11 incident: `next build` used to silently bake
 * "http://localhost:8080" into `.next/routes-manifest.json`'s rewrite
 * destination whenever NEXT_PUBLIC_API_BASE_URL wasn't wired through (the
 * Dockerfile had no `ARG NEXT_PUBLIC_API_BASE_URL`, so docker.yml's
 * --build-arg was silently discarded by BuildKit). A standalone `next
 * start` reads that manifest verbatim and never re-evaluates
 * next.config.ts's rewrites() at request time, so no runtime env var could
 * fix it after the fact — see requirement.md v5.35.0 for the full chain.
 * Admin is a separate Next.js app/build from web; see services/web's copy
 * of this script for the same guard on that app.
 *
 * This script runs the REAL command Docker's builder stage runs — `npm run
 * build` — twice, and inspects the REAL build artifact on disk. It's slow
 * (two full production builds), so it's not wired into a fast unit-test
 * loop (admin has no Vitest/unit-test setup at all — only Playwright e2e);
 * run it explicitly or wire it into CI as its own step.
 *
 *   Scenario A — NEXT_PUBLIC_API_BASE_URL unset (reproduces the incident):
 *     the build MUST fail loudly. If it instead succeeds with "localhost"
 *     baked into the manifest, that IS the regression this guards against.
 *
 *   Scenario B — NEXT_PUBLIC_API_BASE_URL set to a real, non-localhost
 *     value: the build MUST succeed, and the manifest's rewrite destination
 *     must carry that value through, and never contain "localhost".
 *
 * Usage: node scripts/verify-production-build-requires-api-base-url.mjs
 * Exit code 0 = both guards held. Non-zero = a regression was detected
 * (details printed to stdout/stderr).
 */
import { spawnSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const APP_ROOT = path.resolve(fileURLToPath(import.meta.url), "..", "..");
const MANIFEST_PATH = path.join(APP_ROOT, ".next", "routes-manifest.json");
// Matches the real production topology (docker-compose.prod.yml, docker.yml's
// --build-arg) — deliberately NOT localhost, so any leak of the old fallback
// is unambiguous.
const REAL_API_BASE_URL = "http://api:8080";

let failures = 0;

function fail(message) {
  console.error(`FAIL: ${message}`);
  failures += 1;
}

function pass(message) {
  console.log(`PASS: ${message}`);
}

function readManifestRewrites() {
  if (!existsSync(MANIFEST_PATH)) return null;
  const manifest = JSON.parse(readFileSync(MANIFEST_PATH, "utf8"));
  return JSON.stringify(manifest.rewrites ?? {});
}

function runBuild(env) {
  const result = spawnSync("npm", ["run", "build"], { cwd: APP_ROOT, env, encoding: "utf8" });
  const output = `${result.stdout ?? ""}${result.stderr ?? ""}`;
  console.log(`exit code: ${result.status}`);
  return { status: result.status, output };
}

// ── Scenario A: var unset — must fail loudly, never bake in localhost ──────
console.log(`\n--- Scenario A: NEXT_PUBLIC_API_BASE_URL unset (incident repro) [${APP_ROOT}] ---`);
{
  const env = { ...process.env };
  env.NEXT_PUBLIC_API_BASE_URL = undefined; // omitted entirely from the child's env (verified empirically, not just assumed)
  const { status, output } = runBuild(env);

  if (status === 0) {
    const destinations = readManifestRewrites();
    if (destinations?.includes("localhost")) {
      fail(
        `production build succeeded WITHOUT NEXT_PUBLIC_API_BASE_URL and silently baked "localhost" into .next/routes-manifest.json — this is the exact 2026-08-11 regression. Destinations: ${destinations}`
      );
    } else {
      fail(
        `production build succeeded WITHOUT NEXT_PUBLIC_API_BASE_URL set at all (expected it to throw instead). Manifest destinations: ${destinations}`
      );
    }
  } else if (!output.includes("NEXT_PUBLIC_API_BASE_URL")) {
    fail(
      `production build failed as expected, but NOT for the expected reason (output doesn't mention NEXT_PUBLIC_API_BASE_URL) — investigate before trusting this guard. Output tail:\n${output.slice(-2000)}`
    );
  } else {
    pass("production build without NEXT_PUBLIC_API_BASE_URL failed loudly, naming the missing var.");
  }
}

// ── Scenario B: var set to a real value — must succeed, never localhost ────
console.log(`\n--- Scenario B: NEXT_PUBLIC_API_BASE_URL=${REAL_API_BASE_URL} (correct config) ---`);
{
  const env = { ...process.env, NEXT_PUBLIC_API_BASE_URL: REAL_API_BASE_URL };
  const { status, output } = runBuild(env);

  if (status !== 0) {
    fail(
      `production build FAILED with NEXT_PUBLIC_API_BASE_URL=${REAL_API_BASE_URL} set — a correctly configured build must succeed. Output tail:\n${output.slice(-2000)}`
    );
  } else {
    const destinations = readManifestRewrites();
    if (!destinations || !destinations.includes(REAL_API_BASE_URL)) {
      fail(`manifest rewrite destination does not carry through ${REAL_API_BASE_URL}. Got: ${destinations}`);
    } else if (destinations.includes("localhost")) {
      fail(`manifest still contains "localhost" even with the var correctly set. Got: ${destinations}`);
    } else {
      pass(`manifest correctly carries ${REAL_API_BASE_URL} through with no "localhost". Got: ${destinations}`);
    }
  }
}

console.log(`\n${failures === 0 ? "ALL GUARDS HELD" : `${failures} GUARD(S) FAILED`}`);
process.exit(failures === 0 ? 0 : 1);
