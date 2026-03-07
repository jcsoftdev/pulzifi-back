import { chromium, type Browser } from "patchright";
import type {
  IBrowserService,
  ExtractOptions,
  PreviewOptions,
  ProgressCallback,
} from "../../domain/services/browser-service";
import type { ExtractionResult } from "../../domain/entities/extraction-result";
import type { PreviewResult } from "../../domain/entities/preview-result";
import type { IImageProcessor } from "../../domain/services/image-processor";
import { BrowserError, NavigationError } from "../../domain/errors/scraper-errors";
import { DEFAULT_VIEWPORT } from "../../domain/value-objects/viewport";
import { createStealthContext, navigateWithProtections } from "./context-factory";
import { extractContent } from "./content-extractor";
import { extractSections } from "./section-extractor";
import { mapSemanticElements } from "./element-mapper";
import { scrollFullPage } from "./page-scroller";
import { waitForRenderStable } from "./render-waiter";
import { log, logError, createTimer } from "../logger";

const MAX_CONCURRENT = parseInt(process.env.MAX_CONCURRENT_PAGES || "3", 10);

export class PatchrightBrowserService implements IBrowserService {
  private browser: Browser | null = null;
  private launching = false;
  private activeTasks = 0;
  private taskQueue: Array<() => void> = [];

  constructor(private imageProcessor: IImageProcessor) {}

