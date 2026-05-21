import { describe, it, expect } from "bun:test";
import { Hono } from "hono";
import { bodyLimitMiddleware } from "./body-limit";

function createTestApp(maxBytes: number): Hono {
  const app = new Hono();
  app.use("/test", bodyLimitMiddleware(maxBytes));
  app.post("/test", async (c) => {
    const body = await c.req.text();
    return c.json({ length: body.length });
  });
  return app;
}

describe("bodyLimitMiddleware", () => {
  const MAX_BYTES = 16 * 1024; // 16 KB

  it("allows a small body (1 KB)", async () => {
    const app = createTestApp(MAX_BYTES);
    const body = "x".repeat(1024);
    const res = await app.fetch(
      new Request("http://localhost/test", {
        method: "POST",
        headers: { "Content-Type": "text/plain", "Content-Length": String(body.length) },
        body,
      }),
    );
    expect(res.status).toBe(200);
    const json = await res.json();
    expect(json.length).toBe(1024);
  });

  it("allows a body exactly at the limit", async () => {
    const app = createTestApp(MAX_BYTES);
    const body = "x".repeat(MAX_BYTES);
    const res = await app.fetch(
      new Request("http://localhost/test", {
        method: "POST",
        headers: { "Content-Type": "text/plain", "Content-Length": String(body.length) },
        body,
      }),
    );
    expect(res.status).toBe(200);
  });

  it("rejects a body 1 byte over the limit via Content-Length header", async () => {
    const app = createTestApp(MAX_BYTES);
    const body = "x".repeat(MAX_BYTES + 1);
    const res = await app.fetch(
      new Request("http://localhost/test", {
        method: "POST",
        headers: { "Content-Type": "text/plain", "Content-Length": String(body.length) },
        body,
      }),
    );
    expect(res.status).toBe(413);
    const json = await res.json();
    expect(json.error).toContain("too large");
  });

  it("rejects a 1 MB body", async () => {
    const app = createTestApp(MAX_BYTES);
    const body = "x".repeat(1024 * 1024);
    const res = await app.fetch(
      new Request("http://localhost/test", {
        method: "POST",
        headers: { "Content-Type": "text/plain", "Content-Length": String(body.length) },
        body,
      }),
    );
    expect(res.status).toBe(413);
  });

  it("does not interfere with GET requests", async () => {
    const app = new Hono();
    app.use("/ping", bodyLimitMiddleware(MAX_BYTES));
    app.get("/ping", (c) => c.json({ ok: true }));

    const res = await app.fetch(new Request("http://localhost/ping"));
    expect(res.status).toBe(200);
  });
});
