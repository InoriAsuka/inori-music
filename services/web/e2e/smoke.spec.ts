/**
 * smoke.spec.ts — inori-music web player smoke tests (Phase 213 / Phase 251)
 *
 * Covers three critical paths:
 *   1. Login flow — unauthenticated users redirect to /login; valid credentials
 *      grant access to the home page.
 *   2. Search — the search page renders and accepts input.
 *   3. Playback stub — the player bar is present after navigating to /tracks.
 *
 * Prerequisites:
 *   - Dev server running on http://localhost:3000 (or E2E_BASE_URL)
 *   - E2E_USERNAME / E2E_PASSWORD env vars set to a valid account
 *   - API server reachable at NEXT_PUBLIC_API_BASE_URL (default http://localhost:8080)
 *
 * Run: npx playwright test
 * CI:  E2E_USERNAME=ci_viewer E2E_PASSWORD=ci-password-123 npx playwright test
 */

import { test, expect, type Page } from "@playwright/test";

const USERNAME = process.env.E2E_USERNAME ?? "ci_viewer";
const PASSWORD = process.env.E2E_PASSWORD ?? "ci-password-123";

/** Shared login helper — fills the form and waits for redirect. */
async function login(page: Page) {
  await page.goto("/login");
  await page.locator("#username").fill(USERNAME);
  await page.locator("#password").fill(PASSWORD);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 10_000 });
}

// ─── 1. Auth ─────────────────────────────────────────────────────────────────

test("unauthenticated visit redirects to /login", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/login/);
});

test("login with valid credentials lands on home page", async ({ page }) => {
  await login(page);
  // Home page renders the "Home" heading
  await expect(page.locator("h1").filter({ hasText: /home/i })).toBeVisible({
    timeout: 8_000,
  });
});

// ─── 2. Search ───────────────────────────────────────────────────────────────

test("search page renders and accepts input", async ({ page }) => {
  await login(page);
  await page.goto("/search");

  // The search input has placeholder "Search tracks, artists…"
  const input = page.getByPlaceholder(/search/i);
  await expect(input).toBeVisible({ timeout: 6_000 });
  await input.fill("test");
  await expect(input).toHaveValue("test");
});

// ─── 3. Player bar on /tracks ────────────────────────────────────────────────

test("tracks page renders and player bar shows idle state", async ({ page }) => {
  await login(page);
  await page.goto("/tracks");

  // Tracks heading
  await expect(page.locator("h1").filter({ hasText: /tracks/i })).toBeVisible({
    timeout: 8_000,
  });

  // Player bar shows "No track playing" before any track is selected
  await expect(page.getByText(/no track playing/i)).toBeVisible({
    timeout: 6_000,
  });
});
