import type { Page } from "patchright";

/**
 * Waits for the DOM to stabilize by observing mutations.
 * Returns when no mutations have occurred for `quietMs` milliseconds.
 */
export async function waitForRenderStable(
  page: Page,
  options: { quietMs?: number; timeoutMs?: number } = {},
): Promise<void> {
  const { quietMs = 500, timeoutMs = 10000 } = options;

  await page.evaluate(
    ({ quietMs, timeoutMs }) => {
      return new Promise<void>((resolve) => {
        let timer: ReturnType<typeof setTimeout>;
        const deadline = setTimeout(() => {
          observer.disconnect();
          resolve();
        }, timeoutMs);

        const observer = new MutationObserver(() => {
          clearTimeout(timer);
          timer = setTimeout(() => {
            observer.disconnect();
            clearTimeout(deadline);
            resolve();
          }, quietMs);
        });

        observer.observe(document.body, {
          childList: true,
          subtree: true,
          attributes: true,
        });

        // Start quiet timer in case DOM is already stable
        timer = setTimeout(() => {
          observer.disconnect();
          clearTimeout(deadline);
          resolve();
        }, quietMs);
      });
    },
    { quietMs, timeoutMs },
  );
}
