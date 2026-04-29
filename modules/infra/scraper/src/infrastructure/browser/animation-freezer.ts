import type { Page } from "patchright";
import { log } from "../logger";

/**
 * Phase 1: Wait for JS-driven text animations (typewriters, tickers, counters)
 * to reach a stable state, then kill all timers.
 *
 * Strategy:
 * 1. Observe DOM text mutations via MutationObserver
 * 2. Wait until no text changes for `quietMs` (default 800ms) — this catches
 *    the natural pause after a typewriter finishes typing a word.
 * 3. Then clear all timers to freeze the state.
 * 4. Timeout after `maxWaitMs` to avoid blocking on continuous animations.
 *
 * Call this early — right after navigation, before scrolling.
 */
export async function freezeJsTimers(
  page: Page,
  options: { quietMs?: number; maxWaitMs?: number } = {},
): Promise<void> {
  const { quietMs = 800, maxWaitMs = 5000 } = options;

  try {
    await page.evaluate(
      ({ quietMs, maxWaitMs }) => {
        return new Promise<void>((resolve) => {
          let quietTimer: ReturnType<typeof setTimeout> | null = null;
          let hasTextMutation = false;

          const observer = new MutationObserver((mutations) => {
            // Only react to characterData (text node) changes
            const hasText = mutations.some(
              (m) =>
                m.type === "characterData" ||
                (m.type === "childList" &&
                  (m.addedNodes.length > 0 || m.removedNodes.length > 0)),
            );
            if (!hasText) return;

            hasTextMutation = true;

            // Reset quiet period timer
            if (quietTimer !== null) {
              clearTimeout(quietTimer);
            }
            quietTimer = setTimeout(() => {
              // Text has been quiet — freeze now
              observer.disconnect();
              killTimers();
              resolve();
            }, quietMs);
          });

          observer.observe(document.body, {
            characterData: true,
            childList: true,
            subtree: true,
          });

          // If no text mutation at all after a short grace period, freeze immediately
          const graceTimer = setTimeout(() => {
            if (!hasTextMutation) {
              observer.disconnect();
              if (quietTimer !== null) clearTimeout(quietTimer);
              killTimers();
              resolve();
            }
          }, 500);

          // Hard timeout: freeze no matter what
          const hardTimeout = setTimeout(() => {
            observer.disconnect();
            if (quietTimer !== null) clearTimeout(quietTimer);
            clearTimeout(graceTimer);
            killTimers();
            resolve();
          }, maxWaitMs);

          function killTimers() {
            clearTimeout(hardTimeout);
            clearTimeout(graceTimer);
            // Clear all timers set by the page
            const lastId = window.setTimeout(() => {}, 0);
            for (let i = 1; i <= lastId; i++) {
              window.clearInterval(i);
              window.clearTimeout(i);
            }
            // Cancel pending rAF callbacks
            const lastRafId = window.requestAnimationFrame(() => {});
            for (let i = 1; i <= lastRafId; i++) {
              window.cancelAnimationFrame(i);
            }
          }
        });
      },
      { quietMs, maxWaitMs },
    );
  } catch {
    log("extract", "JS timer freeze failed (non-critical)", {});
  }
}

/**
 * Phase 2: Freeze CSS animations, media, and Web Animations API.
 * Also re-clears JS timers to catch any created during scroll/rendering.
 * Call this right before taking the screenshot for visual determinism.
 */
export async function freezeAnimations(page: Page): Promise<void> {
  try {
    await page.evaluate(() => {
      // 1. CSS: snap all animations/transitions to their final frame
      const style = document.createElement("style");
      style.textContent = `
        *, *::before, *::after {
          animation-duration: 0s !important;
          animation-delay: 0s !important;
          animation-iteration-count: 1 !important;
          transition-duration: 0s !important;
          transition-delay: 0s !important;
          scroll-behavior: auto !important;
        }
      `;
      document.head.appendChild(style);

      // 2. JS: clear all timers again (catch scroll-triggered ones) and neutralize
      const lastId = window.setTimeout(() => {}, 0);
      for (let i = 1; i <= lastId; i++) {
        window.clearInterval(i);
        window.clearTimeout(i);
      }
      window.setTimeout = (() => 0) as unknown as typeof window.setTimeout;
      window.setInterval = (() => 0) as unknown as typeof window.setInterval;

      // 3. Media: pause all video and audio elements
      document
        .querySelectorAll("video, audio")
        .forEach((el) => (el as HTMLMediaElement).pause());

      // 4. Web Animations API: finish (or cancel infinite) animations
      if (typeof document.getAnimations === "function") {
        document.getAnimations().forEach((a) => {
          try {
            a.finish();
          } catch {
            a.cancel();
          }
        });
      }
    });

    // 5. Settle: ~3 frames at 60fps for CSS overrides to take effect
    await new Promise((r) => setTimeout(r, 50));
  } catch {
    log("extract", "animation freeze failed (non-critical)", {});
  }
}
