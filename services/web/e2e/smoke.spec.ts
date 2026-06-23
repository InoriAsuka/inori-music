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
 *   - E2E_USERNAME / E2E_PASSWORD env vars set to a valid viewer account
 *   - API server reachable at NEXT_PUBLIC_API_BASE_URL
 *
 * Run: npx playwright test
 */

import { test, expect } from "@playwright/test";

const USERNAME = process.env.E2E_USERNAME ?? "viewer@example.com";
const PASSWORD = process.env.E2E_PASSWORD ?? "changeme";

// ─── 1. Login ────────────────────────────────────────────────────────────────

test("unauthenticated visit redirects to /login", async ({ page }) => {
  await page.goto("/");
  await expect(page).toHaveURL(/\/login/);
});

test("login with valid credentials lands on home page", async ({ page }) => {
  await page.goto("/login");

  await page.getByLabel(/username|email/i).fill(USERNAME);
  await page.getByLabel(/password/i).fill(PASSWORD);
  await page.getByRole("button", { name: /sign in|log in/i }).click();

  // After login the app redirects away from /login
  await expect(page).not.toHaveURL(/\/login/, { timeout: 8_000 });

  // Home page has a heading or a recognisable landmark
  await expect(
    page.getByRole("heading", { name: /home/i }).or(
      page.locator("h1").filter({ hasText: /home/i })
    )
  ).toBeVisible({ timeout: 8_000 });
});

// ─── 2. Search ───────────────────────────────────────────────────────────────

test("search page renders and accepts input", async ({ page }) => {
  // Login first
  await page.goto("/login");
  await page.getByLabel(/username|email/i).fill(USERNAME);
  await page.getByLabel(/password/i).fill(PASSWORD);
  await page.getByRole("button", { name: /sign in|log in/i }).click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 8_000 });

  await page.goto("/search");

  const input = page.getByRole("searchbox").or(
    page.getByPlaceholder(/search/i)
  );
  await expect(input).toBeVisible({ timeout: 6_000 });
  await input.fill("test");
  // Input accepted — value is what we typed
  await expect(input).toHaveValue("test");
});

// ─── 3. Player bar visible on /tracks ────────────────────────────────────────

test("tracks page renders and player bar is present in layout", async ({
  page,
}) => {
  // Login first
  await page.goto("/login");
  await page.getByLabel(/username|email/i).fill(USERNAME);
  await page.getByLabel(/password/i).fill(PASSWORD);
  await page.getByRole("button", { name: /sign in|log in/i }).click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 8_000 });

  await page.goto("/tracks");

  // The tracks heading must be visible
  await expect(
    page.getByRole("heading", { name: /tracks/i }).or(
      page.locator("h1").filter({ hasText: /tracks/i })
    )
  ).toBeVisible({ timeout: 8_000 });

  // The persistent player bar renders at the bottom (even before a track is
  // selected it shows the "No track playing" placeholder)
  const playerBar = page
    .locator('[class*="PlayerBar"], [data-testid="player-bar"]')
    .or(page.getByText(/no track playing/i));
  await expect(playerBar.first()).toBeVisible({ timeout: 6_000 });
});
