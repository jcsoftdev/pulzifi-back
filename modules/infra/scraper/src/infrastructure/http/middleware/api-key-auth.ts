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
 * Middleware that validates the X-API-Key header against a known secret.
 *
 * If `apiKey` is empty/undefined (development mode), the middleware logs a
 * WARNING at boot time and allows all requests through.
 *
 * Returns 401 if the header is missing or does not match.
 */
export function apiKeyAuthMiddleware(apiKey: string | undefined): MiddlewareHandler {
  const keyIsSet = typeof apiKey === "string" && apiKey.length > 0;

  if (!keyIsSet) {
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
