/**
 * main-flow.spec.ts — inori-music web player end-to-end main flow.
 *
 * Exercises the full viewer journey in one continuous session:
 *   登录 → 浏览专辑 → 播放曲目（断言 audio src 就绪与播放状态）
 *   → 收藏/取消收藏 → 历史页出现记录 → 登出
 *
 * Two techniques make this reliable without depending on real audio decode
 * or waiting out a full track duration:
 *
 * 1. Audio src / playback-state assertion — useAudio() creates the
 *    HTMLAudioElement via `new Audio()` and never attaches it to the DOM,
 *    so `page.locator("audio")` can't see it. We inject a subclass via
 *    `page.addInitScript` before any navigation that stashes the instance
 *    on `window.__testAudioInstance`, letting us read `.src`/`.paused`
 *    directly — the same object the app actually plays through.
 *
 * 2. History record assertion — a play is only recorded on the audio
 *    element's real "ended" event (see hooks/useAudio.ts), which normally
 *    requires waiting out the track's full duration. Instead we dispatch a
 *    synthetic "ended" Event on the same real audio instance, which drives
 *    the app's actual onEnded handler (POST /api/v1/me/history +
 *    skipToNext) — exercising real application code, not a test-only stub.
 *
 * Prerequisites (same as smoke.spec.ts):
 *   - Dev server running on http://localhost:3000 (or E2E_BASE_URL)
 *   - E2E_USERNAME / E2E_PASSWORD env vars set to a valid account
 *   - API server reachable at NEXT_PUBLIC_API_BASE_URL (default http://localhost:8080)
 *   - At least one album with at least one track seeded — the test skips
 *     itself (rather than failing) if the catalog is empty, since seeding
 *     is an environment concern, not something this spec controls.
 *
 * Run: npx playwright test main-flow.spec.ts
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

/** Capture every HTMLAudioElement the app constructs on window.__testAudioInstance. */
async function installAudioProbe(page: Page) {
  await page.addInitScript(() => {
    const NativeAudio = window.Audio;
    class ProbedAudio extends NativeAudio {
      constructor(...args: ConstructorParameters<typeof NativeAudio>) {
        super(...args);
        // @ts-expect-error test-only global
        window.__testAudioInstance = this;
      }
    }
    // @ts-expect-error overriding the built-in constructor for test observability
    window.Audio = ProbedAudio;
  });
}

test("main flow: login → browse → play → favorite → history → logout", async ({ page }) => {
  await installAudioProbe(page);
  await login(page);

  // ── Browse albums ──────────────────────────────────────────────────────
  await page.goto("/albums");
  const albumLinks = page.locator('a[href^="/albums/"]');
  // Wait for the album list to render (skeleton rows replaced by real rows)
  // before counting — otherwise .count() races the initial data fetch and
  // reads 0 even when the catalog has albums.
  try {
    await albumLinks.first().waitFor({ state: "visible", timeout: 8_000 });
  } catch {
    // fall through — the count check below will skip with a clear reason
  }
  const albumCount = await albumLinks.count();
  test.skip(albumCount === 0, "No albums seeded in this environment — skipping main flow test");

  await albumLinks.first().click();
  await expect(page).toHaveURL(/\/albums\/.+/);

  // Wait for the track list to render (skeleton rows replaced by real rows).
  const trackLinks = page.locator('a[href^="/tracks/"]');
  await expect(trackLinks.first()).toBeVisible({ timeout: 8_000 });
  const trackCount = await trackLinks.count();
  test.skip(trackCount === 0, "Album has no tracks in this environment — skipping main flow test");

  const trackTitle = (await trackLinks.first().textContent())?.trim() ?? "";
  const trackHref = await trackLinks.first().getAttribute("href");
  const trackId = trackHref?.split("/tracks/")[1] ?? "";
  expect(trackId).not.toBe("");

  // ── Play the album (starts from track 0) ───────────────────────────────
  const playbackDescriptorPromise = page.waitForResponse(
    (res) => res.url().includes("/catalog/tracks/") && res.url().includes("/playback") && res.request().method() === "GET",
    { timeout: 8_000 }
  );
  await page.getByRole("button", { name: "Play" }).click();
  const playbackRes = await playbackDescriptorPromise;
  expect(playbackRes.ok()).toBeTruthy();

  // PlayerBar must switch out of the "No track playing" placeholder.
  await expect(page.getByText(/no track playing/i)).toHaveCount(0, { timeout: 8_000 });

  // Assert the real audio element resolved a playable src ("audio src 就绪").
  await expect
    .poll(
      async () =>
        page.evaluate(() => {
          const audio = (window as unknown as { __testAudioInstance?: HTMLAudioElement }).__testAudioInstance;
          return audio?.src ?? "";
        }),
      { timeout: 8_000 }
    )
    .not.toBe("");

  // Assert playback status settled to an active state (playing or paused —
  // NOT the idle/error states), i.e. PlayerBar shows transport controls.
  await expect(page.getByTitle(/^(play|pause)$/i)).toBeVisible({ timeout: 8_000 });

  // ── Favorite / unfavorite ───────────────────────────────────────────────
  const favoriteBtn = page.getByTitle(/add to favorites|remove from favorites/i).first();
  await expect(favoriteBtn).toBeVisible();
  await favoriteBtn.click();
  await expect(page.getByTitle("Remove from favorites").first()).toBeVisible({ timeout: 5_000 });

  await page.getByRole("link", { name: "Favorites" }).click();
  await expect(page).toHaveURL(/\/library\/favorites/);
  await expect(page.getByText(/no favorites yet/i)).toHaveCount(0, { timeout: 8_000 });
  await expect(page.getByText(trackTitle, { exact: false }).first()).toBeVisible({ timeout: 8_000 });

  // Unfavorite directly from the favorites page (covers "取消收藏").
  await page.getByTitle("Remove favorite").first().click();
  await expect(page.getByText(/no favorites yet/i)).toBeVisible({ timeout: 8_000 });

  // ── Simulate the track finishing, then check history ────────────────────
  const historyPostPromise = page.waitForResponse(
    (res) => res.url().includes("/api/v1/me/history") && res.request().method() === "POST",
    { timeout: 8_000 }
  );
  await page.evaluate(() => {
    const audio = (window as unknown as { __testAudioInstance?: HTMLAudioElement }).__testAudioInstance;
    audio?.dispatchEvent(new Event("ended"));
  });
  const historyRes = await historyPostPromise;
  expect(historyRes.ok()).toBeTruthy();

  await page.getByRole("link", { name: "History" }).click();
  await expect(page).toHaveURL(/\/library\/history/);
  await page.getByRole("button", { name: "Events" }).click();
  await expect(page.getByText(trackId).first()).toBeVisible({ timeout: 8_000 });

  // ── Logout ───────────────────────────────────────────────────────────────
  await page.getByTitle("Log out").click();
  await expect(page).toHaveURL(/\/login/, { timeout: 8_000 });
});
