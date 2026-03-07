import { serve } from "bun";
import { createApp } from "./src/infrastructure/http/app";
import { PatchrightBrowserService } from "./src/infrastructure/browser/patchright-browser";
import { SharpImageProcessor } from "./src/infrastructure/image/sharp-processor";
import { ExtractPageHandler } from "./src/application/extract-page/handler";
import { PreviewPageHandler } from "./src/application/preview-page/handler";
import { HealthCheckHandler } from "./src/application/health-check/handler";

const PORT = parseInt(process.env.PORT || "3000", 10);

async function main() {
  console.log("[scraper] Starting...");

  // Initialize infrastructure
  const imageProcessor = new SharpImageProcessor();
  const browserService = new PatchrightBrowserService(imageProcessor);
  await browserService.launch();

  // Initialize application handlers
  const extractHandler = new ExtractPageHandler(browserService);
  const previewHandler = new PreviewPageHandler(browserService);
  const healthHandler = new HealthCheckHandler(browserService);

  // Create HTTP app
  const app = createApp(extractHandler, previewHandler, healthHandler);

  // Start server
  const server = serve({
    fetch: app.fetch,
    port: PORT,
  });

  console.log(`[scraper] Listening on http://0.0.0.0:${PORT}`);

  // Graceful shutdown
  const shutdown = async (signal: string) => {
    console.log(`[scraper] Received ${signal}, shutting down...`);
    server.stop();
    await browserService.shutdown();
    process.exit(0);
  };

  process.on("SIGINT", () => shutdown("SIGINT"));
  process.on("SIGTERM", () => shutdown("SIGTERM"));
}

main().catch((err) => {
  console.error("[scraper] Fatal error:", err);
  process.exit(1);
});
