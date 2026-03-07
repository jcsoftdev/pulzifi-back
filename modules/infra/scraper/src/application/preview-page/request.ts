export interface PreviewPageRequest {
  url: string;
  block_ads_cookies?: boolean;
  /** How many levels of meaningful sections to discover (1–5, default 2). */
  section_depth?: number;
}
