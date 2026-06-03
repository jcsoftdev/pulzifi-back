import { test, expect, request, type APIRequestContext } from "@playwright/test";

// Hybrid API-driven E2E against the real Docker stack (docker-compose.e2e.yml):
// login (BFF, sets the access_token cookie) → create workspace → create page →
// run the REAL Playwright scraper against a controlled fixture → detect a content
// change → assert the content_change alert.
//
// Why API-driven and not UI-clicking: the frontend login form is a Payload-CMS block
// (no standalone /login route), so driving it would require standing up + migrating +
// seeding Payload. The BFF login endpoint sets the same auth cookie the UI relies on,
// so this exercises the identical backend pipeline (real scraper, real S3, real change
// detection, real alert) without the CMS provisioning. The only thing skipped is DOM
// clicking — no pipeline logic.

const TENANT = "jcsoftdev-inc";
const BASE = process.env.E2E_BASE_URL ?? "http://jcsoftdev-inc.lvh.me:3000";
const FIXTURE_SWITCH = process.env.E2E_FIXTURE_SWITCH ?? "http://localhost:8088/switch";
const TARGET_URL = "http://e2e-fixture:8080/page";
const EMAIL = "ajcarlos032@gmail.com";
const PASSWORD = "Asdfgh@1";
const H = { "X-Tenant": TENANT } as const;

type Check = { status: string } & Record<string, unknown>;

async function pollLatestCheck(
  api: APIRequestContext,
  pageId: string,
  predicate: (c: Check) => boolean,
  timeoutMs = 120_000,
): Promise<Check> {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const res = await api.get(`${BASE}/api/v1/monitoring/checks/page/${pageId}`, { headers: H });
    if (res.ok()) {
      const checks: Check[] = (await res.json()).checks ?? [];
      const latest = checks[0]; // list endpoint returns newest-first
      if (latest && predicate(latest)) return latest;
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
  throw new Error(`pollLatestCheck timed out for page ${pageId}`);
}

test("full pipeline (API): login → workspace → page → scraper → detect change → alert", async () => {
  // The context keeps a cookie jar; BFF login sets HttpOnly access_token on this origin.
  const api = await request.newContext({ baseURL: BASE });

  // 1. Login (BFF) — sets the access_token cookie used by all /api/v1 routes.
  const login = await api.post(`${BASE}/api/auth/login`, {
    headers: H,
    data: { email: EMAIL, password: PASSWORD },
  });
  expect(login.status(), "BFF login").toBe(200);

  // 2. Create workspace.
  const wsRes = await api.post(`${BASE}/api/v1/workspaces`, {
    headers: H,
    data: { name: "E2E Workspace", type: "personal", tags: [] },
  });
  expect(wsRes.status(), "create workspace").toBe(201);
  const workspaceId: string = (await wsRes.json()).id;
  expect(workspaceId).toBeTruthy();

  // 3. Create the monitored page pointing at the in-network fixture.
  const pageRes = await api.post(`${BASE}/api/v1/pages`, {
    headers: H,
    data: { workspace_id: workspaceId, name: "E2E Fixture Page", url: TARGET_URL, tags: [] },
  });
  expect(pageRes.status(), "create page").toBe(201);
  const pageId: string = (await pageRes.json()).id;
  expect(pageId).toBeTruthy();

  // 4. Baseline check — first run stores the baseline, detects no change.
  const run1 = await api.post(`${BASE}/api/v1/monitoring/checks/page/${pageId}/run`, { headers: H });
  expect(run1.status(), "baseline run accepted").toBe(202);
  await pollLatestCheck(api, pageId, (c) => c.status === "success");

  // 5. Mutate the fixture v1 → v2 (deterministic content change).
  const switchRes = await api.post(FIXTURE_SWITCH);
  expect(switchRes.ok(), "fixture switch").toBeTruthy();

  // The scheduler debounces "Run Now" within 5s of the previous claim (scheduler.go:
  // an atomic UPDATE that no-ops if last_checked_at is < 5s old, to dedupe concurrent
  // triggers). Wait past that window so the second run actually schedules a new check.
  await new Promise((r) => setTimeout(r, 6000));

  // 6. Second check — now the scraper sees changed content.
  const run2 = await api.post(`${BASE}/api/v1/monitoring/checks/page/${pageId}/run`, { headers: H });
  expect(run2.status(), "second run accepted").toBe(202);
  await pollLatestCheck(api, pageId, (c) => c.status === "success" && c.change_detected === true);

  // 7. Assert the content_change alert was created for this page.
  const alertsRes = await api.get(`${BASE}/api/v1/alerts/all`, { headers: H });
  expect(alertsRes.ok(), "list alerts").toBeTruthy();
  type Alert = { page_id: string; type: string } & Record<string, unknown>;
  const alerts: Alert[] = (await alertsRes.json()).data ?? [];
  const alert = alerts.find((a) => a.page_id === pageId && a.type === "content_change");
  expect(alert, "content_change alert exists for the page").toBeTruthy();

  await api.dispose();
});
