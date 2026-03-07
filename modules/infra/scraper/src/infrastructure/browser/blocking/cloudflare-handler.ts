import type { Page } from "patchright";

const CF_TITLES = [
  "Just a moment...",
  "Attention Required! | Cloudflare",
  "Please Wait... | Cloudflare",
];

const CF_MAX_WAIT_MS = parseInt(
  process.env.CF_MAX_WAIT_MS || "20000",
  10,
);

/**
 * Detects if a Cloudflare challenge page is shown.
 */
export async function isCloudflareChallenge(page: Page): Promise<boolean> {
  const title = await page.title();
  return CF_TITLES.some((t) => title.includes(t));
}

/**
 * Waits for a Cloudflare challenge to resolve by polling the title.
 * Returns true if the challenge was resolved, false if it timed out.
 */
export async function waitForCloudflare(page: Page): Promise<boolean> {
  const start = Date.now();
  const pollInterval = 1000;

  while (Date.now() - start < CF_MAX_WAIT_MS) {
    const isCF = await isCloudflareChallenge(page);
    if (!isCF) return true;
    await page.waitForTimeout(pollInterval);
  }

  return false;
}
