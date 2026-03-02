import { Hono } from 'hono';
import { chromium, Browser, BrowserContext, Page } from 'playwright';

const app = new Hono();
const port = Number(process.env.PORT || 3000);

const STABILITY_MS = 500;  // no DOM mutations for this long = render complete
const MAX_RENDER_WAIT_MS = 15000; // safety cap for render detection

let browser: Browser;

async function initBrowser(): Promise<void> {
  try {
    browser = await chromium.launch({
      headless: true,
      args: [
        '--no-sandbox',
        '--disable-setuid-sandbox',
        '--disable-dev-shm-usage',
        '--disable-gpu',
        '--disable-blink-features=AutomationControlled',
        '--disable-infobars',
        '--window-size=1366,768',
        '--lang=en-US,en'
      ]
    });
    console.log('Browser initialized successfully');
  } catch (err) {
    console.error('Failed to initialize browser:', err);
    process.exit(1);
  }
}

const CHROME_USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36';

/**
 * Adds stealth init scripts to a browser context to avoid bot detection.
 * These run before any page JavaScript executes.
 */
async function addStealthScripts(context: BrowserContext): Promise<void> {
  await context.addInitScript(() => {
    try {
      // Hide navigator.webdriver (primary bot detection signal)
      Object.defineProperty(navigator, 'webdriver', {
        get: () => undefined,
      });
    } catch (_) {}

    try {
      // Fake Chrome runtime object (missing in headless)
      if (!(window as any).chrome) {
        (window as any).chrome = {};
      }
      if (!(window as any).chrome.runtime) {
        (window as any).chrome.runtime = {
          connect: () => {},
          sendMessage: () => {},
        };
      }
    } catch (_) {}

    try {
      // Spoof navigator.plugins length (headless has 0)
      // Use a simple array-like object — PluginArray/Plugin constructors
      // are restricted in some Chromium builds
      Object.defineProperty(navigator, 'plugins', {
        get: () => {
          return [
            { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format', length: 0 },
            { name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: '', length: 0 },
            { name: 'Native Client', filename: 'internal-nacl-plugin', description: '', length: 0 },
          ] as any;
        },
      });
    } catch (_) {}

    try {
      // Consistent navigator.languages and platform
      Object.defineProperty(navigator, 'languages', {
        get: () => ['en-US', 'en'],
      });
      Object.defineProperty(navigator, 'platform', {
        get: () => 'Win32',
      });
    } catch (_) {}

    try {
      // Realistic hardware values
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        get: () => 4,
      });
      Object.defineProperty(navigator, 'deviceMemory', {
        get: () => 8,
      });
    } catch (_) {}

    try {
      // Fix Permissions.query for notifications
      const originalQuery = window.Permissions?.prototype?.query;
      if (originalQuery) {
        window.Permissions.prototype.query = function (params: any) {
          if (params.name === 'notifications') {
            return Promise.resolve({ state: 'prompt', onchange: null } as PermissionStatus);
          }
          return originalQuery.call(this, params);
        };
      }
    } catch (_) {}

    try {
      // WebGL vendor/renderer spoofing (hide SwiftShader)
      const getParameterProto = WebGLRenderingContext.prototype.getParameter;
      WebGLRenderingContext.prototype.getParameter = function (param: number) {
        const UNMASKED_VENDOR_WEBGL = 0x9245;
        const UNMASKED_RENDERER_WEBGL = 0x9246;
        if (param === UNMASKED_VENDOR_WEBGL) return 'Google Inc. (NVIDIA)';
        if (param === UNMASKED_RENDERER_WEBGL) return 'ANGLE (NVIDIA, NVIDIA GeForce GTX 1650 Direct3D11 vs_5_0 ps_5_0, D3D11)';
        return getParameterProto.call(this, param);
      };

      // Also patch WebGL2
      if (typeof WebGL2RenderingContext !== 'undefined') {
        const getParameterProto2 = WebGL2RenderingContext.prototype.getParameter;
        WebGL2RenderingContext.prototype.getParameter = function (param: number) {
          const UNMASKED_VENDOR_WEBGL = 0x9245;
          const UNMASKED_RENDERER_WEBGL = 0x9246;
          if (param === UNMASKED_VENDOR_WEBGL) return 'Google Inc. (NVIDIA)';
          if (param === UNMASKED_RENDERER_WEBGL) return 'ANGLE (NVIDIA, NVIDIA GeForce GTX 1650 Direct3D11 vs_5_0 ps_5_0, D3D11)';
          return getParameterProto2.call(this, param);
        };
      }
    } catch (_) {}
  });
}

