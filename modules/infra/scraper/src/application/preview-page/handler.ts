import type { IBrowserService, ProgressCallback } from "../../domain/services/browser-service";
import type { PreviewResult } from "../../domain/entities/preview-result";
import type { PreviewPageRequest } from "./request";

export class PreviewPageHandler {
  constructor(private browserService: IBrowserService) {}

  async handle(
    request: PreviewPageRequest,
    onProgress?: ProgressCallback,
  ): Promise<PreviewResult> {
    return this.browserService.preview(
      {
        url: request.url,
        blockAdsCookies: request.block_ads_cookies ?? false,
        sectionDepth: request.section_depth,
      },
      onProgress,
    );
  }
}
