import type { Page } from "patchright";

/**
 * Common selectors for cookie consent banners.
 * These are removed from the DOM to prevent overlays in screenshots.
 */
const COOKIE_BANNER_SELECTORS = [
  // ID-based
  "#cookie-banner",
  "#cookie-consent",
  "#cookie-notice",
  "#cookieConsent",
  "#CybotCookiebotDialog",
  "#onetrust-banner-sdk",
  "#onetrust-consent-sdk",
  "#consent-banner",
  "#gdpr-cookie-notice",
  "#cc-main",
  // Class-based
  ".cookie-banner",
  ".cookie-consent",
  ".cookie-notice",
  ".consent-banner",
  ".cc-banner",
  ".cc-window",
  ".cc-dialog",
  ".gdpr-banner",
  // Data attribute-based
  '[class*="cookie-banner"]',
  '[class*="cookie"]',
  '[class*="Cookie"]',
  '[class*="cookie-consent"]',
  '[class*="CookieConsent"]',
  '[id*="cookie-banner"]',
  '[id*="cookie-consent"]',
  '[aria-label*="cookie"]',
  '[aria-label*="Cookie"]',
  '[aria-label*="consent"]',
];

/**
 * Removes cookie consent banners from the page DOM.
 */
export async function removeCookieBanners(page: Page): Promise<void> {
  await page.evaluate((selectors: string[]) => {
    for (const selector of selectors) {
      try {
        const elements = document.querySelectorAll(selector);
        elements.forEach((el) => el.remove());
      } catch {
        // Invalid selector — skip
      }
    }
    // Also remove elements with high z-index that look like overlays
    const allElements = document.querySelectorAll("div, section, aside");
    for (const el of allElements) {
      const style = getComputedStyle(el);
      if (
        style.position === "fixed" &&
        parseInt(style.zIndex || "0") > 9000 &&
        el.querySelector('button, [class*="accept"], [class*="agree"]')
      ) {
        el.remove();
      }
    }
  }, COOKIE_BANNER_SELECTORS);
}
