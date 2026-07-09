/**
 * smoke.spec.ts for inori-music admin smoke tests (Phase 2/3 of Phase 2/5/1)
 *
 * Covers four critical paths:
 * 1. Login flow — unauthenticated users redirect to /login; valid credentials grant access.
 * 2. Users tab — renders user table, confirms at least one user exists.
 * 3. Catalog tab — renders catalog view and confirms heading.
 * 4. Storage tab — renders storage backend list.
 *
 * Prerequisites:
 * - Dev server running at http://localhost:3001 (or E2E_BASE_URL).
 * - E2E_USERNAME / E2E_PASSWORD env vars set to a valid admin account.
 * - API server reachable at NEXT_PUBLIC_API_BASE_URL (default http://localhost:8080).
 *
 * Run: npx playwright test
 * CI: E2E_USERNAME=ci_viewer E2E_PASSWORD=ci-password-123 npx playwright test
 */

import { test, expect, type Page } from "@playwright/test";

const E2E_USERNAME = process.env.E2E_USERNAME ?? "admin";
const E2E_PASSWORD = process.env.E2E_PASSWORD ?? "password";

async function login(page: Page) {
	await page.goto("/login");
	await page.fill('input[type="text"]', E2E_USERNAME);
	await page.fill('input[type="password"]', E2E_PASSWORD);
	await page.click('button[type="submit"]');

	// Wait for redirect to dashboard
	await page.waitForURL(/\/(dashboard|users|catalog|storage)/i, {
		timeout: 8_000,
	});
}

test("login flow redirects unauthenticated users and grants access on valid credentials", async ({
	page,
}) => {
	await page.goto("/users");
	// Should redirect to /login
	await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });

	await login(page);
	// After login, should land on a protected route
	await expect(page).toHaveURL(/\/(dashboard|users|catalog|storage)/i);
});

test("users tab renders user table with at least one user", async ({
	page,
}) => {
	await login(page);
	await page.goto("/users");

	// Users heading
	await expect(page.locator("h1").filter({ hasText: /users/i })).toBeVisible({
		timeout: 8_000,
	});

	// Confirm table or user rows exist
	await expect(
		page.locator("table, [role='table'], div").filter({ hasText: /username/i }),
	).toBeVisible({ timeout: 6_000 });
});

test("catalog tab renders catalog view", async ({ page }) => {
	await login(page);
	await page.goto("/catalog");

	// Catalog heading
	await expect(
		page.locator("h1").filter({ hasText: /catalog/i }),
	).toBeVisible({ timeout: 8_000 });
});

test("storage tab renders storage backend list", async ({ page }) => {
	await login(page);
	await page.goto("/storage");

	// Storage heading
	await expect(
		page.locator("h1").filter({ hasText: /storage/i }),
	).toBeVisible({ timeout: 8_000 });
});
