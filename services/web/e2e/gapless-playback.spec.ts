/**
 * gapless-playback.spec.ts — v5.2.0 dual-element gapless engine e2e coverage.
 *
 * Two scenarios, best-effort per the v5.2.0 plan (.plan/20260706-021-v5.2.0-web-audio-engine.md):
 *
 * 1. "two-track queue auto-advances" — plays an album/queue with >=2 tracks,
 *    dispatches a synthetic "ended" event on the currently active probed
 *    audio element (same technique as main-flow.spec.ts — real decode/full
 *    track duration isn't practical in CI), and asserts the player advances
 *    to the second track (currentIndex/track title changes) without the
 *    "No track playing" placeholder reappearing — i.e. the swap-to-standby
 *    path in useAudio.ts fired instead of falling all the way back to idle.
 *
 * 2. "refresh restores queue and position without autoplay" — starts
 *    playback, waits for position to advance, reloads the page, and asserts
 *    the PlayerBar shows the same track (queue restored) while playback
 *    remains paused (no autoplay) per the persisted state's `restoredPending`
 *    contract in store/player.ts.
 *
 * Prerequisites (same as main-flow.spec.ts):
 *   - Dev server running on http://localhost:3000 (or E2E_BASE_URL)
 *   - E2E_USERNAME / E2E_PASSWORD env vars set to a valid account
 *   - API server reachable at NEXT_PUBLIC_API_BASE_URL (default http://localhost:8080)
 *   - At least one album with at least 2 tracks seeded — tests skip
 *     themselves (rather than failing) if the catalog doesn't meet this,
 *     since seeding is an environment concern.
 *
 * NOTE: unverified pending live env — see final report ("what I could not
 * verify locally"). Written to match this repo's established Playwright
 * conventions (main-flow.spec.ts / smoke.spec.ts) but not executed against a
 * running dev server + API + seeded data in this session.
 *
 * Run: npx playwright test gapless-playback.spec.ts
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

/**
 * Captures every HTMLAudioElement the app constructs on
 * window.__testAudioInstances (plural — the gapless engine creates two).
 * The most-recently-played one is exposed as __testAudioInstance for
 * compatibility with the single-element assertions used elsewhere.
 */
async function installAudioProbe(page: Page) {
  await page.addInitScript(() => {
    const NativeAudio = window.Audio;
    class ProbedAudio extends NativeAudio {
      constructor(...args: ConstructorParameters<typeof NativeAudio>) {
        super(...args);
        // @ts-expect-error test-only global
        window.__testAudioInstances = window.__testAudioInstances ?? [];
        // @ts-expect-error test-only global
        window.__testAudioInstances.push(this);
        this.addEventListener("play", () => {
          // @ts-expect-error test-only global
          window.__testAudioInstance = this;
        });
      }
    }
    window.Audio = ProbedAudio;
  });
}

/** Dispatches a synthetic "ended" event on whichever probed element is currently active/playing. */
async function dispatchEndedOnActiveAudio(page: Page) {
  await page.evaluate(() => {
    const audio = (window as unknown as { __testAudioInstance?: HTMLAudioElement }).__testAudioInstance;
    audio?.dispatchEvent(new Event("ended"));
  });
}