type ChallengeType = 'turnstile' | 'managed' | 'block' | 'none';

/**
 * Detects if the page is showing a Cloudflare challenge.
 */
async function detectCloudflareChallenge(page: Page): Promise<ChallengeType> {
  return page.evaluate((): string => {
    const title = document.title.toLowerCase();
    const bodyText = document.body?.innerText || '';

    // WAF block page
    if (title.includes('attention required') || title.includes('access denied') || bodyText.includes('Sorry, you have been blocked')) {
      return 'block';
    }

    // Turnstile widget
    const hasTurnstile = !!document.querySelector('.cf-turnstile') ||
      !!document.querySelector('iframe[src*="challenges.cloudflare.com"]');

    // Managed challenge
    const hasManagedChallenge = !!document.querySelector('#challenge-running') ||
      !!document.querySelector('#challenge-stage');

    const isChallengePage = title.includes('just a moment') ||
      bodyText.includes('Verify you are human') ||
      bodyText.includes('Checking your browser');

    if (hasTurnstile && isChallengePage) return 'turnstile';
    if ((hasManagedChallenge || isChallengePage) && !hasTurnstile) return 'managed';

    return 'none';
  }) as Promise<ChallengeType>;
}

/**
 * Attempts to handle a Cloudflare challenge by waiting for auto-resolve
 * or clicking the Turnstile checkbox if present.
 */
async function handleCloudflareChallenge(page: Page): Promise<void> {
  const challengeType = await detectCloudflareChallenge(page);

  if (challengeType === 'none') return;

  console.log(`Cloudflare challenge detected: ${challengeType}`);

  if (challengeType === 'block') {
    throw new CloudflareError('Cloudflare WAF block — site has blocked this request');
  }

  // For turnstile challenges, try clicking the checkbox inside the iframe
  if (challengeType === 'turnstile') {
    try {
      const cfFrame = page.frames().find(f => f.url().includes('challenges.cloudflare.com'));
      if (cfFrame) {
        const checkbox = await cfFrame.$('input[type="checkbox"], .cb-lb');
        if (checkbox) {
          console.log('Clicking Turnstile checkbox');
          await checkbox.click();
        }
      }
    } catch (e) {
      console.log('Could not click Turnstile checkbox, waiting for auto-resolve');
    }
  }

  // Poll for challenge resolution
  const CF_POLL_INTERVAL_MS = 2000;
  const CF_MAX_WAIT_MS = 20000;
  const deadline = Date.now() + CF_MAX_WAIT_MS;

  while (Date.now() < deadline) {
    await new Promise(r => setTimeout(r, CF_POLL_INTERVAL_MS));

    const currentChallenge = await detectCloudflareChallenge(page);
    if (currentChallenge === 'none') {
      console.log('Cloudflare challenge resolved');
      return;
    }

    if (currentChallenge === 'block') {
      throw new CloudflareError('Cloudflare WAF block — site has blocked this request');
    }
  }

  throw new CloudflareError('CLOUDFLARE_CHALLENGE_TIMEOUT — challenge did not resolve within 20s');
}

class CloudflareError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'CloudflareError';
  }
}

/**
 * Scrolls through the page in viewport-sized steps to trigger
 * IntersectionObserver-based lazy loading, then scrolls back to top.
 * Uses 500ms pause per step to give lazy loaders time to react.
 */
