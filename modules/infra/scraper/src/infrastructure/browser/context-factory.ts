import type { Browser, BrowserContext, Page } from "patchright";
import { DEFAULT_VIEWPORT } from "../../domain/value-objects/viewport";
import { applyFingerprint } from "./stealth/fingerprint-manager";
import { enableAdBlocking } from "./blocking/ad-blocker";
import { removeCookieBanners } from "./blocking/cookie-blocker";
import {
  isCloudflareChallenge,
  waitForCloudflare,
} from "./blocking/cloudflare-handler";
import { log, createTimer } from "../logger";

const NAV_TIMEOUT_MS = parseInt(process.env.NAV_TIMEOUT_MS || "30000", 10);

export interface ContextResult {
  context: BrowserContext;
  page: Page;
}

/**
 * Creates a new browser context with stealth fingerprinting applied.
 */
export async function createStealthContext(
  browser: Browser,
): Promise<ContextResult> {
  const context = await browser.newContext({
    viewport: DEFAULT_VIEWPORT,
    ignoreHTTPSErrors: true,
    javaScriptEnabled: true,
    bypassCSP: true,
  });

  // Apply realistic fingerprint
  await applyFingerprint(context);

  const page = await context.newPage();
  page.setDefaultTimeout(NAV_TIMEOUT_MS);
  page.setDefaultNavigationTimeout(NAV_TIMEOUT_MS);

  return { context, page };
}

/**
 * Navigates to a URL with ad blocking, cookie banner removal,
 * and Cloudflare challenge handling.
 */
export async function navigateWithProtections(
  page: Page,
  url: string,
  blockAdsCookies: boolean,
): Promise<void> {
  if (blockAdsCookies) {
    const adTimer = createTimer();
    await enableAdBlocking(page);
    log("navigate", "ad blocking enabled", { url, elapsed: adTimer.elapsed() });
  }

  const gotoTimer = createTimer();
  await page.goto(url, {
    waitUntil: "domcontentloaded",
    timeout: NAV_TIMEOUT_MS,
  });
  log("navigate", "page.goto completed (domcontentloaded)", { url, elapsed: gotoTimer.elapsed() });

  // Handle Cloudflare challenge
  if (await isCloudflareChallenge(page)) {
    log("navigate", "cloudflare challenge detected, waiting...", { url });
    const cfTimer = createTimer();
    const resolved = await waitForCloudflare(page);
    log("navigate", `cloudflare challenge ${resolved ? "resolved" : "timed out"}`, { url, elapsed: cfTimer.elapsed() });
  }

  // Wait for network idle after initial load
  const idleTimer = createTimer();
  try {
    await page.waitForLoadState("networkidle", { timeout: 10000 });
    log("navigate", "network idle reached", { url, elapsed: idleTimer.elapsed() });
  } catch {
    log("navigate", "network idle timed out (proceeding)", { url, elapsed: idleTimer.elapsed() });
  }

  if (blockAdsCookies) {
    const cookieTimer = createTimer();
    await removeCookieBanners(page);
    log("navigate", "cookie banners removed", { url, elapsed: cookieTimer.elapsed() });
  }
}
