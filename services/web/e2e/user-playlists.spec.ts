/**
 * Personal playlist workflow.
 *
 * Requires a running web/API stack plus E2E_USERNAME and E2E_PASSWORD. The test
 * skips without credentials or when fewer than two catalog tracks are seeded;
 * it is structure for a credentialed environment, not a mocked API test.
 */
import { expect, test, type Page } from "@playwright/test";

const USERNAME = process.env.E2E_USERNAME;
const PASSWORD = process.env.E2E_PASSWORD;

async function login(page: Page) {
  await page.goto("/login");
  await page.locator("#username").fill(USERNAME ?? "");
  await page.locator("#password").fill(PASSWORD ?? "");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page).not.toHaveURL(/\/login/, { timeout: 10_000 });
}

async function addTrackAt(page: Page, index: number, playlistName: string) {
  await page.getByRole("button", { name: "Track options" }).nth(index).click();
  await page.getByRole("menuitem", { name: "Add to playlist" }).click();
  const dialog = page.getByRole("dialog", { name: "Add to playlist" });
  await expect(dialog).toBeVisible();
  const playlistButton = dialog.getByRole("button", { name: playlistName, exact: true });
  await playlistButton.click();
  await expect(playlistButton).toBeDisabled();
  await dialog.getByRole("button", { name: "Close" }).click();
}

test("personal playlist: create, add, reorder, play all, delete", async ({ page }) => {
  test.skip(!USERNAME || !PASSWORD, "Set E2E_USERNAME and E2E_PASSWORD to run the personal-playlist flow");
  await login(page);

  const playlistName = `E2E playlist ${Date.now()}`;
  await page.goto("/library/playlists");
  await page.getByRole("button", { name: "New" }).click();
  const createDialog = page.getByRole("dialog", { name: "New playlist" });
  await createDialog.getByRole("textbox", { name: "Playlist name" }).fill(playlistName);
  await createDialog.getByRole("button", { name: "Create" }).click();
  await expect(page.getByRole("link", { name: new RegExp(playlistName) })).toBeVisible();

  await page.goto("/tracks");
  const trackLinks = page.locator('a[href^="/tracks/"]');
  try {
    await trackLinks.nth(1).waitFor({ state: "visible", timeout: 8_000 });
  } catch {
    // The explicit skip below records catalog seeding as an environment concern.
  }
  test.skip((await trackLinks.count()) < 2, "At least two catalog tracks are required");

  await addTrackAt(page, 0, playlistName);
  await addTrackAt(page, 1, playlistName);

  await page.goto("/library/playlists");
  await page.getByRole("link", { name: new RegExp(playlistName) }).click();
  await expect(page.getByText("2 tracks")).toBeVisible();

  const handles = page.getByRole("button", { name: /^Reorder / });
  const firstLabel = await handles.first().getAttribute("aria-label");
  await handles.first().dragTo(handles.nth(1));
  await expect(handles.nth(1)).toHaveAttribute("aria-label", firstLabel ?? "");

  await page.getByRole("button", { name: "Play all" }).click();
  await expect(page.getByText(/no track playing/i)).toHaveCount(0, { timeout: 8_000 });

  await page.getByRole("link", { name: "My Playlists" }).click();
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByRole("button", { name: `Delete ${playlistName}` }).click();
  await expect(page.getByRole("link", { name: new RegExp(playlistName) })).toHaveCount(0);
});
