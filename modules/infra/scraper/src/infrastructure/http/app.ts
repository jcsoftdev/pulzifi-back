import { Hono } from "hono";
import { streamSSE } from "hono/streaming";
import { ExtractPageHandler } from "../../application/extract-page/handler";
import { PreviewPageHandler } from "../../application/preview-page/handler";
import { HealthCheckHandler } from "../../application/health-check/handler";
import { ScraperError } from "../../domain/errors/scraper-errors";
import type { ExtractPageRequest } from "../../application/extract-page/request";
import type { PreviewPageRequest } from "../../application/preview-page/request";
import { log, logError, createRequestId, createTimer } from "../logger";

export function createApp(
  extractHandler: ExtractPageHandler,
  previewHandler: PreviewPageHandler,
  healthHandler: HealthCheckHandler,
): Hono {
  const app = new Hono();

  // Health check
  app.get("/health", (c) => {
    const result = healthHandler.handle();
    const status = result.status === "ok" ? 200 : 503;
    return c.json(result, status);
  });

  // Extract endpoint
  app.post("/extract", async (c) => {
    const reqId = createRequestId();
    const timer = createTimer();
    const body = await c.req.json<ExtractPageRequest>();

    if (!body.url) {
      log("http", "extract request missing url", { reqId });
      return c.json({ error: "url is required" }, 400);
    }

    const sectionsCount = body.sections?.length ?? 0;
    log("http", "extract request received", {
      reqId,
      url: body.url,
      hasSelector: !!(body.selector || body.selector_xpath),
      sections: sectionsCount,
      blockAdsCookies: body.block_ads_cookies ?? false,
    });

    try {
      const result = await extractHandler.handle(body);
      log("http", "extract request completed", {
        reqId,
        url: body.url,
        elapsed: timer.elapsed(),
        titleLength: result.title.length,
        htmlLength: result.html.length,
        textLength: result.text.length,
        screenshotSize: result.screenshot_base64.length,
        selectorMatched: result.selector_matched,
        sectionsExtracted: result.sections?.length ?? 0,
      });
      return c.json(result);
    } catch (err) {
      if (err instanceof ScraperError) {
        logError("http", `extract failed [${err.code}]`, err, { reqId, url: body.url, elapsed: timer.elapsed() });
        return c.json({ error: err.message, code: err.code }, err.statusCode as any);
      }
      logError("http", "extract unexpected error", err, { reqId, url: body.url, elapsed: timer.elapsed() });
      return c.json({ error: "Internal server error" }, 500);
    }
  });

  // Preview endpoint — SSE stream or JSON based on Accept header
  app.post("/preview", async (c) => {
    const reqId = createRequestId();
    const timer = createTimer();
    const body = await c.req.json<PreviewPageRequest>();

    if (!body.url) {
      log("http", "preview request missing url", { reqId });
      return c.json({ error: "url is required" }, 400);
    }

    const accept = c.req.header("Accept") || "";
    const wantsJSON = accept.includes("application/json");
    const mode = wantsJSON ? "json" : "sse";

    log("http", "preview request received", {
      reqId,
      url: body.url,
      mode,
      sectionDepth: body.section_depth,
      blockAdsCookies: body.block_ads_cookies ?? false,
    });

    // JSON mode — backward compat with Preview() Go method
    if (wantsJSON) {
      try {
        const result = await previewHandler.handle(body);
        log("http", "preview request completed", {
          reqId,
          url: body.url,
          mode,
          elapsed: timer.elapsed(),
          elements: result.elements.length,
          pageHeight: result.page_height,
        });
        return c.json(result);
      } catch (err) {
        if (err instanceof ScraperError) {
          logError("http", `preview failed [${err.code}]`, err, { reqId, url: body.url, elapsed: timer.elapsed() });
          return c.json({ error: err.message, code: err.code }, err.statusCode as any);
        }
        logError("http", "preview unexpected error", err, { reqId, url: body.url, elapsed: timer.elapsed() });
        return c.json({ error: "Internal server error" }, 500);
      }
    }

    // SSE stream mode — for PreviewStream() Go method
    return streamSSE(c, async (stream) => {
      try {
        const result = await previewHandler.handle(body, (step, totalSteps, message) => {
          log("http", `preview sse progress ${step}/${totalSteps}: ${message}`, { reqId });
          stream.writeSSE({
            event: "progress",
            data: JSON.stringify({ step, total_steps: totalSteps, message }),
          });
        });

        log("http", "preview sse completed", {
          reqId,
          url: body.url,
          elapsed: timer.elapsed(),
          elements: result.elements.length,
          pageHeight: result.page_height,
        });

        await stream.writeSSE({
          event: "result",
          data: JSON.stringify(result),
        });
      } catch (err) {
        const message =
          err instanceof Error ? err.message : "Unknown error";
        logError("http", "preview sse error", err, { reqId, url: body.url, elapsed: timer.elapsed() });
        await stream.writeSSE({
          event: "error",
          data: JSON.stringify({ error: message }),
        });
      }
    });
  });

  return app;
}