  private get launchOptions() {
    const executablePath = process.env.CHROMIUM_PATH || undefined;
    return {
      headless: true as const,
      executablePath,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--disable-dev-shm-usage",
        "--disable-gpu",
        "--no-first-run",
        "--disable-extensions",
        "--disable-background-networking",
        "--disable-default-apps",
        "--disable-sync",
        "--disable-translate",
      ],
    };
  }

  async launch(): Promise<void> {
    this.launching = true;
    try {
      this.browser = await chromium.launch(this.launchOptions);
      this.browser.on("disconnected", () => {
        console.warn("[browser] Chromium disconnected unexpectedly");
        this.browser = null;
      });
      console.log("[browser] Chromium launched");
    } finally {
      this.launching = false;
    }
  }

  /** Re-launch browser if it has crashed */
  private async ensureBrowser(): Promise<Browser> {
    if (this.browser?.isConnected()) return this.browser;
    if (this.launching) {
      // Wait for in-flight launch
      await new Promise<void>((resolve) => {
        const check = () => {
          if (!this.launching) return resolve();
          setTimeout(check, 100);
        };
        check();
      });
      if (this.browser?.isConnected()) return this.browser;
    }
    console.log("[browser] Re-launching Chromium...");
    await this.launch();
    return this.browser!;
  }

  isHealthy(): boolean {
    return this.browser !== null && this.browser.isConnected();
  }

  async shutdown(): Promise<void> {
    if (this.browser) {
      await this.browser.close();
      this.browser = null;
      console.log("[browser] Chromium closed");
    }
  }

  private async acquireSlot(): Promise<void> {
    if (this.activeTasks < MAX_CONCURRENT) {
      this.activeTasks++;
      return;
    }
    return new Promise<void>((resolve) => {
      this.taskQueue.push(() => {
        this.activeTasks++;
        resolve();
      });
    });
  }

  private releaseSlot(): void {
    this.activeTasks--;
    const next = this.taskQueue.shift();
    if (next) next();
  }

  async extract(options: ExtractOptions): Promise<ExtractionResult> {
    const timer = createTimer();
    const url = options.url;
    log("extract", "starting extraction", { url, blockAds: options.blockAdsCookies, hasSelector: !!options.selector, sectionsCount: options.sections?.length ?? 0 });

    const browser = await this.ensureBrowser();

    log("extract", "waiting for concurrency slot", { url, activeTasks: this.activeTasks, queued: this.taskQueue.length });
    await this.acquireSlot();
    log("extract", "slot acquired", { url });

    const stepTimer = createTimer();
    const { context, page } = await createStealthContext(browser);
    log("extract", "stealth context created", { url, elapsed: stepTimer.elapsed() });

    try {
      const navTimer = createTimer();
      await navigateWithProtections(page, options.url, options.blockAdsCookies);
      log("extract", "navigation completed", { url, elapsed: navTimer.elapsed() });

      // Scroll to trigger lazy loading, then wait for stability
      const scrollTimer = createTimer();
      await scrollFullPage(page);
      log("extract", "full page scroll completed", { url, elapsed: scrollTimer.elapsed() });

      const renderTimer = createTimer();
      await waitForRenderStable(page);
      log("extract", "render stabilized", { url, elapsed: renderTimer.elapsed() });

      // Extract content
      const selectorConfig = options.selector
        ? {
            selector: options.selector.selector,
            selectorXpath: options.selector.selectorXpath,
            selectorOffsets: options.selector.selectorOffsets,
          }
        : undefined;

      const contentTimer = createTimer();
      const content = await extractContent(page, selectorConfig);
      log("extract", "content extracted", { url, elapsed: contentTimer.elapsed(), titleLength: content.title.length, htmlLength: content.html.length, textLength: content.text.length, selectorMatched: content.selectorMatched });

      // Take full-page screenshot
      const screenshotTimer = createTimer();
      let screenshotBase64 = "";
      if (selectorConfig && (selectorConfig.selector || selectorConfig.selectorXpath)) {
        // Screenshot of matched element
        const el = selectorConfig.selector
          ? await page.$(selectorConfig.selector)
          : selectorConfig.selectorXpath
            ? (await page.$$(`:xpath=${selectorConfig.selectorXpath}`))[0] || null
            : null;

        if (el) {
          const clip = await el.boundingBox();
          if (clip) {
            const offsets = selectorConfig.selectorOffsets;
            const adjustedClip = offsets
              ? {
                  x: clip.x - (offsets.left || 0),
                  y: clip.y - (offsets.top || 0),
                  width: clip.width + (offsets.left || 0) + (offsets.right || 0),
                  height: clip.height + (offsets.top || 0) + (offsets.bottom || 0),
                }
              : clip;
            const buf = await page.screenshot({ clip: adjustedClip, type: "png" });
            screenshotBase64 = await this.imageProcessor.pngToWebpBase64(Buffer.from(buf));
            log("extract", "element screenshot captured", { url, elapsed: screenshotTimer.elapsed(), clip: adjustedClip });
          }
        }
      }

      if (!screenshotBase64) {
        const buf = await page.screenshot({ fullPage: true, type: "png" });
        screenshotBase64 = await this.imageProcessor.cropToWidthAndConvert(
          Buffer.from(buf),
          DEFAULT_VIEWPORT.width,
        );
        log("extract", "full-page screenshot captured", { url, elapsed: screenshotTimer.elapsed(), size: screenshotBase64.length });
      }

      // Extract sections if provided
      let sections: ExtractionResult["sections"];
      if (options.sections && options.sections.length > 0) {
        const sectionTimer = createTimer();
        sections = await extractSections(page, options.sections, this.imageProcessor);
        const matched = sections?.filter((s) => s.selector_matched).length ?? 0;
        log("extract", "sections extracted", { url, elapsed: sectionTimer.elapsed(), total: sections?.length ?? 0, matched });
      }

      log("extract", "extraction completed", { url, totalElapsed: timer.elapsed() });

      return {
        title: content.title,
        html: content.html,
        text: content.text,
        screenshot_base64: screenshotBase64,
        selector_matched: content.selectorMatched,
        sections,
      };
    } catch (err: any) {
      logError("extract", "extraction failed", err, { url, elapsed: timer.elapsed() });
      if (err.message?.includes("net::ERR_") || err.message?.includes("Navigation")) {
        throw new NavigationError(err.message, options.url);
      }
      throw err;
    } finally {
      await context.close().catch(() => {});
      this.releaseSlot();
      log("extract", "slot released", { url, activeTasks: this.activeTasks, queued: this.taskQueue.length });
    }
  }

  async preview(
    options: PreviewOptions,
    onProgress?: ProgressCallback,
  ): Promise<PreviewResult> {
    const timer = createTimer();
    const url = options.url;
    log("preview", "starting preview", { url, blockAds: options.blockAdsCookies, sectionDepth: options.sectionDepth });

    const browser = await this.ensureBrowser();

    log("preview", "waiting for concurrency slot", { url, activeTasks: this.activeTasks, queued: this.taskQueue.length });
    await this.acquireSlot();
    log("preview", "slot acquired", { url });

    const { context, page } = await createStealthContext(browser);

    try {
      const totalSteps = 5;

      onProgress?.(1, totalSteps, "Navigating to page...");
      const navTimer = createTimer();
      await navigateWithProtections(page, options.url, options.blockAdsCookies);
      log("preview", "step 1/5: navigation completed", { url, elapsed: navTimer.elapsed() });

      onProgress?.(2, totalSteps, "Scrolling page for lazy content...");
      const scrollTimer = createTimer();
      await scrollFullPage(page);
      await waitForRenderStable(page);
      log("preview", "step 2/5: scroll + render stable", { url, elapsed: scrollTimer.elapsed() });

      onProgress?.(3, totalSteps, "Capturing screenshot...");
      const screenshotTimer = createTimer();
      const buf = await page.screenshot({ fullPage: true, type: "png" });
      const screenshotBase64 = await this.imageProcessor.cropToWidthAndConvert(
        Buffer.from(buf),
        DEFAULT_VIEWPORT.width,
      );
      log("preview", "step 3/5: screenshot captured", { url, elapsed: screenshotTimer.elapsed(), size: screenshotBase64.length });

      const pageHeight = await page.evaluate(() =>
        Math.max(document.body.scrollHeight, document.documentElement.scrollHeight),
      );

      onProgress?.(4, totalSteps, "Mapping page elements...");
      const mapTimer = createTimer();
      const depth = Math.max(1, Math.min(5, options.sectionDepth ?? 3));
      const elements = await mapSemanticElements(page, depth);
      log("preview", "step 4/5: elements mapped", { url, elapsed: mapTimer.elapsed(), elementsFound: elements.length, depth });

      onProgress?.(5, totalSteps, "Done!");
      log("preview", "step 5/5: preview completed", { url, totalElapsed: timer.elapsed(), pageHeight, elements: elements.length });

      return {
        screenshot_base64: screenshotBase64,
        viewport: DEFAULT_VIEWPORT,
        page_height: pageHeight,
        elements,
      };
    } catch (err: any) {
      logError("preview", "preview failed", err, { url, elapsed: timer.elapsed() });
      if (err.message?.includes("net::ERR_") || err.message?.includes("Navigation")) {
        throw new NavigationError(err.message, options.url);
      }
      throw err;
    } finally {
      await context.close().catch(() => {});
      this.releaseSlot();
      log("preview", "slot released", { url, activeTasks: this.activeTasks, queued: this.taskQueue.length });
    }
  }
}
