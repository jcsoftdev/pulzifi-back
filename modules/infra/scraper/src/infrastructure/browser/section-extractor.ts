import type { Page } from "patchright";
import type { SectionResult } from "../../domain/entities/extraction-result";
import type { SectionConfig } from "../../domain/value-objects/selector-config";
import type { IImageProcessor } from "../../domain/services/image-processor";
import { extractElementContent } from "./content-extractor";
import { log, logError, createTimer } from "../logger";

/**
 * Extracts content and screenshots from multiple page sections.
 * Each section is identified by a CSS selector or XPath.
 */
export async function extractSections(
  page: Page,
  sections: SectionConfig[],
  imageProcessor: IImageProcessor,
): Promise<SectionResult[]> {
  log("sections", `extracting ${sections.length} section(s)`);
  const results: SectionResult[] = [];

  for (const section of sections) {
    const sectionTimer = createTimer();
    const { html, text, matched } = await extractElementContent(page, {
      selector: section.selector,
      selectorXpath: section.selectorXpath,
      selectorOffsets: section.selectorOffsets,
    });

    let screenshotBase64 = "";
    if (matched) {
      const element = section.selector
        ? await page.$(section.selector)
        : section.selectorXpath
          ? (await page.$$(`:xpath=${section.selectorXpath}`))[0] || null
          : null;

      if (element) {
        try {
          // Scroll element into view to ensure it's rendered and visible
          await element.scrollIntoViewIfNeeded().catch(() => {});

          const offsets = section.selectorOffsets;
          const hasOffsets = offsets && (offsets.top || offsets.right || offsets.bottom || offsets.left);

          if (hasOffsets) {
            // When offsets are provided, use clip-based screenshot with adjusted region
            const clip = await element.boundingBox();
            if (clip && clip.width > 0 && clip.height > 0) {
              const adjustedClip = {
                x: Math.max(0, clip.x - (offsets.left || 0)),
                y: Math.max(0, clip.y - (offsets.top || 0)),
                width: clip.width + (offsets.left || 0) + (offsets.right || 0),
                height: clip.height + (offsets.top || 0) + (offsets.bottom || 0),
              };
              const screenshotBuffer = await page.screenshot({
                clip: adjustedClip,
                type: "png",
              });
              screenshotBase64 = await imageProcessor.pngToWebpBase64(
                Buffer.from(screenshotBuffer),
              );
              log("sections", `section "${section.id}" screenshot captured (clip)`, { elapsed: sectionTimer.elapsed(), clip: adjustedClip });
            }
          }

          // Default: use element.screenshot() which handles scrolling and clipping automatically
          if (!screenshotBase64) {
            const screenshotBuffer = await element.screenshot({ type: "png" });
            screenshotBase64 = await imageProcessor.pngToWebpBase64(
              Buffer.from(screenshotBuffer),
            );
            log("sections", `section "${section.id}" screenshot captured (element)`, { elapsed: sectionTimer.elapsed() });
          }
        } catch (err) {
          logError("sections", `section "${section.id}" screenshot failed`, err);
        }
      }
    }

    log("sections", `section "${section.id}" done`, { matched, htmlLength: html.length, textLength: text.length, hasScreenshot: screenshotBase64.length > 0, elapsed: sectionTimer.elapsed() });

    results.push({
      id: section.id,
      screenshot_base64: screenshotBase64,
      html,
      text,
      selector_matched: matched,
    });
  }

  log("sections", `all sections extracted`, { total: results.length, matched: results.filter((r) => r.selector_matched).length });
  return results;
}
