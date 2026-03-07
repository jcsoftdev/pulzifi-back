import type { ExtractionResult } from "../entities/extraction-result";
import type { PreviewResult } from "../entities/preview-result";
import type { SelectorConfig, SectionConfig } from "../value-objects/selector-config";

export interface ExtractOptions {
  url: string;
  blockAdsCookies: boolean;
  selector?: SelectorConfig;
  sections?: SectionConfig[];
}

export interface PreviewOptions {
  url: string;
  blockAdsCookies: boolean;
  /** How many levels of meaningful sections to discover (1–5, default 2). */
  sectionDepth?: number;
}

export interface ProgressCallback {
  (step: number, totalSteps: number, message: string): void;
}

export interface IBrowserService {
  extract(options: ExtractOptions): Promise<ExtractionResult>;
  preview(
    options: PreviewOptions,
    onProgress?: ProgressCallback,
  ): Promise<PreviewResult>;
  isHealthy(): boolean;
  shutdown(): Promise<void>;
}
