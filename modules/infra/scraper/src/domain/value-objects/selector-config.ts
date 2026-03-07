export interface SelectorOffsets {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

export interface SelectorConfig {
  selector?: string;
  selectorXpath?: string;
  selectorOffsets?: SelectorOffsets;
}

export interface SectionConfig {
  id: string;
  selector?: string;
  selectorXpath?: string;
  selectorOffsets?: SelectorOffsets;
}
