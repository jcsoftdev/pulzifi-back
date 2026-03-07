import type { Page } from "patchright";

/**
 * Scrolls the page incrementally to trigger lazy-loaded content,
 * then scrolls back to top.
 */
export async function scrollFullPage(
  page: Page,
  options: { stepPx?: number; delayMs?: number } = {},
): Promise<void> {
  const { stepPx = 400, delayMs = 100 } = options;

  await page.evaluate(
    async ({ stepPx, delayMs }) => {
      const totalHeight = document.body.scrollHeight;
      let currentPosition = 0;

      while (currentPosition < totalHeight) {
        window.scrollBy(0, stepPx);
        currentPosition += stepPx;
        await new Promise((r) => setTimeout(r, delayMs));
      }

      // Scroll back to top
      window.scrollTo(0, 0);
      await new Promise((r) => setTimeout(r, 200));
    },
    { stepPx, delayMs },
  );
}
