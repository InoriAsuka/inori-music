import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright config — smoke tests for inori-music web player.
 *
 * Targets: http://localhost:3000 (dev server must be running).
 * Run: npx playwright test
 */
export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  retries: 1,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
