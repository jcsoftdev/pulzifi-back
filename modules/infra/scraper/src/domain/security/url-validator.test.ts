import { describe, it, expect } from "bun:test";
import { validateUrl, UrlValidationError } from "./url-validator";

/** Fake resolver that always returns a public IP — simulates a healthy DNS lookup. */
const publicResolver = async (_hostname: string): Promise<string[]> => [
  "93.184.216.34",
];

/** Fake resolver that always returns a private IP — simulates DNS rebinding. */
const privateResolver = async (_hostname: string): Promise<string[]> => [
  "192.168.1.100",
];

/** Fake resolver that always returns loopback — simulates DNS rebinding to localhost. */
const loopbackResolver = async (_hostname: string): Promise<string[]> => [
  "127.0.0.1",
];

describe("validateUrl — SSRF protection", () => {
  // ─── Valid public URLs ────────────────────────────────────────────────────
  it("accepts a valid public https URL (mocked DNS)", async () => {
    await expect(
      validateUrl("https://example.com", publicResolver),
    ).resolves.toBeUndefined();
  });

  it("accepts a valid public http URL (mocked DNS)", async () => {
    await expect(
      validateUrl("http://example.com", publicResolver),
    ).resolves.toBeUndefined();
  });

  it("accepts URL with path and query string (mocked DNS)", async () => {
    await expect(
      validateUrl("https://example.com/path?q=1&page=2", publicResolver),
    ).resolves.toBeUndefined();
  });

  // ─── Protocol validation ──────────────────────────────────────────────────
  it("rejects file:// protocol", async () => {
    await expect(validateUrl("file:///etc/passwd", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects ftp:// protocol", async () => {
    await expect(validateUrl("ftp://example.com/file", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects javascript: protocol", async () => {
    await expect(validateUrl("javascript:alert(1)", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects data: URI scheme", async () => {
    await expect(validateUrl("data:text/html,<h1>test</h1>", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects a bare hostname with no protocol", async () => {
    await expect(validateUrl("example.com", publicResolver)).rejects.toThrow(UrlValidationError);
  });

  // ─── Private IP ranges ────────────────────────────────────────────────────
  it("rejects localhost (127.0.0.1)", async () => {
    await expect(validateUrl("http://127.0.0.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects loopback hostname", async () => {
    await expect(validateUrl("http://localhost/admin", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 127.0.0.2 (loopback range)", async () => {
    await expect(validateUrl("http://127.0.0.2/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 10.0.0.1 (RFC 1918 class A)", async () => {
    await expect(validateUrl("http://10.0.0.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 10.255.255.255 (RFC 1918 class A edge)", async () => {
    await expect(validateUrl("http://10.255.255.255/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 172.16.0.1 (RFC 1918 class B)", async () => {
    await expect(validateUrl("http://172.16.0.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 172.31.255.255 (RFC 1918 class B edge)", async () => {
    await expect(validateUrl("http://172.31.255.255/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 192.168.1.1 (RFC 1918 class C)", async () => {
    await expect(validateUrl("http://192.168.1.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 169.254.0.1 (link-local)", async () => {
    await expect(validateUrl("http://169.254.0.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 169.254.169.254 (AWS metadata endpoint)", async () => {
    await expect(
      validateUrl("http://169.254.169.254/latest/meta-data/", publicResolver),
    ).rejects.toThrow(UrlValidationError);
  });

  it("rejects 100.64.0.1 (CGNAT shared address space)", async () => {
    await expect(validateUrl("http://100.64.0.1/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects 0.0.0.0", async () => {
    await expect(validateUrl("http://0.0.0.0/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects IPv6 loopback ::1", async () => {
    await expect(validateUrl("http://[::1]/", publicResolver)).rejects.toThrow(UrlValidationError);
  });

  it("rejects IPv6 unique-local fc00::/7", async () => {
    await expect(validateUrl("http://[fc00::1]/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  it("rejects IPv6 link-local fe80::/10", async () => {
    await expect(validateUrl("http://[fe80::1]/", publicResolver)).rejects.toThrow(
      UrlValidationError,
    );
  });

  // ─── DNS rebinding simulation (mocked DNS resolver) ───────────────────────
  it("rejects URL whose hostname resolves to a private IP (DNS rebinding)", async () => {
    await expect(
      validateUrl("https://evil.example.com/", privateResolver),
    ).rejects.toThrow(UrlValidationError);
  });

  it("rejects URL whose hostname resolves to loopback (DNS rebinding)", async () => {
    await expect(
      validateUrl("https://evil.example.com/", loopbackResolver),
    ).rejects.toThrow(UrlValidationError);
  });

  it("accepts URL whose hostname resolves to a public IP", async () => {
    await expect(
      validateUrl("https://example.com/", publicResolver),
    ).resolves.toBeUndefined();
  });

  // ─── Error shape ──────────────────────────────────────────────────────────
  it("UrlValidationError carries the original URL", async () => {
    try {
      await validateUrl("http://127.0.0.1/", publicResolver);
      throw new Error("should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(UrlValidationError);
      expect((err as UrlValidationError).url).toBe("http://127.0.0.1/");
    }
  });
});

describe("validateUrl — SSRF host allowlist", () => {
  /** Resolver that always returns a private IP (10.x range) for allowlist tests. */
  const tenResolver = async (_hostname: string): Promise<string[]> => ["10.0.0.5"];

  it("allows an allowlisted hostname that resolves to a private IP", async () => {
    await expect(
      validateUrl("http://e2e-fixture:8080/page", tenResolver, new Set(["e2e-fixture"])),
    ).resolves.toBeUndefined();
  });

  it("still blocks a non-allowlisted hostname resolving to a private IP", async () => {
    await expect(
      validateUrl("http://internal-svc:8080/", tenResolver, new Set(["e2e-fixture"])),
    ).rejects.toThrow(UrlValidationError);
  });

  it("empty allowlist preserves current blocking behavior", async () => {
    await expect(
      validateUrl("http://e2e-fixture:8080/page", tenResolver, new Set()),
    ).rejects.toThrow(UrlValidationError);
  });

  it("allowlist match is case-insensitive", async () => {
    await expect(
      validateUrl("http://E2E-Fixture:8080/page", tenResolver, new Set(["e2e-fixture"])),
    ).resolves.toBeUndefined();
  });
});
