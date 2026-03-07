import type { IBrowserService } from "../../domain/services/browser-service";

export class HealthCheckHandler {
  constructor(private browserService: IBrowserService) {}

  handle(): { status: string } {
    return {
      status: this.browserService.isHealthy() ? "ok" : "unhealthy",
    };
  }
}
