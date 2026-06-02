import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./apps/web/e2e",
  timeout: 120_000,
  expect: { timeout: 15_000 },
  retries: 0,
  reporter: "list",
  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://jcsoftdev-inc.lvh.me:3000",
    trace: "retain-on-failure",
    ignoreHTTPSErrors: true,
  },
});
