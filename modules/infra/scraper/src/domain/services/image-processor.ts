export interface IImageProcessor {
  pngToWebpBase64(pngBuffer: Buffer, quality?: number): Promise<string>;
  pngToBase64(pngBuffer: Buffer): string;
  cropToWidthAndConvert(
    pngBuffer: Buffer,
    width: number,
    quality?: number,
  ): Promise<string>;
}
