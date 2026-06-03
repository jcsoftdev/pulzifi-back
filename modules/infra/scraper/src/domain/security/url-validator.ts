import { dns } from "bun";

export class UrlValidationError extends Error {
  constructor(
    message: string,
    public readonly url: string,
  ) {
    super(message);
    this.name = "UrlValidationError";
  }
}

/**
 * Resolver function signature — accepts a hostname and returns a list of
 * resolved IP address strings. The default implementation uses Bun's built-in
 * DNS resolver. Tests can inject a fake resolver to simulate DNS rebinding.
 */
export type DnsResolver = (hostname: string) => Promise<string[]>;

const defaultResolver: DnsResolver = async (hostname: string): Promise<string[]> => {
  try {
    // Bun's dns.lookup returns typed DNSLookup records ({ address, family, ttl }),
    // so we map to the plain address strings the DnsResolver contract requires.
    // (The older dns.resolve returns record objects too, which broke isPrivateIP.)
    const v4 = await dns.lookup(hostname, { family: 4 }).catch(() => []);
    const v6 = await dns.lookup(hostname, { family: 6 }).catch(() => []);
    return [...v4, ...v6].map((r) => r.address);
  } catch {
    return [];
  }
};

/**
 * Returns true if the given IPv4 address string falls within any blocked range:
 *   127.0.0.0/8    — loopback
 *   10.0.0.0/8     — RFC 1918
 *   172.16.0.0/12  — RFC 1918
 *   192.168.0.0/16 — RFC 1918
 *   169.254.0.0/16 — link-local (incl. AWS metadata)
 *   100.64.0.0/10  — CGNAT shared address space
 *   0.0.0.0/8      — "This" network
 */
function isPrivateIPv4(ip: string): boolean {
  const parts = ip.split(".").map(Number);
  if (parts.length !== 4 || parts.some((p) => isNaN(p) || p < 0 || p > 255)) {
    return false;
  }
  const [a, b, c, d] = parts;

  if (a === 0) return true;                                           // 0.0.0.0/8
  if (a === 127) return true;                                         // 127.0.0.0/8 loopback
  if (a === 10) return true;                                          // 10.0.0.0/8
  if (a === 172 && b >= 16 && b <= 31) return true;                  // 172.16.0.0/12
  if (a === 192 && b === 168) return true;                            // 192.168.0.0/16
  if (a === 169 && b === 254) return true;                            // 169.254.0.0/16
  if (a === 100 && b >= 64 && b <= 127) return true;                  // 100.64.0.0/10 CGNAT

  void c; void d; // suppress unused-var lint
  return false;
}

/**
 * Returns true if the given IPv6 address string falls within any blocked range:
 *   ::1            — loopback
 *   fc00::/7       — unique-local (fc00:: and fd00::)
 *   fe80::/10      — link-local
 */
function isPrivateIPv6(ip: string): boolean {
  // Normalise: strip surrounding brackets if present
  const addr = ip.replace(/^\[|\]$/g, "").toLowerCase();

  if (addr === "::1") return true;

  // fc00::/7 — covers fc00:: through fdff::
  if (addr.startsWith("fc") || addr.startsWith("fd")) return true;

  // fe80::/10 — link-local: fe80–febf
  if (addr.startsWith("fe8") || addr.startsWith("fe9") ||
      addr.startsWith("fea") || addr.startsWith("feb")) {
    return true;
  }

  return false;
}

function isPrivateIP(ip: string): boolean {
  // Fail closed: any entry that is not a plain string (e.g. a resolver that
  // returns DNS record objects) is treated as unsafe rather than crashing.
  if (typeof ip !== "string") return true;
  if (ip.includes(":")) return isPrivateIPv6(ip);
  return isPrivateIPv4(ip);
}

/** Well-known private hostnames that should always be blocked. */
const BLOCKED_HOSTNAMES = new Set(["localhost", "broadcasthost"]);

/**
 * Validates that a URL is safe to scrape:
 *  1. Protocol must be http or https
 *  2. Hostname must not be a known private hostname
 *  3. All IPs the hostname resolves to must be public
 *
 * @param rawUrl    The URL string submitted by the caller
 * @param resolver  Optional custom DNS resolver (used in tests)
 * @param allowlist Optional set of hostnames (lowercased) permitted to bypass
 *                  private-IP blocking. Empty by default (secure-by-default).
 *                  Used exclusively by the E2E stack via SSRF_HOST_ALLOWLIST.
 */
export async function validateUrl(
  rawUrl: string,
  resolver: DnsResolver = defaultResolver,
  allowlist: Set<string> = new Set(),
): Promise<void> {
  let parsed: URL;
  try {
    parsed = new URL(rawUrl);
  } catch {
    throw new UrlValidationError(`Invalid URL: ${rawUrl}`, rawUrl);
  }

  // 1. Protocol check
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new UrlValidationError(
      `Forbidden URL: only http and https protocols are allowed`,
      rawUrl,
    );
  }

  const hostname = parsed.hostname.replace(/^\[|\]$/g, ""); // strip IPv6 brackets

  // Allowlist short-circuit (secure-by-default: empty allowlist = no effect).
  // Used only by the E2E stack via SSRF_HOST_ALLOWLIST to reach a controlled fixture.
  if (allowlist.has(hostname.toLowerCase())) {
    return;
  }

  // 2. Block known private hostnames
  if (BLOCKED_HOSTNAMES.has(hostname.toLowerCase())) {
    throw new UrlValidationError(`Forbidden URL: blocked hostname`, rawUrl);
  }

  // 3. If hostname is already an IP address, validate it directly
  if (isPrivateIP(hostname)) {
    throw new UrlValidationError(
      `Forbidden URL: private/internal IP address`,
      rawUrl,
    );
  }

  // 4. DNS resolution check — defend against DNS rebinding
  const resolvedIPs = await resolver(hostname);
  for (const ip of resolvedIPs) {
    if (isPrivateIP(ip)) {
      throw new UrlValidationError(
        `Forbidden URL: hostname resolves to a private/internal IP address`,
        rawUrl,
      );
    }
  }
}
