import type { Page, ElementHandle } from "patchright";
import type { SelectorConfig } from "../../domain/value-objects/selector-config";
import { log } from "../logger";

export interface ExtractedContent {
  title: string;
  html: string;
  text: string;
  selectorMatched: boolean;
}

/**
 * Resolves a target element from a selector config.
 * Tries CSS selector first, then XPath.
 */
async function resolveElement(
  page: Page,
  config: SelectorConfig,
): Promise<ElementHandle | null> {
  if (config.selector) {
    const el = await page.$(config.selector);
    if (el) {
      log("content", "element resolved via CSS selector", { selector: config.selector });
      return el;
    }
    log("content", "CSS selector did not match", { selector: config.selector });
  }
  if (config.selectorXpath) {
    const els = await page.$$(`:xpath=${config.selectorXpath}`);
    if (els.length > 0) {
      log("content", "element resolved via XPath", { xpath: config.selectorXpath });
      return els[0];
    }
    log("content", "XPath selector did not match", { xpath: config.selectorXpath });
  }
  return null;
}

/**
 * Extracts title, HTML, and text content from a page.
 * If a selector is provided, extracts from the matched element only.
 */
export async function extractContent(
  page: Page,
  selector?: SelectorConfig,
): Promise<ExtractedContent> {
  const title = await page.title();
  log("content", "page title extracted", { title: title.substring(0, 100) });

  if (selector && (selector.selector || selector.selectorXpath)) {
    log("content", "attempting selector-based extraction", { selector: selector.selector, xpath: selector.selectorXpath });
    const element = await resolveElement(page, selector);
    if (element) {
      const html = await element.innerHTML();
      const text = await element.innerText().catch(() => "");
      log("content", "selector-based extraction succeeded", { htmlLength: html.length, textLength: text.length });
      return { title, html, text, selectorMatched: true };
    }
    log("content", "selector not matched, falling back to full page extraction");
  }

  const html = await page.content();
  const text = await page.evaluate(() => {
    const body = document.body;
    if (!body) return "";
    const clone = body.cloneNode(true) as HTMLElement;
    clone.querySelectorAll("script, style, noscript").forEach((el) => el.remove());
    return clone.innerText || "";
  });

  log("content", "full page content extracted", { htmlLength: html.length, textLength: text.length });

  return {
    title,
    html,
    text,
    selectorMatched: !selector || (!selector.selector && !selector.selectorXpath),
  };
}

/**
 * Extracts content from a specific element by selector config.
 * Returns null content if selector doesn't match.
 */
export async function extractElementContent(
  page: Page,
  config: SelectorConfig,
): Promise<{ html: string; text: string; matched: boolean }> {
  const element = await resolveElement(page, config);
  if (!element) {
    return { html: "", text: "", matched: false };
  }
  const html = await element.innerHTML();
  const text = await element.innerText().catch(() => "");
  return { html, text, matched: true };
}
