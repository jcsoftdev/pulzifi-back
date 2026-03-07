import type { BoundingRect, Viewport } from "../value-objects/viewport";

export interface PreviewElement {
  selector: string;
  xpath: string;
  tag: string;
  rect: BoundingRect;
  text_preview: string;
  semantic_role: string;
}

export interface PreviewResult {
  screenshot_base64: string;
  viewport: Viewport;
  page_height: number;
  elements: PreviewElement[];
}
