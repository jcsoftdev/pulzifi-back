import type { Page } from "patchright";
import { log } from "../logger";

/**
 * Waits for all visible <img> elements on the page to finish loading.
 * Handles lazy-loaded images that have been triggered by scrolling but
 * haven't completed their network fetch yet. Each image gets an individual
 * timeout to avoid blocking on broken images.
 */
export async function waitForImages(
  page: Page,
  options: { perImageTimeoutMs?: number; totalTimeoutMs?: number } = {},
): Promise<void> {
  const { perImageTimeoutMs = 5000, totalTimeoutMs = 15000 } = options;

  try {
    const result = await page.evaluate(
      async ({ perImageTimeoutMs, totalTimeoutMs }) => {
        const deadline = Date.now() + totalTimeoutMs;
        const images = Array.from(document.querySelectorAll("img"));

        // Filter to images that have a src and are not already complete
        const pending = images.filter(
          (img) => img.src && !img.complete && img.naturalWidth === 0,
        );

        if (pending.length === 0) return { total: images.length, waited: 0 };

        const waitForOne = (img: HTMLImageElement): Promise<void> =>
          new Promise<void>((resolve) => {
            // Already loaded while we were setting up
            if (img.complete && img.naturalWidth > 0) {
              resolve();
              return;
            }

            const remaining = Math.min(
              perImageTimeoutMs,
              deadline - Date.now(),
            );
            if (remaining <= 0) {
              resolve();
              return;
            }

            const timer = setTimeout(resolve, remaining);
            img.addEventListener(
              "load",
              () => {
                clearTimeout(timer);
                resolve();
              },
              { once: true },
            );
            img.addEventListener(
              "error",
              () => {
                clearTimeout(timer);
                resolve();
              },
              { once: true },
            );
          });

        await Promise.all(pending.map(waitForOne));
        return { total: images.length, waited: pending.length };
      },
      { perImageTimeoutMs, totalTimeoutMs },
    );

    log("extract", "images loaded", {
      total: result.total,
      waited: result.waited,
    });
  } catch {
    log("extract", "image load wait failed (non-critical)", {});
  }
}
