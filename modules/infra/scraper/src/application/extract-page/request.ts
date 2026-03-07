import type { SelectorOffsets } from "../../domain/value-objects/selector-config";

export interface SectionRequest {
  id: string;
  selector?: string;
  selectorXpath?: string;
  selectorOffsets?: SelectorOffsets;
}

export interface ExtractPageRequest {
  url: string;
  block_ads_cookies?: boolean;
  selector?: string;
  selector_xpath?: string;
  selector_offsets?: SelectorOffsets;
  sections?: SectionRequest[];
}
