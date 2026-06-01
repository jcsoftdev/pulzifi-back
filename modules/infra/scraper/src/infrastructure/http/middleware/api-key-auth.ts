import type { MiddlewareHandler } from "hono";

/**
 * Constant-time string comparison to defend against timing attacks.
 * Returns true if `a` and `b` are equal.
 */
function safeEqual(a: string, b: string): boolean {
  if (a.length !== b.length) {
    // Still iterate to prevent length-based timing leakage
    let diff = 0;
    const maxLen = Math.max(a.length, b.length);
    for (let i = 0; i < maxLen; i++) {
      diff |= (a.charCodeAt(i) ?? 0) ^ (b.charCodeAt(i) ?? 0);
    }
    return diff === 0 && a.length === b.length;
  }
  let diff = 0;
  for (let i = 0; i < a.length; i++) {
    diff |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return diff === 0;
}

/**
 * Detects whether the process is running in production.
 * Checks NODE_ENV and ENVIRONMENT env vars (either triggers production mode).
 */
function isProduction(): boolean {
  return (
    process.env.NODE_ENV === "production" ||
    process.env.ENVIRONMENT === "production"
  );
}

/**
 * Middleware that validates the X-API-Key header against a known secret.
 *
 * Production (NODE_ENV=production or ENVIRONMENT=production):
 *   If SCRAPER_API_KEY is unset or empty → process exits with a fatal error.
 *   All requests without a valid key → 401 Unauthorized.
 *
 * Development (default):
 *   If SCRAPER_API_KEY is unset → logs a WARNING and allows all requests through.
 *
 * Returns 401 if the header is missing or does not match.
 */
export function apiKeyAuthMiddleware(apiKey: string | undefined): MiddlewareHandler {
  const keyIsSet = typeof apiKey === "string" && apiKey.length > 0;

  if (!keyIsSet) {
    if (isProduction()) {
      console.error(
        "[security] FATAL: SCRAPER_API_KEY is not set in production. " +
        "Refusing to start without scraper authentication.",
      );
      process.exit(1);
    }
    console.warn(
      "[security] WARNING: SCRAPER_API_KEY is not set. " +
      "All requests are allowed. Set this variable in production.",
    );
  }

  return async (c, next) => {
    // Dev mode: no key configured → pass through
    if (!keyIsSet) {
      return next();
    }

    const provided = c.req.header("x-api-key") ?? "";

    if (!safeEqual(provided, apiKey as string)) {
      return c.json({ error: "Unauthorized" }, 401);
    }

    return next();
  };
}
