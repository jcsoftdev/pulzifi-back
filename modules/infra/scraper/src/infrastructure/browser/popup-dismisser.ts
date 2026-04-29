import type { Page } from "patchright";
import { log } from "../logger";

/**
 * Removes popup modals, cookie banners, and overlay elements from the DOM
 * before content extraction and screenshotting. These elements appear
 * inconsistently between page loads and cause false-positive change detection.
 *
 * Strategy:
 * 1. Click visible close/dismiss buttons on overlays
 * 2. Wait briefly for dismiss animations
 * 3. Remove remaining fixed/absolute overlays that cover significant viewport area
 * 4. Remove orphaned backdrop/overlay elements
 * 5. Restore body scroll (popups often set overflow:hidden)
 */
export async function dismissPopups(page: Page): Promise<number> {
  try {
    const removed = await page.evaluate(() => {
      let removedCount = 0;
      const vw = window.innerWidth;
      const vh = window.innerHeight;
      const viewportArea = vw * vh;

      // Phase 1: Try clicking close buttons on visible overlays
      const closeSelectors = [
        // Common close button patterns
        '[class*="close" i][role="button"]',
        '[class*="close" i]:is(button, a, span, div)',
        '[aria-label*="close" i]',
        '[aria-label*="dismiss" i]',
        '[class*="dismiss" i]',
        'button[class*="modal" i][class*="close" i]',
        // Cookie consent dismiss
        '[class*="cookie" i] button',
        '[id*="cookie" i] button',
        '[class*="consent" i] button',
      ];

      for (const sel of closeSelectors) {
        try {
          const btns = document.querySelectorAll(sel);
          for (const btn of btns) {
            const el = btn as HTMLElement;
            const rect = el.getBoundingClientRect();
            // Only click buttons that are visible and in the viewport
            if (rect.width > 0 && rect.height > 0 && rect.top >= 0 && rect.left >= 0) {
              el.click();
            }
          }
        } catch {
          // Ignore click errors
        }
      }

      // Phase 2: Remove fixed/absolute overlays covering significant viewport area
      const allElements = document.querySelectorAll("*");
      const toRemove: Element[] = [];

      for (const el of allElements) {
        const style = window.getComputedStyle(el);
        const position = style.position;

        if (position !== "fixed" && position !== "absolute") continue;

        // Skip tiny elements (icons, tooltips)
        const rect = el.getBoundingClientRect();
        if (rect.width < 100 || rect.height < 100) continue;

        const elArea = rect.width * rect.height;
        const coverageRatio = elArea / viewportArea;

        // Element covers > 30% of viewport — likely a modal or overlay
        if (coverageRatio > 0.3) {
          // Check z-index — overlays typically have high z-index
          const zIndex = parseInt(style.zIndex, 10);
          const hasHighZ = !isNaN(zIndex) && zIndex > 10;
          const isOverlay =
            hasHighZ ||
            style.backgroundColor.includes("rgba") || // semi-transparent backdrop
            el.classList.toString().toLowerCase().match(/modal|popup|overlay|dialog|banner|cookie|consent|newsletter/) ||
            el.id.toLowerCase().match(/modal|popup|overlay|dialog|banner|cookie|consent|newsletter/);

          if (isOverlay) {
            toRemove.push(el);
          }
        }
      }

      // Remove identified overlays (children first to avoid double-removal)
      const uniqueToRemove = toRemove.filter(
        (el) => !toRemove.some((other) => other !== el && other.contains(el)),
      );

      for (const el of uniqueToRemove) {
        el.remove();
        removedCount++;
      }

      // Phase 3: Restore body scroll in case popup set overflow:hidden
      document.body.style.overflow = "";
      document.documentElement.style.overflow = "";

      return removedCount;
    });

    if (removed > 0) {
      log("extract", "popups dismissed", { removed });
      // Brief settle after DOM changes
      await new Promise((r) => setTimeout(r, 100));
    }

    return removed;
  } catch {
    log("extract", "popup dismissal failed (non-critical)", {});
    return 0;
  }
}
