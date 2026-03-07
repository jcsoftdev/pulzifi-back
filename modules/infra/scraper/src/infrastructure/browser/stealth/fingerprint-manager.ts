import { FingerprintGenerator } from "fingerprint-generator";
import { FingerprintInjector } from "fingerprint-injector";
import type { BrowserContext } from "patchright";

const generator = new FingerprintGenerator();
const injector = new FingerprintInjector();

/**
 * Generates a realistic browser fingerprint and injects it
 * into a browser context for anti-detection.
 */
export async function applyFingerprint(
  context: BrowserContext,
): Promise<void> {
  const fingerprint = generator.getFingerprint({
    browsers: ["chrome"],
    operatingSystems: ["linux"],
    devices: ["desktop"],
  });

  await injector.attachFingerprintToPlaywright(context as any, fingerprint);
}
