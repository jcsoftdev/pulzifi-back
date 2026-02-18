import { Hono } from 'hono';
import { chromium, Browser, BrowserContext, Page } from 'playwright';

const app = new Hono();
const port = Number(process.env.PORT || 3000);

let browser: Browser;

async function initBrowser(): Promise<void> {
  try {
    browser = await chromium.launch({
      headless: true,
      args: [
        '--no-sandbox',
        '--disable-setuid-sandbox',
        '--disable-dev-shm-usage',
        '--disable-gpu'
      ]
    });
    console.log('Browser initialized successfully');
  } catch (err) {
    console.error('Failed to initialize browser:', err);
    process.exit(1);
  }
}

app.post('/extract', async (c) => {
  const body = await c.req.json().catch(() => null);
  const url = typeof body === 'object' && body !== null && 'url' in body ? (body as { url?: string }).url : undefined;
  if (!url) {
    return c.json({ error: 'URL is required' }, 400);
  }

  const startedAt = Date.now();
  let context: BrowserContext | undefined;
  let page: Page | undefined;
  try {
    context = await browser.newContext({
      viewport: { width: 1366, height: 768 },
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36'
    });
    page = await context.newPage();

    console.log(`Navigating to ${url}`);
    await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });

    // Wait for stability similar to original Go code
    await page.waitForTimeout(1000);

    const pageMetrics = await page.evaluate(() => ({
      scrollWidth: document.documentElement.scrollWidth,
      scrollHeight: document.documentElement.scrollHeight,
      clientWidth: document.documentElement.clientWidth,
      clientHeight: document.documentElement.clientHeight
    }));

    const title = await page.title();
    const content = await page.content(); // HTML
    
    // Get visible text
    const text = await page.evaluate((): string => {
        return document.body.innerText;
    });

    // Take full page screenshot
    const screenshot = await page.screenshot({
      fullPage: true,
      type: 'png',
      animations: 'disabled'
    });

    const durationMs = Date.now() - startedAt;
    console.log(
      `Extraction completed for ${url} in ${durationMs}ms (title_len=${title.length}, html_len=${content?.length || 0}, text_len=${text.length}, screenshot_bytes=${screenshot.length}, page_scroll=${pageMetrics.scrollWidth}x${pageMetrics.scrollHeight}, page_client=${pageMetrics.clientWidth}x${pageMetrics.clientHeight})`
    );

    await page.close();
    await context.close();

    return c.json({
      title,
      html: content || '',
      text: text,
      screenshot_base64: screenshot.toString('base64')
    });
  } catch (error) {
    console.error(`Extraction error for ${url}:`, error);
    if (page) await page.close().catch(() => {});
    if (context) await context.close().catch(() => {});
    return c.json({ error: error instanceof Error ? error.message : 'Unknown error' }, 500);
  }
});

app.get('/health', (c) => {
  return c.json({ status: 'ok', browser_connected: !!browser && browser.isConnected() });
});

await initBrowser();
console.log(`Extractor service running on port ${port}`);

export default {
  port,
  fetch: app.fetch
};
