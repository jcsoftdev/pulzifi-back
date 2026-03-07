import sharp from "sharp";
import type { IImageProcessor } from "../../domain/services/image-processor";

const SCREENSHOT_QUALITY = parseInt(
  process.env.SCREENSHOT_QUALITY || "80",
  10,
);

export class SharpImageProcessor implements IImageProcessor {
  async pngToWebpBase64(pngBuffer: Buffer, quality?: number): Promise<string> {
    const webpBuffer = await sharp(pngBuffer)
      .webp({ quality: quality ?? SCREENSHOT_QUALITY })
      .toBuffer();
    return webpBuffer.toString("base64");
  }

  pngToBase64(pngBuffer: Buffer): string {
    return pngBuffer.toString("base64");
  }

  /**
   * Crops a PNG to a specific width (keeping full height) and converts to WebP base64.
   * Used to strip horizontal overflow from fullPage screenshots.
   */
  async cropToWidthAndConvert(
    pngBuffer: Buffer,
    width: number,
    quality?: number,
  ): Promise<string> {
    const metadata = await sharp(pngBuffer).metadata();
    const imgWidth = metadata.width ?? width;
    const imgHeight = metadata.height ?? 1;

    // Only crop if the image is wider than desired
    const s =
      imgWidth > width
        ? sharp(pngBuffer).extract({
            left: 0,
            top: 0,
            width,
            height: imgHeight,
          })
        : sharp(pngBuffer);

    const webpBuffer = await s
      .webp({ quality: quality ?? SCREENSHOT_QUALITY })
      .toBuffer();
    return webpBuffer.toString("base64");
  }
}
