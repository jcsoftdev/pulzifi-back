import type { MiddlewareHandler } from "hono";

/**
 * Middleware that rejects request bodies larger than `maxBytes`.
 *
 * Strategy:
 *  1. For non-POST/PUT/PATCH requests (no body), pass through immediately.
 *  2. Check Content-Length header first — fast path, no body read needed.
 *  3. Read the body as an ArrayBuffer and check its actual byte length
 *     (handles chunked transfer without Content-Length).
 */
export function bodyLimitMiddleware(maxBytes: number): MiddlewareHandler {
  return async (c, next) => {
    const method = c.req.method.toUpperCase();

    // GET / HEAD / DELETE etc. have no body — skip
    if (method !== "POST" && method !== "PUT" && method !== "PATCH") {
      return next();
    }

    // Fast path: Content-Length header
    const contentLengthHeader = c.req.header("content-length");
    if (contentLengthHeader !== undefined) {
      const contentLength = parseInt(contentLengthHeader, 10);
      if (!isNaN(contentLength) && contentLength > maxBytes) {
        return c.json(
          { error: `Request body too large. Maximum allowed: ${maxBytes} bytes` },
          413,
        );
      }
    }

    // Slow path: read and check actual body bytes (chunked transfer encoding)
    const rawBody = await c.req.arrayBuffer();
    if (rawBody.byteLength > maxBytes) {
      return c.json(
        { error: `Request body too large. Maximum allowed: ${maxBytes} bytes` },
        413,
      );
    }

    // Re-inject the body so downstream handlers can still read it
    // Hono stores the raw body internally after arrayBuffer() is called,
    // so subsequent c.req.json() / c.req.text() calls still work.
    return next();
  };
}
