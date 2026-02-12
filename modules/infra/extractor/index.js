const express = require('express');
const { chromium } = require('playwright');
const bodyParser = require('body-parser');

const app = express();
const port = process.env.PORT || 3000;

// Increase limit for large screenshots/HTML
app.use(bodyParser.json({ limit: '50mb' }));

let browser;

async function initBrowser() {
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

app.post('/extract', async (req, res) => {
  const { url } = req.body;
  if (!url) {
    return res.status(400).json({ error: 'URL is required' });
  }

  let context;
  let page;
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

    const title = await page.title();
    const content = await page.content(); // HTML
    
    // Get visible text
    const text = await page.evaluate(() => {
        return document.body.innerText;
    });

    const screenshot = await page.screenshot({ fullPage: false, type: 'png' });

    await page.close();
    await context.close();

    res.json({
      title,
      html: content,
      text: text,
      screenshot_base64: screenshot.toString('base64')
    });
  } catch (error) {
    console.error(`Extraction error for ${url}:`, error);
    if (page) await page.close().catch(() => {});
    if (context) await context.close().catch(() => {});
    res.status(500).json({ error: error.message });
  }
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok', browser_connected: !!browser && browser.isConnected() });
});

initBrowser().then(() => {
  app.listen(port, () => {
    console.log(`Extractor service running on port ${port}`);
  });
});
