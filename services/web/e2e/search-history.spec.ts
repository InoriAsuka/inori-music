/**
 * search-history.spec.ts — search history dropdown behavior (v5.1.0).
 *
 * Covers the localStorage-backed recent-searches dropdown on /search:
 *   type a query → history recorded → clearing the field and refocusing
 *   shows the dropdown → selecting an entry fills the query → clearing
 *   history empties the dropdown.
 *
 * Does not depend on catalog data being seeded — history persists purely
 * client-side regardless of whether the query matches anything.
 *
 * Prerequisites (same as smoke.spec.ts):
 *   - Dev server running on http://localhost:3000 (or E2E_BASE_URL)
 *   - E2E_USERNAME / E2E_PASSWORD env vars set to a valid account
 *
 * Run: npx playwright test search-history.spec.ts
 */

import { test, expect, type Page } from "@playwright/test";

const USERNAME = process.env.E2E_USERNAME ?? "ci_viewer";
const PASSWORD = process.env.E2E_PASSWORD ?? "ci-password-123";

async function login(page: Page) {
  await page.goto("/login");
  await page.locator("#username").fill(USERNAME);
  await page.locator("#password").fill(PASSWORD);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 10_000 });
}

test("search history: recorded on search, shown on refocus, selectable, clearable", async ({ page }) => {
  await login(page);
  await page.goto("/search");

  const input = page.getByPlaceholder(/search/i);
  await expect(input).toBeVisible({ timeout: 6_000 });

  // Typing a query and letting the debounce fire records it to history.
  await input.fill("neon shrine");
  await page.waitForTimeout(500); // debounce is 300ms

  // Clear the field and refocus — the recent-searches dropdown should appear.
  await input.fill("");
  await input.blur();
  await input.click();

  await expect(page.getByText(/recent searches/i)).toBeVisible({ timeout: 4_000 });
  const entry = page.getByText("neon shrine", { exact: true });
  await expect(entry).toBeVisible();

  // Selecting the entry fills the input.
  await entry.click();
  await expect(input).toHaveValue("neon shrine");

  // Clear the field and refocus, then clear all history via the trash icon.
  await input.fill("");
  await input.blur();
  await input.click();
  await expect(page.getByText(/recent searches/i)).toBeVisible({ timeout: 4_000 });
  await page.getByTitle(/clear history/i).click();
  await expect(page.getByText(/recent searches/i)).toHaveCount(0);
});
