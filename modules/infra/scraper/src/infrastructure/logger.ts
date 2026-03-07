let requestCounter = 0;

export function createRequestId(): string {
  return `req-${++requestCounter}-${Date.now().toString(36)}`;
}

export function log(
  tag: string,
  message: string,
  meta?: Record<string, unknown>,
) {
  const ts = new Date().toISOString();
  const metaStr = meta ? ` ${JSON.stringify(meta)}` : "";
  console.log(`${ts} [${tag}] ${message}${metaStr}`);
}

export function logError(
  tag: string,
  message: string,
  error?: unknown,
  meta?: Record<string, unknown>,
) {
  const ts = new Date().toISOString();
  const errMsg =
    error instanceof Error ? error.message : error ? String(error) : "";
  const metaStr = meta ? ` ${JSON.stringify(meta)}` : "";
  console.error(`${ts} [${tag}] ${message}: ${errMsg}${metaStr}`);
}

export function createTimer() {
  const start = performance.now();
  return {
    elapsed: () => `${(performance.now() - start).toFixed(0)}ms`,
  };
}
