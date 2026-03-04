import { Hono } from 'hono';
import { chromium, Browser, BrowserContext, Page } from 'rebrowser-playwright';

const app = new Hono();
const port = Number(process.env.PORT || 3000);

const STABILITY_MS = 500;  // no DOM mutations for this long = render complete
const MAX_RENDER_WAIT_MS = 15000; // safety cap for render detection

let browser: Browser;

const CHROME_USER_AGENT = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36';

async function initBrowser(): Promise<void> {
  try {
    browser = await chromium.launch({
      headless: true,
      args: [
        '--no-sandbox',
        '--disable-setuid-sandbox',
        '--disable-dev-shm-usage',
        '--disable-infobars',
        '--window-size=1920,1080',
        '--lang=en-US,en',
        '--disable-background-networking',
        '--disable-component-update',
        '--disable-domain-reliability',
        '--disable-sync',
        '--metrics-recording-only',
        '--no-first-run',
        '--no-default-browser-check',
      ],
    });
    console.log('Chromium browser initialized successfully (rebrowser-playwright)');
  } catch (err) {
    console.error('Failed to initialize browser:', err);
    process.exit(1);
  }
}

/**
 * Adds stealth init scripts to hide automation signals.
 * rebrowser-playwright handles CDP leaks; these cover the remaining JS-level signals.
 */
