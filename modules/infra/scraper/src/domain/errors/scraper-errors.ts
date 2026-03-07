export class ScraperError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number = 500,
  ) {
    super(message);
    this.name = "ScraperError";
  }
}

export class ExtractionError extends ScraperError {
  constructor(message: string) {
    super(message, "EXTRACTION_ERROR", 500);
    this.name = "ExtractionError";
  }
}

export class BrowserError extends ScraperError {
  constructor(message: string) {
    super(message, "BROWSER_ERROR", 503);
    this.name = "BrowserError";
  }
}

export class NavigationError extends ScraperError {
  constructor(message: string, public readonly url: string) {
    super(message, "NAVIGATION_ERROR", 502);
    this.name = "NavigationError";
  }
}

export class TimeoutError extends ScraperError {
  constructor(message: string) {
    super(message, "TIMEOUT_ERROR", 504);
    this.name = "TimeoutError";
  }
}