async function scrollFullPage(page: Page): Promise<void> {
  await page.evaluate(async () => {
    const delay = (ms: number) => new Promise(r => setTimeout(r, ms));
    const step = Math.floor(window.innerHeight * 0.7);
    let currentPos = 0;
    let maxScroll = document.body.scrollHeight;

    while (currentPos < maxScroll) {
      window.scrollTo(0, currentPos);
      await delay(500);
      currentPos += step;
      const newMax = document.body.scrollHeight;
      if (newMax > maxScroll + step * 5) break; // cap infinite scroll growth
      maxScroll = newMax;
    }
    // Hit the very bottom
    window.scrollTo(0, document.body.scrollHeight);
    await delay(800);
    // Back to top
    window.scrollTo(0, 0);
    await delay(300);
  });
}

/**
 * Forces ALL lazy-loaded resources to load:
 * - <img> with data-src, loading="lazy", placeholder src (data:image/gif, data:image/png)
 * - <picture>/<source> with data-srcset
 * - <video> poster images
 * - <iframe> with data-src
 * - Elements with data-bg / data-background-image (CSS bg lazy pattern)
 * - Removes common lazy CSS classes (lazy, lazyload, lazyloading)
 */
async function forceLoadAllResources(page: Page): Promise<void> {
  await page.evaluate(() => {
    const isPlaceholder = (src: string | null | undefined): boolean => {
      if (!src) return true;
      return src === 'about:blank' ||
        src.startsWith('data:image/svg') ||
        src.startsWith('data:image/gif') ||
        src.startsWith('data:image/png') ||
        src.startsWith('data:image/webp') ||
        src === '';
    };

    // <img> tags
    for (const img of document.querySelectorAll('img')) {
      img.loading = 'eager';
      img.classList.remove('lazy', 'lazyload', 'lazyloading');
      img.removeAttribute('data-ll-status');

      for (const attr of ['data-src', 'data-lazy-src', 'data-original', 'data-lazy']) {
        const val = img.getAttribute(attr);
        if (val && isPlaceholder(img.getAttribute('src'))) {
          img.src = val;
          break;
        }
      }
      const lazySrcset = img.getAttribute('data-srcset');
      if (lazySrcset) {
        img.srcset = lazySrcset;
      }
      // Handle data-sizes
      const lazySizes = img.getAttribute('data-sizes');
      if (lazySizes && !img.sizes) {
        img.sizes = lazySizes;
      }
    }

    // <picture> / <source> elements
    for (const source of document.querySelectorAll('picture source, video source')) {
      const dataSrcset = source.getAttribute('data-srcset');
      if (dataSrcset) {
        source.setAttribute('srcset', dataSrcset);
      }
      const dataSrc = source.getAttribute('data-src');
      if (dataSrc && !source.getAttribute('src')) {
        source.setAttribute('src', dataSrc);
      }
    }

    // <video> poster
    for (const video of document.querySelectorAll('video')) {
      const dataPoster = video.getAttribute('data-poster');
      if (dataPoster && !video.poster) {
        video.poster = dataPoster;
      }
    }

    // <iframe> with data-src
    for (const iframe of document.querySelectorAll('iframe')) {
      const dataSrc = iframe.getAttribute('data-src');
      if (dataSrc && isPlaceholder(iframe.getAttribute('src'))) {
        iframe.src = dataSrc;
      }
    }

    // CSS background images via data attributes (common lazy pattern)
    for (const el of document.querySelectorAll<HTMLElement>('[data-bg], [data-background-image], [data-bg-src]')) {
      const bg = el.getAttribute('data-bg') || el.getAttribute('data-background-image') || el.getAttribute('data-bg-src');
      if (bg && !el.style.backgroundImage) {
        el.style.backgroundImage = `url('${bg}')`;
      }
    }
  });
}

