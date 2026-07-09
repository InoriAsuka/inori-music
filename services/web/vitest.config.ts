import { defineConfig } from "vitest/config";

/**
 * Vitest config — unit tests for store/ and lib/ (pure logic, no DOM needed).
 *
 * Playwright (e2e/) covers UI/browser behavior; Vitest covers state machines
 * and utility functions in isolation. Run: npx vitest run
 */
export default defineConfig({
  test: {
    environment: "node",
    include: ["**/*.test.ts"],
    exclude: ["node_modules/**", "e2e/**", ".next/**"],
  },
});
