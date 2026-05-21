import { describe, it, expect, beforeEach, afterEach } from "bun:test";
import { Hono } from "hono";
import { apiKeyAuthMiddleware } from "./api-key-auth";

function createTestApp(apiKey: string | undefined): Hono {
  const app = new Hono();
  app.use("/secure", apiKeyAuthMiddleware(apiKey));
  app.post("/secure", (c) => c.json({ ok: true }));
  return app;
}

describe("apiKeyAuthMiddleware", () => {
  it("returns 200 when correct key is provided", async () => {
    const app = createTestApp("test-secret-key");
    const res = await app.fetch(
      new Request("http://localhost/secure", {
        method: "POST",
        headers: { "X-API-Key": "test-secret-key" },
      }),
    );
    expect(res.status).toBe(200);
  });

  it("returns 401 when key is missing", async () => {
    const app = createTestApp("test-secret-key");
    const res = await app.fetch(
      new Request("http://localhost/secure", { method: "POST" }),
    );
    expect(res.status).toBe(401);
    const json = await res.json();
    expect(json.error).toBeDefined();
  });

  it("returns 401 when wrong key is provided", async () => {
    const app = createTestApp("test-secret-key");
    const res = await app.fetch(
      new Request("http://localhost/secure", {
        method: "POST",
        headers: { "X-API-Key": "wrong-key" },
      }),
    );
    expect(res.status).toBe(401);
  });

  it("returns 401 when key is empty string", async () => {
    const app = createTestApp("test-secret-key");
    const res = await app.fetch(
      new Request("http://localhost/secure", {
        method: "POST",
        headers: { "X-API-Key": "" },
      }),
    );
    expect(res.status).toBe(401);
  });

  it("allows through when env var is undefined (dev mode)", async () => {
    const app = createTestApp(undefined);
    const res = await app.fetch(
      new Request("http://localhost/secure", { method: "POST" }),
    );
    expect(res.status).toBe(200);
  });

  it("allows through when env var is empty string (dev mode)", async () => {
    const app = createTestApp("");
    const res = await app.fetch(
      new Request("http://localhost/secure", { method: "POST" }),
    );
    expect(res.status).toBe(200);
  });

  it("does not allow a partial match to pass", async () => {
    const app = createTestApp("full-secret-key");
    const res = await app.fetch(
      new Request("http://localhost/secure", {
        method: "POST",
        headers: { "X-API-Key": "full-secret" },
      }),
    );
    expect(res.status).toBe(401);
  });

  it("is case-sensitive", async () => {
    const app = createTestApp("SecretKey");
    const res = await app.fetch(
      new Request("http://localhost/secure", {
        method: "POST",
        headers: { "X-API-Key": "secretkey" },
      }),
    );
    expect(res.status).toBe(401);
  });
});