/**
 * Waits for all pending resources: images, fonts, and CSS background images.
 */
async function waitForAllResources(page: Page): Promise<void> {
  // Wait for fonts
  await page.evaluate(() => {
    return Promise.race([
      (document as any).fonts?.ready || Promise.resolve(),
      new Promise(r => setTimeout(r, 5000)),
    ]);
  });

  // Wait for pending <img> elements + preload CSS background image URLs
  await page.evaluate(() => {
    return new Promise<void>((resolve) => {
      const timeout = setTimeout(resolve, 15000);
      const promises: Promise<void>[] = [];

      // <img> elements
      for (const img of document.querySelectorAll('img')) {
        if (img.src && !img.complete) {
          promises.push(new Promise<void>(r => {
            img.addEventListener('load', () => r(), { once: true });
            img.addEventListener('error', () => r(), { once: true });
          }));
        }
      }

      // CSS background images — collect unique URLs from visible elements
      const bgUrls = new Set<string>();
      const visibleElements = document.querySelectorAll<HTMLElement>(
        'header, nav, main, section, article, footer, .hero, .banner, .slider, .carousel, [class*="hero"], [class*="banner"], [class*="slider"], [class*="carousel"], [style*="background"]'
      );
      for (const el of visibleElements) {
        try {
          const bg = getComputedStyle(el).backgroundImage;
          if (bg && bg !== 'none') {
            const matches = bg.matchAll(/url\(["']?(.*?)["']?\)/g);
            for (const m of matches) {
              if (m[1] && !m[1].startsWith('data:')) {
                bgUrls.add(m[1]);
              }
            }
          }
        } catch (_) {}
      }

      for (const bgUrl of bgUrls) {
        promises.push(new Promise<void>(r => {
          const img = new Image();
          img.onload = () => r();
          img.onerror = () => r();
          img.src = bgUrl;
        }));
      }

      if (promises.length === 0) {
        clearTimeout(timeout);
        resolve();
        return;
      }

      Promise.all(promises).then(() => {
        clearTimeout(timeout);
        resolve();
      });
    });
  });

  // Ensure the browser has painted after all resources loaded
  await page.evaluate(() => {
    return new Promise<void>(r => requestAnimationFrame(() => requestAnimationFrame(() => r())));
  });
}

/**
 * Waits until the page DOM stops mutating, which signals that JavaScript
 * rendering (React, Vue, Angular, vanilla JS) has completed.
 *
 * Uses MutationObserver: every DOM change resets a stability timer.
 * Once no mutations occur for STABILITY_MS, the page is considered rendered.
 * A hard cap of MAX_RENDER_WAIT_MS prevents infinite waits on pages with
 * constant DOM activity (animations, tickers, live feeds).
 */
async function waitForRenderComplete(page: Page): Promise<void> {
  await page.evaluate(({ stabilityMs, maxWaitMs }) => {
    return new Promise<void>((resolve) => {
      let timer: ReturnType<typeof setTimeout>;

      const done = () => {
        clearTimeout(timer);
        clearTimeout(maxTimer);
        observer.disconnect();
        resolve();
      };

      const maxTimer = setTimeout(done, maxWaitMs);

      const observer = new MutationObserver(() => {
        clearTimeout(timer);
        timer = setTimeout(done, stabilityMs);
      });

      observer.observe(document.documentElement, {
        childList: true,
        subtree: true,
        attributes: true,
        characterData: true,
      });

      // If the DOM is already stable (static page / SSR), resolve after stabilityMs
      timer = setTimeout(done, stabilityMs);
    });
  }, { stabilityMs: STABILITY_MS, maxWaitMs: MAX_RENDER_WAIT_MS });
}

/**
 * URL patterns for known ad networks, trackers, and analytics services.
 * Used to block requests when block_ads_cookies is enabled.
 */
const AD_URL_PATTERNS: RegExp[] = [
  // Google Ads & DoubleClick
  /googlesyndication\.com/,
  /doubleclick\.net/,
  /googleadservices\.com/,
  /google-analytics\.com/,
  /googletagmanager\.com/,
  /pagead2\.googlesyndication\.com/,
  /adservice\.google\./,
  // Facebook / Meta
  /facebook\.net\/.*\/fbevents/,
  /connect\.facebook\.net/,
  /facebook\.com\/tr/,
  // Common ad networks
  /adnxs\.com/,
  /adsrvr\.org/,
  /amazon-adsystem\.com/,
  /criteo\.com/,
  /criteo\.net/,
  /outbrain\.com/,
  /taboola\.com/,
  /moatads\.com/,
  /rubiconproject\.com/,
  /pubmatic\.com/,
  /openx\.net/,
  /casalemedia\.com/,
  /sharethis\.com/,
  /addthis\.com/,
  /popads\.net/,
  /adcolony\.com/,
  /serving-sys\.com/,
  // Analytics & trackers
  /hotjar\.com/,
  /fullstory\.com/,
  /mixpanel\.com/,
  /segment\.com\/analytics/,
  /hubspot\.com\/.*tracking/,
  /bat\.bing\.com/,
  /clarity\.ms/,
  // Generic ad paths
  /\/ads\//,
  /\/adsbygoogle/,
];

/**
 * Blocks ad/tracker network requests on a browser context using route interception.
 */
async function setupAdBlocking(context: BrowserContext): Promise<void> {
  await context.route('**/*', (route) => {
    const url = route.request().url();

    // Check URL against ad patterns
    if (AD_URL_PATTERNS.some(pattern => pattern.test(url))) {
      return route.abort();
    }

    return route.continue();
  });
}

/**
 * Injects a persistent cookie/consent element remover as an init script.
 * Runs before any page JS. Uses a MutationObserver to continuously find
 * and remove from the DOM any element that has "cookie" or "consent" in
 * its classes, id, or — for leaf nodes — text content.
 */
async function setupCookieBannerBlocker(context: BrowserContext): Promise<void> {
  await context.addInitScript(() => {
    const SEARCH_TERMS = ['cookie', 'consent', 'gdpr', 'rgpd'];

    /** Walk up from el to find the topmost ancestor that still matches cookie criteria */
    function findBannerRoot(el: HTMLElement): HTMLElement {
      let root = el;
      let current = el.parentElement;
      while (current && current !== document.body && current !== document.documentElement) {
        const cls = (current.className || '').toString().toLowerCase();
        const id = (current.id || '').toLowerCase();
        if (SEARCH_TERMS.some(t => cls.includes(t) || id.includes(t))) {
          root = current;
        }
        current = current.parentElement;
      }
      return root;
    }

    function removeCookieElements(): void {
      if (!document.body) return;

      const toRemove = new Set<HTMLElement>();
      const allElements = document.body.querySelectorAll<HTMLElement>('*');

      for (const el of allElements) {
        if (el.getAttribute('data-cookie-removed') === '1') continue;

        // 1. Check classes for cookie/consent terms
        const hasCookieClass = Array.from(el.classList).some(cls =>
          SEARCH_TERMS.some(t => cls.toLowerCase().includes(t))
        );

        // 2. Check id
        const id = (el.id || '').toLowerCase();
        const hasCookieId = SEARCH_TERMS.some(t => id.includes(t));

        // 3. Check aria-label
        const aria = (el.getAttribute('aria-label') || '').toLowerCase();
        const hasCookieAria = SEARCH_TERMS.some(t => aria.includes(t));

        // 4. For leaf nodes (no element children), check text content
        const hasCookieText = el.children.length === 0 &&
          SEARCH_TERMS.some(t => (el.textContent || '').toLowerCase().includes(t));

        if (hasCookieClass || hasCookieId || hasCookieAria) {
          // Walk up to find the root banner container and remove the whole thing
          toRemove.add(findBannerRoot(el));
        } else if (hasCookieText) {
          // For text-only matches on leaf nodes, walk up to check if parent is a banner
          const parent = el.parentElement;
          if (parent) {
            const style = getComputedStyle(parent);
            const isOverlay = style.position === 'fixed' || style.position === 'sticky';
            const isDialog = parent.getAttribute('role') === 'dialog' || parent.getAttribute('role') === 'alertdialog';
            if (isOverlay || isDialog) {
              toRemove.add(findBannerRoot(parent));
            }
          }
        }
      }

      for (const el of toRemove) {
        try {
          el.remove();
        } catch (_) {
          // Fallback: hide if removal fails
          el.style.setProperty('display', 'none', 'important');
          el.setAttribute('data-cookie-removed', '1');
        }
      }

      // Restore body scroll that cookie banners may have locked
      if (document.body) {
        document.body.style.removeProperty('overflow');
        document.body.classList.remove('no-scroll', 'modal-open', 'cookie-modal-open');
      }
      document.documentElement.style.removeProperty('overflow');
    }

    // Run on DOMContentLoaded
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', removeCookieElements);
    } else {
      removeCookieElements();
    }

    // MutationObserver: remove cookie elements as they're added to the DOM
    const observer = new MutationObserver(removeCookieElements);

    const startObserving = () => {
      observer.observe(document.documentElement, { childList: true, subtree: true });
    };

    if (document.documentElement) {
      startObserving();
    } else {
      document.addEventListener('DOMContentLoaded', startObserving);
    }

    // Periodic sweep for the first 15 seconds (catches late JS-injected banners)
    let ticks = 0;
    const interval = setInterval(() => {
      removeCookieElements();
      if (++ticks >= 30) clearInterval(interval);
    }, 500);
  });
}

/**
 * Final aggressive pass: remove ALL cookie-related elements from the DOM
 * right before taking the screenshot.
 */
async function finalCookieBannerCleanup(page: Page): Promise<void> {
  await page.evaluate(() => {
    const SEARCH_TERMS = ['cookie', 'consent', 'gdpr', 'rgpd'];

    function findBannerRoot(el: HTMLElement): HTMLElement {
      let root = el;
      let current = el.parentElement;
      while (current && current !== document.body && current !== document.documentElement) {
        const cls = (current.className || '').toString().toLowerCase();
        const id = (current.id || '').toLowerCase();
        if (SEARCH_TERMS.some(t => cls.includes(t) || id.includes(t))) {
          root = current;
        }
        current = current.parentElement;
      }
      return root;
    }

    const toRemove = new Set<HTMLElement>();
    const allElements = document.body.querySelectorAll<HTMLElement>('*');

    for (const el of allElements) {
      // Check classes
      const hasCookieClass = Array.from(el.classList).some(cls =>
        SEARCH_TERMS.some(t => cls.toLowerCase().includes(t))
      );

      // Check id
      const hasCookieId = SEARCH_TERMS.some(t => (el.id || '').toLowerCase().includes(t));

      // Check aria-label
      const hasCookieAria = SEARCH_TERMS.some(t => (el.getAttribute('aria-label') || '').toLowerCase().includes(t));

      // Leaf node text check
      const hasCookieText = el.children.length === 0 &&
        SEARCH_TERMS.some(t => (el.textContent || '').toLowerCase().includes(t));

      if (hasCookieClass || hasCookieId || hasCookieAria) {
        toRemove.add(findBannerRoot(el));
      } else if (hasCookieText) {
        const parent = el.parentElement;
        if (parent) {
          const style = getComputedStyle(parent);
          if (style.position === 'fixed' || style.position === 'sticky' ||
            parent.getAttribute('role') === 'dialog' || parent.getAttribute('role') === 'alertdialog') {
            toRemove.add(findBannerRoot(parent));
          }
        }
      }
    }

    for (const el of toRemove) {
      try { el.remove(); } catch (_) {}
    }

    // Restore body scroll
    document.body.style.removeProperty('overflow');
    document.documentElement.style.removeProperty('overflow');
    document.body.classList.remove('no-scroll', 'modal-open', 'cookie-modal-open');
  });
}

app.post('/extract', async (c) => {
  const body = await c.req.json().catch(() => null);
  const url = typeof body === 'object' && body !== null && 'url' in body ? (body as { url?: string }).url : undefined;
  if (!url) {
    return c.json({ error: 'URL is required' }, 400);
  }

  const blockAdsCookies = typeof body === 'object' && body !== null && 'block_ads_cookies' in body
    ? Boolean((body as { block_ads_cookies?: boolean }).block_ads_cookies)
    : false;

  const startedAt = Date.now();
  let context: BrowserContext | undefined;
  let page: Page | undefined;
  try {
    context = await browser.newContext({
      viewport: { width: 1366, height: 768 },
      userAgent: CHROME_USER_AGENT,
      locale: 'en-US',
      timezoneId: 'America/New_York',
      javaScriptEnabled: true,
      extraHTTPHeaders: {
        'Accept-Language': 'en-US,en;q=0.9',
        'sec-ch-ua': '"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"',
        'sec-ch-ua-mobile': '?0',
        'sec-ch-ua-platform': '"Windows"',
      },
    });

    await addStealthScripts(context);

    // Block ad requests and cookie banners when enabled
    if (blockAdsCookies) {
      await setupAdBlocking(context);
      await setupCookieBannerBlocker(context);
      console.log(`Ad blocking + cookie banner blocking enabled for ${url}`);
    }

    page = await context.newPage();

    console.log(`Navigating to ${url}`);

    // Use 'domcontentloaded' for fast initial load, then wait for rendering
    // and network to settle separately. 'load' and 'networkidle' can stall
    // on slow third-party resources causing timeouts.
    await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });

    // Handle Cloudflare challenges before proceeding with extraction
    await handleCloudflareChallenge(page);

    // Wait for 'load' event (CSS, scripts, above-fold images) — cap at 15s
    await page.waitForLoadState('load').catch(() => {
      console.log(`Load event timeout for ${url}, proceeding`);
    });

    // Wait for network to settle (third-party scripts, API calls)
    await page.waitForLoadState('networkidle').catch(() => {
      console.log(`Initial networkidle timeout for ${url}, proceeding`);
    });

    // Wait for JS frameworks to finish rendering
    await waitForRenderComplete(page);

    // Scroll full page to trigger lazy-loaded content (IntersectionObserver)
    await scrollFullPage(page);

    // Force lazy resources: img data-src, picture/source, video poster, bg images
    await forceLoadAllResources(page);

    // Wait for network to settle AGAIN after scroll triggered new requests
    await page.waitForLoadState('networkidle').catch(() => {
      console.log(`Post-scroll networkidle timeout for ${url}, proceeding`);
    });

    // Wait for all images, fonts, CSS background images + ensure paint
    await waitForAllResources(page);

    // Final settle for any CSS transitions / late paints
    await page.waitForTimeout(1000);

    // Final cookie banner cleanup right before screenshot
    if (blockAdsCookies) {
      await finalCookieBannerCleanup(page);
      await page.waitForTimeout(300);
    }

    const pageMetrics = await page.evaluate(() => ({
      scrollWidth: document.documentElement.scrollWidth,
      scrollHeight: document.documentElement.scrollHeight,
      clientWidth: document.documentElement.clientWidth,
      clientHeight: document.documentElement.clientHeight
    }));

    const title = await page.title();
    const content = await page.content();

    const text = await page.evaluate((): string => {
      return document.body.innerText;
    });

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
    if (page) await page.close().catch(() => { });
    if (context) await context.close().catch(() => { });

    const statusCode = error instanceof CloudflareError ? 403 : 500;
    return c.json({ error: error instanceof Error ? error.message : 'Unknown error' }, statusCode);
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