test("gapless: two-track queue auto-advances to the next track on ended", async ({ page }) => {
  await installAudioProbe(page);
  await login(page);

  await page.goto("/albums");
  const albumLinks = page.locator('a[href^="/albums/"]');
  try {
    await albumLinks.first().waitFor({ state: "visible", timeout: 8_000 });
  } catch {
    // fall through — skip check below reports the reason
  }
  const albumCount = await albumLinks.count();
  test.skip(albumCount === 0, "No albums seeded in this environment — skipping gapless test");

  await albumLinks.first().click();
  await expect(page).toHaveURL(/\/albums\/.+/);

  const trackLinks = page.locator('a[href^="/tracks/"]');
  await expect(trackLinks.first()).toBeVisible({ timeout: 8_000 });
  const trackCount = await trackLinks.count();
  test.skip(trackCount < 2, "Album needs at least 2 tracks for gapless auto-advance — skipping");

  const firstTitle = (await trackLinks.nth(0).textContent())?.trim() ?? "";
  const secondTitle = (await trackLinks.nth(1).textContent())?.trim() ?? "";
  expect(firstTitle).not.toBe("");
  expect(secondTitle).not.toBe("");
  expect(firstTitle).not.toBe(secondTitle);

  // Start playback from the first track.
  const playbackDescriptorPromise = page.waitForResponse(
    (res) =>
      res.url().includes("/catalog/tracks/") && res.url().includes("/playback") && res.request().method() === "GET",
    { timeout: 8_000 }
  );
  await page.getByRole("button", { name: "Play" }).click();
  await playbackDescriptorPromise;
  await expect(page.getByText(/no track playing/i)).toHaveCount(0, { timeout: 8_000 });

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

  // Assert the PlayerBar shows the first track's title before advancing.
  await expect(page.getByText(firstTitle, { exact: false }).first()).toBeVisible({ timeout: 8_000 });

  // Simulate the first track finishing — drives useAudio's onEnded ->
  // skipToNext(), which either swaps to a preloaded standby element or
  // falls back to resolving the next track directly (both paths are
  // exercised by the app's real code, just like main-flow.spec.ts's
  // single-track ended assertion).
  const secondPlaybackPromise = page
    .waitForResponse(
      (res) =>
        res.url().includes("/catalog/tracks/") && res.url().includes("/playback") && res.request().method() === "GET",
      { timeout: 8_000 }
    )
    .catch(() => null); // gapless swap path may not issue a fresh request if already preloaded
  await dispatchEndedOnActiveAudio(page);
  await secondPlaybackPromise;

  // PlayerBar must now show the second track, and never regress to the
  // "no track playing" idle placeholder in between.
  await expect(page.getByText(secondTitle, { exact: false }).first()).toBeVisible({ timeout: 8_000 });
  await expect(page.getByText(/no track playing/i)).toHaveCount(0);
});

test("persistence: refresh restores queue and current track without autoplay", async ({ page }) => {
  await installAudioProbe(page);
  await login(page);

  await page.goto("/albums");
  const albumLinks = page.locator('a[href^="/albums/"]');
  try {
    await albumLinks.first().waitFor({ state: "visible", timeout: 8_000 });
  } catch {
    // fall through
  }
  const albumCount = await albumLinks.count();
  test.skip(albumCount === 0, "No albums seeded in this environment — skipping persistence test");

  await albumLinks.first().click();
  await expect(page).toHaveURL(/\/albums\/.+/);

  const trackLinks = page.locator('a[href^="/tracks/"]');
  await expect(trackLinks.first()).toBeVisible({ timeout: 8_000 });
  const trackCount = await trackLinks.count();
  test.skip(trackCount === 0, "Album has no tracks in this environment — skipping persistence test");

  const trackTitle = (await trackLinks.first().textContent())?.trim() ?? "";

  const playbackDescriptorPromise = page.waitForResponse(
    (res) =>
      res.url().includes("/catalog/tracks/") && res.url().includes("/playback") && res.request().method() === "GET",
    { timeout: 8_000 }
  );
  await page.getByRole("button", { name: "Play" }).click();
  await playbackDescriptorPromise;
  await expect(page.getByText(/no track playing/i)).toHaveCount(0, { timeout: 8_000 });

  // Let the position tick advance a bit and the throttled localStorage write flush.
  await page.waitForTimeout(6_000);

  await page.reload();

  // Queue/current track restored — PlayerBar shows the same track title,
  // and does NOT fall back to the idle placeholder.
  await expect(page.getByText(/no track playing/i)).toHaveCount(0, { timeout: 8_000 });
  await expect(page.getByText(trackTitle, { exact: false }).first()).toBeVisible({ timeout: 8_000 });

  // No autoplay: the transport control must show "Play" (not "Pause"),
  // i.e. playback status settled to "paused", never auto-resumed.
  await expect(page.getByTitle("Play")).toBeVisible({ timeout: 8_000 });
});
