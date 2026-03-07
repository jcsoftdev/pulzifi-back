import {
  PlaywrightBlocker,
} from "@ghostery/adblocker-playwright";
import type { Page } from "patchright";

let cachedBlocker: PlaywrightBlocker | null = null;

/**
 * Returns a singleton Ghostery ad blocker instance.
 * Uses EasyList + EasyPrivacy + uBlock Origin filters.
 */
async function getBlocker(): Promise<PlaywrightBlocker> {
  if (!cachedBlocker) {
    cachedBlocker = await PlaywrightBlocker.fromPrebuiltAdsAndTracking();
  }
  return cachedBlocker;
}

/**
 * Enables network-level ad/tracker blocking on a page.
 */
export async function enableAdBlocking(page: Page): Promise<void> {
  const blocker = await getBlocker();
  // The Ghostery blocker expects a Playwright page — patchright is API-compatible
  await blocker.enableBlockingInPage(page as any);
}
