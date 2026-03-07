export interface SectionResult {
  id: string;
  screenshot_base64: string;
  html: string;
  text: string;
  selector_matched: boolean;
}

export interface ExtractionResult {
  title: string;
  html: string;
  text: string;
  screenshot_base64: string;
  selector_matched: boolean;
  sections?: SectionResult[];
}