async function addStealthScripts(context: BrowserContext): Promise<void> {
  await context.addInitScript(() => {
    // 1. Hide webdriver flag
    try {
      Object.defineProperty(navigator, 'webdriver', {
        get: () => undefined,
      });
    } catch (_) {}

    // 2. Fake chrome runtime (headless Chrome lacks this)
    try {
      if (!(window as any).chrome) {
        (window as any).chrome = {};
      }
      if (!(window as any).chrome.runtime) {
        (window as any).chrome.runtime = {
          connect: () => {},
          sendMessage: () => {},
          id: undefined,
        };
      }
      if (!(window as any).chrome.loadTimes) {
        (window as any).chrome.loadTimes = () => ({});
      }
      if (!(window as any).chrome.csi) {
        (window as any).chrome.csi = () => ({});
      }
      if (!(window as any).chrome.app) {
        (window as any).chrome.app = {
          isInstalled: false,
          InstallState: { DISABLED: 'disabled', INSTALLED: 'installed', NOT_INSTALLED: 'not_installed' },
          RunningState: { CANNOT_RUN: 'cannot_run', READY_TO_RUN: 'ready_to_run', RUNNING: 'running' },
        };
      }
    } catch (_) {}

    // 3. Permissions API — return realistic results
    try {
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

    // 4. Fake plugins (headless Chrome has 0 plugins — dead giveaway)
    try {
      Object.defineProperty(navigator, 'plugins', {
        get: () => {
          const fakePlugins = [
            { name: 'Chrome PDF Plugin', description: 'Portable Document Format', filename: 'internal-pdf-viewer', length: 1 },
            { name: 'Chrome PDF Viewer', description: '', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', length: 1 },
            { name: 'Native Client', description: '', filename: 'internal-nacl-plugin', length: 2 },
          ];
          const arr = Object.create(PluginArray.prototype);
          fakePlugins.forEach((p, i) => {
            const plugin = Object.create(Plugin.prototype);
            Object.defineProperties(plugin, {
              name: { value: p.name, enumerable: true },
              description: { value: p.description, enumerable: true },
              filename: { value: p.filename, enumerable: true },
              length: { value: p.length, enumerable: true },
            });
            Object.defineProperty(arr, i, { value: plugin, enumerable: true });
          });
          Object.defineProperty(arr, 'length', { value: fakePlugins.length });
          arr.item = (i: number) => arr[i] || null;
          arr.namedItem = (name: string) => fakePlugins.find(p => p.name === name) ? arr[fakePlugins.findIndex(p => p.name === name)] : null;
          arr.refresh = () => {};
          return arr;
        },
      });
    } catch (_) {}

    // 5. Fake mimeTypes
    try {
      Object.defineProperty(navigator, 'mimeTypes', {
        get: () => {
          const fakeMimes = [
            { type: 'application/pdf', suffixes: 'pdf', description: 'Portable Document Format' },
            { type: 'application/x-google-chrome-pdf', suffixes: 'pdf', description: 'Portable Document Format' },
          ];
          const arr = Object.create(MimeTypeArray.prototype);
          fakeMimes.forEach((m, i) => {
            const mime = Object.create(MimeType.prototype);
            Object.defineProperties(mime, {
              type: { value: m.type, enumerable: true },
              suffixes: { value: m.suffixes, enumerable: true },
              description: { value: m.description, enumerable: true },
            });
            Object.defineProperty(arr, i, { value: mime, enumerable: true });
          });
          Object.defineProperty(arr, 'length', { value: fakeMimes.length });
          arr.item = (i: number) => arr[i] || null;
          arr.namedItem = (name: string) => fakeMimes.find(m => m.type === name) ? arr[fakeMimes.findIndex(m => m.type === name)] : null;
          return arr;
        },
      });
    } catch (_) {}

    // 6. Spoof languages
    try {
      Object.defineProperty(navigator, 'languages', {
        get: () => ['en-US', 'en'],
      });
    } catch (_) {}

    // 7. Spoof hardwareConcurrency (headless sometimes returns odd values)
    try {
      Object.defineProperty(navigator, 'hardwareConcurrency', {
        get: () => 8,
      });
    } catch (_) {}

    // 8. Spoof deviceMemory
    try {
      Object.defineProperty(navigator, 'deviceMemory', {
        get: () => 8,
      });
    } catch (_) {}

    // 9. WebGL vendor/renderer spoofing (headless leaks "Google SwiftShader")
    try {
      const getParameterOrig = WebGLRenderingContext.prototype.getParameter;
      WebGLRenderingContext.prototype.getParameter = function (param: number) {
        if (param === 37445) return 'Google Inc. (NVIDIA)';       // UNMASKED_VENDOR_WEBGL
        if (param === 37446) return 'ANGLE (NVIDIA, NVIDIA GeForce GTX 1650, OpenGL 4.5)'; // UNMASKED_RENDERER_WEBGL
        return getParameterOrig.call(this, param);
      };
      const getParameterOrig2 = WebGL2RenderingContext.prototype.getParameter;
      WebGL2RenderingContext.prototype.getParameter = function (param: number) {
        if (param === 37445) return 'Google Inc. (NVIDIA)';
        if (param === 37446) return 'ANGLE (NVIDIA, NVIDIA GeForce GTX 1650, OpenGL 4.5)';
        return getParameterOrig2.call(this, param);
      };
    } catch (_) {}

    // 10. Fix screen dimensions to match viewport
    try {
      Object.defineProperty(screen, 'width', { get: () => 1920 });
      Object.defineProperty(screen, 'height', { get: () => 1080 });
      Object.defineProperty(screen, 'availWidth', { get: () => 1920 });
      Object.defineProperty(screen, 'availHeight', { get: () => 1040 });
      Object.defineProperty(screen, 'colorDepth', { get: () => 24 });
      Object.defineProperty(screen, 'pixelDepth', { get: () => 24 });
    } catch (_) {}

    // 11. Hide Notification permission denial (headless denies by default)
    try {
      const origNotification = window.Notification;
      if (origNotification) {
        Object.defineProperty(origNotification, 'permission', {
          get: () => 'default',
        });
      }
    } catch (_) {}

    // 12. Prevent iframe contentWindow detection
    try {
      const origHTMLIFrameElement = HTMLIFrameElement.prototype;
      const origContentWindow = Object.getOwnPropertyDescriptor(origHTMLIFrameElement, 'contentWindow');
      if (origContentWindow) {
        Object.defineProperty(origHTMLIFrameElement, 'contentWindow', {
          get: function () {
            const win = origContentWindow.get?.call(this);
            if (!win) return win;
            // Patch the nested window to also hide webdriver
            try {
              Object.defineProperty(win.navigator, 'webdriver', { get: () => undefined });
            } catch (_) {}
            return win;
          },
        });
      }
    } catch (_) {}
  });
}

type ChallengeType = 'turnstile' | 'managed' | 'block' | 'none';

/**
 * Detects if the page is showing a Cloudflare challenge.
 */
async function detectCloudflareChallenge(page: Page): Promise<ChallengeType> {
  try {
    return await page.evaluate((): string => {
      const title = document.title.toLowerCase();
      const bodyText = document.body?.innerText || '';

      // WAF block page
      if (title.includes('attention required') || title.includes('access denied') || title.includes('blocked') || bodyText.includes('Sorry, you have been blocked')) {
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
  } catch (_) {
    // Page navigated during evaluate (e.g. Cloudflare auto-redirect) — treat as no challenge
    return 'none';
  }
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

/**
 * Creates a new browser context with stealth scripts and optional ad/cookie blocking.
 */
async function createBrowserContext(blockAdsCookies: boolean, url: string): Promise<BrowserContext> {
  if (!browser || !browser.isConnected()) {
    console.log('Browser disconnected — reinitializing...');
    await initBrowser();
  }
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 },
    screen: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
    userAgent: CHROME_USER_AGENT,
    locale: 'en-US',
    timezoneId: 'America/New_York',
    javaScriptEnabled: true,
    extraHTTPHeaders: {
      'Accept-Language': 'en-US,en;q=0.9',
      'sec-ch-ua': '"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"',
      'sec-ch-ua-mobile': '?0',
      'sec-ch-ua-platform': '"Windows"',
      'Upgrade-Insecure-Requests': '1',
    },
  });

  await addStealthScripts(context);

  if (blockAdsCookies) {
    await setupAdBlocking(context);
    await setupCookieBannerBlocker(context);
    console.log(`Ad blocking + cookie banner blocking enabled for ${url}`);
  }

  return context;
}

/**
 * Navigates to a URL and waits for the page to be fully rendered.
 */
async function navigateAndWaitForRender(page: Page, url: string, blockAdsCookies: boolean): Promise<void> {
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await handleCloudflareChallenge(page);
  await page.waitForLoadState('load').catch(() => {
    console.log(`Load event timeout for ${url}, proceeding`);
  });
  await page.waitForLoadState('networkidle').catch(() => {
    console.log(`Initial networkidle timeout for ${url}, proceeding`);
  });
  await waitForRenderComplete(page);
  await scrollFullPage(page);
  await forceLoadAllResources(page);
  await page.waitForLoadState('networkidle').catch(() => {
    console.log(`Post-scroll networkidle timeout for ${url}, proceeding`);
  });
  await waitForAllResources(page);
  await page.waitForTimeout(1000);

  if (blockAdsCookies) {
    await finalCookieBannerCleanup(page);
    await page.waitForTimeout(300);
  }
}

/**
 * Generates an element map of selectable structural elements on the page.
 * Returns elements with their CSS selector, XPath, bounding rect, and text preview.
 */
async function generateElementMap(page: Page): Promise<Array<{
  selector: string;
  xpath: string;
  tag: string;
  rect: { x: number; y: number; w: number; h: number };
  text_preview: string;
  semantic_role: string;
}>> {
  return page.evaluate(() => {
    const STRUCTURAL_TAGS = new Set([
      'section', 'article', 'div', 'table', 'ul', 'ol', 'nav',
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'p', 'form', 'main',
      'header', 'footer', 'aside', 'figure', 'figcaption', 'details',
      'blockquote', 'pre', 'fieldset', 'dl',
    ]);
    const MIN_WIDTH = 50;
    const MIN_HEIGHT = 20;
    const MAX_ELEMENTS = 200;

    function getUniqueSelector(el: Element): string {
      if (el.id) return `#${CSS.escape(el.id)}`;
      const dataTestId = el.getAttribute('data-testid');
      if (dataTestId) return `[data-testid="${CSS.escape(dataTestId)}"]`;
      const dataId = el.getAttribute('data-id');
      if (dataId) return `[data-id="${CSS.escape(dataId)}"]`;

      const parts: string[] = [];
      let current: Element | null = el;
      while (current && current !== document.body && current !== document.documentElement) {
        let part = current.tagName.toLowerCase();
        if (current.id) {
          parts.unshift(`#${CSS.escape(current.id)} > ${part}`);
          break;
        }
        const parent = current.parentElement;
        if (parent) {
          const siblings = Array.from(parent.children).filter(c => c.tagName === current!.tagName);
          if (siblings.length > 1) {
            const idx = siblings.indexOf(current) + 1;
            part += `:nth-of-type(${idx})`;
          }
        }
        parts.unshift(part);
        current = parent;
      }
      return parts.join(' > ');
    }

    function getXPath(el: Element): string {
      if (el.id) return `//*[@id="${el.id}"]`;
      const parts: string[] = [];
      let current: Element | null = el;
      while (current && current !== document.body) {
        let tag = current.tagName.toLowerCase();
        const parent = current.parentElement;
        if (parent) {
          const siblings = Array.from(parent.children).filter(c => c.tagName === current!.tagName);
          if (siblings.length > 1) {
            const idx = siblings.indexOf(current) + 1;
            tag += `[${idx}]`;
          }
        }
        parts.unshift(tag);
        current = parent;
      }
      return '//' + parts.join('/');
    }

    function getSemanticRole(el: Element): string {
      const role = el.getAttribute('role');
      if (role) return role;
      const tag = el.tagName.toLowerCase();
      const roleMap: Record<string, string> = {
        nav: 'navigation', header: 'banner', footer: 'contentinfo',
        main: 'main', aside: 'complementary', form: 'form',
        table: 'table', article: 'article', section: 'region',
      };
      return roleMap[tag] || 'generic';
    }

    const elements: Array<{
      selector: string; xpath: string; tag: string;
      rect: { x: number; y: number; w: number; h: number };
      text_preview: string; semantic_role: string;
    }> = [];

    const allElements = document.querySelectorAll(Array.from(STRUCTURAL_TAGS).join(','));
    for (const el of allElements) {
      if (elements.length >= MAX_ELEMENTS) break;
      const rect = el.getBoundingClientRect();
      if (rect.width < MIN_WIDTH || rect.height < MIN_HEIGHT) continue;
      const style = getComputedStyle(el);
      if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') continue;

      const text = (el.textContent || '').trim().slice(0, 100);
      elements.push({
        selector: getUniqueSelector(el),
        xpath: getXPath(el),
        tag: el.tagName.toLowerCase(),
        rect: {
          x: Math.round(rect.x + window.scrollX),
          y: Math.round(rect.y + window.scrollY),
          w: Math.round(rect.width),
          h: Math.round(rect.height),
        },
        text_preview: text,
        semantic_role: getSemanticRole(el),
      });
    }

    return elements;
  });
}

/**
 * Finds an element by CSS selector (fallback XPath), scrolls it into view,
 * and returns a clipped screenshot and element-only text/HTML.
 */
async function extractWithSelector(
  page: Page,
  selector: string,
  selectorXpath: string,
  offsets: { top: number; right: number; bottom: number; left: number },
): Promise<{
  screenshot: Buffer;
  html: string;
  text: string;
  selectorMatched: boolean;
}> {
  // Try CSS selector first, then XPath
  let elementHandle = selector ? await page.$(selector).catch(() => null) : null;
  if (!elementHandle && selectorXpath) {
    const xpathResults = await page.$(`xpath=${selectorXpath}`).catch(() => null);
    elementHandle = xpathResults;
  }

  if (!elementHandle) {
    // Fallback to full page
    const screenshot = await page.screenshot({ fullPage: true, type: 'png', animations: 'disabled' });
    const html = await page.content();
    const text = await page.evaluate(() => document.body.innerText);
    return { screenshot, html, text, selectorMatched: false };
  }

  await elementHandle.scrollIntoViewIfNeeded().catch(() => {});
  await page.waitForTimeout(500);

  const box = await elementHandle.boundingBox();
  if (!box) {
    const screenshot = await page.screenshot({ fullPage: true, type: 'png', animations: 'disabled' });
    const html = await page.content();
    const text = await page.evaluate(() => document.body.innerText);
    return { screenshot, html, text, selectorMatched: false };
  }

  // Apply offsets
  const clip = {
    x: Math.max(0, box.x - offsets.left),
    y: Math.max(0, box.y - offsets.top),
    width: box.width + offsets.left + offsets.right,
    height: box.height + offsets.top + offsets.bottom,
  };

  const screenshot = await page.screenshot({ clip, type: 'png', animations: 'disabled' });

  const elementData = await elementHandle.evaluate((el: Element) => ({
    html: el.outerHTML,
    text: (el as HTMLElement).innerText || el.textContent || '',
  }));

  return {
    screenshot,
    html: elementData.html,
    text: elementData.text,
    selectorMatched: true,
  };
}

app.post('/preview', async (c) => {
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
    context = await createBrowserContext(blockAdsCookies, url);
    page = await context.newPage();

    console.log(`Preview: Navigating to ${url}`);
    await navigateAndWaitForRender(page, url, blockAdsCookies);

    const viewport = await page.evaluate(() => ({
      width: document.documentElement.clientWidth,
      height: document.documentElement.clientHeight,
    }));

    const pageHeight = await page.evaluate(() => document.documentElement.scrollHeight);

    // Suppress horizontal overflow so fullPage screenshot doesn't capture blank space on the right
    await page.evaluate(() => {
      document.documentElement.style.overflowX = 'hidden';
      document.body.style.overflowX = 'hidden';
    });

    const screenshot = await page.screenshot({
      fullPage: true,
      type: 'png',
      animations: 'disabled',
      clip: { x: 0, y: 0, width: viewport.width, height: pageHeight },
    });

    const elements = await generateElementMap(page);

    const durationMs = Date.now() - startedAt;
    console.log(`Preview completed for ${url} in ${durationMs}ms (elements=${elements.length})`);

    await page.close();
    await context.close();

    return c.json({
      screenshot_base64: screenshot.toString('base64'),
      viewport,
      page_height: pageHeight,
      elements,
    });
  } catch (error) {
    console.error(`Preview error for ${url}:`, error);
    if (page) await page.close().catch(() => {});
    if (context) await context.close().catch(() => {});

    const statusCode = error instanceof CloudflareError ? 403 : 500;
    return c.json({ error: error instanceof Error ? error.message : 'Unknown error' }, statusCode);
  }
});

app.post('/extract', async (c) => {
  const body = await c.req.json().catch(() => null);
  const url = typeof body === 'object' && body !== null && 'url' in body ? (body as { url?: string }).url : undefined;
  if (!url) {
    return c.json({ error: 'URL is required' }, 400);
  }

  const blockAdsCookies = typeof body === 'object' && body !== null && 'block_ads_cookies' in body
    ? Boolean((body as { block_ads_cookies?: boolean }).block_ads_cookies)
    : false;

  const selectorCSS = typeof body === 'object' && body !== null && 'selector' in body
    ? String((body as { selector?: string }).selector || '')
    : '';

  const selectorXPath = typeof body === 'object' && body !== null && 'selector_xpath' in body
    ? String((body as { selector_xpath?: string }).selector_xpath || '')
    : '';

  const selectorOffsets = typeof body === 'object' && body !== null && 'selector_offsets' in body
    ? (body as { selector_offsets?: { top?: number; right?: number; bottom?: number; left?: number } }).selector_offsets || {}
    : {};
  const offsets = {
    top: Number(selectorOffsets.top) || 0,
    right: Number(selectorOffsets.right) || 0,
    bottom: Number(selectorOffsets.bottom) || 0,
    left: Number(selectorOffsets.left) || 0,
  };

  const hasSelector = selectorCSS !== '' || selectorXPath !== '';

  const startedAt = Date.now();
  let context: BrowserContext | undefined;
  let page: Page | undefined;
  try {
    context = await createBrowserContext(blockAdsCookies, url);
    page = await context.newPage();

    console.log(`Navigating to ${url}`);
    await navigateAndWaitForRender(page, url, blockAdsCookies);

    let title: string;
    let content: string;
    let text: string;
    let screenshot: Buffer;
    let selectorMatched = true;

    if (hasSelector) {
      // Element-targeted extraction
      title = await page.title();
      const result = await extractWithSelector(page, selectorCSS, selectorXPath, offsets);
      screenshot = result.screenshot;
      content = result.html;
      text = result.text;
      selectorMatched = result.selectorMatched;
    } else {
      // Full page extraction (original behavior)
      title = await page.title();
      content = await page.content();
      text = await page.evaluate((): string => document.body.innerText);
      const pageWidth = await page.evaluate(() => document.documentElement.clientWidth);
      const fullHeight = await page.evaluate(() => document.documentElement.scrollHeight);
      await page.evaluate(() => {
        document.documentElement.style.overflowX = 'hidden';
        document.body.style.overflowX = 'hidden';
      });
      screenshot = await page.screenshot({
        fullPage: true,
        type: 'png',
        animations: 'disabled',
        clip: { x: 0, y: 0, width: pageWidth, height: fullHeight },
      });
    }

    const durationMs = Date.now() - startedAt;
    console.log(
      `Extraction completed for ${url} in ${durationMs}ms (title_len=${title.length}, html_len=${content?.length || 0}, text_len=${text.length}, screenshot_bytes=${screenshot.length}, selector_matched=${selectorMatched})`
    );

    await page.close();
    await context.close();

    return c.json({
      title,
      html: content || '',
      text: text,
      screenshot_base64: screenshot.toString('base64'),
      selector_matched: selectorMatched,
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
