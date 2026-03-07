import type { IBrowserService } from "../../domain/services/browser-service";
import type { ExtractionResult } from "../../domain/entities/extraction-result";
import type { ExtractPageRequest } from "./request";

export class ExtractPageHandler {
  constructor(private browserService: IBrowserService) {}

  async handle(request: ExtractPageRequest): Promise<ExtractionResult> {
    const selectorConfig =
      request.selector || request.selector_xpath
        ? {
            selector: request.selector,
            selectorXpath: request.selector_xpath,
            selectorOffsets: request.selector_offsets,
          }
        : undefined;

    const sections = request.sections?.map((s) => ({
      id: s.id,
      selector: s.selector,
      selectorXpath: s.selectorXpath,
      selectorOffsets: s.selectorOffsets,
    }));

    return this.browserService.extract({
      url: request.url,
      blockAdsCookies: request.block_ads_cookies ?? false,
      selector: selectorConfig,
      sections,
    });
  }
}
