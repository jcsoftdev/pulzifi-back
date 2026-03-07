export interface Viewport {
  width: number;
  height: number;
}

export interface BoundingRect {
  x: number;
  y: number;
  w: number;
  h: number;
}

export const DEFAULT_VIEWPORT: Viewport = { width: 1440, height: 900 };
